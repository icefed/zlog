package zlog

import (
	"context"

	"golang.org/x/exp/slog"
)

// NewNopHandler returns a slog handler that discards all log messages.
func NewNopHandler() slog.Handler {
	return &nopHandler{}
}

var _ slog.Handler = (*nopHandler)(nil)

type nopHandler struct{}

func (n *nopHandler) Enabled(_ context.Context, _ slog.Level) bool  { return false }
func (n *nopHandler) Handle(_ context.Context, _ slog.Record) error { return nil }
func (n *nopHandler) WithAttrs(_ []slog.Attr) slog.Handler          { return n }
func (n *nopHandler) WithGroup(_ string) slog.Handler               { return n }

// NewNopSlogger returns a slog logger that discards all log messages.
func NewNopSlogger() *slog.Logger {
	return slog.New(NewNopHandler())
}

// NewNopLogger returns a nop logger that discards all log messages.
func NewNopLogger() *Logger {
	var l *Logger
	return l
}
