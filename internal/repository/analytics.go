package repository

import (
	"signmeup/internal/firebase"
	"signmeup/internal/models"
)

func (fr *FirebaseRepository) SaveAnalyticsForQueue(queue *models.Queue, analytics *models.QueueAnalytics) error {
	_, err := fr.firestoreClient.Collection(models.FirestoreQueuesCollection).Doc(queue.ID).Set(firebase.Context, queue)
	return err
}

func (fr *FirebaseRepository) CreateCourseAnalytics(ca *models.CourseAnalytics) error {
	_, _, err := fr.firestoreClient.Collection(models.FirestoreCourseAnalyticsCollection).Add(firebase.Context, ca)
	return err
}
