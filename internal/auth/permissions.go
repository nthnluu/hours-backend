package auth

import (
	"net/http"
	"signmeup/internal/models"
	repo "signmeup/internal/repository"

	"github.com/go-chi/chi/v5"
)

func RequireStaffForCourse(courseURLParam string) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := GetUserFromRequest(r)
			if err != nil {
				rejectUnauthorizedRequest(w)
				return
			}

			if !isStaffForCourse(user, chi.URLParam(r, courseURLParam)) {
				rejectForbiddenRequest(w)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RequireAdminForCourse(courseURLParam string) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := GetUserFromRequest(r)
			if err != nil {
				rejectUnauthorizedRequest(w)
				return
			}

			if !isAdminForCourse(user, chi.URLParam(r, courseURLParam)) {
				rejectForbiddenRequest(w)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RequireStaffForQueue(queueURLParam string) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := GetUserFromRequest(r)
			if err != nil {
				rejectUnauthorizedRequest(w)
				return
			}

			qID := chi.URLParam(r, queueURLParam)
			q, err := repo.Repository.GetQueue(qID)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if !isStaffForCourse(user, q.CourseID) {
				rejectForbiddenRequest(w)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func isStaffForCourse(u *models.User, courseID string) bool {
	if _, ok := u.CoursePermissions[courseID]; !ok {
		return false
	}

	return true
}

func isAdminForCourse(u *models.User, courseID string) bool {
	var ok bool
	var p models.CoursePermission

	if p, ok = u.CoursePermissions[courseID]; !ok {
		return false
	}

	return p == models.CourseAdmin
}
