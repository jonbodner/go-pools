package buffers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBufferQueue(t *testing.T) {
	item1 := NewFixedBuffer(1)
	item2 := NewFixedBuffer(1)
	q := BufferQueue{}

	assert.Equal(t, 0, q.Len())

	q.Push(item1)
	assert.Equal(t, 1, q.Len())

	q.Push(item2)
	assert.Equal(t, 2, q.Len())

	i1 := q.Pop()
	assert.Equal(t, item1, i1)
	assert.Equal(t, 1, q.Len())

	i2 := q.Pop()
	assert.Equal(t, item2, i2)
	assert.Equal(t, 0, q.Len())

	i3 := q.Pop()
	assert.Nil(t, i3)
	assert.Equal(t, 0, q.Len())
}
