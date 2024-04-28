package node

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/0xsequence/bundler"
	"github.com/0xsequence/bundler/config"
	"github.com/0xsequence/bundler/endorser"
	"github.com/0xsequence/bundler/ipfs"
	"github.com/0xsequence/bundler/lib/calldata"
	"github.com/0xsequence/bundler/lib/collector"
	"github.com/0xsequence/bundler/lib/debugger"
	"github.com/0xsequence/bundler/lib/provider"
	"github.com/0xsequence/bundler/lib/registry"
	"github.com/0xsequence/bundler/lib/store"
	"github.com/0xsequence/bundler/lib/utils"
	"github.com/0xsequence/bundler/mempool"
	"github.com/0xsequence/bundler/p2p"
	"github.com/0xsequence/bundler/rpc"
	"github.com/0xsequence/ethkit/ethrpc"
	"github.com/0xsequence/ethkit/ethwallet"
	"github.com/go-chi/httplog/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"golang.org/x/sync/errgroup"
)

type Node struct {
	Config *config.Config
	Logger *httplog.Logger
	Host   *p2p.Host
	RPC    *rpc.RPC

	Mempool   mempool.Interface
	Archive   *bundler.Archive
	Ingress   *bundler.Ingress
	Collector *collector.Collector
	Registry  registry.Interface
	Pruner    *bundler.Pruner
	Provider  *provider.Batched

	ctx       context.Context
	ctxStopFn context.CancelFunc
	running   int32
}

func NewNode(cfg *config.Config) (*Node, error) {
	var err error

	cfg.GitCommit = bundler.GITCOMMIT

	// Logging
	loggerOptions := httplog.Options{
		LogLevel:         httplog.LevelByName(cfg.Logging.Level),
		JSON:             cfg.Logging.JSON,
		Concise:          cfg.Logging.Concise,
		RequestHeaders:   cfg.Logging.RequestHeaders,
		ResponseHeaders:  cfg.Logging.ResponseHeaders,
		MessageFieldName: "message",
		LevelFieldName:   "severity",
		TimeFieldFormat:  time.RFC3339Nano,
		Tags: map[string]string{
			"serviceName":    cfg.Logging.ServiceName,
			"serviceVersion": bundler.GITCOMMIT,
		},
		QuietDownRoutes: []string{
			"/",
			"/ping",
		},
		QuietDownPeriod: 10 * time.Second,
		SourceFieldName: cfg.Logging.Source,
	}
	if !cfg.Logging.JSON {
		loggerOptions.TimeFieldFormat = time.RFC3339
	}
	logger := httplog.NewLogger("bundler", loggerOptions)

	// Wallet
	mnmonic := cfg.Mnemonic
	if mnmonic == "" {
		// TODO: Maybe persist the wallet in a file?
		entropy, err := ethwallet.RandomEntropy(256)
		if err != nil {
			return nil, err
		}

		mnmonic, err = ethwallet.EntropyToMnemonic(entropy)
		if err != nil {
			return nil, err
		}

		logger.Info("=> no mnemonic provided, using temporal wallet")
	}

	// Identity
	identity, err := p2p.NewIdentity(mnmonic)
	if err != nil {
		return nil, err
	}

	// Provider
	base, err := ethrpc.NewProvider(cfg.NetworkConfig.RpcUrl)
	if err != nil {
		return nil, err
	}
	client := utils.NewHttpRpcMetricsClient()
	base.SetHTTPClient(&http.Client{
		Transport: client,
	})

	// Extended provider
	extended := provider.NewExtended(base, true, true)
	batched := provider.NewBatched(extended, 10*time.Millisecond)

	// ChainID
	chainID, err := batched.ChainID(context.Background())
	if err != nil {
		return nil, err
	}

	logger.Info("=> setup node identity", "id", identity.ID.String())

	// Metrics
	prom := prometheus.NewRegistry()
	promPrefix := prometheus.WrapRegistererWithPrefix("bundler_", prom)
	promPrefix = prometheus.WrapRegistererWith(prometheus.Labels{"id": identity.ID.String()}, promPrefix)
	promPrefix = prometheus.WrapRegistererWith(prometheus.Labels{"chain_id": chainID.String()}, promPrefix)

	promPrefix.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	client.UseRegistry(promPrefix, cfg.NetworkConfig.RpcUrl)

	// Debugger
	debugger, err := debugger.NewDebugger(cfg.DebuggerConfig, context.Background(), logger, promPrefix, cfg.NetworkConfig.RpcUrl)
	if err != nil {
		return nil, err
	}

	// Endorser
	endorser := endorser.NewEndorser(logger, promPrefix, batched, debugger)

	// Store
	// TODO: Add custom store path
	store, err := store.CreateInstanceStore(identity.ID.String() + "-" + chainID.String())
	if err != nil {
		logger.Warn("=> unable to create instance store", "error", err)
	} else {
		logger.Info("=> setup instance store", "path", store.String())
	}

	// p2p host
	host, err := p2p.NewHost(&cfg.P2PHostConfig, logger.Logger, promPrefix, identity, chainID)
	if err != nil {
		return nil, err
	}

	// IPFS Client
	ipfs := ipfs.NewClient(promPrefix, cfg.NetworkConfig.IPFSUrl)

	// Collector
	collector, err := collector.NewCollector(&cfg.CollectorConfig, logger, promPrefix, batched)
	if err != nil {
		return nil, err
	}

	// Gas model
	var calldataModel calldata.CostModel
	if cfg.LinearCalldataModel != nil {
		calldataModel = calldata.NewLinearModel(
			cfg.LinearCalldataModel.FixedCost,
			cfg.LinearCalldataModel.ZeroByteCost,
			cfg.LinearCalldataModel.NonZeroByteCost,
		)
	} else {
		logger.Info("=> using default calldata model")
		calldataModel = calldata.DefaultModel()
	}

	// Endorser registry
	registry, err := registry.NewRegistry(&cfg.RegistryConfig, logger, promPrefix, batched)
	if err != nil {
		return nil, err
	}

	// Mempool
	mempool, err := mempool.NewMempool(&cfg.MempoolConfig, logger, promPrefix, endorser, collector, ipfs, calldataModel, registry)
	if err != nil {
		return nil, err
	}

	// Ingress
	ingress := bundler.NewIngress(&cfg.MempoolConfig, logger, promPrefix, mempool, collector, host)

	// Archive
	archive := bundler.NewArchive(&cfg.ArchiveConfig, host, logger, promPrefix, store, ipfs, mempool)

	// Pruner
	pruner := bundler.NewPruner(cfg.PrunerConfig, logger, promPrefix, mempool, endorser, registry)

	// RPC
	rpc, err := rpc.NewRPC(cfg, logger, promPrefix, prom, host, mempool, archive, batched.Provider, collector, endorser, ipfs, registry)
	if err != nil {
		return nil, err
	}

	//
	// Server
	//
	server := &Node{
		Config:    cfg,
		Logger:    logger,
		Host:      host,
		RPC:       rpc,
		Mempool:   mempool,
		Archive:   archive,
		Ingress:   ingress,
		Collector: collector,
		Registry:  registry,
		Pruner:    pruner,
		Provider:  batched,
	}

	return server, nil
}

