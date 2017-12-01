# Architecture

## Package Overview

 - `cmd`: command main packages
   - `freenote`: freenote CLI tool
   - `freenoted`: freenote server
 - `config`: configuration
 - `ids`: ID helper functions
 - `notes`: note model and handling
 - `page`: pagination model and handling
 - `rest`: REST API handler and helpers
 - `stats`: stats measurement for expvar
 - `store`: backing store handlers
 - `users`: user account model and handling
 - `web`: UI content files (HTML/CSS/JS)

## REST API Handler

Becuase the REST API is pretty straightforward, we don't use any third-party
mux; in fact, we don't use any mux at all. We just use a cascading handler,
wherein /users goes to the users collection handler; that looks for an ID,
and if it finds one, passes control to the user handler; that looks for one
of the user subroutes, and if it finds one, passes control to the appropriate
handler; and so forth. The mux is entirely hard-coded, and therefore highly
efficient. Were the API to get much more complicated, we'd likely want to move
to a proper mux, as performance is not top priority for this server.

`ServeHTTP` handles requests to begin with, establishing DB connectivity,
authenticating the user, authorizing the route, then routing the request. It
bundles up the request details into a `requestContext` instance, which just
holds request-specific variables to avoid unnecessary re-parsing or excessive
argument counts in the handler methods. This is used in favor of Request.Context
for type safety.

auth.go contains handlers for authentication, authorization, and session handling.
The only session data used is a session token to maintain authentication for web
clients.

contenttype.go contains handlers for parsing requests of arbitrary content types
and marshalling responses in arbitrary content types, based on the Content-Type
and Accept headers, respectively.

hypermedia.go handles decorating response objects with link collections prior to
marshaling, and includes helpers for generating the most common link relations.

notes.go and users.go contain helpers for decorating users and notes, respectively.

debug.go contains a stub handler for `/debug` that always returns 404 and only
runs if neither the `debug` nor `dev` build tags are supplied. If either is
supplied, debug_dev.go will run instead, which routes handlers for expvar and
pprof.

## Backing Store

Only two types of data are in the backing store, users and notes. These can be
stored in the same database, or different databases. A backing store driver must
fulfull the interfaces defined in store.go.

There are currently two backing stores implemented, an embedded database using
BoltDB via Storm, and an external database using MongoDB. Further stores are
planned for future versions.

## User Account Handling

Authentication is handled by the `Password` type. Rather than including credentials
directly in the `User`, all the details of password management are segregated out.
The `Password` struct contains the password version, salt, and hash. The version
indicates what hash/salt implementation was used, allowing the security to be
updated at will while maintaining backward compatibility.

Authorization is handled by the simple AccessLevel enumeration, with a series of
users levels, each with greater access than the one below it, allowing for access
control by greater-than/less-than comparison against the constants. This is used
primarily for route pattern authorization, and for administrative access (i.e. the
ability to manage data owned by another user).

Lastly there is a recovery mode, whereby the system can generate a random password
to an account with full administrative access, which is available for a limited
time after the application starts. The password is printed in the log. This can be
used for new installations (when no other administrative users exist yet) or for
system recovery, when no administrative users are able to log in.

## Web UI and Embedded Content

The web UI is a static HTML/CSS/JS site which interacts with the REST API. The
content files are embedded using `github.com/aprice/embed`, and served directly by
the application. Given the `dev` build tag, static files will be served from disk,
assuming the working directory is the `github.com/aprice/freenote` root.

### Architectural Decisions

The UI uses no external libraries, frameworks, or polyfills. Built-in JavaScript
functionality in modern browsers is not overly complex and provides all the
needed functionality for the UI.

With all modern browsers auto-updating, we're free to target relatively recent
versions of all mainstream browsers on desktop and mobile. Modern browsers are
getting much better at standards compliance and cross-browser compatibility,
which means we get the advantage of modern features. So the basic rule of thumb
for any core functionality is any feature must be supported by Firefox, Chrome,
and Safari versions at least 6 months old, without vendor prefixes. Any less
broadly supported feature must degrade completely transparently when not
available.

The UI design is fully responsive and adaptive based on state (identified by
classes) and media. It's progressive, supporting offline access from IndexedDB
and Web Cache, and uses a Service Worker to manage content caching.

The UI is relatively slim, coming in at under 250kB before minification and
compression (under 200kB after), allowing for load times on the order of 500ms.

### UI Architecture

The UI is comprised of a few main components:

- `window.App` is the main application manager, which handles showing messages,
errors, and confirmation panels, as well as the application startup process. It
is defined in `app.js`.
- `window.User` handles authentication and user (password) editing. It is
defined in `auth.js`.
- `window.NoteList` maintains the note list, selecting notes, and creating
notes. It is defined in `notes.js`.
- `window.NoteEditor` maints the note editor, saving notes, and deleting notes.
It is defined in `notes.js`.
- `window.DataStore` is an opaque reference to `IndexedDBDataStore`, or
`RESTDataStore` if IndexedDB is not available on the client. It is set in `db.js`.
- `IndexedDBDataStore` is a note store that acts as a proxy to `RESTDataStore`,
using IndexedDB for caching and offline access. It is defined in `db.js`.
- `RESTDataStore` is a note store backed by the REST API. It is defined in
`db.js`.
- A service worker is defined in `worker.js`, responsible for caching static
content for offline use.

`utils.js` includes a handful of helper and utility functions, primarily
shortcuts to frequently-used calls.

The application starts with a call to `App.init()` from the window's `onload`
handler. This then calls `User.init()`, `NoteList.init()`, and
`NoteEditor.init()`. `User.init()` loads the current user from `localStorage`
if available, and either way, refreshes the current session from the server.
`NoteList.init()` establishes handlers and loads the note list from the data
store. `NoteEditor.init()` just establishes event handlers.

When a note or note list is retrieved from the data store, assuming IndexedDB
is available, it first pulls from the local database to serve the request
quickly. It then makes a REST API call in the background, and when that call
returns, it fires a `notefetched` (if a single note was retrieved in full) or
`summaryfetched` (if a list of notes were retrieved with summaries) event on
the window; these events are handled by `NoteList` and `NoteEditor` and used to
refresh the display when fresh data is received from the server.

When a note is saved, it is put in IndexedDB with a `pending` flag; if it is a
new note, it is also given a temporary ID. When a note is deleted, it is updated
in IndexedDB with a `delete` flag. At regular intervals, the local database is
synced with the remote database, by:

- Sending any `pending` notes
- Deleting any `delete` notes
- Retrieving any notes with a modification time newer than the most recent
currently in the local database

This allows us to continue operating while offline and sync up when connectivity
is restored. It also allows for immediate responsiveness to user actions.