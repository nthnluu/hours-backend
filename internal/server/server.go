package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
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
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedHeaders:   []string{"Cookie", "Content-Type"},
		ExposedHeaders:   []string{"Set-Cookie"},
		AllowCredentials: true,
		// Enable Debugging for testing, consider disabling in production
		Debug: false,
	})

	handler := c.Handler(router)
	log.Fatal(http.ListenAndServe(":8080", handler))
}
