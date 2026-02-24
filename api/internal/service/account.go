package service

import (
	"BankTransactions/internal/domain"
	"BankTransactions/internal/logger"
	"BankTransactions/internal/models"
	"fmt"
)

type AccountSaver interface {
	SaveAccount(a *models.Account) error
	GetBalance(id int) (int, error)
	IsAccountExists(passport int) (bool, error)
}

type AccountService struct {
	storage AccountSaver
	log     logger.Logger
}

func NewAccountService(log logger.Logger, stg AccountSaver) *AccountService {
	return &AccountService{
		storage: stg,
		log:     log,
	}
}

func (s *AccountService) CreateAccount(a *models.Account) error {
	const op = "service.CreateAccount"

	exist, err := s.storage.IsAccountExists(a.PassportId)

	if err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}
	if exist == true {
		return domain.UserUniqueError
	}

	if err = s.storage.SaveAccount(a); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
