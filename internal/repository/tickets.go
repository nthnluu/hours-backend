package repository

import (
	"signmeup/internal/firebase"
	"signmeup/internal/models"
	"signmeup/internal/qerrors"

	"github.com/mitchellh/mapstructure"
)

func (fr *FirebaseRepository) GetUserFromTicket(queueID string, ticketID string) (*models.User, error) {
	ts, err := fr.firestoreClient.Collection(models.FirestoreQueuesCollection).Doc(queueID).Collection(models.FirestoreTicketsCollection).Doc(ticketID).Get(firebase.FirebaseContext)
	if err != nil {
		return nil, err
	}

	if !ts.Exists() {
		return nil, qerrors.EntityNotFound
	}

	var ticket *models.Ticket
	err = mapstructure.Decode(ts.Data(), &ticket)
	if err != nil {
		return nil, err
	}

	return ticket.CreatedBy, nil
}
