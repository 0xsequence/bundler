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
	"github.com/0xsequence/ethkit/go-ethereum/common/hexutil"
	"github.com/0xsequence/ethkit/go-ethereum/rpc"
	"github.com/go-chi/httplog/v2"
	"github.com/prometheus/client_golang/prometheus"
)

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

func (a *AnvilDebugger) CodeAt(ctx context.Context, addr common.Address) ([]byte, error) {
	var result hexutil.Bytes
	err := a.client.Call(&result, "eth_getCode", addr.Hex(), "latest")
	return result, err
}

func (a *AnvilDebugger) tryDebugTraceCall(ctx context.Context, args *DebugCallArgs, contextArgs *DebugContextArgs) (*TransactionTrace, error) {
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

	start_context := time.Now()

	if (contextArgs != nil) && (len(contextArgs.CodeReplacements) > 0) {
		for _, cr := range contextArgs.CodeReplacements {
			a.performSetCode(ctx, cr.Address, cr.Code)
		}
	}
	if (contextArgs != nil) && (len(contextArgs.SlotReplacements) > 0) {
		for _, sr := range contextArgs.SlotReplacements {
			a.performSetStorageAt(ctx, sr.Address, sr.Slot, sr.Value)
		}
	}

	start := time.Now()
	err := a.client.CallContext(ctx2, res, "debug_traceCall", params)

	a.needsReset = true
	go a.Reset()

	if err != nil {
		return nil, err
	}

	a.logger.Debug("anvil debug trace call", "duration", time.Since(start))
	a.logger.Debug("anvil debug trace call with context", "duration", time.Since(start_context))

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
	return a.DebugTraceCallContext(ctx, args, nil)
}

func (a *AnvilDebugger) DebugTraceCallContext(ctx context.Context, args *DebugCallArgs, contextArgs *DebugContextArgs) (*TransactionTrace, error) {
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
		res, err := a.tryDebugTraceCall(ctx, args, contextArgs)
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

func (a *AnvilDebugger) performSetCode(ctx context.Context, addr common.Address, code []byte) error {
	hexCode := "0x" + common.Bytes2Hex(code)

	ctx2, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	start := time.Now()
	err := a.client.CallContext(ctx2, nil, "anvil_setCode", addr.Hex(), hexCode)

	if err != nil {
		return err
	}

	a.logger.Debug("anvil set code", "duration", time.Since(start))

	return nil
}

func (a *AnvilDebugger) SetCode(ctx context.Context, addr common.Address, code []byte) error {
	a.lock.Lock()
	defer a.lock.Unlock()

	if a.cancel == nil {
		err := a.tryStartUnlocked()
		if err != nil {
			return err
		}
	}

	return a.performSetCode(ctx, addr, code)
}

func (a *AnvilDebugger) performSetStorageAt(ctx context.Context, addr common.Address, slot common.Hash, value common.Hash) error {
	ctx2, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	start := time.Now()
	err := a.client.CallContext(ctx2, nil, "anvil_setStorageAt", addr.Hex(), slot.Hex(), value.Hex())

	if err != nil {
		return err
	}

	a.logger.Debug("anvil set storage", "duration", time.Since(start))

	return nil
}

func (a *AnvilDebugger) SetStorageAt(ctx context.Context, addr common.Address, slot common.Hash, value common.Hash) error {
	a.lock.Lock()
	defer a.lock.Unlock()

	if a.cancel == nil {
		err := a.tryStartUnlocked()
		if err != nil {
			return err
		}
	}

	return a.performSetStorageAt(ctx, addr, slot, value)
}
