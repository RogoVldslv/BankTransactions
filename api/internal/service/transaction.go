package service

import (
	"BankTransactions/internal/domain"
	redisutils "BankTransactions/internal/lib/redis"
	"BankTransactions/internal/logger"
	"BankTransactions/internal/models"
	"context"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type TransferRepository interface {
	AccountExistsWithId(accountID int) (bool, error)
	SaveTransaction(t *models.Transaction) error
	IncrementBalance(accountID int, amount int64) error
	Debit(accountId int, amount int64) error
}

type TransactionStorage interface {
	WithinTransaction(fn func(repo TransferRepository) error) error
	GetTgId(accountID int) (int64, error)
}

type TransactionService struct {
	storage TransactionStorage
	log     logger.Logger
	rClient *redis.Client
}

func NewTransactionService(log logger.Logger, stg TransactionStorage, rClient *redis.Client) *TransactionService {
	return &TransactionService{
		storage: stg,
		log:     log,
		rClient: rClient,
	}
}

func (s *TransactionService) TransferMoney(ctx context.Context, t *models.Transaction) error {
	const op = "service.TransferMoney"

	err := s.storage.WithinTransaction(func(repo TransferRepository) error {
		exists, err := repo.AccountExistsWithId(t.FromUserId)
		if err != nil {
			return fmt.Errorf("%s:%w", op, err)
		}
		if !exists {
			return fmt.Errorf("%s:%w", op, domain.ErrFromAccountNotFind)
		}

		exists, err = repo.AccountExistsWithId(t.ToUserId)
		if err != nil {
			return fmt.Errorf("%s:%w", op, err)
		}
		if !exists {
			return fmt.Errorf("%s:%w", op, domain.ErrToAccountNotFind)
		}

		err = repo.Debit(t.FromUserId, t.Amount)

		if err != nil {
			return fmt.Errorf("%s:%w", op, err)
		}

		if err = repo.SaveTransaction(t); err != nil {
			return fmt.Errorf("%s:%w", op, err)
		}

		if err = repo.IncrementBalance(t.ToUserId, t.Amount); err != nil {
			return fmt.Errorf("%s:%w", op, err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}

	var tgID int64

	tgID, err = s.storage.GetTgId(t.ToUserId)
	if errors.Is(err, domain.ErrTelegramIDNotFound) {
		s.log.Debug("tg id no found, failed to push notification")
		return nil
	}
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err = redisutils.LPushTransfer(ctx, s.rClient, t, tgID); err != nil {
		s.log.Error("failed to push transfer notification", "error", err)
	} else {
		s.log.Debug("notification has been sent")
	}

	return nil
}
