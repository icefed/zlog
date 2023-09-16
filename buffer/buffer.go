package buffer

import (
	"sync"
	"unicode/utf8"
	"unsafe"
)

// buffer size
const (
	smallBufferSize = 512
	maxBufferSize   = 4 << 10
)

// growBuf is a buffer for growing
var growBuf = make([]byte, smallBufferSize)

// Buffer is a single bytes buffer
type Buffer []byte

// bufferPool for reusing buffers
var bufferPool = sync.Pool{
	New: func() any {
		buf := make([]byte, 0, smallBufferSize)
		return (*Buffer)(&buf)
	},
}

// New returns a new buffer
func New() *Buffer {
	return bufferPool.Get().(*Buffer)
}

func (b *Buffer) Free() {
	if cap(*b) <= maxBufferSize {
		*b = (*b)[:0]
		bufferPool.Put(b)
	}
}

func (b *Buffer) Reset() {
	*b = (*b)[:0]
}

func (b *Buffer) Len() int {
	return len(*b)
}

func (b *Buffer) LastByte() *byte {
	n := len(*b)
	if n <= 0 {
		return nil
	}
	return &(*b)[n-1]
}

// Truncate buffer to n bytes
// If n is negative, then do nothing
func (b *Buffer) Truncate(n int) {
	if n < 0 {
		return
	}
	if n > len(*b) {
		b.Grow(n - len(*b))
	}
	*b = (*b)[:n]
}

// Grow buffer with n bytes
// If n is negative, then do nothing
func (b *Buffer) Grow(n int) {
	if n < 0 {
		return
	}
	l := len(*b)
	m := cap(*b)
	if m-l >= n {
		*b = (*b)[:l+n]
		return
	}

	if n < smallBufferSize {
		*b = append(*b, growBuf...)
		*b = (*b)[:l+n]
		return
	}

	c := l + n
	if c < 2*m {
		c = 2 * m
	}
	*b = append(*b, make([]byte, c-l)...)
	*b = (*b)[:l+n]
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

func (b *Buffer) String() string {
	return unsafe.String(unsafe.SliceData(*b), len(*b))
}
