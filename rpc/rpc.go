package rpc

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/0xsequence/bundler"
	"github.com/0xsequence/bundler/admin"
	"github.com/0xsequence/bundler/collector"
	"github.com/0xsequence/bundler/config"
	"github.com/0xsequence/bundler/contracts/gen/solabis/abivalidator"
	"github.com/0xsequence/bundler/endorser"
	"github.com/0xsequence/bundler/ipfs"
	"github.com/0xsequence/bundler/mempool"
	"github.com/0xsequence/bundler/p2p"
	"github.com/0xsequence/bundler/proto"
	"github.com/0xsequence/bundler/registry"
	"github.com/0xsequence/bundler/sender"
	"github.com/0xsequence/bundler/types"
	"github.com/0xsequence/ethkit/ethrpc"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httplog/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type metrics struct {
	methodTime *prometheus.HistogramVec
}

func createMetrics(reg prometheus.Registerer) *metrics {
	methodTime := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "rpc_method_time",
		Help:    "Method execution time",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 18),
	}, []string{"method"})

	if reg != nil {
		reg.MustRegister(methodTime)
	}

	return &metrics{
		methodTime: methodTime,
	}
}

type RPC struct {
	Config     *config.Config
	Log        *httplog.Logger
	Host       *p2p.Host
	HTTP       *http.Server
	Metrics    *metrics
	Registerer prometheus.Registerer
	Gatherer   prometheus.Gatherer

	mempool   mempool.Interface
	archive   *bundler.Archive
	collector *collector.Collector
	sender    sender.Interface
	executor  *abivalidator.OperationValidator
	ipfs      ipfs.Interface
	admin     *admin.Admin
	registry  registry.Interface

	running   int32
	startTime time.Time
}

func NewRPC(
	cfg *config.Config,
	logger *httplog.Logger,
	metrics prometheus.Registerer,
	gatherer prometheus.Gatherer,
	host *p2p.Host,
	mempool mempool.Interface,
	archive *bundler.Archive,
	provider *ethrpc.Provider,
	collector *collector.Collector,
	endorser endorser.Interface,
	ipfs ipfs.Interface,
	registry registry.Interface,
) (*RPC, error) {
	if !common.IsHexAddress(cfg.NetworkConfig.ValidatorContract) {
		return nil, fmt.Errorf("\"%v\" is not a valid operation validator contract", cfg.NetworkConfig.ValidatorContract)
	}
	validatorContract := common.HexToAddress(cfg.NetworkConfig.ValidatorContract)

	// HTTP Server
	httpServer := &http.Server{
		// Addr:              cfg.Service.Listen,
		Addr:              fmt.Sprintf(":%d", cfg.RPCPort),
		ReadTimeout:       45 * time.Second,
		WriteTimeout:      45 * time.Second,
		IdleTimeout:       45 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	executor, err := abivalidator.NewOperationValidator(validatorContract, provider)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to validator contract")
	}

	factory := sender.NewMnemonicWalletFactory(provider, cfg.Mnemonic)
	sender := sender.NewSender(&cfg.SendersConfig, logger, factory, provider, mempool, endorser, executor, collector, registry)
	sender.SetRegisterer(metrics)

	admin := admin.NewAdmin(logger, ipfs, mempool, registry)

	s := &RPC{
		archive:   archive,
		mempool:   mempool,
		sender:    sender,
		collector: collector,
		executor:  executor,
		ipfs:      ipfs,
		admin:     admin,
		registry:  registry,

		Config:     cfg,
		Log:        logger,
		Metrics:    createMetrics(metrics),
		Registerer: metrics,
		Gatherer:   gatherer,
		Host:       host,
		HTTP:       httpServer,
		startTime:  time.Now().UTC(),
	}
	return s, nil
}

