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
func (rh *requestHandler) doSession(w http.ResponseWriter, r *http.Request) {
	defer stats.Measure("req", "session", r.Method)()
	var err error
	switch r.Method {
	case http.MethodOptions:
		rh.preflight(w, r, nil, http.MethodGet, http.MethodPost, http.MethodDelete)
		return
	case http.MethodGet:
		if rh.user.Access == users.LevelAnon {
			statusResponse(w, http.StatusUnauthorized)
			return
		}
		sendResponse(w, r, decorateUser(rh.user, true, true, rh.baseURI), http.StatusOK)
		return
	case http.MethodPost:
		var user users.User
		if username := strings.ToLower(r.FormValue("username")); username != "" {
			user, err = rh.db.UserStore().UserByName(username)
			if handleError(w, err) {
				return
			}
			var ok bool
			if ok, err = user.Password.Verify(r.FormValue("password")); !ok || err != nil {
				http.Error(w, "Authentication Failed", http.StatusUnauthorized)
				return
			}
		} else {
			user, err = authenticate(w, r, rh.db.UserStore())
			if handleError(w, err) {
				return
			}
		}
		user.CleanSessions()
		sess, err := user.NewSession()
		if handleError(w, err) {
			return
		}
		if err = rh.db.UserStore().SaveUser(&user); handleError(w, err) {
			return
		}
		writeSessionCookie(w, sess)
		sendResponse(w, r, decorateUser(user, true, true, rh.baseURI), http.StatusOK)
	case http.MethodDelete:
		if rh.user.Access == users.LevelAnon {
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
		if payload.UserID != rh.user.ID {
			statusResponse(w, http.StatusForbidden)
			return
		}
		deleteSessionCookie(w)
		rh.user.CleanSessions()
		idx := -1
		for i, v := range rh.user.Sessions {
			if v.ID == payload.SessionID {
				idx = i
				break
			}
		}
		if idx >= 0 {
			rh.user.Sessions = append(rh.user.Sessions[:idx], rh.user.Sessions[idx:]...)
			if err := rh.db.UserStore().SaveUser(&rh.user); handleError(w, err) {
				return
			}
		}
	default:
		w.Header().Add("Allow", "GET, POST, DELETE")
		statusResponse(w, http.StatusMethodNotAllowed)
	}
}

// users/?.*
func (rh *requestHandler) doUsers(w http.ResponseWriter, r *http.Request) {
	if len(rh.path) > 1 {
		rh.doUser(w, r)
		return
	}
	defer stats.Measure("req", "users", r.Method)()
	switch r.Method {
	case http.MethodOptions:
		rh.preflight(w, r, nil, http.MethodGet, http.MethodPost)
		return
	case http.MethodGet:
		if rh.user.Access < users.LevelAdmin {
			handleError(w, errUnauthorized)
			return
		}
		pageReq := page.Page{
			Length: 10,
			SortBy: "username",
		}
		pageReq.FromQueryString(r.URL, []string{"username", "displayname"})
		pageRes, total, err := rh.db.UserStore().Users(pageReq)
		if handleError(w, err) {
			return
		}
		pageReq.HasMore = total > (pageReq.Start + pageReq.Length)
		sendResponse(w, r, decorateUsers(pageRes, pageReq, rh.user.Access >= users.LevelAdmin, rh.baseURI), http.StatusOK)
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
		newUser.Username = strings.ToLower(newUser.Username)
		if err = users.ValidateUsername(newUser.Username); badRequest(w, err) {
			return
		}
		if _, err = rh.db.UserStore().UserByName(newUser.Username); err == nil {
			badRequest(w, errors.New("username already in use"))
			return
		}
		if err = rh.db.UserStore().SaveUser(&newUser); handleError(w, err) {
			log.Println("error saving user: ", err)
			return
		}
		wn := notes.WelcomeNote(newUser.ID)
		err = rh.db.NoteStore().SaveNote(&wn)
		if err != nil {
			log.Println("saving welcome note failed: ", err)
		}
		pl := struct {
			decoratedUser
			Password string
		}{
			decorateUser(newUser, true, true, rh.baseURI),
			pw,
		}
		w.Header().Add("Location", fmt.Sprintf("%s/users/%s", rh.baseURI, newUser.ID))
		sendResponse(w, r, pl, http.StatusCreated)
		return
	default:
		w.Header().Add("Allow", "GET, POST")
		statusResponse(w, http.StatusMethodNotAllowed)
	}
}

// users/(id|username)/?.*
func (rh *requestHandler) doUser(w http.ResponseWriter, r *http.Request) {
	idOrName := rh.popSegment()
	var (
		err     error
		owner   users.User
		ownerID uuid.UUID
	)
	if ownerID, err = ids.ParseID(idOrName); err == nil {
		owner, err = rh.db.UserStore().UserByID(ownerID)
	} else {
		owner, err = rh.db.UserStore().UserByName(idOrName)
	}
	if handleError(w, err) {
		return
	}
	rh.owner = owner
	nextHandler := rh.popSegment()
	if nextHandler == "password" {
		rh.doPassword(w, r)
		return
	} else if nextHandler == "notes" {
		rh.doNotes(w, r)
		return
	} else if len(nextHandler) > 1 {
		statusResponse(w, http.StatusNotFound)
		return
	}

	defer stats.Measure("req", "user", r.Method)()

	switch r.Method {
	case http.MethodOptions:
		rh.preflight(w, r, nil, http.MethodGet, http.MethodPut, http.MethodDelete)
		return
	case http.MethodGet:
		self := owner.ID == rh.user.ID
		sendResponse(w, r, decorateUser(owner, self, self, rh.baseURI), http.StatusOK)
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
		if err = rh.db.UserStore().SaveUser(updateUser); handleError(w, err) {
			return
		}
		w.Header().Add("Location", fmt.Sprintf("%s/users/%s", rh.baseURI, updateUser.ID))
		sendResponse(w, r, decorateUser(*updateUser, true, true, rh.baseURI), http.StatusOK)
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
func (rh *requestHandler) doPassword(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodOptions:
		rh.preflight(w, r, nil, http.MethodPut)
		return
	case http.MethodPut:
		var err error
		if rh.owner.ID != rh.user.ID && rh.user.Access < users.LevelAdmin {
			statusResponse(w, http.StatusForbidden)
			return
		}
		pwr := struct{ Password string }{}
		if r.Header.Get("Content-Type") == "text/plain" {
			var body []byte
			body, err = ioutil.ReadAll(r.Body)
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
		if rh.owner.Password, err = users.NewPassword(pwr.Password); handleError(w, err) {
			return
		}
		sess, err := rh.owner.NewSession()
		if handleError(w, err) {
			return
		}
		writeSessionCookie(w, sess)
		if err = rh.db.UserStore().SaveUser(&rh.owner); handleError(w, err) {
			return
		}
		w.WriteHeader(http.StatusOK)
	default:
		w.Header().Add("Allow", http.MethodPut)
		statusResponse(w, http.StatusMethodNotAllowed)
	}
}

// users/{id}/notes/?.*
func (rh *requestHandler) doNotes(w http.ResponseWriter, r *http.Request) {
	if len(rh.path) > 1 {
		rh.doNote(w, r)
		return
	}
	var err error
	folderPath := r.URL.Query().Get("folder")
	defer stats.Measure("req", "notes", r.Method)()
	switch r.Method {
	case http.MethodOptions:
		rh.preflight(w, r, nil, http.MethodGet, http.MethodPost)
		return
	case http.MethodGet:
		if !authorizeUser(rh.user, rh.owner) {
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
		q := store.NoteQuery{
			Owner:  rh.owner.ID,
			Page:   pageReq,
			Folder: folderPath,
		}
		list, total, err = rh.db.NoteStore().QueryNotes(q)
		if handleError(w, err) {
			return
		}
		pageReq.HasMore = total > (pageReq.Start + pageReq.Length)
		sendResponse(w, r, decorateNotes(rh.owner, list, folderPath, pageReq, authorizeUser(rh.user, rh.owner), rh.baseURI), http.StatusOK)
	case http.MethodPost:
		note := new(notes.Note)
		var err error
		if err = parseRequest(r, note); badRequest(w, err) {
			return
		}
		note.ID = uuid.NewV4()
		note.Owner = rh.owner.ID
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
		ensureMarkdownBody(note, rh.sanitizer)
		if err = rh.db.NoteStore().SaveNote(note); handleError(w, err) {
			return
		}
		ensureHTMLBody(note, rh.sanitizer)
		w.Header().Add("Location", fmt.Sprintf("%s/users/%s/notes/%s", rh.baseURI, rh.owner.ID, note.ID))
		sendResponse(w, r, decorateNote(*note, true, rh.baseURI), http.StatusCreated)
	default:
		w.Header().Add("Allow", "GET, POST")
		statusResponse(w, http.StatusMethodNotAllowed)
	}
}

// users/{id}/notes/{id}
func (rh *requestHandler) doNote(w http.ResponseWriter, r *http.Request) {
	var (
		noteID uuid.UUID
		err    error
		note   notes.Note
	)
	if noteID, err = ids.ParseID(rh.popSegment()); badRequest(w, err) {
		return
	}
	if note, err = rh.db.NoteStore().NoteByID(noteID); handleError(w, err) {
		return
	}
	if note.Owner != rh.owner.ID {
		http.NotFound(w, r)
		return
	}
	// If there were anything chained after this, we'd add note to the context.
	defer stats.Measure("req", "note", r.Method)()
	switch r.Method {
	case http.MethodOptions:
		rh.preflight(w, r, nil, http.MethodGet, http.MethodPut, http.MethodDelete)
		return
	case http.MethodGet:
		//TODO: Sharing
		if !authorizeNote(rh.user, note) {
			statusResponse(w, http.StatusForbidden)
			return
		}
		ensureHTMLBody(&note, rh.sanitizer)
		//TODO: text/markdown, text/plain Accept support & front matter addition
		sendResponse(w, r, decorateNote(note, authorizeNoteWrite(rh.user, note), rh.baseURI), http.StatusOK)
	case http.MethodPut:
		//TODO: Conflict checking (etag, modified, etc)
		//TODO: Sharing
		//TODO: text/markdown, text/plain, text/html Content-Type support & front matter parsing
		if !authorizeUser(rh.user, rh.owner) {
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
		ensureMarkdownBody(note, rh.sanitizer)
		if err = rh.db.NoteStore().SaveNote(note); handleError(w, err) {
			return
		}
		ensureHTMLBody(note, rh.sanitizer)
		sendResponse(w, r, decorateNote(*note, authorizeNoteWrite(rh.user, *note), rh.baseURI), http.StatusOK)
		return
	case http.MethodDelete:
		if !authorizeNote(rh.user, note) {
			statusResponse(w, http.StatusForbidden)
			return
		}
		if err := rh.db.NoteStore().DeleteNote(noteID); handleError(w, err) {
			return
		}
		statusResponse(w, http.StatusNoContent)
	default:
		w.Header().Add("Allow", "GET, PUT, DELETE")
		statusResponse(w, http.StatusMethodNotAllowed)
	}
}
