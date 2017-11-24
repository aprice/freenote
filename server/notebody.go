package server

import (
	"strings"

	"github.com/lunny/html2md"
	"github.com/microcosm-cc/bluemonday"
	blackfriday "gopkg.in/russross/blackfriday.v2"

	"github.com/aprice/freenote/notes"
)

var bfRender blackfriday.Renderer
var bfExt = 0 |
	blackfriday.NoIntraEmphasis |
	blackfriday.Tables |
	blackfriday.FencedCode |
	blackfriday.Autolink |
	blackfriday.Strikethrough |
	blackfriday.SpaceHeadings |
	blackfriday.HeadingIDs |
	blackfriday.AutoHeadingIDs |
	blackfriday.BackslashLineBreak |
	blackfriday.DefinitionLists

func init() {
	var bfFlags = 0 |
		blackfriday.UseXHTML |
		blackfriday.NofollowLinks |
		blackfriday.Safelink |
		blackfriday.HrefTargetBlank |
		blackfriday.FootnoteReturnLinks

	bfRender = blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
		Flags: bfFlags,
	})

	html2md.AddRule("pre", &html2md.Rule{
		Patterns: []string{"pre"},
		Replacement: func(innerHTML string, attrs []string) string {
			body := strings.TrimSpace(innerHTML)
			if !strings.HasPrefix(body, "<code>") {
				return innerHTML
			}
			body = strings.TrimPrefix(body, "<code>")
			body = strings.TrimSuffix(body, "</code>")
			return "\n\n```" + body + "```\n"
		},
	})
}

func ensureMarkdownBody(note *notes.Note, p *bluemonday.Policy) {
	if note.Body == "" && note.HTMLBody != "" {
		note.Body = html2md.Convert(note.HTMLBody)
	}
}

func ensureHTMLBody(note *notes.Note, p *bluemonday.Policy) {
	if note.HTMLBody == "" && note.Body != "" {
		raw := blackfriday.Run([]byte(note.Body), blackfriday.WithRenderer(bfRender), blackfriday.WithExtensions(bfExt))
		note.HTMLBody = string(p.SanitizeBytes(raw))
	}
}
