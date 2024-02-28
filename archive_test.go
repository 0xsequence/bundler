package bundler_test

import (
	"context"
	"testing"

	"github.com/0xsequence/bundler"
	"github.com/0xsequence/bundler/mocks"
	"github.com/go-chi/httplog/v2"
	"github.com/stretchr/testify/assert"
)

func TestOperations(t *testing.T) {
	logger := httplog.NewLogger("")
	ipfs := &mocks.MockIpfs{}
	host := &mocks.MockP2p{}
	mempool := &mocks.MockMempool{}

	archive := bundler.NewArchive(host, logger, ipfs, mempool)

	mempool.On("KnownOperations").Return([]string{
		"0x123",
		"0x456",
	}).Once()

	ctx, cancel := context.WithCancel(context.Background())
	ops := archive.Operations(ctx)
	assert.Equal(t, ops.Mempool, []string{
		"0x123",
		"0x456",
	})
	assert.Equal(t, ops.Archive, "")
	cancel()
}
