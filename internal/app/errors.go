package app

import "errors"

var (
	ErrNotFound    = errors.New("not found")
	ErrInvalidArgs = errors.New("invalid arguments")
)
