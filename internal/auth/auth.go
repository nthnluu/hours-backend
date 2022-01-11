package auth

import (
	"fmt"
	"log"
	"net/http"
	"signmeup/internal/firebase"
)

var repository Repository

// GetUserByID retrieves the User associated with the given ID.
func GetUserByID(id string) (*User, error) {
	user, err := repository.Get(id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetUserByID retrieves the User associated with the given ID.
func GetUserByEmail(email string) (*User, error) {
	userID, err := repository.GetIDByEmail(email)
	if err != nil {
		return nil, err
	}
	return GetUserByID(userID)
}

// GetUserByID retrieves the User associated with the given ID.
func UpdateUser(user *UpdateUserRequest) error {
	return repository.Update(user)
}

// GetUserByID retrieves the User associated with the given ID.
func UpdateUserByEmail(u *UpdateUserByEmailRequest) error {
	user, err := GetUserByEmail(u.Email)
	if err != nil {
		return err
	}
	return UpdateUser(&UpdateUserRequest{
		ID: user.ID,
		IsAdmin: u.IsAdmin,
	})
}

// verifySessionCookie verifies that the given session cookie is valid and returns the associated User if valid.
func verifySessionCookie(sessionCookie *http.Cookie) (*User, error) {
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
