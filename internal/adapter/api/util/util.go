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
