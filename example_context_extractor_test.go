package zlog_test

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/icefed/zlog"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ...
		// Pretend that we read and parsed the token, and the user authentication succeeded
		ctx := context.WithValue(context.Background(), userKey{}, user{
			Name: "test@test.com",
			Id:   "a2067a0a-6b0b-4ee5-a049-16bdb8ed6ff5",
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func LogMiddleware(log *zlog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		duration := time.Since(start)
		log.InfoContext(r.Context(), "Received request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("duration", duration.String()),
		)
	})
}

func hello(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello, World!"))
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

type userKey struct{}
type user struct {
	Name string
	Id   string
}

// The following is an example of printing a user request in http server. The log
// contains user information and can be used as an audit log.

func ExampleContextExtractor_userContext() {
	h := zlog.NewJSONHandler(&zlog.Config{
		HandlerOptions: slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	})
	h = h.WithOptions(zlog.WithContextExtractor(userContextExtractor))
	log := zlog.New(h)

	httpHandler := http.HandlerFunc(hello)
	// set auth middleware
	handler := AuthMiddleware(httpHandler)
	// set log middleware
	handler = LogMiddleware(log, handler)

	log.Info("starting server, listening on port 8080")
	http.ListenAndServe(":8080", handler)

	// Send a request using curl.
	// $ curl http://localhost:8080/api/v1/products
	// Hello, World!

	// Output like:
	// {"time":"2023-09-09T19:51:55.683+08:00","level":"INFO","msg":"starting server, listening on port 8080"}
	// {"time":"2023-09-09T19:52:04.228+08:00","level":"INFO","msg":"Received request","user":{"name":"test@test.com","id":"a2067a0a-6b0b-4ee5-a049-16bdb8ed6ff5"},"method":"GET","path":"/api/v1/products","duration":"6.221µs"}
}
