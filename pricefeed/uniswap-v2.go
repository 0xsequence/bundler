package pricefeed

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/0xsequence/bundler/config"
	"github.com/0xsequence/bundler/pricefeed/abis"
	"github.com/0xsequence/ethkit/ethcontract"
	"github.com/0xsequence/ethkit/ethrpc"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/davecgh/go-spew/spew"
	"github.com/go-chi/httplog/v2"
)

const EXPIRATION_TIME = 1 * time.Minute

type UniswapV2Feed struct {
	cfg *config.UniswapV2Reference

	mutex sync.RWMutex

	inverse    bool
	lastUpdate time.Time
	reserve0   *big.Int
	reserve1   *big.Int
	decimals0  int
	decimals1  int

	logger   *httplog.Logger
	contract *ethcontract.Contract

	Provider *ethrpc.Provider
}

func NewUniswapV2Feed(provider *ethrpc.Provider, logger *httplog.Logger, cfg *config.UniswapV2Reference) (*UniswapV2Feed, error) {
	abi := ethcontract.MustParseABI(abis.UNISWAP_V2)
	contract := ethcontract.NewContractCaller(common.HexToAddress(cfg.Pool), abi, provider)

	return &UniswapV2Feed{
		cfg: cfg,

		mutex: sync.RWMutex{},

		logger:   logger,
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
		spew.Dump(err)
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
	return time.Since(f.lastUpdate) < EXPIRATION_TIME
}

func (f *UniswapV2Feed) Start(ctx context.Context) error {
	token0, token1, err := f.fetchTokens()
	if err != nil {
		return fmt.Errorf("uniswap-v2: error fetching tokens: %w", err)
	}

	// If token0 is base token, then inverse is false
	// If token1 is base token, then inverse is true
	// If neither token0 nor token1 is base token, then return error
	if token0.String() == f.cfg.BaseToken {
		f.inverse = false
	} else if token1.String() == f.cfg.BaseToken {
		f.inverse = true
	} else {
		return fmt.Errorf("neither token0 nor token1 is base token")
	}

	// Fetch the decimals of token0 and token1
	f.decimals0, err = FetchDecimals(f.Provider, token0)
	if err != nil {
		return fmt.Errorf("uniswap-v2: error fetching decimals: %w", err)
	}

	f.decimals1, err = FetchDecimals(f.Provider, token1)
	if err != nil {
		return fmt.Errorf("uniswap-v2: error fetching decimals: %w", err)
	}

	for ctx.Err() == nil {
		reserve0, reserve1, _, err := f.fetchReserves()
		if err != nil {
			f.logger.Warn("uniswap-v2: error fetching reserves", "pool", f.cfg.Pool, "error", err)
		}

		f.mutex.Lock()
		f.reserve0 = reserve0
		f.reserve1 = reserve1
		f.lastUpdate = time.Now()
		f.mutex.Unlock()

		r, _ := f.FromNative(big.NewInt(1))
		f.logger.Info("uniswap-v2: fetched token rate", "rate", r.String())

		time.Sleep(5 * time.Second)
	}

	return nil
}

func (f *UniswapV2Feed) getReservesNative0() (r0, r1 *big.Int, d0, d1 int, err error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	if !f.Ready() {
		return nil, nil, 0, 0, fmt.Errorf("uniswap-v2: feed not ready")
	}

	if f.inverse {
		return f.reserve1, f.reserve0, f.decimals1, f.decimals0, nil
	}

	return f.reserve0, f.reserve1, f.decimals0, f.decimals1, nil
}

func normalizeDecimals(r0, r1 *big.Int, d0, d1 int) (nr0, nr1 *big.Int) {
	// Convert r0 and r1 to the same decimals (the highest one)
	if d0 > d1 {
		nr0 = r0
		nr1 = new(big.Int).Mul(r1, new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(d0-d1)), nil))
	} else {
		nr0 = new(big.Int).Mul(r0, new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(d1-d0)), nil))
		nr1 = r1
	}

	return nr0, nr1
}

func (f *UniswapV2Feed) FromNative(native *big.Int) (*big.Int, error) {
	r0, r1, d0, d1, err := f.getReservesNative0()
	if err != nil {
		return nil, err
	}

	nr0, nr1 := normalizeDecimals(r0, r1, d0, d1)

	return new(big.Int).Div(new(big.Int).Mul(native, nr1), nr0), nil
}

func (f *UniswapV2Feed) ToNative(value *big.Int) (*big.Int, error) {
	r0, r1, d0, d1, err := f.getReservesNative0()
	if err != nil {
		return nil, err
	}

	nr0, nr1 := normalizeDecimals(r0, r1, d0, d1)

	return new(big.Int).Div(new(big.Int).Mul(value, nr0), nr1), nil
}
