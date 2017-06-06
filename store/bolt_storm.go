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

func (s *StormNoteStore) DeleteNote(id uuid.UUID) error {
	err := s.db.Select(q.Eq("ID", id)).Delete(new(notes.Note))
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

func (s *StormUserStore) SaveUser(user *users.User) error {
	err := s.db.Save(user)
	if err == storm.ErrAlreadyExists {
		err = s.db.Update(user)
	}
	return err
}

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
