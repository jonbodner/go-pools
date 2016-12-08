package buffers

type BufferQueue []*FixedBuffer

func (q *BufferQueue) Len() int {
	return len(*q)
}

func (q *BufferQueue) Push(b *FixedBuffer) {
	b.Reset()
	*q = append(*q, b)
}

func (q *BufferQueue) Pop() *FixedBuffer {
	qlen := len(*q)
	if qlen == 0 {
		return nil
	}
	var result *FixedBuffer
	result, *q = (*q)[qlen-1], (*q)[:qlen-1]
	return result
}
