package app

import (
	"errors"
)

var (
	ErrInvalidArgs       = errors.New("invalid arguments")
	ErrHostAlreadyExists = errors.New("host already exists in ssh config file")
)
