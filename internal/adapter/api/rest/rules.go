package rest

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/dedpnd/unifier/internal/adapter/api/util"
	"github.com/dedpnd/unifier/internal/adapter/store"
	"github.com/dedpnd/unifier/internal/core/worker"
	"github.com/dedpnd/unifier/internal/models"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type RulesHandler struct {
	Logger *zap.Logger
	Store  store.Storage
	Pool   worker.Pool
}

const IntServerError = "internal server error"

func (h RulesHandler) GetAllRules(res http.ResponseWriter, req *http.Request) {
	data, err := h.Store.GetAllRules(req.Context())
	if err != nil {
		h.Logger.With(zap.Error(err)).Error("failed get all records from database")
		http.Error(res, IntServerError, http.StatusInternalServerError)
		return
	}

	resBodyBytes := new(bytes.Buffer)
	if err := json.NewEncoder(resBodyBytes).Encode(&data); err != nil {
		http.Error(res, IntServerError, http.StatusBadRequest)
		return
	}

	res.Header().Set("Content-Type", "application/json")

	_, err = res.Write(resBodyBytes.Bytes())
	if err != nil {
		h.Logger.With(zap.Error(err)).Error("failed write record to response")
		http.Error(res, IntServerError, http.StatusInternalServerError)
		return
	}
}

func (h RulesHandler) CreateRule(res http.ResponseWriter, req *http.Request) {
	token, ok := util.GetTokenFromContext(req.Context())
	if !ok {
		h.Logger.Error("invalid jwt token")
		http.Error(res, IntServerError, http.StatusInternalServerError)
		return
	}

	pBody := models.Config{}

	if err := json.NewDecoder(req.Body).Decode(&pBody); err != nil {
		http.Error(res, `invalid parsing JSON`, http.StatusBadRequest)
		return
	}

	id, err := h.Store.CreateRule(req.Context(), pBody, token.ID)
	if err != nil {
		h.Logger.With(zap.Error(err)).Error("failed save rule")
		http.Error(res, IntServerError, http.StatusInternalServerError)
		return
	}

	pID := strconv.Itoa(id)
	h.Pool.AddWorker(pID, pBody)

	res.WriteHeader(http.StatusOK)
}

func (h RulesHandler) DeleteRule(res http.ResponseWriter, req *http.Request) {
	token, ok := util.GetTokenFromContext(req.Context())
	if !ok {
		h.Logger.Error("invalid jwt token")
		http.Error(res, IntServerError, http.StatusInternalServerError)
		return
	}

	id := chi.URLParam(req, "id")

	pID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(res, `failde convert id to int`, http.StatusBadRequest)
		return
	}

	dr, err := h.Store.GetRuleByID(req.Context(), pID)
	if err != nil {
		h.Logger.With(zap.Error(err)).Error("failed get rule")
		http.Error(res, IntServerError, http.StatusInternalServerError)
		return
	}

	if dr.ID == 0 {
		http.Error(res, "not found", http.StatusNotFound)
		return
	}

	if *dr.Owner != token.ID {
		http.Error(res, "forbidden", http.StatusForbidden)
		return
	}

	err = h.Store.DeleteRule(req.Context(), pID)
	if err != nil {
		h.Logger.With(zap.Error(err)).Error("failed delete rule")
		http.Error(res, IntServerError, http.StatusInternalServerError)
		return
	}

	h.Pool.DeleteWorker(id)

	res.WriteHeader(http.StatusOK)
}
