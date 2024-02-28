package bundler_test

import (
	"context"
	"testing"
	"time"

	"github.com/0xsequence/bundler"
	"github.com/0xsequence/bundler/ipfs"
	"github.com/0xsequence/bundler/mocks"
	"github.com/0xsequence/bundler/proto"
	"github.com/go-chi/httplog/v2"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func TestDoArchive(t *testing.T) {
	logger := httplog.NewLogger("")
	mipfs := &mocks.MockIpfs{}
	host := &mocks.MockP2p{}
	mempool := &mocks.MockMempool{}

	archive := bundler.NewArchive(host, logger, mipfs, mempool)

	cid, _ := ipfs.Cid([]byte("hello test"))

	host.On("HostID").Return(peer.ID("")).Once()
	mipfs.On("Report", mock.Anything).Return(cid, nil).Once()

	mempool.On("ForgetOps", time.Minute).Return([]string{
		"0x123",
		"0x456",
	}).Once()

	messageType := proto.MessageType_ARCHIVE
	host.On("Broadcast", proto.Message{
		Type:    &messageType,
		Message: &bundler.ArchiveMessage{ArchiveCid: cid},
	}).Return(nil).Once()

	ctx, cancel := context.WithCancel(context.Background())
	archive.TriggerArchive(ctx, time.Minute, false)
	cancel()

	mempool.AssertExpectations(t)
	host.AssertExpectations(t)
	mipfs.AssertExpectations(t)
}
