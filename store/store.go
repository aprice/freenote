package store

import (
	"errors"

	uuid "github.com/satori/go.uuid"

	"github.com/aprice/freenote/config"
	"github.com/aprice/freenote/notes"
	"github.com/aprice/freenote/page"
	"github.com/aprice/freenote/users"
)

// ErrNotFound should be returned when any database query finds no
// results. Driver-specific not found errors should never be returned.
var ErrNotFound = errors.New("requested resource not found")

// Session implementations handle access to the backing store(s) for
// notes and users for a single session. They may optionally also be an
// io.Closer, and if they are, they can expect to be closed after each
// request.
type Session interface {
	NoteStore() NoteStore
	UserStore() UserStore
}

// NoteStore implementations handle access to the backing store for notes.
type NoteStore interface {
	NoteByID(id uuid.UUID) (notes.Note, error)
	QueryNotes(query NoteQuery) ([]notes.Note, int, error)
	//FoldersByFolder(userID uuid.UUID, folder string) ([]string, error)
	//Tags(userID uuid.UUID) ([]string, error)
	SaveNote(note *notes.Note) error
	DeleteNote(id uuid.UUID) error
}

// NoteQuery holds parameters for a Note store query.
type NoteQuery struct {
	Owner  uuid.UUID
	Folder string
	Tag    string
	Text   string
	Page   page.Page
}

// UserStore implementations handle access to the backing store for users.
type UserStore interface {
	UserByID(id uuid.UUID) (users.User, error)
	UserByName(username string) (users.User, error)
	Users(page page.Page) ([]users.User, int, error)
	SaveUser(user *users.User) error
	DeleteUser(id uuid.UUID) error
}

// NewSession returns a new database session for the given configuration,
// including selecting the appropriate database driver.
func NewSession(conf config.Config) (Session, error) {
	if conf.BoltDB != "" {
		return NewStormStore(conf)
	} else if conf.Mongo != config.NilConnection {
		return NewMongoStore(conf)
	}
	return nil, errors.New("No backing store confiured")
}
