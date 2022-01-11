package queue

import (
	"errors"
)

var (
	InvalidQueueError = errors.New("the provided queue is not valid")
)
