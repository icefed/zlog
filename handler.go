package zlog

import (
	"context"
	"io"
	"log/slog"
	"os"
	"slices"
	"time"

	"golang.org/x/term"

	"github.com/icefed/zlog/buffer"
)

// JSONHandler implements the slog.Handler interface, transforming r.Record
// into the JSON format, and follows therules set by slog.Handler.
//
// Additionally, it provides support for the development mode, akin to zap,
// which allows built-in attributes to be output in a human-friendly format.
type JSONHandler struct {
	c *Config
	// the writer is a terminal file descriptor.
	isTerm bool

	groups                 []string
	preformattedGroupAttrs []byte
}

// ContextExtractor get attributes from context, that can be used in slog.Handler.
type ContextExtractor func(context.Context) []slog.Attr

// Config the configuration for the JSONHandler.
type Config struct {
	// If nil, a default handler is used.
	slog.HandlerOptions

	// Development enables development mode like zap development Logger,
	// that will write logs in human-friendly format,
	// also level will be colored if output is a terminal.
	Development bool

	// Writer is the writer to use. If nil, os.Stderr is used.
	Writer io.Writer

	// TimeFormatter is the time formatter to use for buildin attribute time value. If nil, use format RFC3339Milli as default.
	TimeFormatter AppendTimeFunc

	// TimeDurationAsInt format time.Duration as int if true, otherwise format as string using time.Duration.String method, eg: 3m10s.
	TimeDurationAsInt bool

	// built-in attribute keys, use slog's default if not set.
	// https://pkg.go.dev/log/slog#pkg-constants
	TimeKey    string
	LevelKey   string
	MessageKey string
	SourceKey  string

	// StacktraceEnabled enables stack trace for slog.Record.
	StacktraceEnabled bool
	// StacktraceLevel means which slog.Level from we should enable stack trace.
	// Default is slog.LevelError.
	StacktraceLevel slog.Leveler
	// StacktraceKey is the key for stacktrace field, default is "stacktrace".
	StacktraceKey string

	// ContextExtractors will be used in Handler.Handle
	ContextExtractors []ContextExtractor

	// If a group has no Attrs (even if it has a non-empty key), ignore it.
	IgnoreEmptyGroup bool
}

func (c *Config) copy() *Config {
	newConfig := *c
	newConfig.ContextExtractors = slices.Clone(c.ContextExtractors)
	return &newConfig
}

// AppendTimeFunc append the formatted value to buf and returns the extended buffer.
type AppendTimeFunc func(buf []byte, t time.Time) []byte

// RFC3339Milli define the time format as RFC3339 with millisecond precision.
const RFC3339Milli = "2006-01-02T15:04:05.999Z07:00"

var defaultConfig = Config{
	HandlerOptions: slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelInfo,
	},
	Development: false,
	// use stderr as default writer
	Writer: os.Stderr,
	TimeFormatter: func(buf []byte, t time.Time) []byte {
		return t.AppendFormat(buf, RFC3339Milli)
	},
	TimeKey:           slog.TimeKey,
	LevelKey:          slog.LevelKey,
	MessageKey:        slog.MessageKey,
	SourceKey:         slog.SourceKey,
	StacktraceEnabled: false,
	StacktraceLevel:   slog.LevelError,
	StacktraceKey:     "stacktrace",
}

