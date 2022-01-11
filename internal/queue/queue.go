package queue

import "log"

var repository Repository

// CreateQueue creates a course using the provided Title, Description, and Course ID.
func CreateQueue(queue *CreateQueueRequest) (*Queue, error) {
	createdQueue, err := repository.CreateQueue(queue)
	if err != nil {
		return nil, err
	}

	return createdQueue, nil
}

// CreateTicket creates a ticket within the given queue.
func CreateTicket(ticket *CreateTicketRequest) (*Ticket, error) {
	createdTicket, err := repository.CreateTicket(ticket)
	if err != nil {
		return nil, err
	}

	return createdTicket, nil
}

func init() {
	repo, err := NewFirebaseRepository()
	if err != nil {
		log.Panicf("error creating Firebase user repository: %v\n", err)
	}

	repository = repo
}
