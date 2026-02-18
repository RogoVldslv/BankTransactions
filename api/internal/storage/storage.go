package storage

import (
	"BankTransactions/internal/models"
	"errors"
)

var (
	UserUniqueError = errors.New("user is already exists")
)

type Storage interface {
	GetBalance(accountId *int) (int, error)
	SaveAccount(a *models.Account) error
}
