package buffers

import (
	"bytes"
	"errors"
	"io"
)

var ErrAtCapacity = errors.New("buffers.FixedBuffer: at capacity")

type FixedBuffer struct {
	b   []byte
	off int
}

func NewFixedBuffer(size uint64) *FixedBuffer {
	f := &FixedBuffer{b: make([]byte, size)}
	f.Reset()
	return f
}

func (f *FixedBuffer) NewReader() *bytes.Reader {
	return bytes.NewReader(f.b[f.off:])
}

// ReadFrom reads data from r until EOF or capacity of buffer and appends it to the buffer, but does not grow the buffer.
// See: https://golang.org/pkg/io/#ReaderFrom
func (f *FixedBuffer) ReadFrom(input io.Reader) (n int64, err error) {
	for err == nil {
		l := len(f.b)

		if cap(f.b)-l == 0 {
			break
		}

		m, err := input.Read(f.b[l:cap(f.b)])
		n += int64(m)
		f.b = f.b[0 : l+m]
		if err == io.EOF {
			return n, nil
		}
	}
	return n, err
}

func (f *FixedBuffer) Len() int {
	return len(f.b)
}

func (f *FixedBuffer) Cap() int {
	return cap(f.b)
}

func (f *FixedBuffer) Reset() {
	f.off = 0
	f.b = f.b[0:0]
}

func (f *FixedBuffer) WriteTo(w io.Writer) (n int64, err error) {
	l := len(f.b[f.off:])
	m, err := w.Write(f.b[f.off:])
	f.off += m
	n = int64(m)
	if err != nil {
		return n, err
	}
	if m != l {
		return n, io.ErrShortWrite
	}
	f.Reset()
	return n, err
}

func (f *FixedBuffer) WriteString(s string) (n int, err error) {
	off := f.off + len(f.b)
	l := cap(f.b) - off
	n = copy(f.b[off:l], s)
	f.b = f.b[f.off : off+n]
	return n, err
}

func (f *FixedBuffer) Write(p []byte) (n int, err error) {
	off := f.off + len(f.b)
	l := cap(f.b) - off
	n = copy(f.b[off:l], p)
	f.b = f.b[f.off : off+n]
	return n, err
}
