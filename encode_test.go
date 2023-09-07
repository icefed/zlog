package zlog

import (
	"bytes"
	"io/fs"
	"net"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/icefed/zlog/buffer"
	"golang.org/x/exp/slog"
)

var testTime = time.Date(2023, 8, 16, 1, 2, 3, 666666666, time.UTC)

func TestTextEncoder(t *testing.T) {
	t.Run("no_replace", func(t *testing.T) {
		h := NewJSONHandler(&Config{
			HandlerOptions: slog.HandlerOptions{
				AddSource: true,
				Level:     slog.LevelDebug,
			},
		})
		buf := buffer.New()
		defer buf.Free()
		enc := newTextEncoder(h, buf)

		tests := []struct {
			key   string
			value any
			want  string
		}{
			{
				key:   slog.TimeKey,
				value: testTime,
				want:  "2023-08-16T01:02:03.666Z",
			}, {
				key:   slog.LevelKey,
				value: slog.LevelInfo,
				want:  "INFO",
			}, {
				key:   slog.MessageKey,
				value: "test msg",
				want:  "test msg",
			}, {
				key:   "error",
				value: fs.ErrNotExist,
				want:  "file does not exist",
			}, {
				key: slog.SourceKey,
				value: &slog.Source{
					File: "test.go",
					Line: 300,
				},
				want: "test.go:300",
			}, {
				key:   "stacktrace",
				value: &stacktrace{getPC()},
				want:  wantPCFunction + "\n\t" + wantPCFile + ":" + strconv.Itoa(wantPCLine),
			}, {
				key:   "ip",
				value: net.ParseIP("127.0.0.1"),
				want:  "127.0.0.1",
			},
		}
		for _, test := range tests {
			enc.Append(test.key, test.value)
			if string(buf.Bytes()) != test.want {
				t.Errorf("got %v, want %v", string(buf.Bytes()), test.want)
			}
			buf.Reset()
		}
	})
	t.Run("replace", func(t *testing.T) {
		h := NewJSONHandler(&Config{
			HandlerOptions: slog.HandlerOptions{
				AddSource: true,
				Level:     slog.LevelDebug,
			},
		})
		buf := buffer.New()
		defer buf.Free()

		tests := []struct {
			key         string
			value       any
			want        string
			replaceAttr func(_ []string, a slog.Attr) slog.Attr
		}{
			{
				key:   slog.LevelKey,
				value: slog.LevelDebug,
				want:  "",
				replaceAttr: func(_ []string, a slog.Attr) slog.Attr {
					if a.Key == slog.LevelKey && a.Value.Any().(slog.Level) == slog.LevelDebug {
						return slog.Attr{}
					}
					return a
				},
			}, {
				key:   slog.LevelKey,
				value: slog.LevelError,
				want:  "20",
				replaceAttr: func(_ []string, a slog.Attr) slog.Attr {
					if a.Key == slog.LevelKey && a.Value.Any().(slog.Level) == slog.LevelError {
						a.Value = slog.AnyValue(20)
						return a
					}
					return a
				},
			}, {
				key:   "replacepc",
				value: getPC(),
				want:  `replacedpc`,
				replaceAttr: func(_ []string, a slog.Attr) slog.Attr {
					if a.Key == "replacepc" {
						a.Value = slog.AnyValue("replacedpc")
					}
					return a
				},
			},
		}
		for _, test := range tests {
			if test.replaceAttr != nil {
				h = h.WithOptions(WithReplaceAttr(test.replaceAttr))
			}
			enc := newTextEncoder(h, buf)
			enc.Append(test.key, test.value)
			if string(buf.Bytes()) != test.want {
				t.Errorf("got %v, want %v", string(buf.Bytes()), test.want)
			}
			buf.Reset()
		}
	})
}

