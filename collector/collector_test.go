package collector_test

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/0xsequence/bundler/collector"
	"github.com/0xsequence/bundler/config"
	"github.com/0xsequence/bundler/mocks"
	"github.com/0xsequence/bundler/types"
	gethTypes "github.com/0xsequence/ethkit/go-ethereum/core/types"
	"github.com/go-chi/httplog/v2"
	"github.com/stretchr/testify/require"
)

func TestValidatePayment(t *testing.T) {
	provider := &mocks.MockRPCProvider{}

	var tag *big.Int

	provider.
		On("BlockByNumber", context.Background(), tag).
		Return(gethTypes.NewBlockWithHeader(&gethTypes.Header{BaseFee: big.NewInt(10000000000)}), nil)

	c, err := collector.NewCollector(
		&config.CollectorConfig{},
		httplog.NewLogger("collector"),
		provider,
	)
	require.NoError(t, err)

	bytes := make([]byte, 20)
	_, err = rand.Read(bytes)
	require.NoError(t, err)

	go func() {
		if err := c.Run(context.Background()); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
		}
	}()

	for c.BaseFee() == nil {
		time.Sleep(time.Second)
	}

	op := types.Operation{
		MaxFeePerGas:               big.NewInt(100000000000),
		PriorityFeePerGas:          big.NewInt(1000000000),
		BaseFeeScalingFactor:       big.NewInt(3),
		BaseFeeNormalizationFactor: big.NewInt(1000000000),
	}

	err = c.ValidatePayment(&op)
	require.NoError(t, err)

	op.MaxFeePerGas = big.NewInt(9999999999)

	err = c.ValidatePayment(&op)
	require.ErrorIs(t, err, collector.InsufficientFeeError)
}
