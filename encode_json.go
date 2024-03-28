package zlog

import (
	"encoding"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
	"strconv"
	"time"
	"unicode"

	"github.com/icefed/zlog/buffer"
)

type jsonEncoder struct {
	buf *buffer.Buffer

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
	if enc.replaceAttr != nil && a.Value.Kind() != slog.KindGroup {
		a.Value = a.Value.Resolve()
		a = enc.replaceAttr(enc.openGroups, a)
		// If ReplaceAttr returns an Attr with Key == "", the attribute is discarded.
		if a.Key == "" {
			return
		}
	}
	a.Value = a.Value.Resolve()
	// If an Attr's key and value are both the zero value, ignore the Attr.
	if a.Equal(slog.Attr{}) {
		return
	}
	if a.Value.Kind() == slog.KindGroup {
		groupAttrs := a.Value.Group()
		// If a group's key is empty, inline the group's Attrs.
		// If a group has no Attrs (even if it has a non-empty key), ignore it.
		notEmptyGroupOrKey := len(groupAttrs) != 0 && a.Key != ""
		if notEmptyGroupOrKey {
			enc.OpenGroup(a.Key)
		}
		for i := range groupAttrs {
			enc.AppendAttr(groupAttrs[i])
		}
		if notEmptyGroupOrKey {
			enc.CloseGroup()
		}
		return
	}

	enc.addKey(a.Key)
	enc.addValue(a.Value)
}

