package auth

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"log"
	"net/http"
)

func Routes() *chi.Mux {
	router := chi.NewRouter()
	router.Get("/", getAllUsersHandler)
	router.Post("/", createUserHandler)
	router.Get("/{userID}", getUserHandler)
	router.Delete("/{userID}", deleteUserHandler)
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
