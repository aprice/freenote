package store

import (
	"io"

	uuid "github.com/satori/go.uuid"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/aprice/freenote/config"
	"github.com/aprice/freenote/notes"
	"github.com/aprice/freenote/page"
	"github.com/aprice/freenote/users"
)

var session *mgo.Session

func initMongoDB(conf config.Config) error {
	var err error
	di := mgo.DialInfo{
		Addrs:    []string{conf.Mongo.Host},
		Database: conf.Mongo.Namespace,
	}
	if conf.Mongo.User != "" {
		di.Username = conf.Mongo.User
		di.Password = conf.Mongo.User
	}
	session, err = mgo.DialWithInfo(&di)
	return err
}

func NewMongoStore(conf config.Config) (Session, error) {
	if session == nil {
		err := initMongoDB(conf)
		if err != nil {
			return nil, err
		}
	}
	if err := session.Ping(); err != nil {
		return nil, err
	}
	sess := session.Copy()
	return &MongoStore{sess.DB(conf.Mongo.Namespace)}, nil
}

var _ Session = (*MongoStore)(nil)
var _ io.Closer = (*MongoStore)(nil)

type MongoStore struct {
	db *mgo.Database
}

func (s *MongoStore) NoteStore() NoteStore {
	return &MongoNoteStore{s.db.C("Notes")}
}

func (s *MongoStore) UserStore() UserStore {
	return &MongoUserStore{s.db.C("Users")}
}

func (s *MongoStore) Close() error {
	s.db.Session.Close()
	return nil
}

var _ NoteStore = (*MongoNoteStore)(nil)

type MongoNoteStore struct {
	c *mgo.Collection
}

func (s *MongoNoteStore) NoteByID(id uuid.UUID) (notes.Note, error) {
	var result notes.Note
	err := s.c.FindId(id).One(&result)
	return result, mongoError(err)
}

func (s *MongoNoteStore) NotesByOwner(userID uuid.UUID, page page.Page) ([]notes.Note, int, error) {
	result := []notes.Note{}
	q := s.c.Find(bson.M{"owner": userID})
	total, err := q.Count()
	if err != nil {
		return nil, -1, err
	}
	//TODO: Allow controlled sort field & direction
	err = q.Sort("-modified").Skip(page.Start).Limit(page.Length).All(&result)
	if err == nil && len(result) == 0 {
		err = ErrNotFound
	}
	return result, total, mongoError(err)
}

func (s *MongoNoteStore) NotesByFolder(userID uuid.UUID, folder string, page page.Page) ([]notes.Note, int, error) {
	result := []notes.Note{}
	q := s.c.Find(bson.M{"owner": userID, "folder": folder})
	total, err := q.Count()
	if err != nil {
		return nil, -1, err
	}
	//TODO: Allow controlled sort field & direction
	err = q.Sort("-modified").Skip(page.Start).Limit(page.Length).All(&result)
	if err == nil && len(result) == 0 {
		err = ErrNotFound
	}
	return result, total, mongoError(err)
}

func (s *MongoNoteStore) FoldersByFolder(userID uuid.UUID, folder string) ([]string, error) {
	result := []string{}
	err := s.c.Find(bson.M{"owner": userID, "folder": folder}).Sort("folder").Distinct("folder", &result)
	if err == nil && len(result) == 0 {
		err = ErrNotFound
	}
	return result, mongoError(err)
}

func (s *MongoNoteStore) NotesByTag(userID uuid.UUID, tag string, page page.Page) ([]notes.Note, error) {
	result := []notes.Note{}
	//TODO: Allow controlled sort field & direction
	err := s.c.Find(bson.M{"owner": userID, "tags": tag}).Sort("-modified").Skip(page.Start).Limit(page.Length).All(&result)
	if err == nil && len(result) == 0 {
		err = ErrNotFound
	}
	return result, mongoError(err)
}

func (s *MongoNoteStore) Tags(userID uuid.UUID, page page.Page) ([]string, error) {
	result := []string{}
	err := s.c.Find(bson.M{"owner": userID}).Distinct("tags", &result)
	if err == nil && len(result) == 0 {
		err = ErrNotFound
	}
	return result, mongoError(err)
}

func (s *MongoNoteStore) SaveNote(note *notes.Note) error {
	if note.ID == uuid.Nil {
		note.ID = uuid.NewV4()
	}
	_, err := s.c.UpsertId(note.ID, note)
	return mongoError(err)
}

func (s *MongoNoteStore) DeleteNote(id uuid.UUID) error {
	err := s.c.Remove(bson.M{"_id": id})
	return mongoError(err)
}

var _ UserStore = (*MongoUserStore)(nil)

type MongoUserStore struct {
	c *mgo.Collection
}

func (s *MongoUserStore) UserByID(id uuid.UUID) (users.User, error) {
	var result users.User
	err := s.c.FindId(id).One(&result)
	return result, mongoError(err)
}

func (s *MongoUserStore) UserByName(username string) (users.User, error) {
	var result users.User
	err := s.c.Find(bson.M{"username": username}).One(&result)
	return result, mongoError(err)
}

func (s *MongoUserStore) Users(page page.Page) ([]users.User, int, error) {
	result := []users.User{}
	q := s.c.Find(nil)
	total, err := q.Count()
	if err != nil {
		return nil, -1, err
	}
	//TODO: Allow controlled sort field & direction
	err = q.Sort("username").Skip(page.Start).Limit(page.Length).All(&result)
	if err == nil && len(result) == 0 {
		err = ErrNotFound
	}
	return result, total, mongoError(err)
}

func (s *MongoUserStore) SaveUser(user *users.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.NewV4()
	}
	_, err := s.c.UpsertId(user.ID, user)
	return mongoError(err)
}

func (s *MongoUserStore) DeleteUser(id uuid.UUID) error {
	err := s.c.Remove(bson.M{"_id": id})
	return mongoError(err)
}

func mongoError(err error) error {
	if err == nil {
		return nil
	}
	if err == mgo.ErrNotFound {
		return ErrNotFound
	}
	return err
}
