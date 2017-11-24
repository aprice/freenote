package rest

import (
	"fmt"
	"strings"

	"github.com/aprice/freenote/notes"
	"github.com/aprice/freenote/page"
	"github.com/aprice/freenote/users"
)

// DecoratedNote represents a Note with hypermedia links.
type DecoratedNote struct {
	Links Links `json:"_links" xml:"Links>Link"`
	notes.Note
	XMLName struct{} `json:"-" xml:"Note"`
}

// DecorateNote decorates a Note with hypermedia links.
func DecorateNote(note notes.Note, canWrite bool, baseURI string) DecoratedNote {
	links := Links{}
	links.RecordRUD(fmt.Sprintf("%s/users/%s/notes/%s", baseURI, note.Owner, note.ID), canWrite)
	links.Add(Link{
		Rel:  "author",
		Href: fmt.Sprintf("%s/users/%s", baseURI, note.Owner),
	})
	return DecoratedNote{Note: note, Links: links}
}

// DecoratedNotes represents a collection of notes with hypermedia links for the
// collection and notes.
type DecoratedNotes struct {
	Links   Links           `json:"_links" xml:"Links>Link"`
	Notes   []DecoratedNote `json:"notes" xml:"Page>Note"`
	XMLName struct{}        `json:"-" xml:"Notes"`
}

// DecorateNotes decorates a collection of Notes with hypermedia links for the
// collection and notes.
func DecorateNotes(owner users.User, values []notes.Note, folder string, page page.Page, canWrite bool, baseURI string) DecoratedNotes {
	links := Links{}
	decorated := make([]DecoratedNote, len(values))
	var base string
	if folder == "" {
		base = fmt.Sprintf("%s/users/%s/notes", baseURI, owner.ID)
	} else {
		base = fmt.Sprintf("%s/users/%s/notes/%s", baseURI, owner.ID, folder)
	}
	for i := range values {
		idx := strings.Index(values[i].Body, "\n")
		if idx > 0 {
			values[i].Body = values[i].Body[:idx]
		}
		decorated[i] = DecorateNote(values[i], canWrite, baseURI)
	}
	links.CollectionCR(base, page, canWrite)
	return DecoratedNotes{Notes: decorated, Links: links}
}
