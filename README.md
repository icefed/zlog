# zlog - JSON structured handler/logger for Golang slog

[![GoDoc](https://godoc.org/github.com/icefed/zlog?status.svg)](https://pkg.go.dev/github.com/icefed/zlog)
[![Go Report Card](https://goreportcard.com/badge/github.com/icefed/zlog)](https://goreportcard.com/report/github.com/icefed/zlog)
![Build Status](https://github.com/icefed/zlog/actions/workflows/test.yml/badge.svg)
[![Coverage](https://img.shields.io/codecov/c/github/icefed/zlog)](https://codecov.io/gh/icefed/zlog)
[![License](https://img.shields.io/github/license/icefed/zlog)](./LICENSE)

## Features
- JSON Structured logging
- Logger with format method(xxxxf)
- Development mode with human-friendly output
- WithCallerSkip to skip caller
- Context extractor for Record context
- Custom time formatter for buildin attribute time value

## Usage

Because zlog implements the slog.Handler interface, you can create a zlog.JSONHander and use slog.Logger.
```go
import (
    "golang.org/x/exp/slog"
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

	"github.com/icefed/zlog"
	"golang.org/x/exp/slog"
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
	"github.com/icefed/zlog"
	"golang.org/x/exp/slog"
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
![](examples/development.png)

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

The following is an example of printing a user request in http server. The log contains user information and can be used as an audit log.

```go
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ...
		// Pretend that we read and parsed the token, and the user authentication succeeded
		ctx := context.WithValue(context.Background(), userKey{}, user{
			Name: "test@test.com",
			Id:   "a2067a0a-6b0b-4ee5-a049-16bdb8ed6ff5",
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func LogMiddleware(log *zlog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		duration := time.Since(start)
		log.InfoContext(r.Context(), "Received request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("duration", duration.String()),
		)
	})
}

func hello(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello, World!"))
}

func main() {
	h := zlog.NewJSONHandler(&zlog.Config{
		HandlerOptions: slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	})
	h = h.WithOptions(zlog.WithContextExtractor(userContextExtractor))
	log := zlog.New(h)

	httpHandler := http.HandlerFunc(hello)
	// set auth middleware
	handler := AuthMiddleware(httpHandler)
	// set log middleware
	handler = LogMiddleware(log, handler)

	log.Info("starting server, listening on port 8080")
	http.ListenAndServe(":8080", handler)
}

type userKey struct{}
type user struct {
	Name string
	Id   string
}
```

Send a request using curl.
```bash
curl http://localhost:8080/api/v1/products
Hello, World!
```

Outputs:
```
{"time":"2023-09-09T19:51:55.683+08:00","level":"INFO","msg":"starting server, listening on port 8080"}
{"time":"2023-09-09T19:52:04.228+08:00","level":"INFO","msg":"Received request","user":{"name":"test@test.com","id":"a2067a0a-6b0b-4ee5-a049-16bdb8ed6ff5"},"method":"GET","path":"/api/v1/products","duration":"6.221µs"}
```

The example of OpenTelemetry TraceContextExtractor.
[TraceContext](https://pkg.go.dev/github.com/icefed/zlog#example-ContextExtractor-TraceContext)


## Benchmarks
