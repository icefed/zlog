package zlog

import (
	"runtime"
	"strconv"
	"strings"

	"github.com/icefed/zlog/buffer"
	"golang.org/x/exp/slog"
)

func buildSource(pc uintptr) *slog.Source {
	fs := runtime.CallersFrames([]uintptr{pc})
	f, _ := fs.Next()
	return &slog.Source{
		Function: f.Function,
		File:     f.File,
		Line:     f.Line,
	}
}

func formatSourceValueFromPC(buf *buffer.Buffer, pc uintptr) {
	fs := runtime.CallersFrames([]uintptr{pc})
	f, _ := fs.Next()

	defer func() {
		buf.WriteByte(':')
		*buf = strconv.AppendInt(*buf, int64(f.Line), 10)
	}()

	i := strings.LastIndexByte(f.File, '/')
	if i < 0 {
		buf.WriteString(f.File)
		return
	}
	i = strings.LastIndexByte(f.File[:i], '/')
	if i < 0 {
		buf.WriteString(f.File)
		return
	}
	buf.WriteString(f.File[i+1:])
}

func formatSourceValue(buf *buffer.Buffer, s *slog.Source) {
	defer func() {
		buf.WriteByte(':')
		*buf = strconv.AppendInt(*buf, int64(s.Line), 10)
	}()

	i := strings.LastIndexByte(s.File, '/')
	if i < 0 {
		buf.WriteString(s.File)
		return
	}
	i = strings.LastIndexByte(s.File[:i], '/')
	if i < 0 {
		buf.WriteString(s.File)
		return
	}
	buf.WriteString(s.File[i+1:])
}
