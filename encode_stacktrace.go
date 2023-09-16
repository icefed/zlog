package zlog

import (
	"runtime"
	"strconv"
	"sync"

	"github.com/icefed/zlog/buffer"
)

// stacktrace define the stack trace source.
// For ReplaceAttr, *stacktrace is the type of value in slog.Attr.
type stacktrace struct {
	pc uintptr
}

type stacktracePCs []uintptr

var stacktracePCsPool = sync.Pool{
	New: func() any {
		pcs := make([]uintptr, 64)
		return (*stacktracePCs)(&pcs)
	},
}

func newStacktracePCs() *stacktracePCs {
	return stacktracePCsPool.Get().(*stacktracePCs)
}

func (st *stacktracePCs) Free() {
	stacktracePCsPool.Put(st)
}

func formatStacktrace(buf *buffer.Buffer, sourcepc uintptr) {
	sfs := runtime.CallersFrames([]uintptr{sourcepc})
	sf, _ := sfs.Next()

	writeCaller := func(fun, file string, line int, more bool) {
		buf.WriteString(fun)
		buf.Write([]byte("\n\t"))
		buf.WriteString(file)
		buf.WriteByte(':')
		*buf = strconv.AppendInt(*buf, int64(line), 10)
		if more {
			buf.WriteByte('\n')
		}
	}

	pcs := *newStacktracePCs()
	defer pcs.Free()
	n := runtime.Callers(1, pcs)
	more := n > 0

	fs := runtime.CallersFrames(pcs[:n])
	var f runtime.Frame
	found := false
	for more {
		f, more = fs.Next()
		if found {
			writeCaller(f.Function, f.File, f.Line, more)
			continue
		}
		if f.Function == sf.Function && f.File == sf.File && f.Line == sf.Line {
			writeCaller(f.Function, f.File, f.Line, more)
			found = true
		}
	}

	if !found {
		// write source
		writeCaller(sf.Function, sf.File, sf.Line, more)
	}
}
