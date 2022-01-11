package course

import (
	"log"
)

var repository Repository

// CreateCourse creates a course using the provided Title, Code, and Term.
func GetCourse(course *GetCourseRequest) (*Course, error) {
	gottedCourse, err := repository.Get(course)
	if err != nil {
		return nil, err
	}

	return gottedCourse, nil
}

// CreateCourse creates a course using the provided Title, Code, and Term.
func CreateCourse(course *CreateCourseRequest) (*Course, error) {
	createdCourse, err := repository.Create(course)
	if err != nil {
		return nil, err
	}

	return createdCourse, nil
}

// GetCourseByID returns a course with the given ID.
func GetCourseByID(id string) (*Course, error) {
	course, err := repository.Get(&GetCourseRequest{id})
	if err != nil {
		return nil, err
	}

	return course, nil
}

func init() {
	repo, err := NewFirebaseRepository()
	if err != nil {
		log.Panicf("error creating Firebase user repository: %v\n", err)
	}

	repository = repo
}
