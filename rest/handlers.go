package rest

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/russross/blackfriday"
	uuid "github.com/satori/go.uuid"

	"github.com/aprice/freenote/notes"
	"github.com/aprice/freenote/page"
	"github.com/aprice/freenote/users"
	"github.com/lunny/html2md"
)

// session
func (s *Server) doSession(rc requestContext, w http.ResponseWriter, r *http.Request) {
	var err error
	switch r.Method {
	case http.MethodOptions:
		w.Header().Add("Allow", "GET, POST, DELETE")
		w.WriteHeader(http.StatusNoContent)
	case http.MethodGet:
		if rc.user.ID == uuid.Nil {
			statusResponse(w, http.StatusUnauthorized)
			return
		}
		sendResponse(w, r, decorateUser(rc.user, true, true, s.conf), http.StatusOK)
		return
	case http.MethodPost:
		var user users.User
		if username := r.FormValue("username"); username != "" {
			user, err = rc.db.UserStore().UserByName(username)
			if handleError(w, err) {
				return
			}
			if ok, err := user.Password.Verify(r.FormValue("password")); !ok || err != nil {
				http.Error(w, "Authentication Failed", http.StatusUnauthorized)
				return
			}
		} else {
			user, err = s.authenticate(w, r, rc.db.UserStore())
			if handleError(w, err) {
				return
			}
		}
		rc.user.CleanSessions()
		sessID := user.NewSession()
		if err = rc.db.UserStore().SaveUser(&user); handleError(w, err) {
			return
		}
		writeSessionCookie(w, user.ID, sessID)
		sendResponse(w, r, decorateUser(user, true, true, s.conf), http.StatusOK)
	case http.MethodDelete:
		var payload struct {
			UserID    uuid.UUID
			SessionID uuid.UUID
		}
		if r.ContentLength == 0 {
			payload.UserID, payload.SessionID, err = parseSessionCookie(r)
			if badRequest(w, err) {
				return
			}
		} else if err := parseRequest(r, payload); badRequest(w, err) {
			return
		}
		if payload.UserID != rc.user.ID {
			statusResponse(w, http.StatusForbidden)
			return
		}
		deleteSessionCookie(w)
		rc.user.CleanSessions()
		idx := -1
		for i, v := range rc.user.Sessions {
			if v.ID == payload.SessionID {
				idx = i
				break
			}
		}
		if idx >= 0 {
			rc.user.Sessions = append(rc.user.Sessions[:idx], rc.user.Sessions[idx:]...)
			if err := rc.db.UserStore().SaveUser(&rc.user); handleError(w, err) {
				return
			}
		}
	default:
		w.Header().Add("Allow", "GET, POST, DELETE")
		statusResponse(w, http.StatusMethodNotAllowed)
	}
}

// users/?.*
func (s *Server) doUsers(rc requestContext, w http.ResponseWriter, r *http.Request) {
	if len(rc.path) > 1 {
		s.doUser(rc, w, r)
		return
	}
	switch r.Method {
	case http.MethodOptions:
		w.Header().Add("Allow", "GET, POST")
		w.WriteHeader(http.StatusNoContent)
	case http.MethodGet:
		pageReq := page.FromQueryString(r.URL, []string{"username", "displayname"})
		pageRes, err := rc.db.UserStore().Users(pageReq)
		if handleError(w, err) {
			return
		}
		sendResponse(w, r, decorateUsers(pageRes, pageReq, rc.user.Access >= users.LevelAdmin, s.conf), http.StatusOK)
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
		pw, newUser.Password, err = users.RandomPassword()
		if handleError(w, err) {
			return
		}
		//TODO: Give this to the user stead of the log
		log.Println(pw)
		if err = users.ValidateUsername(newUser.Username); badRequest(w, err) {
			return
		}
		if err = rc.db.UserStore().SaveUser(&newUser); handleError(w, err) {
			log.Println("error saving user")
			return
		}
		w.Header().Add("Location", fmt.Sprintf("%s/users/%s", s.conf.BaseURI, newUser.ID))
		sendResponse(w, r, decorateUser(newUser, true, true, s.conf), http.StatusCreated)
		return
	default:
		w.Header().Add("Allow", "GET, POST")
		statusResponse(w, http.StatusMethodNotAllowed)
	}
}

