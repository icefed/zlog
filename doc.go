/*
Package zlog implements JSON-structured logging for slog.Handler interface.
It provides the ability to print stacktrace, a development mode that renders logs
in a user-friendly output like a zap library, flexible custom time formatters and
other features, and almost fully compatible with slog.Handler rules, while
maintaining high performance.

zlog also provides a Logger for better use the formatted logging methods (xxxxf).

# Getting Started

Create a JSONHandler with default [Config].

	handler := zlog.NewJSONHandler(nil)

Use slog.Logger with zlog.JSONHandler.

	log := slog.New(handler)
	log.Info("hello world", slog.String("foo", "bar"))
	log.WithGroup("request").
		With(slog.String("method", "GET")).
		Info("received request", "params", params)

Use zlog.Logger with zlog.JSONHandler.

	log := zlog.New(handler)
	log.Info("hello world", slog.String("foo", "bar"))
	// ...
	log.Errorf("Read file %s failed: %s", filePath, err)

# Use default zlog.Logger

	zlog.Warn("hello world", slog.String("foo", "bar"))

	// set the logger you created as the default.
	log := zlog.New(handler)
	zlog.SetDefault(log)
	zlog.Warn("hello world", slog.String("foo", "bar"))

# Custom Configuration

Create a JSONHandler with custom [Config].

	h := zlog.NewJSONHandler(&zlog.Config{
		HandlerOptions: slog.HandlerOptions{
			// enable AddSource
			AddSource: true,
			// set the level
			Level: slog.LevelDebug,
		}
		// enable development mode
		Development: true,
	})

	// use WithOptions to override the handler Config.
	h = h.WithOptions(zlog.WithStacktraceEnabled(true))
	log:= slog.New(h)

# TraceContext

The Context may contain some values that you want to print in each log. You need
to implement the [ContextExtractor] function, which extracts the value you want
and returns []slog.Attr.

The traceContextExtractor function extracts trace span information from the Context.

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
*/
package zlog
