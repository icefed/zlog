package zlog

import (
	"io/fs"
	"log/slog"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/icefed/zlog/buffer"
)

var testTime = time.Date(2023, 8, 16, 1, 2, 3, 666666666, time.UTC)

func TestTextEncoder(t *testing.T) {
	t.Run("no_replace", func(t *testing.T) {
		h := NewJSONHandler(&Config{
			HandlerOptions: slog.HandlerOptions{
				AddSource: true,
				Level:     slog.LevelDebug,
			},
		})
		buf := buffer.New()
		defer buf.Free()
		enc := newTextEncoder(h, buf)

		tests := []struct {
			name  string
			key   string
			value any
			want  string
		}{
			{
				name:  "time",
				key:   slog.TimeKey,
				value: testTime,
				want:  "2023-08-16T01:02:03.666Z",
			}, {
				name:  "level",
				key:   slog.LevelKey,
				value: slog.LevelInfo,
				want:  "INFO",
			}, {
				name:  "msg",
				key:   slog.MessageKey,
				value: "test msg",
				want:  "test msg",
			}, {
				name:  "error",
				key:   "error",
				value: fs.ErrNotExist,
				want:  "file does not exist",
			}, {
				name: "source",
				key:  slog.SourceKey,
				value: &slog.Source{
					File: "test.go",
					Line: 300,
				},
				want: "test.go:300",
			}, {
				name:  "stacktrace",
				key:   "stacktrace",
				value: &stacktrace{getPC()},
				want:  wantPCFunction + "\n\t" + wantPCFile + ":" + strconv.Itoa(wantPCLine),
			}, {
				name:  "ip",
				key:   "ip",
				value: net.ParseIP("127.0.0.1"),
				want:  "127.0.0.1",
			},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				enc.Append(test.key, test.value)
				if string(buf.Bytes()) != test.want {
					t.Errorf("got %v, want %v", string(buf.Bytes()), test.want)
				}
				buf.Reset()
			})
		}
	})
	t.Run("replace", func(t *testing.T) {
		h := NewJSONHandler(&Config{
			HandlerOptions: slog.HandlerOptions{
				AddSource: true,
				Level:     slog.LevelDebug,
			},
		})
		buf := buffer.New()
		defer buf.Free()

		tests := []struct {
			name        string
			key         string
			value       any
			want        string
			replaceAttr func(_ []string, a slog.Attr) slog.Attr
		}{
			{
				name:  "ignore level",
				key:   slog.LevelKey,
				value: slog.LevelDebug,
				want:  "",
				replaceAttr: func(_ []string, a slog.Attr) slog.Attr {
					if a.Key == slog.LevelKey && a.Value.Any().(slog.Level) == slog.LevelDebug {
						return slog.Attr{}
					}
					return a
				},
			}, {
				name:  "replace level to int",
				key:   slog.LevelKey,
				value: slog.LevelError,
				want:  "20",
				replaceAttr: func(_ []string, a slog.Attr) slog.Attr {
					if a.Key == slog.LevelKey && a.Value.Any().(slog.Level) == slog.LevelError {
						a.Value = slog.AnyValue(20)
						return a
					}
					return a
				},
			}, {
				name:  "replace pc to string",
				key:   "replacepc",
				value: getPC(),
				want:  `replacedpc`,
				replaceAttr: func(_ []string, a slog.Attr) slog.Attr {
					if a.Key == "replacepc" {
						a.Value = slog.AnyValue("replacedpc")
					}
					return a
				},
			},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				if test.replaceAttr != nil {
					h = h.WithOptions(WithReplaceAttr(test.replaceAttr))
				}
				enc := newTextEncoder(h, buf)
				enc.Append(test.key, test.value)
				if string(buf.Bytes()) != test.want {
					t.Errorf("got %v, want %v", string(buf.Bytes()), test.want)
				}
				buf.Reset()
			})
		}
	})
}
