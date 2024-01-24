package worker

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dedpnd/unifier/internal/adapter/store"
	"github.com/dedpnd/unifier/internal/models"
	"go.uber.org/zap"
)

type Pool struct {
	logger   *zap.Logger
	kafkaURL string
	p        map[string]workerEntity
}

type workerEntity struct {
	ID     string
	Config models.Config
	Stop   chan bool
}

func StartPool(kAddr string, str store.Storage, lg *zap.Logger) (Pool, error) {
	p := Pool{
		logger:   lg,
		kafkaURL: kAddr,
		p:        make(map[string]workerEntity),
	}

	rules, err := str.GetAllRules(context.Background())
	if err != nil {
		return Pool{}, fmt.Errorf("failed get all rule from storage: %w", err)
	}

	for i := range rules {
		id := strconv.Itoa(rules[i].ID)
		p.AddWorker(id, rules[i].Rule)
	}

	return p, nil
}

func (p Pool) AddWorker(id string, rule models.Config) {
	p.p[id] = workerEntity{
		ID:     id,
		Config: rule,
		Stop:   make(chan bool),
	}

	go func() {
		if err := Start(context.Background(), p.kafkaURL, p.p[id], p.logger); err != nil {
			p.logger.With(zap.Error(err)).Error("Worker has error", zap.String("ID", id))
		}
	}()
}

func (p Pool) DeleteWorker(id string) {
	wrk := p.p[id]

	// Останнавливаем воркер
	wrk.Stop <- true

	// Удаляем конфигурацию
	delete(p.p, id)
}

func (p Pool) StopPool() {
	for i := range p.p {
		p.p[i].Stop <- true
	}
}
