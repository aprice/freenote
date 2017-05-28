package rest

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/aprice/freenote/store"
	"github.com/aprice/freenote/users"
	uuid "github.com/satori/go.uuid"
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
	} else if userID, sessID, err := parseSessionCookie(r); err != http.ErrNoCookie {
		if err == errAuthCookieInvalid {
			deleteSessionCookie(w)
			return users.User{}, errAuthCookieInvalid
		}
		user, err := us.UserByID(userID)
		if err != nil {
			deleteSessionCookie(w)
			return users.User{}, err
		}
		if !user.ValidateSession(sessID) {
			deleteSessionCookie(w)
			return users.User{}, errAuthCookieInvalid
		}
		writeSessionCookie(w, user.ID, sessID)
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

func parseSessionCookie(r *http.Request) (userID uuid.UUID, sessID uuid.UUID, err error) {
	userID = uuid.Nil
	sessID = uuid.Nil
	c, err := r.Cookie("auth")
	if err != nil {
		return
	}
	parts := strings.Split(c.Value, "|")
	if len(parts) != 2 {
		err = errAuthCookieInvalid
		return
	}
	bytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		err = errAuthCookieInvalid
		return
	}
	userID, err = uuid.FromBytes(bytes)
	if err != nil {
		err = errAuthCookieInvalid
		return
	}
	bytes, err = base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		err = errAuthCookieInvalid
		return
	}
	sessID, err = uuid.FromBytes(bytes)
	if err != nil {
		err = errAuthCookieInvalid
		return
	}
	return
}

func writeSessionCookie(w http.ResponseWriter, userID uuid.UUID, sessID uuid.UUID) {
	c := &http.Cookie{
		Name: "auth",
		Value: fmt.Sprintf("%s|%s",
			base64.RawURLEncoding.EncodeToString(userID.Bytes()),
			base64.RawURLEncoding.EncodeToString(sessID.Bytes())),
		MaxAge: 60 * 60 * 24 * 90,
	}
	http.SetCookie(w, c)
}

func deleteSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:    "auth",
		Value:   "",
		Expires: time.Now(),
		MaxAge:  -1,
	})
}
