package rest

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/acme/autocert"

	uuid "github.com/satori/go.uuid"

	"github.com/aprice/freenote/config"
	"github.com/aprice/freenote/notes"
	"github.com/aprice/freenote/store"
	"github.com/aprice/freenote/users"
	"github.com/aprice/freenote/web"
	"github.com/microcosm-cc/bluemonday"
)

type Server struct {
	conf          config.Config
	sessionCookie string
	fs            http.Handler
	sanitizer     *bluemonday.Policy
	svr           *http.Server
	tlsSvr        *http.Server
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

func NewServer(conf config.Config) (*Server, error) {
	s := &Server{
		conf:          conf,
		sessionCookie: "sess",
		fs:            web.GetEmbeddedContent(),
		sanitizer:     bluemonday.UGCPolicy(),
	}
	s.sanitizer.AllowAttrs("class").Matching(regexp.MustCompile("^language-[a-zA-Z0-9]+$")).OnElements("code")
	s.svr = &http.Server{
		Addr:    fmt.Sprintf(":%d", conf.Port),
		Handler: s,
	}
	if len(conf.LetsEncryptHosts) > 0 {
		log.Printf("Let's Encrypt! Hosts: %v", conf.LetsEncryptHosts)
		m := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(conf.LetsEncryptHosts...),
			Cache:      autocert.DirCache("/tmp/freenoted/acme"),
		}
		s.tlsSvr = &http.Server{
			Addr:      fmt.Sprintf(":%d", conf.TLSPort),
			Handler:   s,
			TLSConfig: &tls.Config{GetCertificate: m.GetCertificate},
		}
	} else if conf.CertFile != "" {
		log.Printf("Using provided SSL certificate: %s", conf.CertFile)
		cer, err := tls.LoadX509KeyPair(conf.CertFile, conf.KeyFile)
		if err != nil {
			return nil, err
		}
		s.tlsSvr = &http.Server{
			Addr:    fmt.Sprintf(":%d", conf.TLSPort),
			Handler: s,
			TLSConfig: &tls.Config{
				Certificates: []tls.Certificate{cer},
			},
		}
	}
	return s, nil
}

func (s *Server) Start() {
	go func() {
		log.Printf("http listening on %d", s.conf.Port)
		log.Println(s.svr.ListenAndServe())
	}()
	if s.tlsSvr != nil {
		go func() {
			log.Printf("https listening on %d", s.conf.TLSPort)
			log.Println(s.tlsSvr.ListenAndServeTLS("", ""))
		}()
	}
}

func (s *Server) Stop() {
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		ctx, cxl := context.WithTimeout(context.Background(), 5*time.Second)
		defer cxl()
		s.svr.Shutdown(ctx)
		wg.Done()
	}()
	if s.tlsSvr != nil {
		wg.Add(1)
		go func() {
			ctx, cxl := context.WithTimeout(context.Background(), 5*time.Second)
			cxl()
			s.tlsSvr.Shutdown(ctx)
			wg.Done()
		}()
	}
	wg.Wait()
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
	case "session":
		s.doSession(rc, w, r)
	case "debug":
		doDebug(w, r)
	default:
		s.fs.ServeHTTP(w, r)
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
