package server

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
	"log"
	"net/http"
	"signmeup/internal/auth"
	"signmeup/internal/config"
	"signmeup/internal/course"
	"signmeup/internal/queue"
)

func Routes() *chi.Mux {
	router := chi.NewRouter()
	router.Use(
		middleware.Logger, // Log API Request Calls
	)

	router.Route("/v1", func(r chi.Router) {
		r.Mount("/users", auth.Routes())
		r.Mount("/courses", course.Routes())
		r.Mount("/queues", queue.Routes())
		// mount routes here...
	})

	return router
}

func Start() {
	if config.Config == nil {
		log.Panic("❌ Missing or invalid configuration!")
	}

	router := Routes()
	c := cors.New(cors.Options{
		AllowedOrigins:   config.Config.AllowedOrigins,
		AllowedHeaders:   []string{"Cookie", "Content-Type"},
		AllowedMethods:   []string{"GET", "POST", "DELETE"},
		ExposedHeaders:   []string{"Set-Cookie"},
		AllowCredentials: true,
	})

	handler := c.Handler(router)
	log.Printf("Server is listening on port %v\n", config.Config.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", config.Config.Port), handler))
}
