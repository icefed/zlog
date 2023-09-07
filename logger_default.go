package zlog

import (
	"context"

	"golang.org/x/exp/slog"
)

var defaultLogger = New(NewJSONHandler(nil))

// SetDefault sets the default logger. If l is nil, the default logger is not changed.
func SetDefault(l *Logger) {
	if l == nil {
		return
	}
	defaultLogger = l
}

// With calls Logger.With on the default logger.
func With(args ...any) *Logger {
	return defaultLogger.With(args...)
}

// WithGroup calls Logger.WithGroup on the default logger.
func WithGroup(name string) *Logger {
	return defaultLogger.WithGroup(name)
}

// Log calls Logger.Log on the default logger.
func Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	defaultLogger.Log(ctx, level, msg, args...)
}

// LogAttrs calls Logger.LogAttrs on the default logger.
func LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	defaultLogger.LogAttrs(ctx, level, msg, attrs...)
}

// Debug calls Logger.Debug on the default logger.
func Debug(msg string, args ...any) {
	defaultLogger.Debug(msg, args...)
}

// Debugf calls Logger.Debugf on the default logger.
func Debugf(format string, args ...any) {
	defaultLogger.Debugf(format, args...)
}

// DebugContext calls Logger.DebugContext on the default logger.
func DebugContext(ctx context.Context, msg string, args ...any) {
	defaultLogger.DebugContext(ctx, msg, args...)
}

// DebugContextf calls Logger.DebugContextf on the default logger.
func DebugContextf(ctx context.Context, format string, args ...any) {
	defaultLogger.DebugContextf(ctx, format, args...)
}

// Error calls Logger.Error on the default logger.
func Error(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
}

// Errorf calls Logger.Errorf on the default logger.
func Errorf(format string, args ...any) {
	defaultLogger.Errorf(format, args...)
}

// ErrorContext calls Logger.ErrorContext on the default logger.
func ErrorContext(ctx context.Context, msg string, args ...any) {
	defaultLogger.ErrorContext(ctx, msg, args...)
}

// ErrorContextf calls Logger.ErrorContextf on the default logger.
func ErrorContextf(ctx context.Context, format string, args ...any) {
	defaultLogger.ErrorContextf(ctx, format, args...)
}

// Info calls Logger.Info on the default logger.
func Info(msg string, args ...any) {
	defaultLogger.Info(msg, args...)
}

// Infof calls Logger.Infof on the default logger.
func Infof(format string, args ...any) {
	defaultLogger.Infof(format, args...)
}

// InfoContext calls Logger.InfoContext on the default logger.
func InfoContext(ctx context.Context, msg string, args ...any) {
	defaultLogger.InfoContext(ctx, msg, args...)
}

// InfoContextf calls Logger.InfoContextf on the default logger.
func InfoContextf(ctx context.Context, format string, args ...any) {
	defaultLogger.InfoContextf(ctx, format, args...)
}

// Warn calls Logger.Warn on the default logger.
func Warn(msg string, args ...any) {
	defaultLogger.Warn(msg, args...)
}

// Warnf calls Logger.Warnf on the default logger.
func Warnf(format string, args ...any) {
	defaultLogger.Warnf(format, args...)
}

// WarnContext calls Logger.WarnContext on the default logger.
func WarnContext(ctx context.Context, msg string, args ...any) {
	defaultLogger.WarnContext(ctx, msg, args...)
}

// WarnContextf calls Logger.WarnContextf on the default logger.
func WarnContextf(ctx context.Context, format string, args ...any) {
	defaultLogger.WarnContextf(ctx, format, args...)
}
