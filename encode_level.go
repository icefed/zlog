package zlog

import (
	"log/slog"
	"slices"

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

type lvlEscape struct {
	slog.Level
	string
}

var levelColorList []lvlEscape

func init() {
	UseDefaultLevelColors()
}

// formatColorLevelValue returns the string representation of the level.
func formatColorLevelValue(buf *buffer.Buffer, l slog.Level) {
	var mode lvlEscape
	for _, mode = range levelColorList {
		if l >= mode.Level {
			break
		}
	}
	buf.WriteString(mode.string)
	buf.WriteString(l.String())
	buf.WriteString(reset)
}

// UseDefaultLevelColors resets  the colors levels to the default configuration
func UseDefaultLevelColors() {
	levelColorList = []lvlEscape{
		{slog.LevelError, red},
		{slog.LevelWarn, yellow},
		{slog.LevelInfo, blue},
		{slog.LevelDebug, magenta},
	}
}

// SetLevelColor adds (or overwrites) the global level color used in Development mode for a given logging
// level, or anything above it and up to the next level.
func SetLevelColor(l slog.Level, escape string) {
	// We need to maintain the `levelColorList` in reverse order so the search works.

	// And to avoid the need for lock etc (just in case someone calls this when logging calls might be made) we
	// replace levelColorList whole sale with a new slice than edit it in place. That way either either get the
	// whole new slice or the whole old slice.

	newMode := lvlEscape{l, escape}
	newList := slices.Clone(levelColorList)

	targetPos := slices.IndexFunc(newList, func(mode lvlEscape) bool {
		return l >= mode.Level
	})

	if targetPos == -1 {
		// Position not found, stick it on the end
		newList = append(newList, newMode)
	} else if newList[targetPos].Level == l {
		// Replace the current one
		newList[targetPos] = newMode
	} else {
		newList = slices.Insert(newList, targetPos, newMode)
	}

	levelColorList = newList
}
