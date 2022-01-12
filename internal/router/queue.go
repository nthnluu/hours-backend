package router

import (
	"encoding/json"
	"net/http"
	"signmeup/internal/auth"
	"signmeup/internal/models"
	repo "signmeup/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func QueueRoutes() *chi.Mux {
	router := chi.NewRouter()
	router.With(auth.RequireAuth(true)).Post("/", createQueueHandler)
	router.With(auth.RequireAuth(true)).Post("/{queueID}", createTicketHandler)
	return router
}

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

func createTicketHandler(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTicketRequest
	queueID := chi.URLParam(r, "queueID")

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req.QueueID = queueID

	ticket, err := repo.Repository.CreateTicket(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, ticket)
}
