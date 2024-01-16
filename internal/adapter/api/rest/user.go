package rest

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/dedpnd/unifier/internal/adapter/store"
	"github.com/dedpnd/unifier/internal/core/auth"
	"github.com/dedpnd/unifier/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	Store store.Storage
}

type LoginBody struct {
	Login    *string `json:"login"`
	Password *string `json:"password"`
}

func (h UserHandler) Register(res http.ResponseWriter, req *http.Request) {
	pBody := LoginBody{}

	if err := json.NewDecoder(req.Body).Decode(&pBody); err != nil {
		log.Println(fmt.Errorf("invalid parsing JSON: %w", err))
		http.Error(res, `Invalid parsing JSON`, http.StatusBadRequest)
		return
	}

	if pBody.Login == nil || pBody.Password == nil {
		http.Error(res, `Login or password incorrect`, http.StatusBadRequest)
		return
	}

	data, err := h.Store.GetUserByLogin(req.Context(), *pBody.Login)
	if err != nil {
		log.Println(fmt.Errorf("failed get user from database: %w", err))
		http.Error(res, IntServerError, http.StatusInternalServerError)
		return
	}

	if data.ID > 0 {
		http.Error(res, `Login exist`, http.StatusConflict)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(*pBody.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println(fmt.Errorf("failed get hash from password: %w", err))
		http.Error(res, IntServerError, http.StatusInternalServerError)
		return
	}

	id, err := h.Store.CreateUser(req.Context(), models.User{Login: *pBody.Login, Hash: string(hash)})
	if err != nil {
		log.Println(fmt.Errorf("failed create user: %w", err))
		http.Error(res, IntServerError, http.StatusInternalServerError)
		return
	}

	token, err := auth.GetJWT(id, *pBody.Login)
	if err != nil {
		log.Println(fmt.Errorf("failed create jwt token: %w", err))
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
		log.Println(fmt.Errorf("invalid parsing JSON: %w", err))
		http.Error(res, `Invalid parsing JSON`, http.StatusBadRequest)
		return
	}

	if pBody.Login == nil || pBody.Password == nil {
		http.Error(res, `Login or password incorrect`, http.StatusBadRequest)
		return
	}

	data, err := h.Store.GetUserByLogin(req.Context(), *pBody.Login)
	if err != nil {
		log.Println(fmt.Errorf("failed get user from database: %w", err))
		http.Error(res, IntServerError, http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(data.Hash), []byte(*pBody.Password)); err != nil {
		http.Error(res, `Login or password incorrect`, http.StatusBadRequest)
		return
	}

	token, err := auth.GetJWT(data.ID, *pBody.Login)
	if err != nil {
		log.Println(fmt.Errorf("failed create jwt token: %w", err))
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
