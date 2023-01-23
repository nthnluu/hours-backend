package router

import (
	"net/http"
	"signmeup/internal/analytics"
	"signmeup/internal/auth"
	"signmeup/internal/middleware"
	"signmeup/internal/models"
	repo "signmeup/internal/repository"
	"time"

	"github.com/go-chi/chi/v5"

	"golang.org/x/sync/errgroup"
)

func AnalyticsRoutes() *chi.Mux {
	router := chi.NewRouter()

	router.Use(auth.AuthCtx())
	router.With(auth.RequireCourseAdmin())

	router.Route("/course/{courseID}", func(router chi.Router) {
		// Sets "courseID" from URL param in the context
		router.Use(middleware.CourseCtx())

		// Create/update analytics for a course (returns it as well)
		router.Post("/", generateAnalyticsHandler)

		// Reads (most-recent) analytics for a course
		router.Get("/", getAnalyticsHandler)
	})

	router.Route("/user/{userID}", func(router chi.Router) {
		router.Get("/", getUserAnalytics)
	})

	return router
}

// POST: /course/{courseID}/
//
// Should create QueueAnalytics for the queues in the given range (if the QueueAnalytics don't exist
// on the queue already). Also, generates the CourseAnalytics using the QueueAnalytics. Writes
// the CourseAnalytics to a top-level analytics collection, and responds with it.
//
// It's unclear whether we can short-circuit reading QueueAnalytics by reading CourseAnalytics
// that have time rangers that intersect with the currently requested range.
//
// Another consideration is that if we have a cache of CourseAnalytics objects and the analytics
// schema has changed, we might need to recalculate QueueAnalytics and CourseAnalytics objects. That
// is, we can't blindly read existing *Analytics objects, since the schema/data we serve might have
// changed since our last computation.
func generateAnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	// Get the start and end ranges
	courseID := r.Context().Value("courseID").(string)
	rangeStart := time.Now()
	rangeEnd := time.Now()

	// Get the existing course analytics from the database

	// Find all the queues in that range that do _not_ have analytics data of the current
	// version (or that do not have any analytics at all)
	queues, err := repo.Repository.GetQueuesInRange(courseID, rangeStart, rangeEnd)
	if err != nil {
		// TODO(neil): Handle error canonically
	}

	// Generate/refresh QueueAnalytics for each of these queues (if needed)
	wg := errgroup.Group{}

	for _, queue := range queues {
		queue := queue

		wg.Go(func() error {
			// Function that takes the queue and says whether we need to
			if analytics.ShouldGenerateAnalytics(queue) {
				tickets, err := repo.Repository.GetTicketsForQueue(queue.ID)
				if err != nil {
					return err
				}

				analytics := analytics.GenerateAnalyticsFromTickets(tickets)
				queue.Analytics = analytics

				err = saveAnalyticsForQueue(queue, analytics)
				if err != nil {
					return err
				}
			}

			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		// TODO(neil): Handle the error
	}

	// Using the new QueueAnalytics (and any existing ones of the currect version), generate
	// CourseAnalytics.
	courseAnalytics := analytics.GenerateCourseAnalyticsFromQueues(queues)
	courseAnalytics.StartRange = rangeStart
	courseAnalytics.EndRange = rangeEnd
	courseAnalytics.CourseID = courseID

	// Write the CourseAnalytics to the database
}

// Uses the firebase client to save the given analytics onto the given queue
func saveAnalyticsForQueue(queue *models.Queue, analytics *models.QueueAnalytics) error {
	return nil
}

// GET: /course/{courseID}/
//
// Reads the most recent CourseAnalytics from the collection.
func getAnalyticsHandler(w http.ResponseWriter, r *http.Request) {

}

// GET: /user/{userID}
// Note: the userID is not parsed out of the request
//
// TODO: We need a separate read path for this. It is not handled by repository/analytics.
func getUserAnalytics(w http.ResponseWriter, r *http.Request) {

}
