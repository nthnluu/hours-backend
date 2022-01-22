package models

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
	DisplayName string `json:"displayName" mapstructure:"displayName" firebase:"displayName"`
	Email       string `json:"email" mapstructure:"email" firebase:"displayName"`
	PhoneNumber string `json:"phoneNumber,omitempty" mapstructure:"phoneNumber" firebase:"displayName"`
	PhotoURL    string `json:"photoUrl,omitempty" mapstructure:"photoUrl" firebase:"displayName"`
	IsAdmin     bool   `json:"isAdmin,omitempty" mapstructure:"isAdmin" firebase:"displayName"`
	Pronouns    string `json:"pronouns,omitempty" mapstructure:"pronouns" firebase:"displayName"`
	MeetingLink string `json:"meetingLink,omitempty" mapstructure:"meetingLink" firebase:"displayName"`
	// Map from course ID to CoursePermission
	CoursePermissions map[string]CoursePermission `json:"coursePermissions" mapstructure:"coursePermissions" firebase:"displayName"`
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
