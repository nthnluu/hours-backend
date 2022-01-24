package auth

import (
	"net/http"
	"signmeup/internal/models"
	repo "signmeup/internal/repository"
)

func RequireStaffForCourse() func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := GetUserFromRequest(r)
			if err != nil {
				rejectUnauthorizedRequest(w)
				return
			}

			courseID := r.Context().Value("courseID").(string)
			if !hasCourseStaffPermission(user, courseID) {
				rejectForbiddenRequest(w)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RequireCourseAdmin() func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := GetUserFromRequest(r)
			if err != nil {
				rejectUnauthorizedRequest(w)
				return
			}

			courseID := r.Context().Value("courseID").(string)
			if !hasCourseAdminPermission(user, courseID) {
				rejectForbiddenRequest(w)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RequireQueueStaff() func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := GetUserFromRequest(r)
			if err != nil {
				rejectUnauthorizedRequest(w)
				return
			}

			qID := r.Context().Value("queueID").(string)
			q, err := repo.Repository.GetQueue(qID)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if !hasCourseStaffPermission(user, q.CourseID) {
				rejectForbiddenRequest(w)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func hasCourseStaffPermission(u *models.User, courseID string) bool {
	if u.IsAdmin {
		return true
	}

	if _, ok := u.CoursePermissions[courseID]; !ok {
		return false
	}

	return true
}

func hasCourseAdminPermission(u *models.User, courseID string) bool {
	if u.IsAdmin {
		return true
	}

	var ok bool
	var p models.CoursePermission
	if p, ok = u.CoursePermissions[courseID]; !ok {
		return false
	}
	return p == models.CourseAdmin
}
