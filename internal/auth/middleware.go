package auth

import (
	"context"
	"net/http"
	"signmeup/internal/config"
)

// RequireAuth is a middleware that rejects requests without a valid session cookie. The User associated with the
// request is added to the request context, and can be accessed via GetUserFromRequest.
func RequireAuth(adminOnly bool) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenCookie, err := r.Cookie(config.Config.SessionCookieName)
			if err != nil {
				// Missing session cookie.
				rejectUnauthorizedRequest(w)
				return
			}

			// Verify the session cookie. In this case an additional check is added to detect
			// if the user's Firebase session was revoked, user deleted/disabled, etc.
			user, err := verifySessionCookie(tokenCookie)
			if err != nil {
				// Missing session cookie.
				rejectUnauthorizedRequest(w)
				return
			}

			if adminOnly && !user.IsAdmin {
				rejectUnauthorizedRequest(w)
				return
			}

			// create a new request context containing the authenticated user
			ctxWithUser := context.WithValue(r.Context(), "currentUser", user)
			rWithUser := r.WithContext(ctxWithUser)

			next.ServeHTTP(w, rWithUser)
		})
	}
}

// GetUserFromRequest returns a User if it exists within the request context. Only works with routes that implement the
// RequireAuth middleware.
func GetUserFromRequest(r *http.Request) (*User, error) {
	user := r.Context().Value("currentUser").(*User)
	if user != nil {
		return user, nil
	}

	return nil, UserNotFoundError
}

// Helpers

func rejectUnauthorizedRequest(w http.ResponseWriter) {
	http.Error(w, "You must be authenticated to access this resource", http.StatusUnauthorized)
}
