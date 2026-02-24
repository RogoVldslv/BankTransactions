package service

import (
	"BankTransactions/internal/domain"
	"BankTransactions/internal/logger"
	"BankTransactions/internal/models"
	"fmt"

	"gorm.io/gorm"
)

type TransactionStorage interface {
	SaveTransaction(t *models.Transaction) error
	AccountExistsWithId(accountId int) (bool, error)
	CanAfford(accountId int, amount int64) (bool, error)
	GetDB() *gorm.DB
	IncrementBalance(accountId int, amount int64) error
	DecrementBalance(accountId int, amount int64) error
}

type TransactionService struct {
	storage TransactionStorage
	log     logger.Logger
}

func NewTransactionService(log logger.Logger, stg TransactionStorage) *TransactionService {
	return &TransactionService{
		storage: stg,
		log:     log,
	}
}

func (s *TransactionService) ManualSaveTransaction(t *models.Transaction) error {
	const op = "service.ManualSaveTransaction"
	err := s.storage.SaveTransaction(t)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *TransactionService) TransferMoney(t *models.Transaction) error {
	const op = "service.TransferMoney"

	return s.storage.GetDB().Transaction(func(tx *gorm.DB) error {
		exists, err := s.storage.AccountExistsWithId(t.FromUserId)
		if err != nil {
			return fmt.Errorf("%s:%w", op, err)
		}
		if !exists {
			return fmt.Errorf("%s:%w", op, domain.ErrFromAccountNotFind)
		}

		exists, err = s.storage.AccountExistsWithId(t.ToUserId)
		if err != nil {
			return fmt.Errorf("%s:%w", op, err)
		}
		if !exists {
			return fmt.Errorf("%s:%w", op, domain.ErrToAccountNotFind)
		}

		canAfford, err := s.storage.CanAfford(t.FromUserId, t.Amount)

		if err != nil {
			return fmt.Errorf("%s:%w", op, err)
		}
		if !canAfford {
			return fmt.Errorf("%s:%w", op, domain.ErrInsufficientFunds)
		}

		if err = s.storage.SaveTransaction(t); err != nil {
			return fmt.Errorf("%s:%w", op, err)
		}

		if err = s.storage.DecrementBalance(t.FromUserId, t.Amount); err != nil {
			return fmt.Errorf("%s:%w", op, err)
		}

		if err = s.storage.IncrementBalance(t.ToUserId, t.Amount); err != nil {
			return fmt.Errorf("%s:%w", op, err)
		}

		return nil
	})

}
