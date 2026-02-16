package main

import (
	"BankTransactions/internal/config"
	"fmt"
)

func main() {
	cfg := config.MustLoadConfig()

	fmt.Println(cfg)

	// init log

	// init db

	// init router

	// handlers

	// run server
}
