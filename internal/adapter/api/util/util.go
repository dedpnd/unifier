package util

import (
	"context"

	"github.com/dedpnd/unifier/internal/core/auth"
)

type contextKey int

const (
	ContextKeyToken contextKey = iota
)

func GetTokenFromContext(ctx context.Context) (auth.Claims, bool) {
	caller, ok := ctx.Value(ContextKeyToken).(auth.Claims)
	return caller, ok
}

func SetTokenToContext(ctx context.Context, pl auth.Claims) context.Context {
	return context.WithValue(ctx, ContextKeyToken, pl)
}
