package examples

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/icefed/zlog"
)

func ExampleSlogWithZlogHandler() {
	h := zlog.NewJSONHandler(&zlog.Config{
		HandlerOptions: slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	})
	log := slog.New(h)

	log.Info("hello world")
	// ...
	log.With(slog.String("app", "test")).
		Error("db execution failed", "error", fmt.Errorf("some error"))
	log.LogAttrs(context.Background(), slog.LevelInfo, "this is a info message", slog.String("app", "test"))
}
