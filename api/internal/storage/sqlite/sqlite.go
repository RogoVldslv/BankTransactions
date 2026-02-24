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

func (r *Storage) GetBalance(id int) (int, error) {
	const op = "storage.sqlite.GetBalance"
	var balance int
	err := r.DB.Model(&models.Account{}).Where("id = ?", id).
		Pluck("balance", &balance).Error
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return balance, nil
}

func (r *Storage) SaveAccount(a *models.Account) error {
	const op = "storage.sqlite.SaveAccount"
	if err := r.DB.Create(a).Error; err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *Storage) SaveTransaction(t *models.Transaction) error {
	const op = "storage.sqlite.SaveTransaction"
	if err := r.DB.Create(t).Error; err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *Storage) IsAccountExists(passport int) (bool, error) {
	const op = "storage.sqlite.IsAccountExists"
	var exiting models.Account
	err := r.DB.Where("passport_id = ?", passport).First(&exiting).Error

	if err == nil {
		return true, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return false, fmt.Errorf("%s: %w", op, err)
}

func MustNew(log logger.Logger, cfg *config.Config) *Storage {
	const op = "storage.sqlite.New"

	newLogger := logo.New(
		stdlog.New(os.Stdout, "\r\n", stdlog.LstdFlags), // стандартный вывод
		logo.Config{
			SlowThreshold:             time.Second, // Логировать только запросы дольше 1 сек
			LogLevel:                  logo.Info,   // ИЛИ logger.Error, чтобы не видеть "record not found"
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
