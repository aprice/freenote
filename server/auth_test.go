package server

import (
	"fmt"
	"testing"

	"github.com/aprice/freenote/users"
)

func TestAuthorize(t *testing.T) {
	user := users.New("test")
	url := fmt.Sprintf("/users/%s/notes", user.ID)
	ok := authorize(url, user)
	if !ok {
		t.Errorf("User %s can't access own notes at %s", user.ID, url)
	}
}
