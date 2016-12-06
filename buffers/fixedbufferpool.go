package buffers

import (
	"bytes"
	"sync/atomic"
	"time"
)

type BufferAvailableFunc func(*bytes.Buffer)

type FixedBufferPool struct {
	bufPool  chan *bytes.Buffer
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
	pool := make(chan *bytes.Buffer, maxBufferCount)
	for i := uint32(0); i < maxBufferCount; i++ {
		buf := bytes.NewBuffer(make([]byte, maxBytesPerBuffer))
		buf.Reset()
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

// AsyncCallbackWithBuffer will use a go-routine to callback the given function with a non-nil buffer
// when one becomes available.  Once it the callback is done with the buffer it will be
// automatically put back into the pool for reuse.
func (p *FixedBufferPool) AsyncCallbackWithBuffer(cbFunc BufferAvailableFunc) {
	go func() {
		buf := p.WaitForGet(876000 * time.Hour) // wait for a really long time (100yrs)
		if buf != nil {
			cbFunc(buf)
			p.Put(buf)
		}
	}()
}

func (p *FixedBufferPool) WaitForGet(maxWait time.Duration) *bytes.Buffer {
	duration := 100 * time.Millisecond
	if duration > maxWait {
		duration = maxWait
	}
	start := time.Now()

	var result *bytes.Buffer
ResultCheck:
	for result == nil {
		select {
		case buf, ok := <-p.bufPool:
			result = buf
			if !ok {
				break ResultCheck
			}
		case <-time.After(duration):
		default:
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
