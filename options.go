package zlog

import (
	"io"
	"time"

	"golang.org/x/exp/slog"
)

// Option is the option for JSONHandler.
type Option interface {
	apply(*JSONHandler)
}

type optionFunc struct {
	f func(*JSONHandler)
}

func (o optionFunc) apply(h *JSONHandler) {
	o.f(h)
}

// WithAddSource enables add source field.
func WithAddSource(addSource bool) Option {
	return optionFunc{func(h *JSONHandler) {
		h.c.AddSource = addSource
	}}
}

// WithLevel sets the level.
func WithLevel(level slog.Leveler) Option {
	return optionFunc{func(h *JSONHandler) {
		h.c.Level = level
	}}
}

// WithReplaceAttr sets the replaceAttr.
func WithReplaceAttr(replaceAttr func(groups []string, a slog.Attr) slog.Attr) Option {
	return optionFunc{func(h *JSONHandler) {
		h.c.ReplaceAttr = replaceAttr
	}}
}

// WithDevelopment enables development mode like zap development Logger.
func WithDevelopment(development bool) Option {
	return optionFunc{func(h *JSONHandler) {
		h.c.Development = development
	}}
}

// WithWriter sets the writer.
func WithWriter(w io.Writer) Option {
	return optionFunc{func(h *JSONHandler) {
		h.c.Writer = w
		h.writer = newSafeWriter(w)
		h.isTerm = isTerminal(w)
	}}
}

// WithTimeFormatter sets the time formatter.
func WithTimeFormatter(formatter func([]byte, time.Time) []byte) Option {
	return optionFunc{func(h *JSONHandler) {
		if formatter != nil {
			h.c.TimeFormatter = formatter
		}
	}}
}

// WithStacktraceEnabled enables stacktrace for slog.Record.
func WithStacktraceEnabled(enabled bool) Option {
	return optionFunc{func(h *JSONHandler) {
		h.c.StacktraceEnabled = enabled
	}}
}

// WithStacktraceLevel sets the level for stacktrace.
func WithStacktraceLevel(level slog.Leveler) Option {
	return optionFunc{func(h *JSONHandler) {
		h.c.StacktraceLevel = level
	}}
}

// WithStacktraceKey sets the key for stacktrace field.
func WithStacktraceKey(key string) Option {
	return optionFunc{func(h *JSONHandler) {
		h.c.StacktraceKey = key
	}}
}

// WithContextExtractor adds context extractors.
func WithContextExtractor(extractors ...ContextExtractor) Option {
	return optionFunc{func(h *JSONHandler) {
		h.c.ContextExtractors = append(h.c.ContextExtractors, extractors...)
	}}
}
