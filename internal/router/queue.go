package router

import (
	"context"
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
	router.Use(auth.AuthCtx())

	// Queue creation
	// TODO: handle permissions here
	router.Post("/create", createQueueHandler)

	router.Route("/{queueID}", func(r chi.Router) {
		// Sets "queueID" from URL param in the context
		r.Use(QueueCtx())

		// Queue modification
		router.With(auth.RequireStaffForQueue()).Put("/edit", editQueueHandler)
		router.With(auth.RequireStaffForQueue()).Patch("/cutoff", cutoffQueueHandler)
		router.With(auth.RequireStaffForQueue()).Patch("/shuffle", shuffleQueueHandler)
		router.With(auth.RequireStaffForQueue()).Delete("/", deleteQueueHandler)

		// Ticket modification
		router.Post("/ticket", createTicketHandler)
		router.Patch("/ticket", editTicketHandler)
		router.Delete("/ticket", deleteTicketHandler)
	})

	return router
}

func QueueCtx() func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			queueID := chi.URLParam(r, "queueID")

			ctx := context.WithValue(r.Context(), "queueID", queueID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
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

// POST: /shuffle
func shuffleQueueHandler(w http.ResponseWriter, r *http.Request) {
	var req *models.ShuffleQueueRequest

	queueID := chi.URLParam(r, "queueID")
	req.QueueID = queueID

	err := repo.Repository.ShuffleQueue(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// POST: /edit
func editQueueHandler(w http.ResponseWriter, r *http.Request) {
	var req models.EditQueueRequest

	req.QueueID = chi.URLParam(r, "queueID")
	err := repo.Repository.EditQueue(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("Successfully edited queue " + req.QueueID))
}

// POST: /cutoff/{queueID}
func cutoffQueueHandler(w http.ResponseWriter, r *http.Request) {
	var req models.CutoffQueueRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = repo.Repository.CutoffQueue(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// POST: /delete
func deleteQueueHandler(w http.ResponseWriter, r *http.Request) {
	var req models.DeleteQueueRequest
	req.QueueID = chi.URLParam(r, "queueID")
	err := repo.Repository.DeleteQueue(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("Successfully edited queue " + req.QueueID))
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
