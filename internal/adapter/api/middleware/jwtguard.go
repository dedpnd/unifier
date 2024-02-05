package middleware

import (
	"errors"
	"net/http"

	"github.com/dedpnd/unifier/internal/adapter/api/util"
	"github.com/dedpnd/unifier/internal/core/auth"
)

func JWTguard(h http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		c, err := req.Cookie("token")
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				res.WriteHeader(http.StatusUnauthorized)
				return
			}
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		token := c.Value
		pl, err := auth.VerifyJWTandGetPayload(token)
		if err != nil {
			res.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := req.Context()
		r := req.WithContext(util.SetTokenToContext(ctx, pl))

		h.ServeHTTP(res, r)
	})
}
