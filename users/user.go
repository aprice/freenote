package users

import (
	"errors"
	"time"

	uuid "github.com/satori/go.uuid"
)

var ErrAuthenticationFailed = errors.New("authentication failed, incorrect username or password")
var ErrUsernameTooShort = errors.New("username too short")
var ErrUsernameTooLong = errors.New("username too long")

// SessionLifetime is the duration that a user session will be valid.
const SessionLifetime = 90 * 24 * time.Hour

// User represents a credentialed user in the system.
type User struct {
	ID          uuid.UUID   `json:"id" xml:"id,attr" bson:"_id"`
	Username    string      `json:"username" storm:"unique"`
	DisplayName string      `json:"name"`
	Password    *Password   `json:"password,omitempty" xml:"-"`
	Access      AccessLevel `json:"access"`
	Sessions    []*Session  `json:"sessions,omitempty" xml:"-"`
}

// New creates a new user with the given username and default access.
func New(username string) User {
	return User{
		ID:       uuid.NewV4(),
		Username: username,
		Access:   LevelUser,
		Sessions: make([]*Session, 0),
	}
}

// ValidateSession checks the validity of a user session.
func (u *User) ValidateSession(sessID uuid.UUID, key string) bool {
	if u.Sessions == nil || len(u.Sessions) == 0 {
		return false
	}
	for _, sess := range u.Sessions {
		if sess.ID == sessID && sess.Expires.After(time.Now()) {
			if ok, err := sess.Key.Verify(key); ok && err == nil {
				sess.Expires = time.Now().Add(SessionLifetime)
				return true
			}
		}
	}
	return false
}

// NewSession creates a new session for this user.
func (u *User) NewSession() (session Session, err error) {
	key, pw, err := RandomPassword(16)
	if err != nil {
		return Session{}, err
	}
	sid := uuid.NewV4()
	sess := Session{
		ID:      sid,
		UserID:  u.ID,
		Expires: time.Now().Add(SessionLifetime),
		Key:     pw,
		Secret:  key,
	}
	u.Sessions = append(u.Sessions, &sess)
	return sess, nil
}

// CleanSessions removes expired sessions for this user.
func (u *User) CleanSessions() {
	if u.Sessions == nil || len(u.Sessions) == 0 {
		return
	}
	s := make([]*Session, 0, len(u.Sessions))
	for _, sess := range u.Sessions {
		if sess.Expires.After(time.Now()) {
			s = append(s, sess)
		}
	}
	u.Sessions = s
}

// Session represents a session for this user
type Session struct {
	ID      uuid.UUID `json:"id"`
	UserID  uuid.UUID `json:"-"`
	Expires time.Time
	Key     *Password
	Secret  string `json:"-"`
}

// AccessLevel indicates a user's permissions.
type AccessLevel uint8

const (
	// LevelAnon indicates an unauthenticated user
	LevelAnon AccessLevel = iota
	// LevelGuest indicates an authenticated user with no permissions
	LevelGuest
	// LevelUser indicates an authenticated user with default permissions
	LevelUser
	// LevelAdmin indicates an administrator
	LevelAdmin
	// LevelRecovery indicates the recovery user
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

// ParseAccessLevel from string as returned by AccessLevel.String().
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

// ValidateUsername checks that a username meets requirements.
func ValidateUsername(name string) error {
	if len(name) < 3 {
		return ErrUsernameTooShort
	}
	if len(name) > 24 {
		return ErrUsernameTooLong
	}
	return nil
}
