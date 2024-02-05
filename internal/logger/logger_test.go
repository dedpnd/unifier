package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestInit(t *testing.T) {
	// Передаем уровень логирования в функцию Init
	level := "info"
	l, err := Init(level)

	// Проверяем, что функция не вернула ошибку
	assert.NoError(t, err)
	// Проверяем, что логгер был успешно создан
	assert.NotNil(t, l)

	assert.Equal(t, l.Core().Enabled(zap.InfoLevel), true)

	level = "warn"
	l, err = Init(level)
	assert.NoError(t, err)
	assert.NotNil(t, l)
	assert.Equal(t, l.Core().Enabled(zap.InfoLevel), false)
	assert.Equal(t, l.Core().Enabled(zap.WarnLevel), true)
}
