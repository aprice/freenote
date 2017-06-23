package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"

	"io/ioutil"

	"github.com/aprice/freenote/config"
	"github.com/aprice/freenote/notes"
	"github.com/aprice/freenote/store"
	"github.com/aprice/freenote/users"
	uuid "github.com/satori/go.uuid"
)

const testBoltDB = "test.db"
const testUsername = "test"
const testPassword = "swordfish"

var testConfig = config.Config{
	Port:   18000,
	BoltDB: testBoltDB,
}

func TestFirstLoad(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	userID, s, err := setupTest()
	defer cleanupTest()
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/session",
		strings.NewReader(fmt.Sprintf("username=%s&password=%s", testUsername, testPassword)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Server responded %d: %s", w.Code, w.Body)
	}
	cookie := w.Header().Get("Set-Cookie")

	req = httptest.NewRequest("GET", "/session", nil)
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Accept", "application/json")
	w = httptest.NewRecorder()
	s.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
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
	req.Header.Set("Accept", "application/json")
	w = httptest.NewRecorder()
	s.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
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

func TestCreate(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	userID, s, err := setupTest()
	defer cleanupTest()
	if err != nil {
		t.Fatal(err)
	}

	noteBodyFile, err := os.Open("testdata/cheatsheet.md")
	if err != nil {
		t.Fatal(err)
	}
	body, err := ioutil.ReadAll(noteBodyFile)
	if err != nil {
		t.Fatal(err)
	}

	noteHTMLFile, err := os.Open("testdata/cheatsheet.html")
	if err != nil {
		t.Fatal(err)
	}
	html, err := ioutil.ReadAll(noteHTMLFile)
	if err != nil {
		t.Fatal(err)
	}

	note := notes.Note{Title: "Sample Note"}
	note.Body = string(body)
	b, err := json.Marshal(note)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("POST", fmt.Sprintf("/users/%s/notes", userID),
		bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(testUsername, testPassword)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("Server responded %d: %s", w.Code, w.Body)
	}

	note = notes.Note{}
	err = json.NewDecoder(w.Body).Decode(&note)
	if err != nil {
		t.Fatal(err)
	}
	re := regexp.MustCompile(`\s+`)
	expected := re.ReplaceAllString(string(body), " ")
	actual := re.ReplaceAllString(note.Body, " ")
	if actual != expected {
		t.Errorf("Markdown does not match expected value.\nExpected:\n%q\n-------\nActual:\n%q\n",
			expected, actual)
	}
	expected = re.ReplaceAllString(string(html), " ")
	actual = re.ReplaceAllString(note.HTMLBody, " ")
	if actual != expected {
		t.Errorf("HTML does not match expected value.\nExpected:\n%q\n-------\nActual:\n%q\n",
			expected, actual)
	}
}

func setupTest() (userID uuid.UUID, server *Server, err error) {
	userID, err = createTestUser()
	if err != nil {
		return
	}
	server, err = NewServer(testConfig)
	return
}

func cleanupTest() {
	os.Remove(testBoltDB)
}

func createTestUser() (uuid.UUID, error) {
	db, err := store.NewSession(testConfig)
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