// users/(id|username)/?.*
func (s *Server) doUser(rc requestContext, w http.ResponseWriter, r *http.Request) {
	var err error
	if rc.ownerID, err = uuid.FromString(rc.pathSegment(1)); err == nil {
		rc.owner, err = rc.db.UserStore().UserByID(rc.ownerID)
	} else {
		rc.owner, err = rc.db.UserStore().UserByName(rc.pathSegment(1))
	}
	if handleError(w, err) {
		return
	}
	rc.ownerID = rc.owner.ID

	if rc.pathSegment(2) == "password" {
		s.doPassword(rc, w, r)
		return
	} else if rc.pathSegment(2) == "notes" {
		s.doNotes(rc, w, r)
		return
	} else if len(rc.path) > 2 {
		statusResponse(w, http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodOptions:
		w.Header().Add("Allow", "GET, PUT, DELETE")
		w.WriteHeader(http.StatusNoContent)
	case http.MethodGet:
		self := rc.ownerID == rc.user.ID
		sendResponse(w, r, decorateUser(rc.owner, self, self, s.conf), http.StatusOK)
	case http.MethodPut:
		//TODO: Conflict checking (etag, modified, etc)
		user := new(users.User)
		var err error
		if err = parseRequest(r, user); badRequest(w, err) {
			return
		}
		// No fuckery allowed
		if user.ID != rc.ownerID {
			http.Error(w, "Bad Request: cant't change user ID", http.StatusBadRequest)
			return
		}
		if user.Username != rc.owner.Username {
			http.Error(w, "Bad Request: can't change username", http.StatusBadRequest)
			return
		}
		// Password change is via a different route
		user.Password = rc.owner.Password
		if err = rc.db.UserStore().SaveUser(user); handleError(w, err) {
			return
		}
		w.Header().Add("Location", fmt.Sprintf("%s/users/%s", s.conf.BaseURI, user.ID))
		sendResponse(w, r, decorateUser(*user, true, true, s.conf), http.StatusOK)
		return
	case http.MethodDelete:
	//TODO: Delete user & all notes
	default:
		w.Header().Add("Allow", "GET, PUT, DELETE")
		statusResponse(w, http.StatusMethodNotAllowed)
	}
}

// users/{id}/password
func (s *Server) doPassword(rc requestContext, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodOptions:
		w.Header().Add("Allow", http.MethodPut)
		w.WriteHeader(http.StatusNoContent)
	case http.MethodPut:
		var err error
		if rc.ownerID != rc.user.ID && rc.user.Access < users.LevelAdmin {
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
		} else {
			if err = parseRequest(r, &pwr); badRequest(w, err) {
				return
			}
		}
		if err = users.ValidatePassword(pwr.Password); badRequest(w, err) {
			return
		}
		rc.owner.Password, err = users.NewPassword(pwr.Password)
		if handleError(w, err) {
			return
		}
		if err = rc.db.UserStore().SaveUser(&rc.owner); handleError(w, err) {
			return
		}
		w.WriteHeader(http.StatusOK)
	default:
		w.Header().Add("Allow", http.MethodPut)
		statusResponse(w, http.StatusMethodNotAllowed)
	}
}

