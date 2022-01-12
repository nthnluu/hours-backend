package repository

import (
	"fmt"
	"signmeup/internal/firebase"
	"signmeup/internal/models"
	"signmeup/internal/qerrors"

	"cloud.google.com/go/firestore"
	"github.com/mitchellh/mapstructure"
)

// CreateQueue creates a queue with the given details, and returns the queue.
func (fr *FirebaseRepository) CreateQueue(cqr *models.CreateQueueRequest) (queue *models.Queue, err error) {
	course, err := fr.getCourse(cqr.CourseID)
	if err != nil {
		return nil, err
	}

	queue = &models.Queue{
		Title:       cqr.Title,
		Description: cqr.Description,
		CourseID:    course.ID,
		Course:      course,
		IsActive:    true,
	}

	ref, _, err := fr.firestoreClient.Collection(models.FirestoreQueuesCollection).Add(firebase.FirebaseContext, map[string]interface{}{
		"title":       queue.Title,
		"description": queue.Description,
		"courseID":    queue.CourseID,
		"course": map[string]interface{}{
			"id":    queue.Course.ID,
			"title": queue.Course.Title,
			"code":  queue.Course.Code,
		},
		"tickets": []string{},
		"active":  queue.IsActive,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating queue: %v\n", err)
	}

	queue.ID = ref.ID

	return
}

func (fr *FirebaseRepository) CreateTicket(c *models.CreateTicketRequest) (ticket *models.Ticket, err error) {
	queue, err := fr.getQueue(c.QueueID)
	if err != nil {
		return nil, qerrors.InvalidQueueError
	}

	ticket = &models.Ticket{
		Queue:     queue,
		CreatedBy: c.CreatedBy,
		Status:    models.StatusWaiting,
	}

	// Add ticket to the queue's ticket collection
	ref, _, err := fr.firestoreClient.Collection(models.FirestoreQueuesCollection).Doc(queue.ID).Collection(models.FirestoreTicketsCollection).Add(firebase.FirebaseContext, map[string]interface{}{
		"createdBy": ticket.CreatedBy,
		"status":    ticket.Status,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating ticket: %v\n", err)
	}
	queue.ID = ref.ID

	// Add ticket to the queue's ticket array
	_, err = fr.firestoreClient.Collection(models.FirestoreQueuesCollection).Doc(queue.ID).Update(firebase.FirebaseContext, []firestore.Update{
		{
			Path:  "tickets",
			Value: append(queue.Tickets, queue.ID),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error adding ticket to queue: %v\n", err)
	}

	return
}

// getQueue gets the Queue from the queues map corresponding to the provided queue ID.
func (r FirebaseRepository) getQueue(id string) (*models.Queue, error) {
	r.queuesLock.RLock()
	defer r.queuesLock.RUnlock()

	if val, ok := r.queues[id]; ok {
		return val, nil
	} else {
		return nil, fmt.Errorf("No profile found for ID %v\n", id)
	}
}

// initializeQueuesListener starts a snapshot listener
func (fr *FirebaseRepository) initializeQueuesListener() {
	handleDoc := func(doc *firestore.DocumentSnapshot) error {
		fr.queuesLock.Lock()
		defer fr.queuesLock.Unlock()

		var q models.Queue
		err := mapstructure.Decode(doc.Data(), &q)
		if err != nil {
			return err
		}

		c, err := fr.getCourse(q.CourseID)
		if err != nil {
			fmt.Println(q.CourseID)
			return err
		}

		q.Course = c
		fr.queues[doc.Ref.ID] = &q

		return nil
	}

	done := make(chan bool)
	go fr.createCollectionInitializer(models.FirestoreQueuesCollection, &done, handleDoc)
	<-done
}
