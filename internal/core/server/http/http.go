package http

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/go-chi/chi/v5"
)

func GracefulServer(addr string, r chi.Router, cb func()) {
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
			log.Fatal(err.Error())
		}
	}()

	wg.Wait()
}
