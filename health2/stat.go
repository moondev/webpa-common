package health2

import "sync/atomic"

type Resetter interface {
	Reset()
}

type Value interface {
	Resetter
	Set(value int64)
	Add(delta int64)
}

type basicStat struct {
	value int64
}

func (b *basicStat) Reset() {
	atomic.StoreInt64(&b.value, 0)
}

func (b *basicStat) Set(value int64) {
	atomic.StoreInt64(&b.value, value)
}

func (b *basicStat) Add(delta int64) {
	atomic.AddInt64(&b.value, delta)
}
