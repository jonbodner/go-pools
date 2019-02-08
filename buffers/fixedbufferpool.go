package buffers

import (
	"context"
	"sync/atomic"
	"time"
)

type BufferAvailableFunc func(*FixedBuffer)

type FixedBufferPool struct {
	bufPool  chan *FixedBuffer
	gets     uint64
	puts     uint64
	timeouts uint64
	count    uint32
}

const (
	DefaultBufferMax   = 4096
	DefaultBufferCount = 5
)

func NewDefaultFixedBufferPool() *FixedBufferPool {
	return NewFixedBufferPool(DefaultBufferMax, DefaultBufferCount)
}

func NewFixedBufferPool(maxBytesPerBuffer uint64, maxBufferCount uint32) *FixedBufferPool {
	// use a buffered channel as our pool of available byte buffers
	// it's tested to be faster than a list with mutexes
	pool := make(chan *FixedBuffer, maxBufferCount)
	for i := uint32(0); i < maxBufferCount; i++ {
		buf := NewFixedBuffer(maxBytesPerBuffer)
		pool <- buf
	}
	return &FixedBufferPool{bufPool: pool, count: maxBufferCount}
}

func (p *FixedBufferPool) Len() int {
	return int(p.count)
}

func (p *FixedBufferPool) Close() error {
	close(p.bufPool)
	return nil
}

// Get selects an arbitrary buffer from the Pool
// It may return nil if none is available.
func (p *FixedBufferPool) Get() *FixedBuffer {
	var buf *FixedBuffer
	select {
	case buf = <-p.bufPool:
	default:
	}

	if buf != nil {
		atomic.AddUint64(&p.gets, 1)
	}
	return buf
}

func (p *FixedBufferPool) AsyncCallbackWithBuffer(cbFunc BufferAvailableFunc) {
	p.AsyncCallbackWithBufferContext(context.Background(), cbFunc)
}

// AsyncCallbackWithBuffer will use a go-routine to callback the given function with a non-nil buffer
// when one becomes available.  Once it the callback is done with the buffer it will be
// automatically put back into the pool for reuse.
func (p *FixedBufferPool) AsyncCallbackWithBufferContext(ctx context.Context, cbFunc BufferAvailableFunc) {
	go func() {
		buf := p.WaitForGetContext(ctx, 876000*time.Hour) // wait for a really long time (100yrs)
		if buf != nil {
			defer p.Put(buf)
			cbFunc(buf)
		}
	}()
}

func (p *FixedBufferPool) WaitForGet(maxWait time.Duration) *FixedBuffer {
	return p.WaitForGetContext(context.Background(), maxWait)
}
func (p *FixedBufferPool) WaitForGetContext(ctx context.Context, maxWait time.Duration) *FixedBuffer {
	remaining := maxWait

	for remaining > 0 {
		duration := 100 * time.Millisecond
		if duration > remaining {
			duration = remaining
		}
		select {
		case buf, ok := <-p.bufPool:
			if ok {
				return buf
			}
		case <-time.After(duration):
			atomic.AddUint64(&p.timeouts, 1)
		case <-ctx.Done():
			return nil
		}
		remaining -= duration
	}

	return nil
}

func (p *FixedBufferPool) Put(b *FixedBuffer) {
	b.Reset()
	p.bufPool <- b

	atomic.AddUint64(&p.puts, 1)
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
