package collector

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/0xsequence/bundler/config"
	"github.com/0xsequence/bundler/pricefeed"
	"github.com/0xsequence/bundler/types"
	"github.com/0xsequence/ethkit/ethrpc"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/go-chi/httplog/v2"
)

type Collector struct {
	cfg *config.CollectorConfig

	listening   bool
	lastBaseFee *big.Int
	priorityFee *big.Int

	feeds map[common.Address]pricefeed.Feed

	logger *httplog.Logger

	Provider *ethrpc.Provider
}

var _ Interface = &Collector{}

func NewCollector(cfg *config.CollectorConfig, logger *httplog.Logger, provider *ethrpc.Provider) (*Collector, error) {
	feeds := make(map[common.Address]pricefeed.Feed)

	priorityFee := new(big.Int).SetInt64(cfg.PriorityFee)

	c := &Collector{
		cfg:         cfg,
		feeds:       feeds,
		logger:      logger,
		priorityFee: priorityFee,
		Provider:    provider,
	}

	for _, ref := range cfg.References {
		feed, err := pricefeed.FeedForReference(&ref, logger, provider)
		if err != nil {
			return nil, err
		}

		if err := c.AddFeed(ref.Token, feed); err != nil {
			return nil, err
		}
	}

	return c, nil
}

func (c *Collector) AddFeed(tokenAddr string, feed pricefeed.Feed) error {
	addr := common.HexToAddress(tokenAddr)
	if addr == (common.Address{}) {
		return fmt.Errorf("collector: invalid token address: %s", tokenAddr)
	}

	if _, ok := c.feeds[addr]; ok {
		return fmt.Errorf("collector: duplicate token address: %s", tokenAddr)
	}

	c.logger.Info("collector: added feed", "token", tokenAddr, "feed", feed.Name())
	c.feeds[common.HexToAddress(tokenAddr)] = feed

	return nil
}

func (c *Collector) BaseFee() *big.Int {
	return c.lastBaseFee
}

func (c *Collector) PriorityFee() *big.Int {
	return c.priorityFee
}

func (c *Collector) Run(ctx context.Context) error {
	if c.listening {
		return fmt.Errorf("collector: already running")
	}

	c.listening = true
	for ctx.Err() == nil {
		c.FetchBaseFee(ctx)

		time.Sleep(5 * time.Second)
	}

	return nil
}

func (c *Collector) Feeds() []pricefeed.Feed {
	feeds := make([]pricefeed.Feed, 0, len(c.feeds))
	for _, feed := range c.feeds {
		feeds = append(feeds, feed)
	}
	return feeds
}

func (c *Collector) FetchBaseFee(ctx context.Context) {
	block, err := c.Provider.BlockByNumber(ctx, nil)
	if err != nil {
		c.logger.Warn("collector: error fetching block", "error", err)
		return
	}

	c.lastBaseFee = block.BaseFee()
	c.logger.Debug("collector: base fee fetched", "fee", c.lastBaseFee.String())
}

func (c *Collector) MinFeePerGas(feeToken common.Address) (*big.Int, error) {
	if c.lastBaseFee == nil {
		return nil, fmt.Errorf("collector: base fee not fetched")
	}

	minFeePerGas := new(big.Int).Add(c.lastBaseFee, c.priorityFee)

	if feeToken != (common.Address{}) {
		feed, ok := c.feeds[feeToken]
		if !ok {
			return nil, fmt.Errorf("collector: unsupported fee token: %s", feeToken.Hex())
		}

		var err error
		minFeePerGas, err = feed.FromNative(minFeePerGas)
		if err != nil {
			return nil, fmt.Errorf("collector: error converting fee to native token: %w", err)
		}
	}

	if c.cfg.PriorityFeeMul != 0 {
		scalar := new(big.Float).SetFloat64(c.cfg.PriorityFeeMul)
		applied := new(big.Float).Mul(new(big.Float).SetInt(minFeePerGas), scalar)
		minFeePerGas, _ = applied.Int(nil)
	}

	return minFeePerGas, nil
}

func (c *Collector) MeetsPayment(op *types.Operation) (bool, error) {
	minFeePerGas, err := c.MinFeePerGas(op.FeeToken)
	if err != nil {
		return false, err
	}

	if op.MaxFeePerGas.Cmp(minFeePerGas) < 0 {
		return false, fmt.Errorf("collector: operation payment too low: %s < %s", op.MaxFeePerGas.String(), minFeePerGas.String())
	}

	return true, nil
}
