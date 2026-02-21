package handlers

import (
	resp "BankTransactions/internal/lib/responses"
	"BankTransactions/internal/lib/sl"
	"BankTransactions/internal/logger"
	"BankTransactions/internal/models"
	"BankTransactions/internal/storage"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func CreatePostTransaction(log logger.Logger, stg storage.Storage) gin.HandlerFunc {
	const op = "handlers.PostTransaction"
	log = log.With("op", op)
	return func(c *gin.Context) {
		var t models.Transaction
		if err := c.ShouldBindJSON(&t); err != nil {
			var ve validator.ValidationErrors
			if errors.As(err, &ve) {
				log.Info("validation error", sl.Err(ve))
				c.JSON(http.StatusBadRequest, resp.Error(ve.Error()))
				return
			}
			log.Info("failed to parse json", sl.Err(err))
			c.JSON(http.StatusBadRequest, resp.Error("bad data"))
			return
		}

		if err := stg.SaveTransaction(&t); err != nil {
			log.Error("failed to save transaction", sl.Err(err))
			c.JSON(http.StatusInternalServerError, resp.Error("internal error"))
			return
		}

		log.Info("transaction us saved")

		c.JSON(http.StatusOK, resp.Ok())

	}
}
