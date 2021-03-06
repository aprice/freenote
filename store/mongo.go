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

// NewMongoStore initializes a new Storm/Bolt data store.
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

// MongoStore handles the MongoDB backing store.
type MongoStore struct {
	db *mgo.Database
}

// NoteStore returns the NoteStore for this session.
func (s *MongoStore) NoteStore() NoteStore {
	return &MongoNoteStore{s.db.C("Notes")}
}

// UserStore returns the UserStore for this session.
func (s *MongoStore) UserStore() UserStore {
	return &MongoUserStore{s.db.C("Users")}
}

// Close this session.
func (s *MongoStore) Close() error {
	s.db.Session.Close()
	return nil
}

// MongoNoteStore handles the MongoDB-backed Note store.
type MongoNoteStore struct {
	c *mgo.Collection
}

// NoteByID retrieves a single note by its unique ID.
func (s *MongoNoteStore) NoteByID(id uuid.UUID) (notes.Note, error) {
	var result notes.Note
	err := s.c.FindId(id).One(&result)
	return result, mongoError(err)
}

// QueryNotes queries the collection of notes with the parameters given in query,
// and returns the requested page of notes, the total notes matching the query
// (ignoring pagination), and any error encountered.
func (s *MongoNoteStore) QueryNotes(query NoteQuery) ([]notes.Note, int, error) {
	result := []notes.Note{}
	qry := bson.M{}
	if query.Owner != uuid.Nil {
		qry["owner"] = query.Owner
	}
	if query.Folder != "" {
		qry["folder"] = query.Folder
	}
	if query.Tag != "" {
		qry["tags"] = query.Tag
	}
	if query.ModifiedSince.After(epoch) {
		qry["modified"] = bson.M{"$gt": query.ModifiedSince}
	}
	if query.Text != "" {
		//TODO: Full text search
	}
	q := s.c.Find(qry)
	total, err := q.Count()
	if err != nil {
		return nil, -1, err
	}
	//TODO: Allow controlled sort field & direction
	err = q.Sort("-modified").Skip(query.Page.Start).Limit(query.Page.Length).All(&result)
	if err == nil && len(result) == 0 {
		err = ErrNotFound
	}
	return result, total, mongoError(err)
}

// FoldersByFolder returns the list of folders with the given folder prefix.
func (s *MongoNoteStore) FoldersByFolder(userID uuid.UUID, folder string) ([]string, error) {
	result := []string{}
	err := s.c.Find(bson.M{"owner": userID, "folder": folder}).Sort("folder").Distinct("folder", &result)
	if err == nil && len(result) == 0 {
		err = ErrNotFound
	}
	return result, mongoError(err)
}

// Tags returns the list of tags used by the given user.
func (s *MongoNoteStore) Tags(userID uuid.UUID) ([]string, error) {
	result := []string{}
	err := s.c.Find(bson.M{"owner": userID}).Distinct("tags", &result)
	if err == nil && len(result) == 0 {
		err = ErrNotFound
	}
	return result, mongoError(err)
}

// SaveNote saves a new or updated note to the data store.
func (s *MongoNoteStore) SaveNote(note *notes.Note) error {
	if note.ID == uuid.Nil {
		note.ID = uuid.NewV4()
	}
	_, err := s.c.UpsertId(note.ID, note)
	return mongoError(err)
}

// DeleteNote deletes the note with the given ID from the data store.
func (s *MongoNoteStore) DeleteNote(id uuid.UUID) error {
	err := s.c.Remove(bson.M{"_id": id})
	return mongoError(err)
}

// MongoUserStore handles the MongoDB-backed Note store.
type MongoUserStore struct {
	c *mgo.Collection
}

// UserByID retrieves a single user by its unique ID
func (s *MongoUserStore) UserByID(id uuid.UUID) (users.User, error) {
	var result users.User
	err := s.c.FindId(id).One(&result)
	return result, mongoError(err)
}

// UserByName retrieves a single user by its unique username.
func (s *MongoUserStore) UserByName(username string) (users.User, error) {
	var result users.User
	err := s.c.Find(bson.M{"username": username}).One(&result)
	return result, mongoError(err)
}

// Users retrieves a page of users. It returns the page of users and the total
// number of users.
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

// SaveUser saves a new or updated user to the data store.
func (s *MongoUserStore) SaveUser(user *users.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.NewV4()
	}
	_, err := s.c.UpsertId(user.ID, user)
	return mongoError(err)
}

// DeleteUser deletes a user with the given ID from the data store.
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
