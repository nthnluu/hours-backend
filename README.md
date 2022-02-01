<img src="https://i.imgur.com/tNHmFW3.png" alt=""/>

<h1 align="center">Hours Backend</h1>

<div align="center">
 <b>
  The Go backend for Hours — a real-time office hour management system used at Brown University.
 </b>
</div>

## File structure
```
├── internal
│   ├── api   // TODO
│   └── auth
│   │   └── middleware.go   // middlewares and helpers for checking user authentication from request.
│   │   └── permissions.go    // middlewares for checking user permissions.
│   └── config    // application configuration
│   └── firebase    // defines global varibles with the initialized Firebase app and context.
│   └── models    // type definitions 
│   └── qerrors   // definitions for errors that can be sent back to the client.
│   └── repository    // encapsulates logic for accessing entities from Firestore.
│   └── router    // route definitions and handlers.
│   └── server    // the HTTP server.

```