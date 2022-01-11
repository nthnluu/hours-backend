package course

import (
	"cloud.google.com/go/firestore"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"signmeup/internal/firebase"
	"sync"
	"time"
)

// Repository encapsulates the logic to access courses from a database.
type Repository interface {
	// Get returns the course corresponding to the specified course ID.
	Get(id string) (*Course, error)
	// Create saves a new course into the database.
	Create(course *CreateCourseRequest) (*Course, error)
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

func (r *firebaseRepository) Create(c *CreateCourseRequest) (course *Course, err error) {
	course = &Course{
		Title:      c.Title,
		Code:       c.Code,
		Term:       c.Term,
		IsArchived: false,
	}
	ref, _, err := r.firestoreClient.Collection(FirestoreCoursesCollection).Add(firebase.FirebaseContext, map[string]interface{}{
		"title":      course.Title,
		"code":       course.Code,
		"term":       course.Term,
		"isArchived": course.IsArchived,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating course: %v\n", err)
	}

	course.ID = ref.ID
	return
}

func (r *firebaseRepository) Get(id string) (*Course, error) {
	course, err := r.getCourse(id)
	if err != nil {
		return nil, err
	}

	return course, nil
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
