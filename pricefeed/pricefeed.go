package pricefeed

import (
	"context"
	"fmt"
	"math/big"

	"github.com/0xsequence/bundler/config"
	"github.com/0xsequence/ethkit/ethrpc"
	"github.com/go-chi/httplog/v2"
)

type Feed interface {
	Ready() bool
	Name() string
	FromNative(amount *big.Int) (*big.Int, error)
	ToNative(amount *big.Int) (*big.Int, error)
	Factors() (*big.Int, *big.Int, error)
	Start(ctx context.Context) error
}

func FeedForReference(cfg *config.PriceReference, logger *httplog.Logger, provider ethrpc.Interface) (Feed, error) {
	if cfg.UniswapV2 != nil {
		return NewUniswapV2Feed(provider, logger, cfg.UniswapV2)
	}

	return nil, fmt.Errorf("pricefeed: unknown reference type")
}
