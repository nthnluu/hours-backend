package models

import (
	"time"
)

// Auth stuff

type CoursePermission string

const (
	CourseAdmin CoursePermission = "ADMIN"
	CourseStaff                  = "STAFF"
)

const (
	FirestoreUserProfilesCollection = "user_profiles"
)

// Profile is a collection of standard profile information for a user.
// This struct separates client-safe profile information from internal user metadata.
type Profile struct {
	DisplayName       string                      `json:"displayName" mapstructure:"displayName"`
	Email             string                      `json:"email" mapstructure:"email"`
	PhoneNumber       string                      `json:"phoneNumber,omitempty" mapstructure:"phoneNumber"`
	PhotoURL          string                      `json:"photoUrl" mapstructure:"photoUrl"`
	IsAdmin           bool                        `json:"isAdmin" mapstructure:"isAdmin"`
	CoursePermissions map[string]CoursePermission `json:"coursePermissions" mapstructure:"coursePermissions"`
}

// User represents a registered user.
type User struct {
	*Profile
	ID                 string `json:"id" mapstructure:"id"`
	Disabled           bool
	CreationTimestamp  int64
	LastLogInTimestamp int64
}

// CreateUserRequest is the parameter struct for the CreateUser function.
type CreateUserRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"displayName"`
}

// UpdateUserRequest is the parameter struct for the UpdateUser function.
type UpdateUserRequest struct {
	ID          string `json:"id,omitempty"`
	DisplayName string `json:"displayName"`
	IsAdmin     bool   `json:"isAdmin"`
}

// UpdateUserRequest is the parameter struct for the UpdateUser function.
type UpdateUserByEmailRequest struct {
	Email   string `json:"email"`
	IsAdmin bool   `json:"isAdmin"`
}

// Course stuff

var (
	FirestoreCoursesCollection = "courses"
)

type Course struct {
	ID                string                      `json:"id" mapstructure:"id"`
	Title             string                      `json:"title" mapstructure:"title"`
	Code              string                      `json:"code" mapstructure:"code"`
	Term              string                      `json:"term" mapstructure:"term"`
	IsArchived        bool                        `json:"isArchived" mapstructure:"isArchived"`
	CoursePermissions map[string]CoursePermission `json:"coursePermissions" mapstructure:"coursePermissions"`
}

type GetCourseRequest struct {
	CourseID string `json:"courseID"`
}

type CreateCourseRequest struct {
	Title     string `json:"title"`
	Code      string `json:"code"`
	Term      string `json:"term"`
	CreatedBy *User  `json:"omitempty"`
}

type DeleteCourseRequest struct {
	CourseID string `json:"courseID"`
}

type EditCourseRequest struct {
	CourseID string `json:"courseID"`
	Title    string `json:"title"`
	Code     string `json:"code"`
	Term     string `json:"term"`
}

type AddCoursePermissionRequest struct {
	CourseID   string `json:"courseID"`
	Email      string `json:"email"`
	Permission string `json:"permission"`
}

type RemoveCoursePermissionRequest struct {
	CourseID string `json:"courseID"`
	UserID   string `json:"userID"`
}

// Queue stuff

var (
	FirestoreQueuesCollection  = "queues"
	FirestoreTicketsCollection = "tickets"
)

type Queue struct {
	ID          string   `json:"id" mapstructure:"id"`
	Title       string   `json:"title" mapstructure:"title"`
	Description string   `json:"code" mapstructure:"code"`
	CourseID    string   `json:"courseID" mapstructure:"courseID"`
	Course      *Course  `json:"course" mapstructure:"course,omitempty"`
	IsCutOff    bool     `json:"isCutOff" mapstructure:"isCutOff,omitempty"`
	Tickets     []string `json:"tickets" mapstructure:"tickets"`

	// Deprecated
	IsActive bool `json:"isActive" mapstructure:"isActive"`
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
	Title       string `json:"title"`
	Description string `json:"description"`
	CourseID    string `json:"courseID"`
}

// EditQueueRequest is the parameter struct to the EditQueue function.
type EditQueueRequest struct {
	QueueID     string `json:"queueID,omitempty"`
	Title       string `json:"title"`
	Description string `json:"description"`
	IsActive    bool   `json:"isActive"`
}

// DeleteQueueRequest is the parameter struct to the CreateQueue function.
type DeleteQueueRequest struct {
	QueueID string `json:"queueID,omitempty"`
}

// CutoffQueueRequest is the parameter struct to the CutoffQueue function.
type CutoffQueueRequest struct {
	IsCutoff bool   `json:"isCutOff"`
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

// CreateTicketRequest is the parameter struct to the CreateTicket function.
type EditTicketRequest struct {
	ID          string       `json:"id" mapstructure:"id"`
	QueueID     string       `json:"queueID,omitempty"`
	Status      TicketStatus `json:"status" mapstructure:"status"`
	Description string       `json:"description"`
}

// CreateTicketRequest is the parameter struct to the CreateTicket function.
type DeleteTicketRequest struct {
	ID      string `json:"id" mapstructure:"id"`
	QueueID string `json:"queueID,omitempty"`
}
