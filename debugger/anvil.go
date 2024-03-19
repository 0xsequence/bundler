package debugger

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/0xsequence/ethkit/go-ethereum/rpc"
	"github.com/go-chi/httplog/v2"
	"github.com/prometheus/client_golang/prometheus"
)

type anvilMetrics struct {
	startAttempts  prometheus.Counter
	startSuccesses prometheus.Counter
	startFailures  prometheus.Counter

	stopOperations  prometheus.Counter
	resetOperations prometheus.Counter

	debugTraceCallOperations prometheus.Counter
	debugTraceCallRetry      prometheus.Counter
	debugTraceCallSuccesses  prometheus.Counter
	debugTraceCallFailures   prometheus.Counter

	anvilRunning prometheus.Gauge

	ipcWaitFailures       *prometheus.CounterVec
	ipcWaitFailureError   prometheus.Labels
	ipcWaitFailureTimeout prometheus.Labels

	debugCallDuration   prometheus.Histogram
	ipcFileWaitDuration prometheus.Histogram
}

func createAnvilMetrics(reg prometheus.Registerer) *anvilMetrics {
	startAttempts := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "anvil_start_attempts",
		Help: "Number of attempts to start anvil",
	})

	startSuccesses := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "anvil_start_successes",
		Help: "Number of successful starts of anvil",
	})

	startFailures := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "anvil_start_failures",
		Help: "Number of failed starts of anvil",
	})

	stopOperations := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "anvil_stop_operations",
		Help: "Number of stop operations on anvil",
	})

	resetOperations := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "anvil_reset_operations",
		Help: "Number of reset operations on anvil",
	})

	debugTraceCallOperations := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "anvil_debug_trace_call_operations",
		Help: "Number of debug trace call operations on anvil",
	})

	debugTraceCallSuccesses := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "anvil_debug_trace_call_successes",
		Help: "Number of successful debug trace call operations on anvil",
	})

	debugTraceCallRetry := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "anvil_debug_trace_call_retry",
		Help: "Number of retries for debug trace call operations on anvil",
	})

	debugTraceCallFailures := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "anvil_debug_trace_call_failures",
		Help: "Number of failed debug trace call operations on anvil",
	})

	anvilRunning := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "anvil_running",
		Help: "Anvil running state",
	})

	debugCallDuration := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "anvil_debug_call_duration",
		Help: "Duration of anvil debug call operations",
		Buckets: []float64{
			0.25, // 0.25 seconds
			0.5,  // 0.5 seconds
			1,    // 1 second
			2.5,  // 2.5 seconds
			5,    // 5 seconds
			10,   // 10 seconds
			15,   // 15 seconds
			30,   // 30 seconds
			45,   // 45 seconds
			60,   // 60 seconds
		},
	})

	ipcWaitFailures := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "anvil_ipc_wait_failures",
		Help: "Number of failures waiting for anvil ipc file",
	}, []string{"reason"})

	ipcFileWaitDuration := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "anvil_ipc_file_wait_duration",
		Help:    "Duration of waiting for anvil ipc file",
		Buckets: prometheus.DefBuckets,
	})

	if reg != nil {
		reg.MustRegister(
			startAttempts, startSuccesses, startFailures,
			stopOperations, resetOperations,
			debugTraceCallOperations, debugTraceCallSuccesses,
			debugTraceCallRetry,
			ipcWaitFailures, debugTraceCallFailures,
			anvilRunning, debugCallDuration, ipcFileWaitDuration,
		)
	}

	return &anvilMetrics{
		startAttempts:            startAttempts,
		startSuccesses:           startSuccesses,
		startFailures:            startFailures,
		stopOperations:           stopOperations,
		resetOperations:          resetOperations,
		debugTraceCallOperations: debugTraceCallOperations,
		debugTraceCallSuccesses:  debugTraceCallSuccesses,
		debugTraceCallRetry:      debugTraceCallRetry,
		debugTraceCallFailures:   debugTraceCallFailures,
		anvilRunning:             anvilRunning,
		debugCallDuration:        debugCallDuration,
		ipcWaitFailures:          ipcWaitFailures,
		ipcFileWaitDuration:      ipcFileWaitDuration,

		ipcWaitFailureError:   prometheus.Labels{"reason": "error"},
		ipcWaitFailureTimeout: prometheus.Labels{"reason": "timeout"},
	}
}

