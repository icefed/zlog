package benchmarks

import (
	"io"

	"github.com/rs/zerolog"
)

func newZerolog() zerolog.Logger {
	return zerolog.New(io.Discard).With().Timestamp().Logger()
}

func newDisabledZerolog() zerolog.Logger {
	return newZerolog().Level(zerolog.Disabled)
}

func zerologAnyFields(e *zerolog.Event) *zerolog.Event {
	return e.
		Any("string", testString).
		Any("longstring", testMessage).
		Any("strings", testStrings).
		Any("int", testInt).
		Any("ints", testInts).
		Any("time", testTime).
		Any("times", testTimes).
		Any("struct", testStruct).
		Any("structs", testStructs).
		Any("error", testErr)
}

func zerologFields(e *zerolog.Event) *zerolog.Event {
	return e.
		Str("string", testString).
		Str("longstring", testMessage).
		Strs("strings", testStrings).
		Int("int", testInt).
		Ints("ints", testInts).
		Time("time", testTime).
		Times("times", testTimes).
		Any("struct", testStruct).
		Any("structs", testStructs).
		Err(testErr)
}

func zerologContext(c zerolog.Context) zerolog.Context {
	return c.
		Str("string", testString).
		Str("longstring", testMessage).
		Strs("strings", testStrings).
		Int("int", testInt).
		Ints("ints", testInts).
		Time("time", testTime).
		Times("times", testTimes).
		Interface("struct", testStruct).
		Interface("structs", testStructs).
		Err(testErr)
}
