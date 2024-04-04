module github.com/icefed/zlog/benchmarks

go 1.21

replace github.com/icefed/zlog => ../

require (
	github.com/icefed/zlog v0.0.0
	github.com/rs/zerolog v1.30.0
	go.uber.org/zap v1.26.0
	go.uber.org/zap/exp v0.2.0
)

require (
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/term v0.18.0 // indirect
)
