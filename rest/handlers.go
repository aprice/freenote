package rest

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	uuid "github.com/satori/go.uuid"

	"github.com/aprice/freenote/ids"
	"github.com/aprice/freenote/notes"
	"github.com/aprice/freenote/page"
	"github.com/aprice/freenote/stats"
	"github.com/aprice/freenote/store"
	"github.com/aprice/freenote/users"
)

// session
func (s *Server) doSession(w http.ResponseWriter, r *http.Request) {
	defer stats.Measure("req", "session", r.Method)()
	var err error
	db, _ := store.FromContext(r.Context())
	switch r.Method {
	case http.MethodOptions:
		w.Header().Add("Allow", "GET, POST, DELETE")
		w.WriteHeader(http.StatusNoContent)
	case http.MethodGet:
		user, loggedIn := users.FromContext(r.Context())
		if !loggedIn || user.ID == uuid.Nil {
			statusResponse(w, http.StatusUnauthorized)
			return
		}
		sendResponse(w, r, decorateUser(user, true, true, s.conf), http.StatusOK)
		return
	case http.MethodPost:
		var user users.User
		if username := r.FormValue("username"); username != "" {
			user, err = db.UserStore().UserByName(username)
			if handleError(w, err) {
				return
			}
			if ok, err := user.Password.Verify(r.FormValue("password")); !ok || err != nil {
				http.Error(w, "Authentication Failed", http.StatusUnauthorized)
				return
			}
		} else {
			user, err = s.authenticate(w, r, db.UserStore())
			if handleError(w, err) {
				return
			}
		}
		user.CleanSessions()
		sess, err := user.NewSession()
		if handleError(w, err) {
			return
		}
		if err = db.UserStore().SaveUser(&user); handleError(w, err) {
			return
		}
		writeSessionCookie(w, sess)
		sendResponse(w, r, decorateUser(user, true, true, s.conf), http.StatusOK)
	case http.MethodDelete:
		user, loggedIn := users.FromContext(r.Context())
		if !loggedIn {
			handleError(w, errNoAuth)
			return
		}
		var payload struct {
			UserID    uuid.UUID
			SessionID uuid.UUID
		}
		if r.ContentLength == 0 {
			sess, err := parseSessionCookie(r)
			if badRequest(w, err) {
				return
			}
			payload.SessionID = sess.ID
			payload.UserID = sess.UserID
		} else if err := parseRequest(r, payload); badRequest(w, err) {
			return
		}
		if payload.UserID != user.ID {
			statusResponse(w, http.StatusForbidden)
			return
		}
		deleteSessionCookie(w)
		user.CleanSessions()
		idx := -1
		for i, v := range user.Sessions {
			if v.ID == payload.SessionID {
				idx = i
				break
			}
		}
		if idx >= 0 {
			user.Sessions = append(user.Sessions[:idx], user.Sessions[idx:]...)
			if err := db.UserStore().SaveUser(&user); handleError(w, err) {
				return
			}
		}
	default:
		w.Header().Add("Allow", "GET, POST, DELETE")
		statusResponse(w, http.StatusMethodNotAllowed)
	}
}

// users/?.*
func (s *Server) doUsers(w http.ResponseWriter, r *http.Request) {
	if len(r.URL.Path) > 1 {
		s.doUser(w, r)
		return
	}
	defer stats.Measure("req", "users", r.Method)()
	db, _ := store.FromContext(r.Context())
	user, _ := users.FromContext(r.Context())
	switch r.Method {
	case http.MethodOptions:
		w.Header().Add("Allow", "GET, POST")
		w.WriteHeader(http.StatusNoContent)
	case http.MethodGet:
		if user.Access < users.LevelAdmin {
			handleError(w, errUnauthorized)
			return
		}
		pageReq := page.Page{
			Length: 10,
			SortBy: "username",
		}
		pageReq.FromQueryString(r.URL, []string{"username", "displayname"})
		pageRes, total, err := db.UserStore().Users(pageReq)
		if handleError(w, err) {
			return
		}
		pageReq.HasMore = total > (pageReq.Start + pageReq.Length)
		sendResponse(w, r, decorateUsers(pageRes, pageReq, user.Access >= users.LevelAdmin, s.conf), http.StatusOK)
	case http.MethodPost:
		//TODO: User creation controls
		//TODO: New user verification
		var newUser users.User
		var err error
		if err = parseRequest(r, &newUser); badRequest(w, err) {
			return
		}
		newUser.ID = uuid.NewV4()
		var pw string
		pw, newUser.Password, err = users.RandomPassword(12)
		if handleError(w, err) {
			return
		}
		if err = users.ValidateUsername(newUser.Username); badRequest(w, err) {
			return
		}
		if _, err := db.UserStore().UserByName(newUser.Username); err == nil {
			badRequest(w, errors.New("username already in use"))
			return
		}
		if err = db.UserStore().SaveUser(&newUser); handleError(w, err) {
			log.Println("error saving user")
			return
		}
		wn := notes.WelcomeNote(newUser.ID)
		err = db.NoteStore().SaveNote(&wn)
		if err != nil {
			log.Println("Saving welcome note failed: ", err)
		}
		pl := struct {
			decoratedUser
			Password string
		}{
			decorateUser(newUser, true, true, s.conf),
			pw,
		}
		w.Header().Add("Location", fmt.Sprintf("%s/users/%s", s.conf.BaseURI, newUser.ID))
		sendResponse(w, r, pl, http.StatusCreated)
		return
	default:
		w.Header().Add("Allow", "GET, POST")
		statusResponse(w, http.StatusMethodNotAllowed)
	}
}