// NewJSONHandler creates a slog handler that writes log messages as JSON.
// If config is nil, a default configuration is used.
func NewJSONHandler(config *Config) *JSONHandler {
	var c Config
	if config == nil {
		c = defaultConfig
	} else {
		c = *config
		if c.Writer == nil {
			c.Writer = defaultConfig.Writer
		}
		if c.TimeFormatter == nil {
			c.TimeFormatter = defaultConfig.TimeFormatter
		}
		if c.StacktraceLevel == nil {
			c.StacktraceLevel = defaultConfig.StacktraceLevel
		}
		if c.StacktraceKey == "" {
			c.StacktraceKey = defaultConfig.StacktraceKey
		}
		if c.TimeKey == "" {
			c.TimeKey = defaultConfig.TimeKey
		}
		if c.LevelKey == "" {
			c.LevelKey = defaultConfig.LevelKey
		}
		if c.MessageKey == "" {
			c.MessageKey = defaultConfig.MessageKey
		}
		if c.SourceKey == "" {
			c.SourceKey = defaultConfig.SourceKey
		}
		c.ContextExtractors = slices.Clone(c.ContextExtractors)
	}

	handler := &JSONHandler{
		c:      &c,
		isTerm: isTerminal(c.Writer),
	}
	return handler
}

// Enabled reports whether the handler handles records at the given level. The handler ignores records whose level is lower.
// https://pkg.go.dev/log/slog#Handler
func (h *JSONHandler) Enabled(_ context.Context, level slog.Level) bool {
	if h.c.Level == nil {
		return level >= defaultConfig.Level.Level()
	}
	return level >= h.c.Level.Level()
}

// CapturePC returns true if the handler has AddSource option enabled or the stacktrace
// is enabled at the given level.
// Logger should set PC in the slog.Record if this function returns true.
func (h *JSONHandler) CapturePC(level slog.Level) bool {
	return h.c.AddSource || h.stacktraceEnabled(level)
}

// WithOptions return a new handler with the given options.
// Options will override the hander's config.
func (h *JSONHandler) WithOptions(opts ...Option) *JSONHandler {
	newHandler := h.clone()
	for i := range opts {
		opts[i].apply(newHandler.c)
	}
	newHandler.isTerm = isTerminal(newHandler.c.Writer)
	return newHandler
}

// stacktraceEnabled reports whether the handler should record the stack trace of a slog.Record at the given level.
func (h *JSONHandler) stacktraceEnabled(level slog.Level) bool {
	if !h.c.StacktraceEnabled {
		return false
	}
	return level >= h.c.StacktraceLevel.Level()
}

// Handle formats its argument Record as a JSON object on a single line.
// https://pkg.go.dev/log/slog#Handler
func (h *JSONHandler) Handle(ctx context.Context, r slog.Record) error {
	buf := buffer.New()
	defer buf.Free()

	if h.c.Development {
		h.encodeDevelopment(ctx, r, buf)
	} else {
		h.encode(ctx, r, buf)
	}

	_, err := h.c.Writer.Write(buf.Bytes())
	return err
}

func (h *JSONHandler) contextAttrs(ctx context.Context, f func(slog.Attr)) {
	for _, ex := range h.c.ContextExtractors {
		if ex == nil {
			continue
		}
		attrs := ex(ctx)
		for i := range attrs {
			f(attrs[i])
		}
	}
}

const (
	lineEnding = '\n'
)

func (h *JSONHandler) encodeDevelopment(ctx context.Context, r slog.Record, buf *buffer.Buffer) {
	tenc := newTextEncoder(h, buf)
	// time
	// If r.Time is the zero time, ignore the time.
	if !r.Time.IsZero() {
		tenc.Append(h.c.TimeKey, r.Time)
		buf.WriteString("  ")
	}
	// level
	tenc.Append(h.c.LevelKey, r.Level)
	// source
	// If r.PC is zero, ignore it.
	if h.c.AddSource && r.PC != 0 {
		buf.WriteByte('\t')
		tenc.Append(h.c.SourceKey, r.PC)
	}
	// message
	if r.Message != "" {
		buf.WriteByte('\t')
		tenc.Append(h.c.MessageKey, r.Message)
	}

	size := len(buf.Bytes())

	buf.WriteByte('\t')
	enc := newJSONEncoder(h, buf)
	buf.WriteByte('{')
	// preformatted attrs
	enc.AppendFormatted(h.preformattedGroupAttrs)
	// add context attrs
	h.contextAttrs(ctx, func(attr slog.Attr) {
		enc.AppendAttr(attr)
	})
	// add record attrs
	r.Attrs(func(attr slog.Attr) bool {
		enc.AppendAttr(attr)
		return true
	})
	enc.CloseGroups()
	buf.WriteByte('}')

	if buf.Len()-size == 3 {
		buf.Truncate(size)
	}

	if *buf.LastByte() != lineEnding {
		buf.WriteByte(lineEnding)
	}
	// stack trace
	if h.stacktraceEnabled(r.Level) && r.PC != 0 {
		tenc.Append(h.c.StacktraceKey, &stacktrace{r.PC})
	}
	if *buf.LastByte() != lineEnding {
		buf.WriteByte(lineEnding)
	}
}