type AnvilDebugger struct {
	ID     string
	RpcUrl string

	lock    sync.Mutex
	logger  *slog.Logger
	metrics *anvilMetrics
	ctx     context.Context

	client     *rpc.Client
	cancel     context.CancelFunc
	ipcAddr    string
	needsReset bool
}

var _ Interface = &AnvilDebugger{}

func NewAnvilDebugger(ctx context.Context, logger *httplog.Logger, metrics prometheus.Registerer, rpcUrl string) (*AnvilDebugger, error) {
	if err := checkExists(); err != nil {
		return nil, err
	}

	// Generate random hex id
	id := make([]byte, 8)
	if _, err := rand.Read(id); err != nil {
		return nil, err
	}
	idstr := common.Bytes2Hex(id)

	// Suffix with a random string to avoid conflicts
	// TODO: Allow config to set the ipc path
	ipcAddr := "/tmp/anvil-" + idstr + ".ipc"

	logger2 := logger.With("anvil_id", idstr)

	return &AnvilDebugger{
		ID:      idstr,
		ctx:     ctx,
		ipcAddr: ipcAddr,
		logger:  logger2,
		metrics: createAnvilMetrics(metrics),

		RpcUrl: rpcUrl,
	}, nil
}

func checkExists() error {
	cmd := exec.Command("anvil", "--version")

	// Getting the output of the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("anvil not found: %s", err)
	}

	// The output should include the "anvil" text
	if !strings.Contains(string(output), "anvil") {
		return fmt.Errorf("anvil command unexpected output: %s", output)
	}

	return nil
}

func (a *AnvilDebugger) Lock() *sync.Mutex {
	return &a.lock
}

func (a *AnvilDebugger) waitForIPC(timeout time.Duration) bool {
	start := time.Now()
	for {
		if _, err := os.Stat(a.ipcAddr); err == nil {
			a.metrics.ipcFileWaitDuration.Observe(float64(time.Since(start)))
			return true
		} else if os.IsNotExist(err) {
			if time.Since(start) > timeout {
				a.metrics.ipcWaitFailures.With(a.metrics.ipcWaitFailureTimeout).Inc()
				return false
			}
			time.Sleep(100 * time.Millisecond)
		} else {
			a.metrics.ipcWaitFailures.With(a.metrics.ipcWaitFailureError).Inc()
			return false
		}
	}
}

func (a *AnvilDebugger) Start() error {
	a.lock.Lock()
	defer a.lock.Unlock()
	return a.startLocked()
}

func (a *AnvilDebugger) startLocked() error {
	a.metrics.startAttempts.Inc()

	if a.ctx == nil || a.RpcUrl == "" || a.ipcAddr == "" {
		a.metrics.startFailures.Inc()
		return fmt.Errorf("anvil debugger not initialized")
	}

	if a.cancel != nil {
		a.metrics.startFailures.Inc()
		return fmt.Errorf("anvil already started")
	}

	cmd := exec.Command("anvil", "--fork-url", a.RpcUrl, "--ipc", a.ipcAddr, "--port", "0")
	if err := cmd.Start(); err != nil {
		a.metrics.startFailures.Inc()
		return err
	}

	endProc := func() {
		if err := cmd.Process.Kill(); err != nil {
			a.logger.Warn("Error killing anvil process", "err", err)
		}

		// Waiting for the process to exit
		if _, err := cmd.Process.Wait(); err != nil {
			a.logger.Warn("Error waiting for anvil process to exit", "err", err)
		}
	}

	if !a.waitForIPC(5 * time.Second) {
		endProc()
		a.metrics.startFailures.Inc()
		return fmt.Errorf("anvil timeout waiting for ipc file")
	}

	// Create new client
	rc, err := rpc.Dial(a.ipcAddr)
	if err != nil {
		endProc()
		a.metrics.startFailures.Inc()
		return err
	}

	ctx2, cancel := context.WithCancel(a.ctx)
	a.cancel = cancel

	// Listen for ctx.Done() and kill the process
	go func() {
		<-ctx2.Done()
		a.logger.Info("anvil stopping...", "ipc", a.ipcAddr)

		endProc()
		a.Stop()
	}()

	a.client = rc
	a.logger.Info("anvil started", "ipc", a.ipcAddr)
	a.metrics.startSuccesses.Inc()

	return nil
}

