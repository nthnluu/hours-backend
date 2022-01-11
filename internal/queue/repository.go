package queue

import (
	"cloud.google.com/go/firestore"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"signmeup/internal/course"
	"signmeup/internal/firebase"
	"sync"
	"time"
)

// Repository encapsulates the logic to access queues from a database.
type Repository interface {
	// CreateQueue saves a new queue into the database.
	CreateQueue(course *CreateQueueRequest) (*Queue, error)
	// CreateTicket saves a new ticket into the database.
	CreateTicket(course *CreateTicketRequest) (*Ticket, error)
	// EditTicket edits a ticket.
	EditTicket(c *EditTicketRequest) error
	// Deleteticket deletes a ticket.
	DeleteTicket(c *DeleteTicketRequest) error
}

type firebaseRepository struct {
	firestoreClient *firestore.Client

	queuesLock *sync.RWMutex
	queues     map[string]*Queue
}

const (
	FirestoreQueuesCollection  = "queues"
	FirestoreTicketsCollection = "tickets"
)

// NewFirebaseRepository creates a new user repository with Firebase as the database.
func NewFirebaseRepository() (Repository, error) {
	repository := &firebaseRepository{
		queuesLock: &sync.RWMutex{},
		queues:     make(map[string]*Queue),
	}

	firestoreClient, err := firebase.FirebaseApp.Firestore(firebase.FirebaseContext)
	if err != nil {
		return nil, fmt.Errorf("Firestore client error: %v\n", err)
	}
	repository.firestoreClient = firestoreClient

	var wg sync.WaitGroup

	wg.Add(1)
	log.Println("⏳ Starting queues collection listener...")
	go func() {
		err := repository.startQueuesListener(&wg)
		if err != nil {
			log.Fatalf("queues collection listner error: %v\n", err)
		}
	}()
	wg.Wait()

	return repository, nil
}

func (r *firebaseRepository) CreateQueue(c *CreateQueueRequest) (queue *Queue, err error) {
	queueCourse, err := course.GetCourseByID(c.CourseID)
	if err != nil {
		return nil, err
	}

	queue = &Queue{
		Title:       c.Title,
		Description: c.Description,
		CourseID:    queueCourse.ID,
		Course:      queueCourse,
		IsActive:    true,
	}

	ref, _, err := r.firestoreClient.Collection(FirestoreQueuesCollection).Add(firebase.FirebaseContext, map[string]interface{}{
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

func (r *firebaseRepository) CreateTicket(c *CreateTicketRequest) (ticket *Ticket, err error) {
	// Get the queue that this ticket belongs to.
	queue, err := r.getQueue(c.QueueID)
	if err != nil {
		return nil, InvalidQueueError
	}

	// Construct ticket.
	ticket = &Ticket{
		Queue:     queue,
		CreatedBy: c.CreatedBy,
		Status:    StatusWaiting,
		Description: c.Description,
	}

	// Add ticket to the queue's ticket collection
	ref, _, err := r.firestoreClient.Collection(FirestoreQueuesCollection).Doc(queue.ID).Collection(FirestoreTicketsCollection).Add(firebase.FirebaseContext, map[string]interface{}{
		"createdBy": 	ticket.CreatedBy.Profile,
		"createdAt": 	time.Now(),
		"status":    	ticket.Status,
		"description":	ticket.Description,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating ticket: %v\n", err)
	}

	// Add ticket to the queue's ticket array
	_, err = r.firestoreClient.Collection(FirestoreQueuesCollection).Doc(queue.ID).Update(firebase.FirebaseContext, []firestore.Update{
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

func (r *firebaseRepository) EditTicket(c *EditTicketRequest) error {
	// Get the queue that this ticket belongs to.
	queue, err := r.getQueue(c.QueueID)
	if err != nil {
		return InvalidQueueError
	}

	// Edit ticket in collection.
	_, err = r.firestoreClient.Collection(FirestoreQueuesCollection).Doc(queue.ID).Collection(FirestoreTicketsCollection).Doc(c.ID).Update(firebase.FirebaseContext, []firestore.Update{
		{
			Path: "status",
			Value: c.Status,
		}, {
			Path: "description",
			Value: c.Description,
		},
	})
	return err
}

func (r *firebaseRepository) DeleteTicket(c *DeleteTicketRequest) error {
	// Get the queue that this ticket belongs to.
	queue, err := r.getQueue(c.QueueID)
	if err != nil {
		return InvalidQueueError
	}

	// Remove ticket from tickets.
	_, err = r.firestoreClient.Collection(FirestoreQueuesCollection).Doc(queue.ID).Collection(FirestoreTicketsCollection).Doc(c.ID).Delete(firebase.FirebaseContext)
	if err != nil {
		return err
	}

	// Remove ticket from queue.
	_, err = r.firestoreClient.Collection(FirestoreQueuesCollection).Doc(queue.ID).Update(firebase.FirebaseContext, []firestore.Update{
		{
			Path:  "tickets",
			Value: firestore.ArrayRemove(c.ID),
		},
	})
	return err
}

// getQueue gets the Queue from the queues map corresponding to the provided queue ID.
func (r firebaseRepository) getQueue(id string) (*Queue, error) {
	r.queuesLock.RLock()
	defer r.queuesLock.RUnlock()

	if val, ok := r.queues[id]; ok {
		return val, nil
	} else {
		return nil, fmt.Errorf("No profile found for ID %v\n", id)
	}
}

func (r firebaseRepository) startQueuesListener(wg *sync.WaitGroup) error {
	it := r.firestoreClient.Collection(FirestoreQueuesCollection).Snapshots(firebase.FirebaseContext)
	var doOnce sync.Once

	for {
		snap, err := it.Next()
		// DeadlineExceeded will be returned when ctx is cancelled.
		if status.Code(err) == codes.DeadlineExceeded {
			return nil
		}
		if err != nil {
			return fmt.Errorf("Snapshots.Next: %v", err)
		}
		if snap != nil {
			r.queuesLock.Lock()

			for {
				doc, err := snap.Documents.Next()
				if err == iterator.Done {
					doOnce.Do(func() {
						log.Println("✅ Started queues collection listener.")
						wg.Done()
					})
					r.queuesLock.Unlock()
					break
				}
				if err != nil {
					return fmt.Errorf("Documents.Next: %v", err)
				}

				var queue Queue
				err = mapstructure.Decode(doc.Data(), &queue)
				if err != nil {
					return err
				}

				c, err := course.GetCourseByID(queue.CourseID)
				if err != nil {
					fmt.Println(queue.CourseID)
					return err
				}
				queue.Course = c
				queue.ID = doc.Ref.ID
				r.queues[doc.Ref.ID] = &queue
			}
		}
	}
}
