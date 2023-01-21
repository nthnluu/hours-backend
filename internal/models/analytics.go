package models

import "time"

// Percentiles is a generic struct for storing percentiles for any distribution of data.
type Percentiles struct {
	P50 int
	P90 int
	P99 int
}

// QueueAnalytics are added to each document in the queues collection under the analytics field.
// TODO(neil): If this is the struct we get back from Firebase, we'll need struct properties.
type QueueAnalytics struct {
	// TimeToSeen (TTS) is a distribution over the delta between creating a ticket and the ticket
	// being claimed. It does not include data for students who aren't claimed (since their TTS is
	// undefined).
	TimeToSeen Percentiles

	// NumberTAs is the number of TAs who claimed at least 1 student in the queue.
	NumberTAs int

	// StudentsSeen is the number of tickets in the queue with a StatusComplete status
	StudentsSeen int
	// StudentsRemaining are the tickets with StatusWaiting
	StudentsRemaining int
	// StudentsMissing are the tickets with StatusMissing
	StudentsMissing int
}

// CourseAnalytics are stored within a separate analytics collection, and are created using a mix
// of QueueAnalytics from queues within a certain time interval as well as individual queries.
//
// TODO(neil): Think about whether this structuring of data makes it easy to not recalculate data
// that we've already calculated.
type CourseAnalytics struct {
	CourseID string

	// StartTime and EndTime form the closed interval over which these analytics have been created
	StartTime time.Time
	EndTime   time.Time

	TimeToSeen Percentiles

	StudentsSeen     Percentiles
	StudensRemaining Percentiles
	StudentsMissing  Percentiles
}