// users/{id}/notes/?.*
func (s *Server) doNotes(rc requestContext, w http.ResponseWriter, r *http.Request) {
	var folderPath string
	var err error
	if rc.noteID, err = uuid.FromString(rc.pathSegment(3)); err == nil {
		rc.note, err = rc.db.NoteStore().NoteByID(rc.noteID)
		if handleError(w, err) {
			return
		}
		s.doNote(rc, w, r)
		return
	} else if len(rc.path) > 3 {
		folderPath = strings.Join(rc.path[3:], "/")
	}
	switch r.Method {
	case http.MethodOptions:
		w.Header().Add("Allow", "GET, POST")
		w.WriteHeader(http.StatusNoContent)
	case http.MethodGet:
		if rc.ownerID != rc.user.ID && rc.user.Access < users.LevelAdmin {
			statusResponse(w, http.StatusForbidden)
			return
		}
		//TODO: Option to list folders instead of notes
		//TODO: Option to filter by tag
		//TODO: Full text search
		var list []notes.Note
		pageReq := page.FromQueryString(r.URL, []string{"title", "created", "modified"})
		if folderPath == "" {
			list, err = rc.db.NoteStore().NotesByOwner(rc.ownerID, pageReq)
		} else {
			list, err = rc.db.NoteStore().NotesByFolder(rc.ownerID, folderPath, pageReq)
		}
		if handleError(w, err) {
			return
		}
		sendResponse(w, r, decorateNotes(rc.owner, list, folderPath, pageReq, true, s.conf), http.StatusOK)
	case http.MethodPost:
		note := new(notes.Note)
		var err error
		if err = parseRequest(r, note); badRequest(w, err) {
			return
		}
		note.ID = uuid.NewV4()
		note.Owner = rc.ownerID
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
		if err = rc.db.NoteStore().SaveNote(note); handleError(w, err) {
			return
		}
		w.Header().Add("Location", fmt.Sprintf("%s/users/%s/notes/%s", s.conf.BaseURI, rc.ownerID, note.ID))
		sendResponse(w, r, decorateNote(*note, true, s.conf), http.StatusCreated)
	default:
		w.Header().Add("Allow", "GET, POST")
		statusResponse(w, http.StatusMethodNotAllowed)
	}
}

// users/{id}/notes/{id}
func (s *Server) doNote(rc requestContext, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodOptions:
		w.Header().Add("Allow", "GET, PUT, DELETE")
		w.WriteHeader(http.StatusNoContent)
	case http.MethodGet:
		//TODO: Sharing
		if rc.ownerID != rc.user.ID && rc.user.Access < users.LevelAdmin {
			statusResponse(w, http.StatusForbidden)
			return
		}
		rc.note.HTMLBody = string(blackfriday.MarkdownCommon([]byte(rc.note.Body)))
		//TODO: text/markdown, text/plain Accept support & front matter addition
		sendResponse(w, r, decorateNote(rc.note, rc.note.Owner == rc.user.ID, s.conf), http.StatusOK)
	case http.MethodPut:
		//TODO: Conflict checking (etag, modified, etc)
		//TODO: Sharing
		//TODO: text/markdown, text/plain, text/html Content-Type support & front matter parsing
		note := new(notes.Note)
		var err error
		if err = parseRequest(r, note); badRequest(w, err) {
			return
		}
		if note.ID != rc.note.ID {
			// No fuckery allowed
			http.Error(w, "Bad Request: cant't change ID", http.StatusBadRequest)
			return
		}
		if note.Body == "" && note.HTMLBody != "" {
			note.Body = html2md.Convert(note.HTMLBody)
		}
		if err = rc.db.NoteStore().SaveNote(note); handleError(w, err) {
			return
		}
		if note.HTMLBody == "" && note.Body != "" {
			rc.note.HTMLBody = string(blackfriday.MarkdownCommon([]byte(rc.note.Body)))
		}
		sendResponse(w, r, decorateNote(*note, rc.note.Owner == rc.user.ID, s.conf), http.StatusOK)
		return
	case http.MethodDelete:
		if err := rc.db.NoteStore().DeleteNote(rc.noteID); handleError(w, err) {
			return
		}
		statusResponse(w, http.StatusNoContent)
	default:
		w.Header().Add("Allow", "GET, PUT, DELETE")
		statusResponse(w, http.StatusMethodNotAllowed)
	}
}
