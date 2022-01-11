package firebase

import (
	"context"
	firebaseSDK "firebase.google.com/go"
	"google.golang.org/api/option"
)

// FirebaseApp is a global variable to hold the initialized Firebase App object
var FirebaseApp *firebaseSDK.App
var FirebaseContext context.Context

func initializeFirebaseApp() {
	ctx := context.Background()
	opt := option.WithCredentialsFile("signmeup-46c73-firebase-adminsdk-9mll6-d60aa0677c.json")
	app, err := firebaseSDK.NewApp(ctx, nil, opt)
	if err != nil {
		panic(err.Error())
	}

	FirebaseApp = app
	FirebaseContext = ctx
}

func init() {
	initializeFirebaseApp()
}
