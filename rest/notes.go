package rest

import (
	"fmt"
	"strings"

	"github.com/aprice/freenote/notes"
	"github.com/aprice/freenote/page"
	"github.com/aprice/freenote/users"
)

type DecoratedNote struct {
	Links Links `json:"_links" xml:"Links>Link"`
	notes.Note
	XMLName struct{} `json:"-" xml:"Note"`
}

func DecorateNote(note notes.Note, canWrite bool, baseURI string) DecoratedNote {
	links := Links{}
	links.RecordRUD(fmt.Sprintf("%s/users/%s/notes/%s", baseURI, note.Owner, note.ID), canWrite)
	links.Add(Link{
		Rel:  "author",
		Href: fmt.Sprintf("%s/users/%s", baseURI, note.Owner),
	})
	return DecoratedNote{Note: note, Links: links}
}

type DecoratedNotes struct {
	Links   Links        `json:"_links" xml:"Links>Link"`
	Notes   []notes.Note `json:"notes" xml:"Page>Note"`
	XMLName struct{}     `json:"-" xml:"Notes"`
}

func DecorateNotes(owner users.User, values []notes.Note, folder string, page page.Page, canWrite bool, baseURI string) DecoratedNotes {
	links := Links{}
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
	}
	links.CollectionCR(base, page, canWrite)
	return DecoratedNotes{Notes: values, Links: links}
}
