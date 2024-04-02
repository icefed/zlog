package examples

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

// This example shows how to use the context, and print OpenTelemetry trace in log.
// If you have an api server that supports OpenTelemetry, you can use this example in your
// handler middleware and print trace in each log.
func ExampleContextExtractorTraceContext() {
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

	// Output like:
	// {"time":"2023-09-08T20:12:14.733","level":"INFO","msg":"call parentFun","trace":{"traceID":"95f0717d9da16177176efdbc7c06bfbd","spanID":"7718edf7b2a8388d"}}
	// {"time":"2023-09-08T20:12:14.733","level":"INFO","msg":"call childFun","trace":{"traceID":"95f0717d9da16177176efdbc7c06bfbd","spanID":"ef83f673951742b0"}}
}
