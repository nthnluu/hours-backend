package auth

import (
	"cloud.google.com/go/firestore"
	firebaseAuth "firebase.google.com/go/auth"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"queue/internal/firebase"
	"strings"
	"sync"
)

// Repository encapsulates the logic to access users from a database.
type Repository interface {
	// Get returns the users corresponding to the specified user ID.
	Get(id string) (*User, error)
	// Count returns the number of users.
	Count() int
	// List returns the list of all registered users.
	List() ([]*User, error)
	// Create saves a new user into the database.
	Create(user *CreateUserRequest) (*User, error)
	// Update updates the user with given ID in the database.
	Update(user *UpdateUserRequest) (*User, error)
	// Delete removes the user with given ID from the database.
	Delete(id string) error
}

// firebaseRepository queries and persists users in Firebase.
type firebaseRepository struct {
	authClient      *firebaseAuth.Client
	firestoreClient *firestore.Client

	profilesLock *sync.RWMutex
	profiles     map[string]*Profile
}

const (
	FirestoreUserProfilesCollection = "user_profiles"
)

// NewFirebaseRepository creates a new user repository with Firebase as the database.
func NewFirebaseRepository() (Repository, error) {
	repository := &firebaseRepository{
		profilesLock: &sync.RWMutex{},
		profiles:     make(map[string]*Profile),
	}

	authClient, err := firebase.FirebaseApp.Auth(firebase.FirebaseContext)
	if err != nil {
		return nil, fmt.Errorf("Auth client error: %v\n", err)
	}
	repository.authClient = authClient

	firestoreClient, err := firebase.FirebaseApp.Firestore(firebase.FirebaseContext)
	if err != nil {
		return nil, fmt.Errorf("Firestore client error: %v\n", err)
	}
	repository.firestoreClient = firestoreClient

	go func() {
		err := repository.startUserProfilesListener()
		if err != nil {
			log.Fatalf("user profiles collection listner error: %v\n", err)
		}
	}()

	return repository, nil
}

// Queries

func (r firebaseRepository) Get(id string) (*User, error) {
	if err := validateID(id); err != nil {
		return nil, err
	}

	fbUser, err := r.authClient.GetUser(firebase.FirebaseContext, id)
	if err != nil {
		return nil, UserNotFoundError
	}

	// TODO: Refactor email verification and user profile creation into separate function.

	// Check if the Firebase user has a valid Brown email address
	if strings.Split(fbUser.Email, "@")[1] != "brown.edu" {
		// invalid brown email, delete the user from Firebase Auth
		_ = r.authClient.DeleteUser(firebase.FirebaseContext, fbUser.UID)
		return nil, InvalidEmailError
	}

	profile, err := r.getUserProfile(fbUser.UID)
	if err != nil {
		// no profile for the user found, create one.
		profile = &Profile{
			DisplayName: fbUser.DisplayName,
			Email:       fbUser.Email,
			PhoneNumber: "",
			PhotoURL:    "",
			// if there are no registered users, make the first one an admin
			IsAdmin: r.getUserCount() == 0,
		}
		_, err = r.firestoreClient.Collection(FirestoreUserProfilesCollection).Doc(fbUser.UID).Set(firebase.FirebaseContext, map[string]interface{}{
			"displayName": profile.DisplayName,
			"email":       profile.Email,
			"id":          fbUser.UID,
			"isAdmin":     profile.IsAdmin,
		})
		if err != nil {
			return nil, fmt.Errorf("error creating user profile: %v\n", err)
		}
	}

	return fbUserToUserRecord(fbUser, profile), nil
}

func (r firebaseRepository) Count() int {
	return len(r.profiles)
}

func (r firebaseRepository) List() ([]*User, error) {
	var users []*User
	iter := r.authClient.Users(firebase.FirebaseContext, "")
	for {
		fbUser, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error listing user_mgt: %s\n", err)
		}

		profile, err := r.getUserProfile(fbUser.UID)
		if err != nil {
			return nil, err
		}
		user := fbUserToUserRecord(fbUser.UserRecord, profile)

		users = append(users, user)
	}

	return users, nil
}

// Operations

