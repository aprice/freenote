package rest

import (
	"fmt"

	"github.com/aprice/freenote/config"
	"github.com/aprice/freenote/notes"
	"github.com/aprice/freenote/page"
	"github.com/aprice/freenote/users"
)

type decoratedNote struct {
	Links Links `json:"_links" xml:"Links>Link"`
	notes.Note
	XMLName struct{} `json:"-" xml:"Note"`
}

func decorateNote(note notes.Note, canWrite bool, conf config.Config) decoratedNote {
	links := Links{}
	links.RecordRUD(fmt.Sprintf("%s/users/%s/notes/%s", conf.BaseURI, note.Owner, note.ID), canWrite)
	links.Add(Link{
		Rel:  "author",
		Href: fmt.Sprintf("%s/users/%s", conf.BaseURI, note.Owner),
	})
	return decoratedNote{Note: note, Links: links}
}

type decoratedNotes struct {
	Links   Links        `json:"_links" xml:"Links>Link"`
	Notes   []notes.Note `json:"notes" xml:"Page>Note"`
	XMLName struct{}     `json:"-" xml:"Notes"`
}

func decorateNotes(owner users.User, values []notes.Note, folder string, page page.Page, canWrite bool, conf config.Config) decoratedNotes {
	links := Links{}
	var base string
	if folder == "" {
		base = fmt.Sprintf("%s/users/%s/notes", conf.BaseURI, owner.ID)
	} else {
		base = fmt.Sprintf("%s/users/%s/notes/%s", conf.BaseURI, owner.ID, folder)
	}
	links.CollectionCR(base, page, canWrite)
	return decoratedNotes{Notes: values, Links: links}
}
