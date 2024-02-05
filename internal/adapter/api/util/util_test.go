package util

import (
	"context"
	"testing"

	"github.com/dedpnd/unifier/internal/core/auth"
	"github.com/stretchr/testify/assert"
)

func TestSetTokenToContext(t *testing.T) {
	// Создаем контекст
	ctx := context.Background()

	// Создаем некоторый тестовый токен
	testToken := auth.Claims{
		// ваш токен
	}

	// Вызываем функцию SetTokenToContext
	ctxWithToken := SetTokenToContext(ctx, testToken)

	// Проверяем, что токен успешно установлен в контексте
	tokenFromContext, ok := GetTokenFromContext(ctxWithToken)
	assert.True(t, ok)
	assert.Equal(t, testToken, tokenFromContext)
}

func TestGetTokenFromContext(t *testing.T) {
	// Создаем контекст
	ctx := context.Background()

	// Создаем некоторый тестовый токен
	testToken := auth.Claims{
		// ваш токен
	}

	// Устанавливаем токен в контекст
	ctxWithToken := context.WithValue(ctx, ContextKeyToken, testToken)

	// Вызываем функцию GetTokenFromContext
	tokenFromContext, ok := GetTokenFromContext(ctxWithToken)

	// Проверяем, что токен успешно получен из контекста
	assert.True(t, ok)
	assert.Equal(t, testToken, tokenFromContext)
}