func (r firebaseRepository) Create(user *CreateUserRequest) (*User, error) {
	if err := user.Validate(); err != nil {
		return nil, err
	}

	// Create a user in Firebase Auth.
	u := (&firebaseAuth.UserToCreate{}).Email(user.Email).Password(user.Password)
	fbUser, err := r.authClient.CreateUser(firebase.FirebaseContext, u)
	if err != nil {
		return nil, fmt.Errorf("error creating user: %v\n", err)
	}

	// Create a user profile in Firestore.
	profile := &Profile{
		DisplayName: user.DisplayName,
		Email:       user.Email,
		PhoneNumber: "",
		PhotoURL:    "",
	}
	_, err = r.firestoreClient.Collection(FirestoreUserProfilesCollection).Doc(fbUser.UID).Set(firebase.FirebaseContext, map[string]interface{}{
		"permissions": []string{},
		"displayName": profile.DisplayName,
		"email":       profile.Email,
		"id":          fbUser.UID,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating user profile: %v\n", err)
	}

	return fbUserToUserRecord(fbUser, profile), nil
}

// TODO
func (r firebaseRepository) Update(user *UpdateUserRequest) (*User, error) {
	return nil, nil
}

func (r firebaseRepository) Delete(id string) error {
	// Delete account from Firebase Authentication.
	err := r.authClient.DeleteUser(firebase.FirebaseContext, id)
	if err != nil {
		return DeleteUserError
	}

	// Delete profile from user_profiles Firestore collection.
	_, err = r.firestoreClient.Collection("user_profiles").Doc(id).Delete(firebase.FirebaseContext)
	if err != nil {
		return DeleteUserError
	}

	return nil
}

// Helpers

// startUserProfilesListener attaches a listener to the user_profiles collection and updates the
// userProfiles map.
// This allows us to keep an in-memory copy of the user_profiles Firestore collection, so we don't have to query
// Firestore each time we need to access a profile.
func (r firebaseRepository) startUserProfilesListener() error {
	it := r.firestoreClient.Collection(FirestoreUserProfilesCollection).Snapshots(firebase.FirebaseContext)
	for {
		snap, err := it.Next()
		// DeadlineExceeded will be returned when ctx is cancelled.
		if status.Code(err) == codes.DeadlineExceeded {
			return nil
		}
		if err != nil {
			return fmt.Errorf("Snapshots.Next: %v", err)
		}
		if snap != nil {
			for {
				doc, err := snap.Documents.Next()
				if err == iterator.Done {
					break
				}
				if err != nil {
					return fmt.Errorf("Documents.Next: %v", err)
				}

				var userProfile Profile
				err = mapstructure.Decode(doc.Data(), &userProfile)
				if err != nil {
					return err
				}

				r.profilesLock.Lock()
				r.profiles[doc.Ref.ID] = &userProfile
				r.profilesLock.Unlock()
			}
		}
	}
}

// fbUserToUserRecord combines a Firebase UserRecord and a Profile into a User
func fbUserToUserRecord(fbUser *firebaseAuth.UserRecord, profile *Profile) *User {
	// TODO: Refactor such that displayName, email, and profile photo are pulled from firebase auth and not the user profile stored in Firestore.
	return &User{
		ID:                 fbUser.UID,
		Profile:            profile,
		Disabled:           fbUser.Disabled,
		CreationTimestamp:  fbUser.UserMetadata.CreationTimestamp,
		LastLogInTimestamp: fbUser.UserMetadata.LastLogInTimestamp,
	}
}

// getUserProfile gets the Profile from the userProfiles map corresponding to the provided user ID.
func (r firebaseRepository) getUserProfile(id string) (*Profile, error) {
	r.profilesLock.RLock()
	defer r.profilesLock.RUnlock()

	if val, ok := r.profiles[id]; ok {
		return val, nil
	} else {
		return nil, fmt.Errorf("No profile found for ID %v\n", id)
	}
}

// getUserCount returns the number of user profiles.
func (r firebaseRepository) getUserCount() int {
	r.profilesLock.RLock()
	defer r.profilesLock.RUnlock()

	return len(r.profiles)
}
