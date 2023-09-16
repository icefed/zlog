package zlog

import (
	"io/fs"
	"math/big"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/icefed/zlog/buffer"
	"golang.org/x/exp/slog"
)

func TestJSONEncoderNoReplace(t *testing.T) {
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
}

func TestJSONEncoderReplace(t *testing.T) {
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
}

func TestJSONEncoderAddAny(t *testing.T) {
	h := NewJSONHandler(&Config{
		HandlerOptions: slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	})
	buf := buffer.New()
	defer buf.Free()

	vTimePtr := &testTime
	dur := time.Second * 10
	vDuration := &dur
	str := "stringvalue"
	vStringPtr := &str
	b := true
	vBoolPtr := &b
	i := 32
	vIntPtr := &i
	i8 := int8(32)
	vInt8Ptr := &i8
	i16 := int16(32)
	vInt16Ptr := &i16
	i32 := int32(32)
	vInt32Ptr := &i32
	i64 := int64(32)
	vInt64Ptr := &i64
	ui := uint(32)
	vUintPtr := &ui
	ui8 := uint8(32)
	vUint8Ptr := &ui8
	ui16 := uint16(32)
	vUint16Ptr := &ui16
	ui32 := uint32(32)
	vUint32Ptr := &ui32
	ui64 := uint64(32)
	vUint64Ptr := &ui64
	f32 := float32(0.32)
	vFloat32Ptr := &f32
	f64 := float64(0.32)
	vFloat64Ptr := &f64
	var nilErrArr []error
	var nilByteArr []byte
	var nilStrArr []string
	var nilBoolArr []bool
	var nilIntArr []int
	var nilInt8Arr []int8
	var nilInt16Arr []int16
	var nilInt32Arr []int32
	var nilInt64Arr []int64
	var nilUintArr []uint
	var nilUint16Arr []uint16
	var nilUint32Arr []uint32
	var nilUint64Arr []uint64
	var nilFloat32Arr []float32
	var nilFloat64Arr []float64
	var nilDurationArr []time.Duration
	var nilTimeArr []time.Time

	tests := []struct {
		value any
		want  string
	}{
		{
			value: (*slog.Source)(nil),
			want:  `null`,
		}, {
			value: (*stacktrace)(nil),
			want:  `null`,
		}, {
			value: vTimePtr,
			want:  `"2023-08-16T01:02:03.666666666Z"`,
		}, {
			value: (*time.Time)(nil),
			want:  `null`,
		}, {
			value: vDuration,
			want:  `10000000000`,
		}, {
			value: (*time.Duration)(nil),
			want:  `null`,
		}, {
			value: vStringPtr,
			want:  `"stringvalue"`,
		}, {
			value: (*string)(nil),
			want:  `null`,
		}, {
			value: vBoolPtr,
			want:  `true`,
		}, {
			value: (*bool)(nil),
			want:  `null`,
		}, {
			value: vIntPtr,
			want:  `32`,
		}, {
			value: (*int)(nil),
			want:  `null`,
		}, {
			value: vInt8Ptr,
			want:  `32`,
		}, {
			value: (*int8)(nil),
			want:  `null`,
		}, {
			value: vInt16Ptr,
			want:  `32`,
		}, {
			value: (*int16)(nil),
			want:  `null`,
		}, {
			value: vInt32Ptr,
			want:  `32`,
		}, {
			value: (*int32)(nil),
			want:  `null`,
		}, {
			value: vInt64Ptr,
			want:  `32`,
		}, {
			value: (*int64)(nil),
			want:  `null`,
		}, {
			value: vUintPtr,
			want:  `32`,
		}, {
			value: (*uint)(nil),
			want:  `null`,
		}, {
			value: vUint8Ptr,
			want:  `32`,
		}, {
			value: (*uint8)(nil),
			want:  `null`,
		}, {
			value: vUint16Ptr,
			want:  `32`,
		}, {
			value: (*uint16)(nil),
			want:  `null`,
		}, {
			value: vUint32Ptr,
			want:  `32`,
		}, {
			value: (*uint32)(nil),
			want:  `null`,
		}, {
			value: vUint64Ptr,
			want:  `32`,
		}, {
			value: (*uint64)(nil),
			want:  `null`,
		}, {
			value: vFloat32Ptr,
			want:  `0.32`,
		}, {
			value: (*float32)(nil),
			want:  `null`,
		}, {
			value: vFloat64Ptr,
			want:  `0.32`,
		}, {
			value: (*float64)(nil),
			want:  `null`,
		}, {
			value: []error{fs.ErrInvalid, fs.ErrPermission, nil},
			want:  `["invalid argument","permission denied",null]`,
		}, {
			value: nilErrArr,
			want:  `null`,
		}, {
			value: []byte(`testbytes😀`),
			want:  `"testbytes😀"`,
		}, {
			value: nilByteArr,
			want:  `""`,
		}, {
			value: []byte{0, 1, 2, 3, '4'},
			want:  `"AAECAzQ="`,
		}, {
			value: []string{"a", "b", "c"},
			want:  `["a","b","c"]`,
		}, {
			value: nilStrArr,
			want:  `null`,
		}, {
			value: []bool{true, false, true},
			want:  `[true,false,true]`,
		}, {
			value: nilBoolArr,
			want:  `null`,
		}, {
			value: []int{1, 2, 3, -1},
			want:  `[1,2,3,-1]`,
		}, {
			value: nilIntArr,
			want:  `null`,
		}, {
			value: []int8{1, 2, 3},
			want:  `[1,2,3]`,
		}, {
			value: nilInt8Arr,
			want:  `null`,
		}, {
			value: []int16{1, 2, 3},
			want:  `[1,2,3]`,
		}, {
			value: nilInt16Arr,
			want:  `null`,
		}, {
			value: []int32{1, 2, 3},
			want:  `[1,2,3]`,
		}, {
			value: nilInt32Arr,
			want:  `null`,
		}, {
			value: []int64{1, 2, 3},
			want:  `[1,2,3]`,
		}, {
			value: nilInt64Arr,
			want:  `null`,
		}, {
			value: []uint{1, 2, 3},
			want:  `[1,2,3]`,
		}, {
			value: nilUintArr,
			want:  `null`,
		}, {
			value: []uint8{1, 2, 3},
			want:  `"AQID"`,
		}, {
			value: []uint16{1, 2, 3},
			want:  `[1,2,3]`,
		}, {
			value: nilUint16Arr,
			want:  `null`,
		}, {
			value: []uint32{1, 2, 3},
			want:  `[1,2,3]`,
		}, {
			value: nilUint32Arr,
			want:  `null`,
		}, {
			value: []uint64{1, 2, 3},
			want:  `[1,2,3]`,
		}, {
			value: nilUint64Arr,
			want:  `null`,
		}, {
			value: []float32{0.1, 0.2, 0.3},
			want:  `[0.1,0.2,0.3]`,
		}, {
			value: nilFloat32Arr,
			want:  `null`,
		}, {
			value: []float64{0.1, 0.2, 0.3},
			want:  `[0.1,0.2,0.3]`,
		}, {
			value: nilFloat64Arr,
			want:  `null`,
		}, {
			value: []time.Duration{time.Second, time.Minute, time.Hour},
			want:  `[1000000000,60000000000,3600000000000]`,
		}, {
			value: nilDurationArr,
			want:  `null`,
		}, {
			value: []time.Time{testTime, testTime.Add(time.Hour), testTime.Add(time.Hour * 2)},
			want:  `["2023-08-16T01:02:03.666666666Z","2023-08-16T02:02:03.666666666Z","2023-08-16T03:02:03.666666666Z"]`,
		}, {
			value: nilTimeArr,
			want:  `null`,
		}, {
			value: big.NewInt(32),
			want:  `32`,
		}, {
			value: (*big.Int)(nil),
			want:  `null`,
		}, {
			value: net.ParseIP("192.168.1.1"),
			want:  `"192.168.1.1"`,
		}, {
			value: (*net.IP)(nil),
			want:  `null`,
		}, {
			value: &net.AddrError{Err: "invalid argument", Addr: "127.0.0.1"},
			want:  `"address 127.0.0.1: invalid argument"`,
		}, {
			value: (*net.AddrError)(nil),
			want:  `null`,
		}, {
			value: &net.NS{Host: "localhost"},
			want:  `{"Host":"localhost"}`,
		},
	}
	for _, test := range tests {
		enc := newJSONEncoder(h, buf)
		enc.addAny(test.value)
		if string(buf.Bytes()) != test.want {
			t.Errorf("test %T, got %v, want %v", test.value, string(buf.Bytes()), test.want)
		}
		buf.Reset()
	}
}
