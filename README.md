# Hours backend
This is the Go backend 


## File structure
```
├── internal
│   ├── api           // package containing code related to authentication
│   │   ├── auth.go    // contains public functions for doing auth-related things (such as creating users)
│   │   ├── errors.go  // contains definitions for possible errors
│   │   └── models.go  // models relating to authentication
│   │   └── repository.go  // encapsulates the logic to access users from a database
│   │   └── routes.go     // HTTP route handlers relating to authentication
│   └── firebase
│   │   └── firebase.go   // contains a global variable pointing to an initialized firebase app
│   └── server
│       └── server.go     // contains the http server. routes exported from a package should be mounted here.
└── main.go
```

## remarks
- users without a valid @brown.edu email are automatically deleted from Firebase authentication.
- the first registered user will automatically be made admin.
- ticket statuses are mapped to ints:
  - 0: waiting
  - 1: claimed
  - 2: missing
  - 3: complete
- 
## todo
- lots of refactoring to do.
- firestore collecction listners need concurrency things