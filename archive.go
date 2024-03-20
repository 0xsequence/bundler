package bundler

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/0xsequence/bundler/config"
	"github.com/0xsequence/bundler/ipfs"
	"github.com/0xsequence/bundler/mempool"
	"github.com/0xsequence/bundler/p2p"
	"github.com/0xsequence/bundler/proto"
	"github.com/0xsequence/bundler/store"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/cyberphone/json-canonicalization/go/src/webpki.org/jsoncanonicalizer"
	"github.com/go-chi/httplog/v2"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/prometheus/client_golang/prometheus"
)

type archiveMetrics struct {
	receivedArchive        *prometheus.CounterVec
	receivedInvalidArchive *prometheus.CounterVec

	seenArchiveCids prometheus.Gauge

	archiveSkipEmpty prometheus.Counter
	doneArchive      prometheus.Histogram
	failedArchive    prometheus.Counter

	failedStoreArchive prometheus.Counter

	receivedArchiveNew   prometheus.Labels
	receivedArchiveKnown prometheus.Labels

	invalidArchiveBadMsgReason prometheus.Labels
	invalidArchiveBadCidReason prometheus.Labels
}

func createArchiveMetrics(reg prometheus.Registerer) *archiveMetrics {
	receivedArchive := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "archive_received_count",
		Help: "Number of received archives",
	}, []string{"status"})

	receivedInvalidArchive := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "archive_received_invalid_count",
		Help: "Number of received invalid archives",
	}, []string{"status"})

	seenArchiveCids := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "archive_seen_cids",
		Help: "Number of seen archive cids, pending records to be archived",
	})

	archiveSkipEmpty := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "archive_skip_empty_count",
		Help: "Number of empty archives skipped",
	})

	failedStoreArchive := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "archive_failed_store_count",
		Help: "Number of failed store archives",
	})

	doneArchive := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "archive_done_operations",
		Help: "Number of operations archived",
		Buckets: []float64{
			0,
			1,
			2,
			5,
			10,
			15,
			20,
			25,
			50,
			75,
			100,
			150,
			200,
			500,
			1000,
			2000,
		},
	})

	failedArchive := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "archive_failed_count",
		Help: "Number of failed archives",
	})

	receivedArchiveNew := prometheus.Labels{"status": "new"}
	receivedArchiveKnown := prometheus.Labels{"status": "known"}

	invalidArchiveBadMsgReason := prometheus.Labels{"reason": "bad_message"}
	invalidArchiveBadCidReason := prometheus.Labels{"reason": "bad_cid"}

	if reg != nil {
		reg.MustRegister(receivedArchive)
		reg.MustRegister(receivedInvalidArchive)
		reg.MustRegister(seenArchiveCids)
		reg.MustRegister(doneArchive)
		reg.MustRegister(failedArchive)
		reg.MustRegister(archiveSkipEmpty)
	}

	return &archiveMetrics{
		receivedArchive:        receivedArchive,
		receivedInvalidArchive: receivedInvalidArchive,

		seenArchiveCids: seenArchiveCids,

		archiveSkipEmpty: archiveSkipEmpty,
		doneArchive:      doneArchive,
		failedArchive:    failedArchive,

		failedStoreArchive: failedStoreArchive,

		receivedArchiveNew:   receivedArchiveNew,
		receivedArchiveKnown: receivedArchiveKnown,

		invalidArchiveBadMsgReason: invalidArchiveBadMsgReason,
		invalidArchiveBadCidReason: invalidArchiveBadCidReason,
	}
}

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

	ipfs  ipfs.Interface
	store store.Store

	runEvery     time.Duration
	forgetAfter  time.Duration
	seenArchives map[string]string

	Host    p2p.Interface
	Logger  *httplog.Logger
	Metrics *archiveMetrics

	PrevArchive string

	Mempool mempool.Interface
}

func NewArchive(cfg *config.ArchiveConfig, host p2p.Interface, logger *httplog.Logger, metrics prometheus.Registerer, store store.Store, ipfs ipfs.Interface, mempool mempool.Interface) *Archive {
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
		ipfs:  ipfs,
		store: store,

		runEvery:    runEvery,
		forgetAfter: forgetAfter,

		Host:    host,
		Logger:  logger,
		Metrics: createArchiveMetrics(metrics),

		seenArchives: make(map[string]string),

		Mempool: mempool,
	}
}

