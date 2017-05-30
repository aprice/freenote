package store

import (
	"github.com/asdine/storm"
	uuid "github.com/satori/go.uuid"

	"github.com/aprice/freenote/config"
	"github.com/aprice/freenote/notes"
	"github.com/aprice/freenote/page"
	"github.com/aprice/freenote/users"
)

func NewStormStore(conf config.Config) (Session, error) {
	db, err := storm.Open(conf.BoltDB)
	if err != nil {
		return nil, err
	}
	return &StormStore{db}, nil
}

type StormStore struct {
	db *storm.DB
}

func (s *StormStore) NoteStore() NoteStore {
	store := &StormNoteStore{s.db.From("notes")}
	store.db.Init(&notes.Note{})
	return store
}

func (s *StormStore) UserStore() UserStore {
	store := &StormUserStore{s.db.From("users")}
	store.db.Init(&users.User{})
	return store
}

func (s *StormStore) Close() error {
	return s.db.Close()
}

type StormNoteStore struct {
	db storm.Node
}

func (s *StormNoteStore) NoteByID(id uuid.UUID) (notes.Note, error) {
	var result notes.Note
	err := s.db.One("ID", id, &result)
	return result, stormError(err)
}

func (s *StormNoteStore) NotesByOwner(userID uuid.UUID, page page.Page) ([]notes.Note, error) {
	var result []notes.Note
	//TODO: User-controlled sorting
	err := s.db.Find("Owner", userID, &result, storm.Limit(page.Length), storm.Skip(page.Start))
	return result, stormError(err)
}

func (s *StormNoteStore) NotesByFolder(userID uuid.UUID, folder string, page page.Page) ([]notes.Note, error) {
	var result []notes.Note
	//TODO: User-controlled sorting
	err := s.db.Find("Folder", folder, &result)
	return result, stormError(err)
}

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

func (s *StormNoteStore) DeleteNote(note notes.Note) error {
	err := s.db.DeleteStruct(note)
	return stormError(err)
}

type StormUserStore struct {
	db storm.Node
}

func (s *StormUserStore) UserByID(id uuid.UUID) (users.User, error) {
	var result users.User
	err := s.db.One("ID", id, &result)
	return result, stormError(err)
}

func (s *StormUserStore) UserByName(username string) (users.User, error) {
	var result users.User
	err := s.db.One("Username", username, &result)
	return result, stormError(err)
}

func (s *StormUserStore) Users(page page.Page) ([]users.User, error) {
	var result []users.User
	//TODO: User-controlled sorting
	err := s.db.All(&result, storm.Limit(page.Length), storm.Skip(page.Start))
	return result, stormError(err)
}

func (s *StormUserStore) SaveUser(user *users.User) error {
	err := s.db.Save(user)
	if err == storm.ErrAlreadyExists {
		err = s.db.Update(user)
	}
	return err
}

func (s *StormUserStore) DeleteUser(user users.User) error {
	err := s.db.DeleteStruct(user)
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
