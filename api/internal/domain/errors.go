package domain

import "errors"

var (
	UserUniqueError = errors.New("user is already exists")
)
