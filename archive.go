package bundler

import (
	"context"
	"encoding/json"
	"sort"
	"sync"
	"time"

	"github.com/0xsequence/bundler/config"
	"github.com/0xsequence/bundler/ipfs"
	"github.com/0xsequence/bundler/mempool"
	"github.com/0xsequence/bundler/p2p"
	"github.com/0xsequence/bundler/proto"
	"github.com/go-chi/httplog/v2"
	"github.com/libp2p/go-libp2p/core/peer"
)

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

	ipfs ipfs.Interface

	runEvery     time.Duration
	forgetAfter  time.Duration
	seenArchives map[string]string

	Host   p2p.Interface
	Logger *httplog.Logger

	PrevArchive string

	Mempool mempool.Interface
}

func NewArchive(cfg *config.ArchiveConfig, host p2p.Interface, logger *httplog.Logger, ipfs ipfs.Interface, mempool mempool.Interface) *Archive {
	var runEvery time.Duration
	if cfg.RunEveryMillis != 0 {
		runEvery = time.Duration(cfg.RunEveryMillis) * time.Millisecond
	} else {
		runEvery = 5 * time.Minute
	}

	var forgetAfter time.Duration
	if cfg.ForgetAfterSeconds != 0 {
		forgetAfter = time.Duration(cfg.ForgetAfterSeconds) * time.Second
	} else {
		forgetAfter = 15 * time.Minute
	}

	if logger != nil {
		logger.Info("archive: initialized", "run_every", runEvery, "forget_after", forgetAfter)
	}

	return &Archive{
		lock: sync.Mutex{},
		ipfs: ipfs,

		runEvery:    runEvery,
		forgetAfter: forgetAfter,

		Host:   host,
		Logger: logger,

		seenArchives: make(map[string]string),

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

		a.seenArchives[peer.String()] = amsg.ArchiveCid
	})

	for ctx.Err() == nil {
		time.Sleep(a.runEvery)
		a.TriggerArchive(ctx, a.forgetAfter, false)
	}
}

func (a *Archive) SeenArchives() map[string]string {
	a.lock.Lock()
	defer a.lock.Unlock()

	// Copy the map
	seen := make(map[string]string)
	for k, v := range a.seenArchives {
		seen[k] = v
	}

	return seen
}

func (a *Archive) Stop(ctx context.Context) {
	a.Logger.Info("archive: stopping..")
	a.TriggerArchive(ctx, 0, true)

	// Delay 250ms to ensure the archive is published
	// to the network before exiting
	time.Sleep(250 * time.Millisecond)
	a.Logger.Info("archive: stopped, published last archive", "cid", a.PrevArchive)
}

func (a *Archive) TriggerArchive(ctx context.Context, age time.Duration, force bool) {
	// Get the operations that should be archive
	archive := a.Mempool.ForgetOps(age)
	err := a.doArchive(ctx, archive, force)
	if err != nil {
		a.Logger.Error("archive: error archiving", "ops", len(archive), "error", err)
	}
}

func (a *Archive) doArchive(ctx context.Context, ops []string, force bool) error {
	if len(ops) == 0 && !force {
		return nil
	}

	a.lock.Lock()
	defer a.lock.Unlock()

	snapshot := &ArchiveSnapshot{
		Timestamp:    uint64(time.Now().Unix()),
		Identity:     a.Host.HostID().String(),
		Operations:   ops,
		SeenArchives: a.seenArchives,
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

	cid, err := a.ipfs.Report(body)
	if err != nil {
		return err
	}

	a.Logger.Info("archive: archived", "ops", len(ops), "cid", cid)

	a.PrevArchive = cid
	a.seenArchives = make(map[string]string)

	// Broadcast the archive
	err = a.Host.Broadcast(proto.Message{
		Type:    proto.MessageType_ARCHIVE,
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
