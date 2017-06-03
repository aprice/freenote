package rest

import (
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"

	uuid "github.com/satori/go.uuid"

	"github.com/aprice/freenote/config"
	"github.com/aprice/freenote/notes"
	"github.com/aprice/freenote/store"
	"github.com/aprice/freenote/users"
	"github.com/aprice/freenote/web"
)

type Server struct {
	conf          config.Config
	sessionCookie string
	fs            http.Handler
}

type requestContext struct {
	db      store.Session
	user    users.User
	ownerID uuid.UUID
	owner   users.User
	noteID  uuid.UUID
	note    notes.Note
	path    []string
}

func (rc requestContext) pathSegment(idx int) string {
	if idx < 0 || idx > len(rc.path)-1 {
		return ""
	}
	return rc.path[idx]
}

func NewServer(conf config.Config) *Server {
	return &Server{
		conf:          conf,
		sessionCookie: "sess",
		fs:            web.GetEmbeddedContent(),
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	db, err := store.NewSession(s.conf)
	if err != nil {
		statusResponse(w, http.StatusInternalServerError)
		log.Println(err)
		return
	}
	if clo, ok := db.(io.Closer); ok {
		defer clo.Close()
	}
	user, err := s.authenticate(w, r, db.UserStore())
	switch err {
	case errNoAuth, nil:
	case errAuthCookieInvalid:
		statusResponse(w, http.StatusUnauthorized)
		return
	case users.ErrAuthenticationFailed:
		http.Error(w, "Authentication Failed", http.StatusUnauthorized)
		return
	default:
		handleError(w, err)
		return
	}
	if user.ID != uuid.Nil && rand.Float64() < 0.1 {
		user.CleanSessions()
		db.UserStore().SaveUser(&user)
	}
	ok, err := s.authorize(r, user)
	if handleError(w, err) {
		return
	}
	if !ok {
		if user.Access == users.LevelAnon {
			statusResponse(w, http.StatusUnauthorized)
		} else {
			statusResponse(w, http.StatusForbidden)
		}
		return
	}
	var pp []string
	path := strings.Trim(r.URL.Path, "/")
	path = strings.ToLower(path)
	if path == "" {
		pp = make([]string, 1)
	} else {
		pp = strings.Split(path, "/")
	}
	rc := requestContext{
		db:   db,
		user: user,
		path: pp,
	}
	switch pp[0] {
	case "users":
		s.doUsers(rc, w, r)
		return
	case "session":
		s.doSession(rc, w, r)
		return
	default:
		w.Header().Add("Pragma", "no-cache")
		s.fs.ServeHTTP(w, r)
		return
	}
}

// Handle an error; returns true if a response was sent, false otherwise
func handleError(w http.ResponseWriter, err error) bool {
	if err == store.ErrNotFound {
		statusResponse(w, http.StatusNotFound)
		return true
	}
	if err == users.ErrAuthenticationFailed {
		http.Error(w, "Authentication Failed", http.StatusUnauthorized)
		return true
	}
	if err != nil {
		log.Printf("%T: %[1]q", err)
		statusResponse(w, http.StatusInternalServerError)
		return true
	}
	return false
}

func statusResponse(w http.ResponseWriter, statusCode int) {
	http.Error(w, http.StatusText(statusCode), statusCode)
}

func badRequest(w http.ResponseWriter, err error) bool {
	if err == errUnsupportedMediaType {
		statusResponse(w, http.StatusUnsupportedMediaType)
		return true
	}
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest)+": "+err.Error(), http.StatusBadRequest)
		return true
	}
	return false
}
