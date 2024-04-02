package zlog

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"time"

	"github.com/icefed/zlog/buffer"
)

type Logger struct {
	h *JSONHandler

	callerSkip int
}

// New creates a new Logger. NewJSONHandler(nil) will be used if h is nil.
func New(h *JSONHandler) *Logger {
	if h == nil {
		h = NewJSONHandler(nil)
	}
	l := &Logger{
		h: h,
	}

	return l
}

func (l *Logger) clone() *Logger {
	return &Logger{
		h: l.h,
	}
}

// Enabled reports whether the handler handles records at the given level.
func (l *Logger) Enabled(ctx context.Context, level slog.Level) bool {
	if l == nil {
		return false
	}
	return l.h.Enabled(ctx, level)
}

// With returns a new logger with the given arguments.
func (l *Logger) With(args ...any) *Logger {
	if l == nil {
		return l
	}
	newLogger := l.clone()
	newLogger.h = l.h.WithAttrs(argsToAttrs(args...)).(*JSONHandler)
	return newLogger
}

// WithContext returns a new logger with the given handler options.
func (l *Logger) WithOptions(opts ...Option) *Logger {
	if l == nil {
		return l
	}
	newLogger := l.clone()
	newLogger.h = l.h.WithOptions(opts...)
	return newLogger
}

// WithCallerSkip returns a new logger with the given caller skip.
// argument 'skip' will be added to the caller skip in the logger, which is passed
// as the first parameter 'skip' when calling runtime.Callers to get the source's pc.
// If 'skip' is 1, then skip a caller frame, if you have set callerskip, 'skip' can
// also be negative to subtract callerskip.
func (l *Logger) WithCallerSkip(skip int) *Logger {
	if l == nil {
		return l
	}
	newLogger := l.clone()
	newLogger.callerSkip += skip
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

// WithGroup returns a new logger with the given group name.
func (l *Logger) WithGroup(name string) *Logger {
	if l == nil {
		return l
	}
	newLogger := l.clone()
	newLogger.h = l.h.WithGroup(name).(*JSONHandler)
	return newLogger
}

// Handler returns the handler.
func (l *Logger) Handler() *JSONHandler {
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
	if l.h.CapturePC(level) {
		var pcs [1]uintptr
		// skip runtime.Callers, log, log's caller, and l.callerSkip
		runtime.Callers(3+l.callerSkip, pcs[:])
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
	if l.h.CapturePC(level) {
		var pcs [1]uintptr
		// skip runtime.Callers, logAttrs, logAttrs's caller, and l.callerSkip
		runtime.Callers(3+l.callerSkip, pcs[:])
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

func (l *Logger) logf(ctx context.Context, level slog.Level, format string, args ...any) {
	if !l.Enabled(ctx, level) {
		return
	}

	buf := buffer.New()
	defer buf.Free()
	*buf = fmt.Appendf(*buf, format, args...)
	l.log(ctx, level, buf.String())
}

// Log prints log as a JSON object on a single line with the given level and message.
// Log follows the rules of slog.Logger.Log, args can be slog.Attr or will be
// converted to slog.Attr in pairs.
func (l *Logger) Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	l.log(ctx, level, msg, args...)
}

// Logf prints log as a JSON object on a single line with the given level, message and attrs.
func (l *Logger) LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	l.logAttrs(ctx, level, msg, attrs...)
}

// Debug prints log message at the debug level.
func (l *Logger) Debug(msg string, args ...any) {
	l.log(nil, slog.LevelDebug, msg, args...)
}

// Debugf prints log message at the debug level, fmt.Sprintf is used to format.
func (l *Logger) Debugf(format string, args ...any) {
	l.logf(nil, slog.LevelDebug, format, args...)
}

// DebugContext prints log message at the debug level with context.
func (l *Logger) DebugContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, slog.LevelDebug, msg, args...)
}

// DebugContextf prints log message at the debug level with context, fmt.Sprintf is used to format.
func (l *Logger) DebugContextf(ctx context.Context, format string, args ...any) {
	l.logf(ctx, slog.LevelDebug, format, args...)
}

// Info prints log message at the info level.
func (l *Logger) Info(msg string, args ...any) {
	l.log(context.Background(), slog.LevelInfo, msg, args...)
}

// Infof prints log message at the info level, fmt.Sprintf is used to format.
func (l *Logger) Infof(format string, args ...any) {
	l.logf(context.Background(), slog.LevelInfo, format, args...)
}

// InfoContext prints log message at the info level with context.
func (l *Logger) InfoContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, slog.LevelInfo, msg, args...)
}

// InfoContextf prints log message at the info level with context, fmt.Sprintf is used to format.
func (l *Logger) InfoContextf(ctx context.Context, format string, args ...any) {
	l.logf(ctx, slog.LevelInfo, format, args...)
}

// Warn prints log message at the warn level.
func (l *Logger) Warn(msg string, args ...any) {
	l.log(context.Background(), slog.LevelWarn, msg, args...)
}

// Warnf prints log message at the warn level, fmt.Sprintf is used to format.
func (l *Logger) Warnf(format string, args ...any) {
	l.logf(nil, slog.LevelWarn, format, args...)
}

// WarnContext prints log message at the warn level with context.
func (l *Logger) WarnContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, slog.LevelWarn, msg, args...)
}

// WarnContextf prints log message at the warn level with context, fmt.Sprintf is used to format.
func (l *Logger) WarnContextf(ctx context.Context, format string, args ...any) {
	l.logf(ctx, slog.LevelWarn, format, args...)
}

// Error prints log message at the error level.
func (l *Logger) Error(msg string, args ...any) {
	l.log(context.Background(), slog.LevelError, msg, args...)
}

// Errorf prints log message at the error level, fmt.Sprintf is used to format.
func (l *Logger) Errorf(format string, args ...any) {
	l.logf(nil, slog.LevelError, format, args...)
}

// ErrorContext prints log message at the error level with context.
func (l *Logger) ErrorContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, slog.LevelError, msg, args...)
}

// ErrorContextf prints log message at the error level with context, fmt.Sprintf is used to format.
func (l *Logger) ErrorContextf(ctx context.Context, format string, args ...any) {
	l.logf(ctx, slog.LevelError, format, args...)
}
