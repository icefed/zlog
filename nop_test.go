package zlog

import (
	"context"
	"log/slog"
	"testing"
)

func TestNopHandler(t *testing.T) {
	nopHandler := NewNopHandler()
	nopHandler.Enabled(context.Background(), slog.LevelError)
	nopHandler.Handle(nil, slog.Record{})
	nopHandler.WithAttrs(nil)
	nopHandler.WithGroup("group")

	nop := NewNopSlogger()
	// no output
	nop.Info("This is a test")
	nop.With("key", "value").Info("This is a test")
	nop.DebugContext(context.Background(), "This is a test")
}

func TestNopLogger(t *testing.T) {
	log := NewNopLogger().WithOptions(WithAddSource(true))

	log.With("key", "value").Log(context.Background(), slog.LevelDebug, "debug", "key", "value")
	log.WithGroup("group").Log(context.Background(), slog.LevelDebug, "debug", "key", "value")
	log.Log(context.Background(), slog.LevelDebug, "debug", "key", "value")
	log.LogAttrs(context.Background(), slog.LevelDebug, "debug", slog.String("key", "value"))
	log.Debug("debug", "key", "value")
	log.Debugf("debugf: %s: %d", "key", 1)
	log.DebugContext(context.Background(), "debug", "key", "value")
	log.DebugContextf(context.Background(), "debugf: %s: %d", "key", 1)
	log.Error("error", "key", "value")
	log.Errorf("errorf: %s: %d", "key", 1)
	log.ErrorContext(context.Background(), "error", "key", "value")
	log.ErrorContextf(context.Background(), "errorf: %s: %d", "key", 1)
	log.Info("info", "key", "value")
	log.Infof("infof: %s: %d", "key", 1)
	log.InfoContext(context.Background(), "info", "key", "value")
	log.InfoContextf(context.Background(), "infof: %s: %d", "key", 1)
	log.Warn("warn", "key", "value")
	log.Warnf("warnf: %s: %d", "key", 1)
	log.WarnContext(context.Background(), "warn", "key", "value")
	log.WarnContextf(context.Background(), "warnf: %s: %d", "key", 1)
}
