package auth

import (
	"fmt"
	"log"
	"net/http"
	"queue/internal/firebase"
)

var repository Repository

func GetUserByID(id string) (*User, error) {
	user, err := repository.Get(id)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func GetAllUsers() ([]*User, error) {
	users, err := repository.List()
	if err != nil {
		return nil, err
	}

	return users, nil
}

// CreateUser creates a user using the provided Email, Password, and DisplayName.
func CreateUser(user *CreateUserRequest) (*User, error) {
	createdUser, err := repository.Create(user)
	if err != nil {
		return nil, EmailExistsError
	}

	return createdUser, nil
}

func UpdateUser(user *UpdateUserRequest) (*User, error) {
	return &User{
		Profile:            nil,
		ID:                 "",
		Disabled:           false,
		CreationTimestamp:  0,
		LastLogInTimestamp: 0,
	}, nil
}

func DeleteUser(user *User) (*User, error) {
	err := repository.Delete(user.ID)
	return user, err
}

// VerifySessionCookie verifies that the given session cookie is valid and returns the associated User if valid.
func VerifySessionCookie(sessionCookie *http.Cookie) (*User, error) {
	authClient, err := firebase.FirebaseApp.Auth(firebase.FirebaseContext)
	if err != nil {
		log.Fatalf("error getting Auth client: %v\n", err)
	}

	decoded, err := authClient.VerifySessionCookieAndCheckRevoked(firebase.FirebaseContext, sessionCookie.Value)
	if err != nil {
		return nil, fmt.Errorf("error verifying cookie: %v\n", err)
	}

	user, err := GetUserByID(decoded.UID)
	if err != nil {
		return nil, fmt.Errorf("error getting user from cookie: %v\n", err)
	}

	return user, nil
}

func init() {
	repo, err := NewFirebaseRepository()
	if err != nil {
		log.Panicf("error creating Firebase user repository: %v\n", err)
	}

	repository = repo
}
