package pricefeed

import (
	"context"
	"fmt"

	"github.com/0xsequence/bundler/config"
	"github.com/0xsequence/ethkit/ethrpc"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/go-chi/httplog/v2"
	"github.com/prometheus/client_golang/prometheus"
)

type Feed interface {
	Ready() bool
	Name() string
	Snapshot() (*Snapshot, error)
	Start(ctx context.Context) error
}

func FeedForReference(cfg *config.PriceReference, logger *httplog.Logger, metrics prometheus.Registerer, provider ethrpc.Interface) (Feed, error) {
	if !common.IsHexAddress(cfg.Token) {
		return nil, fmt.Errorf("\"%v\" is not a token address", cfg.Token)
	}
	token := common.HexToAddress(cfg.Token)

	if token == (common.Address{}) {
		if cfg.UniswapV2 != nil {
			return nil, fmt.Errorf("no feed required for native token")
		}
		return NativeFeed{}, nil
	}

	if cfg.UniswapV2 != nil {
		return NewUniswapV2Feed(provider, logger, metrics, cfg.UniswapV2)
	}

	return nil, fmt.Errorf("pricefeed: unknown reference type")
}
