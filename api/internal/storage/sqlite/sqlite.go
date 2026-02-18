package sqlite

import (
	"BankTransactions/internal/config"
	"BankTransactions/internal/lib/sl"
	"BankTransactions/internal/logger"
	"BankTransactions/internal/models"
	"BankTransactions/internal/storage"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Storage struct {
	DB *gorm.DB
}

func (r *Storage) GetBalance(accountId *int) (int, error) {
	const op = "storage.sqlite.GetBalance"
	var balance int
	err := r.DB.Model(&models.Account{}).Where("id = ?", accountId).
		Pluck("balance", &balance).Error
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return balance, nil
}

func (r *Storage) SaveAccount(a *models.Account) error {
	const op = "storage.sqlite.SaveAccount"
	var exiting models.Account
	err := r.DB.Where("passport_id = ?", a.PassportId).First(&exiting).Error
	if err == nil {
		return storage.UserUniqueError
	}

	if !(errors.Is(err, gorm.ErrRecordNotFound)) {
		return fmt.Errorf("%s: %w", op, err)
	}
	if err = r.DB.Create(a).Error; err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func MustNew(log logger.Logger, cfg *config.Config) *Storage {
	const op = "storage.sqlite.New"

	db, err := gorm.Open(sqlite.Open(cfg.StoragePath), &gorm.Config{})
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
