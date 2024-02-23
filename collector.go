package bundler

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/0xsequence/bundler/config"
	"github.com/0xsequence/bundler/pricefeed"
	"github.com/0xsequence/ethkit/ethrpc"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/go-chi/httplog/v2"
)

type Collector struct {
	listening      bool
	lastBaseFee    *big.Int
	minPriorityFee *big.Int

	feeds map[common.Address]*pricefeed.Feed

	logger *httplog.Logger

	Provider *ethrpc.Provider
}

func NewCollector(cfg *config.CollectorConfig, logger *httplog.Logger, provider *ethrpc.Provider) (*Collector, error) {
	feeds := make(map[common.Address]*pricefeed.Feed)

	for _, ref := range cfg.References {
		feed, err := pricefeed.FeedForReference(&ref, logger, provider)
		if err != nil {
			return nil, err
		}

		addr := common.HexToAddress(ref.Token)
		if addr == (common.Address{}) {
			return nil, fmt.Errorf("collector: invalid token address: %s", ref.Token)
		}

		if _, ok := feeds[addr]; ok {
			return nil, fmt.Errorf("collector: duplicate token address: %s", ref.Token)
		}

		logger.Info("collector: added feed", "token", ref.Token, "feed", feed.Name())
		feeds[common.HexToAddress(ref.Token)] = &feed
	}

	return &Collector{
		feeds:    feeds,
		logger:   logger,
		Provider: provider,
	}, nil
}

func (c *Collector) BaseFee() *big.Int {
	return c.lastBaseFee
}

func (c *Collector) MinPriorityFee() *big.Int {
	return c.minPriorityFee
}

func (c *Collector) Run(ctx context.Context) error {
	if c.listening {
		return fmt.Errorf("collector: already running")
	}

	c.listening = true
	for ctx.Err() == nil {
		c.FetchBaseFee(ctx)
		c.FetchMinPriorityFee(ctx)

		time.Sleep(5 * time.Second)
	}

	return nil
}

func (c *Collector) Feeds() []*pricefeed.Feed {
	feeds := make([]*pricefeed.Feed, 0, len(c.feeds))
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

func (c *Collector) FetchMinPriorityFee(ctx context.Context) {
	block, err := c.Provider.BlockByNumber(ctx, nil)
	if err != nil {
		c.logger.Warn("collector: error fetching block", "error", err)
		return
	}

	// Load all transactions and find the lowest one that has priority fee != 0
	txs := block.Transactions()
	var lowest *big.Int
	for _, tx := range txs {
		if tx.GasTipCap().Cmp(big.NewInt(0)) != 0 {
			if lowest == nil || tx.GasTipCap().Cmp(lowest) < 0 {
				lowest = tx.GasTipCap()
			}
		}
	}

	if lowest == nil {
		c.logger.Debug("collector: no transactions with priority fee found")
		return
	}

	c.minPriorityFee = lowest
	c.logger.Debug("collector: lowest priority fee found", "fee", lowest.String())
}