func (enc *jsonEncoder) AppendTime(key string, t time.Time) {
	if enc.replaceAttr != nil {
		attr := enc.replaceAttr(enc.openGroups, slog.Time(key, t))
		// If ReplaceAttr returns an Attr with Key == "", the attribute is discarded.
		if attr.Key == "" {
			return
		}
		enc.addKey(key)
		attr.Value = attr.Value.Resolve()
		if attr.Value.Kind() == slog.KindTime {
			enc.addTime2Formatter(t)
		} else {
			enc.addValue(attr.Value)
		}
		return
	}
	enc.addKey(key)
	enc.addTime2Formatter(t)
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

// addValue handle slog.Value except slog.KindLogValuer and slog.KindGroup
func (enc *jsonEncoder) addValue(v slog.Value) {
	switch v.Kind() {
	case slog.KindAny:
		enc.addAny(v.Any())
	case slog.KindBool:
		enc.addBool(v.Bool())
	case slog.KindDuration:
		enc.addDuration(v.Duration())
	case slog.KindFloat64:
		enc.addFloat64(v.Float64())
	case slog.KindInt64:
		enc.addInt64(v.Int64())
	case slog.KindString:
		enc.safeAddString(v.String())
	case slog.KindTime:
		enc.addTime(v.Time())
	case slog.KindUint64:
		enc.addUint64(v.Uint64())
	default:
		panic(fmt.Sprintf("bad kind: %s", v.Kind()))
	}
}

func (enc *jsonEncoder) addAny(v any) {
	switch v := v.(type) {
	case *slog.Source: // source
		if isNil(v) {
			enc.addNil()
			return
		}
		enc.addSource(v)
	case *stacktrace: // stacktrace
		if isNil(v) {
			enc.addNil()
			return
		}
		enc.addStacktrace(v)
	case *time.Time:
		if v == nil {
			enc.addNil()
			return
		}
		enc.addTime(*v)
	case *time.Duration:
		if v == nil {
			enc.addNil()
			return
		}
		enc.addDuration(*v)
	case *string:
		if v == nil {
			enc.addNil()
			return
		}
		enc.safeAddString(*v)
	case *bool:
		if v == nil {
			enc.addNil()
			return
		}
		enc.addBool(*v)
	case *int:
		if v == nil {
			enc.addNil()
			return
		}
		enc.addInt64(int64(*v))
	case *int8:
		if v == nil {
			enc.addNil()
			return
		}
		enc.addInt64(int64(*v))
	case *int16:
		if v == nil {
			enc.addNil()
			return
		}
		enc.addInt64(int64(*v))
	case *int32:
		if v == nil {
			enc.addNil()
			return
		}
		enc.addInt64(int64(*v))
	case *int64:
		if v == nil {
			enc.addNil()
			return
		}
		enc.addInt64(*v)
	case *uint:
		if v == nil {
			enc.addNil()
			return
		}
		enc.addUint64(uint64(*v))
	case *uint8:
		if v == nil {
			enc.addNil()
			return
		}
		enc.addUint64(uint64(*v))
	case *uint16:
		if v == nil {
			enc.addNil()
			return
		}
		enc.addUint64(uint64(*v))
	case *uint32:
		if v == nil {
			enc.addNil()
			return
		}
		enc.addUint64(uint64(*v))
	case *uint64:
		if v == nil {
			enc.addNil()
			return
		}
		enc.addUint64(*v)
	case *float32:
		if v == nil {
			enc.addNil()
			return
		}
		enc.addFloat32(float64(*v))
	case *float64:
		if v == nil {
			enc.addNil()
			return
		}
		enc.addFloat64(*v)
	case []error:
		enc.addErrorArray(v)
	case []byte:
		enc.addBytes(v)
	case []string:
		enc.addStringArray(v)
	case []bool:
		enc.addBoolArray(v)
	case []int:
		enc.addIntArray(v)
	case []int8:
		enc.addInt8Array(v)
	case []int16:
		enc.addInt16Array(v)
	case []int32:
		enc.addInt32Array(v)
	case []int64:
		enc.addInt64Array(v)
	case []uint:
		enc.addUintArray(v)
	// case []uint8: // same as []byte
	case []uint16:
		enc.addUint16Array(v)
	case []uint32:
		enc.addUint32Array(v)
	case []uint64:
		enc.addUint64Array(v)
	case []float32:
		enc.addFloat32Array(v)
	case []float64:
		enc.addFloat64Array(v)
	case []time.Duration:
		enc.addDurationArray(v)
	case []time.Time:
		enc.addTimeArray(v)
	case json.Marshaler: // json.Marshaler
		if isNil(v) {
			enc.addNil()
			return
		}
		data, err := v.MarshalJSON()
		if err != nil {
			enc.safeAddString(fmt.Sprintf("!ERROR:%v", err))
			return
		}
		if !json.Valid(data) {
			enc.safeAddString(fmt.Sprintf("!ERROR: invalid MarshalJSON output:%v", v))
			return
		}
		enc.addRawMessage(data)
	case encoding.TextMarshaler: // encoding.TextMarshaler
		if isNil(v) {
			enc.addNil()
			return
		}
		data, err := v.MarshalText()
		if err != nil {
			enc.safeAddString(fmt.Sprintf("!ERROR:%v", err))
			return
		}
		enc.safeAddString((*buffer.Buffer)(&data).String())
	case error: // handle error after json.Marshaler
		if isNil(v) {
			enc.addNil()
			return
		}
		enc.safeAddString(v.Error())
	default:
		je := json.NewEncoder(&ioWriter{enc.buf})
		je.SetEscapeHTML(false)
		if err := je.Encode(v); err != nil {
			enc.safeAddString(fmt.Sprintf("!ERROR:%v", err))
		}
	}
}

func isNil(v any) bool {
	return v == nil || (reflect.ValueOf(v).Kind() == reflect.Ptr && reflect.ValueOf(v).IsNil())
}

func (enc *jsonEncoder) addErrorArray(errors []error) {
	if errors == nil {
		enc.addNil()
		return
	}
	enc.buf.WriteByte('[')
	for i, err := range errors {
		if i > 0 {
			enc.buf.WriteByte(',')
		}
		if isNil(err) {
			enc.addNil()
		} else {
			enc.safeAddString(err.Error())
		}
	}
	enc.buf.WriteByte(']')
}

func (enc *jsonEncoder) addBytes(bytes []byte) {
	if len(bytes) == 0 {
		enc.addString("")
		return
	}

	printable := true
	for _, r := range (*buffer.Buffer)(&bytes).String() {
		if !unicode.IsPrint(r) {
			printable = false
			break
		}
	}
	if printable {
		enc.safeAddString((*buffer.Buffer)(&bytes).String())
		return
	}
	// not printable, encode as base64
	enc.buf.WriteByte('"')
	encodedLen := base64.StdEncoding.EncodedLen(len(bytes))
	enc.buf.Grow(encodedLen + 1)
	base64.StdEncoding.Encode((*enc.buf)[enc.buf.Len()-encodedLen-1:], bytes)
	enc.buf.Truncate(enc.buf.Len() - 1)
	enc.buf.WriteByte('"')
}

func (enc *jsonEncoder) addStringArray(arr []string) {
	if arr == nil {
		enc.addNil()
		return
	}
	enc.buf.WriteByte('[')
	for i, s := range arr {
		if i > 0 {
			enc.buf.WriteByte(',')
		}
		enc.safeAddString(s)
	}
	enc.buf.WriteByte(']')
}

func (enc *jsonEncoder) addBoolArray(arr []bool) {
	if arr == nil {
		enc.addNil()
		return
	}
	enc.buf.WriteByte('[')
	for i, b := range arr {
		if i > 0 {
			enc.buf.WriteByte(',')
		}
		enc.addBool(b)
	}
	enc.buf.WriteByte(']')
}

func (enc *jsonEncoder) addIntArray(arr []int) {
	if arr == nil {
		enc.addNil()
		return
	}
	enc.buf.WriteByte('[')
	for i, n := range arr {
		if i > 0 {
			enc.buf.WriteByte(',')
		}
		enc.addInt64(int64(n))
	}
	enc.buf.WriteByte(']')
}

func (enc *jsonEncoder) addInt8Array(arr []int8) {
	if arr == nil {
		enc.addNil()
		return
	}
	enc.buf.WriteByte('[')
	for i, n := range arr {
		if i > 0 {
			enc.buf.WriteByte(',')
		}
		enc.addInt64(int64(n))
	}
	enc.buf.WriteByte(']')
}

func (enc *jsonEncoder) addInt16Array(arr []int16) {
	if arr == nil {
		enc.addNil()
		return
	}
	enc.buf.WriteByte('[')
	for i, n := range arr {
		if i > 0 {
			enc.buf.WriteByte(',')
		}
		enc.addInt64(int64(n))
	}
	enc.buf.WriteByte(']')
}

func (enc *jsonEncoder) addInt32Array(arr []int32) {
	if arr == nil {
		enc.addNil()
		return
	}
	enc.buf.WriteByte('[')
	for i, n := range arr {
		if i > 0 {
			enc.buf.WriteByte(',')
		}
		enc.addInt64(int64(n))
	}
	enc.buf.WriteByte(']')
}

func (enc *jsonEncoder) addInt64Array(arr []int64) {
	if arr == nil {
		enc.addNil()
		return
	}
	enc.buf.WriteByte('[')
	for i, n := range arr {
		if i > 0 {
			enc.buf.WriteByte(',')
		}
		enc.addInt64(n)
	}
	enc.buf.WriteByte(']')
}

func (enc *jsonEncoder) addUintArray(arr []uint) {
	if arr == nil {
		enc.addNil()
		return
	}
	enc.buf.WriteByte('[')
	for i, n := range arr {
		if i > 0 {
			enc.buf.WriteByte(',')
		}
		enc.addUint64(uint64(n))
	}
	enc.buf.WriteByte(']')
}

func (enc *jsonEncoder) addUint16Array(arr []uint16) {
	if arr == nil {
		enc.addNil()
		return
	}
	enc.buf.WriteByte('[')
	for i, n := range arr {
		if i > 0 {
			enc.buf.WriteByte(',')
		}
		enc.addUint64(uint64(n))
	}
	enc.buf.WriteByte(']')
}

func (enc *jsonEncoder) addUint32Array(arr []uint32) {
	if arr == nil {
		enc.addNil()
		return
	}
	enc.buf.WriteByte('[')
	for i, n := range arr {
		if i > 0 {
			enc.buf.WriteByte(',')
		}
		enc.addUint64(uint64(n))
	}
	enc.buf.WriteByte(']')
}

func (enc *jsonEncoder) addUint64Array(arr []uint64) {
	if arr == nil {
		enc.addNil()
		return
	}
	enc.buf.WriteByte('[')
	for i, n := range arr {
		if i > 0 {
			enc.buf.WriteByte(',')
		}
		enc.addUint64(n)
	}
	enc.buf.WriteByte(']')
}

func (enc *jsonEncoder) addFloat32Array(arr []float32) {
	if arr == nil {
		enc.addNil()
		return
	}
	enc.buf.WriteByte('[')
	for i, n := range arr {
		if i > 0 {
			enc.buf.WriteByte(',')
		}
		enc.addFloat32(float64(n))
	}
	enc.buf.WriteByte(']')
}

func (enc *jsonEncoder) addFloat64Array(arr []float64) {
	if arr == nil {
		enc.addNil()
		return
	}
	enc.buf.WriteByte('[')
	for i, n := range arr {
		if i > 0 {
			enc.buf.WriteByte(',')
		}
		enc.addFloat64(n)
	}
	enc.buf.WriteByte(']')
}

func (enc *jsonEncoder) addDurationArray(arr []time.Duration) {
	if arr == nil {
		enc.addNil()
		return
	}
	enc.buf.WriteByte('[')
	for i, d := range arr {
		if i > 0 {
			enc.buf.WriteByte(',')
		}
		enc.addDuration(d)
	}
	enc.buf.WriteByte(']')
}

func (enc *jsonEncoder) addTimeArray(arr []time.Time) {
	if arr == nil {
		enc.addNil()
		return
	}
	enc.buf.WriteByte('[')
	for i, t := range arr {
		if i > 0 {
			enc.buf.WriteByte(',')
		}
		enc.addTime(t)
	}
	enc.buf.WriteByte(']')
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

	enc.safeAddString(buf.String())
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

func (enc *jsonEncoder) addFloat32(f float64) {
	*enc.buf = strconv.AppendFloat(*enc.buf, f, 'f', -1, 32)
}

func (enc *jsonEncoder) addFloat64(f float64) {
	*enc.buf = strconv.AppendFloat(*enc.buf, f, 'f', -1, 64)
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

func (enc *jsonEncoder) addNil() {
	enc.buf.WriteString("null")
}

func (enc *jsonEncoder) addRawMessage(data []byte) {
	enc.buf.Write(data)
}

func (enc *jsonEncoder) addString(s string) {
	enc.buf.WriteByte('"')
	enc.buf.WriteString(s)
	enc.buf.WriteByte('"')
}

func (enc *jsonEncoder) safeAddString(s string) {
	jsonEncodeString(enc.buf, s)
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