func (s *Node) Run() error {
	if s.IsRunning() {
		return fmt.Errorf("server already running")
	}

	oplog := s.Logger.With("op", "run")
	oplog.Info("=> run service")

	// Running
	atomic.StoreInt32(&s.running, 1)

	// Server root context
	s.ctx, s.ctxStopFn = context.WithCancel(context.Background())

	// Subprocess run context
	g, ctx := errgroup.WithContext(s.ctx)

	// RPC
	g.Go(func() error {
		oplog.Info("-> rpc: run")
		return s.RPC.Run(ctx)
	})

	// Provider
	g.Go(func() error {
		oplog.Info("-> provider: run")
		return s.Provider.Run(ctx)
	})

	// Node
	g.Go(func() error {
		oplog.Info("-> p2p: run")
		return s.Host.Run(ctx)
	})

	// Ingress processor
	g.Go(func() error {
		oplog.Info("-> ingress: run")
		s.Ingress.Run(ctx)
		return nil
	})

	// Archive
	g.Go(func() error {
		oplog.Info("-> archive: run")
		s.Archive.Run(ctx)
		return nil
	})

	// Collector
	g.Go(func() error {
		oplog.Info("-> collector: run")
		s.Collector.Run(ctx)
		return nil
	})

	// Collector feeds
	feeds := s.Collector.Feeds()
	for _, feed := range feeds {
		feed := feed
		g.Go(func() error {
			oplog.Info("-> collector: feed: run", "feed", feed.Name())
			err := feed.Start(ctx)
			if err != nil {
				oplog.Error("-> collector: feed: error", "feed", feed.Name(), "error", err)
			}
			return err
		})
	}

	// Pruner
	g.Go(func() error {
		oplog.Info("-> pruner: run")
		s.Pruner.Run(ctx)
		return nil
	})

	// Once run context is done, trigger a server-stop.
	go func() {
		<-ctx.Done()
		s.Stop()
	}()

	// Wait for subprocesses to finish
	return g.Wait()
}

func (s *Node) Stop() {
	if !s.IsRunning() || s.IsStopping() {
		return
	}

	s.Logger.Info("-> bundler: shutdown server")

	// Stopping
	atomic.StoreInt32(&s.running, 2)

	// Shutdown signal with grace period of 30 seconds
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Force shutdown after grace period
	go func() {
		<-shutdownCtx.Done()
		if shutdownCtx.Err() == context.DeadlineExceeded {
			s.Fatal("graceful shutdown timed out.. forced exit.")
		}
	}()

	var wg sync.WaitGroup

	// TODO: stop all various internal services

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.RPC.Stop(shutdownCtx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.Archive.Stop(shutdownCtx)
	}()

	wg.Wait()

	// Stop the P2P layer last
	// as the node may have some messages to send

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.Host.Stop(shutdownCtx)
	}()

	// Wait for subprocesses to gracefully stop
	wg.Wait()
	s.ctxStopFn()
	atomic.StoreInt32(&s.running, 0)
}

func (s *Node) IsRunning() bool {
	return atomic.LoadInt32(&s.running) >= 1
}

func (s *Node) IsStopping() bool {
	return atomic.LoadInt32(&s.running) == 2
}

func (s *Node) Fatal(format string, v ...interface{}) {
	s.Logger.Error(fmt.Sprintf(format, v...))
	os.Exit(1)
}

func (s *Node) End() {
	s.Logger.Info("-> bundler: bye")
}
