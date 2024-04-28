package bundler_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/0xsequence/bundler"
	"github.com/0xsequence/bundler/config"
	"github.com/0xsequence/bundler/ipfs"
	"github.com/0xsequence/bundler/lib/mocks"
	"github.com/0xsequence/bundler/p2p"
	"github.com/go-chi/httplog/v2"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOperations(t *testing.T) {
	logger := httplog.NewLogger("")
	ipfs := &mocks.MockIPFS{}
	host := &mocks.MockP2p{}
	mempool := &mocks.MockMempool{}

	archive := bundler.NewArchive(&config.ArchiveConfig{}, host, logger, nil, "", ipfs, mempool)

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
	mipfs := &mocks.MockIPFS{}
	host := &mocks.MockP2p{}
	mempool := &mocks.MockMempool{}

	archive := bundler.NewArchive(&config.ArchiveConfig{}, host, logger, nil, "", mipfs, mempool)

	cid, _ := ipfs.Cid([]byte("hello test"))

	host.On("Address").Return("0x3BC1C2e7120F1a2cf4535C752BE921ABeD2dc14b", nil).Once()
	host.On("Sign", mock.Anything).Return([]byte{0x01, 0x02, 0x03}, nil).Once()
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
		assert.Equal(t, obj.Archive.Identity, "0x3BC1C2e7120F1a2cf4535C752BE921ABeD2dc14b")
		assert.Equal(t, obj.Signature, "0x010203")
	}).Return(cid, nil).Once()

	mempool.On("ForgetOps", time.Minute).Return([]string{
		"0x123",
		"0x456",
	}).Once()

	host.On("Broadcast", mock.Anything, p2p.ArchiveTopic, &bundler.ArchiveMessage{ArchiveCid: cid}).Return(nil).Once()

	ctx, cancel := context.WithCancel(context.Background())
	archive.TriggerArchive(ctx, time.Minute, false)
	cancel()

	mempool.AssertExpectations(t)
	host.AssertExpectations(t)
	mipfs.AssertExpectations(t)
}

func TestListenArchives(t *testing.T) {
	logger := httplog.NewLogger("")
	mipfs := &mocks.MockIPFS{}
	host := &mocks.MockP2p{}
	mempool := &mocks.MockMempool{}

	archive := bundler.NewArchive(&config.ArchiveConfig{}, host, logger, nil, "", mipfs, mempool)

	cHandler := make(chan p2p.MsgHandler)
	host.On("HandleTopic", mock.Anything, p2p.ArchiveTopic, mock.Anything).Run(func(args mock.Arguments) {
		cHandler <- args[2].(p2p.MsgHandler)
	}).Return(nil).Once()

	ctx, cancel := context.WithCancel(context.Background())
	go archive.Run(ctx)

	cid, _ := ipfs.Cid([]byte("hello test 2"))

	handler := <-cHandler

	ad := &bundler.ArchiveMessage{ArchiveCid: cid}
	data, err := json.Marshal(ad)
	assert.Nil(t, err)
	handler(ctx, peer.ID("123"), data)

	for len(archive.SeenArchives()) == 0 {
	}

	expectSeenArchives := map[string]string{}
	expectSeenArchives[peer.ID("123").String()] = cid
	assert.Equal(t, archive.SeenArchives(), expectSeenArchives)

	cid2, _ := ipfs.Cid([]byte("hello test 3"))

	// Doing an archive should broadcast this seen one
	host.On("Address").Return("0x3BC1C2e7120F1a2cf4535C752BE921ABeD2dc14b", nil).Once()
	host.On("Sign", mock.Anything).Return([]byte{0x01, 0x02, 0x03}, nil).Once()
	mipfs.On("Report", mock.Anything).Run(func(args mock.Arguments) {
		data := args.Get(0).([]byte)
		var obj bundler.SignedArchiveSnapshot
		err := json.Unmarshal(data, &obj)
		assert.Nil(t, err)

		assert.Equal(t, obj.Archive.SeenArchives, expectSeenArchives)
		assert.Equal(t, obj.Archive.Operations, []string{})
		assert.Equal(t, obj.Archive.PrevArchive, "")
		assert.Equal(t, obj.Archive.Identity, "0x3BC1C2e7120F1a2cf4535C752BE921ABeD2dc14b")
	}).Return(cid2, nil).Once()
	mempool.On("ForgetOps", time.Minute).Return([]string{}).Once()

	host.On("Broadcast", mock.Anything, p2p.ArchiveTopic, &bundler.ArchiveMessage{ArchiveCid: cid2}).Return(nil).Once()

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
	mipfs := &mocks.MockIPFS{}
	host := &mocks.MockP2p{}
	mempool := &mocks.MockMempool{}

	archive := bundler.NewArchive(&config.ArchiveConfig{}, host, logger, nil, "", mipfs, mempool)

	cid1, _ := ipfs.Cid([]byte("hello test 1"))
	cid2, _ := ipfs.Cid([]byte("hello test 2"))

	host.On("Address").Return("0x3BC1C2e7120F1a2cf4535C752BE921ABeD2dc14b", nil).Twice()
	host.On("Sign", mock.Anything).Return([]byte{0x01, 0x02, 0x03}, nil).Twice()
	mempool.On("ForgetOps", time.Minute).Return([]string{}).Twice()
	host.On("Broadcast", mock.Anything, mock.Anything, mock.Anything).Return(nil).Twice()

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
	mipfs := &mocks.MockIPFS{}
	host := &mocks.MockP2p{}
	mempool := &mocks.MockMempool{}

	archive := bundler.NewArchive(&config.ArchiveConfig{
		RunEveryMillis:     1,
		ForgetAfterSeconds: 13,
	}, host, logger, nil, "", mipfs, mempool)

	cid, _ := ipfs.Cid([]byte("hello test"))

	done := make(chan struct{})

	host.On("Address").Return("0x3BC1C2e7120F1a2cf4535C752BE921ABeD2dc14b", nil).Once()
	host.On("Sign", mock.Anything).Return([]byte{0x01, 0x02, 0x03}, nil).Once()
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
		assert.Equal(t, obj.Archive.Identity, "0x3BC1C2e7120F1a2cf4535C752BE921ABeD2dc14b")
		done <- struct{}{}
	}).Return(cid, nil).Once()

	mempool.On("ForgetOps", time.Second*13).Return([]string{
		"0x123",
	}).Once()

	mempool.On("ForgetOps", time.Second*13).Return([]string{}).Maybe()

	host.On("Broadcast", mock.Anything, p2p.ArchiveTopic, &bundler.ArchiveMessage{ArchiveCid: cid}).Return(nil).Once()
	host.On("HandleTopic", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

	ctx, cancel := context.WithCancel(context.Background())
	go archive.Run(ctx)

	<-done

	cancel()

	mempool.AssertExpectations(t)
	host.AssertExpectations(t)
	mipfs.AssertExpectations(t)
}
