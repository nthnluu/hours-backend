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
	router.With(auth.RequireAuth(true)).Post("/", createCourseHandler)
	return router
}

func createCourseHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateCourseRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	course, err := CreateCourse(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, course)
}
