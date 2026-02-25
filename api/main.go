package main

import (
	"BankTransactions/internal/config"
	"BankTransactions/internal/http_server/handlers"
	redisutils "BankTransactions/internal/lib/redis"
	"BankTransactions/internal/logger"
	"BankTransactions/internal/storage"
	"BankTransactions/internal/storage/sqlite"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
)

const (
	localEnv = "local"
	prodEnv  = "prod"
)

func main() {
	cfg := config.MustLoadConfig()

	var log logger.Logger = mustSetupLogger(cfg.Env)

	var stg storage.Storage = sqlite.MustNew(log, cfg)

	rdb := redisutils.NewRedisClient(cfg.RedisAddress)

	r := gin.Default()

	r.POST("/account", handlers.CreatePostAccount(log, stg))

	r.POST("/transaction", func(c *gin.Context) {
		if c.Query("manual") == "true" {
			handlers.CreatePostManualTransaction(log, stg)(c)
			return
		}
		handlers.TransferMoney(log, stg, rdb)(c)
	})
	log.Info("server is starting")
	if err := r.Run(cfg.Address); err != nil {
		log.Error(fmt.Sprintf("failed to start server on address: %s", cfg.Address))
		os.Exit(1)
	}

}

func mustSetupLogger(env string) *slog.Logger {
	var logger *slog.Logger
	switch env {
	case prodEnv:
		logger = slog.New(slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{Level: slog.LevelInfo},
		),
		)
	case localEnv:
		logger = slog.New(slog.NewTextHandler(
			os.Stdout,
			&slog.HandlerOptions{Level: slog.LevelDebug},
		),
		)
	default:
		log.Fatal("env must be local or prod")
	}
	return logger
}
