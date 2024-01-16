package util

import (
	"context"

	"github.com/dedpnd/unifier/internal/core/auth"
)

type contextKey string

func (c contextKey) String() string {
	return string(c)
}

var (
	ContextKeyToken = contextKey("deleteCaller")
)

func GetTokenFromContext(ctx context.Context) (auth.Claims, bool) {
	caller, ok := ctx.Value(ContextKeyToken).(auth.Claims)
	return caller, ok
}
