package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
	"queue/internal/auth"
)

func Routes() *chi.Mux {
	router := chi.NewRouter()
	router.Use(
		middleware.Logger, // Log API Request Calls
	)

	router.Route("/v1", func(r chi.Router) {
		r.Mount("/users", auth.Routes())
	})

	return router
}

func Start() {
	router := Routes()
	log.Fatal(http.ListenAndServe(":8080", router))
}
