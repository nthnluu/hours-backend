package queue

import (
	"signmeup/internal/auth"
	"signmeup/internal/course"
)

type Queue struct {
	ID          string         `json:"id" mapstructure:"id"`
	Title       string         `json:"title" mapstructure:"title"`
	Description string         `json:"code" mapstructure:"code"`
	CourseID    string         `json:"courseID" mapstructure:"courseID"`
	Course      *course.Course `json:"course" mapstructure:"course,omitempty"`
	IsActive    bool           `json:"isActive" mapstructure:"isActive"`
	Tickets     []string       `json:"tickets" mapstructure:"tickets"`
}

// CreateQueueRequest is the parameter struct to the CreateQueue function.
type CreateQueueRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	CourseID    string `json:"courseID"`
}

type TicketStatus int64

const (
	StatusWaiting TicketStatus = iota
	StatusClaimed
	StatusMissing
	StatusComplete
)

type Ticket struct {
	ID        string       `json:"id" mapstructure:"id"`
	Queue     *Queue       `json:"queue" mapstructure:"queue"`
	CreatedBy *auth.User   `json:"createdBy" mapstructure:"createdBy"`
	Status    TicketStatus `json:"status" mapstructure:"status"`
}

// CreateTicketRequest is the parameter struct to the CreateTicket function.
type CreateTicketRequest struct {
	Description string     `json:"description"`
	QueueID     string     `json:"queueID,omitempty"`
	CreatedBy   *auth.User `json:"createdBy,omitempty"`
}
