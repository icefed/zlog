package zlog

import (
	"fmt"
	"io/fs"
	"log/slog"
	"math/big"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/icefed/zlog/buffer"
)

type fakeJSONMarshaler struct {
	json []byte
	err  error
}

func (f fakeJSONMarshaler) MarshalJSON() ([]byte, error) {
	return f.json, f.err
}

func TestJSONEncoder(t *testing.T) {
	h := NewJSONHandler(&Config{
		HandlerOptions: slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelDebug,
		},
		TimeDurationAsInt: true,
		IgnoreEmptyGroup:  true,
	})
	buf := buffer.New()
	defer buf.Free()

	tests := []struct {
		name        string
		groups      []string
		key         string
		value       any
		want        string
		replaceAttr func(_ []string, a slog.Attr) slog.Attr
	}{
		{
			name:  "time",
			key:   slog.TimeKey,
			value: testTime,
			want:  `"time":"2023-08-16T01:02:03.666Z"`,
		}, {
			name:   "level with group",
			groups: []string{"g"},
			key:    slog.LevelKey,
			value:  slog.LevelInfo,
			want:   `"g":{"level":"INFO"`,
		}, {
			name:  "msg",
			key:   slog.MessageKey,
			value: "test msg",
			want:  `"msg":"test msg"`,
		}, {
			name:   "error with group",
			groups: []string{"g", "g2"},
			key:    "error",
			value:  fs.ErrNotExist,
			want:   `"g":{"g2":{"error":"file does not exist"`,
		}, {
			name: "source",
			key:  slog.SourceKey,
			value: &slog.Source{
				File: "test.go",
				Line: 300,
			},
			want: `"source":"test.go:300"`,
		}, {
			name:  "stacktrace",
			key:   "stacktrace",
			value: &stacktrace{getPC()},
			want:  `"stacktrace":"` + wantPCFunction + "\\n\\t" + wantPCFile + ":" + strconv.Itoa(wantPCLine) + `"`,
		}, {
			name:  "ip",
			key:   "ip",
			value: net.ParseIP("127.0.0.1"),
			want:  `"ip":"127.0.0.1"`,
		}, {
			name:  "formatted",
			key:   "formatted",
			value: []byte(`"env":"prod","app":"web"`),
			want:  `"env":"prod","app":"web"`,
		}, {
			name:  "group",
			key:   "group",
			value: slog.GroupValue(slog.Attr{Key: "env", Value: slog.StringValue("prod")}, slog.Attr{Key: "app", Value: slog.StringValue("web")}),
			want:  `"group":{"env":"prod","app":"web"}`,
		},
		{name: "bool", key: "bool", value: true, want: `"bool":true`},
		{name: "duration", key: "duration", value: time.Second * 10, want: `"duration":10000000000`},
		{name: "float64", key: "float64", value: float64(0.32), want: `"float64":0.32`},
		{name: "int64", key: "int64", value: int64(32), want: `"int64":32`},
		{name: "uint64", key: "uint64", value: uint64(111), want: `"uint64":111`},
		{name: "string", key: "string", value: "stringvalue", want: `"string":"stringvalue"`},
		{
			name:  "escapedstring",
			key:   "escapedstring",
			value: "😀z\nz\rz\tz\"z\\z\u2028z\x00z\xff",
			want:  `"escapedstring":"😀z\nz\rz\tz\"z\\z\u2028z\u0000z\ufffd"`,
		}, {
			name:  "MarshalJSON error",
			key:   "marshaljsonerror",
			value: &fakeJSONMarshaler{err: fmt.Errorf("MarshalJSON error")},
			want:  `"marshaljsonerror":"!ERROR:MarshalJSON error"`,
		}, {
			name:  "MarshalJSON get invalid",
			key:   "marshaljsongetinvalid",
			value: &fakeJSONMarshaler{json: []byte(`"invali"d"`)},
			want:  `"marshaljsongetinvalid":"!ERROR:invalid MarshalJSON output:\"invali\"d\""`,
		}, {
			name:  "MarshalText error",
			key:   "marshaltexterror",
			value: &fakeTextMarshaler{err: fmt.Errorf("MarshalText error")},
			want:  `"marshaltexterror":"!ERROR:MarshalText error"`,
		}, {
			name:  "json encode error",
			key:   "jsonencodeerror",
			value: make(chan int),
			want:  `"jsonencodeerror":"!ERROR:json: unsupported type: chan int"`,
		}, {
			name:  "ignore level",
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
			name:  "error+1",
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
			name:  "replace time",
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
			name:  "replace time to string",
			key:   "timestring2",
			value: testTime,
			want:  `"timestring2":"2023-08-16T01:02:03.666666666Z"`,
			replaceAttr: func(_ []string, a slog.Attr) slog.Attr {
				if a.Key == "timestring2" {
					a.Value = slog.StringValue("2023-08-16T01:02:03.666666666Z")
				}
				return a
			},
		}, {
			name:  "testtime",
			key:   "testtime",
			value: testTime,
			want:  `"testtime":"2023-08-16T01:02:03.666Z"`,
		}, {
			name:  "deletetime",
			key:   "deletetime",
			value: testTime,
			want:  ``,
			replaceAttr: func(_ []string, a slog.Attr) slog.Attr {
				if a.Key == "deletetime" {
					a.Key = ""
				}
				return a
			},
		}, {
			name: "source",
			key:  slog.SourceKey,
			value: &slog.Source{
				File: "test.go",
				Line: 300,
			},
			want: `"source":"test.go:300"`,
		}, {
			name:  "replacepc",
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
			name:  "stacktrace",
			key:   "stacktrace",
			value: &stacktrace{getPC()},
			want:  `"stacktrace":"` + wantPCFunction + "\\n\\t" + wantPCFile + ":" + strconv.Itoa(wantPCLine) + `"`,
		}, {
			name:  "group",
			key:   "group",
			value: slog.GroupValue(slog.Attr{Key: "env", Value: slog.StringValue("prod")}, slog.Attr{Key: "app", Value: slog.StringValue("web")}),
			want:  `"group":{"env":"prod","app":"web"}`,
		}, {
			name:  "ignore ip",
			key:   "ipaddr",
			value: &net.IPAddr{IP: net.IPv4(127, 0, 0, 1)},
			replaceAttr: func(_ []string, a slog.Attr) slog.Attr {
				if a.Key == "ipaddr" {
					a.Key = ""
				}
				return a
			},
		}, {
			name:   "replace in group",
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
		t.Run(test.name, func(t *testing.T) {
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
				enc.AppendMessage(test.key, v)
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
		})
	}
}

