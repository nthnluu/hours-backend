package repository

import (
	"cloud.google.com/go/firestore"
	"fmt"
	"github.com/golang/glog"
	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"google.golang.org/api/iterator"
	"log"
	"net/http"
	"signmeup/internal/config"
	"signmeup/internal/firebase"
	"signmeup/internal/models"
	"signmeup/internal/qerrors"
	"strings"

	firebaseAuth "firebase.google.com/go/auth"
)

func (fr *FirebaseRepository) initializeUserProfilesListener() {
	handleDocs := func(docs []*firestore.DocumentSnapshot) error {
		newProfiles := make(map[string]*models.Profile)
		for _, doc := range docs {
			if !doc.Exists() {
				continue
			}

			var c models.Profile
			err := mapstructure.Decode(doc.Data(), &c)
			if err != nil {
				log.Panicf("Error destructuring document: %v", err)
				return err
			}

			newProfiles[doc.Ref.ID] = &c
		}

		fr.profilesLock.Lock()
		defer fr.profilesLock.Unlock()
		fr.profiles = newProfiles

		return nil
	}

	done := make(chan bool)
	query := fr.firestoreClient.Collection(models.FirestoreUserProfilesCollection).Query
	go func() {
		err := fr.createCollectionInitializer(query, &done, handleDocs)
		if err != nil {
			log.Panicf("error creating user profiles collection listener: %v\n", err)
		}
	}()
	<-done
}

// VerifySessionCookie verifies that the given session cookie is valid and returns the associated User if valid.
func (fr *FirebaseRepository) VerifySessionCookie(sessionCookie *http.Cookie) (*models.User, error) {
	decoded, err := fr.authClient.VerifySessionCookieAndCheckRevoked(firebase.Context, sessionCookie.Value)

	if err != nil {
		return nil, fmt.Errorf("error verifying cookie: %v\n", err)
	}

	user, err := fr.GetUserByID(decoded.UID)
	if err != nil {
		return nil, fmt.Errorf("error getting user from cookie: %v\n", err)
	}

	return user, nil
}

func (fr *FirebaseRepository) GetUserByID(id string) (*models.User, error) {
	if err := validateID(id); err != nil {
		return nil, err
	}

	fbUser, err := fr.authClient.GetUser(firebase.Context, id)
	if err != nil {
		return nil, qerrors.UserNotFoundError
	}

	// TODO: Refactor email verification and user profile creation into separate function.

	// Check the Firebase user's email against the list of allowed domains.
	if len(config.Config.AllowedEmailDomains) > 0 {
		domain := strings.Split(fbUser.Email, "@")[1]
		if !contains(config.Config.AllowedEmailDomains, domain) {
			// invalid email domain, delete the user from Firebase Auth
			_ = fr.authClient.DeleteUser(firebase.Context, fbUser.UID)
			return nil, qerrors.InvalidEmailError
		}
	}

	profile, err := fr.getUserProfile(fbUser.UID)
	if err != nil {
		// no profile for the user found, create one.
		profile = &models.Profile{
			DisplayName: fbUser.DisplayName,
			Email:       fbUser.Email,
			PhotoURL:    fbUser.PhotoURL,
			// if there are no registered users, make the first one an admin
			IsAdmin: fr.getUserCount() == 0,
		}
		_, err = fr.firestoreClient.Collection(models.FirestoreUserProfilesCollection).Doc(fbUser.UID).Set(firebase.Context, map[string]interface{}{
			"coursePermissions": make(map[string]models.CoursePermission),
			"displayName":       profile.DisplayName,
			"email":             profile.Email,
			"photoUrl":          profile.PhotoURL,
			"meetingLink":       "",
			"pronouns":          "",
			"id":                fbUser.UID,
			"isAdmin":           profile.IsAdmin,
			"notifications":     make([]models.Notification, 0),
		})
		if err != nil {
			return nil, fmt.Errorf("error creating user profile: %v\n", err)
		}

		// Go through each of the invites and execute them.
		iter := fr.firestoreClient.Collection(models.FirestoreInvitesCollection).Where("email", "==", fbUser.Email).Documents(firebase.Context)
		for {
			// Get this document.
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return nil, err
			}
			// Decode this document.
			var invite models.CourseInvite
			err = mapstructure.Decode(doc.Data(), &invite)
			if err != nil {
				return nil, err
			}
			// Execute the invite.
			err = fr.AddPermission(&models.AddCoursePermissionRequest{
				CourseID:   invite.CourseID,
				Email:      invite.Email,
				Permission: invite.Permission,
			})
			if err != nil {
				glog.Warningf("there was a problem adding course permission to a user: %v\n", err)
			}

			// Delete the doc.
			_, err = doc.Ref.Delete(firebase.Context)
			if err != nil {
				glog.Warningf("there was a problem deleting invite: %v\n", err)
			}
		}
	}

	return fbUserToUserRecord(fbUser, profile), nil
}

