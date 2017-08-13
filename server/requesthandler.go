package server

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	uuid "github.com/satori/go.uuid"

	"github.com/aprice/freenote"
	"github.com/aprice/freenote/config"
	"github.com/aprice/freenote/notes"
	"github.com/aprice/freenote/store"
	"github.com/aprice/freenote/users"
)

type requestHandler struct {
	conf      config.Config
	sanitizer *bluemonday.Policy
	baseURI   string
	path      string
	db        store.Session
	user      users.User
	owner     users.User
	note      notes.Note
}

func NewRequestHandler(r *http.Request, conf config.Config, sanitizer *bluemonday.Policy) (*requestHandler, error) {
	db, err := store.NewSession(conf)
	if err != nil {
		return nil, err
	}

	var baseURI string
	if conf.ForceTLS {
		baseURI = conf.BaseURI
	}
	u := strings.TrimPrefix(conf.BaseURI, "http")
	u = strings.TrimPrefix(u, "s")
	if p := r.Header.Get("X-Forwarded-Proto"); p != "" {
		baseURI = p + u
	}
	if r.TLS == nil {
		baseURI = "http" + u
	}
	baseURI = "https" + u

	rh := &requestHandler{
		conf:      conf,
		sanitizer: sanitizer,
		baseURI:   baseURI,
		path:      r.URL.Path,
		db:        db,
	}
	return rh, nil
}

func (rh *requestHandler) close() {
	if clo, ok := rh.db.(io.Closer); ok {
		clo.Close()
	}
}

func (rh *requestHandler) handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Vary", "Accept")
	if freenote.Version != "" {
		w.Header().Add("X-Freenote-Version", freenote.Version+"-"+freenote.Build)
	}
	var err error
	rh.user, err = authenticate(w, r, rh.db.UserStore())
	switch err {
	case errNoAuth, nil:
	case errAuthCookieInvalid:
		http.Error(w, "Unauthorized: Invalid Session Cookie", http.StatusUnauthorized)
		return
	default:
		handleError(w, err)
		return
	}
	if rh.user.ID != uuid.Nil && rand.Float64() < 0.01 {
		rh.user.CleanSessions()
		rh.db.UserStore().SaveUser(&rh.user)
	}
	if !authorize(r.URL.Path, rh.user) {
		if rh.user.Access == users.LevelAnon {
			statusResponse(w, http.StatusUnauthorized)
		} else {
			statusResponse(w, http.StatusForbidden)
		}
		return
	}

	first := rh.popSegment()
	switch first {
	case "users":
		rh.doUsers(w, r)
	case "session":
		rh.doSession(w, r)
	}
}

func (rh *requestHandler) preflight(w http.ResponseWriter, r *http.Request, body []byte, methods ...string) {
	methList := strings.ToUpper(strings.Join(methods, ", "))
	headers := w.Header()
	reqHeader := r.Header.Get("Access-Control-Request-Headers")
	if reqHeader == "" {
		reqHeader = "*"
	}
	var origin string
	if r.TLS == nil {
		origin = fmt.Sprintf("%s:%d", rh.conf.BaseURI, rh.conf.Port)
	} else {
		origin = fmt.Sprintf("%s:%d", rh.conf.BaseURI, rh.conf.TLSPort)
	}

	headers.Add("Vary", "Origin")
	headers.Add("Vary", "Access-Control-Request-Methods")
	headers.Add("Vary", "Access-Control-Request-Headers")

	headers.Set("Allow", methList)
	headers.Set("Access-Control-Allow-Origin", origin)
	headers.Set("Access-Control-Allow-Methods", methList)
	headers.Set("Access-Control-Allow-Headers", reqHeader)
	headers.Set("Access-Control-Allow-Credentials", "true")
	headers.Set("Access-Control-Max-Age", "86400")

	if len(body) == 0 {
		w.WriteHeader(http.StatusNoContent)
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}
}

func (rh *requestHandler) popSegment() string {
	path := strings.Trim(rh.path, "/")
	seg := path
	if idx := strings.Index(seg, "/"); idx > 0 {
		seg = seg[:idx]
		rh.path = path[len(seg):]
	} else {
		rh.path = "/"
	}
	return seg
}
