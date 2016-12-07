package buffers

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFillBufferToCap(t *testing.T) {
	b := NewFixedBuffer(bytes.MinRead)
	assert.Equal(t, 0, b.Len())
	assert.Equal(t, bytes.MinRead, b.Cap())

	s := strings.Repeat("F", bytes.MinRead+1)

	n, err := b.ReadFrom(bytes.NewReader([]byte(s)))

	assert.Nil(t, err)
	assert.Equal(t, bytes.MinRead, int(n))
	assert.Equal(t, bytes.MinRead, b.Len())
	assert.Equal(t, bytes.MinRead, b.Cap())
}

type SlowReader struct {
	C int
}

func (s *SlowReader) Read(p []byte) (n int, err error) {
	if s.C == 0 {
		str := strings.Repeat("F", cap(p))
		m := copy(p, []byte(str))
		s.C++
		return m, err
	}
	m := copy(p, []byte("asdf"))
	s.C++
	return m, io.EOF
}

func TestSlowReader(t *testing.T) {
	b := NewFixedBuffer(bytes.MinRead)
	assert.Equal(t, 0, b.Len())
	assert.Equal(t, bytes.MinRead, b.Cap())

	sr := &SlowReader{}
	n, err := b.ReadFrom(sr)
	assert.Equal(t, 1, sr.C)

	assert.Nil(t, err)
	assert.Equal(t, bytes.MinRead, int(n))
	assert.Equal(t, bytes.MinRead, b.Len())
	assert.Equal(t, bytes.MinRead, b.Cap())
}
