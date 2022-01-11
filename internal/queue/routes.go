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
	router.With(auth.RequireAuth(false)).Post("/create", createQueueHandler)
	router.With(auth.RequireAuth(false)).Post("/ticket/create/{queueID}", createTicketHandler)
	router.With(auth.RequireAuth(false)).Post("/ticket/edit/{queueID}", editTicketHandler)
	router.With(auth.RequireAuth(false)).Post("/ticket/delete/{queueID}", deleteTicketHandler)
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

func editTicketHandler(w http.ResponseWriter, r *http.Request) {
	var req EditTicketRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	queueID := chi.URLParam(r, "queueID")
	req.QueueID = queueID

	err = EditTicket(&req)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("Successfully edited ticket " + req.ID))
}

func deleteTicketHandler(w http.ResponseWriter, r *http.Request) {
	var req DeleteTicketRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	queueID := chi.URLParam(r, "queueID")
	req.QueueID = queueID

	err = DeleteTicket(&req)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("Successfully edited ticket " + req.ID))
}
