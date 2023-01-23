package analytics

import (
	"signmeup/internal/models"
	"sort"
)

const (
	QUEUE_ANALYTICS_VERSION  = 1
	COURSE_ANALYTICS_VERSION = 1
)

// Returns whether the queue has no analytics, or the analytics that it does have is of an old
// version.
func ShouldGenerateAnalytics(queue *models.Queue) bool {
	return queue.Analytics == nil || queue.Analytics.Version < QUEUE_ANALYTICS_VERSION
}

// Uses the queue to generate queue analytics. Does not need/use any Firebase connection.
func GenerateAnalyticsFromTickets(tickets []*models.Ticket) *models.QueueAnalytics {
	analytics := &models.QueueAnalytics{
		Version: QUEUE_ANALYTICS_VERSION,

		TimeToSeen: make([]int, 0),

		StudentsSeen:    make([]string, 0),
		StudentsUnseen:  make([]string, 0),
		StudentsMissing: make([]string, 0),
	}

	analytics.Version = QUEUE_ANALYTICS_VERSION

	for _, ticket := range tickets {
		if ticket.Status == models.StatusComplete || ticket.Status == models.StatusClaimed {
			analytics.StudentsSeen = append(analytics.StudentsSeen, ticket.User.UserID)

			timeToSeen := ticket.ClaimedAt.Sub(ticket.CreatedAt).Seconds()
			analytics.TimeToSeen = append(analytics.TimeToSeen, int(timeToSeen))
		} else if ticket.Status == models.StatusWaiting || ticket.Status == models.StatusReturned {
			analytics.StudentsUnseen = append(analytics.StudentsUnseen, ticket.User.UserID)
		} else if ticket.Status == models.StatusMissing {
			analytics.StudentsMissing = append(analytics.StudentsMissing, ticket.User.UserID)
		} else {
			// TODO(neil): Handle error
		}
	}

	return analytics
}

func QueuesToQueueAnalytics(queues []*models.Queue) []*models.QueueAnalytics {
	var analytics []*models.QueueAnalytics

	for _, queue := range queues {
		analytics = append(analytics, queue.Analytics)
	}

	return analytics
}

func GenerateCourseAnalyticsFromQueues(queues []*models.Queue) *models.CourseAnalytics {
	queueAnalytics := QueuesToQueueAnalytics(queues)
	var courseAnalytics models.CourseAnalytics

	var timeToSeen []int
	for _, queueAnalytics := range queueAnalytics {
		courseAnalytics.NumStudentsSeen += len(queueAnalytics.StudentsSeen)
		courseAnalytics.NumStudentsUnseen += len(queueAnalytics.StudentsUnseen)
		courseAnalytics.NumStudentsMissing += len(queueAnalytics.StudentsMissing)
		timeToSeen = append(timeToSeen, queueAnalytics.TimeToSeen...)
	}
	courseAnalytics.TimeToSeen = CalculatePercentiles(timeToSeen)

	return &courseAnalytics
}

func CalculatePercentiles(data []int) models.Percentiles {
	if len(data) == 0 {
		return models.Percentiles{}
	}

	sort.Ints(data)

	calculatePercentile := func(percentile float64) float64 {
		rank := percentile / 100 * float64(len(data)-1)
		rankInt := int(rank)

		// If the rank is an integer, return the value at that index
		if rank == float64(rankInt) {
			return float64(data[rankInt])
		}

		// Otherwise, linearly interpolate
		baseline := data[rankInt]
		interpolation := (rank - float64(rankInt)) * float64(data[rankInt+1]-data[rankInt])

		return float64(baseline) + interpolation
	}

	return models.Percentiles{
		P50: calculatePercentile(50),
		P90: calculatePercentile(90),
		P99: calculatePercentile(99),
	}
}
