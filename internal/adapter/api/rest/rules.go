package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/dedpnd/unifier/internal/adapter/api/util"
	"github.com/dedpnd/unifier/internal/adapter/store"
	"github.com/dedpnd/unifier/internal/core/worker"
	"github.com/dedpnd/unifier/internal/models"
	"github.com/go-chi/chi/v5"
)

type RulesHandler struct {
	Store store.Storage
	Pool  worker.Pool
}

const IntServerError = "Internal server error"

func (h RulesHandler) GetAllRules(res http.ResponseWriter, req *http.Request) {
	data, err := h.Store.GetAllRules(req.Context())
	if err != nil {
		log.Println(fmt.Errorf("failed get all records from database: %w", err))
		http.Error(res, IntServerError, http.StatusInternalServerError)
		return
	}

	resBodyBytes := new(bytes.Buffer)
	if err := json.NewEncoder(resBodyBytes).Encode(&data); err != nil {
		log.Println(fmt.Errorf("invalid stringify JSON: %w", err))
		http.Error(res, IntServerError, http.StatusBadRequest)
		return
	}

	res.Header().Set("Content-Type", "application/json")

	_, err = res.Write(resBodyBytes.Bytes())
	if err != nil {
		log.Println(fmt.Errorf("failed write record to response: %w", err))
		http.Error(res, IntServerError, http.StatusInternalServerError)
		return
	}
}

func (h RulesHandler) CreateRule(res http.ResponseWriter, req *http.Request) {
	token, ok := util.GetTokenFromContext(req.Context())
	if !ok {
		log.Println(fmt.Errorf("invalid jwt token"))
		http.Error(res, IntServerError, http.StatusInternalServerError)
		return
	}

	pBody := models.Config{}

	if err := json.NewDecoder(req.Body).Decode(&pBody); err != nil {
		log.Println(fmt.Errorf("invalid parsing JSON: %w", err))
		http.Error(res, `Invalid parsing JSON`, http.StatusBadRequest)
		return
	}

	id, err := h.Store.CreateRule(req.Context(), pBody, token.ID)
	if err != nil {
		log.Println(fmt.Errorf("failed save records: %w", err))
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
		log.Println(fmt.Errorf("invalid jwt token"))
		http.Error(res, IntServerError, http.StatusInternalServerError)
		return
	}

	id := chi.URLParam(req, "id")

	pID, err := strconv.Atoi(id)
	if err != nil {
		log.Println(fmt.Errorf("failed convert id: %w", err))
		http.Error(res, `Failde convert id to int`, http.StatusBadRequest)
		return
	}

	dr, err := h.Store.GetRuleByID(req.Context(), pID)
	if err != nil {
		log.Println(fmt.Errorf("failed get rule row: %w", err))
		http.Error(res, IntServerError, http.StatusInternalServerError)
		return
	}

	if dr.ID == 0 {
		http.Error(res, "Not Found", http.StatusNotFound)
		return
	}

	if dr.Owner != token.ID {
		http.Error(res, "Forbidden", http.StatusForbidden)
		return
	}

	err = h.Store.DeleteRule(req.Context(), pID)
	if err != nil {
		log.Println(fmt.Errorf("failed delele rules row: %w", err))
		http.Error(res, IntServerError, http.StatusInternalServerError)
		return
	}

	h.Pool.DeleteWorker(id)

	res.WriteHeader(http.StatusOK)
}
