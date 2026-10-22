// Package apperror holds ready-made errors so callers do not construct them inline.
// The directory is named error; the package name stays apperror so it does not
// shadow the builtin error type.
package apperror

import "errors"

var (
	ErrNotFound      = errors.New("not found")
	ErrUserNotFound  = errors.New("user not found")
	ErrUsernameTaken = errors.New("username already exists")
	ErrEmailTaken    = errors.New("email already exists")
	ErrInvalidStatus = errors.New("invalid status")
)
