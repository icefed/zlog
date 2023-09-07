package buffer

import (
	"sync"
	"unicode/utf8"
)

// Initial buffer size
const initBufferSize = 512

// Buffer is a single bytes buffer
type Buffer []byte

// bufferPool for reusing buffers
var bufferPool = sync.Pool{
	New: func() any {
		buf := make([]byte, 0, initBufferSize)
		return (*Buffer)(&buf)
	},
}

// New returns a new buffer
func New() *Buffer {
	return bufferPool.Get().(*Buffer)
}

func (b *Buffer) Free() {
	b.Reset()
	bufferPool.Put(b)
}

func (b *Buffer) Reset() {
	*b = (*b)[:0]
}

func (b *Buffer) Len() int {
	return len(*b)
}

func (b *Buffer) LastByte() *byte {
	n := b.Len()
	if n <= 0 {
		return nil
	}
	return &(*b)[n-1]
}

func (b *Buffer) Truncate(n int) {
	if n > len(*b) {
		return
	}
	*b = (*b)[:n]
}

func (b *Buffer) Write(data []byte) {
	*b = append(*b, data...)
}

func (b *Buffer) WriteString(str string) {
	*b = append(*b, str...)
}

func (b *Buffer) WriteByte(c byte) {
	*b = append(*b, c)
}

func (b *Buffer) WriteRune(r rune) {
	*b = utf8.AppendRune(*b, r)
}

func (b *Buffer) Bytes() []byte {
	return *b
}
