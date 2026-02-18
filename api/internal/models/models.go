package models

import (
	"gorm.io/gorm"
)

type Transaction struct {
	gorm.Model
	FromUserId int    `gorm:"not null;index" json:"from_id" binding:"required"`
	ToUserId   int    `gorm:"not null;index" json:"to_id" binding:"required"`
	Amount     int64  `gorm:"not null" json:"amount" binding:"required;gte=0"`
	Message    string `gorm:"nullable;type:varchar(64)" json:"message" binding:"max=32"`
}

type Account struct {
	gorm.Model
	Name       string `gorm:"not null" json:"name" binding:"required"`
	Balance    int    `gorm:"not null; default:0" json:"balance"`
	PassportId int    `gorm:"not null" json:"passport_id" binding:"required,min=10000000,max=99999999" `
}

func AllModels() []any {
	return []any{&Transaction{}, &Account{}}
}
