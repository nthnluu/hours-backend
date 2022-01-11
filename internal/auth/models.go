package auth

import (
	"fmt"
	"strings"
)

type CoursePermission string

const (
	CourseAdmin CoursePermission = "ADMIN"
	CourseStaff                  = "STAFF"
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

// Validate checks a CreateUserRequest struct for errors.
func (u *CreateUserRequest) Validate() error {
	if err := validateEmail(u.Email); err != nil {
		return err
	}

	if err := validatePassword(u.Password); err != nil {
		return err
	}

	if err := validateDisplayName(u.DisplayName); err != nil {
		return err
	}

	return nil
}

// UpdateUserRequest is the parameter struct for the UpdateUser function.
type UpdateUserRequest struct {
	ID 			string `json:"id,omitempty"`
	DisplayName string `json:"displayName"`
	IsAdmin		bool `json:"isAdmin"`
}

// UpdateUserRequest is the parameter struct for the UpdateUser function.
type UpdateUserByEmailRequest struct {
	Email 		string `json:"email"`
	IsAdmin		bool `json:"isAdmin"`
}

// Validators.

func validateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email must be a non-empty string")
	}
	if parts := strings.Split(email, "@"); len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("malformed email string: %q", email)
	}
	return nil
}

func validatePassword(val string) error {
	if len(val) < 6 {
		return fmt.Errorf("password must be a string at least 6 characters long")
	}
	return nil
}

func validateDisplayName(val string) error {
	if val == "" {
		return fmt.Errorf("display name must be a non-empty string")
	}
	return nil
}

func validateID(id string) error {
	if id == "" {
		return fmt.Errorf("id must be a non-empty string")
	}
	if len(id) > 128 {
		return fmt.Errorf("id string must not be longer than 128 characters")
	}
	return nil
}
