package router

import (
	"encoding/json"
	"github.com/golang/glog"
	"log"
	"net/http"
	"signmeup/internal/auth"
	"signmeup/internal/config"
	"signmeup/internal/firebase"
	"signmeup/internal/models"
	repo "signmeup/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func AuthRoutes() *chi.Mux {
	router := chi.NewRouter()

	// Auth routes that require authentication
	router.Route("/", func(r chi.Router) {
		r.Use(auth.AuthCtx())

		// Information about the current user
		r.Get("/me", getMeHandler)
		r.Get("/{userID}", getUserHandler)

		// Update the current user's information
		r.Post("/update", updateUserHandler)
		r.With(auth.RequireAdmin()).Post("/updateByEmail", updateUserByEmailHandler)

		// Notification clearing
		r.Post("/clearNotification", clearNotificationHandler)
		r.Post("/clearAllNotifications", clearAllNotificationsHandler)

		// Favorite courses
		r.Post("/addFavoriteCourses", addFavoriteCourseHandler)
		r.Post("/removeFavoriteCourses", removeFavoriteCourseHandler)
	})

	// Alter the current session. No auth middlewares required.
	router.Post("/session", createSessionHandler)
	router.Post("/signout", signOutHandler)

	return router
}

// GET: /me
func getMeHandler(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetUserFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	render.JSON(w, r, struct {
		*models.Profile
		ID string `json:"id"`
	}{user.Profile, user.ID})
}

func getUserHandler(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	user, err := repo.Repository.GetUserByID(userID)
	if err != nil {
		// TODO(nthnluu): Refactor into helper function
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		resp := make(map[string]string)
		resp["message"] = "User Not Found"
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			glog.Warningln("failed to marshall json error response")
		}
		_, err = w.Write(jsonResp)
		if err != nil {
			glog.Warningf("failed to write response: %v\n", err)
			log.Println(err)
		}
		return
	}
	render.JSON(w, r, user)
}

// POST: /update
func updateUserHandler(w http.ResponseWriter, r *http.Request) {
	var req *models.UpdateUserRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	user, err := auth.GetUserFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	req.UserID = user.ID

	err = repo.Repository.UpdateUser(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	_, err = w.Write([]byte("successfully edited user " + req.UserID))
	if err != nil {
		glog.Warningf("failed to write response: %v\n", err)
		return
	}
}

// POST: /updateByEmail
func updateUserByEmailHandler(w http.ResponseWriter, r *http.Request) {
	var req *models.MakeAdminByEmailRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = repo.Repository.MakeAdminByEmail(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	_, err = w.Write([]byte("Successfully edited user " + req.Email))
	if err != nil {
		return
	}
}

// POST: /session
func createSessionHandler(w http.ResponseWriter, r *http.Request) {
	authClient, err := firebase.App.Auth(firebase.Context)
	if err != nil {
		log.Fatalf("error getting Auth client: %v\n", err)
	}

	var req struct {
		Token string `json:"token"`
	}

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set session expiration to 5 days.
	expiresIn := config.Config.SessionCookieExpiration

	// Create the session cookie. This will also verify the ID token in the process.
	// The session cookie will have the same claims as the ID token.
	// To only allow session cookie setting on recent sign-in, auth_time in ID token
	// can be checked to ensure user was recently signed in before creating a session cookie.
	cookie, err := authClient.SessionCookie(firebase.Context, req.Token, expiresIn)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var sameSite http.SameSite
	if config.Config.IsHTTPS {
		sameSite = http.SameSiteNoneMode
	} else {
		sameSite = http.SameSiteLaxMode
	}

	http.SetCookie(w, &http.Cookie{
		Name:     config.Config.SessionCookieName,
		Value:    cookie,
		MaxAge:   int(expiresIn.Seconds()),
		HttpOnly: true,
		SameSite: sameSite,
		Secure:   config.Config.IsHTTPS,
		Path:     "/",
	})

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("success"))
	return
}

// POST: /signout
func signOutHandler(w http.ResponseWriter, r *http.Request) {
	var sameSite http.SameSite
	if config.Config.IsHTTPS {
		sameSite = http.SameSiteNoneMode
	} else {
		sameSite = http.SameSiteLaxMode
	}

	http.SetCookie(w, &http.Cookie{
		Name:     config.Config.SessionCookieName,
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: sameSite,
		Secure:   config.Config.IsHTTPS,
		Path:     "/",
	})

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("success"))
	return
}

// POST: notification clear
func clearNotificationHandler(w http.ResponseWriter, r *http.Request) {
	var req *models.ClearNotificationRequest

	user, err := auth.GetUserFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
	}

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	req.UserID = user.ID

	err = repo.Repository.ClearNotification(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("Successfully cleared notification"))
}

// POST: notification clear all
func clearAllNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetUserFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
	}

	req := models.ClearAllNotificationsRequest{UserID: user.ID}

	err = repo.Repository.ClearAllNotifications(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("Successfully cleared notification"))
}

// POST: add a favorite course
func addFavoriteCourseHandler(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetUserFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
	}

	var req *models.AddFavoriteCourseRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		glog.Warningln(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = repo.Repository.AddFavoriteCourse(user.ID, req.CourseID)
	if err != nil {
		glog.Warningln(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("Successfully added favorite course"))
}

// POST: remove a favorite course
func removeFavoriteCourseHandler(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetUserFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
	}

	var req *models.RemoveFavoriteCourseRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		glog.Warningln(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = repo.Repository.RemoveFavoriteCourse(user.ID, req.CourseID)
	if err != nil {
		glog.Warningln(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("Successfully removed favorite course"))
}