func (s *RPC) Run(ctx context.Context) error {
	if s.IsRunning() {
		return fmt.Errorf("rpc: already running")
	}

	s.Log.Info(fmt.Sprintf("-> rpc: listening on %s", s.HTTP.Addr), "op", "run")

	atomic.StoreInt32(&s.running, 1)
	defer atomic.StoreInt32(&s.running, 0)

	// Setup HTTP server handler
	s.HTTP.Handler = s.handler()

	// Handle stop signal to ensure clean shutdown
	go func() {
		<-ctx.Done()
		s.Stop(context.Background())
	}()

	// Run the senders
	go s.sender.Run(ctx)

	// Start the http server and serve!
	err := s.HTTP.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (s *RPC) Stop(timeoutCtx context.Context) {
	if !s.IsRunning() || s.IsStopping() {
		return
	}
	atomic.StoreInt32(&s.running, 2)

	s.Log.Info("-> rpc: stopping..", "op", "stop")
	s.HTTP.Shutdown(timeoutCtx)
	s.Log.Info("-> rpc: stopped.", "op", "stop")
}

func (s *RPC) IsRunning() bool {
	return atomic.LoadInt32(&s.running) == 1
}

func (s *RPC) IsStopping() bool {
	return atomic.LoadInt32(&s.running) == 2
}

func (s *RPC) GetLogger(ctx context.Context) *slog.Logger {
	lg := httplog.LogEntry(ctx)
	return &lg
}

func (s *RPC) handler() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RealIP)

	// Metrics and heartbeat
	// r.Use(telemetry.Collector(s.Config.Telemetry, []string{"/rpc"}))
	r.Use(middleware.NoCache)
	// r.Use(honeybadger.Handler)
	r.Use(middleware.Heartbeat("/ping"))

	// HTTP request logger
	r.Use(httplog.RequestLogger(s.Log, []string{"/", "/ping", "/status", "/metrics", "/favicon.ico"}))

	// CORS
	// r.Use(s.corsHandler())
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	// Timeout any request after 28 seconds as Cloudflare has a 30 second limit anyways.
	//r.Use(middleware.Timeout(28 * time.Second))

	// Rate limiting
	// r.Use(httprate.LimitByIP(200, 1*time.Minute))

	// Static routes
	r.Get("/", s.metered(indexHandler))
	r.Get("/favicon.ico", s.metered(http.HandlerFunc(stubHandler(""))))
	r.Get("/status", s.metered(s.statusPage))
	r.Get("/peers", s.metered(s.peersPage))

	// Add prometheus metrics
	r.Get("/metrics", s.metered(promhttp.HandlerFor(s.Gatherer, promhttp.HandlerOpts{Registry: s.Registerer}).ServeHTTP))

	// Mount rpc endpoints
	bundlerRPCHandler := proto.NewBundlerServer(s)
	r.Post("/rpc/Bundler/*", s.metered(bundlerRPCHandler.ServeHTTP))

	// TODO: Add JWT for Admin space
	adminRPCHandler := proto.NewAdminServer(s.admin)
	r.Post("/rpc/Admin/*", s.metered(adminRPCHandler.ServeHTTP))

	return r
}

func (s *RPC) metered(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		s.Metrics.methodTime.WithLabelValues(r.URL.Path).Observe(time.Since(start).Seconds())
	})
}

// Ping is a health-check that returns an empty message.
func (s *RPC) Ping(ctx context.Context) (bool, error) {
	return true, nil
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("."))
}

func (s *RPC) SendOperation(ctx context.Context, pop *proto.Operation) (string, error) {
	op, err := types.NewOperationFromProto(pop)
	if err != nil {
		return "", err
	}

	// Always PIN these operations to IPFS
	// as they are being sent by the user, and
	// it is useful for debugging
	go op.ReportToIPFS(s.ipfs)

	err = s.mempool.AddOperation(ctx, op, true)
	if err != nil {
		return "", err
	}

	// If the operation is fine, broadcast it to the network
	s.Host.Broadcast(ctx, p2p.OperationTopic, op.ToProtoPure())

	return op.Hash(), nil
}

func (s RPC) Mempool(ctx context.Context) (*proto.MempoolView, error) {
	return s.mempool.Inspect(), nil
}

func (s RPC) Operations(ctx context.Context) (*proto.Operations, error) {
	return s.archive.Operations(ctx), nil
}

func (s *RPC) FeeAsks(ctx context.Context) (*proto.FeeAsks, error) {
	return s.collector.FeeAsks()
}

func stubHandler(respBody string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(respBody))
	})
}

var _ proto.Bundler = &RPC{}
