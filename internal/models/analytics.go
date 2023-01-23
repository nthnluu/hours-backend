package models

import "time"

// Percentiles is a generic struct for storing percentiles for any distribution of data
type Percentiles struct {
	P50 float64
	P90 float64
	P99 float64
}

// QueueAnalytics are added to each document in the queues collection under the analytics field.
//
// They are lazily computed: we calculate them only when a queue is needed for some CourseAnalytics
// query.
//
// TODO(neil): If this is the struct we get back from Firebase, we'll need struct properties.
type QueueAnalytics struct {
	// The version of the analytics struct being stored on the queue document. We regenerate
	// analytics for a queue if the version of the already-stored analytics is less than the version
	// of the code running the analytics.
	Version int

	// TimeToSeen contains the number of seconds between each claimed ticket's creation time and
	// claim time.
	//
	// The length of this field should be equal to StudentsSeen.
	TimeToSeen []int

	// The IDs of the students seen (StatusClaimed), remaining (StatusWaiting or StatusReturned),
	// and missing (StatusMissing).
	StudentsSeen    []string
	StudentsUnseen  []string
	StudentsMissing []string
}

// CourseAnalytics are stored within a separate analytics collection, and are created using a mix
// of QueueAnalytics from queues within a certain time interval as well as individual queries.
type CourseAnalytics struct {
	// The CourseID for which these analytics are calculated
	CourseID string

	// StartRange and EndRange form the closed interval over which these analytics have been created
	StartRange time.Time
	EndRange   time.Time

	// TimeToSeen (TTS) is a distribution over the delta between creating a ticket and the ticket
	// being claimed. It does not include data for students who aren't claimed (since their TTS is
	// undefined).
	TimeToSeen Percentiles

	// The aggregate number of students who were seen, remaining, and missing over all queues
	// in the time range. Absolute numbers are given here, and the client can compute percentages
	// if needed.
	NumStudentsSeen    int
	NumStudentsUnseen  int
	NumStudentsMissing int
}
