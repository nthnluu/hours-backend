package qerrors

import "errors"

var (
	// Generic errors
	InvalidBody    = errors.New("invalid body")
	EntityNotFound = errors.New("entity not found")

	// Course errors
	CourseNotFoundError = errors.New("course not found")

	// User errors
	DeleteUserError    = errors.New("an error occurred while deleting user")
	UserNotFoundError  = errors.New("user not found")
	InvalidEmailError  = errors.New("invalid Brown email address")
	InvalidDisplayName = errors.New("invalid display name provided")

	// Queue errors
	InvalidQueueError = errors.New("the provided queue is not valid")
)
