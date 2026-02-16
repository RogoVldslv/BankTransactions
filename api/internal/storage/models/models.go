package models

import (
	"gorm.io/gorm"
)

type Transaction struct {
	gorm.Model
	FromUserId int    `gorm:"not null;index"`
	ToUserId   int    `gorm:"not null;index"`
	Amount     int64  `gorm:"not null"`
	Message    string `gorm:"nullable;size:64;"`
}

type Account struct {
	gorm.Model
	Name    string `gorm:"not null"`
	Balance int64  `gorm:"not null"`
}

func AllModels() []any {
	return []any{&Transaction{}, &Account{}}
}
