package middleware

import (
	"log"
	"net/http"
	"time"
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

func Logger() func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			start := time.Now()

			o := &responseObserver{ResponseWriter: res}
			h.ServeHTTP(o, req)

			duration := time.Since(start)

			log.Printf(`Fetch URL: [ method:%v, uri:%v, status: %v, size: %v, duration:%v ]`,
				req.Method, req.RequestURI, o.status, o.written, duration)
		})
	}
}
