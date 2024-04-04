package benchmarks

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"go.uber.org/zap"
)

var (
	testErr     = fmt.Errorf("ERROR")
	testString  = "Hello World"
	testStrings = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	testInt     = 99
	testInts    = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}
	testTime    = time.Now()
	testTimes   = []time.Time{
		testTime, testTime, testTime, testTime, testTime,
		testTime, testTime, testTime, testTime, testTime,
	}
	testStruct = testStructure{
		CreatedAt: time.Now(),
		Name:      "testName",
		ID:        "e5fab56f-36e6-4d1f-a162-f33cfead0e8a",
		Sum:       666,
	}
	testStructs = []testStructure{
		testStruct, testStruct, testStruct, testStruct, testStruct,
		testStruct, testStruct, testStruct, testStruct, testStruct,
	}
	testMessage = "Package slog provides structured logging, in which log records include a message, a severity level, and various other attributes expressed as key-value pairs."
)

type testStructure struct {
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	ID        string    `json:"id"`
	Sum       int       `json:"sum"`
}

func BenchmarkDisabledWithoutFields(b *testing.B) {
	b.Run("slog", func(b *testing.B) {
		logger := newDisabledSlog()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(testMessage)
			}
		})
	})
	b.Run("slog with zlog", func(b *testing.B) {
		logger := newDisabledSlogWithZlog()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(testMessage)
			}
		})
	})
	b.Run("zlog", func(b *testing.B) {
		logger := newDisabledZlog()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(testMessage)
			}
		})
	})
	b.Run("slog with zap", func(b *testing.B) {
		logger := newDisabledSlogWithZap()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(testMessage)
			}
		})
	})
	b.Run("zap", func(b *testing.B) {
		logger := newZapLogger(zap.ErrorLevel)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(testMessage)
			}
		})
	})
	b.Run("zerolog", func(b *testing.B) {
		logger := newDisabledZerolog()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info().Msg(testMessage)
			}
		})
	})
}

func BenchmarkDisabledAddingFields(b *testing.B) {
	b.Run("slog", func(b *testing.B) {
		logger := newDisabledSlog()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.LogAttrs(context.Background(), slog.LevelInfo, testMessage, slogFields()...)
			}
		})
	})
	b.Run("slog with zlog", func(b *testing.B) {
		logger := newDisabledSlogWithZlog()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.LogAttrs(context.Background(), slog.LevelInfo, testMessage, slogFields()...)
			}
		})
	})
	b.Run("zlog", func(b *testing.B) {
		logger := newDisabledZlog()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.LogAttrs(context.Background(), slog.LevelInfo, testMessage, slogFields()...)
			}
		})
	})
	b.Run("slog with zap", func(b *testing.B) {
		logger := newDisabledSlogWithZap()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.LogAttrs(context.Background(), slog.LevelInfo, testMessage, slogFields()...)
			}
		})
	})
	b.Run("zap", func(b *testing.B) {
		logger := newZapLogger(zap.ErrorLevel)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(testMessage, zapFields()...)
			}
		})
	})
	b.Run("zerolog", func(b *testing.B) {
		logger := newDisabledZerolog()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				zerologFields(logger.Info()).Msg(testMessage)
			}
		})
	})
}

func BenchmarkWithoutFields(b *testing.B) {
	b.Run("slog", func(b *testing.B) {
		logger := newSlog()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(testMessage)
			}
		})
	})
	b.Run("slog with zlog", func(b *testing.B) {
		logger := newSlogWithZlog()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(testMessage)
			}
		})
	})
	b.Run("zlog", func(b *testing.B) {
		logger := newZlog()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(testMessage)
			}
		})
	})
	b.Run("slog with zap", func(b *testing.B) {
		logger := newSlogWithZap()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(testMessage)
			}
		})
	})
	b.Run("zap", func(b *testing.B) {
		logger := newZapLogger(zap.DebugLevel)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(testMessage)
			}
		})
	})
	b.Run("zerolog", func(b *testing.B) {
		logger := newZerolog()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info().Msg(testMessage)
			}
		})
	})
}

func BenchmarkAccumulatedContext(b *testing.B) {
	b.Run("slog", func(b *testing.B) {
		logger := newSlog(slogFields()...)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(testMessage)
			}
		})
	})
	b.Run("slog with zlog", func(b *testing.B) {
		logger := newSlogWithZlog(slogFields()...)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(testMessage)
			}
		})
	})
	b.Run("zlog", func(b *testing.B) {
		logger := newZlog(slogFields()...)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(testMessage)
			}
		})
	})
	b.Run("slog with zap", func(b *testing.B) {
		logger := newSlogWithZap(slogFields()...)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(testMessage)
			}
		})
	})
	b.Run("zap", func(b *testing.B) {
		logger := newZapLogger(zap.DebugLevel).With(zapFields()...)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(testMessage)
			}
		})
	})
	b.Run("zerolog", func(b *testing.B) {
		logger := zerologContext(newZerolog().With()).Logger()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info().Msg(testMessage)
			}
		})
	})
}

func BenchmarkAddingFields(b *testing.B) {
	b.Run("slog", func(b *testing.B) {
		logger := newSlog()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.LogAttrs(context.Background(), slog.LevelInfo, testMessage, slogFields()...)
			}
		})
	})
	b.Run("slog with zlog", func(b *testing.B) {
		logger := newSlogWithZlog()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.LogAttrs(context.Background(), slog.LevelInfo, testMessage, slogFields()...)
			}
		})
	})
	b.Run("zlog", func(b *testing.B) {
		logger := newZlog()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.LogAttrs(context.Background(), slog.LevelInfo, testMessage, slogFields()...)
			}
		})
	})
	b.Run("slog with zap", func(b *testing.B) {
		logger := newSlogWithZap()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.LogAttrs(context.Background(), slog.LevelInfo, testMessage, slogFields()...)
			}
		})
	})
	b.Run("zap", func(b *testing.B) {
		logger := newZapLogger(zap.DebugLevel)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(testMessage, zapFields()...)
			}
		})
	})
	b.Run("zerolog", func(b *testing.B) {
		logger := newZerolog()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				zerologFields(logger.Info()).Msg(testMessage)
			}
		})
	})
}

func BenchmarkKVArgs(b *testing.B) {
	b.Run("slog", func(b *testing.B) {
		logger := newSlog()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(testMessage, kvArgs()...)
			}
		})
	})
	b.Run("slog with zlog", func(b *testing.B) {
		logger := newSlogWithZlog()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(testMessage, kvArgs()...)
			}
		})
	})
	b.Run("zlog", func(b *testing.B) {
		logger := newZlog()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(testMessage, kvArgs()...)
			}
		})
	})
	b.Run("slog with zap", func(b *testing.B) {
		logger := newSlogWithZap()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(testMessage, kvArgs()...)
			}
		})
	})
	b.Run("zap", func(b *testing.B) {
		logger := newZapLogger(zap.DebugLevel).Sugar()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Infow(testMessage, kvArgs()...)
			}
		})
	})
	b.Run("zerolog", func(b *testing.B) {
		logger := newZerolog()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				zerologAnyFields(logger.Info()).Msg(testMessage)
			}
		})
	})
}
