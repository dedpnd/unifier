package http

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func GracefulServer(addr string, r chi.Router, lg *zap.Logger, cb func()) {
	var wg sync.WaitGroup
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	wg.Add(1)

	go func() {
		defer func() {
			cb()
			stop()
			wg.Done()
		}()

		<-ctx.Done()
	}()

	go func() {
		if err := http.ListenAndServe(addr, r); err != nil {
			lg.Fatal(err.Error())
		}
	}()

	wg.Wait()
}
