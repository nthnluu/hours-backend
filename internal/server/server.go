package server

import (
	"fmt"
	"log"
	"net/http"
	"signmeup/internal/config"
	rtr "signmeup/internal/router"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
)

func Routes() *chi.Mux {
	router := chi.NewRouter()
	router.Use(
		middleware.Logger, // Log API Request Calls
	)

	router.Route("/", func(r chi.Router) {
		r.Mount("/", rtr.HealthRoutes())
	})

	router.Route("/v1", func(r chi.Router) {
		r.Mount("/users", rtr.AuthRoutes())
		r.Mount("/courses", rtr.CourseRoutes())
		r.Mount("/queues", rtr.QueueRoutes())
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
		AllowedMethods:   []string{"GET", "POST", "DELETE", "PATCH"},
		ExposedHeaders:   []string{"Set-Cookie"},
		AllowCredentials: true,
	})

	handler := c.Handler(router)
	log.Printf("Server is listening on port %v\n", config.Config.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", config.Config.Port), handler))
}
