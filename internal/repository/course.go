package repository

import (
	"fmt"
	"log"
	"signmeup/internal/firebase"
	"signmeup/internal/models"
	"signmeup/internal/qerrors"

	"cloud.google.com/go/firestore"
	"github.com/mitchellh/mapstructure"
)

func (fr *FirebaseRepository) initializeCoursesListener() {
	handleDoc := func(doc *firestore.DocumentSnapshot) error {
		fr.coursesLock.Lock()
		defer fr.coursesLock.Unlock()

		var c models.Course
		err := mapstructure.Decode(doc.Data(), &c)
		if err != nil {
			log.Panicf("Error destructuring document: %v", err)
			return err
		}

		c.ID = doc.Ref.ID
		fr.courses[doc.Ref.ID] = &c

		return nil
	}

	done := make(chan bool)
	go func() {
		err := fr.createCollectionInitializer(models.FirestoreCoursesCollection, &done, handleDoc)
		if err != nil {
			log.Panicf("error creating course collection listner: %v\n", err)
		}
	}()
	<-done
}

// GetCourseByID gets the Course from the courses map corresponding to the provided course ID.
func (fr *FirebaseRepository) GetCourseByID(ID string) (*models.Course, error) {
	fr.coursesLock.RLock()
	defer fr.coursesLock.RUnlock()

	if val, ok := fr.courses[ID]; ok {
		return val, nil
	} else {
		return nil, qerrors.CourseNotFoundError
	}
}

func (fr *FirebaseRepository) CreateCourse(c *models.CreateCourseRequest) (course *models.Course, err error) {
	course = &models.Course{
		Title:             c.Title,
		Code:              c.Code,
		Term:              c.Term,
		IsArchived:        false,
		CoursePermissions: map[string]models.CoursePermission{},
	}

	course.CoursePermissions[c.CreatedBy.ID] = models.CourseAdmin
	ref, _, err := fr.firestoreClient.Collection(models.FirestoreCoursesCollection).Add(firebase.FirebaseContext, map[string]interface{}{
		"title":             course.Title,
		"code":              course.Code,
		"term":              course.Term,
		"isArchived":        course.IsArchived,
		"coursePermissions": course.CoursePermissions,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating course: %v\n", err)
	}
	course.ID = ref.ID

	// update user profile to include new permission
	// TODO(n-young): refactor when update user is implmted. :)
	_, err = fr.firestoreClient.Collection(models.FirestoreUserProfilesCollection).Doc(c.CreatedBy.ID).Set(firebase.FirebaseContext, map[string]interface{}{
		"coursePermissions": map[string]interface{}{
			course.ID: models.CourseAdmin,
		},
	}, firestore.MergeAll)
	if err != nil {
		return nil, err
	}
	return course, nil
}

func (fr *FirebaseRepository) DeleteCourse(c *models.DeleteCourseRequest) error {
	// Get this course's info.
	course, err := fr.GetCourseByID(c.CourseID)
	if err != nil {
		return err
	}

	// Delete this course from all users with permissions.
	for k := range course.CoursePermissions {
		_, err = fr.firestoreClient.Collection(models.FirestoreUserProfilesCollection).Doc(k).Update(firebase.FirebaseContext, []firestore.Update{
			{
				Path:  "coursePermissions." + course.ID,
				Value: firestore.Delete,
			},
		})
		if err != nil {
			return err
		}
	}

	// Delete the course.
	_, err = fr.firestoreClient.Collection(models.FirestoreCoursesCollection).Doc(c.CourseID).Delete(firebase.FirebaseContext)
	return err
}

func (fr *FirebaseRepository) EditCourse(c *models.EditCourseRequest) error {
	_, err := fr.firestoreClient.Collection(models.FirestoreCoursesCollection).Doc(c.CourseID).Update(firebase.FirebaseContext, []firestore.Update{
		{Path: "title", Value: c.Title},
		{Path: "term", Value: c.Term},
		{Path: "code", Value: c.Code},
	})
	return err
}

func (fr *FirebaseRepository) AddPermission(c *models.AddCoursePermissionRequest) error {
	// Get user by email.
	user, err := fr.GetUserByEmail(c.Email)
	if err != nil {
		return err
	}
	// Set course-side permissions.
	_, err = fr.firestoreClient.Collection(models.FirestoreCoursesCollection).Doc(c.CourseID).Update(firebase.FirebaseContext, []firestore.Update{
		{
			Path:  "coursePermissions." + user.ID,
			Value: c.Permission,
		},
	})
	if err != nil {
		return err
	}
	// Set user-side permissions.
	_, err = fr.firestoreClient.Collection(models.FirestoreUserProfilesCollection).Doc(user.ID).Update(firebase.FirebaseContext, []firestore.Update{
		{
			Path:  "coursePermissions." + c.CourseID,
			Value: c.Permission,
		},
	})
	return err
}

func (fr *FirebaseRepository) RemovePermission(c *models.RemoveCoursePermissionRequest) error {
	_, err := fr.firestoreClient.Collection(models.FirestoreCoursesCollection).Doc(c.CourseID).Update(firebase.FirebaseContext, []firestore.Update{
		{
			Path:  "coursePermissions." + c.UserID,
			Value: firestore.Delete,
		},
	})
	if err != nil {
		return err
	}
	_, err = fr.firestoreClient.Collection(models.FirestoreUserProfilesCollection).Doc(c.UserID).Update(firebase.FirebaseContext, []firestore.Update{
		{
			Path:  "coursePermissions." + c.CourseID,
			Value: firestore.Delete,
		},
	})
	return err
}
