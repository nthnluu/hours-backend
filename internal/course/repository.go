package course

import (
	"cloud.google.com/go/firestore"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"signmeup/internal/auth"
	"signmeup/internal/firebase"
	"sync"
	"time"
)

// Repository encapsulates the logic to access courses from a database.
type Repository interface {
	// Get returns the course corresponding to the specified course ID.
	Get(course *GetCourseRequest) (*Course, error)
	// Create saves a new course into the database.
	Create(course *CreateCourseRequest) (*Course, error)
	// Delete deletes a new course from a database.
	Delete(course *DeleteCourseRequest) error
	// Edit edits a course's details.
	Edit(c *EditCourseRequest) error
	// AddPermission adds an admin to a course.
	AddPermission(c *AddCoursePermissionRequest) error
	// RemovePermission removes an admin from a course.
	RemovePermission(c *RemoveCoursePermissionRequest) error
}

type firebaseRepository struct {
	firestoreClient *firestore.Client

	coursesLock *sync.RWMutex
	courses     map[string]*Course
}

const (
	FirestoreCoursesCollection = "courses"
)

// NewFirebaseRepository creates a new user repository with Firebase as the database.
func NewFirebaseRepository() (Repository, error) {
	repository := &firebaseRepository{
		coursesLock: &sync.RWMutex{},
		courses:     make(map[string]*Course),
	}

	firestoreClient, err := firebase.FirebaseApp.Firestore(firebase.FirebaseContext)
	if err != nil {
		return nil, fmt.Errorf("Firestore client error: %v\n", err)
	}
	repository.firestoreClient = firestoreClient

	var wg sync.WaitGroup
	wg.Add(1)
	log.Println("⏳ Starting courses collection listener...")
	go func() {
		err := repository.startCoursesListener(&wg)
		if err != nil {
			log.Fatalf("courses collection listner error: %v\n", err)
		}
	}()

	wg.Wait()

	time.Sleep(10)

	return repository, nil
}

func (r *firebaseRepository) Get(c *GetCourseRequest) (*Course, error) {
	course, err := r.getCourse(c.CourseID)
	if err != nil {
		return nil, err
	}

	return course, nil
}

func (r *firebaseRepository) Create(c *CreateCourseRequest) (course *Course, err error) {
	course = &Course{
		Title:             c.Title,
		Code:              c.Code,
		Term:              c.Term,
		IsArchived:        false,
		CoursePermissions: map[string]auth.CoursePermission{},
	}

	course.CoursePermissions[c.CreatedBy.ID] = auth.CourseAdmin
	ref, _, err := r.firestoreClient.Collection(FirestoreCoursesCollection).Add(firebase.FirebaseContext, map[string]interface{}{
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
	_, err = r.firestoreClient.Collection(auth.FirestoreUserProfilesCollection).Doc(c.CreatedBy.ID).Set(firebase.FirebaseContext, map[string]interface{}{
		"coursePermissions": map[string]interface{}{
			course.ID: auth.CourseAdmin,
		},
	}, firestore.MergeAll)
	if err != nil {
		return nil, err
	}
	return course, nil
}

func (r *firebaseRepository) Delete(c *DeleteCourseRequest) error {
	// Get this course's info.
	course, err := r.getCourse(c.CourseID)
	if err != nil {
		return err
	}

	// Delete this course from all users with permissions.
	for k := range course.CoursePermissions {
		_, err = r.firestoreClient.Collection(auth.FirestoreUserProfilesCollection).Doc(k).Update(firebase.FirebaseContext, []firestore.Update{
			{
				Path: "coursePermissions." + course.ID,
				Value: firestore.Delete,
			},
		})
		if err != nil {
			return err
		}
	}

	// Delete the course.
	_, err = r.firestoreClient.Collection(FirestoreCoursesCollection).Doc(c.CourseID).Delete(firebase.FirebaseContext)
	return err
}

func (r *firebaseRepository) Edit(c *EditCourseRequest) error {
	_, err := r.firestoreClient.Collection(FirestoreCoursesCollection).Doc(c.CourseID).Update(firebase.FirebaseContext, []firestore.Update{
		{ Path: "title", Value: c.Title, },
		{ Path: "term", Value: c.Term, },
		{ Path: "code", Value: c.Code, },
	})
	if err != nil {
		return err
	}
	return err
}

func (r *firebaseRepository) AddPermission(c *AddCoursePermissionRequest) error {
	_, err := r.firestoreClient.Collection(FirestoreCoursesCollection).Doc(c.CourseID).Update(firebase.FirebaseContext, []firestore.Update{
		{
			Path: "coursePermissions." + c.UserID,
			Value: c.Permission,
		},
	})
	if err != nil {
		return err
	}
	_, err = r.firestoreClient.Collection(auth.FirestoreUserProfilesCollection).Doc(c.UserID).Update(firebase.FirebaseContext, []firestore.Update{
		{
			Path: "coursePermissions." + c.CourseID,
			Value: c.Permission,
		},
	})
	return err
}

func (r *firebaseRepository) RemovePermission(c *RemoveCoursePermissionRequest) error {
	_, err := r.firestoreClient.Collection(FirestoreCoursesCollection).Doc(c.CourseID).Update(firebase.FirebaseContext, []firestore.Update{
		{
			Path: "coursePermissions." + c.UserID,
			Value: firestore.Delete,
		},
	})
	if err != nil {
		return err
	}
	_, err = r.firestoreClient.Collection(auth.FirestoreUserProfilesCollection).Doc(c.UserID).Update(firebase.FirebaseContext, []firestore.Update{
		{
			Path: "coursePermissions." + c.CourseID,
			Value: firestore.Delete,
		},
	})
	return err
}

// Helpers

// getCourse gets the Course from the courses map corresponding to the provided course ID.
func (r firebaseRepository) getCourse(id string) (*Course, error) {
	r.coursesLock.RLock()
	defer r.coursesLock.RUnlock()

	if val, ok := r.courses[id]; ok {
		return val, nil
	} else {
		return nil, CourseNotFoundError
	}
}

func (r firebaseRepository) startCoursesListener(wg *sync.WaitGroup) error {
	it := r.firestoreClient.Collection(FirestoreCoursesCollection).Snapshots(firebase.FirebaseContext)
	var doOnce sync.Once

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
			r.coursesLock.Lock()

			for {
				doc, err := snap.Documents.Next()
				if err == iterator.Done {
					doOnce.Do(func() {
						log.Println("✅ Started courses collection listener.")
						wg.Done()
					})

					r.coursesLock.Unlock()
					break
				}
				if err != nil {
					return fmt.Errorf("Documents.Next: %v", err)
				}

				var course Course
				err = mapstructure.Decode(doc.Data(), &course)
				if err != nil {
					return err
				}

				course.ID = doc.Ref.ID
				r.courses[doc.Ref.ID] = &course
			}
		}
	}
}
