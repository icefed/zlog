package zlog

import (
	"encoding"
	"fmt"
	"log/slog"
	"time"

	"github.com/icefed/zlog/buffer"
)

// textEncoder encode buildin attributes for development mode.
type textEncoder struct {
	buf *buffer.Buffer

	coloredLevel  bool
	timeFormatter func([]byte, time.Time) []byte
	replaceAttr   func(groups []string, a slog.Attr) slog.Attr
}

func newTextEncoder(h *JSONHandler, buf *buffer.Buffer) *textEncoder {
	return &textEncoder{
		buf:           buf,
		coloredLevel:  h.needColoredLevel(),
		timeFormatter: h.c.TimeFormatter,
		replaceAttr:   h.c.ReplaceAttr,
	}
}

func (enc *textEncoder) Append(key string, v any) {
	if enc.replaceAttr != nil {
		var a slog.Attr
		switch v := v.(type) {
		// source PC
		case uintptr:
			a = enc.replaceAttr(nil, slog.Any(key, buildSource(v)))
		default:
			a = enc.replaceAttr(nil, slog.Any(key, v))
		}
		if a.Key == "" {
			return
		}
		enc.addValue(a.Value)
		return
	}
	// source PC
	switch v := v.(type) {
	// source PC
	case uintptr:
		formatSourceValueFromPC(enc.buf, v)
	default:
		enc.addValue(slog.AnyValue(v))
	}
}

func (enc *textEncoder) addValue(v slog.Value) {
	v = v.Resolve()
	switch v.Kind() {
	case slog.KindString:
		enc.buf.WriteString(v.String())
	case slog.KindTime:
		*enc.buf = enc.timeFormatter(*enc.buf, v.Time())
	case slog.KindAny:
		if l, ok := v.Any().(slog.Level); ok {
			if enc.coloredLevel {
				formatColorLevelValue(enc.buf, l)
			} else {
				enc.buf.WriteString(l.String())
			}
			return
		}
		if s, ok := v.Any().(*slog.Source); ok && s != nil {
			formatSourceValue(enc.buf, s)
			return
		}
		if st, ok := v.Any().(*stacktrace); ok && st != nil {
			formatStacktrace(enc.buf, st.pc)
			return
		}
		if tm, ok := v.Any().(encoding.TextMarshaler); ok {
			data, err := tm.MarshalText()
			if err != nil {
				*enc.buf = fmt.Appendf(*enc.buf, "!ERROR:%v", err)
				return
			}
			enc.buf.Write(data)
			return
		}
		*enc.buf = fmt.Appendf(*enc.buf, "%+v", v.Any())
	default:
		*enc.buf = fmt.Appendf(*enc.buf, "%+v", v.Any())
	}
}
