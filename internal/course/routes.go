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
	router.With(auth.RequireAuth(true)).Post("/create", createCourseHandler)
	router.With(auth.RequireAuth(true)).Post("/delete/{courseID}", deleteCourseHandler)
	router.With(auth.RequireAuth(true)).Post("/edit/{courseID}", editCourseHandler)
	router.With(auth.RequireAuth(true)).Post("/addPermission/{courseID}", addCoursePermissionHandler)
	router.With(auth.RequireAuth(true)).Post("/removePermission/{courseID}", removeCoursePermissionHandler)
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

func deleteCourseHandler(w http.ResponseWriter, r *http.Request) {
	courseID := chi.URLParam(r, "courseID")

	err := DeleteCourse(&DeleteCourseRequest{courseID})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("Successfully deleted course " + courseID))
}

func editCourseHandler(w http.ResponseWriter, r *http.Request) {
	var req EditCourseRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req.CourseID = chi.URLParam(r, "courseID")

	err = EditCourse(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("Successfully edited course " + req.CourseID))
}

func addCoursePermissionHandler(w http.ResponseWriter, r *http.Request) {
	var req AddCoursePermissionRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req.CourseID = chi.URLParam(r, "courseID")

	err = AddCoursePermission(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("Successfully added course permission to " + req.CourseID))
}

func removeCoursePermissionHandler(w http.ResponseWriter, r *http.Request) {
	var req RemoveCoursePermissionRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req.CourseID = chi.URLParam(r, "courseID")

	err = RemoveCoursePermission(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("Successfully removed course permission from " + req.CourseID))
}
