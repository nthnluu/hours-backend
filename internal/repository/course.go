package repository

import (
	"fmt"
	"github.com/golang/glog"
	"google.golang.org/api/iterator"
	"signmeup/internal/firebase"
	"signmeup/internal/models"
	"signmeup/internal/qerrors"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/mitchellh/mapstructure"
)

// GetCourseByID gets the Course from the courses map corresponding to the provided course ID.
func (fr *FirebaseRepository) GetCourseByID(ID string) (*models.Course, error) {
	doc, err := fr.firestoreClient.Collection(models.FirestoreCoursesCollection).Doc(ID).Get(firebase.Context)
	if err != nil {
		return nil, err
	}

	if !doc.Exists() {
		return nil, qerrors.CourseNotFoundError
	}

	var c models.Course
	err = mapstructure.Decode(doc.Data(), &c)
	if err != nil {
		glog.Fatalf("Error destructuring course document: %v", err)
		return nil, qerrors.CourseNotFoundError
	}

	c.ID = doc.Ref.ID
	return &c, nil
}

func (fr *FirebaseRepository) GetCourseByInfo(code string, term string) (*models.Course, error) {
	iter := fr.firestoreClient.Collection(models.FirestoreCoursesCollection).Where("code", "==", code).Where("term", "==", term).Limit(1).Documents(firebase.Context)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, qerrors.CourseNotFoundError
		}

		// Return the first result of the query.
		var c models.Course
		err = mapstructure.Decode(doc.Data(), &c)
		if err != nil {
			glog.Fatalf("Error destructuring course document: %v", err)
			return nil, qerrors.CourseNotFoundError
		}

		c.ID = doc.Ref.ID
		return &c, nil
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

// DeleteCoursesByTerm deletes all courses within the given term.
func (fr *FirebaseRepository) DeleteCoursesByTerm(term string) error {
	iter := fr.firestoreClient.Collection(models.FirestoreCoursesCollection).Where("term", "==", term).Documents(firebase.Context)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		_, err = doc.Ref.Delete(firebase.Context)
		if err != nil {
			glog.Fatalf("Error deleting course document: %v", err)
			return err
		}
	}
	return nil
}
