package server

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
	"time"

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

// TestFirstLoad simulates the initial load of the UI.
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
		t.Fatalf("Server responded %d: %s", w.Code, truncate(w.Body.String(), 50))
	}
	cookie := w.Header().Get("Set-Cookie")

	req = httptest.NewRequest("GET", "/session", nil)
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Accept", "application/json")
	w = httptest.NewRecorder()
	s.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Server responded %d: %s", w.Code, truncate(w.Body.String(), 50))
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
		t.Fatalf("Server responded %d: %s", w.Code, truncate(w.Body.String(), 50))
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

// TestCRUDNote exercises the full note CRUD operations.
func TestCRUDNote(t *testing.T) {
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
	bodyBytes, err := ioutil.ReadAll(noteBodyFile)
	if err != nil {
		t.Fatal(err)
	}
	body := string(bodyBytes)

	noteHTMLFile, err := os.Open("testdata/cheatsheet.html")
	if err != nil {
		t.Fatal(err)
	}
	htmlBytes, err := ioutil.ReadAll(noteHTMLFile)
	if err != nil {
		t.Fatal(err)
	}
	html := string(htmlBytes)
	// Create
	note := notes.Note{Title: "Sample Note"}
	note.Body = body
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
		t.Fatalf("Server responded %d: %s", w.Code, truncate(w.Body.String(), 50))
	}

	note = notes.Note{}
	err = json.NewDecoder(w.Body).Decode(&note)
	if err != nil {
		t.Fatal(err)
	}
	expected := body
	actual := note.Body
	if actual != expected {
		t.Errorf("Markdown does not match expected value.\nExpected:\n%q\n-------\nActual:\n%q\n",
			expected, actual)
	}
	expected = normalizeHTML(html)
	actual = normalizeHTML(note.HTMLBody)
	if actual != expected {
		t.Errorf("HTML does not match expected value.\nExpected:\n%q\n-------\nActual:\n%q\n",
			expected, actual)
	}

	// Retrieve
	req = httptest.NewRequest("GET", fmt.Sprintf("/users/%s/notes/%s", userID, note.ID), nil)
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(testUsername, testPassword)
	w = httptest.NewRecorder()
	s.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Server responded %d: %s", w.Code, truncate(w.Body.String(), 50))
	}

	note = notes.Note{}
	err = json.NewDecoder(w.Body).Decode(&note)
	if err != nil {
		t.Fatal(err)
	}
	expected = body
	actual = note.Body
	if actual != expected {
		t.Errorf("Markdown does not match expected value.\nExpected:\n%q\n-------\nActual:\n%q\n",
			expected, actual)
	}
	expected = normalizeHTML(html)
	actual = normalizeHTML(note.HTMLBody)
	if actual != expected {
		t.Errorf("HTML does not match expected value.\nExpected:\n%q\n-------\nActual:\n%q\n",
			expected, actual)
	}

	// Update
	note.Title = "Sample Note Updated"
	note.HTMLBody = ""
	note.Modified = time.Now()
	b, err = json.Marshal(note)
	if err != nil {
		t.Fatal(err)
	}
	req = httptest.NewRequest("PUT", fmt.Sprintf("/users/%s/notes/%s", userID, note.ID),
		bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(testUsername, testPassword)
	w = httptest.NewRecorder()
	s.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Server responded %d: %s", w.Code, truncate(w.Body.String(), 50))
	}

	note = notes.Note{}
	err = json.NewDecoder(w.Body).Decode(&note)
	if err != nil {
		t.Fatal(err)
	}
	expected = body
	actual = note.Body
	if actual != expected {
		t.Errorf("Markdown does not match expected value.\nExpected:\n%q\n-------\nActual:\n%q\n",
			expected, actual)
	}
	expected = normalizeHTML(html)
	actual = normalizeHTML(note.HTMLBody)
	if actual != expected {
		t.Errorf("HTML does not match expected value.\nExpected:\n%q\n-------\nActual:\n%q\n",
			expected, actual)
	}

	// Delete
	req = httptest.NewRequest("DELETE", fmt.Sprintf("/users/%s/notes/%s", userID, note.ID), nil)
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(testUsername, testPassword)
	w = httptest.NewRecorder()
	s.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("Server responded %d: %s", w.Code, truncate(w.Body.String(), 50))
	}

	req = httptest.NewRequest("GET", fmt.Sprintf("/users/%s/notes/%s", userID, note.ID), nil)
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(testUsername, testPassword)
	w = httptest.NewRecorder()
	s.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound && w.Code != http.StatusGone {
		t.Fatalf("Server responded %d: %s", w.Code, truncate(w.Body.String(), 50))
	}
}

// TODO: Figure out why IDs aren't consistent
var normalizeSpaces = regexp.MustCompile(`(?ms)\s+`)
var normalizeIDs = regexp.MustCompile(`(?i)\s+id="[^"]*"\s*`)
var normalizeTags = regexp.MustCompile(`>\s+<`)

func normalizeHTML(in string) string {
	in = strings.TrimSpace(in)
	in = normalizeIDs.ReplaceAllString(in, " ")
	in = normalizeSpaces.ReplaceAllString(in, " ")
	in = normalizeTags.ReplaceAllString(in, "><")
	return in
}

func setupTest() (userID uuid.UUID, server *Server, err error) {
	userID, err = createTestUser()
	if err != nil {
		return
	}
	server, err = New(testConfig)
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

func truncate(s string, l int) string {
	if len(s) > l {
		return s[:l]
	}
	return s
}
