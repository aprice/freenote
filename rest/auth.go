package rest

import (
	"encoding/base64"
	"errors"
	"net/http"
	"regexp"
	"time"

	uuid "github.com/satori/go.uuid"

	"github.com/aprice/freenote/store"
	"github.com/aprice/freenote/users"
)

var errNoAuth = errors.New("no authentication provided")
var errAuthFailed = errors.New("authentication failed")
var errAuthCookieInvalid = errors.New("auth cookie invalid")

// Supported authentication: HTTP Basic, HTTP Bearer, Cookie
func (s *Server) authenticate(w http.ResponseWriter, r *http.Request, us store.UserStore) (users.User, error) {
	if uname, pass, ok := r.BasicAuth(); ok {
		if uname == users.RecoveryAdminName {
			user, err := users.AuthenticateAdmin(pass)
			if err != nil {
				return users.User{}, err
			} else {
				return user, nil
			}
		}
		user, err := us.UserByName(uname)
		if err != nil {
			return users.User{}, err
		}
		ok, err = user.Password.Verify(pass)
		if err != nil {
			return users.User{}, err
		} else if !ok {
			return users.User{}, errAuthFailed
		} else {
			return user, nil
		}
	} else if sess, err := parseSessionCookie(r); err != http.ErrNoCookie {
		if err == errAuthCookieInvalid {
			deleteSessionCookie(w)
			return users.User{}, errAuthCookieInvalid
		}
		user, err := us.UserByID(sess.UserID)
		if err != nil {
			deleteSessionCookie(w)
			return users.User{}, err
		}
		if !user.ValidateSession(sess.ID, sess.Secret) {
			deleteSessionCookie(w)
			return users.User{}, errAuthCookieInvalid
		}
		refreshSessionCookie(w, r)
		return user, nil
	}

	return users.User{}, errNoAuth
}

var userOwnedPat = regexp.MustCompile("/users/([^/]+)/?.*")

// General authz based on path & method with no object details, prevents 404 fishing
func (s *Server) authorize(r *http.Request, user users.User) (bool, error) {
	if pts := userOwnedPat.FindStringSubmatch(r.URL.Path); len(pts) > 0 {
		if pts[0] == user.ID.String() || pts[0] == user.Username {
			return true, nil
		} else if user.Access >= users.LevelAdmin {
			return true, nil
		}
		//TODO: Sharing!
		return false, nil
	} else if r.URL.Path == "/users/" || r.URL.Path == "/users" {
		if user.Access >= users.LevelAdmin {
			return true, nil
		}
		return false, nil
	}
	return true, nil
}

func parseSessionCookie(r *http.Request) (users.Session, error) {
	sess := users.Session{}
	c, err := r.Cookie("auth")
	if err != nil {
		return sess, err
	}
	b, err := base64.RawURLEncoding.DecodeString(c.Value)
	if err != nil {
		err = errAuthCookieInvalid
		return sess, err
	}
	sess.UserID, err = uuid.FromBytes(b[:16])
	if err != nil {
		err = errAuthCookieInvalid
		return users.Session{}, err
	}
	sess.ID, err = uuid.FromBytes(b[16:32])
	if err != nil {
		err = errAuthCookieInvalid
		return users.Session{}, err
	}
	sess.Secret = string(b[32:])
	return sess, nil
}

func writeSessionCookie(w http.ResponseWriter, sess users.Session) {
	b := make([]byte, 32+len(sess.Secret))
	copy(b, sess.UserID.Bytes())
	copy(b[16:], sess.ID.Bytes())
	copy(b[32:], []byte(sess.Secret))
	c := &http.Cookie{
		Name:   "auth",
		Value:  base64.RawURLEncoding.EncodeToString(b),
		MaxAge: 60 * 60 * 24 * 90,
		Path:   "/",
	}
	http.SetCookie(w, c)
}

func refreshSessionCookie(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("auth")
	if c != nil && err == nil {
		c.MaxAge = 60 * 60 * 24 * 90
		c.Path = "/"
		http.SetCookie(w, c)
	}
}

func deleteSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:    "auth",
		Value:   "",
		Expires: time.Now(),
		MaxAge:  -1,
		Path:    "/",
	})
}
