package zlog

import (
	"io"
	"log/slog"
	"time"
)

// Option is the option for JSONHandler.
type Option interface {
	apply(*Config)
}

type optionFunc struct {
	f func(*Config)
}

func (o optionFunc) apply(c *Config) {
	o.f(c)
}

// WithAddSource enables add source field.
func WithAddSource(addSource bool) Option {
	return optionFunc{func(c *Config) {
		c.AddSource = addSource
	}}
}

// WithLevel sets the level.
func WithLevel(level slog.Leveler) Option {
	return optionFunc{func(c *Config) {
		c.Level = level
	}}
}

// WithReplaceAttr sets the replaceAttr.
func WithReplaceAttr(replaceAttr func(groups []string, a slog.Attr) slog.Attr) Option {
	return optionFunc{func(c *Config) {
		c.ReplaceAttr = replaceAttr
	}}
}

// WithDevelopment enables development mode like zap development Logger.
func WithDevelopment(development bool) Option {
	return optionFunc{func(c *Config) {
		c.Development = development
	}}
}

// WithWriter sets the writer.
func WithWriter(w io.Writer) Option {
	return optionFunc{func(c *Config) {
		c.Writer = w
	}}
}

// WithTimeFormatter sets the time formatter.
func WithTimeFormatter(formatter func([]byte, time.Time) []byte) Option {
	return optionFunc{func(c *Config) {
		if formatter != nil {
			c.TimeFormatter = formatter
		}
	}}
}

// WithStacktraceEnabled enables stacktrace for slog.Record.
func WithStacktraceEnabled(enabled bool) Option {
	return optionFunc{func(c *Config) {
		c.StacktraceEnabled = enabled
	}}
}

// WithStacktraceLevel sets the level for stacktrace.
func WithStacktraceLevel(level slog.Leveler) Option {
	return optionFunc{func(c *Config) {
		c.StacktraceLevel = level
	}}
}

// WithStacktraceKey sets the key for stacktrace field.
func WithStacktraceKey(key string) Option {
	return optionFunc{func(c *Config) {
		c.StacktraceKey = key
	}}
}

// WithContextExtractor adds context extractors.
func WithContextExtractor(extractors ...ContextExtractor) Option {
	return optionFunc{func(c *Config) {
		c.ContextExtractors = append(c.ContextExtractors, extractors...)
	}}
}
