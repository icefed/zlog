package zlog

import (
	"log/slog"
	"testing"

	"github.com/icefed/zlog/buffer"
)

func TestFormatSourceValue(t *testing.T) {
	tests := []struct {
		name   string
		source slog.Source
		want   string
	}{
		{
			name: "multi level path",
			source: slog.Source{
				File: "logtest/source/value/test.go",
				Line: 12,
			},
			want: "value/test.go:12",
		}, {
			name: "two level path",
			source: slog.Source{
				File: "value/test2.go",
				Line: 15,
			},
			want: "value/test2.go:15",
		}, {
			name: "single file",
			source: slog.Source{
				File: "test3.go",
				Line: 20,
			},
			want: "test3.go:20",
		},
	}

	buf := buffer.New()
	defer buf.Free()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			formatSourceValue(buf, &test.source)
			if string(buf.Bytes()) != test.want {
				t.Errorf("got %v, want %v", string(buf.Bytes()), test.want)
			}
			buf.Reset()
		})
	}
}
