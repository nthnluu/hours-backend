package router

import (
	"encoding/json"
	"log"
	"net/http"
	"signmeup/internal/auth"
	"signmeup/internal/middleware"
	"signmeup/internal/models"
	repo "signmeup/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func QueueRoutes() *chi.Mux {
	router := chi.NewRouter()
	router.Use(auth.AuthCtx())

	// Queue creation
	// We can't do /{courseID}/create since that will conflate with the ^/{queueID} routes
	router.With(middleware.CourseCtx(), auth.RequireStaffForCourse()).Post("/create/{courseID}", createQueueHandler)

	router.Route("/{queueID}", func(router chi.Router) {
		// Sets "queueID" from URL param in the context
		router.Use(middleware.QueueCtx())

		// Queue modification
		router.With(auth.RequireQueueStaff()).Post("/edit", editQueueHandler)
		router.With(auth.RequireQueueStaff()).Patch("/cutoff", cutoffQueueHandler)
		router.With(auth.RequireQueueStaff()).Patch("/shuffle", shuffleQueueHandler)
		router.With(auth.RequireQueueStaff()).Delete("/", deleteQueueHandler)

		// Ticket modification
		router.Post("/ticket", createTicketHandler)
		router.Patch("/ticket", editTicketHandler)
		router.Post("/ticket/delete", deleteTicketHandler)
	})

	return router
}

// POST: /{courseID}/create
func createQueueHandler(w http.ResponseWriter, r *http.Request) {
	var req models.CreateQueueRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req.CourseID = r.Context().Value("courseID").(string)

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

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req.QueueID = r.Context().Value("queueID").(string)

	err = repo.Repository.ShuffleQueue(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// POST: /edit
func editQueueHandler(w http.ResponseWriter, r *http.Request) {
	var req models.EditQueueRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req.QueueID = r.Context().Value("queueID").(string)

	err = repo.Repository.EditQueue(&req)
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

	req.QueueID = r.Context().Value("queueID").(string)
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

	user, err := auth.GetUserFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
	}

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req.QueueID = r.Context().Value("queueID").(string)
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
	req.QueueID = r.Context().Value("queueID").(string)

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
	req.QueueID = r.Context().Value("queueID").(string)

	err = repo.Repository.DeleteTicket(req)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("Successfully edited ticket " + req.ID))
}
