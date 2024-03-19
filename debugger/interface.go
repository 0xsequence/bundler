package debugger

import (
	"context"

	"github.com/0xsequence/ethkit/go-ethereum/common"
)

type DebugCallArgs struct {
	From common.Address
	To   common.Address
	Data []byte
}

type DebugOverride struct {
	Code      *string                     `json:"code"`
	StateDiff map[common.Hash]common.Hash `json:"stateDiff"`
}

type DebugOverrideArgs map[common.Address]*DebugOverride

type LogEntry struct {
	PC      float64  `json:"pc"`
	Op      string   `json:"op"`
	Gas     float64  `json:"gas"`
	GasCost float64  `json:"gasCost"`
	Stack   []string `json:"stack"`
	Depth   float64  `json:"depth"`
}

type TransactionTrace struct {
	From        common.Address `json:"from"`
	Failed      bool           `json:"failed"`
	Gas         float64        `json:"gas"`
	ReturnValue string         `json:"returnValue"`
	StructLogs  []LogEntry     `json:"structLogs"`
}

type Interface interface {
	DebugTraceCall(ctx context.Context, args *DebugCallArgs, overrideArgs *DebugOverrideArgs) (*TransactionTrace, error)
}
