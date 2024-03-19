package registry

import (
	"fmt"
	"sync"
	"time"

	"github.com/0xsequence/bundler/config"
	"github.com/0xsequence/bundler/registry/source"
	"github.com/0xsequence/ethkit/go-ethereum/accounts/abi/bind"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/go-chi/httplog/v2"
	"github.com/prometheus/client_golang/prometheus"
)

type metrics struct {
	knownEndorsers   prometheus.Gauge
	temporalBanned   prometheus.Gauge
	permanentBanned  prometheus.Gauge
	trustedEndorsers prometheus.Gauge

	temporalBans     prometheus.Counter
	temporalForgives prometheus.Counter

	discoverRequests prometheus.Counter
	discoverSuccess  prometheus.Counter
	discoverCollided prometheus.Counter
	discoverFailures *prometheus.CounterVec

	askedForEndorser *prometheus.CounterVec

	discoverFailLowReputation   prometheus.Labels
	discoverFailNoReputation    prometheus.Labels
	discoverFailErrorReputation prometheus.Labels

	discoverTime prometheus.Histogram
}

func createMetrics(reg prometheus.Registerer) *metrics {
	knownEndorsers := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "registry_known_endorsers",
		Help: "Number of known endorsers",
	})

	temporalBanned := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "registry_temporal_banned",
		Help: "Number of temporarily banned endorsers",
	})

	permanentBanned := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "registry_permanent_banned",
		Help: "Number of permanently banned endorsers",
	})

	trustedEndorsers := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "registry_trusted_endorsers",
		Help: "Number of trusted endorsers",
	})

	temporalBans := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "registry_temporal_bans",
		Help: "Number of temporary bans",
	})

	temporalForgives := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "registry_temporal_forgives",
		Help: "Number of temporary ban forgives",
	})

	discoverRequests := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "registry_discover_requests",
		Help: "Number of discover requests",
	})

	discoverSuccess := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "registry_discover_success",
		Help: "Number of successful discovers",
	})

	discoverFailures := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "registry_discover_failures",
		Help: "Number of discover failures",
	}, []string{"reason"})

	discoverCollided := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "registry_discover_collided",
		Help: "Number of discover collisions",
	})

	discoverFailLowReputation := prometheus.Labels{
		"reason": "low_reputation",
	}

	discoverFailNoReputation := prometheus.Labels{
		"reason": "no_reputation",
	}

	discoverFailErrorReputation := prometheus.Labels{
		"reason": "error_reputation",
	}

	askedForEndorser := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "registry_asked_for_endorser",
		Help: "Number of times an endorser was asked for",
	}, []string{"status"})

	discoverTime := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "registry_discover_time",
		Help:    "Time taken to discover an endorser",
		Buckets: prometheus.DefBuckets,
	})

	if reg != nil {
		reg.MustRegister(
			knownEndorsers,
			temporalBanned,
			permanentBanned,
			trustedEndorsers,
			temporalBans,
			temporalForgives,
			discoverRequests,
			discoverSuccess,
			discoverFailures,
			discoverCollided,
			discoverTime,
			askedForEndorser,
		)
	}

	return &metrics{
		knownEndorsers:   knownEndorsers,
		temporalBanned:   temporalBanned,
		permanentBanned:  permanentBanned,
		trustedEndorsers: trustedEndorsers,

		temporalBans:     temporalBans,
		temporalForgives: temporalForgives,

		discoverRequests: discoverRequests,
		discoverSuccess:  discoverSuccess,
		discoverFailures: discoverFailures,
		discoverCollided: discoverCollided,

		askedForEndorser: askedForEndorser,

		discoverFailLowReputation:   discoverFailLowReputation,
		discoverFailNoReputation:    discoverFailNoReputation,
		discoverFailErrorReputation: discoverFailErrorReputation,

		discoverTime: discoverTime,
	}
}

type WeightedSource struct {
	Source source.Interface
	Weight float64
}

type Registry struct {
	lock    sync.RWMutex
	logger  *httplog.Logger
	metrics *metrics

	minReputation   float64
	tempBanDuration time.Duration

	knownEndorsers   map[common.Address]EndorserStatus
	temporalBanStart map[common.Address]time.Time

	Sources []*WeightedSource
}

