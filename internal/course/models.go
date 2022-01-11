package course

type Course struct {
	ID         string `json:"id" mapstructure:"id"`
	Title      string `json:"title" mapstructure:"title"`
	Code       string `json:"code" mapstructure:"code"`
	Term       string `json:"term" mapstructure:"term"`
	IsArchived bool   `json:"isArchived" mapstructure:"isArchived"`
}

type CreateCourseRequest struct {
	Title string `json:"title"`
	Code  string `json:"code"`
	Term  string `json:"term"`
}
