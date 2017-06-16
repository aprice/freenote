package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/aprice/freenote/config"
	"github.com/aprice/freenote/notes"
	"github.com/aprice/freenote/store"
	"github.com/aprice/freenote/users"
	uuid "github.com/satori/go.uuid"
)

const testBoltDB = "test.db"
const testUsername = "test"
const testPassword = "swordfish"

func TestFirstLoad(t *testing.T) {
	// Create test user
	c := getTestConfig()
	userID, err := createTestUser(c)
	defer destroyTestDB()
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewServer(c)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("POST", "/session",
		strings.NewReader(fmt.Sprintf("username=%s&password=%s", testUsername, testPassword)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("Server responded %d: %s", w.Code, w.Body)
	}
	cookie := w.Header().Get("Set-Cookie")

	req = httptest.NewRequest("GET", "/session", nil)
	req.Header.Set("Cookie", cookie)
	w = httptest.NewRecorder()
	s.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("Server responded %d: %s", w.Code, w.Body)
	}
	user := new(users.User)
	err = json.NewDecoder(w.Body).Decode(user)
	if err != nil {
		t.Fatal(err)
	}
	if user.ID != userID {
		t.Fatal("Returned user ID does not match created user ID")
	}

	req = httptest.NewRequest("GET", fmt.Sprintf("/users/%s/notes", user.ID.String()), nil)
	req.Header.Set("Cookie", cookie)
	w = httptest.NewRecorder()
	s.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("Server responded %d: %s", w.Code, w.Body)
	}
	notesPayload := struct{ Notes []notes.Note }{}
	err = json.NewDecoder(w.Body).Decode(&notesPayload)
	if err != nil {
		t.Fatal(err)
	}
	if len(notesPayload.Notes) != 1 {
		t.Fatalf("Returned notes not length 1, actually %d", len(notesPayload.Notes))
	}
}

func getTestConfig() config.Config {
	return config.Config{
		Port:   18000,
		BoltDB: testBoltDB,
	}
}

func createTestUser(conf config.Config) (uuid.UUID, error) {
	db, err := store.NewSession(conf)
	if err != nil {
		return uuid.Nil, err
	}
	if closer, ok := db.(io.Closer); ok {
		defer closer.Close()
	}
	testUser := users.New(testUsername)
	testUser.Password, err = users.NewPassword(testPassword)
	if err != nil {
		return uuid.Nil, err
	}
	err = db.UserStore().SaveUser(&testUser)
	if err != nil {
		return uuid.Nil, err
	}
	wn := notes.WelcomeNote(testUser.ID)
	err = db.NoteStore().SaveNote(&wn)
	if err != nil {
		return uuid.Nil, err
	}
	return testUser.ID, nil
}

func destroyTestDB() {
	os.Remove(testBoltDB)
}
