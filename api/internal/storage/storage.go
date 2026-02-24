package storage

import (
	"BankTransactions/internal/models"
)

type Storage interface {
	GetBalance(id int) (int, error)
	SaveAccount(a *models.Account) error
	SaveTransaction(t *models.Transaction) error
	IsAccountExists(passport int) (bool, error)
}
