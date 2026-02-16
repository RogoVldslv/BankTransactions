package main

import (
	"BankTransactions/internal/config"
	"BankTransactions/internal/storage/sqlite"
	"log"
	"log/slog"
	"os"
)

const (
	localEnv = "local"
	prodEnv  = "prod"
)

func main() {
	cfg := config.MustLoadConfig()

	log := mustSetupLogger(cfg.Env)

	rep := sqlite.MustNew(log, cfg)

	_ = rep

	// init router

	// handlers

	// run server
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
