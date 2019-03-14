package throttle

import (
	"sync"
	"time"
)

// New Returns new Limiter instance.
func NewTBA(t time.Duration, n int64) *Limiter {
	return &Limiter{
		capacity: n,
		initTime: time.Now(),
		interval: t / time.Duration(n),
	}
}

func NewTBAPerSec(n int64) *Limiter {
	return &Limiter{
		capacity: n,
		initTime: time.Now(),
		interval: time.Second / time.Duration(n),
	}
}

// New Returns instance the simplified version of Token Bucket Algorithm.
type Limiter struct {
	capacity int64
	interval time.Duration
	initTime time.Time

	available     int64
	availableTick int64

	// TODO remove and switch to lock free operation.
	mtx sync.Mutex
}

func (l *Limiter) GetN() int64 {
	return l.capacity
}

// WaitN blocks till n buckets can be consumed
func (l *Limiter) WaitN(n int64) {
	time.Sleep(l.waitDuration(time.Now(), n))
}

func (l *Limiter) waitDuration(t time.Time, n int64) time.Duration {
	if n <= 0 {
		return 0
	}

	l.mtx.Lock()
	defer l.mtx.Unlock()

	currTick := int64(t.Sub(l.initTime) / l.interval)
	l.adj(currTick)

	l.available -= n
	if l.available >= 0 {
		return 0
	}

	return l.initTime.Add(time.Duration(currTick-l.available) * l.interval).Sub(t)
}

// adjust bucket resource based on current tick
func (l *Limiter) adj(currTick int64) {
	if l.available >= l.capacity {
		return
	}
	l.available += currTick - l.availableTick
	if l.available > l.capacity {
		l.available = l.capacity
	}
	l.availableTick = currTick
}

// Swap change the instance limit at runtime.
func (l *Limiter) Reset(n int64) {
	l.mtx.Lock()
	l.mtx.Unlock()
	l.capacity = n
	l.initTime = time.Now()
	l.interval = time.Second / time.Duration(n)
}
