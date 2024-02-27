package node

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/0xsequence/bundler"
	"github.com/0xsequence/bundler/config"
	"github.com/0xsequence/bundler/endorser"
	"github.com/0xsequence/bundler/ipfs"
	"github.com/0xsequence/bundler/p2p"
	"github.com/0xsequence/bundler/rpc"
	"github.com/0xsequence/ethkit/ethrpc"
	"github.com/go-chi/httplog/v2"
	"golang.org/x/sync/errgroup"
)

type Node struct {
	Config *config.Config
	Logger *httplog.Logger
	Host   *p2p.Host
	RPC    *rpc.RPC

	Mempool   *bundler.Mempool
	Archive   *bundler.Archive
	Ingress   *bundler.Ingress
	Collector *bundler.Collector

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

	// Provider
	provider, err := ethrpc.NewProvider(cfg.NetworkConfig.RpcUrl)
	if err != nil {
		return nil, err
	}

	// Endorser
	endorser := endorser.NewEndorser(provider)

	// wallet
	wallet, err := rpc.SetupWallet(cfg.Mnemonic, 0, provider)
	if err != nil {
		return nil, err
	}
	logger.Info("=> setup node wallet", "address", wallet.Address().String())

	// p2p host
	host, err := p2p.NewHost(cfg, logger.Logger, wallet)
	if err != nil {
		return nil, err
	}

	// IPFS Client
	ipfs := ipfs.NewClient(cfg.NetworkConfig.IpfsUrl)

	// Collector
	collector, err := bundler.NewCollector(&cfg.CollectorConfig, logger, provider)
	if err != nil {
		return nil, err
	}

	// Mempool
	mempool, err := bundler.NewMempool(&cfg.MempoolConfig, logger, endorser, host, collector, ipfs)
	if err != nil {
		return nil, err
	}

	// Ingress
	ingress := bundler.NewIngress(&cfg.MempoolConfig, logger, mempool, collector, host)

	// Archive
	archive := bundler.NewArchive(host, logger, ipfs, mempool)

	// RPC
	rpc, err := rpc.NewRPC(cfg, logger, host, mempool, archive, provider, collector, endorser)
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
		f := *feed
		g.Go(func() error {
			oplog.Info("-> collector: feed: run", "feed", f.Name())
			err := f.Start(ctx)
			if err != nil {
				oplog.Error("-> collector: feed: error", "feed", f.Name(), "error", err)
			}
			return err
		})
	}

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

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.RPC.Stop(shutdownCtx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.Host.Stop(shutdownCtx)
	}()

	// Force shutdown after grace period
	go func() {
		<-shutdownCtx.Done()
		if shutdownCtx.Err() == context.DeadlineExceeded {
			s.Fatal("graceful shutdown timed out.. forced exit.")
		}
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
