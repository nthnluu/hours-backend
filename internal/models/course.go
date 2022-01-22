package models

var (
	FirestoreCoursesCollection = "courses"
)

type Course struct {
	ID                string                      `json:"id" mapstructure:"id"`
	Title             string                      `json:"title" mapstructure:"title"`
	Code              string                      `json:"code" mapstructure:"code"`
	Term              string                      `json:"term" mapstructure:"term"`
	IsArchived        bool                        `json:"isArchived" mapstructure:"isArchived"`
	CoursePermissions map[string]CoursePermission `json:"coursePermissions" mapstructure:"coursePermissions"`
}

type GetCourseRequest struct {
	CourseID string `json:"courseID"`
}

type CreateCourseRequest struct {
	Title     string `json:"title"`
	Code      string `json:"code"`
	Term      string `json:"term"`
	CreatedBy *User  `json:"omitempty"`
}

type DeleteCourseRequest struct {
	CourseID string `json:"courseID"`
}

type EditCourseRequest struct {
	CourseID string `json:"courseID"`
	Title    string `json:"title"`
	Code     string `json:"code"`
	Term     string `json:"term"`
}

type AddCoursePermissionRequest struct {
	CourseID   string `json:"courseID"`
	Email      string `json:"email"`
	Permission string `json:"permission"`
}

type RemoveCoursePermissionRequest struct {
	CourseID string `json:"courseID"`
	UserID   string `json:"userID"`
}
