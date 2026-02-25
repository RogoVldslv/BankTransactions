package storage

import (
	"BankTransactions/internal/models"

	"gorm.io/gorm"
)

type Storage interface {
	GetBalance(int) (int, error)
	SaveAccount(*models.Account) error
	SaveTransaction(*models.Transaction) error
	AccountExistsWithPassport(int) (bool, error)
	CanAfford(int, int64) (bool, error)
	AccountExistsWithId(int) (bool, error)
	IncrementBalance(int, int64) error
	DecrementBalance(int, int64) error
	GetDB() *gorm.DB
}
