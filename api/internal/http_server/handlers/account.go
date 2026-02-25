package handlers

import (
	"BankTransactions/internal/domain"
	resp "BankTransactions/internal/lib/responses"
	"BankTransactions/internal/lib/sl"
	"BankTransactions/internal/logger"
	"BankTransactions/internal/models"
	service2 "BankTransactions/internal/service"
	"BankTransactions/internal/storage"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func CreatePostAccount(log logger.Logger, stg storage.Storage) gin.HandlerFunc {
	const op = "handler.PostAccount"
	log = log.With("op", op)
	service := service2.NewAccountService(log, stg)
	return func(c *gin.Context) {
		var account models.Account
		if err := c.ShouldBindJSON(&account); err != nil {
			var ve validator.ValidationErrors
			if errors.As(err, &ve) {
				log.Info("validation error", sl.Err(ve))
				c.JSON(http.StatusBadRequest, resp.Error(fmt.Sprintf("validation error: %s", ve.Error())))
				return
			}

			log.Info("failed to parse json", sl.Err(err))
			c.JSON(http.StatusBadRequest, resp.Error("bad data"))
			return
		}

		if err := service.CreateAccount(&account); err != nil {
			if errors.Is(err, domain.ErrAccountUnique) {
				log.Info("user with the same passport id already exists")
				c.JSON(http.StatusBadRequest, resp.Error("user with the same passport id already exists"))
				return
			}
			log.Error("failed to save account", sl.Err(err))
			c.JSON(http.StatusInternalServerError, resp.Error("server error"))
			return
		}

		log.Info("account is saved")

		c.JSON(http.StatusOK, resp.Ok())
	}
}
