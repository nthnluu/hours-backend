package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"signmeup/internal/auth"
	"signmeup/internal/middleware"
	"signmeup/internal/models"
	"signmeup/internal/qerrors"
	repo "signmeup/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func CourseRoutes() *chi.Mux {
	router := chi.NewRouter()
	// All course routes require authentication.
	router.Use(auth.AuthCtx())

	// Modifying courses themselves
	router.With(auth.RequireAdmin()).Post("/create", createCourseHandler)

	// Get metadata about a course
	router.Route("/{courseID}", func(router chi.Router) {
		router.Use(middleware.CourseCtx())

		// Anybody authed can read a course
		router.Get("/", getCourseHandler)

		// Only Admins can delete a course
		router.With(auth.RequireAdmin()).Delete("/", deleteCourseHandler)

		// Course modification
		router.With(auth.RequireCourseAdmin()).Post("/edit", editCourseHandler)
		router.With(auth.RequireCourseAdmin()).Post("/addPermission", addCoursePermissionHandler)
		router.With(auth.RequireCourseAdmin()).Post("/removePermission", removeCoursePermissionHandler)
	})
	router.With(auth.RequireAdmin()).Post("/bulkUpload", bulkUploadHandler)

	return router
}

// GET: /{courseID}
func getCourseHandler(w http.ResponseWriter, r *http.Request) {
	courseID := r.Context().Value("courseID").(string)

	course, err := repo.Repository.GetCourseByID(courseID)
	if err != nil {
		if err == qerrors.CourseNotFoundError {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
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

// DELETE: /{courseID}
func deleteCourseHandler(w http.ResponseWriter, r *http.Request) {
	courseID := r.Context().Value("courseID").(string)

	err := repo.Repository.DeleteCourse(&models.DeleteCourseRequest{CourseID: courseID})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("Successfully deleted course " + courseID))
}

// POST: /{courseID}/edit
func editCourseHandler(w http.ResponseWriter, r *http.Request) {
	var req *models.EditCourseRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req.CourseID = r.Context().Value("courseID").(string)

	err = repo.Repository.EditCourse(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("Successfully edited course " + req.CourseID))
}

// POST: /{courseID}/addPermission
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

// POST: /{courseID}/removePermission
func removeCoursePermissionHandler(w http.ResponseWriter, r *http.Request) {
	var req *models.RemoveCoursePermissionRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req.CourseID = r.Context().Value("courseID").(string)

	err = repo.Repository.RemovePermission(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("Successfully removed course permission from " + req.CourseID))
}

// POST: /bulkUpload
func bulkUploadHandler(w http.ResponseWriter, r *http.Request) {
	var req *models.BulkUploadRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = repo.Repository.BulkUpload(req)
	fmt.Println(err)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("Successfully bulk-uploaded"))
}
