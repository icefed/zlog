package zlog

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"golang.org/x/exp/slog"
)

type Logger struct {
	h *JSONHandler

	capturePC bool
}

// New creates a new Logger. NewJSONHandler(nil) will be used if h is nil.
func New(h *JSONHandler) *Logger {
	if h == nil {
		h = NewJSONHandler(nil)
	}
	l := &Logger{
		h:         h,
		capturePC: h.GetAddSource(),
	}

	return l
}

func (l *Logger) clone() *Logger {
	return &Logger{
		h:         l.h,
		capturePC: l.capturePC,
	}
}

func (l *Logger) Enabled(ctx context.Context, level slog.Level) bool {
	if l == nil {
		return false
	}
	return l.h.Enabled(ctx, level)
}

func (l *Logger) With(args ...any) *Logger {
	if l == nil {
		return l
	}
	newLogger := l.clone()
	newLogger.h = l.h.WithAttrs(argsToAttrs(args...)).(*JSONHandler)
	return newLogger
}

func (l *Logger) WithOptions(opts ...Option) *Logger {
	if l == nil {
		return l
	}
	newLogger := l.clone()
	newLogger.h = l.h.WithOptions(opts...)
	newLogger.capturePC = newLogger.h.GetAddSource()
	return newLogger
}

var badKey = "!BADKEY"

func argsToAttrs(args ...any) []slog.Attr {
	var (
		attrs     []slog.Attr
		totalArgs = len(args)
	)
	for i := 0; i < totalArgs; i++ {
		switch arg := args[i].(type) {
		case string:
			if i == totalArgs-1 {
				attrs = append(attrs, slog.String(badKey, arg))
				return attrs
			}
			attrs = append(attrs, slog.Any(arg, args[i+1]))
			i++
		case slog.Attr:
			attrs = append(attrs, arg)
		default:
			attrs = append(attrs, slog.Any(badKey, arg))
		}
	}
	return attrs
}

func (l *Logger) WithGroup(name string) *Logger {
	if l == nil {
		return l
	}
	newLogger := l.clone()
	newLogger.h = l.h.WithGroup(name).(*JSONHandler)
	return newLogger
}

func (l *Logger) Handler() slog.Handler {
	if l == nil {
		return nil
	}
	return l.h
}

func (l *Logger) log(ctx context.Context, level slog.Level, msg string, args ...any) {
	if !l.Enabled(ctx, level) {
		return
	}
	var pc uintptr
	if l.capturePC {
		var pcs [1]uintptr
		runtime.Callers(3, pcs[:])
		pc = pcs[0]
	}

	r := slog.NewRecord(time.Now(), level, msg, pc)
	r.Add(args...)

	if ctx == nil {
		ctx = context.Background()
	}
	// whether error needs to be handled
	_ = l.h.Handle(ctx, r)
}

func (l *Logger) logAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	if !l.Enabled(ctx, level) {
		return
	}
	var pc uintptr
	if l.capturePC {
		var pcs [1]uintptr
		runtime.Callers(3, pcs[:])
		pc = pcs[0]
	}
	r := slog.NewRecord(time.Now(), level, msg, pc)
	r.AddAttrs(attrs...)

	if ctx == nil {
		ctx = context.Background()
	}
	// whether error needs to be handled
	_ = l.h.Handle(ctx, r)
}

func (l *Logger) Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	l.log(ctx, level, msg, args...)
}

func (l *Logger) LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	l.logAttrs(ctx, level, msg, attrs...)
}

func (l *Logger) Debug(msg string, args ...any) {
	l.log(context.Background(), slog.LevelDebug, msg, args...)
}

func (l *Logger) Debugf(format string, args ...any) {
	l.log(context.Background(), slog.LevelDebug, fmt.Sprintf(format, args...))
}

func (l *Logger) DebugContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, slog.LevelDebug, msg, args...)
}

func (l *Logger) DebugContextf(ctx context.Context, format string, args ...any) {
	l.log(ctx, slog.LevelDebug, fmt.Sprintf(format, args...))
}

func (l *Logger) Error(msg string, args ...any) {
	l.log(context.Background(), slog.LevelError, msg, args...)
}

func (l *Logger) Errorf(format string, args ...any) {
	l.log(context.Background(), slog.LevelError, fmt.Sprintf(format, args...))
}

func (l *Logger) ErrorContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, slog.LevelError, msg, args...)
}

func (l *Logger) ErrorContextf(ctx context.Context, format string, args ...any) {
	l.log(ctx, slog.LevelError, fmt.Sprintf(format, args...))
}

func (l *Logger) Info(msg string, args ...any) {
	l.log(context.Background(), slog.LevelInfo, msg, args...)
}

func (l *Logger) Infof(format string, args ...any) {
	l.log(context.Background(), slog.LevelInfo, fmt.Sprintf(format, args...))
}

func (l *Logger) InfoContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, slog.LevelInfo, msg, args...)
}

func (l *Logger) InfoContextf(ctx context.Context, format string, args ...any) {
	l.log(ctx, slog.LevelInfo, fmt.Sprintf(format, args...))
}

func (l *Logger) Warn(msg string, args ...any) {
	l.log(context.Background(), slog.LevelWarn, msg, args...)
}

func (l *Logger) Warnf(format string, args ...any) {
	l.log(context.Background(), slog.LevelWarn, fmt.Sprintf(format, args...))
}

func (l *Logger) WarnContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, slog.LevelWarn, msg, args...)
}

func (l *Logger) WarnContextf(ctx context.Context, format string, args ...any) {
	l.log(ctx, slog.LevelWarn, fmt.Sprintf(format, args...))
}
