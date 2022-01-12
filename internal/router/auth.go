package router

import (
	"encoding/json"
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

	// Information about the current user
	router.With(auth.RequireAuth(false)).Get("/me", getMeHandler)
	router.With(auth.RequireAuth(false)).Get("/{userID}", getUserHandler)

	// Update the current user's information
	router.With(auth.RequireAuth(true)).Post("/update/{userID}", updateUserHandler)
	router.With(auth.RequireAuth(true)).Post("/updateByEmail", updateUserByEmailHandler)

	// Alter the current session
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

	render.JSON(w, r, user.Profile)
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
			log.Fatalf("json marshal fucked. Err: %s", err)
		}
		w.Write(jsonResp)
		return
	}
	render.JSON(w, r, user)
}

// POST: /update/{userId}
func updateUserHandler(w http.ResponseWriter, r *http.Request) {
	var req *models.UpdateUserRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req.ID = chi.URLParam(r, "userID")

	err = repo.Repository.UpdateUser(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("Successfully edited user " + req.ID))
}

// POST: /updateByEmail
func updateUserByEmailHandler(w http.ResponseWriter, r *http.Request) {
	var req *models.UpdateUserByEmailRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = repo.Repository.UpdateUserByEmail(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("Successfully edited user " + req.Email))
}

// POST: /session
func createSessionHandler(w http.ResponseWriter, r *http.Request) {
	authClient, err := firebase.FirebaseApp.Auth(firebase.FirebaseContext)
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
	cookie, err := authClient.SessionCookie(firebase.FirebaseContext, req.Token, expiresIn)
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
