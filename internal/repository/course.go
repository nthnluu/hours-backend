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
		err := mapstructure.Decode(doc, &c)
		if err != nil {
			log.Panicf("Error destructuring document: %v", err)
			return err
		}

		c.ID = doc.Ref.ID
		fr.courses[doc.Ref.ID] = &c

		return nil
	}

	done := make(chan bool)
	go fr.createCollectionInitializer(models.FirestoreCoursesCollection, &done, handleDoc)
	<-done
}

func (r *FirebaseRepository) CreateCourse(ccr *models.CreateCourseRequest) (*models.Course, error) {
	c := &models.Course{
		Title:      ccr.Title,
		Code:       ccr.Code,
		Term:       ccr.Term,
		IsArchived: false,
	}
	ref, _, err := r.firestoreClient.Collection(models.FirestoreCoursesCollection).Add(firebase.FirebaseContext, map[string]interface{}{
		"title":      c.Title,
		"code":       c.Code,
		"term":       c.Term,
		"isArchived": c.IsArchived,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating course: %v\n", err)
	}

	c.ID = ref.ID
	return c, nil
}

func (r *FirebaseRepository) Get(id string) (*models.Course, error) {
	course, err := r.getCourse(id)
	if err != nil {
		return nil, err
	}

	return course, nil
}

// Helpers

// getCourse gets the Course from the courses map corresponding to the provided course ID.
func (r FirebaseRepository) getCourse(ID string) (*models.Course, error) {
	r.coursesLock.RLock()
	defer r.coursesLock.RUnlock()

	if val, ok := r.courses[ID]; ok {
		return val, nil
	} else {
		return nil, qerrors.CourseNotFoundError
	}
}