func (a *Archive) Run(ctx context.Context) {
	if a.ipfs == nil {
		a.Logger.Info("archive: ipfs url not set, not archiving")
		return
	}

	// Try to load the previous archive
	prevArchive, _ := a.store.ReadFile("prev_archive")
	if prevArchive != "" {
		a.Logger.Info("archive: loaded previous archive", "cid", prevArchive)
	} else {
		a.Logger.Info("archive: no previous archive found, starting fresh")
	}

	a.Host.HandleMessageType(proto.MessageType_ARCHIVE, func(peer peer.ID, message []byte) {
		a.lock.Lock()
		defer a.lock.Unlock()

		var amsg *ArchiveMessage
		err := json.Unmarshal(message, &amsg)
		if err != nil {
			a.Logger.Warn("archive: invalid message", "peer", peer)
			a.Metrics.receivedInvalidArchive.With(a.Metrics.invalidArchiveBadMsgReason).Inc()
			return
		}

		if !ipfs.IsCid(amsg.ArchiveCid) {
			a.Logger.Warn("archive: invalid cid", "peer", peer, "cid", amsg.ArchiveCid)
			a.Metrics.receivedInvalidArchive.With(a.Metrics.invalidArchiveBadCidReason).Inc()
			return
		}

		ps := peer.String()
		if _, ok := a.seenArchives[ps]; ok {
			a.Metrics.receivedArchive.With(a.Metrics.receivedArchiveKnown).Inc()
		} else {
			a.Metrics.receivedArchive.With(a.Metrics.receivedArchiveNew).Inc()
		}

		a.seenArchives[ps] = amsg.ArchiveCid

		a.Metrics.seenArchiveCids.Set(float64(len(a.seenArchives)))
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
		a.Metrics.failedArchive.Inc()
		a.Logger.Error("archive: error archiving", "ops", len(archive), "error", err)
	}
}

func (a *Archive) doArchive(_ context.Context, ops []string, force bool) error {
	if len(ops) == 0 && !force {
		a.Metrics.archiveSkipEmpty.Inc()
		return nil
	}

	a.lock.Lock()
	defer a.lock.Unlock()

	addr, err := a.Host.Address()
	if err != nil {
		return err
	}

	snapshot := &ArchiveSnapshot{
		Timestamp:    uint64(time.Now().Unix()),
		Identity:     addr,
		Operations:   ops,
		SeenArchives: a.seenArchives,
		PrevArchive:  a.PrevArchive,
	}

	// Convert to json
	snapshotJson, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}

	// Normalize
	snapshotJson, err = jsoncanonicalizer.Transform(snapshotJson)
	if err != nil {
		return fmt.Errorf("unable to normalize archive json: %w", err)
	}

	sig, err := a.Host.Sign(snapshotJson)
	if err != nil {
		return err
	}

	signedSnapshot := &SignedArchiveSnapshot{
		Archive:   snapshot,
		Signature: "0x" + common.Bytes2Hex(sig),
	}

	body, err := json.Marshal(signedSnapshot)
	if err != nil {
		return err
	}

	body, err = jsoncanonicalizer.Transform(body)
	if err != nil {
		return fmt.Errorf("unable to normalize archive json: %w", err)
	}

	cid, err := a.ipfs.Report(body)
	if err != nil {
		return err
	}

	a.Logger.Info("archive: archived", "ops", len(ops), "cid", cid)

	a.PrevArchive = cid

	// Store the previous archive
	err = a.store.WriteFile("prev_archive", cid)
	if err != nil {
		a.Metrics.failedStoreArchive.Inc()
		a.Logger.Warn("archive: failed to store previous archive", "error", err)
	}

	a.seenArchives = make(map[string]string)
	a.Metrics.seenArchiveCids.Set(0)

	// Broadcast the archive
	err = a.Host.Broadcast(proto.Message{
		Type:    proto.MessageType_ARCHIVE,
		Message: &ArchiveMessage{ArchiveCid: cid},
	})

	if err != nil {
		return err
	}

	a.Metrics.doneArchive.Observe(float64(len(ops)))

	return nil
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
