package zlog

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"testing/slogtest"
	"time"

	cpty "github.com/creack/pty"
)

func TestHandlerSlogtestJson(t *testing.T) {
	var buf bytes.Buffer
	h := NewJSONHandler(&Config{
		HandlerOptions: slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
		IgnoreEmptyGroup: true,
		Writer:           &buf,
	})
	h = h.WithOptions(WithAddSource(true), WithStacktraceEnabled(true), WithStacktraceLevel(slog.LevelDebug), WithStacktraceKey("stacktrace"))

	results := func() []map[string]any {
		var ms []map[string]any
		for _, line := range bytes.Split(buf.Bytes(), []byte{'\n'}) {
			if len(line) == 0 {
				continue
			}
			var m map[string]any
			if err := json.Unmarshal(line, &m); err != nil {
				t.Fatal(err)
			}
			ms = append(ms, m)
		}
		return ms
	}
	err := slogtest.TestHandler(h, results)
	if err != nil {
		t.Fatal(err)
	}
}

func TestHandlerSlogtestDevelopment(t *testing.T) {
	var buf bytes.Buffer
	h := NewJSONHandler(&Config{
		HandlerOptions: slog.HandlerOptions{
			AddSource: true,
		},
		TimeDurationAsInt: true,
		IgnoreEmptyGroup:  true,
		StacktraceEnabled: true,
		StacktraceLevel:   slog.LevelDebug,
	})
	h = h.WithOptions(WithDevelopment(true), WithLevel(slog.LevelDebug), WithWriter(&buf))
	h = h.WithOptions(WithTimeFormatter(func(buf []byte, t time.Time) []byte {
		return t.AppendFormat(buf, RFC3339Milli)
	}))

	results := func() []map[string]any {
		var ms []map[string]any
		for _, line := range bytes.Split(buf.Bytes(), []byte{'\n'}) {
			if len(line) == 0 {
				continue
			}
			// skip stack trace line
			if isStacktraceLine(string(line)) {
				continue
			}

			sepIndex := strings.Index(string(line), "{")
			m := make(map[string]any)
			prefixLen := len(line)
			if sepIndex >= 0 {
				prefixLen = sepIndex
				if err := json.Unmarshal(line[sepIndex:], &m); err != nil {
					t.Fatal(err)
				}
			}
			fields := strings.Fields(string(line[:prefixLen]))
			t, err := time.Parse(RFC3339Milli, fields[0])
			if err == nil {
				m[slog.TimeKey] = t
			}
			for _, field := range fields[1:] {
				switch {
				case isLevelField(field):
					m[slog.LevelKey] = field
				case isSourceField(field):
					m[slog.SourceKey] = field
				default:
					m[slog.MessageKey] = field
				}
			}

			ms = append(ms, m)
		}
		return ms
	}
	err := slogtest.TestHandler(h, results)
	if err != nil {
		t.Fatal(err)
	}
}

func isLevelField(s string) bool {
	switch s {
	case slog.LevelDebug.String(), slog.LevelInfo.String(), slog.LevelWarn.String(), slog.LevelError.String():
		return true
	}
	return false
}

var sourceRegex = regexp.MustCompile(`^(.+\/)?.+.go:\d+$`)

func isSourceField(s string) bool {
	if sourceRegex.MatchString(s) {
		return true
	}
	return false
}

var functionRegex = regexp.MustCompile(`^([?:\w-.]+\/)*[\w.]+$`)
var fileLineRegex = regexp.MustCompile(`^\t\/.*:\d+$`)

func isStacktraceLine(line string) bool {
	if functionRegex.MatchString(line) {
		return true
	}
	if fileLineRegex.MatchString(line) {
		return true
	}
	return false
}

func TestHandlerEmptyConfig(t *testing.T) {
	h := NewJSONHandler(nil)
	log := slog.New(h.WithAttrs([]slog.Attr{}).WithGroup(""))
	if log.Enabled(context.Background(), slog.LevelDebug) {
		t.Error("default config enable debug level")
	}
	if !log.Enabled(context.Background(), slog.LevelError) {
		t.Error("default config not enable error level")
	}

	h2 := NewJSONHandler(&Config{})
	log2 := slog.New(h2)
	if log2.Enabled(context.Background(), slog.LevelDebug) {
		t.Error("default config enable debug level")
	}
	if !log2.Enabled(context.Background(), slog.LevelError) {
		t.Error("default config not enable error level")
	}
}

