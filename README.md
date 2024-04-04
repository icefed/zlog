# zlog - JSON structured handler/logger for Golang slog

[![GoDoc](https://godoc.org/github.com/icefed/zlog?status.svg)](https://pkg.go.dev/github.com/icefed/zlog)
[![Go Report Card](https://goreportcard.com/badge/github.com/icefed/zlog)](https://goreportcard.com/report/github.com/icefed/zlog)
![Build Status](https://github.com/icefed/zlog/actions/workflows/test.yml/badge.svg)
[![Coverage](https://img.shields.io/codecov/c/github/icefed/zlog)](https://codecov.io/gh/icefed/zlog)
[![License](https://img.shields.io/github/license/icefed/zlog)](./LICENSE)

## Features
- JSON Structured logging
- Logger with format method(printf-style)
- Development mode with human-friendly output
- WithCallerSkip to skip caller
- Context extractor for Record context
- Custom time formatter for buildin attribute time value

## Usage

More examples can be found in [examples](https://github.com/icefed/zlog/tree/master/examples).

Because zlog implements the slog.Handler interface, you can create a zlog.JSONHander and use slog.Logger.
```go
import (
    "context"
    "log/slog"

    "github.com/icefed/zlog"
)

func main() {
    h := zlog.NewJSONHandler(&zlog.Config{
		HandlerOptions: slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
    })
    log := slog.New(h)

    log.Info("hello world")
    // ...
    log.With(slog.String("app", "test")).
        Error("db execution failed", "error", err)
    log.LogAttrs(context.Background(), slog.LevelInfo, "this is a info message", slog.String("app", "test"))
}
```

Or you can use zlog.Logger, which implements all slog.Logger methods and is compatible.
Then you can use Infof and other methods that support format format.
```go
import (
	"context"
	"log/slog"

	"github.com/icefed/zlog"
)

func main() {
	h := zlog.NewJSONHandler(&zlog.Config{
		HandlerOptions: slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	})
	log := zlog.New(h)
	log.Info("hello world")

	log.Log(context.Background(), slog.LevelInfo, "this is a info message")
	// ...
	log.Debugf("get value %s from map by key %s", v, k)
}
```

### Development mode

Development mode, like zap development, outputs buildin attributes in Text format for better readability.  If development mode is enabled and writer is a terminal, the level field will be printed in color.
```go
package main

import (
	"log/slog"

	"github.com/icefed/zlog"
)

func main() {
	// start development mode with Config
	h := zlog.NewJSONHandler(&zlog.Config{
		HandlerOptions: slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
		Development: true,
	})

	// turn on development mode with WithOptions
	h = h.WithOptions(zlog.WithDevelopment(true))
	log := zlog.New(h)

	log.Debug("Processing data file", "file", "data.json")
	log.Info("Application started successfully",
		slog.String("version", "1.0.0"),
		slog.String("environment", "dev"))
	log.Warn("Deprecated method 'foo()' is being used", slog.Int("warning_code", 123))
	log.Error("Failed to connect to the database", "error_code", 500, "component", "DatabaseConnection")
}
```

Outputs:
![](images/development.png)

### Enable stack trace

Set StacktraceEnabled to true to enable printing log stack trace, the default print slog.LevelError above the level,
```go
h := zlog.NewJSONHandler(&zlog.Config{
    HandlerOptions: slog.HandlerOptions{
        Level: slog.LevelDebug,
    },
    StacktraceEnabled: true,
})

// set custom stacktrace key
h = h.WithOptions(zlog.WithStacktraceKey("stack"))
```

### Custom time formatter

By default, when printing logs, the time field is formatted with `RFC3339Milli`(`2006-01-02T15:04:05.999Z07:00`). If you want to modify the format, you can configure TimeFormatter in Config.
```go
h := zlog.NewJSONHandler(&zlog.Config{
    HandlerOptions: slog.HandlerOptions{
        Level: slog.LevelDebug,
    },
    TimeFormatter: func(buf []byte,t time.Time) []byte {
        return t.AppendFormat(buf, time.RFC3339Nano)
    },
})

log := zlog.New(h)
log.Info("this is a log message with RFC3339Nano format")

// use int timestamp format with microsecond precision
log = log.WithOptions(zlog.WithTimeFormatter(func(buf []byte, t time.Time) []byte {
    return strconv.AppendInt(buf, t.UnixMicro(), 10)
}))
log.Info("this is a log message in int timestamp format")
```

Outputs:
```
{"time":"2023-09-09T19:02:28.704746+08:00","level":"INFO","msg":"this is a log message with RFC3339Nano format"}
{"time":"1694257348705059","level":"INFO","msg":"this is a log message with int timestamp format"}
```

### Context extractor

We often need to extract the value from the context and print it to the log, for example, an apiserver receives a user request and prints trace and user information to the log.

This example shows how to use the context, and print OpenTelemetry trace in log. If you have an api server that supports OpenTelemetry, you can use this example in your handler middleware and print trace in each log.

```go
package main

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"

	"github.com/icefed/zlog"
)

// traceContextExtractor implement the ContextExtractor, extracts trace context from context
func traceContextExtractor(ctx context.Context) []slog.Attr {
	spanContext := trace.SpanContextFromContext(ctx)
	if spanContext.IsValid() {
		return []slog.Attr{
			slog.Group("trace",
				slog.String("traceID", spanContext.TraceID().String()),
				slog.String("spanID", spanContext.SpanID().String()),
			),
		}
	}
	return nil
}

func parentFun(ctx context.Context, tracer trace.Tracer) {
	ctx, parentSpan := tracer.Start(ctx, "caller1")
	defer parentSpan.End()

	// print log
	slog.InfoContext(ctx, "call parentFun")

	childFun(ctx, tracer)
}

func childFun(ctx context.Context, tracer trace.Tracer) {
	ctx, childSpan := tracer.Start(ctx, "caller2")
	defer childSpan.End()

	// print log
	slog.InfoContext(ctx, "call childFun")
}

func main() {
	// create a logger with traceContextExtractor
	h := zlog.NewJSONHandler(nil)
	h = h.WithOptions(zlog.WithContextExtractor(traceContextExtractor))
	log := slog.New(h)
	slog.SetDefault(log)

	// prepare a call with trace context
	exporter := tracetest.NewInMemoryExporter()
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(tracerProvider)
	tracer := otel.Tracer("api")
	ctx := context.Background()
	parentFun(ctx, tracer)
}
```

Outputs:
```
{"time":"2023-09-08T20:12:14.733","level":"INFO","msg":"call parentFun","trace":{"traceID":"95f0717d9da16177176efdbc7c06bfbd","spanID":"7718edf7b2a8388d"}}
{"time":"2023-09-08T20:12:14.733","level":"INFO","msg":"call childFun","trace":{"traceID":"95f0717d9da16177176efdbc7c06bfbd","spanID":"ef83f673951742b0"}}
```

## Benchmarks

Test modified from [zap benchmarking suite](https://github.com/uber-go/zap/tree/master/benchmarks).

```bench
goos: darwin
goarch: arm64
pkg: github.com/icefed/zlog/benchmarks
BenchmarkDisabledWithoutFields/slog-10          1000000000               0.6019 ns/op          0 B/op          0 allocs/op
BenchmarkDisabledWithoutFields/slog_with_zlog-10                1000000000               0.5913 ns/op          0 B/op          0 allocs/op
BenchmarkDisabledWithoutFields/zlog-10                          1000000000               0.5496 ns/op          0 B/op          0 allocs/op
BenchmarkDisabledWithoutFields/slog_with_zap-10                 1000000000               0.8184 ns/op          0 B/op          0 allocs/op
BenchmarkDisabledWithoutFields/zap-10                           1000000000               0.6048 ns/op          0 B/op          0 allocs/op
BenchmarkDisabledWithoutFields/zerolog-10                       1000000000               0.3016 ns/op          0 B/op          0 allocs/op
BenchmarkDisabledAddingFields/slog-10                           57371786               208.7 ns/op           576 B/op          6 allocs/op
BenchmarkDisabledAddingFields/slog_with_zlog-10                 56256844               208.3 ns/op           576 B/op          6 allocs/op
BenchmarkDisabledAddingFields/zlog-10                           58574133               205.8 ns/op           576 B/op          6 allocs/op
BenchmarkDisabledAddingFields/slog_with_zap-10                  58468917               206.9 ns/op           576 B/op          6 allocs/op
BenchmarkDisabledAddingFields/zap-10                            41587197               295.7 ns/op           864 B/op          6 allocs/op
BenchmarkDisabledAddingFields/zerolog-10                        303398799               39.08 ns/op           88 B/op          2 allocs/op
BenchmarkWithoutFields/slog-10                                  47158194               231.9 ns/op             0 B/op          0 allocs/op
BenchmarkWithoutFields/slog_with_zlog-10                        143867673               81.77 ns/op            0 B/op          0 allocs/op
BenchmarkWithoutFields/zlog-10                                  182174908               64.98 ns/op            0 B/op          0 allocs/op
BenchmarkWithoutFields/slog_with_zap-10                         125818678               95.04 ns/op            0 B/op          0 allocs/op
BenchmarkWithoutFields/zap-10                                   169056685               72.34 ns/op            0 B/op          0 allocs/op
BenchmarkWithoutFields/zerolog-10                               348895234               35.03 ns/op            0 B/op          0 allocs/op
BenchmarkAccumulatedContext/slog-10                             51551946               241.8 ns/op             0 B/op          0 allocs/op
BenchmarkAccumulatedContext/slog_with_zlog-10                   138282912               90.81 ns/op            0 B/op          0 allocs/op
BenchmarkAccumulatedContext/zlog-10                             184607311               66.28 ns/op            0 B/op          0 allocs/op
BenchmarkAccumulatedContext/slog_with_zap-10                    132471319               89.32 ns/op            0 B/op          0 allocs/op
BenchmarkAccumulatedContext/zap-10                              167639162               76.73 ns/op            0 B/op          0 allocs/op
BenchmarkAccumulatedContext/zerolog-10                          314418170               36.53 ns/op            0 B/op          0 allocs/op
BenchmarkAddingFields/slog-10                                    4498432              2639 ns/op            3951 B/op         38 allocs/op
BenchmarkAddingFields/slog_with_zlog-10                          9287110              1300 ns/op            1344 B/op         20 allocs/op
BenchmarkAddingFields/zlog-10                                    9383146              1290 ns/op            1344 B/op         20 allocs/op
BenchmarkAddingFields/slog_with_zap-10                           6758552              1776 ns/op            2218 B/op         23 allocs/op
BenchmarkAddingFields/zap-10                                     8615821              1389 ns/op            1508 B/op         18 allocs/op
BenchmarkAddingFields/zerolog-10                                 9067814              1309 ns/op            2031 B/op         15 allocs/op
BenchmarkKVArgs/slog-10                                          4663335              2562 ns/op            3586 B/op         40 allocs/op
BenchmarkKVArgs/slog_with_zlog-10                                9989289              1186 ns/op             978 B/op         22 allocs/op
BenchmarkKVArgs/zlog-10                                         10343742              1171 ns/op             978 B/op         22 allocs/op
BenchmarkKVArgs/slog_with_zap-10                                 7146567              1661 ns/op            1851 B/op         25 allocs/op
BenchmarkKVArgs/zap-10                                           6913246              1730 ns/op            2352 B/op         24 allocs/op
BenchmarkKVArgs/zerolog-10                                       4937739              2431 ns/op            3355 B/op         40 allocs/op
PASS
ok      github.com/icefed/zlog/benchmarks       471.856s
```
