package store

import (
	"strings"

	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	uuid "github.com/satori/go.uuid"

	"github.com/aprice/freenote/config"
	"github.com/aprice/freenote/notes"
	"github.com/aprice/freenote/page"
	"github.com/aprice/freenote/users"
)

// NewStormStore initializes a new Storm/Bolt data store.
func NewStormStore(conf config.Config) (Session, error) {
	db, err := storm.Open(conf.BoltDB)
	if err != nil {
		return nil, err
	}
	return &StormStore{db}, nil
}

// StormStore handles the Storm/Bolt backing store.
type StormStore struct {
	db *storm.DB
}

// NoteStore returns the NoteStore for this session.
func (s *StormStore) NoteStore() NoteStore {
	store := &StormNoteStore{s.db.From("notes")}
	store.db.Init(&notes.Note{})
	return store
}

// UserStore returns the UserStore for this session.
func (s *StormStore) UserStore() UserStore {
	store := &StormUserStore{s.db.From("users")}
	store.db.Init(&users.User{})
	return store
}

// Close this session.
func (s *StormStore) Close() error {
	return s.db.Close()
}

// StormNoteStore handles the Storm/Bolt backed Note store.
type StormNoteStore struct {
	db storm.Node
}

// NoteByID retrieves a single note by its unique ID.
func (s *StormNoteStore) NoteByID(id uuid.UUID) (notes.Note, error) {
	var result notes.Note
	err := s.db.One("ID", id, &result)
	return result, stormError(err)
}

// NotesByOwner retrieves a page of notes by owner ID. It returns the page of
// notes and the total number of notes owned by the given user.
func (s *StormNoteStore) NotesByOwner(userID uuid.UUID, page page.Page) ([]notes.Note, int, error) {
	var result []notes.Note
	qry := s.db.Select(q.Eq("Owner", userID))
	total, err := qry.Count(new(notes.Note))
	if err != nil {
		return nil, -1, err
	} else if total == 0 {
		return make([]notes.Note, 0), 0, nil
	}
	err = applyPage(qry, page).Find(&result)
	return result, total, stormError(err)
}

// NotesByFolder retrieves a page of notes with a given owner and folder. It
// returns the page of notes and the total number of notes with the given owner
// and folder.
func (s *StormNoteStore) NotesByFolder(userID uuid.UUID, folder string, page page.Page) ([]notes.Note, int, error) {
	var result []notes.Note
	qry := s.db.Select(q.And(q.Eq("Owner", userID), q.Eq("Folder", folder)))
	total, err := qry.Count(new(notes.Note))
	if err != nil {
		return nil, -1, err
	}
	err = applyPage(qry, page).Find(&result)
	return result, total, stormError(err)
}

// SaveNote saves a new or updated note to the data store.
func (s *StormNoteStore) SaveNote(note *notes.Note) error {
	h := note.HTMLBody
	note.HTMLBody = ""
	err := s.db.Save(note)
	if err == storm.ErrAlreadyExists {
		err = s.db.Update(note)
	}
	note.HTMLBody = h
	return err
}

// DeleteNote deletes the note with the given ID from the data store.
func (s *StormNoteStore) DeleteNote(id uuid.UUID) error {
	err := s.db.Select(q.Eq("ID", id)).Delete(new(notes.Note))
	return stormError(err)
}

// StormUserStore handles the Storm/Bolt backed Note store.
type StormUserStore struct {
	db storm.Node
}

// UserByID retrieves a single user by its unique ID
func (s *StormUserStore) UserByID(id uuid.UUID) (users.User, error) {
	var result users.User
	err := s.db.One("ID", id, &result)
	return result, stormError(err)
}

// UserByName retrieves a single user by its unique username.
func (s *StormUserStore) UserByName(username string) (users.User, error) {
	var result users.User
	err := s.db.One("Username", username, &result)
	return result, stormError(err)
}

// Users retrieves a page of users. It returns the page of users and the total
// number of users.
func (s *StormUserStore) Users(page page.Page) ([]users.User, int, error) {
	var result []users.User
	qry := s.db.Select()
	total, err := qry.Count(new(users.User))
	if err != nil {
		return nil, -1, err
	}
	err = applyPage(qry, page).Find(&result)
	return result, total, stormError(err)
}

// SaveUser saves a new or updated user to the data store.
func (s *StormUserStore) SaveUser(user *users.User) error {
	err := s.db.Save(user)
	if err == storm.ErrAlreadyExists {
		err = s.db.Update(user)
	}
	return err
}

// DeleteUser deletes a user with the given ID from the data store.
func (s *StormUserStore) DeleteUser(id uuid.UUID) error {
	err := s.db.Select(q.Eq("ID", id)).Delete(new(users.User))
	return stormError(err)
}

func stormError(err error) error {
	if err == nil {
		return nil
	}
	if err == storm.ErrNotFound {
		return ErrNotFound
	}
	return err
}

func applyPage(qry storm.Query, page page.Page) storm.Query {
	if page.SortDescending {
		qry.Reverse()
	}
	return qry.OrderBy(strings.Title(page.SortBy)).Skip(page.Start).Limit(page.Length)
}
