package models

import "time"

var (
	FirestoreQueuesCollection  = "queues"
	FirestoreTicketsCollection = "tickets"
)

type Queue struct {
	ID                 string          `json:"id" mapstructure:"id"`
	Title              string          `json:"title" mapstructure:"title"`
	Description        string          `json:"code" mapstructure:"code"`
	Location           string          `json:"location" mapstructure:"location"`
	EndTime            time.Time       `json:"endTime" mapstructure:"endTime"`
	ShowMeetingLinks   bool            `json:"showMeetingLinks" mapstructure:"showMeetingLinks"`
	AllowTicketEditing bool            `json:"allowTicketEditing" mapstructure:"allowTicketEditing"`
	CourseID           string          `json:"courseID" mapstructure:"courseID"`
	Course             *Course         `json:"course" mapstructure:"course,omitempty"`
	IsCutOff           bool            `json:"isCutOff" mapstructure:"isCutOff,omitempty"`
	CutoffTicketID     string          `json:"cutoffTicketID" mapstructure:"cutoffTicketID,omitempty"`
	Tickets            []string        `json:"tickets" mapstructure:"tickets"`
	VisibleTickets     []string        `json:"visibleTickets" mapstructure:"visibleTickets"`
	RequireFaceMasks   bool            `json:"requireFaceMasks" mapstructure:"requireFaceMasks"`
	Analytics          *QueueAnalytics `json:"analytics" mapstructure:"analytics"`
}

type TicketStatus string

const (
	StatusWaiting  TicketStatus = "WAITING"
	StatusClaimed  TicketStatus = "CLAIMED"
	StatusMissing  TicketStatus = "MISSING"
	StatusReturned TicketStatus = "RETURNED"
	StatusComplete TicketStatus = "COMPLETE"
)

type TicketUserdata struct {
	UserID      string `json:"UserID" mapstructure:"UserID"`
	Email       string `json:"Email" mapstructure:"Email"`
	PhotoURL    string `json:"PhotoURL" mapstructure:"PhotoURL"`
	DisplayName string `json:"DisplayName" mapstructure:"DisplayName"`
	Pronouns    string `json:"Pronouns" mapstructure:"Pronouns"`
}

type Ticket struct {
	ID          string         `json:"id" mapstructure:"id"`
	User        TicketUserdata `json:"user" mapstructure:"user"`
	Queue       *Queue         `json:"queue" mapstructure:"queue"`
	CreatedAt   time.Time      `json:"createdAt" mapstructure:"createdAt"`
	ClaimedAt   time.Time      `json:"claimedAt,omitempty" mapstructure:"claimedAt"`
	ClaimedBy   string         `json:"claimedBy,omitempty" mapstructure:"claimedBy"`
	CompletedAt time.Time      `json:"completedAt,omitempty" mapstructure:"completedAt"`
	Status      TicketStatus   `json:"status" mapstructure:"status"`
	Description string         `json:"description"`
	Anonymize   bool           `json:"anonymize"`
}

// CreateQueueRequest is the parameter struct to the CreateQueue function.
type CreateQueueRequest struct {
	Title              string    `json:"title"`
	Description        string    `json:"description"`
	Location           string    `json:"location"`
	ShowMeetingLinks   bool      `json:"showMeetingLinks" mapstructure:"showMeetingLinks"`
	AllowTicketEditing bool      `json:"allowTicketEditing" mapstructure:"allowTicketEditing"`
	EndTime            time.Time `json:"endTime"`
	CourseID           string    `json:"courseID"`
	RequireFaceMasks   bool      `json:"requireFaceMasks" mapstructure:"requireFaceMasks"`
}

// EditQueueRequest is the parameter struct to the EditQueue function.
type EditQueueRequest struct {
	QueueID            string    `json:"queueID,omitempty"`
	Title              string    `json:"title"`
	Description        string    `json:"description"`
	Location           string    `json:"location"`
	ShowMeetingLinks   bool      `json:"showMeetingLinks" mapstructure:"showMeetingLinks"`
	AllowTicketEditing bool      `json:"allowTicketEditing" mapstructure:"allowTicketEditing"`
	EndTime            time.Time `json:"endTime"`
	IsCutOff           bool      `json:"isCutOff"`
	RequireFaceMasks   bool      `json:"requireFaceMasks" mapstructure:"requireFaceMasks"`
}

// DeleteQueueRequest is the parameter struct to the CreateQueue function.
type DeleteQueueRequest struct {
	QueueID string `json:"queueID,omitempty"`
}

// CutoffQueueRequest is the parameter struct to the CutoffQueue function.
type CutoffQueueRequest struct {
	IsCutOff       bool   `json:"isCutOff"`
	CutoffTicketID string `json:"cutoffTicketID"`
	QueueID        string `json:",omitempty"`
}

type ShuffleQueueRequest struct {
	QueueID string `json:"queueID,omitempty"`
}

// CreateTicketRequest is the parameter struct to the CreateTicket function.
type CreateTicketRequest struct {
	QueueID     string `json:"queueID,omitempty"`
	CreatedBy   *User  `json:"createdBy,omitempty"`
	Description string `json:"description"`
	Anonymize   bool   `json:"anonymize"`
}

// EditTicketRequest is the parameter struct to the EditTicket function.
type EditTicketRequest struct {
	ID          string       `json:"id" mapstructure:"id"`
	QueueID     string       `json:"queueID,omitempty"`
	OwnerID     string       `json:"ownerID" mapstructure:"ownerID"`
	Status      TicketStatus `json:"status" mapstructure:"status"`
	Description string       `json:"description"`
	ClaimedBy   *User        `json:"claimedBy,omitempty"`
}

// DeleteTicketRequest is the parameter struct to the DeleteTicket function.
type DeleteTicketRequest struct {
	ID      string `json:"id" mapstructure:"id"`
	QueueID string `json:"queueID,omitempty"`
}

// MakeAnnouncementRequest is the parameter struct to the MakeAnnouncement function.
type MakeAnnouncementRequest struct {
	QueueID      string `json:"queueID,omitempty"`
	Announcement string `json:"announcement" mapstructure:"announcement"`
}
