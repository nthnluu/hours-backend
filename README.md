# queue backend
TODO(nthnluu): will write docs


## File structure
```
├── internal
│   ├── auth           // package containing code related to authentication
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