// users/(id|username)/?.*
func (s *Server) doUser(w http.ResponseWriter, r *http.Request) {
	idOrName := popSegment(r)
	db, _ := store.FromContext(r.Context())
	user, _ := users.FromContext(r.Context())
	var (
		err     error
		owner   users.User
		ownerID uuid.UUID
	)
	if ownerID, err = ids.ParseID(idOrName); err == nil {
		owner, err = db.UserStore().UserByID(ownerID)
	} else {
		owner, err = db.UserStore().UserByName(idOrName)
	}
	if handleError(w, err) {
		return
	}
	r = r.WithContext(users.NewOwnerContext(r.Context(), owner))
	nextHandler := popSegment(r)
	if nextHandler == "password" {
		s.doPassword(w, r)
		return
	} else if nextHandler == "notes" {
		s.doNotes(w, r)
		return
	} else if len(nextHandler) > 1 {
		statusResponse(w, http.StatusNotFound)
		return
	}

	defer stats.Measure("req", "user", r.Method)()

	switch r.Method {
	case http.MethodOptions:
		w.Header().Add("Allow", "GET, PUT, DELETE")
		w.WriteHeader(http.StatusNoContent)
	case http.MethodGet:
		self := owner.ID == user.ID
		sendResponse(w, r, decorateUser(owner, self, self, s.conf), http.StatusOK)
	case http.MethodPut:
		//TODO: Conflict checking (etag, modified, etc)
		updateUser := new(users.User)
		var err error
		if err = parseRequest(r, updateUser); badRequest(w, err) {
			return
		}
		// No fuckery allowed
		if updateUser.ID != owner.ID {
			http.Error(w, "Bad Request: cant't change user ID", http.StatusBadRequest)
			return
		}
		if updateUser.Username != owner.Username {
			http.Error(w, "Bad Request: can't change username", http.StatusBadRequest)
			return
		}
		// Password change is via a different route
		updateUser.Password = owner.Password
		updateUser.Sessions = owner.Sessions
		if err = db.UserStore().SaveUser(updateUser); handleError(w, err) {
			return
		}
		w.Header().Add("Location", fmt.Sprintf("%s/users/%s", s.conf.BaseURI, updateUser.ID))
		sendResponse(w, r, decorateUser(*updateUser, true, true, s.conf), http.StatusOK)
		return
	case http.MethodDelete:
		//TODO: Delete user & all notes
		statusResponse(w, http.StatusNotImplemented)
	default:
		w.Header().Add("Allow", "GET, PUT, DELETE")
		statusResponse(w, http.StatusMethodNotAllowed)
	}
}

// users/{id}/password
func (s *Server) doPassword(w http.ResponseWriter, r *http.Request) {
	db, _ := store.FromContext(r.Context())
	user, _ := users.FromContext(r.Context())
	owner, _ := users.OwnerFromContext(r.Context())
	switch r.Method {
	case http.MethodOptions:
		w.Header().Add("Allow", http.MethodPut)
		w.WriteHeader(http.StatusNoContent)
	case http.MethodPut:
		var err error
		if owner.ID != user.ID && user.Access < users.LevelAdmin {
			statusResponse(w, http.StatusForbidden)
			return
		}
		pwr := struct{ Password string }{}
		if r.Header.Get("Content-Type") == "text/plain" {
			body, err := ioutil.ReadAll(r.Body)
			if handleError(w, err) {
				return
			}
			pwr.Password = string(body)
		} else if err = parseRequest(r, &pwr); badRequest(w, err) {
			return
		}
		if err = users.ValidatePassword(pwr.Password); badRequest(w, err) {
			return
		}
		if owner.Password, err = users.NewPassword(pwr.Password); handleError(w, err) {
			return
		}
		sess, err := owner.NewSession()
		if handleError(w, err) {
			return
		}
		writeSessionCookie(w, sess)
		if err = db.UserStore().SaveUser(&owner); handleError(w, err) {
			return
		}
		w.WriteHeader(http.StatusOK)
	default:
		w.Header().Add("Allow", http.MethodPut)
		statusResponse(w, http.StatusMethodNotAllowed)
	}
}

