package collector_test

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"testing"

	"github.com/0xsequence/bundler/collector"
	"github.com/0xsequence/bundler/config"
	"github.com/0xsequence/bundler/contracts/gen/solabis/abiendorser"
	"github.com/0xsequence/bundler/mocks"
	"github.com/0xsequence/bundler/pricefeed"
	"github.com/0xsequence/bundler/proto"
	"github.com/0xsequence/bundler/types"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	gethTypes "github.com/0xsequence/ethkit/go-ethereum/core/types"
	"github.com/0xsequence/go-sequence/lib/prototyp"
	"github.com/go-chi/httplog/v2"
	"github.com/stretchr/testify/require"
)

func TestValidatePayment(t *testing.T) {
	provider := &mocks.MockRPCProvider{}

	var tag *big.Int

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	provider.
		On("BlockByNumber", ctx, tag).
		Return(gethTypes.NewBlockWithHeader(&gethTypes.Header{BaseFee: big.NewInt(10000000000)}), nil)

	c, err := collector.NewCollector(
		&config.CollectorConfig{},
		httplog.NewLogger("collector"),
		nil,
		provider,
	)
	require.NoError(t, err)

	bytes := make([]byte, 20)
	_, err = rand.Read(bytes)
	require.NoError(t, err)

	go func() {
		if err := c.Run(ctx); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
		}
	}()

	for c.BaseFee() == nil {
	}

	op := types.Operation{
		IEndorserOperation: abiendorser.IEndorserOperation{
			MaxFeePerGas:           big.NewInt(100000000000),
			MaxPriorityFeePerGas:   big.NewInt(1000000000),
			FeeScalingFactor:       big.NewInt(3),
			FeeNormalizationFactor: big.NewInt(1000000000),
		},
	}

	err = c.ValidatePayment(&op)
	require.NoError(t, err)

	op.MaxFeePerGas = big.NewInt(9999999999)

	err = c.ValidatePayment(&op)
	require.ErrorIs(t, err, collector.InsufficientFeeError)
}

func TestFeeAsks(t *testing.T) {
	provider := &mocks.MockRPCProvider{}
	mockFeed1 := &mocks.MockFeed{}
	mockFeed2 := &mocks.MockFeed{}

	mockFeed1.On("Name").Return("mockFeed1").Maybe()
	mockFeed2.On("Name").Return("mockFeed2").Maybe()

	feeTokenAddr1 := common.HexToAddress("0xc0ffee254729296a45a3885639AC7E10F9d54979")
	feeTokenAddr2 := common.HexToAddress("0xdeadbeef254729296a45a3885639AC7E10F9d549")

	var tag *big.Int

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	provider.
		On("BlockByNumber", ctx, tag).
		Return(gethTypes.NewBlockWithHeader(&gethTypes.Header{BaseFee: big.NewInt(10000000000)}), nil)

	c, err := collector.NewCollector(
		&config.CollectorConfig{
			PriorityFee: 42,
		},
		httplog.NewLogger("collector"),
		nil,
		provider,
	)

	require.NoError(t, err)

	err = c.AddFeed(feeTokenAddr1.String(), mockFeed1)
	require.NoError(t, err)

	err = c.AddFeed(feeTokenAddr2.String(), mockFeed2)
	require.NoError(t, err)

	go func() {
		if err := c.Run(ctx); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
		}
	}()

	for c.BaseFee() == nil {
	}

	mockFeed1.On("Snapshot").Return(&pricefeed.Snapshot{
		ScalingFactor:       big.NewInt(2),
		NormalizationFactor: big.NewInt(3),
	}, nil).Once()
	mockFeed2.On("Snapshot").Return(&pricefeed.Snapshot{
		ScalingFactor:       big.NewInt(4),
		NormalizationFactor: big.NewInt(5),
	}, nil).Once()

	res, err := c.FeeAsks()
	require.NoError(t, err)

	require.Equal(t, res, &proto.FeeAsks{
		MinBaseFee:     prototyp.NewBigIntFromString("10000000000", 10),
		MinPriorityFee: prototyp.NewBigIntFromString("42", 10),
		AcceptedTokens: map[string]proto.BaseFeeRate{
			feeTokenAddr1.String(): {
				ScalingFactor:       prototyp.NewBigIntFromString("2", 10),
				NormalizationFactor: prototyp.NewBigIntFromString("3", 10),
			},
			feeTokenAddr2.String(): {
				ScalingFactor:       prototyp.NewBigIntFromString("4", 10),
				NormalizationFactor: prototyp.NewBigIntFromString("5", 10),
			},
		},
	})

	// Must ignore the token if one of them fails
	mockFeed1.On("Snapshot").Return(&pricefeed.Snapshot{
		ScalingFactor:       big.NewInt(10),
		NormalizationFactor: big.NewInt(11),
	}, nil).Once()
	mockFeed2.On("Snapshot").Return(nil, fmt.Errorf("mock error")).Once()

	res, err = c.FeeAsks()
	require.NoError(t, err)

	require.Equal(t, res, &proto.FeeAsks{
		MinBaseFee:     prototyp.NewBigIntFromString("10000000000", 10),
		MinPriorityFee: prototyp.NewBigIntFromString("42", 10),
		AcceptedTokens: map[string]proto.BaseFeeRate{
			feeTokenAddr1.String(): {
				ScalingFactor:       prototyp.NewBigIntFromString("10", 10),
				NormalizationFactor: prototyp.NewBigIntFromString("11", 10),
			},
		},
	})
}
