package router

import (
	"encoding/json"
	"log"
	"net/http"
	"signmeup/internal/auth"
	"signmeup/internal/models"
	repo "signmeup/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func QueueRoutes() *chi.Mux {
	router := chi.NewRouter()

	// Queue creation
	router.With(auth.RequireAuth(false)).Post("/create", createQueueHandler)

	// Ticket modification
	// TODO(neil): Make this more semantically REST-y!
	router.With(auth.RequireAuth(false)).Post("/ticket/create/{queueID}", createTicketHandler)
	router.With(auth.RequireAuth(false)).Post("/ticket/edit/{queueID}", editTicketHandler)
	router.With(auth.RequireAuth(false)).Post("/ticket/delete/{queueID}", deleteTicketHandler)

	return router
}

// POST: /create
func createQueueHandler(w http.ResponseWriter, r *http.Request) {
	var req models.CreateQueueRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	queue, err := repo.Repository.CreateQueue(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	render.JSON(w, r, queue)
}

// POST: /ticket/create/{queueID}
func createTicketHandler(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTicketRequest
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

	ticket, err := repo.Repository.CreateTicket(&req)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, ticket)
}

// POST: /ticket/edit/{queueID}
func editTicketHandler(w http.ResponseWriter, r *http.Request) {
	var req *models.EditTicketRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	queueID := chi.URLParam(r, "queueID")
	req.QueueID = queueID

	err = repo.Repository.EditTicket(req)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("Successfully edited ticket " + req.ID))
}

// POST: /ticket/delete/{queueID}
func deleteTicketHandler(w http.ResponseWriter, r *http.Request) {
	var req *models.DeleteTicketRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	queueID := chi.URLParam(r, "queueID")
	req.QueueID = queueID

	err = repo.Repository.DeleteTicket(req)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("Successfully edited ticket " + req.ID))
}
