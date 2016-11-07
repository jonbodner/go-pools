package buffers

import (
	"bytes"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBufferPool(t *testing.T) {
	p := NewFixedBufferPool(5, 1)
	assert.NotNil(t, p)

	b := p.Get()
	assert.NotNil(t, b)
	defer p.Put(b)

	b2 := p.Get()
	assert.Nil(t, b2)

	assert.True(t, b.Cap() == 5)

	assert.Equal(t, uint64(1), p.GetCount())
	assert.Equal(t, uint64(0), p.PutCount())

	p.ResetCounters()
	assert.Equal(t, uint64(0), p.GetCount())
	assert.Equal(t, uint64(0), p.PutCount())
}

func TestBufferReset(t *testing.T) {
	p := NewFixedBufferPool(5, 1)

	b := p.Get()
	b.WriteString("testing")
	p.Put(b)

	b2 := p.Get()
	assert.Equal(t, 0, b2.Len())
	p.Close()
}

func TestBufferGetWaiter(t *testing.T) {
	p := NewFixedBufferPool(5, 1)
	b := p.WaitForGet(20 * time.Second)
	assert.NotNil(t, b)
	var wg1, wg2 sync.WaitGroup

	var b2 *bytes.Buffer
	wg1.Add(1)
	wg2.Add(1)
	go func() {
		wg1.Done()
		b2 = p.WaitForGet(20 * time.Second)
		assert.NotNil(t, b2)
		wg2.Done()
	}()

	wg1.Wait()
	assert.Nil(t, b2)
	time.Sleep(1 * time.Second)
	p.Put(b)
	wg2.Wait()
	assert.NotNil(t, b2)
}

func TestBufferGetWaiterTimeout(t *testing.T) {
	p := NewFixedBufferPool(5, 1)
	b := p.WaitForGet(10 * time.Millisecond)
	assert.NotNil(t, b)
	var wg1, wg2 sync.WaitGroup

	var b2 *bytes.Buffer
	wg1.Add(1)
	wg2.Add(1)
	go func() {
		wg1.Done()
		b2 = p.WaitForGet(10 * time.Millisecond)
		assert.Nil(t, b2)
		wg2.Done()
	}()

	wg1.Wait()
	assert.Nil(t, b2)
	time.Sleep(1 * time.Second)
	p.Put(b)
	wg2.Wait()
	assert.Nil(t, b2)
}

func TestBufferCallback(t *testing.T) {
	p := NewFixedBufferPool(5, 1)
	defer p.Close()

	var wg1 sync.WaitGroup
	var b2 *bytes.Buffer

	wg1.Add(1)
	b := p.Get()
	assert.NotNil(t, b)

	cb := func(buf *bytes.Buffer) {
		b2 = buf
		wg1.Done()
	}

	p.AsyncCallbackWithBuffer(cb)
	time.Sleep(1 * time.Second)
	p.Put(b)
	wg1.Wait()
	assert.NotNil(t, b2)

	b3 := p.Get()
	assert.NotNil(t, b3)
}

func TestBufferCloseWhileWaiting(t *testing.T) {
	p := NewFixedBufferPool(5, 1)
	b := p.Get()
	assert.NotNil(t, b)

	var wg1 sync.WaitGroup
	var b2 *bytes.Buffer

	wg1.Add(1)
	cb := func(buf *bytes.Buffer) {
		b2 = buf
		wg1.Done()
	}

	p.AsyncCallbackWithBuffer(cb)
	time.Sleep(1 * time.Second)
	p.Close()
	time.Sleep(1 * time.Second)

	b3 := p.Get()
	assert.Nil(t, b3)
}