func NewRegistry(cfg *config.RegistryConfig, logger *httplog.Logger, metrics prometheus.Registerer, caller bind.ContractCaller) (*Registry, error) {
	if cfg.MinReputation == 0 {
		logger.Warn("Minimum endorser reputation is not set, using default value", "default", 0.0)
	} else {
		logger.Info("Minimum endorser reputation set", "minReputation", cfg.MinReputation)
	}

	var tempBanDuration time.Duration
	if cfg.TempBanSeconds == 0 {
		tempBanDuration = 24 * time.Hour
		logger.Warn("Temporary ban duration is not set, using default value", "default", tempBanDuration)
	} else {
		tempBanDuration = time.Duration(cfg.TempBanSeconds) * time.Second
		logger.Info("Temporary ban duration set", "duration", tempBanDuration)
	}

	if len(cfg.Sources) == 0 && cfg.MinReputation != 0 && len(cfg.Trusted) == 0 && !cfg.AllowUnusable {
		return nil, fmt.Errorf("unusable node, no endorser can be trusted, set at least one source or trusted endorser; or set min_reputation to 0")
	}

	r := &Registry{
		logger:           logger,
		metrics:          createMetrics(metrics),
		minReputation:    cfg.MinReputation,
		tempBanDuration:  tempBanDuration,
		knownEndorsers:   make(map[common.Address]EndorserStatus),
		temporalBanStart: make(map[common.Address]time.Time),
	}

	for _, s := range cfg.Sources {
		if !common.IsHexAddress(s.Address) {
			return nil, fmt.Errorf("invalid source address: %s", s.Address)
		}

		addr := common.HexToAddress(s.Address)
		logger.Info("adding source", "source", addr, "weight", s.Weight)

		source, err := source.NewContractSource(caller, addr)
		if err != nil {
			return nil, fmt.Errorf("failed to create source: %w", err)
		}

		r.AddSource(source, s.Weight)
	}

	for _, addr := range cfg.Trusted {
		if !common.IsHexAddress(addr) {
			return nil, fmt.Errorf("invalid trusted address: %s", addr)
		}

		logger.Info("trusting endorser", "endorser", addr)
		r.TrustEndorser(common.HexToAddress(addr))
	}

	return r, nil
}

func (r *Registry) AddSource(s source.Interface, weight float64) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.Sources = append(r.Sources, &WeightedSource{
		Source: s,
		Weight: weight,
	})
}

func (r *Registry) doBan(endorser common.Address, banType BanType) {
	if banType == PermanentBan {
		r.metrics.permanentBanned.Inc()
		r.knownEndorsers[endorser] = PermanentBanned
	} else {
		r.metrics.temporalBanned.Inc()
		r.metrics.temporalBans.Inc()
		r.knownEndorsers[endorser] = TemporaryBanned
		r.temporalBanStart[endorser] = time.Now()
	}
}

func (r *Registry) attemptForgiveTempBan(endorser common.Address) {
	r.lock.RLock()
	banStart := r.temporalBanStart[endorser]
	r.lock.RUnlock()

	if banStart.IsZero() {
		return
	}

	// Forgive requires write lock
	r.lock.Lock()
	defer r.lock.Unlock()

	// Check again if it's still banned
	// as it might have been unbanned by another goroutine
	// during the lock exchange
	banStart = r.temporalBanStart[endorser]
	if banStart.IsZero() {
		return
	}

	if time.Since(banStart) < r.tempBanDuration {
		return
	}

	r.metrics.temporalForgives.Inc()
	r.metrics.temporalBanned.Dec()
	r.logger.Info("forgiving temporary ban", "endorser", endorser.String())
	delete(r.knownEndorsers, endorser)
	delete(r.temporalBanStart, endorser)
}

