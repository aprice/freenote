package server

import (
	"github.com/lunny/html2md"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"

	"github.com/aprice/freenote/notes"
)

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
