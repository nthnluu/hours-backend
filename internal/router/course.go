package router

import (
	"encoding/json"
	"net/http"
	"signmeup/internal/auth"
	"signmeup/internal/models"
	repo "signmeup/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func CourseRoutes() *chi.Mux {
	router := chi.NewRouter()

	// Get metadata about a course
	router.Get("/{courseID}", getCourseHandler)

	// Modifying courses themselves
	router.With(auth.RequireAuth(true)).Post("/create", createCourseHandler)
	router.With(auth.RequireAuth(true)).Post("/delete/{courseID}", deleteCourseHandler)
	router.With(auth.RequireAuth(true)).Post("/edit/{courseID}", editCourseHandler)

	// Course permissions
	router.With(auth.RequireAuth(true)).Post("/addPermission/{courseID}", addCoursePermissionHandler)
	router.With(auth.RequireAuth(true)).Post("/removePermission/{courseID}", removeCoursePermissionHandler)
	return router
}

// GET: /{courseID}
func getCourseHandler(w http.ResponseWriter, r *http.Request) {
	courseID := chi.URLParam(r, "courseID")

	course, err := repo.Repository.GetCourse(&models.GetCourseRequest{CourseID: courseID})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, course)
}

// POST: /create
func createCourseHandler(w http.ResponseWriter, r *http.Request) {
	var req *models.CreateCourseRequest

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

	c, err := repo.Repository.CreateCourse(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, c)
}

// POST: /delete/{courseID}
func deleteCourseHandler(w http.ResponseWriter, r *http.Request) {
	courseID := chi.URLParam(r, "courseID")

	err := repo.Repository.DeleteCourse(&models.DeleteCourseRequest{courseID})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("Successfully deleted course " + courseID))
}

// POST: /edit/{courseID}
func editCourseHandler(w http.ResponseWriter, r *http.Request) {
	var req *models.EditCourseRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req.CourseID = chi.URLParam(r, "courseID")

	err = repo.Repository.EditCourse(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("Successfully edited course " + req.CourseID))
}

// POST: /addPermission/{courseID}
func addCoursePermissionHandler(w http.ResponseWriter, r *http.Request) {
	var req *models.AddCoursePermissionRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req.CourseID = chi.URLParam(r, "courseID")

	err = repo.Repository.AddPermission(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("Successfully added course permission to " + req.CourseID))
}

// POST: /removePermission/{courseID}
func removeCoursePermissionHandler(w http.ResponseWriter, r *http.Request) {
	var req *models.RemoveCoursePermissionRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req.CourseID = chi.URLParam(r, "courseID")

	err = repo.Repository.RemovePermission(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("Successfully removed course permission from " + req.CourseID))
}
