package buffers

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteToBuffer(t *testing.T) {
	b := NewFixedBuffer(bytes.MinRead)
	str := "testing"
	n, err := b.Write([]byte(str))
	assert.Nil(t, err)
	assert.Equal(t, n, len(str))
	assert.Equal(t, str, string(b.b))

	str2 := " 1 2 3"
	n, err = b.Write([]byte(str2))
	assert.Nil(t, err)
	assert.Equal(t, len(str2), n)
	assert.Equal(t, len(str)+len(str2), b.Len())
	assert.Equal(t, str+str2, string(b.b))

	out := bytes.NewBuffer(make([]byte, 100))
	b.WriteTo(out)
	assert.Equal(t, b.Len(), 0)

	str3 := "again"
	n, err = b.Write([]byte(str3))
	assert.Nil(t, err)
	assert.Equal(t, n, len(str3))
	assert.Equal(t, str3, string(b.b))
}

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
