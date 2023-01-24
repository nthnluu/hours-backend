package repository

import (
	"fmt"
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
		FaceMaskPolicy:     c.FaceMaskPolicy,
		RejoinCooldown:     c.RejoinCooldown,
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
		"completedTickets":   []string{},
		"pendingTickets":     []string{},
		"isCutOff":           queue.IsCutOff,
		"allowTicketEditing": queue.AllowTicketEditing,
		"showMeetingLinks":   queue.ShowMeetingLinks,
		"faceMaskPolicy":     queue.FaceMaskPolicy,
		"rejoinCooldown":     queue.RejoinCooldown,
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
		{
			Path:  "faceMaskPolicy",
			Value: c.FaceMaskPolicy,
		},
		{
			Path:  "rejoinCooldown",
			Value: c.RejoinCooldown,
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

	rand.Shuffle(len(q.PendingTickets), func(i, j int) {
		q.PendingTickets[i], q.PendingTickets[j] = q.PendingTickets[j], q.PendingTickets[i]
	})

	// TODO: This shuffle operation is not atomic. It reads the current PendingTickets array, shuffles it, and writes
	// it back.
	_, err = fr.firestoreClient.Collection(models.FirestoreQueuesCollection).Doc(c.QueueID).Update(firebase.Context, []firestore.Update{
		{
			Path:  "pendingTickets",
			Value: q.PendingTickets,
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
		Queue:       queue,
		User:        userdata,
		CreatedAt:   time.Now(),
		Status:      models.StatusWaiting,
		Description: c.Description,
		Anonymize:   c.Anonymize,
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

		// Check if any ticket violates the queue cooldown.
		createdByCurrentUser := ticket.User.UserID == c.CreatedBy.ID
		ticketIsComplete := ticket.Status == models.StatusComplete
		canNeverRejoin := queue.RejoinCooldown == -1
		cooldownNotElapsed := time.Now().Sub(ticket.CompletedAt).Minutes() < float64(queue.RejoinCooldown)

		if createdByCurrentUser && ticketIsComplete && (canNeverRejoin || cooldownNotElapsed) {
			return nil, qerrors.QueueCooldownError
		}

		// Errors if the current user has an incomplete ticket in the queue
		if createdByCurrentUser && !ticketIsComplete {
			return nil, qerrors.ActiveTicketError
		}
	}

	// Add ticket to the queue's ticket collection
	ref, _, err := fr.firestoreClient.Collection(models.FirestoreQueuesCollection).Doc(c.QueueID).Collection(models.FirestoreTicketsCollection).Add(firebase.Context, map[string]interface{}{
		"user":        ticket.User,
		"createdAt":   ticket.CreatedAt,
		"status":      ticket.Status,
		"description": ticket.Description,
		"anonymize":   ticket.Anonymize,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating ticket: %v", err)
	}

	// Add ticket to the queue's ticket array and the queue's visible tickets array
	_, err = fr.firestoreClient.Collection(models.FirestoreQueuesCollection).Doc(c.QueueID).Update(firebase.Context, []firestore.Update{
		{Path: "pendingTickets", Value: firestore.ArrayUnion(ref.ID)},
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

	ticketUpdates := []firestore.Update{
		{
			Path:  "status",
			Value: c.Status,
		}, {
			Path:  "description",
			Value: c.Description,
		},
	}

	if c.Status == models.StatusClaimed {
		// The ticket is being claimed.
		ticketUpdates = append(ticketUpdates, firestore.Update{
			Path:  "claimedAt",
			Value: time.Now(),
		})
		ticketUpdates = append(ticketUpdates, firestore.Update{
			Path:  "claimedBy",
			Value: c.ClaimedBy.ID,
		})
		notification := models.Notification{
			Title:     "You've been claimed!",
			Body:      queue.Course.Code,
			Timestamp: time.Now(),
			Type:      models.NotificationClaimed,
		}
		err := fr.AddNotification(c.OwnerID, notification)
		if err != nil {
			glog.Warningf("error sending claim notification: %v\n", err)
		}
	} else if c.Status == models.StatusComplete {
		// Ticket is being marked complete.
		ticketUpdates = append(ticketUpdates, firestore.Update{
			Path:  "completedAt",
			Value: time.Now(),
		})

		// Remove the ticket from the visible tickets array and move it to the completed tickets array.
		_, err := fr.firestoreClient.Collection(models.FirestoreQueuesCollection).Doc(c.QueueID).Update(firebase.Context, []firestore.Update{
			{Path: "pendingTickets", Value: firestore.ArrayRemove(c.ID)},
			{Path: "completedTickets", Value: firestore.ArrayUnion(c.ID)},
		})
		if err != nil {
			return err
		}
	}

	// Edit ticket in collection.
	_, err = fr.firestoreClient.Collection(models.FirestoreQueuesCollection).Doc(c.QueueID).Collection(models.FirestoreTicketsCollection).Doc(c.ID).Update(firebase.Context, ticketUpdates)
	return err
}

func (fr *FirebaseRepository) DeleteTicket(c *models.DeleteTicketRequest) error {
	_, err := fr.firestoreClient.Collection(models.FirestoreQueuesCollection).Doc(c.QueueID).Update(firebase.Context, []firestore.Update{
		{Path: "pendingTickets", Value: firestore.ArrayRemove(c.ID)},
	})
	if err != nil {
		return err
	}

	// Remove ticket from tickets collection.
	_, err = fr.firestoreClient.Collection(models.FirestoreQueuesCollection).Doc(c.QueueID).Collection(models.FirestoreTicketsCollection).Doc(c.ID).Delete(firebase.Context)
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

	for _, ticketID := range queue.PendingTickets {
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
	doc, err := fr.firestoreClient.Collection(models.FirestoreQueuesCollection).Doc(ID).Get(firebase.Context)
	if err != nil {
		return nil, err
	}

	if !doc.Exists() {
		return nil, qerrors.QueueNotFoundError
	}

	var c models.Queue
	err = mapstructure.Decode(doc.Data(), &c)
	if err != nil {
		glog.Fatalf("Error destructuring queue document: %v", err)
		return nil, qerrors.QueueNotFoundError
	}

	c.ID = doc.Ref.ID
	return &c, nil
}
