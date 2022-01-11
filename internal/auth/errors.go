package auth

import (
	"errors"
)

var (
	DeleteUserError   = errors.New("an error occurred while deleting user")
	UserNotFoundError = errors.New("user not found")
	InvalidEmailError = errors.New("invalid Brown email address")
)
