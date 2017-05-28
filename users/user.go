package users

import (
	"errors"
	"log"
	"time"

	uuid "github.com/satori/go.uuid"
)

var ErrAuthenticationFailed = errors.New("authentication failed, incorrect username or password")

const SessionLifetime = 90 * 24 * time.Hour

type User struct {
	ID          uuid.UUID   `json:"id" xml:"id,attr" bson:"_id"`
	Username    string      `json:"username" storm:"unique"`
	DisplayName string      `json:"name"`
	Password    *Password   `json:"password,omitempty" xml:"-"`
	Access      AccessLevel `json:"access"`
	Sessions    []Session   `json:"sessions,omitempty" xml:"-"`
}

func (u *User) ValidateSession(sessID uuid.UUID) bool {
	if u.Sessions == nil || len(u.Sessions) == 0 {
		return false
	}
	for idx, sess := range u.Sessions {
		if sess.ID == sessID && sess.Expires.After(time.Now()) {
			u.Sessions[idx].Expires = time.Now().Add(SessionLifetime)
			return true
		}
	}
	return false
}

func (u *User) NewSession() uuid.UUID {
	id := uuid.NewV4()
	u.Sessions = append(u.Sessions, Session{
		ID:      id,
		Expires: time.Now().Add(SessionLifetime),
	})
	return id
}

func (u *User) CleanSessions() {
	if u.Sessions == nil || len(u.Sessions) == 0 {
		return
	}
	s := make([]Session, 0)
	for _, sess := range u.Sessions {
		if sess.Expires.After(time.Now()) {
			s = append(s, sess)
		}
	}
	log.Printf("Cleaned %s's sessions, %d before cleaning, %d after",
		u.Username, len(u.Sessions), len(s))
	u.Sessions = s
}

type Session struct {
	ID      uuid.UUID `json:"id"`
	Expires time.Time
}

type AccessLevel uint8

const (
	LevelAnon AccessLevel = iota
	LevelGuest
	LevelUser
	LevelAdmin
	LevelRecovery
)

func (al AccessLevel) String() string {
	switch al {
	case LevelGuest:
		return "guest"
	case LevelUser:
		return "user"
	case LevelAdmin:
		return "admin"
	case LevelRecovery:
		return "recovery"
	default:
		return "anonymous"
	}
}

func ParseAccessLevel(in string) AccessLevel {
	switch in {
	case "guest":
		return LevelGuest
	case "user":
		return LevelUser
	case "admin":
		return LevelAdmin
	case "recovery":
		return LevelRecovery
	default:
		return LevelAnon
	}
}

var ErrUsernameTooShort = errors.New("username too short")
var ErrUsernameTooLong = errors.New("username too long")

func ValidateUsername(name string) error {
	if len(name) < 3 {
		return ErrUsernameTooShort
	}
	if len(name) > 24 {
		return ErrUsernameTooLong
	}
	return nil
}
