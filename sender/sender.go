package sender

import (
	"context"
	"math/big"
	"sync"
	"time"

	"github.com/0xsequence/bundler/collector"
	"github.com/0xsequence/bundler/config"
	"github.com/0xsequence/bundler/endorser"
	"github.com/0xsequence/bundler/interfaces"
	"github.com/0xsequence/bundler/mempool"
	"github.com/0xsequence/bundler/registry"
	"github.com/0xsequence/bundler/sender/chiller"
	"github.com/0xsequence/bundler/sender/worker"
	"github.com/0xsequence/bundler/utils"
	"github.com/go-chi/httplog/v2"
	"github.com/prometheus/client_golang/prometheus"
)

type Sender struct {
	logger  *httplog.Logger
	metrics *metrics

	sleepWait time.Duration
	workers   []*worker.Worker
	chiller   *chiller.Chiller

	Collector collector.Interface
	Registry  registry.Interface
	Mempool   mempool.Interface
}

var _ Interface = &Sender{}

func NewSender(
	cfg *config.SendersConfig,
	logger *httplog.Logger,
	factory WalletFactory,
	provider interfaces.Provider,
	mempool mempool.Interface,
	endorser endorser.Interface,
	simulator interfaces.Validator2,
	collector collector.Interface,
	registry registry.Interface,
) *Sender {
	var chillWait time.Duration
	if cfg.ChillWait > 0 {
		chillWait = time.Duration(cfg.ChillWait) * time.Second
	} else {
		chillWait = 1 * time.Second
		logger.Warn("sender: chill wait not set, using default", "chillWait", chillWait)
	}

	var sleepWait time.Duration
	if cfg.SleepWait > 0 {
		sleepWait = time.Duration(cfg.SleepWait) * time.Millisecond
	} else {
		sleepWait = 10 * time.Millisecond
		logger.Warn("sender: sleep wait not set, using default", "sleepWait", sleepWait)
	}

	// Get the minimum balance
	minBalance := big.NewInt(0)
	minBalance, ok := minBalance.SetString(cfg.MinBalance, 10)
	if !ok {
		logger.Warn("sender: invalid min balance", "minBalance", cfg.MinBalance)
	}

	// If minBalance is zero, set it to 0.01 ETH
	if minBalance == nil {
		minBalance = big.NewInt(10000000000000000)
		logger.Warn("sender: min balance not set, using default", "minBalance", minBalance)
	}

	// Create workers
	workers := make([]*worker.Worker, 0, cfg.NumSenders)
	for i := 0; i < int(cfg.NumSenders); i++ {
		wallet, err := factory.GetWallet(i)
		if err != nil || wallet == nil {
			logger.Warn("sender: wallet not available", "id", i, "err", err)
			continue
		}

		worker := worker.NewWorker(provider, collector, endorser, simulator, wallet, big.NewInt(int64(cfg.PriorityFee)), minBalance)
		worker.SetLogger(logger.With("worker", i, "addr", wallet.Address().String()))
		workers = append(workers, worker)
	}

	return &Sender{
		logger:  logger,
		metrics: createMetrics(),

		sleepWait: sleepWait,
		chiller:   chiller.NewChiller(chillWait),
		workers:   workers,

		Collector: collector,
		Registry:  registry,
		Mempool:   mempool,
	}
}

func (s *Sender) SetRegisterer(reg prometheus.Registerer) {
	s.metrics.register(reg)

	for _, worker := range s.workers {
		worker.SetRegisterer(reg)
	}
}

func (s *Sender) Run(ctx context.Context) {
	if len(s.workers) == 0 {
		s.logger.Warn("sender: no workers available")
		return
	}

	input := make(chan *mempool.TrackedOperation)

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(input)
		s.pullWorker(ctx, input)
	}()

	// Start workers
	for _, w := range s.workers {
		wg.Add(1)
		go func(w *worker.Worker) {
			defer wg.Done()
			w.Run(ctx, input)
		}(w)
	}

	// Start handler
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.handlerWorker(ctx)
	}()

	wg.Wait()
}

func (s *Sender) pullWorker(ctx context.Context, input chan<- *mempool.TrackedOperation) {
	for ctx.Err() == nil {
		ops := s.pull(ctx)

		if len(ops) == 0 {
			// Wait 10ms before trying again
			// or until the context is done
			s.metrics.skipRunNoOps.Inc()

			select {
			case <-ctx.Done():
				return
			case <-time.After(s.sleepWait):
				continue
			}
		}

		for _, op := range ops {
			input <- op
		}
	}
}

func (s *Sender) pull(ctx context.Context) []*mempool.TrackedOperation {
	return s.Mempool.ReserveOps(ctx, func(to []*mempool.TrackedOperation) []*mempool.TrackedOperation {
		defer utils.RecordFunctionDuration(time.Now(), s.metrics.selectOpsTime)

		if len(to) == 0 {
			return nil
		}

		// Find the best operation not blocked or chilled
		// we lock the chiller once and do all the checks
		s.chiller.Lock()
		defer s.chiller.Unlock()

		var best *mempool.TrackedOperation
		for _, op := range to {
			if s.chiller.HasLocked(op.Hash()) {
				continue
			}

			if best == nil || s.Collector.Cmp(&op.Operation, &best.Operation) > 0 {
				best = op
			}
		}

		if best != nil {
			return []*mempool.TrackedOperation{best}
		}

		return nil
	})
}

func (s *Sender) handlerWorker(ctx context.Context) {
	// Create fan-in channels
	chills := make([]<-chan string, len(s.workers))
	dones := make([]<-chan string, len(s.workers))
	discards := make([]<-chan string, len(s.workers))
	releases := make([]<-chan *worker.ReleaseOp, len(s.workers))
	bans := make([]<-chan *worker.BanEndorser, len(s.workers))

	for i, w := range s.workers {
		chills[i] = w.Chill()
		dones[i] = w.Done()
		discards[i] = w.Discard()
		releases[i] = w.Release()
		bans[i] = w.Ban()
	}

	chill := utils.FanIn(chills...)
	done := utils.FanIn(dones...)
	discard := utils.FanIn(discards...)
	release := utils.FanIn(releases...)
	ban := utils.FanIn(bans...)

	for {
		select {
		case <-ctx.Done():
			return
		case oph := <-chill:
			s.chiller.Chill(oph)
		case oph := <-done:
			s.chiller.Freeze(oph)
		case oph := <-discard:
			s.Mempool.DiscardOps(ctx, []string{oph})
		case op := <-release:
			s.Mempool.ReleaseOps(ctx, []string{op.Oph}, op.Change)
		case ban := <-ban:
			s.Registry.BanEndorser(ban.Endorser, ban.Type)
		}
	}
}
