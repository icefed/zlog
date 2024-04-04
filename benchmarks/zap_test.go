package benchmarks

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
)

func newZapLogger(lvl zapcore.Level) *zap.Logger {
	ec := zap.NewProductionEncoderConfig()
	enc := zapcore.NewJSONEncoder(ec)
	return zap.New(zapcore.NewCore(
		enc,
		&zaptest.Discarder{},
		lvl,
	))
}

func zapFields() []zap.Field {
	return []zap.Field{
		zap.String("string", testString),
		zap.String("longstring", testMessage),
		zap.Strings("strings", testStrings),
		zap.Int("int", testInt),
		zap.Ints("ints", testInts),
		zap.Time("time", testTime),
		zap.Times("times", testTimes),
		zap.Any("struct", testStruct),
		zap.Any("structs", testStructs),
		zap.Error(testErr),
	}
}