func (h *JSONHandler) encode(ctx context.Context, r slog.Record, buf *buffer.Buffer) {
	enc := newJSONEncoder(h, buf)
	buf.WriteByte('{')
	// time
	// If r.Time is the zero time, ignore the time.
	if !r.Time.IsZero() {
		enc.AppendTime(h.c.TimeKey, r.Time)
	}
	// level
	enc.AppendLevel(h.c.LevelKey, r.Level)
	// source
	// If r.PC is zero, ignore it.
	if h.c.AddSource && r.PC != 0 {
		enc.AppendSourceFromPC(h.c.SourceKey, r.PC)
	}
	// message
	enc.AppendMessage(h.c.MessageKey, r.Message)

	// preformatted attrs
	enc.AppendFormatted(h.preformattedGroupAttrs)
	// add context attrs
	h.contextAttrs(ctx, func(attr slog.Attr) {
		enc.AppendAttr(attr)
	})
	// add record attrs
	r.Attrs(func(attr slog.Attr) bool {
		enc.AppendAttr(attr)
		return true
	})
	enc.CloseGroups()
	// stack trace
	if h.stacktraceEnabled(r.Level) && r.PC != 0 {
		enc.AppendStacktrace(h.c.StacktraceKey, &stacktrace{r.PC})
	}
	buf.WriteByte('}')
	buf.WriteByte(lineEnding)
}

// WithAttrs implements the slog.Handler WithAttrs method.
// https://pkg.go.dev/log/slog#Handler
func (h *JSONHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}
	newHandler := h.clone()
	newHandler.addAttrs(attrs)
	return newHandler
}

func (h *JSONHandler) addAttrs(attrs []slog.Attr) {
	enc := newJSONEncoder(h, (*buffer.Buffer)(&h.preformattedGroupAttrs))
	for i := range attrs {
		enc.AppendAttr(attrs[i])
	}
}

// WithGroup implements the slog.Handler WithGroup method.
// https://pkg.go.dev/log/slog#Handler
func (h *JSONHandler) WithGroup(name string) slog.Handler {
	newHandler := h.clone()
	if name == "" {
		return newHandler
	}
	newHandler.addGroup(name)
	return newHandler
}

func (h *JSONHandler) addGroup(name string) {
	enc := newJSONEncoder(h, (*buffer.Buffer)(&h.preformattedGroupAttrs))
	enc.OpenGroup(name)
	h.groups = append(h.groups, name)
}

func (h *JSONHandler) clone() *JSONHandler {
	newHandler := &JSONHandler{
		c:                      h.c.copy(),
		isTerm:                 h.isTerm,
		groups:                 slices.Clip(h.groups),
		preformattedGroupAttrs: slices.Clip(h.preformattedGroupAttrs),
	}

	return newHandler
}

// needColoredLevel returns true if in development mode and output writer is a terminal.
func (h *JSONHandler) needColoredLevel() bool {
	return h.c.Development && h.isTerm
}

// isTerminal returns true if w is a terminal file descriptor.
func isTerminal(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		return term.IsTerminal(int(f.Fd()))
	}
	return false
}
