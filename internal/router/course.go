package router

import (
	"encoding/json"
	"net/http"
	"signmeup/internal/auth"
	"signmeup/internal/models"
	"signmeup/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func CourseRoutes() *chi.Mux {
	router := chi.NewRouter()
	router.With(auth.RequireAuth(true)).Post("/", createCourseHandler)
	return router
}

func createCourseHandler(w http.ResponseWriter, r *http.Request) {
	var req models.CreateCourseRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c, err := repository.Repository.CreateCourse(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, c)
}
