package qerrors

import "errors"

var (
	// Course errors
	CourseNotFoundError = errors.New("course not found")

	// User errors
	DeleteUserError   = errors.New("an error occurred while deleting user")
	UserNotFoundError = errors.New("user not found")
	InvalidEmailError = errors.New("invalid Brown email address")

	// Queue errors
	InvalidQueueError = errors.New("the provided queue is not valid")
)
