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

type CodeReplacement struct {
	Address common.Address
	Code    []byte
}

type SlotReplacement struct {
	Address common.Address
	Slot    common.Hash
	Value   common.Hash
}

type DebugContextArgs struct {
	CodeReplacements []CodeReplacement
	SlotReplacements []SlotReplacement
}

type Interface interface {
	DebugTraceCall(ctx context.Context, args *DebugCallArgs) (*TransactionTrace, error)
	DebugTraceCallContext(ctx context.Context, args *DebugCallArgs, contextArgs *DebugContextArgs) (*TransactionTrace, error)
}
