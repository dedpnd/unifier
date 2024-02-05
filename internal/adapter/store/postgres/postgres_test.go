package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/dedpnd/unifier/internal/models"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var databaseURL string
var db DataBase

func TestMain(m *testing.M) {
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not construct pool: %s", err)
	}

	err = pool.Client.Ping()
	if err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "16.1-alpine3.18",
		Env: []string{
			"POSTGRES_PASSWORD=test",
			"POSTGRES_USER=test",
			"listen_addresses = '*'",
		},
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	hostAndPort := resource.GetHostPort("5432/tcp")
	databaseURL = fmt.Sprintf("postgres://test:test@%s?sslmode=disable", hostAndPort)

	log.Println("Connecting to database on url: ", databaseURL)

	// Tell docker to hard kill the container in 120 seconds
	err = resource.Expire(120)
	if err != nil {
		log.Fatalf("Expire resource has error: %s", err)
	}

	var sqlDB *sql.DB
	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	pool.MaxWait = 20 * time.Second
	if err = pool.Retry(func() error {
		sqlDB, err = sql.Open("postgres", databaseURL)
		if err != nil {
			return fmt.Errorf("Connection has error: %w", err)
		}

		err = sqlDB.Ping()
		if err != nil {
			return fmt.Errorf("Ping has error: %w", err)
		}

		return nil
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	// Создаем подключение к тестируемой базе данных
	ctx := context.Background()
	db, err = NewDB(ctx, databaseURL, zap.NewNop())
	if err != nil {
		log.Fatalf("Could not create new database: %s", err)
	}

	// Run tests
	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	err = db.Close()
	if err != nil {
		log.Fatalf("Could not close db: %s", err)
	}

	os.Exit(code)
}

func TestNewDB_SuccessfulConnection(t *testing.T) {
	ctx := context.Background()
	db, err := NewDB(ctx, databaseURL, zap.NewNop())
	assert.NoError(t, err)
	assert.NotNil(t, db)
}

func TestClose(t *testing.T) {
	ctx := context.Background()
	db, err := NewDB(ctx, databaseURL, zap.NewNop())
	if err != nil {
		log.Fatalf("Could not create new database: %s", err)
	}

	// Закрываем базу данных
	err = db.Close()

	// Проверяем, что ошибки нет
	assert.NoError(t, err)
}

func TestRunMigrations_SuccessfulMigration(t *testing.T) {
	err := runMigrations(databaseURL)
	assert.NoError(t, err)
}

func TestCreateUser(t *testing.T) {
	ctx := context.Background()
	// Создаем пользователя для теста
	testUser := models.User{
		Login: "testuser",
		Hash:  "hash123",
	}

	// Вызываем функцию, которую тестируем
	id, err := db.CreateUser(ctx, testUser)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, id)
}

func TestGetUserByLogin(t *testing.T) {
	ctx := context.Background()

	// Вызываем функцию, которую тестируем
	user, err := db.GetUserByLogin(ctx, "testuser")
	assert.NoError(t, err)
	assert.Equal(t, "testuser", user.Login)
	assert.Equal(t, "hash123", user.Hash)
}

func TestCreateRule(t *testing.T) {
	ctx := context.Background()

	// Создаем тестовое правило
	testRule := models.Config{
		TopicFrom: "events",
	}

	// Создаем пользователя, который будет владельцем правила
	ownerID, err := db.CreateUser(ctx, models.User{Login: "testowner", Hash: "hash123"})
	assert.NoError(t, err)

	// Вызываем функцию, которую тестируем
	ruleID, err := db.CreateRule(ctx, testRule, ownerID)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, ruleID)
}

func TestGetRuleByID(t *testing.T) {
	ctx := context.Background()

	// Создаем тестовое правило
	testRule := models.Config{
		TopicFrom: "events",
	}

	// Создаем пользователя, который будет владельцем правила
	ownerID, err := db.CreateUser(ctx, models.User{Login: "testowner1", Hash: "hash123"})
	assert.NoError(t, err)

	// Создаем правило
	ruleID, err := db.CreateRule(ctx, testRule, ownerID)
	assert.NoError(t, err)

	// Вызываем функцию, которую тестируем
	rule, err := db.GetRuleByID(ctx, ruleID)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, rule.ID)
	assert.Equal(t, testRule, rule.Rule)
	assert.Equal(t, ownerID, *rule.Owner)
}

func TestDeleteRule(t *testing.T) {
	ctx := context.Background()

	// Создаем тестовое правило
	testRule := models.Config{
		TopicFrom: "events",
	}

	// Создаем пользователя, который будет владельцем правила
	ownerID, err := db.CreateUser(ctx, models.User{Login: "testowner3", Hash: "hash123"})
	assert.NoError(t, err)

	// Создаем правило
	ruleID, err := db.CreateRule(ctx, testRule, ownerID)
	assert.NoError(t, err)

	// Вызываем функцию, которую тестируем
	err = db.DeleteRule(ctx, ruleID)
	assert.NoError(t, err)

	// Пытаемся получить правило после удаления
	r, err := db.GetRuleByID(ctx, ruleID)
	assert.NoError(t, err)
	assert.Equal(t, r, models.Rule{})
}

func TestGetAllRules(t *testing.T) {
	ctx := context.Background()

	// Вызываем функцию, которую тестируем
	rules, err := db.GetAllRules(ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, rules)
}
