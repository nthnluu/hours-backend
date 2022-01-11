package course

import (
	"log"
)

var repository Repository

// GetCourse creates a course using the provided ID.
func GetCourse(course *GetCourseRequest) (*Course, error) {
	gottedCourse, err := repository.Get(course)
	if err != nil {
		return nil, err
	}

	return gottedCourse, nil
}

// GetCourseByID returns a course with the given ID.
func GetCourseByID(id string) (*Course, error) {
	course, err := repository.Get(&GetCourseRequest{id})
	if err != nil {
		return nil, err
	}

	return course, nil
}

// CreateCourse creates a course using the provided Title, Code, and Term.
func CreateCourse(course *CreateCourseRequest) (*Course, error) {
	createdCourse, err := repository.Create(course)
	if err != nil {
		return nil, err
	}

	return createdCourse, nil
}

// GetCourse creates a course using the provided ID.
func DeleteCourse(course *DeleteCourseRequest) error {
	return repository.Delete(course)
}

// EditCourse edits an existing course's details.
func EditCourse(course *EditCourseRequest) error {
	return repository.Edit(course)
}

// AddCoursePermission adds a course admin.
func AddCoursePermission(coursePemission *AddCoursePermissionRequest) error {
	return repository.AddPermission(coursePemission)
}

// RemoveCoursePermission removes a course admin.
func RemoveCoursePermission(coursePemission *RemoveCoursePermissionRequest) error {
	return repository.RemovePermission(coursePemission)
}

func init() {
	repo, err := NewFirebaseRepository()
	if err != nil {
		log.Panicf("error creating Firebase user repository: %v\n", err)
	}

	repository = repo
}
