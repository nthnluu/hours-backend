package auth

import (
	"errors"
)

var (
	EmailExistsError  = errors.New("a user with that email address already exists")
	DeleteUserError   = errors.New("an error occurred while deleting user")
	UserNotFoundError = errors.New("user not found")
	InvalidEmailError = errors.New("invalid Brown email address")
)
