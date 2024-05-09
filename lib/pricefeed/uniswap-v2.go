package pricefeed

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/0xsequence/bundler/config"
	"github.com/0xsequence/bundler/lib/pricefeed/abis"
	"github.com/0xsequence/ethkit/ethcontract"
	"github.com/0xsequence/ethkit/ethrpc"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/go-chi/httplog/v2"
	"github.com/prometheus/client_golang/prometheus"
)

const EXPIRATION_TIME = 1 * time.Minute

type uniswapV2Metrics struct {
	reserve0   prometheus.Gauge
	reserve1   prometheus.Gauge
	lastUpdate prometheus.Gauge

	rate  prometheus.Gauge
	ready prometheus.Gauge

	fetchReservesError prometheus.Counter
	fetchReservesTime  prometheus.Histogram
}

func createUniswapV2Metrics(reg prometheus.Registerer, pool string, inverse bool) *uniswapV2Metrics {
	reserve0 := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "uniswap_v2_reserve0",
	})

	reserve1 := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "uniswap_v2_reserve1",
	})

	lastUpdate := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "uniswap_v2_last_update",
	})

	rate := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "uniswap_v2_rate",
		Help: "The exchange rate of token0 to token1",
	})

	ready := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "uniswap_v2_ready",
	})

	fetchReservesError := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "uniswap_v2_fetch_reserves_error",
		Help: "The number of errors fetching reserves",
	})

	fetchReservesTime := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "uniswap_v2_fetch_reserves_time",
		Help: "The time taken to fetch reserves",
	})

	if reg != nil {
		regTagged := prometheus.WrapRegistererWith(prometheus.Labels{
			"pool":    pool,
			"inverse": fmt.Sprintf("%t", inverse),
		}, reg)

		regTagged.MustRegister(
			reserve0,
			reserve1,
			lastUpdate,
			rate,
			ready,
			fetchReservesError,
			fetchReservesTime,
		)
	}

	return &uniswapV2Metrics{
		reserve0:           reserve0,
		reserve1:           reserve1,
		lastUpdate:         lastUpdate,
		rate:               rate,
		ready:              ready,
		fetchReservesError: fetchReservesError,
		fetchReservesTime:  fetchReservesTime,
	}
}

type UniswapV2Feed struct {
	cfg *config.UniswapV2Reference

	mutex sync.RWMutex

	inverse    bool
	lastUpdate time.Time
	reserve0   *big.Int
	reserve1   *big.Int

	logger  *httplog.Logger
	metrics *uniswapV2Metrics

	contract *ethcontract.Contract

	Provider ethrpc.Interface
}

func NewUniswapV2Feed(provider ethrpc.Interface, logger *httplog.Logger, metrics prometheus.Registerer, cfg *config.UniswapV2Reference) (*UniswapV2Feed, error) {
	abi := ethcontract.MustParseABI(abis.UNISWAP_V2)
	contract := ethcontract.NewContractCaller(common.HexToAddress(cfg.Pool), abi, provider)

	return &UniswapV2Feed{
		cfg: cfg,

		mutex: sync.RWMutex{},

		logger:  logger,
		metrics: createUniswapV2Metrics(metrics, cfg.Pool, false),

		contract: contract,

		Provider: provider,
	}, nil
}

func (f *UniswapV2Feed) fetchReserves() (reserve0, reserve1 *big.Int, timestamp uint32, err error) {
	var result []interface{}
	err = f.contract.Call(nil, &result, "getReserves")
	if err != nil {
		return nil, nil, 0, err
	}

	return result[0].(*big.Int), result[1].(*big.Int), result[2].(uint32), nil
}

func (f *UniswapV2Feed) fetchTokens() (token0, token1 common.Address, err error) {
	var result1 []interface{}
	err = f.contract.Call(nil, &result1, "token0")
	if err != nil {
		return common.Address{}, common.Address{}, err
	}

	token0 = result1[0].(common.Address)

	var result2 []interface{}
	err = f.contract.Call(nil, &result2, "token1")
	if err != nil {
		return common.Address{}, common.Address{}, err
	}

	token1 = result2[0].(common.Address)

	return token0, token1, nil
}

func (f *UniswapV2Feed) Name() string {
	return "uniswap-v2-" + f.cfg.Pool
}

func (f *UniswapV2Feed) Ready() bool {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	ready := time.Since(f.lastUpdate) < EXPIRATION_TIME

	if ready {
		f.metrics.ready.Set(1)
	} else {
		f.metrics.ready.Set(0)
	}

	return ready
}

func (f *UniswapV2Feed) Start(ctx context.Context) error {
	token0, token1, err := f.fetchTokens()
	if err != nil {
		return fmt.Errorf("uniswap-v2: error fetching tokens: %w", err)
	}

	// If token0 is base token, then inverse is false
	// If token1 is base token, then inverse is true
	// If neither token0 nor token1 is base token, then return error
	baseToken := common.HexToAddress(f.cfg.BaseToken)
	if token0 == baseToken {
		f.inverse = false
	} else if token1 == baseToken {
		f.inverse = true
	} else {
		return fmt.Errorf("neither token0 nor token1 is base token %s %s %s", f.cfg.BaseToken, token0.String(), token1.String())
	}

	for ctx.Err() == nil {
		f.doFetch()
		time.Sleep(5 * time.Second)
	}

	return nil
}

func (f *UniswapV2Feed) doFetch() {
	start := time.Now()

	reserve0, reserve1, _, err := f.fetchReserves()
	if err != nil {
		f.metrics.fetchReservesError.Inc()
		f.logger.Warn("uniswap-v2: error fetching reserves", "pool", f.cfg.Pool, "error", err)
		return
	}

	f.metrics.fetchReservesTime.Observe(time.Since(start).Seconds())

	if reserve0 == nil || reserve1 == nil {
		f.metrics.fetchReservesError.Inc()
		f.logger.Warn("uniswap-v2: reserves are nil", "pool", f.cfg.Pool)
		return
	}

	r0float, _ := reserve0.Float64()
	r1float, _ := reserve1.Float64()
	f.metrics.reserve0.Set(r0float)
	f.metrics.reserve1.Set(r1float)
	f.metrics.lastUpdate.Set(float64(time.Now().Unix()))

	f.mutex.Lock()
	f.reserve0 = reserve0
	f.reserve1 = reserve1
	f.lastUpdate = time.Now()
	f.mutex.Unlock()

	s, _ := f.Snapshot()
	if s != nil {
		base, _ := big.NewInt(0).SetString("1000000000000000000", 10)
		price := s.FromNative(base)
		priceFloat, _ := price.Float64()
		f.metrics.rate.Set(priceFloat)
		f.logger.Debug("uniswap-v2: fetched token rate", "rate", price)
	}
}

func (f *UniswapV2Feed) getReservesNative0() (r0, r1 *big.Int, err error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	if !f.Ready() {
		return nil, nil, fmt.Errorf("uniswap-v2: feed not ready")
	}

	if f.inverse {
		return f.reserve0, f.reserve1, nil
	}

	return f.reserve1, f.reserve0, nil
}

func (f *UniswapV2Feed) Snapshot() (*Snapshot, error) {
	r0, r1, err := f.getReservesNative0()
	if err != nil {
		return nil, err
	}

	return &Snapshot{
		ScalingFactor:       r0,
		NormalizationFactor: r1,
	}, nil
}

var _ Feed = (*UniswapV2Feed)(nil)
