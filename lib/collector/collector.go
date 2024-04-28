package collector

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/0xsequence/bundler/config"
	"github.com/0xsequence/bundler/lib/pricefeed"
	"github.com/0xsequence/bundler/lib/types"
	"github.com/0xsequence/bundler/proto"
	"github.com/0xsequence/ethkit/ethrpc"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/0xsequence/go-sequence/lib/prototyp"
	"github.com/go-chi/httplog/v2"
	"github.com/prometheus/client_golang/prometheus"
)

type Collector struct {
	cfg  *config.CollectorConfig
	lock sync.Mutex

	listening   bool
	lastBaseFee *big.Int
	priorityFee *big.Int

	feeds map[common.Address]pricefeed.Feed

	logger  *httplog.Logger
	metrics *metrics

	Provider ethrpc.Interface
}

var _ Interface = &Collector{}

func NewCollector(cfg *config.CollectorConfig, logger *httplog.Logger, metrics prometheus.Registerer, provider ethrpc.Interface) (*Collector, error) {
	feeds := make(map[common.Address]pricefeed.Feed)

	priorityFee := new(big.Int).SetInt64(cfg.PriorityFee)

	c := &Collector{
		cfg:         cfg,
		lock:        sync.Mutex{},
		feeds:       feeds,
		metrics:     createMetrics(metrics),
		logger:      logger,
		priorityFee: priorityFee,
		Provider:    provider,
	}

	for _, ref := range cfg.References {
		feed, err := pricefeed.FeedForReference(&ref, logger, metrics, provider)
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
	if !common.IsHexAddress(tokenAddr) {
		return fmt.Errorf("\"%v\" is not a token address", tokenAddr)
	}
	addr := common.HexToAddress(tokenAddr)

	if _, ok := c.feeds[addr]; ok {
		return fmt.Errorf("collector: duplicate token address: %s", tokenAddr)
	}

	c.logger.Info("collector: added feed", "token", tokenAddr, "feed", feed.Name())
	c.feeds[common.HexToAddress(tokenAddr)] = feed

	return nil
}

func (c *Collector) Feed(tokenStr string) (pricefeed.Feed, error) {
	token := common.HexToAddress(tokenStr)
	feed, ok := c.feeds[token]
	if !ok {
		return nil, fmt.Errorf("collector: no feed for token: %s", tokenStr)
	}
	return feed, nil
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
	start := time.Now()
	block, err := c.Provider.BlockByNumber(ctx, nil)
	if err != nil {
		c.metrics.failedFetchBaseFee.Inc()
		c.logger.Warn("collector: error fetching block", "error", err)
		return
	}

	c.lastBaseFee = block.BaseFee()
	c.metrics.baseFee.Set(float64(c.lastBaseFee.Int64()))
	c.metrics.fetchBaseFeeDuration.Observe(time.Since(start).Seconds())
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

		snap, err := feed.Snapshot()
		if err != nil {
			return nil, fmt.Errorf("collector: error fetching feed snapshot: %w", err)
		}
		minFeePerGas = snap.FromNative(minFeePerGas)
	}

	c.metrics.minFeePerGas.WithLabelValues(feeToken.Hex()).Set(float64(minFeePerGas.Int64()))

	return minFeePerGas, nil
}

func (c *Collector) ValidatePayment(op *types.Operation) error {
	minFeePerGas, err := c.MinFeePerGas(op.FeeToken)
	if err != nil {
		return err
	}

	if op.MaxFeePerGas.Cmp(minFeePerGas) < 0 {
		return fmt.Errorf("collector: maxFeePerGas %v < minFeePerGas %v: %w", op.MaxFeePerGas, minFeePerGas, InsufficientFeeError)
	}

	return nil
}

// Compares two operations and returns which one has the highest relay value
func (c *Collector) Cmp(a, b *types.Operation) int {
	nfA, _ := c.NativeFeesPerGas(a)
	nfB, _ := c.NativeFeesPerGas(b)

	// If the difference of maxFeeA is above 10%, then it takes priority
	// difference: abs(maxFeeA - maxFeeB) / maxFeeA
	diffBase := new(big.Int).Abs(new(big.Int).Sub(nfA.MaxFeePerGas, nfB.MaxFeePerGas))
	diffBase.Mul(diffBase, big.NewInt(100))
	diffBase.Div(diffBase, nfA.MaxFeePerGas)

	if diffBase.Cmp(big.NewInt(10)) >= 0 {
		return nfA.MaxFeePerGas.Cmp(nfB.MaxFeePerGas)
	}

	// The difference of baseFee is too small
	// we use the priorityFee to compare
	return nfA.MaxPriorityFeePerGas.Cmp(nfB.MaxPriorityFeePerGas)
}

func (c *Collector) NativeFeesPerGas(op *types.Operation) (*NativeFees, *pricefeed.Snapshot) {
	maxFee := new(big.Int)
	maxFee.Set(op.MaxFeePerGas)

	priorityFee := new(big.Int)
	priorityFee.Set(op.MaxPriorityFeePerGas)

	var snap *pricefeed.Snapshot

	if !op.NativePayment() {
		feed, err := c.Feed(op.FeeToken.String())
		if err == nil {
			snap, err = feed.Snapshot()
			if err == nil {
				maxFee.Mul(maxFee, op.FeeScalingFactor)
				maxFee.Mul(maxFee, snap.NormalizationFactor)

				d := new(big.Int).Set(snap.ScalingFactor)
				d.Mul(d, op.FeeNormalizationFactor)

				maxFee.Div(maxFee, d)

				priorityFee.Mul(priorityFee, op.FeeScalingFactor)
				priorityFee.Mul(priorityFee, snap.NormalizationFactor)

				d = new(big.Int).Set(snap.ScalingFactor)
				d.Mul(d, op.FeeNormalizationFactor)

				priorityFee.Div(priorityFee, d)
			}
		}
	}

	return &NativeFees{
		MaxFeePerGas:         maxFee,
		MaxPriorityFeePerGas: priorityFee,
	}, snap
}

func (c *Collector) FeeAsks() (*proto.FeeAsks, error) {
	if c.lastBaseFee == nil {
		return nil, fmt.Errorf("collector: base fee not fetched")
	}

	acceptedTokens := make(map[string]proto.BaseFeeRate, len(c.feeds))
	for token, feed := range c.feeds {
		snap, err := feed.Snapshot()
		if err != nil {
			c.logger.Warn("collector: error fetching feed factors", "token", token.Hex(), "error", err)
			continue
		}

		acceptedTokens[token.String()] = proto.BaseFeeRate{
			ScalingFactor:       prototyp.ToBigInt(snap.ScalingFactor),
			NormalizationFactor: prototyp.ToBigInt(snap.NormalizationFactor),
		}
	}

	return &proto.FeeAsks{
		MinBaseFee:     prototyp.ToBigInt(c.lastBaseFee),
		MinPriorityFee: prototyp.ToBigInt(c.priorityFee),
		AcceptedTokens: acceptedTokens,
	}, nil
}
