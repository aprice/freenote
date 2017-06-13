package rest

import (
	"encoding/xml"
	"strings"

	"github.com/aprice/freenote/page"
)

type Link struct {
	Rel     string   `json:"-" xml:"rel,attr"`
	Href    string   `json:"href" xml:"href,attr"`
	Method  string   `json:"method,omitempty" xml:"method,attr,omitempty"`
	XMLName struct{} `json:"-" xml:"link"`
}

func (l Links) Add(link Link) {
	l[link.Rel] = link
}

func (l Links) Canonical(uri string) {
	l.Add(Link{
		Rel:    "canonical",
		Href:   uri,
		Method: "GET",
	})
}

func (l Links) Edit(uri string) {
	l.Add(Link{
		Rel:    "edit",
		Href:   uri,
		Method: "PATCH",
	})
}

func (l Links) Save(uri string) {
	l.Add(Link{
		Rel:    "save",
		Href:   uri,
		Method: "PUT",
	})
}

func (l Links) Delete(uri string) {
	l.Add(Link{
		Rel:    "delete",
		Href:   uri,
		Method: "DELETE",
	})
}

func (l Links) RecordRUD(uri string, canWrite bool) {
	l.Canonical(uri)
	if canWrite {
		l.Edit(uri)
		l.Save(uri)
		l.Delete(uri)
	}
}

func (l Links) Next(base string, curPage page.Page) {
	curPage.Start += curPage.Length
	l.Add(Link{
		Rel:    "next",
		Href:   AppendQueryString(base, curPage.QueryString()),
		Method: "GET",
	})
}

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

func (l Links) Create(uri string) {
	l.Add(Link{
		Rel:    "create",
		Href:   uri,
		Method: "POST",
	})
}

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

type Links map[string]Link

func (l Links) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	linkList := make([]Link, 0, len(l))
	for _, link := range l {
		linkList = append(linkList, link)
	}
	return e.EncodeElement(linkList, start)
}

func AppendQueryString(base, query string) string {
	if strings.Contains("?", base) {
		return base + "&" + query
	}
	return base + "?" + query
}
