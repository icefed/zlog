package zlog

import (
	"bytes"
	"testing"

	"github.com/icefed/zlog/buffer"
	"golang.org/x/exp/slog"
)

func TestFormatColorLevelValue(t *testing.T) {
	tests := []struct {
		level slog.Level
		want  []byte
	}{
		{
			level: slog.LevelDebug,
			want:  []byte("\033[35mDEBUG\033[0m"),
		}, {
			level: slog.LevelDebug - 1,
			want:  []byte("\033[35mDEBUG-1\033[0m"),
		}, {
			level: slog.LevelDebug + 2,
			want:  []byte("\033[35mDEBUG+2\033[0m"),
		}, {
			level: slog.LevelInfo,
			want:  []byte("\033[34mINFO\033[0m"),
		}, {
			level: slog.LevelInfo + 5,
			want:  []byte("\033[33mWARN+1\033[0m"),
		}, {
			level: slog.LevelWarn,
			want:  []byte("\033[33mWARN\033[0m"),
		}, {
			level: slog.LevelError,
			want:  []byte("\033[31mERROR\033[0m"),
		}, {
			level: slog.LevelError + 100,
			want:  []byte("\033[31mERROR+100\033[0m"),
		},
	}

	buf := buffer.New()
	defer buf.Free()

	for _, test := range tests {
		formatColorLevelValue(buf, test.level)
		if !bytes.Equal(buf.Bytes(), test.want) {
			t.Errorf("got %v, want %v", string(buf.Bytes()), string(test.want))
		}
		buf.Reset()
	}
}
