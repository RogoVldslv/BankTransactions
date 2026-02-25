package sqlite

import (
	"BankTransactions/internal/config"
	"BankTransactions/internal/lib/sl"
	"BankTransactions/internal/logger"
	"BankTransactions/internal/models"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	stdlog "log"

	logo "gorm.io/gorm/logger"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Storage struct {
	DB *gorm.DB
}

func (s *Storage) GetDB() *gorm.DB {
	return s.DB
}

func (s *Storage) GetBalance(accountId int) (int, error) {
	const op = "storage.sqlite.GetBalance"
	var balance int
	err := s.DB.Model(&models.Account{}).Where("id = ?", accountId).
		Pluck("balance", &balance).Error
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return balance, nil
}

func (s *Storage) SaveAccount(a *models.Account) error {
	const op = "storage.sqlite.SaveAccount"
	if err := s.DB.Create(a).Error; err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) SaveTransaction(t *models.Transaction) error {
	const op = "storage.sqlite.SaveTransaction"
	if err := s.DB.Create(t).Error; err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) AccountExistsWithPassport(passport int) (bool, error) {
	const op = "storage.sqlite.AccountExistsWithPassport"
	var exiting models.Account
	err := s.DB.Where("passport_id = ?", passport).First(&exiting).Error

	if err == nil {
		return true, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return false, fmt.Errorf("%s: %w", op, err)
}

func (s *Storage) CanAfford(accountId int, amount int64) (bool, error) {
	const op = "storage.sqlite.CanAfford"
	var a models.Account
	err := s.DB.Where("id = ?", accountId).Pluck("balance", &a).Error
	if err != nil {
		return false, fmt.Errorf("%s:%w", op, err)
	}
	return a.Balance >= amount, nil
}

func (s *Storage) AccountExistsWithId(accountId int) (bool, error) {
	const op = "storage.sqlite.AccountExistsWithId"
	var exists models.Account
	err := s.DB.Where("id = ?", accountId).First(&exists).Error
	if err == nil {
		return true, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return false, fmt.Errorf("%s:%w", op, err)
}

func (s *Storage) IncrementBalance(accountId int, amount int64) error {
	const op = "storage.sqlite.IncrementBalance"
	err := s.DB.Model(&models.Account{}).Where("id = ?", accountId).
		Update("balance",
			gorm.Expr("balance + ?", amount),
		).Error
	if err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}
	return nil
}

func (s *Storage) DecrementBalance(accountId int, amount int64) error {
	const op = "storage.sqlite.DecrementBalance"
	err := s.DB.Model(&models.Account{}).Where("id = ?", accountId).
		Update("balance",
			gorm.Expr("balance - ?", amount),
		).Error
	if err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}
	return nil
}

func (s *Storage) GetTgId(accountId int) (int64, error) {
	const op = "storage.sqlite.GetTgId"
	var tgId int64

	err := s.DB.Model(models.Account{}).Where("id = ?", accountId).Pluck("tg_id", &tgId).Error
	if err != nil {
		return 0, fmt.Errorf("%s:%w", op, err)
	}
	return tgId, nil

}

func MustNew(log logger.Logger, cfg *config.Config) *Storage {
	const op = "storage.sqlite.New"

	newLogger := logo.New(
		stdlog.New(os.Stdout, "\r\n", stdlog.LstdFlags),
		logo.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logo.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	db, err := gorm.Open(sqlite.Open(cfg.StoragePath), &gorm.Config{Logger: newLogger})
	if err != nil {
		log.Error("failed to connect with database", sl.Err(err), slog.String("op", op))
		os.Exit(1)
	}

	err = db.AutoMigrate(models.AllModels()...)
	if err != nil {
		log.Error("failed to migrate", sl.Err(err), slog.String("op", op))
		os.Exit(1)
	}

	return &Storage{DB: db}
}
