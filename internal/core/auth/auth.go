package auth

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	ID    int    `json:"id"`
	Login string `json:"login"`
	jwt.RegisteredClaims
}

var jwtKey = []byte("12345")

func GetJWT(id int, login string) (*string, error) {
	var DefaultSession = 15
	var DefaultExpTime = time.Now().Add(time.Duration(DefaultSession) * time.Minute)

	claims := &Claims{
		ID:    id,
		Login: login,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(DefaultExpTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return nil, fmt.Errorf("failed signed jwt: %w", err)
	}

	return &tokenString, nil
}

func VerifyJWTandGetPayload(token string) (Claims, error) {
	claims := &Claims{}

	tkn, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			log.Println(fmt.Errorf("invalid jwt token: %w", err))
			return *claims, fmt.Errorf("failed signature from jwt: %w", err)
		}
		log.Println(fmt.Errorf("invalid jwt token: %w", err))
		return *claims, fmt.Errorf("invalid jwt token: %w", err)
	}

	if !tkn.Valid {
		return *claims, fmt.Errorf("jwt token not valid: %w", err)
	}

	return *claims, nil
}
