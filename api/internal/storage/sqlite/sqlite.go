package sqlite

import (
	"BankTransactions/internal/config"
	"BankTransactions/internal/lib/sl"
	"BankTransactions/internal/storage/models"
	"fmt"
	"log/slog"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Repository struct {
	DB *gorm.DB
}

func (s *Repository) GetBalance(accountId int) (int, error) {
	const op = "storage.sqlite.GetBalance"
	var balance int
	err := s.DB.Model(&models.Account{}).Where("id = ?", accountId).
		Pluck("balance", &balance).Error
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return balance, nil
}

func MustNew(log *slog.Logger, cfg *config.Config) *Repository {
	const op = "storage.sqlite.New"
	log = log.With(slog.String("op", op))

	db, err := gorm.Open(sqlite.Open(cfg.StoragePath), &gorm.Config{})
	if err != nil {
		log.Error("failed to connect with database", sl.Err(err))
		os.Exit(1)
	}

	err = db.AutoMigrate(models.AllModels()...)
	if err != nil {
		log.Error("failed to migrate", sl.Err(err))
		os.Exit(1)
	}

	return &Repository{DB: db}

}
