package rest

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dedpnd/unifier/internal/adapter/store"
	"github.com/dedpnd/unifier/internal/adapter/store/postgres"
	"github.com/dedpnd/unifier/internal/core/auth"
	"github.com/dedpnd/unifier/internal/models"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	Logger *zap.Logger
	Store  store.Storage
}

type LoginBody struct {
	Login    *string `json:"login"`
	Password *string `json:"password"`
}

func (h UserHandler) Register(res http.ResponseWriter, req *http.Request) {
	pBody := LoginBody{}

	if err := json.NewDecoder(req.Body).Decode(&pBody); err != nil {
		http.Error(res, `invalid parsing JSON`, http.StatusBadRequest)
		return
	}

	if pBody.Login == nil || pBody.Password == nil {
		http.Error(res, `login or password incorrect`, http.StatusBadRequest)
		return
	}

	_, err := h.Store.GetUserByLogin(req.Context(), *pBody.Login)
	if err != nil {
		h.Logger.With(zap.Error(err)).Error("failed get user from database")
		http.Error(res, IntServerError, http.StatusInternalServerError)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(*pBody.Password), bcrypt.DefaultCost)
	if err != nil {
		h.Logger.With(zap.Error(err)).Error("failed get hash from password")
		http.Error(res, IntServerError, http.StatusInternalServerError)
		return
	}

	id, err := h.Store.CreateUser(req.Context(), models.User{Login: *pBody.Login, Hash: string(hash)})
	if err != nil {
		if errors.Is(err, postgres.ErrUserUniq) {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		h.Logger.With(zap.Error(err)).Error("failed create user")
		http.Error(res, IntServerError, http.StatusInternalServerError)
		return
	}

	token, err := auth.GetJWT(id, *pBody.Login)
	if err != nil {
		h.Logger.With(zap.Error(err)).Error("failed create jwt token")
		http.Error(res, IntServerError, http.StatusInternalServerError)
		return
	}

	http.SetCookie(res, &http.Cookie{
		Name:  "token",
		Value: *token,
		Path:  "/api/",
	})

	res.WriteHeader(http.StatusOK)
}

func (h UserHandler) Login(res http.ResponseWriter, req *http.Request) {
	pBody := LoginBody{}

	if err := json.NewDecoder(req.Body).Decode(&pBody); err != nil {
		http.Error(res, `invalid parsing JSON`, http.StatusBadRequest)
		return
	}

	if pBody.Login == nil || pBody.Password == nil {
		http.Error(res, `login or password incorrect`, http.StatusBadRequest)
		return
	}

	data, err := h.Store.GetUserByLogin(req.Context(), *pBody.Login)
	if err != nil {
		h.Logger.With(zap.Error(err)).Error("failed get user from database")
		http.Error(res, IntServerError, http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(data.Hash), []byte(*pBody.Password)); err != nil {
		http.Error(res, `login or password incorrect`, http.StatusBadRequest)
		return
	}

	token, err := auth.GetJWT(data.ID, *pBody.Login)
	if err != nil {
		h.Logger.With(zap.Error(err)).Error("failed create jwt token")
		http.Error(res, IntServerError, http.StatusInternalServerError)
		return
	}

	http.SetCookie(res, &http.Cookie{
		Name:  "token",
		Value: *token,
		Path:  "/api/",
	})

	res.WriteHeader(http.StatusOK)
}
