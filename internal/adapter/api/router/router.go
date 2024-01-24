package router

import (
	"github.com/dedpnd/unifier/internal/adapter/api/middleware"
	"github.com/dedpnd/unifier/internal/adapter/api/rest"
	"github.com/dedpnd/unifier/internal/adapter/store"
	"github.com/dedpnd/unifier/internal/core/worker"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func Router(lg *zap.Logger, str store.Storage, pool worker.Pool) (chi.Router, error) {
	r := chi.NewRouter()

	r.Use(middleware.Logger(lg))

	rulesHandler := rest.RulesHandler{
		Logger: lg,
		Store:  str,
		Pool:   pool,
	}

	r.With(middleware.JWTguard).Get("/api/rules", rulesHandler.GetAllRules)
	r.With(middleware.JWTguard).Post("/api/rules", rulesHandler.CreateRule)
	r.With(middleware.JWTguard).Delete("/api/rules/{id}", rulesHandler.DeleteRule)

	userHandler := rest.UserHandler{
		Logger: lg,
		Store:  str,
	}

	r.Post("/api/user/register", userHandler.Register)
	r.Post("/api/user/login", userHandler.Login)

	return r, nil
}
