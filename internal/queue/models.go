package queue

import (
	"signmeup/internal/auth"
	"signmeup/internal/course"
	"time"
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

type TicketStatus string
const (
	StatusWaiting TicketStatus = "WAITING"
	StatusClaimed TicketStatus = "CLAIMED"
	StatusMissing TicketStatus = "MISSING"
	StatusComplete TicketStatus = "COMPLETE"
)

type Ticket struct {
	ID        string       `json:"id" mapstructure:"id"`
	Queue     *Queue       `json:"queue" mapstructure:"queue"`
	CreatedBy *auth.User   `json:"createdBy" mapstructure:"createdBy"`
	CreatedAt time.Time    `json:"createdAt" mapstructure:"createdAt"`
	Status    TicketStatus `json:"status" mapstructure:"status"`
	Description string     `json:"description"`
}

// CreateQueueRequest is the parameter struct to the CreateQueue function.
type CreateQueueRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	CourseID    string `json:"courseID"`
}

// CreateTicketRequest is the parameter struct to the CreateTicket function.
type CreateTicketRequest struct {
	QueueID     string     `json:"queueID,omitempty"`
	CreatedBy   *auth.User `json:"createdBy,omitempty"`
	Description string     `json:"description"`
}

// CreateTicketRequest is the parameter struct to the CreateTicket function.
type EditTicketRequest struct {
	ID        string       `json:"id" mapstructure:"id"`
	QueueID     string     `json:"queueID,omitempty"`
	Status    TicketStatus `json:"status" mapstructure:"status"`
	Description string     `json:"description"`
}

// CreateTicketRequest is the parameter struct to the CreateTicket function.
type DeleteTicketRequest struct {
	ID        string       `json:"id" mapstructure:"id"`
	QueueID     string     `json:"queueID,omitempty"`
}
