package analytics

import (
	"fmt"
	"reflect"
	"signmeup/internal/models"
	"testing"
	"time"
)

func createTicket(userID int, status models.TicketStatus, startTime time.Time, createOffset int, claimOffset int) *models.Ticket {
	ticket := &models.Ticket{
		Status: status,
		User: models.TicketUserdata{
			UserID: fmt.Sprintf("%d", userID),
		},
	}

	ticket.CreatedAt = startTime.Add(time.Duration(createOffset) * time.Second)
	if claimOffset != -1 {
		ticket.ClaimedAt = startTime.Add(time.Duration(claimOffset) * time.Second)
	}

	return ticket
}

func createTickets() []*models.Ticket {
	startTime := time.Now()

	return []*models.Ticket{
		// Complete/claimed
		createTicket(1, models.StatusComplete, startTime, 0, 1),
		createTicket(2, models.StatusComplete, startTime, 0, 5),
		createTicket(3, models.StatusClaimed, startTime, 0, 10),

		// Waiting
		createTicket(4, models.StatusWaiting, startTime, 4, -1),
		createTicket(5, models.StatusWaiting, startTime, 5, -1),
		createTicket(6, models.StatusReturned, startTime, 0, -1),

		// Missing
		createTicket(7, models.StatusMissing, startTime, 0, -1),

		// Waiting student who has already been seen, i.e. someone who signed up again
		createTicket(1, models.StatusWaiting, startTime, 15, -1),
	}
}

func TestGenerateAnalyticsFromTickets(t *testing.T) {
	// Create some synthetic tickets
	tickets := createTickets()
	analytics := GenerateAnalyticsFromTickets(tickets)

	expectedStudentsSeen := []string{"1", "2", "3"}
	if !reflect.DeepEqual(analytics.StudentsSeen, expectedStudentsSeen) {
		t.Errorf("Expected students seen to be %v, got %v", expectedStudentsSeen, analytics.StudentsSeen)
	}

	expectedStudentsWaiting := []string{"4", "5", "6"}
	if !reflect.DeepEqual(analytics.StudentsUnseen, expectedStudentsWaiting) {
		t.Errorf("Expected students waiting to be %v, got %v", expectedStudentsWaiting, analytics.StudentsUnseen)
	}

	expectedStudentsMissing := []string{"7"}
	if !reflect.DeepEqual(analytics.StudentsMissing, expectedStudentsMissing) {
		t.Errorf("Expected students missing to be %v, got %v", expectedStudentsMissing, analytics.StudentsMissing)
	}

	expectedTimeToSeen := []int{1, 5, 10}
	if !reflect.DeepEqual(analytics.TimeToSeen, expectedTimeToSeen) {
		t.Errorf("Expected time to seen to be %v, got %v", expectedTimeToSeen, analytics.TimeToSeen)
	}
}

func TestGenerateCourseAnalytics(t *testing.T) {
	queueAnalytics := GenerateAnalyticsFromTickets(createTickets())
	courseAnalytics := GenerateCourseAnalyticsFromQueues([]*models.Queue{{Analytics: queueAnalytics}})

	if courseAnalytics.NumStudentsSeen != 3 {
		t.Errorf("Expected 3 students seen, got %d", courseAnalytics.NumStudentsSeen)
	}

	if courseAnalytics.NumStudentsUnseen != 3 {
		t.Errorf("Expected 3 students unseen, got %d", courseAnalytics.NumStudentsUnseen)
	}

	if courseAnalytics.NumStudentsMissing != 1 {
		t.Errorf("Expected 1 student missing, got %d", courseAnalytics.NumStudentsMissing)
	}
}

func approximatelyEqual(a float64, b float64) bool {
	return a-b < 0.00001
}

func TestCalculatePercentiles(t *testing.T) {
	basicDistribution := []int{2, 5, 10}
	basicPercentiles := CalculatePercentiles(basicDistribution)
	expectedBasicPercentiles := &models.Percentiles{
		P50: 5,
		P90: 9,
		P99: 9.9,
	}

	if !approximatelyEqual(basicPercentiles.P50, expectedBasicPercentiles.P50) {
		t.Errorf("Expected P50 to be %f, got %f", expectedBasicPercentiles.P50, basicPercentiles.P50)
	}
	if !approximatelyEqual(basicPercentiles.P90, expectedBasicPercentiles.P90) {
		t.Errorf("Expected P90 to be %f, got %f", expectedBasicPercentiles.P90, basicPercentiles.P90)
	}
	if !approximatelyEqual(basicPercentiles.P99, expectedBasicPercentiles.P99) {
		t.Errorf("Expected P99 to be %f, got %f", expectedBasicPercentiles.P99, basicPercentiles.P99)
	}
}
