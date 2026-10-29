package handlers

import (
	"BankTransactions/internal/domain"
	resp "BankTransactions/internal/lib/responses"
	"BankTransactions/internal/lib/sl"
	"BankTransactions/internal/logger"
	"BankTransactions/internal/models"
	servicePackage "BankTransactions/internal/service"
	"BankTransactions/internal/storage"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
)

func TransferMoney(log logger.Logger, stg storage.Storage, rClient *redis.Client) gin.HandlerFunc {
	const op = "handlers.TransferMoney"
	log = log.With("op", op)
	service := servicePackage.NewTransactionService(log, stg, rClient)
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

		if err := service.TransferMoney(c, &t); err != nil {
			if errors.Is(err, domain.ErrToAccountNotFind) {
				log.Info("to account not find", sl.Err(err))
				c.JSON(http.StatusNotFound, resp.Error("account to not found"))
				return
			}
			if errors.Is(err, domain.ErrFromAccountNotFind) {
				log.Info("from account not find", sl.Err(err))
				c.JSON(http.StatusNotFound, resp.Error("account from not find"))
				return
			}
			if errors.Is(err, domain.ErrInsufficientFunds) {
				log.Info("no money", sl.Err(err))
				c.JSON(http.StatusUnprocessableEntity, resp.Error("insufficient funds"))
				return
			}

			log.Error("failed to make transfer", sl.Err(err))
			c.JSON(http.StatusInternalServerError, resp.Error("internal error"))
			return
		}

		log.Info("transaction completed successfully")

		c.JSON(http.StatusOK, resp.Ok())
	}
}