func TestJSONEncoderAddAny(t *testing.T) {
	h := NewJSONHandler(&Config{
		HandlerOptions: slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
		TimeDurationAsInt: true,
		IgnoreEmptyGroup:  true,
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
		name  string
		value any
		want  string
	}{
		{
			name:  "nil source",
			value: (*slog.Source)(nil),
			want:  `null`,
		}, {
			name:  "nil stacktrace",
			value: (*stacktrace)(nil),
			want:  `null`,
		}, {
			name:  "time ptr",
			value: vTimePtr,
			want:  `"2023-08-16T01:02:03.666666666Z"`,
		}, {
			name:  "nil time",
			value: (*time.Time)(nil),
			want:  `null`,
		}, {
			name:  "time duration",
			value: vDuration,
			want:  `10000000000`,
		}, {
			name:  "nil time duration",
			value: (*time.Duration)(nil),
			want:  `null`,
		}, {
			name:  "string ptr",
			value: vStringPtr,
			want:  `"stringvalue"`,
		}, {
			name:  "nil string",
			value: (*string)(nil),
			want:  `null`,
		}, {
			name:  "bool ptr",
			value: vBoolPtr,
			want:  `true`,
		}, {
			name:  "nil bool",
			value: (*bool)(nil),
			want:  `null`,
		}, {
			name:  "int ptr",
			value: vIntPtr,
			want:  `32`,
		}, {
			name:  "nil int",
			value: (*int)(nil),
			want:  `null`,
		}, {
			name:  "int8 ptr",
			value: vInt8Ptr,
			want:  `32`,
		}, {
			name:  "nil int8",
			value: (*int8)(nil),
			want:  `null`,
		}, {
			name:  "int16 ptr",
			value: vInt16Ptr,
			want:  `32`,
		}, {
			name:  "nil int16",
			value: (*int16)(nil),
			want:  `null`,
		}, {
			name:  "int32 ptr",
			value: vInt32Ptr,
			want:  `32`,
		}, {
			name:  "nil int32",
			value: (*int32)(nil),
			want:  `null`,
		}, {
			name:  "int64 ptr",
			value: vInt64Ptr,
			want:  `32`,
		}, {
			name:  "nil int64",
			value: (*int64)(nil),
			want:  `null`,
		}, {
			name:  "uint ptr",
			value: vUintPtr,
			want:  `32`,
		}, {
			name:  "nil uint",
			value: (*uint)(nil),
			want:  `null`,
		}, {
			name:  "uint8 ptr",
			value: vUint8Ptr,
			want:  `32`,
		}, {
			name:  "nil uint8",
			value: (*uint8)(nil),
			want:  `null`,
		}, {
			name:  "uint16 ptr",
			value: vUint16Ptr,
			want:  `32`,
		}, {
			name:  "nil uint16",
			value: (*uint16)(nil),
			want:  `null`,
		}, {
			name:  "uint32 ptr",
			value: vUint32Ptr,
			want:  `32`,
		}, {
			name:  "nil uint32",
			value: (*uint32)(nil),
			want:  `null`,
		}, {
			name:  "uint64 ptr",
			value: vUint64Ptr,
			want:  `32`,
		}, {
			name:  "nil uint64",
			value: (*uint64)(nil),
			want:  `null`,
		}, {
			name:  "float32 ptr",
			value: vFloat32Ptr,
			want:  `0.32`,
		}, {
			name:  "nil float32",
			value: (*float32)(nil),
			want:  `null`,
		}, {
			name:  "float64 ptr",
			value: vFloat64Ptr,
			want:  `0.32`,
		}, {
			name:  "nil float64",
			value: (*float64)(nil),
			want:  `null`,
		}, {
			name:  "error array",
			value: []error{fs.ErrInvalid, fs.ErrPermission, nil},
			want:  `["invalid argument","permission denied",null]`,
		}, {
			name:  "nil error array",
			value: nilErrArr,
			want:  `null`,
		}, {
			name:  "rune array",
			value: []byte(`testbytes😀`),
			want:  `"testbytes😀"`,
		}, {
			name:  "nil byte array",
			value: nilByteArr,
			want:  `""`,
		}, {
			name:  "byte array",
			value: []byte{0, 1, 2, 3, '4'},
			want:  `"AAECAzQ="`,
		}, {
			name:  "string array",
			value: []string{"a", "b", "c"},
			want:  `["a","b","c"]`,
		}, {
			name:  "nil string array",
			value: nilStrArr,
			want:  `null`,
		}, {
			name:  "bool array",
			value: []bool{true, false, true},
			want:  `[true,false,true]`,
		}, {
			name:  "nil bool array",
			value: nilBoolArr,
			want:  `null`,
		}, {
			name:  "int array",
			value: []int{1, 2, 3, -1},
			want:  `[1,2,3,-1]`,
		}, {
			name:  "nil int array",
			value: nilIntArr,
			want:  `null`,
		}, {
			name:  "int8 array",
			value: []int8{1, 2, 3},
			want:  `[1,2,3]`,
		}, {
			name:  "nil int8 array",
			value: nilInt8Arr,
			want:  `null`,
		}, {
			name:  "int16 array",
			value: []int16{1, 2, 3},
			want:  `[1,2,3]`,
		}, {
			name:  "nil int16 array",
			value: nilInt16Arr,
			want:  `null`,
		}, {
			name:  "int32 array",
			value: []int32{1, 2, 3},
			want:  `[1,2,3]`,
		}, {
			name:  "nil int32 array",
			value: nilInt32Arr,
			want:  `null`,
		}, {
			name:  "int64 array",
			value: []int64{1, 2, 3},
			want:  `[1,2,3]`,
		}, {
			name:  "nil int64 array",
			value: nilInt64Arr,
			want:  `null`,
		}, {
			name:  "uint array",
			value: []uint{1, 2, 3},
			want:  `[1,2,3]`,
		}, {
			name:  "nil uint array",
			value: nilUintArr,
			want:  `null`,
		}, {
			name:  "uint8 array",
			value: []uint8{1, 2, 3},
			want:  `"AQID"`,
		}, {
			name:  "nil uint8 array",
			value: []uint16{1, 2, 3},
			want:  `[1,2,3]`,
		}, {
			name:  "nil uint16 array",
			value: nilUint16Arr,
			want:  `null`,
		}, {
			name:  "uint32 array",
			value: []uint32{1, 2, 3},
			want:  `[1,2,3]`,
		}, {
			name:  "nil uint32 array",
			value: nilUint32Arr,
			want:  `null`,
		}, {
			name:  "uint64 array",
			value: []uint64{1, 2, 3},
			want:  `[1,2,3]`,
		}, {
			name:  "nil uint64 array",
			value: nilUint64Arr,
			want:  `null`,
		}, {
			name:  "float32 array",
			value: []float32{0.1, 0.2, 0.3},
			want:  `[0.1,0.2,0.3]`,
		}, {
			name:  "nil float32 array",
			value: nilFloat32Arr,
			want:  `null`,
		}, {
			name:  "float64 array",
			value: []float64{0.1, 0.2, 0.3},
			want:  `[0.1,0.2,0.3]`,
		}, {
			name:  "nil float64 array",
			value: nilFloat64Arr,
			want:  `null`,
		}, {
			name:  "duration array",
			value: []time.Duration{time.Second, time.Minute, time.Hour},
			want:  `[1000000000,60000000000,3600000000000]`,
		}, {
			name:  "nil duration array",
			value: nilDurationArr,
			want:  `null`,
		}, {
			name:  "time array",
			value: []time.Time{testTime, testTime.Add(time.Hour), testTime.Add(time.Hour * 2)},
			want:  `["2023-08-16T01:02:03.666666666Z","2023-08-16T02:02:03.666666666Z","2023-08-16T03:02:03.666666666Z"]`,
		}, {
			name:  "nil time array",
			value: nilTimeArr,
			want:  `null`,
		}, {
			name:  "bigint",
			value: big.NewInt(32),
			want:  `32`,
		}, {
			name:  "nil bigint",
			value: (*big.Int)(nil),
			want:  `null`,
		}, {
			name:  "net.IP",
			value: net.ParseIP("192.168.1.1"),
			want:  `"192.168.1.1"`,
		}, {
			name:  "nil net.IP",
			value: (*net.IP)(nil),
			want:  `null`,
		}, {
			name:  "net.AddrError",
			value: &net.AddrError{Err: "invalid argument", Addr: "127.0.0.1"},
			want:  `"address 127.0.0.1: invalid argument"`,
		}, {
			name:  "nil net.AddrError",
			value: (*net.AddrError)(nil),
			want:  `null`,
		}, {
			name:  "net.NS",
			value: &net.NS{Host: "localhost"},
			want:  `{"Host":"localhost"}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			enc := newJSONEncoder(h, buf)
			enc.addAny(test.value)
			if string(buf.Bytes()) != test.want {
				t.Errorf("test %T, got %v, want %v", test.value, string(buf.Bytes()), test.want)
			}
			buf.Reset()
		})
	}
}

func TestJSONEncoderTimeDurationAsInt(t *testing.T) {
	h := NewJSONHandler(&Config{
		HandlerOptions: slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
		IgnoreEmptyGroup: true,
	})
	buf := buffer.New()
	defer buf.Free()

	tests := []struct {
		name  string
		key   string
		value any
		want  string
	}{
		{
			name:  "durseconds",
			key:   "durseconds",
			value: time.Second * 30,
			want:  `"durseconds":"30s"`,
		}, {
			name:  "durminutes",
			key:   "durminutes",
			value: time.Minute * 30,
			want:  `"durminutes":"30m0s"`,
		}, {
			name:  "durm_s",
			key:   "durm_s",
			value: time.Minute*30 + time.Second*10,
			want:  `"durm_s":"30m10s"`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			enc := newJSONEncoder(h, buf)
			enc.AppendAttr(slog.Attr{
				Key:   test.key,
				Value: slog.AnyValue(test.value),
			})
			if string(buf.Bytes()) != test.want {
				t.Errorf("got %v, want %v", string(buf.Bytes()), test.want)
			}
			buf.Reset()
		})
	}
}
