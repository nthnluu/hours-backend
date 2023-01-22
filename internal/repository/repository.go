package repository

import (
	"fmt"
	"log"
	"sync"

	"signmeup/internal/firebase"
	"signmeup/internal/models"

	firebaseAuth "firebase.google.com/go/auth"

	"cloud.google.com/go/firestore"
)

var Repository *FirebaseRepository

func init() {
	var err error
	Repository, err = NewFirebaseRepository()
	if err != nil {
		log.Panicf("Error creating repository: %v\n", err)
	}

	log.Printf("✅ Successfully created Firebase repository client")
}

type FirebaseRepository struct {
	authClient      *firebaseAuth.Client
	firestoreClient *firestore.Client

	profilesLock *sync.RWMutex
	profiles     map[string]*models.Profile
}

func NewFirebaseRepository() (*FirebaseRepository, error) {
	fr := &FirebaseRepository{
		profilesLock: &sync.RWMutex{},
		profiles:     make(map[string]*models.Profile),
	}

	authClient, err := firebase.App.Auth(firebase.Context)
	if err != nil {
		return nil, fmt.Errorf("Auth client error: %v\n", err)
	}
	fr.authClient = authClient

	firestoreClient, err := firebase.App.Firestore(firebase.Context)
	if err != nil {
		return nil, fmt.Errorf("Firestore client error: %v\n", err)
	}
	fr.firestoreClient = firestoreClient

	// Execute the listeners sequentially, in case later listeners need to utilize data fetched
	// by previous listeners
	initFns := []func(){fr.initializeUserProfilesListener}
	for _, initFn := range initFns {
		fmt.Println("Something was initialized!")
		initFn()
	}

	return fr, nil
}
