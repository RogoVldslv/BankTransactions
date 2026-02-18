package handlers

import (
	resp "BankTransactions/internal/lib/responses"
	"BankTransactions/internal/lib/sl"
	"BankTransactions/internal/logger"
	"BankTransactions/internal/models"
	"BankTransactions/internal/storage"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func CreatePostAccount(log logger.Logger, stg storage.Storage) gin.HandlerFunc {
	const op = "handler.createPostAccount"
	return func(c *gin.Context) {
		var account models.Account
		if err := c.ShouldBindJSON(&account); err != nil {
			var ve validator.ValidationErrors
			if errors.As(err, &ve) {
				log.Info("validation error", sl.Err(ve), slog.String("op", op))
				c.JSON(http.StatusBadRequest, resp.Error(fmt.Sprintf("validation error: %s", ve.Error())))
				return
			}

			log.Info("failed to parse json", sl.Err(err), slog.String("op", op))
			c.JSON(http.StatusBadRequest, resp.Error("bad data"))
			return
		}
		if err := stg.SaveAccount(&account); err != nil {
			if errors.Is(err, storage.UserUniqueError) {
				log.Info("user with the same passport id already exists", slog.String("op", op))
				c.JSON(http.StatusBadRequest, resp.Error("user with the same passport id already exists"))
				return
			}
			log.Error("failed to save account", sl.Err(err), slog.String("op", op))
			c.JSON(http.StatusInternalServerError, resp.Error("server error"))
			return
		}

		log.Info("account is saved", slog.String("op", op))

		c.JSON(http.StatusOK, resp.Ok())
	}
}
