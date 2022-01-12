package repository

import (
	"fmt"
	"signmeup/internal/firebase"
	"sync"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// createCollectionInitializer creates a snapshot iterator over the given collection, and when the
// collection changes, runs a function.
func (fr *FirebaseRepository) createCollectionInitializer(
	collection string, done *chan bool, handleDoc func(doc *firestore.DocumentSnapshot) error) error {

	it := fr.firestoreClient.Collection(collection).Snapshots(firebase.FirebaseContext)
	var doOnce sync.Once

	for {
		snap, err := it.Next()

		// DeadlineExceeded will be returned when ctx is cancelled.
		if status.Code(err) == codes.DeadlineExceeded {
			return nil
		} else if err != nil {
			return fmt.Errorf("Snapshots.Next: %v", err)
		}

		// TODO: Determine why would this happen and handle accordingly
		if snap == nil {
			continue
		}

		for {
			doc, err := snap.Documents.Next()

			// iterator.Done is returned when there are no more items to return
			if err == iterator.Done {
				doOnce.Do(func() {
					*done <- true
				})

				break
			}

			if err != nil {
				return fmt.Errorf("Documents.Next: %v", err)
			}

			handleDoc(doc)
		}
	}
}
