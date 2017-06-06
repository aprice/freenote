package rest

import (
	"fmt"
	"strings"

	"github.com/lunny/html2md"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"

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
	for i := range values {
		idx := strings.Index(values[i].Body, "\n")
		if idx > 0 {
			values[i].Body = values[i].Body[:idx]
		}
	}
	links.CollectionCR(base, page, canWrite)
	return decoratedNotes{Notes: values, Links: links}
}

var bfRender blackfriday.Renderer
var bfExt = 0 |
	blackfriday.EXTENSION_NO_INTRA_EMPHASIS |
	blackfriday.EXTENSION_TABLES |
	blackfriday.EXTENSION_FENCED_CODE |
	blackfriday.EXTENSION_AUTOLINK |
	blackfriday.EXTENSION_STRIKETHROUGH |
	blackfriday.EXTENSION_SPACE_HEADERS |
	blackfriday.EXTENSION_AUTO_HEADER_IDS |
	blackfriday.EXTENSION_BACKSLASH_LINE_BREAK |
	blackfriday.EXTENSION_DEFINITION_LISTS

func init() {
	var bfFlags = 0 |
		blackfriday.HTML_NOFOLLOW_LINKS |
		blackfriday.HTML_SAFELINK |
		blackfriday.HTML_HREF_TARGET_BLANK |
		blackfriday.HTML_FOOTNOTE_RETURN_LINKS
	bfRender = blackfriday.HtmlRenderer(bfFlags, "", "")
}

func ensureMarkdownBody(note *notes.Note, p *bluemonday.Policy) {
	if note.Body == "" && note.HTMLBody != "" {
		note.Body = html2md.Convert(note.HTMLBody)
	}
}

func ensureHTMLBody(note *notes.Note, p *bluemonday.Policy) {
	if note.HTMLBody == "" && note.Body != "" {
		raw := blackfriday.MarkdownOptions([]byte(note.Body), bfRender, blackfriday.Options{Extensions: bfExt})
		note.HTMLBody = string(p.SanitizeBytes(raw))
	}
}
