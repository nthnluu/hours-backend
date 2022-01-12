package repository

import (
	"fmt"
	"signmeup/internal/firebase"
	"signmeup/internal/models"
	"signmeup/internal/qerrors"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/mitchellh/mapstructure"
)

func (fr *FirebaseRepository) CreateQueue(c *models.CreateQueueRequest) (queue *models.Queue, err error) {
	queueCourse, err := fr.GetCourseByID(c.CourseID)
	if err != nil {
		return nil, err
	}

	queue = &models.Queue{
		Title:       c.Title,
		Description: c.Description,
		CourseID:    queueCourse.ID,
		Course:      queueCourse,
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
	// Get the queue that this ticket belongs to.
	queue, err := fr.getQueue(c.QueueID)
	if err != nil {
		return nil, qerrors.InvalidQueueError
	}

	// Construct ticket.
	ticket = &models.Ticket{
		Queue:       queue,
		CreatedBy:   c.CreatedBy,
		Status:      models.StatusWaiting,
		Description: c.Description,
	}

	// Add ticket to the queue's ticket collection
	ref, _, err := fr.firestoreClient.Collection(models.FirestoreQueuesCollection).Doc(queue.ID).Collection(models.FirestoreTicketsCollection).Add(firebase.FirebaseContext, map[string]interface{}{
		"createdBy":   ticket.CreatedBy.Profile,
		"createdAt":   time.Now(),
		"status":      ticket.Status,
		"description": ticket.Description,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating ticket: %v\n", err)
	}

	// Add ticket to the queue's ticket array
	_, err = fr.firestoreClient.Collection(models.FirestoreQueuesCollection).Doc(queue.ID).Update(firebase.FirebaseContext, []firestore.Update{
		{
			Path:  "tickets",
			Value: append(queue.Tickets, ref.ID),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error adding ticket to queue: %v\n", err)
	}
	return
}

func (fr *FirebaseRepository) EditTicket(c *models.EditTicketRequest) error {
	// Get the queue that this ticket belongs to.
	queue, err := fr.getQueue(c.QueueID)
	if err != nil {
		return qerrors.InvalidQueueError
	}

	// Edit ticket in collection.
	_, err = fr.firestoreClient.Collection(models.FirestoreQueuesCollection).Doc(queue.ID).Collection(models.FirestoreTicketsCollection).Doc(c.ID).Update(firebase.FirebaseContext, []firestore.Update{
		{
			Path:  "status",
			Value: c.Status,
		}, {
			Path:  "description",
			Value: c.Description,
		},
	})
	return err
}

func (fr *FirebaseRepository) DeleteTicket(c *models.DeleteTicketRequest) error {
	// Get the queue that this ticket belongs to.
	queue, err := fr.getQueue(c.QueueID)
	if err != nil {
		return qerrors.InvalidQueueError
	}

	// Remove ticket from tickets.
	_, err = fr.firestoreClient.Collection(models.FirestoreQueuesCollection).Doc(queue.ID).Collection(models.FirestoreTicketsCollection).Doc(c.ID).Delete(firebase.FirebaseContext)
	if err != nil {
		return err
	}

	// Remove ticket from queue.
	_, err = fr.firestoreClient.Collection(models.FirestoreQueuesCollection).Doc(queue.ID).Update(firebase.FirebaseContext, []firestore.Update{
		{
			Path:  "tickets",
			Value: firestore.ArrayRemove(c.ID),
		},
	})
	return err
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

		c, err := fr.GetCourseByID(q.CourseID)
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
