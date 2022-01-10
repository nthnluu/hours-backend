package auth

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"log"
	"net/http"
	"queue/internal/firebase"
	"time"
)

func Routes() *chi.Mux {
	router := chi.NewRouter()
	router.Get("/", getAllUsersHandler)
	router.Post("/", createUserHandler)
	router.Get("/{userID}", getUserHandler)
	router.Delete("/{userID}", deleteUserHandler)

	router.With(RequireAuth).Get("/me", getCurrentUserHandler)
	router.Post("/session", createSessionHandler)
	router.Post("/signout", signOutHandler)
	return router
}

func getUserHandler(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	user, err := GetUserByID(userID)
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

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	user := User{
		Profile:            nil,
		ID:                 "",
		Disabled:           false,
		CreationTimestamp:  0,
		LastLogInTimestamp: 0,
	}
	render.JSON(w, r, user)
}

// TODO
func deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	user := User{
		Profile:            nil,
		ID:                 "",
		Disabled:           false,
		CreationTimestamp:  0,
		LastLogInTimestamp: 0,
	}
	render.JSON(w, r, user)
}

// TODO
func getAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	user := User{
		Profile:            nil,
		ID:                 "",
		Disabled:           false,
		CreationTimestamp:  0,
		LastLogInTimestamp: 0,
	}
	render.JSON(w, r, user)
}

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

	// Set session expiration to 90 days.
	expiresIn := time.Hour * 24 * 14

	// Create the session cookie. This will also verify the ID token in the process.
	// The session cookie will have the same claims as the ID token.
	// To only allow session cookie setting on recent sign-in, auth_time in ID token
	// can be checked to ensure user was recently signed in before creating a session cookie.
	cookie, err := authClient.SessionCookie(firebase.FirebaseContext, req.Token, expiresIn)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "signmeup-session",
		Value:    cookie,
		MaxAge:   int(expiresIn.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("success"))
	return
}

func getCurrentUserHandler(w http.ResponseWriter, r *http.Request) {
	//get the authenticated user from the request context
	user := r.Context().Value("currentUser").(*User)
	render.JSON(w, r, user.Profile)
}

func signOutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "signmeup-session",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("success"))
	return
}
