package zlog

import (
	"bytes"
	"runtime"
	"sync"
	"testing"

	"github.com/icefed/zlog/buffer"
)

var (
	wantPCFunction string
	wantPCFile     string
	wantPCLine     int

	getPC = sync.OnceValue(func() uintptr {
		pcs := make([]uintptr, 1)
		runtime.Callers(1, pcs)

		sfs := runtime.CallersFrames(pcs)
		sf, _ := sfs.Next()
		wantPCFunction = sf.Function
		wantPCFile = sf.File
		wantPCLine = sf.Line
		return pcs[0]
	})
)

func stacktraceCaller2(buf *buffer.Buffer) {
	pcs := make([]uintptr, 1)
	runtime.Callers(2, pcs)
	formatStacktrace(buf, pcs[0])
}

func stacktraceCaller1(buf *buffer.Buffer) {
	stacktraceCaller2(buf)
}

func TestFormatStacktrace(t *testing.T) {
	buf := buffer.New()
	defer buf.Free()

	stacktraceCaller1(buf)
	lines := bytes.Split(buf.Bytes(), []byte{'\n'})
	if len(lines) != 8 {
		t.Errorf("got %v, want %v", len(lines), 8)
	}
	buf.Reset()

	// not found
	pcs := make([]uintptr, 1)
	runtime.Callers(1, pcs)
	formatStacktrace(buf, pcs[0])
	lines = bytes.Split(buf.Bytes(), []byte{'\n'})
	if len(lines) != 2 {
		t.Errorf("got %v, want %v", len(lines), 2)
	}
}
