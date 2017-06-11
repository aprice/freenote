package users

import (
	"errors"
	"time"

	uuid "github.com/satori/go.uuid"
)

var ErrAuthenticationFailed = errors.New("authentication failed, incorrect username or password")
var ErrUsernameTooShort = errors.New("username too short")
var ErrUsernameTooLong = errors.New("username too long")

const SessionLifetime = 90 * 24 * time.Hour

type User struct {
	ID          uuid.UUID   `json:"id" xml:"id,attr" bson:"_id"`
	Username    string      `json:"username" storm:"unique"`
	DisplayName string      `json:"name"`
	Password    *Password   `json:"password,omitempty" xml:"-"`
	Access      AccessLevel `json:"access"`
	Sessions    []*Session  `json:"sessions,omitempty" xml:"-"`
}

func New(username string) User {
	return User{
		ID:       uuid.NewV4(),
		Username: username,
		Access:   LevelUser,
		Sessions: make([]*Session, 0),
	}
}

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

type Session struct {
	ID      uuid.UUID `json:"id"`
	UserID  uuid.UUID `json:"-"`
	Expires time.Time
	Key     *Password
	Secret  string `json:"-"`
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

func ValidateUsername(name string) error {
	if len(name) < 3 {
		return ErrUsernameTooShort
	}
	if len(name) > 24 {
		return ErrUsernameTooLong
	}
	return nil
}
