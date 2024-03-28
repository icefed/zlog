package zlog

import (
	"log/slog"

	"github.com/icefed/zlog/buffer"
)

const (
	// Color codes for terminal output.
	black   = "\033[30m"
	red     = "\033[31m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	blue    = "\033[34m"
	magenta = "\033[35m"
	cyan    = "\033[36m"
	white   = "\033[37m"
	reset   = "\033[0m"
)

// formatColorLevelValue returns the string representation of the level.
func formatColorLevelValue(buf *buffer.Buffer, l slog.Level) {
	switch {
	case l < slog.LevelInfo: // LevelDebug
		buf.WriteString(magenta)
	case l < slog.LevelWarn: // LevelInfo
		buf.WriteString(blue)
	case l < slog.LevelError: // LevelWarn
		buf.WriteString(yellow)
	default: // LevelError
		buf.WriteString(red)
	}
	buf.WriteString(l.String())
	buf.WriteString(reset)
}