func TestJSONEncoder(t *testing.T) {
	t.Run("no_replace", func(t *testing.T) {
		h := NewJSONHandler(&Config{
			HandlerOptions: slog.HandlerOptions{
				AddSource: true,
				Level:     slog.LevelDebug,
			},
		})
		buf := buffer.New()
		defer buf.Free()
		enc := newJSONEncoder(h, buf)

		tests := []struct {
			groups []string
			key    string
			value  any
			want   string
		}{
			{
				key:   slog.TimeKey,
				value: testTime,
				want:  `"time":"2023-08-16T01:02:03.666Z"`,
			}, {
				groups: []string{"g"},
				key:    slog.LevelKey,
				value:  slog.LevelInfo,
				want:   `"g":{"level":"INFO"`,
			}, {
				key:   slog.MessageKey,
				value: "test msg",
				want:  `"msg":"test msg"`,
			}, {
				groups: []string{"g", "g2"},
				key:    "error",
				value:  fs.ErrNotExist,
				want:   `"g":{"g2":{"error":"file does not exist"`,
			}, {
				key: slog.SourceKey,
				value: &slog.Source{
					File: "test.go",
					Line: 300,
				},
				want: `"source":"test.go:300"`,
			}, {
				key:   "stacktrace",
				value: &stacktrace{getPC()},
				want:  `"stacktrace":"` + wantPCFunction + "\\n\\t" + wantPCFile + ":" + strconv.Itoa(wantPCLine) + `"`,
			}, {
				key:   "ip",
				value: net.ParseIP("127.0.0.1"),
				want:  `"ip":"127.0.0.1"`,
			}, {
				key:   "formatted",
				value: []byte(`"env":"prod","app":"web"`),
				want:  `"env":"prod","app":"web"`,
			}, {
				key:   "group",
				value: slog.GroupValue(slog.Attr{Key: "env", Value: slog.StringValue("prod")}, slog.Attr{Key: "app", Value: slog.StringValue("web")}),
				want:  `"group":{"env":"prod","app":"web"}`,
			},
			{key: "bool", value: true, want: `"bool":true`},
			{key: "duration", value: time.Second * 10, want: `"duration":10000000000`},
			{key: "float64", value: float64(0.32), want: `"float64":0.32`},
			{key: "int64", value: int64(32), want: `"int64":32`},
			{key: "uint64", value: uint64(111), want: `"uint64":111`},
			{key: "string", value: "stringvalue", want: `"string":"stringvalue"`},
			{
				key:   "escapedstring",
				value: "😀z\nz\rz\tz\"z\\z\u2028z\x00z\xff",
				want:  `"escapedstring":"😀z\nz\rz\tz\"z\\z\u2028z\u0000z\ufffd"`,
			},
		}
		for _, test := range tests {
			for _, group := range test.groups {
				enc.OpenGroup(group)
			}
			switch v := test.value.(type) {
			case time.Time:
				enc.AppendTime(test.key, v)
			case slog.Level:
				enc.AppendLevel(test.key, v)
			case string:
				enc.AppendString(test.key, v)
			case *stacktrace:
				enc.AppendStacktrace(test.key, v)
			case []byte:
				enc.AppendFormatted(v)
			default:
				enc.AppendAttr(slog.Attr{
					Key:   test.key,
					Value: slog.AnyValue(v),
				})
			}
			if string(buf.Bytes()) != test.want {
				t.Errorf("got %v, want %v", string(buf.Bytes()), test.want)
			}
			enc.CloseGroups()
			buf.Reset()
		}
	})
	t.Run("replace", func(t *testing.T) {
		h := NewJSONHandler(&Config{
			HandlerOptions: slog.HandlerOptions{
				AddSource: true,
				Level:     slog.LevelDebug,
			},
		})
		buf := buffer.New()
		defer buf.Free()

		tests := []struct {
			groups      []string
			key         string
			value       any
			want        string
			replaceAttr func(_ []string, a slog.Attr) slog.Attr
		}{
			{
				key:   slog.LevelKey,
				value: slog.LevelDebug,
				want:  "",
				replaceAttr: func(_ []string, a slog.Attr) slog.Attr {
					if a.Key == slog.LevelKey && a.Value.Any().(slog.Level) == slog.LevelDebug {
						return slog.Attr{}
					}
					return a
				},
			}, {
				key:   slog.LevelKey,
				value: slog.LevelError,
				want:  `"level":"ERROR+1"`,
				replaceAttr: func(_ []string, a slog.Attr) slog.Attr {
					if a.Key == slog.LevelKey && a.Value.Any().(slog.Level) == slog.LevelError {
						a.Value = slog.AnyValue(slog.LevelError + 1)
						return a
					}
					return a
				},
			}, {
				key:   "timestring",
				value: "2023-01-01T01:02:03.666Z",
				want:  `"timestring":"2023-08-16T01:02:03.666666666Z"`,
				replaceAttr: func(_ []string, a slog.Attr) slog.Attr {
					if a.Key == "timestring" {
						a.Value = slog.TimeValue(testTime)
					}
					return a
				},
			}, {
				key:   "time",
				value: testTime,
				want:  `"time":"2023-08-16T01:02:03.666Z"`,
			}, {
				groups: []string{"g"},
				key:    "time",
				value:  testTime,
				want:   `"g":{"time":"2023-08-16T01:02:03.666666666Z"`,
			}, {
				key: slog.SourceKey,
				value: &slog.Source{
					File: "test.go",
					Line: 300,
				},
				want: `"source":"test.go:300"`,
			}, {
				key:   "replacepc",
				value: getPC(),
				want:  `"replacepc":"replacedpc"`,
				replaceAttr: func(_ []string, a slog.Attr) slog.Attr {
					if a.Key == "replacepc" {
						a.Value = slog.StringValue("replacedpc")
					}
					return a
				},
			}, {
				key:   "stacktrace",
				value: &stacktrace{getPC()},
				want:  `"stacktrace":"` + wantPCFunction + "\\n\\t" + wantPCFile + ":" + strconv.Itoa(wantPCLine) + `"`,
			}, {
				key:   "group",
				value: slog.GroupValue(slog.Attr{Key: "env", Value: slog.StringValue("prod")}, slog.Attr{Key: "app", Value: slog.StringValue("web")}),
				want:  `"group":{"env":"prod","app":"web"}`,
			}, {
				groups: []string{"s1", "s2"},
				key:    "string",
				value:  "stringvalue",
				want:   `"s1":{"s2":{"string":"replacedstring"`,
				replaceAttr: func(_ []string, a slog.Attr) slog.Attr {
					if a.Key == "string" {
						a.Value = slog.StringValue("replacedstring")
					}
					return a
				},
			},
		}
		for _, test := range tests {
			if test.replaceAttr != nil {
				h = h.WithOptions(WithReplaceAttr(test.replaceAttr))
			}
			enc := newJSONEncoder(h, buf)
			for _, group := range test.groups {
				enc.OpenGroup(group)
			}
			switch v := test.value.(type) {
			case time.Time:
				enc.AppendTime(test.key, v)
			case slog.Level:
				enc.AppendLevel(test.key, v)
			case string:
				enc.AppendString(test.key, v)
			case uintptr:
				enc.AppendSourceFromPC(test.key, v)
			case *stacktrace:
				enc.AppendStacktrace(test.key, v)
			case []byte:
				enc.AppendFormatted(v)
			default:
				enc.AppendAttr(slog.Attr{
					Key:   test.key,
					Value: slog.AnyValue(v),
				})
			}
			if string(buf.Bytes()) != test.want {
				t.Errorf("got %v, want %v", string(buf.Bytes()), test.want)
			}
			enc.CloseGroups()
			buf.Reset()
		}
	})
}

