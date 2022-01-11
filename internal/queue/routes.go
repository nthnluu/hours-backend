package queue

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"log"
	"net/http"
	"signmeup/internal/auth"
)

func Routes() *chi.Mux {
	router := chi.NewRouter()
	router.With(auth.RequireAuth(false)).Post("/", createQueueHandler)
	router.With(auth.RequireAuth(false)).Post("/{queueID}", createTicketHandler)
	return router
}

func createQueueHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateQueueRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	queue, err := CreateQueue(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	render.JSON(w, r, queue)
}

func createTicketHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateTicketRequest
	queueID := chi.URLParam(r, "queueID")
	user, err := auth.GetUserFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
	}

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req.QueueID = queueID
	req.CreatedBy = user

	ticket, err := CreateTicket(&req)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, ticket)
}
