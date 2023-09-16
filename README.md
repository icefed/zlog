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

Test modified from [zap benchmarking suite](https://github.com/uber-go/zap/tree/master/benchmarks).  Only the zap, zerolog, slog, zlog loggers are retained.

Because zap and zerolog support MarshalObject interface to improve object encode performance, but slog does not support it, so non-MarshalObject objects are also tested here.

```bench
goos: darwin
goarch: arm64
pkg: go.uber.org/zap/benchmarks
BenchmarkDisabledWithoutFields/Zap-8    	1000000000	         0.9374 ns/op	       0 B/op	       0 allocs/op
BenchmarkDisabledWithoutFields/Zap.Sugar-8         	314212434	         7.581 ns/op	      16 B/op	       1 allocs/op
BenchmarkDisabledWithoutFields/Zap.SugarFormatting-8         	45816718	        51.76 ns/op	     136 B/op	       6 allocs/op
BenchmarkDisabledWithoutFields/rs/zerolog-8                  	1000000000	         0.4868 ns/op	       0 B/op	       0 allocs/op
BenchmarkDisabledWithoutFields/slog-8                        	1000000000	         0.9439 ns/op	       0 B/op	       0 allocs/op
BenchmarkDisabledWithoutFields/slog.zaphandler-8             	1000000000	         1.177 ns/op	       0 B/op	       0 allocs/op
BenchmarkDisabledWithoutFields/slog.zloghandler-8            	1000000000	         0.9148 ns/op	       0 B/op	       0 allocs/op
BenchmarkDisabledWithoutFields/zlog-8                        	1000000000	         0.8801 ns/op	       0 B/op	       0 allocs/op
BenchmarkDisabledWithoutFields/zlog.Formatting-8             	43461070	        54.55 ns/op	     136 B/op	       6 allocs/op
BenchmarkDisabledAccumulatedContext/Zap-8                    	1000000000	         0.9258 ns/op	       0 B/op	       0 allocs/op
BenchmarkDisabledAccumulatedContext/Zap.Sugar-8              	295636831	         8.935 ns/op	      16 B/op	       1 allocs/op
BenchmarkDisabledAccumulatedContext/Zap.SugarFormatting-8    	44004669	        56.39 ns/op	     136 B/op	       6 allocs/op
BenchmarkDisabledAccumulatedContext/rs/zerolog-8             	1000000000	         0.4896 ns/op	       0 B/op	       0 allocs/op
BenchmarkDisabledAccumulatedContext/slog-8                   	1000000000	         0.9254 ns/op	       0 B/op	       0 allocs/op
BenchmarkDisabledAccumulatedContext/slog.zaphandler-8        	1000000000	         1.215 ns/op	       0 B/op	       0 allocs/op
BenchmarkDisabledAccumulatedContext/slog.zloghandler-8       	1000000000	         0.9326 ns/op	       0 B/op	       0 allocs/op
BenchmarkDisabledAccumulatedContext/zlog-8                   	1000000000	         0.8999 ns/op	       0 B/op	       0 allocs/op
BenchmarkDisabledAccumulatedContext/zlog.Formatting-8        	43064647	        55.56 ns/op	     136 B/op	       6 allocs/op
BenchmarkDisabledAddingFields/Zap-8                          	13642066	       178.7 ns/op	     736 B/op	       5 allocs/op
BenchmarkDisabledAddingFields/Zap.Sugar-8                    	36251174	        57.95 ns/op	     136 B/op	       6 allocs/op
BenchmarkDisabledAddingFields/rs/zerolog-8                   	177689245	        12.26 ns/op	      24 B/op	       1 allocs/op
BenchmarkDisabledAddingFields/slog-8                         	15710580	       153.7 ns/op	     536 B/op	       6 allocs/op
BenchmarkDisabledAddingFields/slog.zaphandler-8              	17893183	       160.2 ns/op	     536 B/op	       6 allocs/op
BenchmarkDisabledAddingFields/slog.zloghandler-8             	16125266	       152.9 ns/op	     536 B/op	       6 allocs/op
BenchmarkDisabledAddingFields/zlog-8                         	15888432	       150.9 ns/op	     536 B/op	       6 allocs/op
BenchmarkWithoutFields/Zap-8                                 	40908684	        66.80 ns/op	       0 B/op	       0 allocs/op
BenchmarkWithoutFields/Zap.Sugar-8                           	33220770	        79.93 ns/op	      16 B/op	       1 allocs/op
BenchmarkWithoutFields/Zap.SugarFormatting-8                 	 1302277	      1834 ns/op	    1919 B/op	      58 allocs/op
BenchmarkWithoutFields/rs/zerolog-8                          	72975961	        33.81 ns/op	       0 B/op	       0 allocs/op
BenchmarkWithoutFields/rs/zerolog.Formatting-8               	 1327540	      1888 ns/op	    1914 B/op	      58 allocs/op
BenchmarkWithoutFields/slog-8                                	12081782	       199.4 ns/op	       0 B/op	       0 allocs/op
BenchmarkWithoutFields/slog.zaphandler-8                     	19321752	       115.6 ns/op	       0 B/op	       0 allocs/op
BenchmarkWithoutFields/slog.zloghandler-8                    	10709412	       218.6 ns/op	       0 B/op	       0 allocs/op
BenchmarkWithoutFields/zlog-8                                	16306533	       161.8 ns/op	       0 B/op	       0 allocs/op
BenchmarkWithoutFields/zlog.Formatting-8                     	 1406833	      1826 ns/op	    1275 B/op	      57 allocs/op
BenchmarkAccumulatedContext/Zap-8                            	32329981	        65.17 ns/op	       0 B/op	       0 allocs/op
BenchmarkAccumulatedContext/Zap.Sugar-8                      	26341386	        89.47 ns/op	      16 B/op	       1 allocs/op
BenchmarkAccumulatedContext/Zap.SugarFormatting-8            	 1243021	      1869 ns/op	    1923 B/op	      58 allocs/op
BenchmarkAccumulatedContext/rs/zerolog-8                     	63809493	        35.78 ns/op	       0 B/op	       0 allocs/op
BenchmarkAccumulatedContext/rs/zerolog.Formatting-8          	 1332926	      1815 ns/op	    1915 B/op	      58 allocs/op
BenchmarkAccumulatedContext/slog-8                           	11705144	       207.7 ns/op	       0 B/op	       0 allocs/op
BenchmarkAccumulatedContext/slog.zaphandler-8                	21758863	       119.9 ns/op	       0 B/op	       0 allocs/op
BenchmarkAccumulatedContext/slog.zloghandler-8               	13387518	       191.4 ns/op	       0 B/op	       0 allocs/op
BenchmarkAccumulatedContext/zlog-8                           	15348927	       159.3 ns/op	       0 B/op	       0 allocs/op
BenchmarkAccumulatedContext/zlog.Formatting-8                	 1408076	      1787 ns/op	    1275 B/op	      57 allocs/op
BenchmarkAddingFields/Zap-8                                  	 3251718	       726.0 ns/op	     739 B/op	       5 allocs/op
BenchmarkAddingFields/Zap.WithoutObjectMarshal-8             	 1783224	      1398 ns/op	    1416 B/op	      18 allocs/op
BenchmarkAddingFields/Zap.Sugar-8                            	 2353765	      1010 ns/op	    1495 B/op	      10 allocs/op
BenchmarkAddingFields/Zap.Sugar.WithoutObjectMarshal-8       	 1472466	      1666 ns/op	    2175 B/op	      23 allocs/op
BenchmarkAddingFields/rs/zerolog-8                           	 6217977	       399.7 ns/op	      24 B/op	       1 allocs/op
BenchmarkAddingFields/rs/zerolog.WithoutObjectMarshal-8      	 1911603	      1266 ns/op	    1660 B/op	      16 allocs/op
BenchmarkAddingFields/slog-8                                 	 1000000	      2487 ns/op	    3366 B/op	      40 allocs/op
BenchmarkAddingFields/slog.zaphandler-8                      	 1605349	      1459 ns/op	    1376 B/op	      24 allocs/op
BenchmarkAddingFields/slog.zloghandler-8                     	 1625652	      1401 ns/op	    1157 B/op	      22 allocs/op
BenchmarkAddingFields/zlog-8                                 	 1781616	      1411 ns/op	    1157 B/op	      22 allocs/op
PASS
ok  	go.uber.org/zap/benchmarks	146.550s
```