// GetUserByEmail retrieves the User associated with the given email.
func (fr *FirebaseRepository) GetUserByEmail(email string) (*models.User, error) {
	userID, err := fr.GetIDByEmail(email)
	if err != nil {
		return nil, err
	}

	return fr.GetUserByID(userID)
}

func (fr *FirebaseRepository) GetIDByEmail(email string) (string, error) {
	// Get user by email.
	iter := fr.firestoreClient.Collection(models.FirestoreUserProfilesCollection).Where("email", "==", email).Documents(firebase.Context)
	doc, err := iter.Next()
	if err != nil {
		return "", err
	}
	// Cast.
	data := doc.Data()
	return data["id"].(string), nil
}

func (fr *FirebaseRepository) UpdateUser(r *models.UpdateUserRequest) error {
	if r.DisplayName == "" {
		return qerrors.InvalidDisplayName
	}

	_, err := fr.firestoreClient.Collection(models.FirestoreUserProfilesCollection).Doc(r.UserID).Update(firebase.Context, []firestore.Update{
		{
			Path:  "displayName",
			Value: r.DisplayName,
		},
		{
			Path:  "pronouns",
			Value: r.Pronouns,
		},
		{
			Path:  "meetingLink",
			Value: r.MeetingLink,
		},
	})

	return err
}

// MakeAdminByEmail makes the user with the given email a site admin.
func (fr *FirebaseRepository) MakeAdminByEmail(u *models.MakeAdminByEmailRequest) error {
	user, err := fr.GetUserByEmail(u.Email)
	if err != nil {
		return err
	}

	_, err = fr.firestoreClient.Collection(models.FirestoreUserProfilesCollection).Doc(user.ID).Update(firebase.Context, []firestore.Update{
		{
			Path:  "isAdmin",
			Value: u.IsAdmin,
		},
	})

	return err
}

func (fr *FirebaseRepository) Count() int {
	fr.profilesLock.RLock()
	defer fr.profilesLock.RUnlock()
	return len(fr.profiles)
}

func (fr *FirebaseRepository) List() ([]*models.User, error) {
	var users []*models.User
	iter := fr.authClient.Users(firebase.Context, "")
	for {
		fbUser, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error listing user_mgt: %s\n", err)
		}

		profile, err := fr.getUserProfile(fbUser.UID)
		if err != nil {
			return nil, err
		}
		user := fbUserToUserRecord(fbUser.UserRecord, profile)

		users = append(users, user)
	}

	return users, nil
}

// Operations

func (fr *FirebaseRepository) AddNotification(userID string, notification models.Notification) error {
	notification.ID = uuid.New().String()
	_, err := fr.firestoreClient.Collection(models.FirestoreUserProfilesCollection).Doc(userID).Update(firebase.Context, []firestore.Update{
		{
			Path:  "notifications",
			Value: firestore.ArrayUnion(notification),
		},
	})

	return err
}

func (fr *FirebaseRepository) ClearNotification(c *models.ClearNotificationRequest) error {
	user, err := fr.GetUserByID(c.UserID)
	if err != nil {
		return err
	}

	newNotifications := make([]models.Notification, 0)
	for _, v := range user.Notifications {
		if v.ID != c.NotificationID {
			newNotifications = append(newNotifications, v)
		}
	}
	_, err = fr.firestoreClient.Collection(models.FirestoreUserProfilesCollection).Doc(c.UserID).Update(firebase.Context, []firestore.Update{
		{
			Path:  "notifications",
			Value: newNotifications,
		},
	})
	return err
}

