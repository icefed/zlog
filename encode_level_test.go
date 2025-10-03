package zlog

import (
	"bytes"
	"log/slog"
	"slices"
	"testing"

	"github.com/icefed/zlog/buffer"
)

func TestFormatColorLevelValue(t *testing.T) {
	tests := []struct {
		name  string
		level slog.Level
		want  []byte
	}{
		{
			name:  "debug",
			level: slog.LevelDebug,
			want:  []byte("\033[35mDEBUG\033[0m"),
		}, {
			name:  "debug-1",
			level: slog.LevelDebug - 1,
			want:  []byte("\033[35mDEBUG-1\033[0m"),
		}, {
			name:  "debug+2",
			level: slog.LevelDebug + 2,
			want:  []byte("\033[35mDEBUG+2\033[0m"),
		}, {
			name:  "info",
			level: slog.LevelInfo,
			want:  []byte("\033[34mINFO\033[0m"),
		}, {
			name:  "info+5",
			level: slog.LevelInfo + 5,
			want:  []byte("\033[33mWARN+1\033[0m"),
		}, {
			name:  "warn",
			level: slog.LevelWarn,
			want:  []byte("\033[33mWARN\033[0m"),
		}, {
			name:  "error",
			level: slog.LevelError,
			want:  []byte("\033[31mERROR\033[0m"),
		}, {
			name:  "error+100",
			level: slog.LevelError + 100,
			want:  []byte("\033[31mERROR+100\033[0m"),
		},
	}

	buf := buffer.New()
	defer buf.Free()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			formatColorLevelValue(buf, test.level)
			if !bytes.Equal(buf.Bytes(), test.want) {
				t.Errorf("got %v, want %v", string(buf.Bytes()), string(test.want))
			}
			buf.Reset()
		})
	}
}

func TestSetLevelColor(t *testing.T) {
	tests := []struct {
		name string
		val  slog.Level
		// The order we expect in output
		want []slog.Level
	}{
		{
			name: "above error",
			val:  slog.LevelError + 100,
			want: []slog.Level{
				slog.LevelError + 100,
				slog.LevelError,
				slog.LevelWarn,
				slog.LevelInfo,
				slog.LevelDebug,
			},
		},
		{
			name: "replace error",
			val:  slog.LevelError,
			want: []slog.Level{
				slog.LevelError,
				slog.LevelWarn,
				slog.LevelInfo,
				slog.LevelDebug,
			},
		},
		{
			name: "in middle",
			val:  slog.LevelError - 2,
			want: []slog.Level{
				slog.LevelError,
				slog.LevelError - 2,
				slog.LevelWarn,
				slog.LevelInfo,
				slog.LevelDebug,
			},
		},
		{
			name: "below debug",
			val:  slog.LevelDebug - 4,
			want: []slog.Level{
				slog.LevelError,
				slog.LevelWarn,
				slog.LevelInfo,
				slog.LevelDebug,
				slog.LevelDebug - 4,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			SetLevelColor(test.val, "<newcolor>")

			var got []slog.Level
			for _, mode := range levelColorList {
				got = append(got, mode.Level)
			}
			if !slices.Equal(got, test.want) {
				t.Errorf("got %v, want %v", got, test.want)
			}
			idx := slices.IndexFunc(levelColorList, func(mode lvlEscape) bool { return mode.Level == test.val })
			if levelColorList[idx].string != "<newcolor>" {
				t.Errorf("Expected to find new value at position %d, got %#v", idx, levelColorList[idx])
			}

			UseDefaultLevelColors()
		})
	}
}
