package models

import "time"

const (
	FirestoreUserProfilesCollection = "user_profiles"
)

type CoursePermission string

const (
	CourseAdmin CoursePermission = "ADMIN"
	CourseStaff CoursePermission = "STAFF"
)

type NotificationType string

const (
	NotificationClaimed      NotificationType = "CLAIMED"
	NotificationAnnouncement NotificationType = "ANNOUNCEMENT"
)

// Profile is a collection of standard profile information for a user.
// This struct separates client-safe profile information from internal user metadata.
type Profile struct {
	DisplayName string `json:"displayName" mapstructure:"displayName" firebase:"displayName"`
	Email       string `json:"email" mapstructure:"email" firebase:"email"`
	PhoneNumber string `json:"phoneNumber,omitempty" mapstructure:"phoneNumber" firebase:"displayName"`
	PhotoURL    string `json:"photoUrl,omitempty" mapstructure:"photoUrl" firebase:"photoUrl"`
	IsAdmin     bool   `json:"isAdmin,omitempty" mapstructure:"isAdmin" firebase:"isAdmin"`
	Pronouns    string `json:"pronouns,omitempty" mapstructure:"pronouns" firebase:"pronouns"`
	MeetingLink string `json:"meetingLink,omitempty" mapstructure:"meetingLink" firebase:"meetingLink"`
	// Map from course ID to CoursePermission
	CoursePermissions map[string]CoursePermission `json:"coursePermissions" mapstructure:"coursePermissions" firebase:"coursePermissions"`
	Notifications     []Notification              `json:"notifications" mapstructure:"notifications" firebase:"notifications"`
	FavoriteCourses   []string                    `json:"favoriteCourses" mapstructure:"favoriteCourses" firebase:"favoriteCourses"`
}

// User represents a registered user.
type User struct {
	*Profile
	ID                 string `json:"id" mapstructure:"id"`
	Disabled           bool
	CreationTimestamp  int64
	LastLogInTimestamp int64
}

type Notification struct {
	ID        string           `json:"id" mapstructure:"id"`
	Title     string           `json:"title" mapstructure:"title"`
	Body      string           `json:"body" mapstructure:"body"`
	Timestamp time.Time        `json:"timestamp" mapstructure:"timestamp"`
	Type      NotificationType `json:"type" mapstructure:"type"`
}

// CreateUserRequest is the parameter struct for the CreateUser function.
type CreateUserRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"displayName"`
}

// UpdateUserRequest is the parameter struct for the UpdateUser function.
type UpdateUserRequest struct {
	// Will be set from context
	UserID      string `json:",omitempty"`
	DisplayName string `json:"displayName"`
	Pronouns    string `json:"pronouns"`
	MeetingLink string `json:"meetingLink"`
}

// MakeAdminByEmailRequest is the parameter struct for the MakeAdminByEmail function.
type MakeAdminByEmailRequest struct {
	Email   string `json:"email"`
	IsAdmin bool   `json:"isAdmin"`
}

// ClearNotificationRequest is the parameter struct for the ClearNotification function.
type ClearNotificationRequest struct {
	UserID         string `json:",omitempty"`
	NotificationID string `json:"notificationId" mapstructure:"notificationId"`
}

// ClearAllNotificationsRequest is the parameter struct for the ClearNotification function.
type ClearAllNotificationsRequest struct {
	UserID string `json:",omitempty"`
}

// AddFavoriteCourseRequest is the parameter struct for the AddFavoriteCourseRequest function.
type AddFavoriteCourseRequest struct {
	CourseID string `json:"courseID" mapstructure:"courseID"`
}

// RemoveFavoriteCourseRequest is the parameter struct for the RemoveFavoriteCourseRequest function.
type RemoveFavoriteCourseRequest struct {
	CourseID string `json:"courseID" mapstructure:"courseID"`
}
