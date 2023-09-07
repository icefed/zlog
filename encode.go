package zlog

import (
	"encoding"
	"encoding/json"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
	"unsafe"

	"golang.org/x/exp/slog"

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

type jsonEncoder struct {
	buf *buffer.Buffer

	je *json.Encoder

	timeFormatter func([]byte, time.Time) []byte
	replaceAttr   func(groups []string, a slog.Attr) slog.Attr
	openGroups    []string
}

func newJSONEncoder(h *JSONHandler, buf *buffer.Buffer) *jsonEncoder {
	return &jsonEncoder{
		buf:           buf,
		timeFormatter: h.c.TimeFormatter,
		openGroups:    h.groups,
		replaceAttr:   h.c.ReplaceAttr,
	}
}

func (enc *jsonEncoder) AppendAttr(a slog.Attr) {
	// If an Attr's key and value are both the zero value, ignore the Attr.
	if a.Equal(slog.Attr{}) {
		return
	}
	if a.Value.Kind() == slog.KindGroup {
		groupAttrs := a.Value.Group()
		// If a group's key is empty, inline the group's Attrs.
		// If a group has no Attrs (even if it has a non-empty key), ignore it.
		emptyGroupOrKey := len(groupAttrs) != 0 && a.Key != ""
		if emptyGroupOrKey {
			enc.OpenGroup(a.Key)
		}
		for i := range groupAttrs {
			enc.AppendAttr(groupAttrs[i])
		}
		if emptyGroupOrKey {
			enc.CloseGroup()
		}
		return
	}

	if enc.replaceAttr != nil {
		a.Value = a.Value.Resolve()
		a = enc.replaceAttr(enc.openGroups, a)
		// If ReplaceAttr returns an Attr with Key == "", the attribute is discarded.
		if a.Key == "" {
			return
		}
	}
	enc.addKey(a.Key)
	a.Value = a.Value.Resolve()
	switch a.Value.Kind() {
	case slog.KindAny:
		enc.addAny(a.Value.Any())
	case slog.KindBool:
		enc.addBool(a.Value.Bool())
	case slog.KindDuration:
		enc.addDuration(a.Value.Duration())
	case slog.KindFloat64:
		enc.addFloat64(a.Value.Float64())
	case slog.KindInt64:
		enc.addInt64(a.Value.Int64())
	case slog.KindString:
		enc.safeAddString(a.Value.String())
	case slog.KindTime:
		if a.Key == slog.TimeKey && len(enc.openGroups) == 0 {
			enc.addTime2Formatter(a.Value.Time())
		} else {
			enc.addTime(a.Value.Time())
		}
	case slog.KindUint64:
		enc.addUint64(a.Value.Uint64())
	default:
		panic(fmt.Sprintf("bad kind: %s", a.Value.Kind()))
	}
}

func (enc *jsonEncoder) AppendTime(key string, t time.Time) {
	if enc.replaceAttr != nil {
		enc.AppendAttr(slog.Time(key, t))
		return
	}
	enc.addKey(key)
	if key == slog.TimeKey && len(enc.openGroups) == 0 {
		enc.addTime2Formatter(t)
	} else {
		enc.addTime(t)
	}
}

func (enc *jsonEncoder) AppendLevel(key string, l slog.Level) {
	if enc.replaceAttr != nil {
		enc.AppendAttr(slog.Any(key, l))
		return
	}
	enc.addKey(key)
	enc.addString(l.String())
}

func (enc *jsonEncoder) AppendString(key string, s string) {
	if enc.replaceAttr != nil {
		enc.AppendAttr(slog.String(key, s))
		return
	}
	enc.addKey(key)
	enc.safeAddString(s)
}

func (enc *jsonEncoder) AppendSourceFromPC(key string, pc uintptr) {
	if enc.replaceAttr != nil {
		enc.AppendAttr(slog.Any(key, buildSource(pc)))
		return
	}
	enc.addKey(key)
	enc.addSourceFromPC(pc)
}

func (enc *jsonEncoder) AppendFormatted(formatted []byte) {
	if len(formatted) == 0 {
		return
	}
	enc.addSeparator()
	enc.buf.Write(formatted)
}

func (enc *jsonEncoder) AppendStacktrace(key string, st *stacktrace) {
	if enc.replaceAttr != nil {
		enc.AppendAttr(slog.Any(key, st))
		return
	}
	enc.addKey(key)
	enc.addStacktrace(st)
}

func (enc *jsonEncoder) OpenGroup(g string) {
	enc.addKey(g)
	enc.buf.WriteByte('{')
	enc.openGroups = append(enc.openGroups, g)
}

func (enc *jsonEncoder) CloseGroup() {
	if len(enc.openGroups) == 0 {
		return
	}
	enc.buf.WriteByte('}')
	enc.openGroups = enc.openGroups[:len(enc.openGroups)-1]
}

func (enc *jsonEncoder) CloseGroups() {
	for i := len(enc.openGroups); i >= 0; i-- {
		enc.CloseGroup()
	}
}

func (enc *jsonEncoder) addAny(v any) {
	// source
	if s, ok := v.(*slog.Source); ok && s != nil {
		enc.addSource(s)
		return
	}

	// stacktrace
	if st, ok := v.(*stacktrace); ok && st != nil {
		enc.addStacktrace(st)
		return
	}

	// error
	if err, ok := v.(error); ok {
		enc.addString(err.Error())
		return
	}

	if enc.je == nil {
		enc.je = json.NewEncoder(&ioWriter{enc.buf})
		enc.je.SetEscapeHTML(false)
	}
	if err := enc.je.Encode(v); err != nil {
		enc.addString(fmt.Sprintf("!ERROR:%v", err))
		return
	}
}

func (enc *jsonEncoder) addSource(s *slog.Source) {
	enc.buf.WriteByte('"')
	formatSourceValue(enc.buf, s)
	enc.buf.WriteByte('"')
}

func (enc *jsonEncoder) addSourceFromPC(pc uintptr) {
	enc.buf.WriteByte('"')
	formatSourceValueFromPC(enc.buf, pc)
	enc.buf.WriteByte('"')
}

func (enc *jsonEncoder) addStacktrace(st *stacktrace) {
	buf := buffer.New()
	defer buf.Free()

	formatStacktrace(buf, st.pc)

	// use unsafe for performance
	enc.safeAddString(unsafe.String(unsafe.SliceData(*buf), buf.Len()))
}

func (enc *jsonEncoder) addBool(b bool) {
	*enc.buf = strconv.AppendBool(*enc.buf, b)
}

func (enc *jsonEncoder) addInt64(i int64) {
	*enc.buf = strconv.AppendInt(*enc.buf, i, 10)
}

func (enc *jsonEncoder) addUint64(i uint64) {
	*enc.buf = strconv.AppendUint(*enc.buf, i, 10)
}

func (enc *jsonEncoder) addFloat64(f float64) {
	data, _ := json.Marshal(f)
	enc.buf.Write(data)
}

func (enc *jsonEncoder) addDuration(d time.Duration) {
	*enc.buf = strconv.AppendInt(*enc.buf, int64(d), 10)
}

func (enc *jsonEncoder) addTime(t time.Time) {
	enc.buf.WriteByte('"')
	*enc.buf = t.AppendFormat(*enc.buf, time.RFC3339Nano)
	enc.buf.WriteByte('"')
}

func (enc *jsonEncoder) addTime2Formatter(t time.Time) {
	enc.buf.WriteByte('"')
	*enc.buf = enc.timeFormatter(*enc.buf, t)
	enc.buf.WriteByte('"')
}

func (enc *jsonEncoder) addString(s string) {
	enc.buf.WriteByte('"')
	enc.buf.WriteString(s)
	enc.buf.WriteByte('"')
}

func (enc *jsonEncoder) safeAddString(s string) {
	jsonEncodeString(enc.buf, s)
}

const hex = "0123456789abcdef"

func jsonEncodeString(buf *buffer.Buffer, s string) {
	buf.WriteByte('"')
	bs := unsafe.Slice(unsafe.StringData(s), len(s))

	start := 0
	for i := 0; i < len(bs); {
		if b := s[i]; b < utf8.RuneSelf {
			if safeSet[b] {
				i++
				continue
			}

			if start < i {
				buf.Write(bs[start:i])
			}
			switch b {
			case '\\', '"':
				buf.WriteByte('\\')
				buf.WriteByte(b)
			case '\t':
				buf.WriteByte('\\')
				buf.WriteByte('t')
			case '\n':
				buf.WriteByte('\\')
				buf.WriteByte('n')
			case '\r':
				buf.WriteByte('\\')
				buf.WriteByte('r')
			default:
				// This encodes bytes < 0x20 except for \n and \r,
				// as well as < and >. The latter are escaped because they
				// can lead to security holes when user-controlled strings
				// are rendered into JSON and served to some browsers.
				buf.WriteString(`\u00`)
				buf.WriteByte(hex[b>>4])
				buf.WriteByte(hex[b&0xF])
			}
			i++
			start = i
			continue
		}
		c, size := utf8.DecodeRune(bs[i:])
		if c == utf8.RuneError && size == 1 {
			if start < i {
				buf.Write(bs[start:i])
			}
			buf.WriteString(`\ufffd`)
			i += size
			start = i
			continue
		}
		// U+2028 is LINE SEPARATOR.
		// U+2029 is PARAGRAPH SEPARATOR.
		// They are both technically valid characters in JSON strings,
		// but don't work in JSONP, which has to be evaluated as JavaScript,
		// and can lead to security holes there. It is valid JSON to
		// escape them, so we do so unconditionally.
		// See http://timelessrepo.com/json-isnt-a-javascript-subset for discussion.
		if c == '\u2028' || c == '\u2029' {
			if start < i {
				buf.Write(bs[start:i])
			}
			buf.WriteString(`\u202`)
			buf.WriteByte(hex[c&0xF])
			i += size
			start = i
			continue
		}
		i += size
	}
	if start < len(s) {
		buf.Write(bs[start:])
	}
	buf.WriteByte('"')
}

var safeSet = [utf8.RuneSelf]bool{
	' ':      true,
	'!':      true,
	'"':      false,
	'#':      true,
	'$':      true,
	'%':      true,
	'&':      true,
	'\'':     true,
	'(':      true,
	')':      true,
	'*':      true,
	'+':      true,
	',':      true,
	'-':      true,
	'.':      true,
	'/':      true,
	'0':      true,
	'1':      true,
	'2':      true,
	'3':      true,
	'4':      true,
	'5':      true,
	'6':      true,
	'7':      true,
	'8':      true,
	'9':      true,
	':':      true,
	';':      true,
	'<':      true,
	'=':      true,
	'>':      true,
	'?':      true,
	'@':      true,
	'A':      true,
	'B':      true,
	'C':      true,
	'D':      true,
	'E':      true,
	'F':      true,
	'G':      true,
	'H':      true,
	'I':      true,
	'J':      true,
	'K':      true,
	'L':      true,
	'M':      true,
	'N':      true,
	'O':      true,
	'P':      true,
	'Q':      true,
	'R':      true,
	'S':      true,
	'T':      true,
	'U':      true,
	'V':      true,
	'W':      true,
	'X':      true,
	'Y':      true,
	'Z':      true,
	'[':      true,
	'\\':     false,
	']':      true,
	'^':      true,
	'_':      true,
	'`':      true,
	'a':      true,
	'b':      true,
	'c':      true,
	'd':      true,
	'e':      true,
	'f':      true,
	'g':      true,
	'h':      true,
	'i':      true,
	'j':      true,
	'k':      true,
	'l':      true,
	'm':      true,
	'n':      true,
	'o':      true,
	'p':      true,
	'q':      true,
	'r':      true,
	's':      true,
	't':      true,
	'u':      true,
	'v':      true,
	'w':      true,
	'x':      true,
	'y':      true,
	'z':      true,
	'{':      true,
	'|':      true,
	'}':      true,
	'~':      true,
	'\u007f': true,
}

func (enc *jsonEncoder) addKey(key string) {
	enc.addSeparator()
	enc.safeAddString(key)
	enc.buf.WriteByte(':')
}

func (enc *jsonEncoder) addSeparator() {
	length := len(enc.buf.Bytes())
	if length == 0 {
		return
	}
	switch enc.buf.Bytes()[length-1] {
	case '{', '[', ':', ',', ' ':
		return
	default:
		enc.buf.WriteByte(',')
	}
}

// ioWriter for json encoder
type ioWriter struct {
	buf *buffer.Buffer
}

func (w *ioWriter) Write(data []byte) (int, error) {
	n := len(data)
	// remove lineEnding char
	if n > 0 && data[n-1] == lineEnding {
		data = data[:n-1]
		n = n - 1
	}
	w.buf.Write(data)
	return n, nil
}

func buildSource(pc uintptr) *slog.Source {
	fs := runtime.CallersFrames([]uintptr{pc})
	f, _ := fs.Next()
	return &slog.Source{
		Function: f.Function,
		File:     f.File,
		Line:     f.Line,
	}
}

func formatSourceValueFromPC(buf *buffer.Buffer, pc uintptr) {
	fs := runtime.CallersFrames([]uintptr{pc})
	f, _ := fs.Next()

	defer func() {
		buf.WriteByte(':')
		*buf = strconv.AppendInt(*buf, int64(f.Line), 10)
	}()

	i := strings.LastIndexByte(f.File, '/')
	if i < 0 {
		buf.WriteString(f.File)
		return
	}
	i = strings.LastIndexByte(f.File[:i], '/')
	if i < 0 {
		buf.WriteString(f.File)
		return
	}
	buf.WriteString(f.File[i+1:])
}

func formatSourceValue(buf *buffer.Buffer, s *slog.Source) {
	defer func() {
		buf.WriteByte(':')
		*buf = strconv.AppendInt(*buf, int64(s.Line), 10)
	}()

	i := strings.LastIndexByte(s.File, '/')
	if i < 0 {
		buf.WriteString(s.File)
		return
	}
	i = strings.LastIndexByte(s.File[:i], '/')
	if i < 0 {
		buf.WriteString(s.File)
		return
	}
	buf.WriteString(s.File[i+1:])
}

// stacktrace define the stack trace source.
// For ReplaceAttr, *stacktrace is the type of value in slog.Attr.
type stacktrace struct {
	pc uintptr
}

type stacktracePCs []uintptr

var stacktracePCsPool = sync.Pool{
	New: func() any {
		pcs := make([]uintptr, 64)
		return (*stacktracePCs)(&pcs)
	},
}

func newStacktracePCs() *stacktracePCs {
	return stacktracePCsPool.Get().(*stacktracePCs)
}

func (st *stacktracePCs) Free() {
	stacktracePCsPool.Put(st)
}

func formatStacktrace(buf *buffer.Buffer, sourcepc uintptr) {
	sfs := runtime.CallersFrames([]uintptr{sourcepc})
	sf, _ := sfs.Next()

	writeCaller := func(fun, file string, line int, more bool) {
		buf.WriteString(fun)
		buf.Write([]byte("\n\t"))
		buf.WriteString(file)
		buf.WriteByte(':')
		*buf = strconv.AppendInt(*buf, int64(line), 10)
		if more {
			buf.WriteByte('\n')
		}
	}

	pcs := *newStacktracePCs()
	defer pcs.Free()
	n := runtime.Callers(1, pcs)
	more := n > 0

	fs := runtime.CallersFrames(pcs[:n])
	var f runtime.Frame
	found := false
	for more {
		f, more = fs.Next()
		if found {
			writeCaller(f.Function, f.File, f.Line, more)
			continue
		}
		if f.Function == sf.Function && f.File == sf.File && f.Line == sf.Line {
			writeCaller(f.Function, f.File, f.Line, more)
			found = true
		}
	}

	if !found {
		// write source
		writeCaller(sf.Function, sf.File, sf.Line, more)
	}
}

const (
	// Color codes for terminal output.
	black   = "\033[30m"
	red     = "\033[31m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	blue    = "\033[34m"
	magenta = "\033[35m"
	cyan    = "\033[36m"
	white   = "\033[37m"
	reset   = "\033[0m"
)

// formatColorLevelValue returns the string representation of the level.
func formatColorLevelValue(buf *buffer.Buffer, l slog.Level) {
	switch {
	case l < slog.LevelInfo: // LevelDebug
		buf.WriteString(magenta)
	case l < slog.LevelWarn: // LevelInfo
		buf.WriteString(blue)
	case l < slog.LevelError: // LevelWarn
		buf.WriteString(yellow)
	default: // LevelError
		buf.WriteString(red)
	}
	buf.WriteString(l.String())
	buf.WriteString(reset)
}
