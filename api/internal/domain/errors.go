package domain

import "errors"

var (
	ErrAccountUnique      = errors.New("account is already exists")
	ErrToAccountNotFind   = errors.New("account to not find")
	ErrFromAccountNotFind = errors.New("account from not find")
	ErrInsufficientFunds  = errors.New("insufficient funds")
)
