package repository

import (
	"fmt"
	"log"
	"signmeup/internal/firebase"
	"signmeup/internal/models"
	"signmeup/internal/qerrors"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/mitchellh/mapstructure"
)

func (fr *FirebaseRepository) initializeCoursesListener() {
	handleDocs := func(docs []*firestore.DocumentSnapshot) error {
		newCourses := make(map[string]*models.Course)
		for _, doc := range docs {
			if !doc.Exists() {
				continue
			}

			var c models.Course
			err := mapstructure.Decode(doc.Data(), &c)
			if err != nil {
				log.Panicf("Error destructuring document: %v", err)
				return err
			}

			c.ID = doc.Ref.ID
			newCourses[doc.Ref.ID] = &c
		}

		fr.coursesLock.Lock()
		defer fr.coursesLock.Unlock()
		fr.courses = newCourses

		return nil
	}

	done := make(chan bool)
	go func() {
		err := fr.createCollectionInitializer(models.FirestoreCoursesCollection, &done, handleDocs)
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

func (fr *FirebaseRepository) GetCourseByInfo(code string, term string) (*models.Course, error) {
	fr.coursesLock.RLock()
	defer fr.coursesLock.RUnlock()

	for _, course := range fr.courses {
		if course.Code == code && course.Term == term {
			return course, nil
		}
	}
	return nil, qerrors.CourseNotFoundError
}

func (fr *FirebaseRepository) CreateCourse(c *models.CreateCourseRequest) (course *models.Course, err error) {
	course = &models.Course{
		Title:             c.Title,
		Code:              c.Code,
		Term:              c.Term,
		IsArchived:        false,
		CoursePermissions: map[string]models.CoursePermission{},
	}

	ref, _, err := fr.firestoreClient.Collection(models.FirestoreCoursesCollection).Add(firebase.Context, map[string]interface{}{
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
		_, err = fr.firestoreClient.Collection(models.FirestoreUserProfilesCollection).Doc(k).Update(firebase.Context, []firestore.Update{
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
	_, err = fr.firestoreClient.Collection(models.FirestoreCoursesCollection).Doc(c.CourseID).Delete(firebase.Context)
	return err
}

func (fr *FirebaseRepository) EditCourse(c *models.EditCourseRequest) error {
	_, err := fr.firestoreClient.Collection(models.FirestoreCoursesCollection).Doc(c.CourseID).Update(firebase.Context, []firestore.Update{
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
		// The user doesn't exist; add an invite to the invites collection and then return.
		_, _, err = fr.firestoreClient.Collection(models.FirestoreInvitesCollection).Add(firebase.Context, map[string]interface{}{
			"email":      c.Email,
			"courseID":   c.CourseID,
			"permission": c.Permission,
		})
		return err
	}
	// Set course-side permissions.
	_, err = fr.firestoreClient.Collection(models.FirestoreCoursesCollection).Doc(c.CourseID).Update(firebase.Context, []firestore.Update{
		{
			Path:  "coursePermissions." + user.ID,
			Value: c.Permission,
		},
	})
	if err != nil {
		return err
	}
	// Set user-side permissions.
	_, err = fr.firestoreClient.Collection(models.FirestoreUserProfilesCollection).Doc(user.ID).Update(firebase.Context, []firestore.Update{
		{
			Path:  "coursePermissions." + c.CourseID,
			Value: c.Permission,
		},
	})
	return err
}

func (fr *FirebaseRepository) RemovePermission(c *models.RemoveCoursePermissionRequest) error {
	_, err := fr.firestoreClient.Collection(models.FirestoreCoursesCollection).Doc(c.CourseID).Update(firebase.Context, []firestore.Update{
		{
			Path:  "coursePermissions." + c.UserID,
			Value: firestore.Delete,
		},
	})
	if err != nil {
		return err
	}
	_, err = fr.firestoreClient.Collection(models.FirestoreUserProfilesCollection).Doc(c.UserID).Update(firebase.Context, []firestore.Update{
		{
			Path:  "coursePermissions." + c.CourseID,
			Value: firestore.Delete,
		},
	})
	return err
}

func (fr *FirebaseRepository) BulkUpload(c *models.BulkUploadRequest) error {
	// SCHEMA: (email, [UTA/HTA], course_code, course_name)

	// Extract data.
	data := make([][]string, 0)
	for _, row := range strings.Split(c.Data, "\n") {
		// Split cols, trim fields, lowercase all but course name.
		cols := strings.Split(row, ",")
		for i := range cols {
			cols[i] = strings.TrimSpace(cols[i])
			if i != 3 {
				cols[i] = strings.ToLower(cols[i])
			}
		}
		// Map permissions.
		if cols[1] == "hta" {
			cols[1] = string(models.CourseAdmin)
		} else {
			cols[1] = string(models.CourseStaff)
		}
		data = append(data, cols)
	}

	// Create courses.
	courses := make(map[string]string)
	for _, row := range data {
		courseCode := row[2]
		courseName := row[3]
		courses[courseCode] = courseName
	}
	for code, name := range courses {
		_, _ = fr.CreateCourse(&models.CreateCourseRequest{
			Title: name,
			Code:  code,
			Term:  c.Term,
		})
	}

	// Create invites.
	for _, row := range data {
		course, err := fr.GetCourseByInfo(row[2], c.Term)
		if err != nil {
			return err
		}
		_ = fr.AddPermission(&models.AddCoursePermissionRequest{
			CourseID:   course.ID,
			Email:      row[0],
			Permission: row[1],
		})
	}
	return nil
}
