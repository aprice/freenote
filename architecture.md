# Architecture

## Package Overview
 - cmd: command main packages
   - freenote: freenote CLI tool
   - freenoted: freenote server
 - config: configuration
 - notes: note model and handling
 - page: pagination model and handling
 - rest: REST API handler and helpers
 - store: backing store handlers
 - users: user account model and handling
 - web: UI content files (HTML/CSS/JS)

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
content files are embedded using github.com/aprice/embed, and served directly by
the application.