func (fr *FirebaseRepository) ClearAllNotifications(c *models.ClearAllNotificationsRequest) error {
	_, err := fr.firestoreClient.Collection(models.FirestoreUserProfilesCollection).Doc(c.UserID).Update(firebase.Context, []firestore.Update{
		{
			Path:  "notifications",
			Value: make([]models.Notification, 0),
		},
	})
	return err
}

// Validate checks a CreateUserRequest struct for errors.
func validate(u *models.CreateUserRequest) error {
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

func (fr *FirebaseRepository) Create(user *models.CreateUserRequest) (*models.User, error) {
	if err := validate(user); err != nil {
		return nil, err
	}

	// Create a user in Firebase Auth.
	u := (&firebaseAuth.UserToCreate{}).Email(user.Email).Password(user.Password)
	fbUser, err := fr.authClient.CreateUser(firebase.Context, u)
	if err != nil {
		return nil, fmt.Errorf("error creating user: %v\n", err)
	}

	// Create a user profile in Firestore.
	profile := &models.Profile{
		DisplayName: user.DisplayName,
		Email:       user.Email,
	}
	_, err = fr.firestoreClient.Collection(models.FirestoreUserProfilesCollection).Doc(fbUser.UID).Set(firebase.Context, map[string]interface{}{
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

func (fr *FirebaseRepository) Delete(id string) error {
	// Delete account from Firebase Authentication.
	err := fr.authClient.DeleteUser(firebase.Context, id)
	if err != nil {
		return qerrors.DeleteUserError
	}

	// Delete profile from user_profiles Firestore collection.
	_, err = fr.firestoreClient.Collection("user_profiles").Doc(id).Delete(firebase.Context)
	if err != nil {
		return qerrors.DeleteUserError
	}

	return nil
}

func (fr *FirebaseRepository) AddFavoriteCourse(userID string, courseID string) error {
	_, err := fr.firestoreClient.Collection(models.FirestoreUserProfilesCollection).Doc(userID).Update(firebase.Context, []firestore.Update{
		{
			Path:  "favoriteCourses",
			Value: firestore.ArrayUnion(courseID),
		},
	})

	return err
}

func (fr *FirebaseRepository) RemoveFavoriteCourse(userID string, courseID string) error {
	_, err := fr.firestoreClient.Collection(models.FirestoreUserProfilesCollection).Doc(userID).Update(firebase.Context, []firestore.Update{
		{
			Path:  "favoriteCourses",
			Value: firestore.ArrayRemove(courseID),
		},
	})

	return err
}

// Helpers

// fbUserToUserRecord combines a Firebase UserRecord and a Profile into a User
func fbUserToUserRecord(fbUser *firebaseAuth.UserRecord, profile *models.Profile) *models.User {
	// TODO: Refactor such that displayName, email, and profile photo are pulled from firebase auth and not the user profile stored in Firestore.
	return &models.User{
		ID:                 fbUser.UID,
		Profile:            profile,
		Disabled:           fbUser.Disabled,
		CreationTimestamp:  fbUser.UserMetadata.CreationTimestamp,
		LastLogInTimestamp: fbUser.UserMetadata.LastLogInTimestamp,
	}
}

// getUserProfile gets the Profile from the userProfiles map corresponding to the provided user ID.
func (fr *FirebaseRepository) getUserProfile(id string) (*models.Profile, error) {
	fr.profilesLock.RLock()
	defer fr.profilesLock.RUnlock()

	if val, ok := fr.profiles[id]; ok {
		return val, nil
	} else {
		return nil, fmt.Errorf("No profile found for ID %v\n", id)
	}
}

// getUserCount returns the number of user profiles.
func (fr *FirebaseRepository) getUserCount() int {
	fr.profilesLock.RLock()
	defer fr.profilesLock.RUnlock()
	return len(fr.profiles)
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

// TODO: Maybe find a better place for this?

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
