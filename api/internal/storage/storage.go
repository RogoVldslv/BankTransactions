package storage

import (
	"BankTransactions/internal/models"
	"BankTransactions/internal/service"

	"gorm.io/gorm"
)

type Storage interface {
	GetBalance(int) (int, error)
	SaveAccount(*models.Account) error
	SaveTransaction(*models.Transaction) error
	AccountExistsWithPassport(int) (bool, error)
	AccountExistsWithId(int) (bool, error)
	IncrementBalance(int, int64) error
	Debit(int, int64) error
	GetDB() *gorm.DB
	GetTgId(accountId int) (int64, error)
	WithinTransaction(fn func(repo service.TransferRepository) error) error
}
