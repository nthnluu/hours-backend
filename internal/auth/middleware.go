package auth

import (
	"context"
	"net/http"
)

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenCookie, err := r.Cookie("signmeup-session")
		if err != nil {
			// Missing session cookie.
			http.Error(w, "You must be authenticated to access this resource", http.StatusUnauthorized)
			return
		}

		// Verify the session cookie. In this case an additional check is added to detect
		// if the user's Firebase session was revoked, user deleted/disabled, etc.
		user, err := VerifySessionCookie(tokenCookie)
		if err != nil {
			// Missing session cookie.
			http.Error(w, "You must be authenticated to access this resource", http.StatusUnauthorized)
			return
		}

		// create a new request context containing the authenticated user
		ctxWithUser := context.WithValue(r.Context(), "currentUser", user)
		rWithUser := r.WithContext(ctxWithUser)

		next.ServeHTTP(w, rWithUser)
	})
}
