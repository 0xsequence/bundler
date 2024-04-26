package chiller

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Chiller struct {
	sync.Mutex

	metrics *metrics

	wait time.Duration

	chilled map[string]time.Time
	frozen  map[string]struct{}
}

func NewChiller(wait time.Duration) *Chiller {
	return &Chiller{
		metrics: createMetrics(),

		wait: wait,

		chilled: make(map[string]time.Time),
		frozen:  make(map[string]struct{}),
	}
}

func (c *Chiller) SetRegisterer(reg prometheus.Registerer) {
	c.metrics.register(reg)
}

func (c *Chiller) Chill(oph string) {
	c.Lock()
	defer c.Unlock()
	c.ChillLocked(oph)
}

func (c *Chiller) ChillLocked(oph string) {
	c.metrics.chilledOps.Inc()
	c.chilled[oph] = time.Now()
}

func (c *Chiller) Freeze(oph string) {
	c.Lock()
	defer c.Unlock()
	c.FreezeLocked(oph)
}

func (c *Chiller) FreezeLocked(oph string) {
	c.metrics.blockedOps.Inc()
	c.frozen[oph] = struct{}{}
}

func (c *Chiller) Has(oph string) bool {
	c.Lock()
	defer c.Unlock()

	return c.HasLocked(oph)
}

func (c *Chiller) HasLocked(oph string) bool {
	_, ok := c.frozen[oph]
	if ok {
		return true
	}

	t, ok := c.chilled[oph]
	if !ok {
		return false
	}

	if time.Since(t) > c.wait {
		c.metrics.chilledOps.Dec()
		delete(c.chilled, oph)
		return false
	}

	return true
}
