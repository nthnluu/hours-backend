package middleware

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func QueueCtx() func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			queueID := chi.URLParam(r, "queueID")

			ctx := context.WithValue(r.Context(), "queueID", queueID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func CourseCtx() func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			queueID := chi.URLParam(r, "courseID")

			ctx := context.WithValue(r.Context(), "courseID", queueID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
