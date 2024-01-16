package main

import (
	"log"

	"github.com/dedpnd/unifier/internal/adapter/api/router"
	"github.com/dedpnd/unifier/internal/adapter/store"
	"github.com/dedpnd/unifier/internal/config"
	h "github.com/dedpnd/unifier/internal/core/server/http"
	"github.com/dedpnd/unifier/internal/core/worker"
)

func main() {
	log.Println("Server start...")

	// Читаем конфигурацию
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal(err.Error())
	}

	// Создаем хранилище
	str, err := store.NewStore(cfg.DatabaseDSN)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Запускаем пул воркеров
	p, err := worker.StartPool(cfg.KafkaAdress, str)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Создаем роутер
	r, err := router.Router(str, p)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Функция для завершения работы
	callback := func() {
		err := str.Close()
		if err != nil {
			log.Fatal(err.Error())
		}

		p.StopPool()
	}

	// Поднимаем сервер
	addr := "localhost:8080"
	h.GracefulServer(addr, r, callback)
}
