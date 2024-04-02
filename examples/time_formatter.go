package examples

import (
	"log/slog"
	"strconv"
	"time"

	"github.com/icefed/zlog"
)

func ExampleTimeFormatter() {
	h := zlog.NewJSONHandler(&zlog.Config{
		HandlerOptions: slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
		TimeFormatter: func(buf []byte, t time.Time) []byte {
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
}
