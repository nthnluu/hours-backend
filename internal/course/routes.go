package course

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"net/http"
	"signmeup/internal/auth"
)

func Routes() *chi.Mux {
	router := chi.NewRouter()
	router.Get("/{courseID}", getCourseHandler)
	router.With(auth.RequireAuth(true)).Post("/", createCourseHandler)
	return router
}

func getCourseHandler(w http.ResponseWriter, r *http.Request) {
	courseID := chi.URLParam(r, "courseID")

	course, err := GetCourse(&GetCourseRequest{courseID})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, course)
}

func createCourseHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateCourseRequest
	user, err := auth.GetUserFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
	}

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req.CreatedBy = user

	course, err := CreateCourse(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, course)
}
