package models

// Auth stuff

const (
	FirestoreUserProfilesCollection = "user_profiles"
)

// Profile is a collection of standard profile information for a user.
// This struct separates client-safe profile information from internal user metadata.
type Profile struct {
	DisplayName string `json:"displayName" mapstructure:"displayName"`
	Email       string `json:"email" mapstructure:"email"`
	PhoneNumber string `json:"phoneNumber,omitempty" mapstructure:"phoneNumber"`
	PhotoURL    string `json:"photoUrl" mapstructure:"photoUrl"`
	IsAdmin     bool   `json:"isAdmin" mapstructure:"isAdmin"`
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
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"displayName"`
}

// Course stuff
var (
	FirestoreCoursesCollection = "courses"
)

type Course struct {
	ID         string `json:"id" mapstructure:"id"`
	Title      string `json:"title" mapstructure:"title"`
	Code       string `json:"code" mapstructure:"code"`
	Term       string `json:"term" mapstructure:"term"`
	IsArchived bool   `json:"isArchived" mapstructure:"isArchived"`
}

type CreateCourseRequest struct {
	Title string `json:"title"`
	Code  string `json:"code"`
	Term  string `json:"term"`
}

// Queue stuff

var (
	FirestoreQueuesCollection  = "queue"
	FirestoreTicketsCollection = "tickets"
)

type Queue struct {
	ID          string   `json:"id" mapstructure:"id"`
	Title       string   `json:"title" mapstructure:"title"`
	Description string   `json:"code" mapstructure:"code"`
	CourseID    string   `json:"courseID" mapstructure:"courseID"`
	Course      *Course  `json:"course" mapstructure:"course,omitempty"`
	IsActive    bool     `json:"isActive" mapstructure:"isActive"`
	Tickets     []string `json:"tickets" mapstructure:"tickets"`
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
	CreatedBy *User        `json:"createdBy" mapstructure:"createdBy"`
	Status    TicketStatus `json:"status" mapstructure:"status"`
}

// CreateTicketRequest is the parameter struct to the CreateTicket function.
type CreateTicketRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	QueueID     string `json:"queueID,omitempty"`
	CreatedBy   *User  `json:"createdBy"`
}
