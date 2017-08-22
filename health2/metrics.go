package health2

import (
	"sync"
	"sync/atomic"

	"github.com/go-kit/kit/metrics"
)

type Counter struct {
	value uint64
}

func (c *Counter) With(labelValues ...string) metrics.Counter {
	return c
}

func (c *Counter) Add(delta float64) {
	atomic.AddUint64(&c.value, uint64(delta))
}

func (c *Counter) Increment() {
	atomic.AddUint64(&c.value, 1)
}

type Gauge struct {
	mutex sync.RWMutex
	value float64
}

func (g *Gauge) With(labelValues ...string) metrics.Gauge {
	return g
}

func (g *Gauge) Set(value float64) {
	g.mutex.Lock()
	g.value = value
	g.mutex.Unlock()
}

func (g *Gauge) Add(value float64) {
	g.mutex.Lock()
	g.value += value
	g.mutex.Unlock()
}

func (g *Gauge) Increment() {
	g.mutex.Lock()
	g.value++
	g.mutex.Unlock()
}
