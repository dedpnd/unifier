package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/dedpnd/unifier/internal/adapter/store/postgres"
	"github.com/dedpnd/unifier/internal/models"
	"go.uber.org/zap"
)

type Storage interface {
	Close() error
	GetUserByLogin(ctx context.Context, login string) (models.User, error)
	CreateUser(ctx context.Context, user models.User) (int, error)
	GetRuleByID(ctx context.Context, id int) (models.Rule, error)
	GetAllRules(ctx context.Context) ([]models.Rule, error)
	CreateRule(ctx context.Context, rule models.Config, owner int) (int, error)
	DeleteRule(ctx context.Context, id int) error
}

func NewStore(dsn string, lg *zap.Logger) (Storage, error) {
	if len(dsn) != 0 {
		dbs, err := postgres.NewDB(context.Background(), dsn, lg)
		if err != nil {
			return nil, fmt.Errorf("failed create database storage: %w", err)
		}
		return dbs, nil
	}

	return nil, errors.New("storage not found")
}
