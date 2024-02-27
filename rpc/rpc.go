package rpc

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/0xsequence/bundler"
	"github.com/0xsequence/bundler/config"
	"github.com/0xsequence/bundler/contracts/gen/solabis/abivalidator"
	"github.com/0xsequence/bundler/endorser"
	"github.com/0xsequence/bundler/p2p"
	"github.com/0xsequence/bundler/proto"
	"github.com/0xsequence/bundler/types"
	"github.com/0xsequence/ethkit/ethrpc"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httplog/v2"
)

type RPC struct {
	Config *config.Config
	Log    *httplog.Logger
	Host   *p2p.Host
	HTTP   *http.Server

	mempool   *bundler.Mempool
	pruner    *bundler.Pruner
	archive   *bundler.Archive
	collector *bundler.Collector
	senders   []*bundler.Sender
	executor  *abivalidator.OperationValidator

	running   int32
	startTime time.Time
}

func NewRPC(cfg *config.Config, logger *httplog.Logger, host *p2p.Host, mempool *bundler.Mempool, archive *bundler.Archive, provider *ethrpc.Provider, collector *bundler.Collector, endorser endorser.Interface) (*RPC, error) {
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

	// Get the chain ID
	chainID, err := provider.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("unable to get chain ID: %w", err)
	}

	senders := make([]*bundler.Sender, 0, cfg.SendersConfig.NumSenders)
	for i := 0; i < int(cfg.SendersConfig.NumSenders); i++ {
		wallet, err := SetupWallet(cfg.Mnemonic, uint32(1+i), provider)
		if err != nil {
			return nil, fmt.Errorf("unable to create wallet for sender %v from hd node: %w", i, err)
		}
		logger.Info(fmt.Sprintf("sender %v: %v", i, wallet.Address()))
		senders = append(senders, bundler.NewSender(uint32(i), wallet, mempool, endorser, executor, collector, chainID))
	}

	pruner := bundler.NewPruner(mempool, endorser, logger)

	s := &RPC{
		archive:   archive,
		mempool:   mempool,
		senders:   senders,
		collector: collector,
		executor:  executor,
		pruner:    pruner,

		Config:    cfg,
		Log:       logger,
		Host:      host,
		HTTP:      httpServer,
		startTime: time.Now().UTC(),
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
	for _, sender := range s.senders {
		go sender.Run(ctx)
	}

	// Run the pruner
	go s.pruner.Run(ctx)

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
	r.Use(httplog.RequestLogger(s.Log, []string{"/", "/ping", "/status", "/favicon.ico"}))

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
	r.Get("/", indexHandler)
	r.Get("/favicon.ico", http.HandlerFunc(stubHandler("")))
	r.Get("/status", s.statusPage)
	r.Get("/peers", s.peersPage)

	// Mount rpc endpoints
	bundlerRPCHandler := proto.NewBundlerServer(s)
	r.Post("/rpc/Bundler/*", bundlerRPCHandler.ServeHTTP)

	// TODO: take config flag with debug_mode true/false
	debugRPCHandler := proto.NewDebugServer(&Debug{RPC: s})
	r.Post("/rpc/Debug/*", debugRPCHandler.ServeHTTP)

	return r
}

// Ping is a healthcheck that returns an empty message.
func (s *RPC) Ping(ctx context.Context) (bool, error) {
	return true, nil
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("."))
}

func (s *RPC) SendOperation(ctx context.Context, pop *proto.Operation) (bool, error) {
	op, err := types.NewOperationFromProto(pop)
	if err != nil {
		return false, err
	}

	// Always PIN these operations to IPFS
	// as they are being sent by the user, and
	// it is useful for debugging
	go s.mempool.ReportToIPFS(op)

	err = s.mempool.AddOperation(ctx, op, true)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (s RPC) Mempool(ctx context.Context) (*proto.MempoolView, error) {
	return s.mempool.Inspect(ctx), nil
}

func (s RPC) Operations(ctx context.Context) (*proto.Operations, error) {
	return s.archive.Operations(ctx), nil
}

func stubHandler(respBody string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(respBody))
	})
}
