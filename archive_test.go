package bundler_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/0xsequence/bundler"
	"github.com/0xsequence/bundler/config"
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

	archive := bundler.NewArchive(&config.ArchiveConfig{}, host, logger, ipfs, mempool)

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

	archive := bundler.NewArchive(&config.ArchiveConfig{}, host, logger, mipfs, mempool)

	cid, _ := ipfs.Cid([]byte("hello test"))

	host.On("HostID").Return(peer.ID("1")).Once()
	mipfs.On("Report", mock.Anything).Run(func(args mock.Arguments) {
		data := args.Get(0).([]byte)
		var obj bundler.SignedArchiveSnapshot
		err := json.Unmarshal(data, &obj)
		assert.Nil(t, err)

		assert.Equal(t, obj.Archive.SeenArchives, make(map[string]string))
		assert.Equal(t, obj.Archive.Operations, []string{
			"0x123",
			"0x456",
		})
		assert.Equal(t, obj.Archive.PrevArchive, "")
		assert.Equal(t, obj.Archive.Identity, peer.ID("1").String())
	}).Return(cid, nil).Once()

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

func TestListenArchives(t *testing.T) {
	logger := httplog.NewLogger("")
	mipfs := &mocks.MockIpfs{}
	host := &mocks.MockP2p{}
	mempool := &mocks.MockMempool{}

	archive := bundler.NewArchive(&config.ArchiveConfig{}, host, logger, mipfs, mempool)

	handlerregistered := make(chan struct{})
	host.On("HandleMessageType", proto.MessageType_ARCHIVE, mock.Anything).Run(func(mock.Arguments) {
		handlerregistered <- struct{}{}
	}).Return(nil).Once()

	ctx, cancel := context.WithCancel(context.Background())
	go archive.Run(ctx)

	cid, _ := ipfs.Cid([]byte("hello test 2"))

	<-handlerregistered

	host.ExtBroadcast(peer.ID("123"), proto.MessageType_ARCHIVE, bundler.ArchiveMessage{ArchiveCid: cid})

	for len(archive.SeenArchives()) == 0 {
	}

	expectSeenArchives := map[string]string{}
	expectSeenArchives[peer.ID("123").String()] = cid
	assert.Equal(t, archive.SeenArchives(), expectSeenArchives)

	cid2, _ := ipfs.Cid([]byte("hello test 3"))

	// Doing an archive should broadcast this seen one
	host.On("HostID").Return(peer.ID("456")).Once()
	mipfs.On("Report", mock.Anything).Run(func(args mock.Arguments) {
		data := args.Get(0).([]byte)
		var obj bundler.SignedArchiveSnapshot
		err := json.Unmarshal(data, &obj)
		assert.Nil(t, err)

		assert.Equal(t, obj.Archive.SeenArchives, expectSeenArchives)
		assert.Equal(t, obj.Archive.Operations, []string{})
		assert.Equal(t, obj.Archive.PrevArchive, "")
		assert.Equal(t, obj.Archive.Identity, peer.ID("456").String())
	}).Return(cid2, nil).Once()
	mempool.On("ForgetOps", time.Minute).Return([]string{}).Once()

	mtype := proto.MessageType_ARCHIVE
	host.On("Broadcast", proto.Message{
		Type: &mtype,
		Message: &bundler.ArchiveMessage{
			ArchiveCid: cid2,
		},
	}).Return(nil).Once()

	archive.TriggerArchive(ctx, time.Minute, true)

	host.AssertExpectations(t)
	mipfs.AssertExpectations(t)
	mempool.AssertExpectations(t)

	// Should have reset the seen archives
	assert.Equal(t, archive.SeenArchives(), map[string]string{})

	cancel()
}

func TestChainArchives(t *testing.T) {
	logger := httplog.NewLogger("")
	mipfs := &mocks.MockIpfs{}
	host := &mocks.MockP2p{}
	mempool := &mocks.MockMempool{}

	archive := bundler.NewArchive(&config.ArchiveConfig{}, host, logger, mipfs, mempool)

	cid1, _ := ipfs.Cid([]byte("hello test 1"))
	cid2, _ := ipfs.Cid([]byte("hello test 2"))

	host.On("HostID").Return(peer.ID("1")).Twice()
	mempool.On("ForgetOps", time.Minute).Return([]string{}).Twice()
	host.On("Broadcast", mock.Anything).Return(nil).Twice()

	mipfs.On("Report", mock.Anything).Return(cid1, nil).Once()

	archive.TriggerArchive(context.Background(), time.Minute, true)

	assert.Equal(t, archive.PrevArchive, cid1)

	mipfs.On("Report", mock.Anything).Run(func(args mock.Arguments) {
		data := args.Get(0).([]byte)
		var obj bundler.SignedArchiveSnapshot
		err := json.Unmarshal(data, &obj)
		assert.Nil(t, err)

		assert.Equal(t, obj.Archive.PrevArchive, cid1)
	}).Return(cid2, nil).Once()

	archive.TriggerArchive(context.Background(), time.Minute, true)

	assert.Equal(t, archive.PrevArchive, cid2)

	host.AssertExpectations(t)
	mipfs.AssertExpectations(t)
	mempool.AssertExpectations(t)
}

func TestRunArchiver(t *testing.T) {
	logger := httplog.NewLogger("")
	mipfs := &mocks.MockIpfs{}
	host := &mocks.MockP2p{}
	mempool := &mocks.MockMempool{}

	archive := bundler.NewArchive(&config.ArchiveConfig{
		RunEveryMillis:     1,
		ForgetAfterSeconds: 13,
	}, host, logger, mipfs, mempool)

	cid, _ := ipfs.Cid([]byte("hello test"))

	done := make(chan struct{})

	host.On("HostID").Return(peer.ID("1")).Once()
	mipfs.On("Report", mock.Anything).Run(func(args mock.Arguments) {
		data := args.Get(0).([]byte)
		var obj bundler.SignedArchiveSnapshot
		err := json.Unmarshal(data, &obj)
		assert.Nil(t, err)

		assert.Equal(t, obj.Archive.SeenArchives, make(map[string]string))
		assert.Equal(t, obj.Archive.Operations, []string{
			"0x123",
		})
		assert.Equal(t, obj.Archive.PrevArchive, "")
		assert.Equal(t, obj.Archive.Identity, peer.ID("1").String())
		done <- struct{}{}
	}).Return(cid, nil).Once()

	mempool.On("ForgetOps", time.Second*13).Return([]string{
		"0x123",
	}).Once()

	mempool.On("ForgetOps", time.Second*13).Return([]string{}).Maybe()

	messageType := proto.MessageType_ARCHIVE
	host.On("Broadcast", proto.Message{
		Type:    &messageType,
		Message: &bundler.ArchiveMessage{ArchiveCid: cid},
	}).Return(nil).Once()
	host.On("HandleMessageType", proto.MessageType_ARCHIVE, mock.Anything).Return(nil).Once()

	ctx, cancel := context.WithCancel(context.Background())
	go archive.Run(ctx)

	<-done

	cancel()

	mempool.AssertExpectations(t)
	host.AssertExpectations(t)
	mipfs.AssertExpectations(t)
}
