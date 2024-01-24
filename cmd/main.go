package main

import (
	"log"

	"github.com/dedpnd/unifier/internal/adapter/api/router"
	"github.com/dedpnd/unifier/internal/adapter/store"
	"github.com/dedpnd/unifier/internal/config"
	h "github.com/dedpnd/unifier/internal/core/server/http"
	"github.com/dedpnd/unifier/internal/core/worker"
	"github.com/dedpnd/unifier/internal/logger"
)

func main() {
	// Создаем логер
	lg, err := logger.Init("info")
	if err != nil {
		log.Fatalln(err.Error())
	}
	lg.Info("Server start...")

	// Читаем конфигурацию
	cfg, err := config.GetConfig()
	if err != nil {
		lg.Fatal(err.Error())
	}

	// Создаем хранилище
	str, err := store.NewStore(cfg.DatabaseDSN, lg)
	if err != nil {
		lg.Fatal(err.Error())
	}

	// Запускаем пул воркеров
	p, err := worker.StartPool(cfg.KafkaAdress, str, lg)
	if err != nil {
		lg.Fatal(err.Error())
	}

	// Создаем роутер
	r, err := router.Router(lg, str, p)
	if err != nil {
		lg.Fatal(err.Error())
	}

	// Функция для завершения работы
	callback := func() {
		err := str.Close()
		if err != nil {
			lg.Fatal(err.Error())
		}

		p.StopPool()
	}

	// Поднимаем сервер
	addr := "localhost:8080"
	h.GracefulServer(addr, r, lg, callback)
}
