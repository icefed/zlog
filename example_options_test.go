package zlog_test

import (
	"github.com/icefed/zlog"
)

// WithOptions can be used in handler and logger.
func ExampleOption_withOptions() {
	// use WithOptions in handler
	// first, create a default handler
	h := zlog.NewJSONHandler(nil)

	// set options for handler
	h = h.WithOptions(
		zlog.WithAddSource(true),
		zlog.WithStacktraceEnabled(true),
		zlog.WithDevelopment(true),
	)

	// use WithOptions in logger
	log := zlog.New(h)

	// set options for logger
	log = log.WithOptions(
		zlog.WithAddSource(false),
	)

	// ...
}
