package buffers2

import (
	"bytes"
	"sync/atomic"
	"time"
)

type FixedBufferPool struct {
	bufPool  chan *bytes.Buffer
	gets     uint64
	puts     uint64
	timeouts uint64
	waiter   chan struct{}
}

const (
	DefaultBufferMax   = 4096
	DefaultBufferCount = 5
)

func NewDefaultFixedBufferPool() *FixedBufferPool {
	return NewFixedBufferPool(DefaultBufferMax, DefaultBufferCount)
}

func NewFixedBufferPool(maxBytesPerBuffer uint64, maxBufferCount uint32) *FixedBufferPool {
	pool := make(chan *bytes.Buffer, maxBufferCount)
	for i := uint32(0); i < maxBufferCount; i++ {
		pool <- bytes.NewBuffer(make([]byte, maxBytesPerBuffer))
	}
	return &FixedBufferPool{bufPool: pool, waiter: make(chan struct{}, maxBufferCount)}
}

func (p *FixedBufferPool) Close() error {
	close(p.waiter)
	close(p.bufPool)
	return nil
}

// Get selects an arbitrary buffer from the Pool
// It may return nil if none is available.
func (p *FixedBufferPool) Get() *bytes.Buffer {
	var buf *bytes.Buffer
	select {
	case buf = <-p.bufPool:
	default:
	}

	if buf != nil {
		atomic.AddUint64(&p.gets, 1)
	}
	return buf
}

func (p *FixedBufferPool) WaitForGet(maxWait time.Duration) *bytes.Buffer {
	result := p.Get()

	if result != nil {
		return result
	}

	duration := 100 * time.Millisecond
	if duration > maxWait {
		duration = maxWait
	}
	start := time.Now()
	for result == nil {
		select {
		case <-p.waiter:
			// signaled
			result = p.Get()
		case <-time.After(duration):
			result = p.Get()
		default:
			result = p.Get()
		}

		// check max time wait
		if result == nil && time.Since(start) > maxWait {
			atomic.AddUint64(&p.timeouts, 1)
			return result
		}
	}

	return result
}

func (p *FixedBufferPool) Put(b *bytes.Buffer) {
	b.Reset()

	p.bufPool <- b

	atomic.AddUint64(&p.puts, 1)
	if len(p.waiter) != cap(p.waiter) {
		p.waiter <- struct{}{}
	}
}

func (p *FixedBufferPool) GetCount() uint64 {
	return atomic.LoadUint64(&p.gets)
}

func (p *FixedBufferPool) PutCount() uint64 {
	return atomic.LoadUint64(&p.puts)
}

func (p *FixedBufferPool) ResetCounters() {
	atomic.StoreUint64(&p.gets, 0)
	atomic.StoreUint64(&p.puts, 0)
	atomic.StoreUint64(&p.timeouts, 0)
}
