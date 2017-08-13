package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"golang.org/x/crypto/acme/autocert"

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

// New creates a new HTTP Server with the given configuration.
func New(conf config.Config) (*Server, error) {
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
	if s.upgrade(w, r) {
		return
	}

	path := strings.Trim(r.URL.Path, "/")
	if idx := strings.Index(path, "/"); idx > 0 {
		path = path[:idx]
	}
	switch path {
	case "session", "users":
		rh, err := NewRequestHandler(r, s.conf, s.sanitizer)
		if err != nil {
			if handleError(w, err) {
				return
			}
		}
		defer rh.close()
		rh.handle(w, r)
	case "debug":
		doDebug(w, r)
	default:
		s.fs.ServeHTTP(w, r)
	}
}

func (s *Server) upgrade(w http.ResponseWriter, r *http.Request) bool {
	// Prefer HTTPS but not HTTPS?
	if s.conf.CanonicalHTTPS && r.TLS == nil && r.Header.Get("X-Forwarded-Proto") != "https" {
		u := s.conf.BaseURI + r.URL.Path
		// Upgrade-Insecure-Requests -> 307 + Vary
		if r.Header.Get("Upgrade-Insecure-Requests") == "1" {
			w.Header().Add("Vary", "Upgrade-Insecure-Requests")
			http.Redirect(w, r, u, http.StatusTemporaryRedirect)
			return true
		}
		// ForceTLS -> 302
		if s.conf.ForceTLS {
			http.Redirect(w, r, u, http.StatusFound)
			return true
		}
	}
	return false
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
