package zlog

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"golang.org/x/exp/slog"
)

func TestLogger(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	l := New(NewJSONHandler(&Config{
		HandlerOptions: slog.HandlerOptions{
			Level: slog.LevelDebug,
			ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey {
					a.Key = ""
				}
				return a
			},
		},
		Writer: buf,
	}))

	check := func(expected string) {
		t.Helper()
		got := buf.String()
		// Remove the trailing newline
		got = got[:len(got)-1]
		if got != expected {
			t.Errorf("got %q, want %q", got, expected)
		}
		buf.Reset()
	}

	// WithCallerSkip
	testWithCallerSkip(l, check, 1, "zlog/logger_test.go:40")

	SetDefault(l)

	l.WithOptions(WithAddSource(true)).Log(context.Background(), slog.LevelDebug, "debug", "key", "value")
	check(`{"level":"DEBUG","source":"zlog/logger_test.go:44","msg":"debug","key":"value"}`)
	l.WithOptions(WithAddSource(true)).LogAttrs(context.Background(), slog.LevelDebug, "debug", slog.String("key", "value"))
	check(`{"level":"DEBUG","source":"zlog/logger_test.go:46","msg":"debug","key":"value"}`)

	WithGroup("g").With("app", "test").Log(context.Background(), slog.LevelDebug, "debug", "key", "value")
	check(`{"level":"DEBUG","msg":"debug","g":{"app":"test","key":"value"}}`)
	With(slog.String("app", "test"), "badkey").LogAttrs(context.Background(), slog.LevelDebug, "debug", slog.String("key", "value"))
	check(`{"level":"DEBUG","msg":"debug","app":"test","!BADKEY":"badkey","key":"value"}`)
	With(1).Log(context.Background(), slog.LevelDebug, "debug", "key", "value")
	check(`{"level":"DEBUG","msg":"debug","!BADKEY":1,"key":"value"}`)
	Log(nil, slog.LevelDebug, "debug", "key", "value")
	check(`{"level":"DEBUG","msg":"debug","key":"value"}`)
	LogAttrs(nil, slog.LevelDebug, "debug", slog.String("key", "value"))
	check(`{"level":"DEBUG","msg":"debug","key":"value"}`)

	Debug("debug", "key", "value")
	check(`{"level":"DEBUG","msg":"debug","key":"value"}`)
	Debugf("debugf: %s: %d", "key", 1)
	check(`{"level":"DEBUG","msg":"debugf: key: 1"}`)
	DebugContext(context.Background(), "debug", "key", "value")
	check(`{"level":"DEBUG","msg":"debug","key":"value"}`)
	DebugContextf(context.Background(), "debugf: %s: %d", "key", 1)
	check(`{"level":"DEBUG","msg":"debugf: key: 1"}`)

	Error("read file failed", "err", "file not found")
	check(`{"level":"ERROR","msg":"read file failed","err":"file not found"}`)
	Errorf("read file failed: %v", fmt.Errorf("file not found"))
	check(`{"level":"ERROR","msg":"read file failed: file not found"}`)
	ErrorContext(context.Background(), "read file failed", "err", "file not found")
	check(`{"level":"ERROR","msg":"read file failed","err":"file not found"}`)
	ErrorContextf(context.Background(), "read file failed: %v", fmt.Errorf("file not found"))
	check(`{"level":"ERROR","msg":"read file failed: file not found"}`)

	Info("info", "progress", 95)
	check(`{"level":"INFO","msg":"info","progress":95}`)
	Infof("infof: %s: %d", "progress", 95)
	check(`{"level":"INFO","msg":"infof: progress: 95"}`)
	InfoContext(context.Background(), "info", "progress", 95)
	check(`{"level":"INFO","msg":"info","progress":95}`)
	InfoContextf(context.Background(), "infof: %s: %d", "progress", 95)
	check(`{"level":"INFO","msg":"infof: progress: 95"}`)

	Warn("request limit exceeded", "duration", 500*time.Millisecond)
	check(`{"level":"WARN","msg":"request limit exceeded","duration":500000000}`)
	Warnf("request limit exceeded: %s: %d", "duration", 500*time.Millisecond)
	check(`{"level":"WARN","msg":"request limit exceeded: duration: 500000000"}`)
	WarnContext(context.Background(), "request limit exceeded", "duration", 500*time.Millisecond)
	check(`{"level":"WARN","msg":"request limit exceeded","duration":500000000}`)
	WarnContextf(context.Background(), "request limit exceeded: %s: %d", "duration", 500*time.Millisecond)
	check(`{"level":"WARN","msg":"request limit exceeded: duration: 500000000"}`)
}

func testWithCallerSkip(log *Logger, check func(expected string), skip int, expectedSource string) {
	log.WithOptions(WithAddSource(true)).WithCallerSkip(1).
		Log(context.Background(), slog.LevelDebug, "debug", "key", "value")
	check(`{"level":"DEBUG","source":"` + expectedSource + `","msg":"debug","key":"value"}`)
}
