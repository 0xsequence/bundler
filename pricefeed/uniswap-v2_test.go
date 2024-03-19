package pricefeed_test

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/0xsequence/bundler/config"
	"github.com/0xsequence/bundler/mocks"
	"github.com/0xsequence/bundler/pricefeed"
	"github.com/0xsequence/ethkit/go-ethereum"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/0xsequence/ethkit/go-ethereum/common/hexutil"
	"github.com/go-chi/httplog/v2"
	"github.com/stretchr/testify/require"
)

func TestUniswapV2Feed(t *testing.T) {
	pool := "0x1111111111111111111111111111111111111111"
	native := "0x2222222222222222222222222222222222222222"
	usdc := "0x3333333333333333333333333333333333333333"
	poolAddress := common.HexToAddress(pool)

	provider := mocks.MockRPCProvider{}

	var tag *big.Int
	provider.On("CallContract", context.Background(), ethereum.CallMsg{
		To:   &poolAddress,
		Data: hexutil.MustDecode("0x0dfe1681"), // token0()
	}, tag).Return(append(hexutil.MustDecode("0x000000000000000000000000"), hexutil.MustDecode(native)...), nil)

	provider.On("CallContract", context.Background(), ethereum.CallMsg{
		To:   &poolAddress,
		Data: hexutil.MustDecode("0xd21220a7"), // token1()
	}, tag).Return(append(hexutil.MustDecode("0x000000000000000000000000"), hexutil.MustDecode(usdc)...), nil)

	provider.On("CallContract", context.Background(), ethereum.CallMsg{
		To:   &poolAddress,
		Data: hexutil.MustDecode("0x0902f1ac"), // getReserves()
	}, tag).Return(hexutil.MustDecode("0x00000000000000000000000000000000000000000000003635c9adc5dea00000000000000000000000000000000000000000000000000000000002ba7def30000000000000000000000000000000000000000000000000000000000000000000"), nil) // 1e21, 3e9, 0

	feed, err := pricefeed.NewUniswapV2Feed(&provider, httplog.NewLogger("pricefeed"), nil, &config.UniswapV2Reference{
		Pool:      pool,
		BaseToken: native,
	})
	require.NoError(t, err)
	require.False(t, feed.Ready())

	go func() {
		if err := feed.Start(context.Background()); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
		}
	}()

	for !feed.Ready() {
		time.Sleep(time.Second)
	}

	snap, err := feed.Snapshot()
	require.NoError(t, err)

	microUSDC := snap.FromNative(big.NewInt(1000000000000000000))
	require.Zero(t, microUSDC.Cmp(big.NewInt(3000000000)))

	wei := snap.ToNative(microUSDC)
	require.Zero(t, wei.Cmp(big.NewInt(1000000000000000000)))
}
