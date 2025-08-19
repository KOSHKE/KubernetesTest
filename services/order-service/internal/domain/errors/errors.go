package errors

import "errors"

var (
	ErrOrderNotFound     = errors.New("order not found")
	ErrOrderAccessDenied = errors.New("access denied")
	ErrInvalidArgument   = errors.New("invalid argument")
)
