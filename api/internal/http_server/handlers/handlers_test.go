package handlers_test

import (
	"BankTransactions/internal/http_server/handlers"
	"BankTransactions/internal/models"
	"BankTransactions/internal/storage"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func mustLog() *slog.Logger {
	return slog.New(slog.NewTextHandler(
		os.Stdout,
		&slog.HandlerOptions{Level: slog.LevelDebug},
	),
	)
}

type mockStorage struct {
	saveAccountFunc     func(a *models.Account) error
	saveTransactionFunc func(t *models.Transaction) error
	getBalanceFunc      func(id int) (int, error)
}

func (m *mockStorage) SaveAccount(a *models.Account) error         { return m.saveAccountFunc(a) }
func (m *mockStorage) SaveTransaction(t *models.Transaction) error { return m.saveTransactionFunc(t) }
func (m *mockStorage) GetBalance(id int) (int, error)              { return m.getBalanceFunc(id) }

func TestCreatePostAccount(t *testing.T) {
	log := mustLog()
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		mockStg := &mockStorage{
			saveAccountFunc: func(a *models.Account) error { return nil },
		}

		r := gin.New()
		r.POST("/account", handlers.CreatePostAccount(log, mockStg))

		w := httptest.NewRecorder()
		body := `{"name": "Vlad", "balance": 100, "passport_id": 12345678}`
		req, _ := http.NewRequest(http.MethodPost, "/account", strings.NewReader(body))

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Duplicate Passport", func(t *testing.T) {
		mockStg := &mockStorage{
			saveAccountFunc: func(a *models.Account) error {
				return storage.UserUniqueError
			},
		}

		r := gin.New()
		r.POST("/account", handlers.CreatePostAccount(log, mockStg))

		w := httptest.NewRecorder()
		body := `{"name": "Vlad", "balance": 100, "passport_id": 12345678}`
		req, _ := http.NewRequest(http.MethodPost, "/account", strings.NewReader(body))

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "user with the same passport id already exists")
	})
}

func TestCreatePostTransaction(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := mustLog()
	t.Run("Invalid JSON", func(t *testing.T) {
		mockStg := &mockStorage{}

		r := gin.New()
		r.POST("/transaction", handlers.CreatePostTransaction(log, mockStg))

		w := httptest.NewRecorder()
		body := `{"amount": "not_a_number"}` // Ошибка типа
		req, _ := http.NewRequest(http.MethodPost, "/transaction", strings.NewReader(body))

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
