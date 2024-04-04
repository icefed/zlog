package benchmarks

import (
	"io"
	"log/slog"

	"go.uber.org/zap/exp/zapslog"
	"go.uber.org/zap/zapcore"
)

func newSlog(fields ...slog.Attr) *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil).WithAttrs(fields))
}

func newDisabledSlog(fields ...slog.Attr) *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}).WithAttrs(fields))
}

func slogFields() []slog.Attr {
	return []slog.Attr{
		slog.String("string", testString),
		slog.String("longstring", testMessage),
		slog.Any("strings", testStrings),
		slog.Int("int", testInt),
		slog.Any("ints", testInts),
		slog.Time("time", testTime),
		slog.Any("times", testTimes),
		slog.Any("struct", testStruct),
		slog.Any("structs", testStructs),
		slog.Any("error", testErr),
	}
}
func kvArgs() []any {
	return []any{
		"string", testString,
		"longstring", testMessage,
		"strings", testStrings,
		"int", testInt,
		"ints", testInts,
		"time", testTime,
		"times", testTimes,
		"struct", testStruct,
		"structs", testStructs,
		"error", testErr,
	}
}

func newSlogWithZap(fields ...slog.Attr) *slog.Logger {
	logger := newZapLogger(zapcore.DebugLevel)
	return slog.New(zapslog.NewHandler(logger.Core(), nil).WithAttrs(fields))
}

func newDisabledSlogWithZap(fields ...slog.Attr) *slog.Logger {
	logger := newZapLogger(zapcore.ErrorLevel)
	return slog.New(zapslog.NewHandler(logger.Core(), nil).WithAttrs(fields))
}
