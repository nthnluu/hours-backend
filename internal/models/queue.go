package models

import "time"

var (
	FirestoreQueuesCollection  = "queues"
	FirestoreTicketsCollection = "tickets"
)

type Queue struct {
	ID          string    `json:"id" mapstructure:"id"`
	Title       string    `json:"title" mapstructure:"title"`
	Description string    `json:"code" mapstructure:"code"`
	Location    string    `json:"location" mapstructure:"location"`
	EndTime     time.Time `json:"endTime" mapstructure:"endTime"`
	CourseID    string    `json:"courseID" mapstructure:"courseID"`
	Course      *Course   `json:"course" mapstructure:"course,omitempty"`
	IsCutOff    bool      `json:"isCutOff" mapstructure:"isCutOff,omitempty"`
	Tickets     []string  `json:"tickets" mapstructure:"tickets"`
}

type TicketStatus string

const (
	StatusWaiting  TicketStatus = "WAITING"
	StatusClaimed  TicketStatus = "CLAIMED"
	StatusMissing  TicketStatus = "MISSING"
	StatusComplete TicketStatus = "COMPLETE"
)

type Ticket struct {
	ID          string       `json:"id" mapstructure:"id"`
	Queue       *Queue       `json:"queue" mapstructure:"queue"`
	CreatedBy   *User        `json:"createdBy" mapstructure:"createdBy"`
	CreatedAt   time.Time    `json:"createdAt" mapstructure:"createdAt"`
	Status      TicketStatus `json:"status" mapstructure:"status"`
	Description string       `json:"description"`
}

// CreateQueueRequest is the parameter struct to the CreateQueue function.
type CreateQueueRequest struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	EndTime     time.Time `json:"endTime"`
	CourseID    string    `json:"courseID"`
}

// EditQueueRequest is the parameter struct to the EditQueue function.
type EditQueueRequest struct {
	QueueID     string `json:"queueID,omitempty"`
	Title       string `json:"title"`
	Description string `json:"description"`
	IsCutOff    bool   `json:"isCutOff"`
}

// DeleteQueueRequest is the parameter struct to the CreateQueue function.
type DeleteQueueRequest struct {
	QueueID string `json:"queueID,omitempty"`
}

// CutoffQueueRequest is the parameter struct to the CutoffQueue function.
type CutoffQueueRequest struct {
	IsCutOff bool   `json:"isCutOff"`
	CourseID string `json:"courseID"`
}

type ShuffleQueueRequest struct {
	QueueID string `json:"queueID,omitempty"`
}

// CreateTicketRequest is the parameter struct to the CreateTicket function.
type CreateTicketRequest struct {
	QueueID     string `json:"queueID,omitempty"`
	CreatedBy   *User  `json:"createdBy,omitempty"`
	Description string `json:"description"`
}

// EditTicketRequest is the parameter struct to the EditTicket function.
type EditTicketRequest struct {
	ID          string       `json:"id" mapstructure:"id"`
	QueueID     string       `json:"queueID,omitempty"`
	Status      TicketStatus `json:"status" mapstructure:"status"`
	Description string       `json:"description"`
}

// DeleteTicketRequest is the parameter struct to the DeleteTicket function.
type DeleteTicketRequest struct {
	ID      string `json:"id" mapstructure:"id"`
	QueueID string `json:"queueID,omitempty"`
}
