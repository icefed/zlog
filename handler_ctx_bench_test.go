package zlog

import (
	"context"
	"io"
	"testing"
)

func BenchmarkContextAttrs(b *testing.B) {
	h := NewJSONHandler(&Config{
		Writer: io.Discard,
	})
	h = h.WithOptions(WithContextExtractor(userContextExtractor))
	log := New(h)
	ctx := context.WithValue(context.Background(), userKey{}, user{
		Name: "test@test.com",
		Id:   "5c81f444-93f9-4cf8-a3b5-c3aeb377a99d",
	})
	for i := 0; i < b.N; i++ {
		log.InfoContext(ctx, "test")
	}
}
