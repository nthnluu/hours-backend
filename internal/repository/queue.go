package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"signmeup/internal/firebase"
	"signmeup/internal/models"
	"signmeup/internal/qerrors"
	"time"

	"github.com/golang/glog"
	"google.golang.org/api/iterator"

	"cloud.google.com/go/firestore"
	"github.com/mitchellh/mapstructure"
)

func (fr *FirebaseRepository) CreateQueue(c *models.CreateQueueRequest) (queue *models.Queue, err error) {
	queueCourse, err := fr.GetCourseByID(c.CourseID)
	if err != nil {
		return nil, err
	}

	queue = &models.Queue{
		Title:              c.Title,
		Description:        c.Description,
		Location:           c.Location,
		EndTime:            c.EndTime,
		CourseID:           queueCourse.ID,
		AllowTicketEditing: c.AllowTicketEditing,
		ShowMeetingLinks:   c.ShowMeetingLinks,
		Course:             queueCourse,
		IsCutOff:           false,
	}

	ref, _, err := fr.firestoreClient.Collection(models.FirestoreQueuesCollection).Add(firebase.Context, map[string]interface{}{
		"title":       queue.Title,
		"description": queue.Description,
		"location":    queue.Location,
		"endTime":     queue.EndTime,
		"courseID":    queue.CourseID,
		"course": map[string]interface{}{
			"id":    queue.Course.ID,
			"title": queue.Course.Title,
			"code":  queue.Course.Code,
		},
		"tickets":            []string{},
		"isCutOff":           queue.IsCutOff,
		"allowTicketEditing": queue.AllowTicketEditing,
		"showMeetingLinks":   queue.ShowMeetingLinks,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating queue: %v", err)
	}

	queue.ID = ref.ID

	return
}

func (fr *FirebaseRepository) EditQueue(c *models.EditQueueRequest) error {
	// Update queue.
	_, err := fr.firestoreClient.Collection(models.FirestoreQueuesCollection).Doc(c.QueueID).Update(firebase.Context, []firestore.Update{
		{
			Path:  "title",
			Value: c.Title,
		}, {
			Path:  "description",
			Value: c.Description,
		}, {
			Path:  "endTime",
			Value: c.EndTime,
		}, {
			Path:  "location",
			Value: c.Location,
		}, {
			Path:  "isCutOff",
			Value: c.IsCutOff,
		}, {
			Path:  "showMeetingLinks",
			Value: c.ShowMeetingLinks,
		}, {
			Path:  "allowTicketEditing",
			Value: c.AllowTicketEditing,
		},
	})
	return err
}

func (fr *FirebaseRepository) DeleteQueue(c *models.DeleteQueueRequest) error {
	// Delete queue.
	_, err := fr.firestoreClient.Collection(models.FirestoreQueuesCollection).Doc(c.QueueID).Delete(firebase.Context)

	return err
}

func (fr *FirebaseRepository) CutoffQueue(c *models.CutoffQueueRequest) error {
	_, err := fr.firestoreClient.Collection(models.FirestoreQueuesCollection).Doc(c.QueueID).Update(firebase.Context, []firestore.Update{
		{Path: "isCutOff", Value: c.IsCutOff},
	})
	return err
}

func (fr *FirebaseRepository) ShuffleQueue(c *models.ShuffleQueueRequest) error {
	q, err := fr.GetQueue(c.QueueID)
	if err != nil {
		return qerrors.InvalidQueueError
	}

	rand.Shuffle(len(q.Tickets), func(i, j int) {
		q.Tickets[i], q.Tickets[j] = q.Tickets[j], q.Tickets[i]
	})

	_, err = fr.firestoreClient.Collection(models.FirestoreQueuesCollection).Doc(c.QueueID).Update(firebase.Context, []firestore.Update{
		{
			Path:  "tickets",
			Value: q.Tickets,
		},
	})

	if err != nil {
		return fmt.Errorf("error shuffling queue: %v", err)
	}

	return nil
}

func (fr *FirebaseRepository) CreateTicket(c *models.CreateTicketRequest) (ticket *models.Ticket, err error) {
	// Get the queue that this ticket belongs to.
	queue, err := fr.GetQueue(c.QueueID)
	if err != nil {
		return nil, qerrors.InvalidQueueError
	}

	userdata := models.TicketUserdata{
		UserID:      c.CreatedBy.ID,
		Email:       c.CreatedBy.Email,
		PhotoURL:    c.CreatedBy.PhotoURL,
		DisplayName: c.CreatedBy.DisplayName,
		Pronouns:    c.CreatedBy.Pronouns,
	}

	ticket = &models.Ticket{
		Queue:        queue,
		User:         userdata,
		CreatedAt:    time.Now(),
		Status:       models.StatusWaiting,
		Description:  c.Description,
		Anonymize:    c.Anonymize,
		BeforeCutoff: !queue.IsCutOff,
	}

	// Check that this user is not already in the queue.
	iter := fr.firestoreClient.Collection(models.FirestoreQueuesCollection).Doc(c.QueueID).Collection(models.FirestoreTicketsCollection).Documents(firebase.Context)
	for {
		// Get next document
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			glog.Warningf("an error occurred while checking for duplicate tickets: %v\n", err)
			return nil, err
		}
		// Check if matches user.
		var ticket models.Ticket
		err = mapstructure.Decode(doc.Data(), &ticket)
		if err != nil {
			return nil, err
		}

		if (ticket.User.UserID == c.CreatedBy.ID) && (ticket.Status != models.StatusComplete) {
			return nil, fmt.Errorf("error creating ticket: user already active in queue")
		}
	}

	// Add ticket to the queue's ticket collection
	ref, _, err := fr.firestoreClient.Collection(models.FirestoreQueuesCollection).Doc(c.QueueID).Collection(models.FirestoreTicketsCollection).Add(firebase.Context, map[string]interface{}{
		"user":         ticket.User,
		"createdAt":    ticket.CreatedAt,
		"status":       ticket.Status,
		"description":  ticket.Description,
		"anonymize":    ticket.Anonymize,
		"beforeCutoff": ticket.BeforeCutoff,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating ticket: %v", err)
	}

	// Add ticket to the queue's ticket array
	queueRef := fr.firestoreClient.Collection(models.FirestoreQueuesCollection).Doc(c.QueueID)
	err = fr.firestoreClient.RunTransaction(firebase.Context, func(ctx context.Context, tx *firestore.Transaction) error {
		doc, err := tx.Get(queueRef)
		if err != nil {
			return err
		}
		tickets, err := doc.DataAt("tickets")
		if err != nil {
			return err
		}
		updatedTickets := append(tickets.([]interface{}), ref.ID)
		return tx.Set(queueRef, map[string]interface{}{
			"tickets": updatedTickets,
		}, firestore.MergeAll)
	})
	if err != nil {
		glog.Errorf("error adding ticket to queue: %v\n", err)
		return nil, fmt.Errorf("error adding ticket to queue: %v", err)
	}

	return
}

func (fr *FirebaseRepository) EditTicket(c *models.EditTicketRequest) error {
	// Validate that this is a valid queue.
	queue, err := fr.GetQueue(c.QueueID)
	if err != nil {
		return qerrors.InvalidQueueError
	}

	updates := []firestore.Update{
		{
			Path:  "status",
			Value: c.Status,
		}, {
			Path:  "description",
			Value: c.Description,
		},
	}

	if c.Status == models.StatusClaimed {
		updates = append(updates, firestore.Update{
			Path:  "claimedAt",
			Value: time.Now(),
		})
		updates = append(updates, firestore.Update{
			Path:  "claimedBy",
			Value: c.ClaimedBy.ID,
		})
		notif := models.Notification{
			Title:     "You've been claimed!",
			Body:      queue.Course.Code,
			Timestamp: time.Now(),
			Type:      models.NotificationClaimed,
		}
		err := fr.AddNotification(c.OwnerID, notif)
		if err != nil {
			glog.Warningf("error sending claim notification: %v\n", err)
		}
	}

	// Edit ticket in collection.
	_, err = fr.firestoreClient.Collection(models.FirestoreQueuesCollection).Doc(c.QueueID).Collection(models.FirestoreTicketsCollection).Doc(c.ID).Update(firebase.Context, updates)
	return err
}

func (fr *FirebaseRepository) DeleteTicket(c *models.DeleteTicketRequest) error {
	// Validate that this is a valid queue.
	_, err := fr.GetQueue(c.QueueID)
	if err != nil {
		return qerrors.InvalidQueueError
	}

	// Remove ticket from tickets.
	_, err = fr.firestoreClient.Collection(models.FirestoreQueuesCollection).Doc(c.QueueID).Collection(models.FirestoreTicketsCollection).Doc(c.ID).Delete(firebase.Context)
	if err != nil {
		return err
	}

	// Remove ticket from queue.
	_, err = fr.firestoreClient.Collection(models.FirestoreQueuesCollection).Doc(c.QueueID).Update(firebase.Context, []firestore.Update{
		{
			Path:  "tickets",
			Value: firestore.ArrayRemove(c.ID),
		},
	})
	return err
}

func (fr *FirebaseRepository) MakeAnnouncement(c *models.MakeAnnouncementRequest) error {
	// Get queue.
	queue, err := fr.GetQueue(c.QueueID)
	if err != nil {
		return qerrors.InvalidQueueError
	}

	// Reject empty announcements.
	if len(c.Announcement) == 0 {
		return qerrors.InvalidBody
	}

	for _, ticketID := range queue.Tickets {
		// Get ticket from collection.
		doc, err := fr.firestoreClient.Collection(models.FirestoreQueuesCollection).Doc(c.QueueID).Collection(models.FirestoreTicketsCollection).Doc(ticketID).Get(firebase.Context)
		if err != nil {
			return err
		}
		// Deserialize.
		var ticket models.Ticket
		err = mapstructure.Decode(doc.Data(), &ticket)
		if err != nil {
			return err
		}
		// If ticket is completed, ignore.
		if ticket.Status == models.StatusComplete {
			continue
		}
		// Add an announcement to the owner of the ticket.
		notification := models.Notification{
			Title:     c.Announcement,
			Body:      queue.Course.Code,
			Timestamp: time.Now(),
			Type:      models.NotificationAnnouncement,
		}
		_ = fr.AddNotification(ticket.User.UserID, notification)
	}
	return nil
}

// GetQueue gets the Queue from the queues map corresponding to the provided queue ID.
func (fr *FirebaseRepository) GetQueue(ID string) (*models.Queue, error) {
	fr.queuesLock.RLock()
	defer fr.queuesLock.RUnlock()

	if val, ok := fr.queues[ID]; ok {
		return val, nil
	} else {
		return nil, errors.New("queue not found")
	}
}

// initializeQueuesListener starts a snapshot listener
func (fr *FirebaseRepository) initializeQueuesListener() {
	handleDocs := func(docs []*firestore.DocumentSnapshot) error {
		newQueues := make(map[string]*models.Queue)
		for _, doc := range docs {
			if !doc.Exists() {
				continue
			}

			var c models.Queue
			err := mapstructure.Decode(doc.Data(), &c)
			if err != nil {
				log.Panicf("Error destructuring document: %v", err)
				return err
			}

			c.ID = doc.Ref.ID
			newQueues[doc.Ref.ID] = &c
		}

		fr.queuesLock.Lock()
		defer fr.queuesLock.Unlock()
		fr.queues = newQueues

		return nil
	}

	done := make(chan bool)
	query := fr.firestoreClient.Collection(models.FirestoreQueuesCollection).Where("endTime", ">", time.Now().AddDate(0, 0, -1))
	go func() {
		err := fr.createCollectionInitializer(query, &done, handleDocs)
		if err != nil {
			log.Panicf("%v collection listener error: %v\n", models.FirestoreQueuesCollection, err)
		}
	}()
	<-done
}