func TestHandlerColoredLevel(t *testing.T) {
	pty, tty, err := cpty.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer pty.Close()
	defer tty.Close()

	log := slog.New(NewJSONHandler(&Config{
		HandlerOptions: slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelDebug,
		},
		Development: true,
		Writer:      pty,
	}))
	log.Debug("test")
	log.Info("test")
	log.Warn("test")
	log.Error("test")
}

type userKey struct{}
type user struct {
	Name string
	Id   string
}

func userContextExtractor(ctx context.Context) []slog.Attr {
	user, ok := ctx.Value(userKey{}).(user)
	if ok {
		return []slog.Attr{
			slog.Group("user", slog.String("name", user.Name), slog.String("id", user.Id)),
		}
	}
	return nil
}

func TestHandlerContextAttrs(t *testing.T) {
	t.Run("json", func(t *testing.T) {
		var buf bytes.Buffer
		h := NewJSONHandler(&Config{
			HandlerOptions: slog.HandlerOptions{
				AddSource: true,
				Level:     slog.LevelDebug,
			},
			Writer: &buf,
		})
		h = h.WithOptions(WithContextExtractor(userContextExtractor), WithContextExtractor(nil))

		results := func() []map[string]any {
			var ms []map[string]any
			for _, line := range bytes.Split(buf.Bytes(), []byte{'\n'}) {
				if len(line) == 0 {
					continue
				}
				var m map[string]any
				if err := json.Unmarshal(line, &m); err != nil {
					t.Fatal(err)
				}
				ms = append(ms, m)
			}
			return ms
		}
		testContextAttrs(t, h, results)
	})
	t.Run("development", func(t *testing.T) {
		var buf bytes.Buffer
		h := NewJSONHandler(&Config{
			HandlerOptions: slog.HandlerOptions{
				AddSource: true,
				Level:     slog.LevelDebug,
			},
			Development: true,
			Writer:      &buf,
		})
		h = h.WithOptions(WithContextExtractor(userContextExtractor))

		results := func() []map[string]any {
			var ms []map[string]any
			for _, line := range bytes.Split(buf.Bytes(), []byte{'\n'}) {
				if len(line) == 0 {
					continue
				}
				m := make(map[string]any)
				i := strings.Index(string(line), "{")
				if i >= 0 {
					if err := json.Unmarshal(line[i:], &m); err != nil {
						t.Fatal(err)
					}
				}
				ms = append(ms, m)
			}
			return ms
		}
		testContextAttrs(t, h, results)
	})
}

func testContextAttrs(t *testing.T, h *JSONHandler, f func() []map[string]any) {
	log := slog.New(h)
	ctx := context.WithValue(context.Background(), userKey{}, user{
		Name: "test@test.com",
		Id:   "5c81f444-93f9-4cf8-a3b5-c3aeb377a99d",
	})

	tests := []struct {
		name         string
		groups       []string
		ctx          context.Context
		wantKeyPaths []string
		wantKeyValue any
	}{
		{
			name:         "no context value",
			ctx:          context.Background(),
			wantKeyPaths: []string{"test"},
			wantKeyValue: true,
		}, {
			name:         "with context value",
			ctx:          ctx,
			wantKeyPaths: []string{"user", "id"},
			wantKeyValue: "5c81f444-93f9-4cf8-a3b5-c3aeb377a99d",
		}, {
			name:         "group with context value",
			groups:       []string{"g"},
			ctx:          ctx,
			wantKeyPaths: []string{"g", "user", "name"},
			wantKeyValue: "test@test.com",
		},
	}
	for _, test := range tests {
		l := log
		for _, group := range test.groups {
			l = l.WithGroup(group)
		}
		l.InfoContext(test.ctx, "test", slog.Bool("test", true))
	}
	results := f()

	if len(tests) != len(results) {
		t.Errorf("got %v, want %v", len(results), len(tests))
	}
	for i, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var v any = results[i]
			for j := 0; j < len(tests[i].wantKeyPaths); j++ {
				p := tests[i].wantKeyPaths[j]
				vm, ok := v.(map[string]any)
				if !ok {
					t.Errorf("got %v, want map[string]any", v)
				}
				v = vm[p]
			}
			if !reflect.DeepEqual(v, tests[i].wantKeyValue) {
				t.Errorf("got %v, want %v", v, tests[i].wantKeyValue)
			}
		})
	}
}
