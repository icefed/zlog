package examples

import (
	"fmt"
	"log/slog"

	"github.com/icefed/zlog"
)

func ExampleStacktrace() {
	h := zlog.NewJSONHandler(&zlog.Config{
		HandlerOptions: slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
		Development:       true,
		StacktraceEnabled: true,
	})

	// set custom stacktrace key
	h = h.WithOptions(zlog.WithStacktraceKey("stack"))

	log := zlog.New(h)
	log.Error("stacktrace", "error", fmt.Errorf("some error"))
}