func (r *Registry) attemptToDiscoverEndorser(endorser common.Address) {
	r.lock.RLock()
	status := r.knownEndorsers[endorser]
	r.lock.RUnlock()

	if status != UnknownEndorser {
		return
	}

	r.metrics.discoverRequests.Inc()

	start := time.Now()
	var totalWeight float64

	for _, source := range r.Sources {
		// Since we don't lock while fetching reputation
		// more than one goroutine might attempt to discover it
		// this is fine, we can optimize it later if it becomes a problem
		reputation, err := source.Source.ReputationForEndorser(endorser)

		if err != nil {
			r.metrics.discoverFailures.With(r.metrics.discoverFailErrorReputation).Inc()
			r.logger.Warn("unable to get reputation from source", "source", source.Source, "endorser", endorser.String(), "error", err)
			continue
		}

		totalWeight += float64(reputation.Int64()) * source.Weight
	}

	if totalWeight >= r.minReputation {
		r.lock.Lock()
		defer r.lock.Unlock()

		if r.knownEndorsers[endorser] != UnknownEndorser {
			r.metrics.discoverCollided.Inc()
			r.logger.Info("discovered new endorser, but it was already discovered", "endorser", endorser.String(), "reputation", totalWeight)
			return
		}

		r.metrics.knownEndorsers.Inc()
		r.metrics.discoverSuccess.Inc()
		r.logger.Info("discovered new endorser", "endorser", endorser.String(), "reputation", totalWeight)
		r.knownEndorsers[endorser] = AcceptedEndorser
	} else {
		// Endorser still unknown, we don't need to do anything
		if totalWeight == 0 {
			r.metrics.discoverFailures.With(r.metrics.discoverFailNoReputation).Inc()
		} else {
			r.metrics.discoverFailures.With(r.metrics.discoverFailLowReputation).Inc()
		}

		r.logger.Info("unable to discover endorser", "endorser", endorser.String(), "reputation", totalWeight)
	}

	r.metrics.discoverTime.Observe(time.Since(start).Seconds())
}

func (r *Registry) TrustEndorser(endorser common.Address) {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.knownEndorsers[endorser] == TemporaryBanned || r.knownEndorsers[endorser] == PermanentBanned {
		r.logger.Warn("trusting a banned endorser! trust takes priority", "endorser", endorser.String())
	}

	r.metrics.knownEndorsers.Inc()
	r.metrics.trustedEndorsers.Inc()
	r.logger.Info("trusting endorser", "endorser", endorser.String(), "prev", r.knownEndorsers[endorser])
	r.knownEndorsers[endorser] = TrustedEndorser
}

func (r *Registry) BanEndorser(endorser common.Address, banType BanType) {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.knownEndorsers[endorser] == TemporaryBanned {
		delete(r.temporalBanStart, endorser)
		r.doBan(endorser, banType)
		return
	}

	if r.knownEndorsers[endorser] == PermanentBanned {
		r.logger.Warn("attempt to ban an already permanently banned endorser", "endorser", endorser.String())
		return
	}

	if r.knownEndorsers[endorser] == TrustedEndorser {
		r.logger.Warn("attempt to ban a trusted endorser", "endorser", endorser.String())
		return
	}

	r.logger.Info("banning endorser", "endorser", endorser.String(), "prev", r.knownEndorsers[endorser], "banType", banType)
	r.doBan(endorser, banType)
}

func (r *Registry) StatusForEndorser(endorser common.Address) EndorserStatus {
	r.attemptForgiveTempBan(endorser)
	r.attemptToDiscoverEndorser(endorser)

	r.lock.RLock()
	defer r.lock.RUnlock()

	status := r.knownEndorsers[endorser]
	r.metrics.askedForEndorser.WithLabelValues(status.String()).Inc()
	return status
}

func (r *Registry) IsAcceptedEndorser(endorser common.Address) bool {
	s := r.StatusForEndorser(endorser)
	return s == AcceptedEndorser || s == TrustedEndorser
}

func (r *Registry) KnownEndorsers() []*KnownEndorser {
	r.lock.RLock()
	defer r.lock.RUnlock()

	endorsers := make([]*KnownEndorser, len(r.knownEndorsers))

	var i int
	for addr, status := range r.knownEndorsers {
		endorsers[i] = &KnownEndorser{
			Address: addr,
			Status:  status,
		}
		i++
	}

	return endorsers
}

var _ Interface = &Registry{}
