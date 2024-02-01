package server

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/0xsequence/bundler"
	"github.com/0xsequence/bundler/config"
	"github.com/0xsequence/bundler/p2p"
	"github.com/0xsequence/bundler/rpc"
	"github.com/go-chi/httplog/v2"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	Config *config.Config
	Logger *httplog.Logger
	// Node   *p2p.Node
	RPC *rpc.RPC

	ctx       context.Context
	ctxStopFn context.CancelFunc
	running   int32
}

func NewServer(cfg *config.Config) (*Server, error) {
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

	// ctx := context.Background()
	// _ = ctx

	// node, err := p2p.NewNode(ctx, cfg, db, logger)
	// if err != nil {
	// 	return nil, err
	// }
	node := &p2p.Node{}

	// priv, pub := node.KeyPair()

	// clog, err := p2p.NewLog(cfg, db, node.PeerID(), node, priv, pub, logger)
	// if err != nil {
	// 	return nil, err
	// }

	// RPC
	rpc, err := rpc.NewRPC(cfg, logger, node)
	if err != nil {
		return nil, err
	}

	//
	// Server
	//
	server := &Server{
		Config: cfg,
		Logger: logger,
		// Node:   node,
		RPC: rpc,
	}

	return server, nil
}

func (s *Server) P2PNodeAddr() string {
	// return s.Node.NodeAddr()
	return ""
}

func (s *Server) Run() error {
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
	// g.Go(func() error {
	// 	oplog.Info("-> p2p: run")
	// 	return s.Node.Run(ctx)
	// })

	// Once run context is done, trigger a server-stop.
	go func() {
		<-ctx.Done()
		s.Stop()
	}()

	// Wait for subprocesses to finish
	return g.Wait()
}

func (s *Server) Stop() {
	if !s.IsRunning() || s.IsStopping() {
		return
	}

	s.Logger.Info("-> bundler: shutdown server")

	// Stopping
	atomic.StoreInt32(&s.running, 2)

	// Shutdown signal with grace period of 30 seconds
	shutdownCtx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.RPC.Stop(shutdownCtx)
	}()

	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	s.Node.Stop()
	// }()

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

func (s *Server) IsRunning() bool {
	return atomic.LoadInt32(&s.running) >= 1
}

func (s *Server) IsStopping() bool {
	return atomic.LoadInt32(&s.running) == 2
}

func (s *Server) Fatal(format string, v ...interface{}) {
	s.Logger.Error(fmt.Sprintf(format, v...))
	os.Exit(1)
}

func (s *Server) End() {
	s.Logger.Info("-> bundler: bye")
}
