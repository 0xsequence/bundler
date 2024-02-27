package bundler

import (
	"context"
	"encoding/json"
	"sort"
	"sync"
	"time"

	"github.com/0xsequence/bundler/ipfs"
	"github.com/0xsequence/bundler/mempool"
	"github.com/0xsequence/bundler/p2p"
	"github.com/0xsequence/bundler/proto"
	"github.com/go-chi/httplog/v2"
	"github.com/libp2p/go-libp2p/core/peer"
)

const ArchiveInterval = 5 * time.Minute
const OpTimeToArchive = 5 * time.Minute

type ArchiveSnapshot struct {
	Timestamp uint64 `json:"time"`
	Identity  string `json:"signer"`

	SeenArchives map[string]string `json:"seen_archives"`
	Operations   []string          `json:"operations"`

	PrevArchive string `json:"prev_archive,omitempty"`
}

type SignedArchiveSnapshot struct {
	Archive   *ArchiveSnapshot `json:"archive"`
	Signature string           `json:"signature,omitempty"`
}

type ArchiveMessage struct {
	ArchiveCid string `json:"archive_cid"`
}

type Archive struct {
	lock sync.Mutex

	ipfs *ipfs.Client

	Host   *p2p.Host
	Logger *httplog.Logger

	PrevArchive  string
	SeenArchives map[string]string

	Mempool mempool.Interface
}

func NewArchive(host *p2p.Host, logger *httplog.Logger, ipfs *ipfs.Client, mempool mempool.Interface) *Archive {
	return &Archive{
		lock: sync.Mutex{},
		ipfs: ipfs,

		Host:   host,
		Logger: logger,

		SeenArchives: make(map[string]string),

		Mempool: mempool,
	}
}

func (a *Archive) Run(ctx context.Context) {
	if a.ipfs == nil {
		a.Logger.Info("archive: ipfs url not set, not archiving")
		return
	}

	a.Host.HandleMessageType(proto.MessageType_ARCHIVE, func(peer peer.ID, message []byte) {
		a.lock.Lock()
		defer a.lock.Unlock()

		var amsg *ArchiveMessage
		err := json.Unmarshal(message, &amsg)
		if err != nil {
			a.Logger.Warn("archive: invalid message", "peer", peer)
			return
		}

		if !ipfs.IsCid(amsg.ArchiveCid) {
			a.Logger.Warn("archive: invalid cid", "peer", peer, "cid", amsg.ArchiveCid)
		}

		a.SeenArchives[peer.String()] = amsg.ArchiveCid
	})

	for ctx.Err() == nil {
		time.Sleep(ArchiveInterval)

		// Get the operations that should be archive
		archive := a.Mempool.ForgetOps(OpTimeToArchive)
		err := a.DoArchive(ctx, archive)
		if err != nil {
			a.Logger.Error("archive: error archiving", "ops", len(archive), "error", err)
		}
	}
}

func (a *Archive) DoArchive(ctx context.Context, ops []string) error {
	if len(ops) == 0 {
		return nil
	}

	a.lock.Lock()
	defer a.lock.Unlock()

	snapshot := &ArchiveSnapshot{
		Timestamp:    uint64(time.Now().Unix()),
		Identity:     a.Host.HostID().String(),
		Operations:   ops,
		SeenArchives: a.SeenArchives,
		PrevArchive:  a.PrevArchive,
	}

	// TODO: Sign the snapshot
	signedSnapshot := &SignedArchiveSnapshot{
		Archive:   snapshot,
		Signature: "",
	}

	body, err := json.Marshal(signedSnapshot)
	if err != nil {
		return err
	}

	cid, err := a.ipfs.ReportToIPFS(body)
	if err != nil {
		return err
	}

	a.Logger.Info("archive: archived", "ops", len(ops), "cid", cid)

	a.PrevArchive = cid
	a.SeenArchives = make(map[string]string)

	// Broadcast the archive
	messageType := proto.MessageType_ARCHIVE
	err = a.Host.Broadcast(proto.Message{
		Type:    &messageType,
		Message: &ArchiveMessage{ArchiveCid: cid},
	})

	return err
}

func (a *Archive) Operations(ctx context.Context) *proto.Operations {
	ops := a.Mempool.KnownOperations()

	// Sort the operations
	sort.Strings(ops)

	return &proto.Operations{
		Mempool: ops,
		Archive: a.PrevArchive,
	}
}
