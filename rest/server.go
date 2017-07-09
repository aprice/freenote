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

	"github.com/microcosm-cc/bluemonday"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/acme/autocert"

	"github.com/aprice/freenote"
	"github.com/aprice/freenote/config"
	"github.com/aprice/freenote/store"
	"github.com/aprice/freenote/users"
	"github.com/aprice/freenote/web"
)

// Server handles HTTP requests.
type Server struct {
	conf      config.Config
	fs        http.Handler
	sanitizer *bluemonday.Policy
	svr       *http.Server
	tlsSvr    *http.Server
}

// NewServer creates a new HTTP Server with the given configuration.
func NewServer(conf config.Config) (*Server, error) {
	s := &Server{
		conf:      conf,
		fs:        web.GetEmbeddedContent(),
		sanitizer: bluemonday.UGCPolicy(),
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

// Start the server (HTTP and HTTPS, if configured)
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

// Stop the server, allowing requests in flight to finish first.
func (s *Server) Stop() {
	wg := new(sync.WaitGroup)
	wg.Add(1)
	// nolint: gas
	go func() {
		ctx, cxl := context.WithTimeout(context.Background(), 5*time.Second)
		defer cxl()
		s.svr.Shutdown(ctx)
		wg.Done()
	}()
	if s.tlsSvr != nil {
		wg.Add(1)
		// nolint: gas
		go func() {
			ctx, cxl := context.WithTimeout(context.Background(), 5*time.Second)
			defer cxl()
			s.tlsSvr.Shutdown(ctx)
			wg.Done()
		}()
	}
	wg.Wait()
}

// ServeHTTP fulfills http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.conf.ForceTLS && r.TLS == nil && r.Header.Get("X-Forwarded-Proto") != "https" {
		u := s.conf.BaseURI + r.URL.Path
		http.Redirect(w, r, u, http.StatusFound)
		return
	}

	w.Header().Add("Vary", "Accept")
	if freenote.Version != "" {
		w.Header().Add("X-Freenote-Version", freenote.Version+" (build "+freenote.Build+")")
	}

	db, err := store.NewSession(s.conf)
	if handleError(w, err) {
		return
	}
	if clo, ok := db.(io.Closer); ok {
		defer clo.Close()
	}
	r = r.WithContext(store.NewContext(r.Context(), db))
	user, err := s.authenticate(w, r, db.UserStore())
	switch err {
	case errNoAuth, nil:
	case errAuthCookieInvalid:
		http.Error(w, "Unauthorized: Invalid Session Cookie", http.StatusUnauthorized)
		//statusResponse(w, http.StatusUnauthorized)
		return
	default:
		handleError(w, err)
		return
	}
	r = r.WithContext(users.NewContext(r.Context(), user))
	if user.ID != uuid.Nil && rand.Float64() < 0.1 {
		user.CleanSessions()
		db.UserStore().SaveUser(&user)
	}
	if !s.authorize(r.URL.Path, user) {
		if user.Access == users.LevelAnon {
			statusResponse(w, http.StatusUnauthorized)
		} else {
			statusResponse(w, http.StatusForbidden)
		}
		return
	}
	first := firstSegment(r.URL.Path)
	switch first {
	case "users":
		stripSegment(r)
		s.doUsers(w, r)
	case "session":
		stripSegment(r)
		s.doSession(w, r)
	case "debug":
		doDebug(w, r)
	default:
		s.fs.ServeHTTP(w, r)
	}
}

func (s *Server) preflight(w http.ResponseWriter, r *http.Request, body []byte, methods ...string) {
	methList := strings.ToUpper(strings.Join(methods, ", "))
	headers := w.Header()
	reqHeader := r.Header.Get("Access-Control-Request-Headers")
	if reqHeader == "" {
		reqHeader = "*"
	}
	var origin string
	if r.TLS == nil {
		origin = fmt.Sprintf("http://%s:%d", r.Host, s.conf.Port)
	} else {
		origin = fmt.Sprintf("https://%s:%d", r.Host, s.conf.TLSPort)
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

func (s *Server) baseURI(r *http.Request) string {
	if s.conf.ForceTLS {
		return s.conf.BaseURI
	}
	u := strings.TrimPrefix(s.conf.BaseURI, "http")
	u = strings.TrimPrefix(u, "s")
	if p := r.Header.Get("X-Forwarded-Proto"); p != "" {
		return p + u
	}
	if r.TLS == nil {
		return "http" + u
	}
	return "https" + u
}

func statusResponse(w http.ResponseWriter, statusCode int) {
	http.Error(w, http.StatusText(statusCode), statusCode)
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
	if err == errUnauthorized {
		statusResponse(w, http.StatusForbidden)
	}
	if err != nil {
		log.Printf("%T: %[1]q", err)
		statusResponse(w, http.StatusInternalServerError)
		return true
	}
	return false
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

func firstSegment(path string) string {
	path = strings.Trim(path, "/")
	if idx := strings.Index(path, "/"); idx > 0 {
		return path[:idx]
	}
	return path
}

func stripSegment(r *http.Request) {
	r.URL.Path = strings.TrimPrefix(r.URL.Path, "/"+firstSegment(r.URL.Path))
}

func popSegment(r *http.Request) string {
	seg := firstSegment(r.URL.Path)
	r.URL.Path = strings.TrimPrefix(r.URL.Path, "/"+seg)
	if r.URL.Path == "" {
		r.URL.Path = "/"
	}
	return seg
}
