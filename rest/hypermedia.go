package rest

import (
	"encoding/xml"
	"strings"

	"github.com/aprice/freenote/page"
)

// Link is a hypermedia/hyperdata related link.
type Link struct {
	Rel     string   `json:"-" xml:"rel,attr"`
	Href    string   `json:"href" xml:"href,attr"`
	Method  string   `json:"method,omitempty" xml:"method,attr,omitempty"`
	XMLName struct{} `json:"-" xml:"link"`
}

// Links represents a collection of Links on an entity.
type Links map[string]Link

// Add a Link to this collection of Links.
func (l Links) Add(link Link) {
	l[link.Rel] = link
}

// Canonical creates a canonical link.
func (l Links) Canonical(uri string) {
	l.Add(Link{
		Rel:    "canonical",
		Href:   uri,
		Method: "GET",
	})
}

// Edit creates an edit link.
func (l Links) Edit(uri string) {
	l.Add(Link{
		Rel:    "edit",
		Href:   uri,
		Method: "PATCH",
	})
}

// Save creates a save link.
func (l Links) Save(uri string) {
	l.Add(Link{
		Rel:    "save",
		Href:   uri,
		Method: "PUT",
	})
}

// Delete creates a delete link.
func (l Links) Delete(uri string) {
	l.Add(Link{
		Rel:    "delete",
		Href:   uri,
		Method: "DELETE",
	})
}

// RecordRUD creates the record retrieve, update, and delete links.
func (l Links) RecordRUD(uri string, canWrite bool) {
	l.Canonical(uri)
	if canWrite {
		l.Edit(uri)
		l.Save(uri)
		l.Delete(uri)
	}
}

// Next creates a next page link.
func (l Links) Next(base string, curPage page.Page) {
	curPage.Start += curPage.Length
	l.Add(Link{
		Rel:    "next",
		Href:   AppendQueryString(base, curPage.QueryString()),
		Method: "GET",
	})
}

// Previous creates a previous page link.
func (l Links) Previous(base string, curPage page.Page) {
	curPage.Start = curPage.Start - curPage.Length
	if curPage.Start < 0 {
		curPage.Start = 0
	}
	l.Add(Link{
		Rel:    "previous",
		Href:   AppendQueryString(base, curPage.QueryString()),
		Method: "GET",
	})
}

// Create creates a create link. Create create.
func (l Links) Create(uri string) {
	l.Add(Link{
		Rel:    "create",
		Href:   uri,
		Method: "POST",
	})
}

// CollectionCR creates the collection create and retrieve links.
func (l Links) CollectionCR(base string, curPage page.Page, canWrite bool) {
	l.Canonical(AppendQueryString(base, curPage.QueryString()))
	if curPage.Start > 0 {
		l.Previous(base, curPage)
	}
	if curPage.HasMore {
		l.Next(base, curPage)
	}
	if canWrite {
		l.Create(base)
	}
}

// MarshalXML fulfills xml.Marshaler
func (l Links) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	linkList := make([]Link, 0, len(l))
	for _, link := range l {
		linkList = append(linkList, link)
	}
	return e.EncodeElement(linkList, start)
}

// AppendQueryString adds a field to a query string.
func AppendQueryString(base, query string) string {
	if strings.Contains("?", base) {
		return base + "&" + query
	}
	return base + "?" + query
}
