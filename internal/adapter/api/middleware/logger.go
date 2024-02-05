package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type responseObserver struct {
	http.ResponseWriter
	status  int
	written int64
}

func (o *responseObserver) Write(p []byte) (n int, err error) {
	n, err = o.ResponseWriter.Write(p)
	o.written += int64(n)
	return
}

func (o *responseObserver) WriteHeader(code int) {
	o.ResponseWriter.WriteHeader(code)
	o.status = code
}

func Logger(lg *zap.Logger) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			start := time.Now()

			o := &responseObserver{ResponseWriter: res}
			h.ServeHTTP(o, req)

			duration := time.Since(start)

			lg.Info("Fetch URL:",
				zap.String("method", req.Method),
				zap.String("uri", req.RequestURI),
				zap.Int("status", o.status),
				zap.Int64("size", o.written),
				zap.Duration("duration", duration),
			)
		})
	}
}