func TestFormatSourceValue(t *testing.T) {
	tests := []struct {
		source slog.Source
		want   string
	}{
		{
			source: slog.Source{
				File: "logtest/source/value/test.go",
				Line: 12,
			},
			want: "value/test.go:12",
		}, {
			source: slog.Source{
				File: "value/test2.go",
				Line: 15,
			},
			want: "value/test2.go:15",
		}, {
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
		formatSourceValue(buf, &test.source)
		if string(buf.Bytes()) != test.want {
			t.Errorf("got %v, want %v", string(buf.Bytes()), test.want)
		}
		buf.Reset()
	}
}

var (
	wantPCFunction string
	wantPCFile     string
	wantPCLine     int

	getPC = sync.OnceValue(func() uintptr {
		pcs := make([]uintptr, 1)
		runtime.Callers(1, pcs)

		sfs := runtime.CallersFrames(pcs)
		sf, _ := sfs.Next()
		wantPCFunction = sf.Function
		wantPCFile = sf.File
		wantPCLine = sf.Line
		return pcs[0]
	})
)

func stacktraceCaller2(buf *buffer.Buffer) {
	pcs := make([]uintptr, 1)
	runtime.Callers(2, pcs)
	formatStacktrace(buf, pcs[0])
}

func stacktraceCaller1(buf *buffer.Buffer) {
	stacktraceCaller2(buf)
}

func TestFormatStacktrace(t *testing.T) {
	buf := buffer.New()
	defer buf.Free()

	stacktraceCaller1(buf)
	lines := bytes.Split(buf.Bytes(), []byte{'\n'})
	if len(lines) != 8 {
		t.Errorf("got %v, want %v", len(lines), 8)
	}
	buf.Reset()

	// not found
	pcs := make([]uintptr, 1)
	runtime.Callers(1, pcs)
	formatStacktrace(buf, pcs[0])
	lines = bytes.Split(buf.Bytes(), []byte{'\n'})
	if len(lines) != 2 {
		t.Errorf("got %v, want %v", len(lines), 2)
	}
}

func TestFormatColorLevelValue(t *testing.T) {
	tests := []struct {
		level slog.Level
		want  []byte
	}{
		{
			level: slog.LevelDebug,
			want:  []byte("\033[35mDEBUG\033[0m"),
		}, {
			level: slog.LevelDebug - 1,
			want:  []byte("\033[35mDEBUG-1\033[0m"),
		}, {
			level: slog.LevelDebug + 2,
			want:  []byte("\033[35mDEBUG+2\033[0m"),
		}, {
			level: slog.LevelInfo,
			want:  []byte("\033[34mINFO\033[0m"),
		}, {
			level: slog.LevelInfo + 5,
			want:  []byte("\033[33mWARN+1\033[0m"),
		}, {
			level: slog.LevelWarn,
			want:  []byte("\033[33mWARN\033[0m"),
		}, {
			level: slog.LevelError,
			want:  []byte("\033[31mERROR\033[0m"),
		}, {
			level: slog.LevelError + 100,
			want:  []byte("\033[31mERROR+100\033[0m"),
		},
	}

	buf := buffer.New()
	defer buf.Free()

	for _, test := range tests {
		formatColorLevelValue(buf, test.level)
		if !bytes.Equal(buf.Bytes(), test.want) {
			t.Errorf("got %v, want %v", string(buf.Bytes()), string(test.want))
		}
		buf.Reset()
	}
}