func (a *AnvilDebugger) Running() bool {
	a.lock.Lock()
	defer a.lock.Unlock()
	return a.cancel != nil
}

func (a *AnvilDebugger) Stop() error {
	a.lock.Lock()
	defer a.lock.Unlock()
	return a.stopLocked()
}

func (a *AnvilDebugger) stopLocked() error {
	a.metrics.stopOperations.Inc()

	if a.cancel == nil {
		return fmt.Errorf("anvil not started")
	}

	a.cancel()
	a.cancel = nil

	// Try to clean up the ipc file
	exec.Command("rm", a.ipcAddr).Run()
	if a.client != nil {
		a.client.Close()
	}

	return nil
}

func (a *AnvilDebugger) Reset() error {
	a.lock.Lock()
	defer a.lock.Unlock()
	return a.resetLocked()
}

func (a *AnvilDebugger) resetLocked() error {
	a.metrics.resetOperations.Inc()

	if a.cancel == nil {
		return fmt.Errorf("anvil not started")
	}

	if a.needsReset {
		start := time.Now()
		params := map[string]string{"jsonRpcUrl": a.RpcUrl}
		err := a.client.Call(nil, "anvil_reset", params)
		if err != nil {
			return err
		}

		a.logger.Debug("anvil reset", "ipc", a.ipcAddr, "duration", time.Since(start))
		a.needsReset = false
	}

	return nil
}

func (a *AnvilDebugger) tryDebugTraceCall(ctx context.Context, args *DebugCallArgs) (*TransactionTrace, error) {
	if err := a.resetLocked(); err != nil {
		return nil, err
	}

	res := &TransactionTrace{}
	params := map[string]string{
		"from": args.From.Hex(),
		"to":   args.To.Hex(),
		"data": "0x" + common.Bytes2Hex(args.Data),
	}

	ctx2, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	start := time.Now()
	err := a.client.CallContext(ctx2, res, "debug_traceCall", params)

	a.needsReset = true
	go a.Reset()

	if err != nil {
		return nil, err
	}

	a.logger.Debug("anvil debug trace call", "duration", time.Since(start))

	res.From = args.From

	return res, nil
}

func (a *AnvilDebugger) tryStartUnlocked() error {
	// Try 3 times, delay 3 seconds between each try
	errs := make([]error, 0, 3)
	for i := 0; i < 3; i++ {
		err := a.startLocked()
		if err == nil {
			return nil
		}

		errs = append(errs, err)
		a.logger.Warn("anvil failed to start, retrying...", "i", i, "err", err)

		time.Sleep(3 * time.Second)
	}

	return fmt.Errorf("failed to start anvil: %v", errs)
}

func (a *AnvilDebugger) DebugTraceCall(ctx context.Context, args *DebugCallArgs) (*TransactionTrace, error) {
	a.metrics.debugTraceCallOperations.Inc()
	start := time.Now()

	a.lock.Lock()
	defer a.lock.Unlock()

	if a.cancel == nil {
		err := a.tryStartUnlocked()
		if err != nil {
			a.metrics.debugTraceCallFailures.Inc()
			return nil, err
		}
	}

	// Try 3 times, delay 3 seconds between each try
	errs := make([]error, 0, 3)

	for i := 0; i < 3; i++ {
		res, err := a.tryDebugTraceCall(ctx, args)
		if err == nil {
			a.metrics.debugTraceCallSuccesses.Inc()
			a.metrics.debugCallDuration.Observe(float64(time.Since(start)))
			return res, nil
		}

		errs = append(errs, err)
		a.metrics.debugTraceCallRetry.Inc()
		a.logger.Warn("anvil failed to debug trace call, retrying...", "i", i, "err", err)
		time.Sleep(3 * time.Second)
	}

	a.metrics.debugTraceCallFailures.Inc()
	return nil, fmt.Errorf("failed to debug trace call: %v", errs)
}
