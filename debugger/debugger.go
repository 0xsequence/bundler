package debugger

import (
	"context"
	"fmt"
	"strings"

	"github.com/0xsequence/bundler/config"
	"github.com/go-chi/httplog/v2"
	"github.com/prometheus/client_golang/prometheus"
)

func NewDebugger(cfg config.DebuggerConfig, ctx context.Context, logger *httplog.Logger, metrics prometheus.Registerer, rpcUrl string) (Interface, error) {
	mls := strings.ToLower(cfg.Mode)
	switch mls {
	case "anvil":
		logger.Info("debugger: anvil debugger enabled")
		return NewAnvilDebugger(ctx, logger, metrics, rpcUrl)
	case "default", "none", "":
		logger.Info("debugger: no debugger enabled")
		return nil, nil
	default:
		return nil, fmt.Errorf("unknown debugger mode: %s", cfg.Mode)
	}
}
