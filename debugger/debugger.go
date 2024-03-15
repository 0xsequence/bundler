package debugger

import (
	"context"
	"fmt"
	"strings"

	"github.com/0xsequence/bundler/config"
	"github.com/go-chi/httplog/v2"
)

func NewDebugger(cfg config.DebuggerConfig, ctx context.Context, logger *httplog.Logger, rpcUrl string) (Interface, error) {
	mls := strings.ToLower(cfg.Mode)
	switch mls {
	case "anvil":
		logger.Info("debugger: anvil debugger enabled")
		return NewAnvilDebugger(ctx, logger, rpcUrl)
	case "default", "none", "":
		logger.Info("debugger: no debugger enabled")
		return nil, nil
	default:
		return nil, fmt.Errorf("unknown debugger mode: %s", cfg.Mode)
	}
}