// users/{id}/notes/?.*
func (s *Server) doNotes(w http.ResponseWriter, r *http.Request) {
	if len(r.URL.Path) > 1 {
		s.doNote(w, r)
		return
	}
	var err error
	db, _ := store.FromContext(r.Context())
	user, _ := users.FromContext(r.Context())
	owner, _ := users.OwnerFromContext(r.Context())
	folderPath := r.URL.Query().Get("folder")
	defer stats.Measure("req", "notes", r.Method)()
	switch r.Method {
	case http.MethodOptions:
		w.Header().Add("Allow", "GET, POST")
		w.WriteHeader(http.StatusNoContent)
	case http.MethodGet:
		if owner.ID != user.ID && user.Access < users.LevelAdmin {
			statusResponse(w, http.StatusForbidden)
			return
		}
		//TODO: Option to list folders instead of notes
		//TODO: Option to filter by tag
		//TODO: Full text search
		var (
			list  []notes.Note
			total int
		)
		pageReq := page.Page{
			Length:         10,
			SortBy:         "modified",
			SortDescending: true,
		}
		pageReq.FromQueryString(r.URL, []string{"modified", "title", "created"})
		if folderPath == "" {
			list, total, err = db.NoteStore().NotesByOwner(owner.ID, pageReq)
		} else {
			list, total, err = db.NoteStore().NotesByFolder(owner.ID, folderPath, pageReq)
		}
		if handleError(w, err) {
			return
		}
		pageReq.HasMore = total > (pageReq.Start + pageReq.Length)
		sendResponse(w, r, decorateNotes(owner, list, folderPath, pageReq, owner.ID == user.ID, s.conf), http.StatusOK)
	case http.MethodPost:
		note := new(notes.Note)
		var err error
		if err = parseRequest(r, note); badRequest(w, err) {
			return
		}
		note.ID = uuid.NewV4()
		note.Owner = owner.ID
		if folderPath != "" {
			note.Folder = folderPath
		}
		if note.Folder != "" {
			parts := strings.Split(note.Folder, "/")
			if _, err = uuid.FromString(parts[0]); err == nil {
				http.Error(w, "Bad Request: root folder cannot be UUID", http.StatusBadRequest)
				return
			}
		}
		ensureMarkdownBody(note, s.sanitizer)
		if err = db.NoteStore().SaveNote(note); handleError(w, err) {
			return
		}
		ensureHTMLBody(note, s.sanitizer)
		w.Header().Add("Location", fmt.Sprintf("%s/users/%s/notes/%s", s.conf.BaseURI, owner.ID, note.ID))
		sendResponse(w, r, decorateNote(*note, true, s.conf), http.StatusCreated)
	default:
		w.Header().Add("Allow", "GET, POST")
		statusResponse(w, http.StatusMethodNotAllowed)
	}
}

// users/{id}/notes/{id}
func (s *Server) doNote(w http.ResponseWriter, r *http.Request) {
	db, _ := store.FromContext(r.Context())
	user, _ := users.FromContext(r.Context())
	owner, _ := users.OwnerFromContext(r.Context())
	var (
		noteID uuid.UUID
		err    error
		note   notes.Note
	)
	if noteID, err = ids.ParseID(popSegment(r)); badRequest(w, err) {
		return
	}
	if note, err = db.NoteStore().NoteByID(noteID); handleError(w, err) {
		return
	}
	// If there were anything chained after this, we'd add note to the context.
	defer stats.Measure("req", "note", r.Method)()
	switch r.Method {
	case http.MethodOptions:
		w.Header().Add("Allow", "GET, PUT, DELETE")
		w.WriteHeader(http.StatusNoContent)
	case http.MethodGet:
		//TODO: Sharing
		if owner.ID != user.ID && user.Access < users.LevelAdmin {
			statusResponse(w, http.StatusForbidden)
			return
		}
		ensureHTMLBody(&note, s.sanitizer)
		//TODO: text/markdown, text/plain Accept support & front matter addition
		sendResponse(w, r, decorateNote(note, note.Owner == user.ID, s.conf), http.StatusOK)
	case http.MethodPut:
		//TODO: Conflict checking (etag, modified, etc)
		//TODO: Sharing
		//TODO: text/markdown, text/plain, text/html Content-Type support & front matter parsing
		if owner.ID != user.ID && user.Access < users.LevelAdmin {
			statusResponse(w, http.StatusForbidden)
			return
		}
		note := new(notes.Note)
		var err error
		if err = parseRequest(r, note); badRequest(w, err) {
			return
		}
		if note.ID != note.ID {
			// No fuckery allowed
			http.Error(w, "Bad Request: cant't change ID", http.StatusBadRequest)
			return
		}
		ensureMarkdownBody(note, s.sanitizer)
		if err = db.NoteStore().SaveNote(note); handleError(w, err) {
			return
		}
		ensureHTMLBody(note, s.sanitizer)
		sendResponse(w, r, decorateNote(*note, note.Owner == user.ID, s.conf), http.StatusOK)
		return
	case http.MethodDelete:
		if owner.ID != user.ID && user.Access < users.LevelAdmin {
			statusResponse(w, http.StatusForbidden)
			return
		}
		if err := db.NoteStore().DeleteNote(noteID); handleError(w, err) {
			return
		}
		statusResponse(w, http.StatusNoContent)
	default:
		w.Header().Add("Allow", "GET, PUT, DELETE")
		statusResponse(w, http.StatusMethodNotAllowed)
	}
}
