package router

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/dedpnd/unifier/internal/adapter/store"
	"github.com/dedpnd/unifier/internal/config"
	"github.com/dedpnd/unifier/internal/core/auth"
	"github.com/dedpnd/unifier/internal/core/worker"
	"github.com/dedpnd/unifier/internal/logger"
	"github.com/go-resty/resty/v2"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"

	_ "github.com/lib/pq"
)

var db *sql.DB
var databaseURL string

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

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	pool.MaxWait = 20 * time.Second
	if err = pool.Retry(func() error {
		db, err = sql.Open("postgres", databaseURL)
		if err != nil {
			return fmt.Errorf("Connection has error: %w", err)
		}

		err = db.Ping()
		if err != nil {
			return fmt.Errorf("Ping has error: %w", err)
		}

		return nil
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	// Run tests
	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func TestRouter(t *testing.T) {
	tests := []struct {
		name          string
		method        string
		authorization bool
		url           string
		body          map[string]interface{}
		expectedCode  int
		expectedBody  string
	}{
		{
			name:   "Register test user",
			method: http.MethodPost,
			url:    "/api/user/register",
			body: map[string]interface{}{
				"login":    "test",
				"password": "test",
			},
			expectedCode: http.StatusOK,
			expectedBody: "",
		},
		{
			name:         "Register test user: invalid body",
			method:       http.MethodPost,
			url:          "/api/user/register",
			body:         map[string]interface{}{},
			expectedCode: http.StatusBadRequest,
			expectedBody: "",
		},
		{
			name:   "Register test user: login must be exist",
			method: http.MethodPost,
			url:    "/api/user/register",
			body: map[string]interface{}{
				"login":    "test",
				"password": "test",
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: "failed to create a new user with login test: user already exists\n",
		},
		{
			name:   "Login test user",
			method: http.MethodPost,
			url:    "/api/user/login",
			body: map[string]interface{}{
				"login":    "test",
				"password": "test",
			},
			expectedCode: http.StatusOK,
			expectedBody: "",
		},
		{
			name:         "Login test user: invalid body",
			method:       http.MethodPost,
			url:          "/api/user/login",
			body:         map[string]interface{}{},
			expectedCode: http.StatusBadRequest,
			expectedBody: "",
		},
		{
			name:   "Login test1 user: login not exist",
			method: http.MethodPost,
			url:    "/api/user/login",
			body: map[string]interface{}{
				"login":    "test1",
				"password": "test1",
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: "login or password incorrect\n",
		},
		{
			name:          "Get all rules",
			method:        http.MethodGet,
			authorization: true,
			url:           "/api/rules",
			expectedCode:  http.StatusOK,
			//nolint:lll // This legal size
			expectedBody: "[{\"id\":1,\"rule\":{\"topicFrom\":\"events\",\"filter\":{\"regexp\":\"\\\"dstHost.ip\\\": \\\"10.10.10.10\\\"\"},\"entityHash\":[\"srcHost.ip\",\"dstHost.port\"],\"unifier\":[{\"name\":\"id\",\"type\":\"string\",\"expression\":\"auditEventLog\"},{\"name\":\"date\",\"type\":\"timestamp\",\"expression\":\"datetime\"},{\"name\":\"ipaddr\",\"type\":\"string\",\"expression\":\"srcHost.ip\"},{\"name\":\"category\",\"type\":\"string\",\"expression\":\"cat\"}],\"extraProcess\":[{\"func\":\"__if\",\"args\":\"category, /Host/Connect/Host/Accept, high\",\"to\":\"category\"},{\"func\":\"__stringConstant\",\"args\":\"test\",\"to\":\"customString1\"}],\"topicTo\":\"test\"},\"owner\":null}]\n",
		},
		{
			name:          "Get all rules: token unauth",
			method:        http.MethodGet,
			authorization: false,
			url:           "/api/rules",
			expectedCode:  http.StatusUnauthorized,
			expectedBody:  "",
		},
		{
			name:          "Add new rule",
			method:        http.MethodPost,
			authorization: true,
			url:           "/api/rules",
			body: map[string]interface{}{
				"topicFrom": "events",
			},
			expectedCode: http.StatusOK,
			expectedBody: "",
		},
		{
			name:          "Add new rule: invalid body",
			method:        http.MethodPost,
			authorization: true,
			url:           "/api/rules",
			body:          nil,
			expectedCode:  http.StatusBadRequest,
			expectedBody:  "",
		},
		{
			name:          "Add new rule: token unauth",
			method:        http.MethodPost,
			authorization: false,
			url:           "/api/rules",
			body:          nil,
			expectedCode:  http.StatusUnauthorized,
			expectedBody:  "",
		},
		{
			name:          "Remove rule",
			method:        http.MethodDelete,
			authorization: true,
			url:           "/api/rules/2",
			expectedCode:  http.StatusOK,
			expectedBody:  "",
		},
		{
			name:          "Remove rule: token unauth",
			method:        http.MethodDelete,
			authorization: false,
			url:           "/api/rules/2",
			expectedCode:  http.StatusUnauthorized,
			expectedBody:  "",
		},
		{
			name:          "Remove rule: invalid id",
			method:        http.MethodDelete,
			authorization: true,
			url:           "/api/rules/sd",
			expectedCode:  http.StatusBadRequest,
			expectedBody:  "",
		},
		{
			name:          "Remove rule: rule id not exist",
			method:        http.MethodDelete,
			authorization: true,
			url:           "/api/rules/5",
			expectedCode:  http.StatusNotFound,
			expectedBody:  "",
		},
		{
			name:          "Remove rule: not owner rule",
			method:        http.MethodDelete,
			authorization: true,
			url:           "/api/rules/1",
			expectedCode:  http.StatusForbidden,
			expectedBody:  "",
		},
	}

	// Создаем логер
	lg, err := logger.Init("error")
	if err != nil {
		assert.NoError(t, err)
	}

	// Читаем конфигурацию
	cfg, err := config.GetConfig()
	if err != nil {
		assert.NoError(t, err)
	}

	// Создаем хранилище
	str, err := store.NewStore(databaseURL, lg)
	if err != nil {
		assert.NoError(t, err)
	}

	// Запускаем пул воркеров
	p, err := worker.StartPool(cfg.KafkaAdress, str, lg)
	if err != nil {
		assert.NoError(t, err)
	}

	// Сорздаем роутер
	r, err := Router(lg, str, p)
	if err != nil {
		assert.NoError(t, err)
	}

	srv := httptest.NewServer(r)
	defer srv.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := resty.New().R()
			req.Method = tt.method
			req.URL = srv.URL + tt.url

			if tt.authorization {
				token, err := auth.GetJWT(1, "test")
				if err != nil {
					assert.NoError(t, err)
				}

				req.SetCookie(&http.Cookie{
					Name:  "token",
					Value: *token,
				})
			}

			if len(tt.body) != 0 {
				req.Body = tt.body
			}

			resp, err := req.Send()
			assert.NoError(t, err, "error making HTTP request")

			assert.Equal(t, tt.expectedCode, resp.StatusCode(), "Response code didn't match expected")

			if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, string(resp.Body()))
			}
		})
	}
}
