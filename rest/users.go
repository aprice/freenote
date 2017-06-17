package rest

import (
	"fmt"

	"github.com/aprice/freenote/page"
	"github.com/aprice/freenote/users"
)

type decoratedUser struct {
	Links Links `json:"_links" xml:"Links>Link"`
	users.User
	XMLName struct{} `json:"-" xml:"User"`
}

func decorateUser(user users.User, canReadNotes, canWrite bool, baseURI string) decoratedUser {
	links := Links{}
	links.RecordRUD(fmt.Sprintf("%s/users/%s", baseURI, user.ID), canWrite)
	if canReadNotes {
		links.Add(Link{
			Rel:    "notes",
			Method: "GET",
			Href:   fmt.Sprintf("%s/users/%s/notes", baseURI, user.ID),
		})
	}
	user.Password = nil
	user.Sessions = nil
	return decoratedUser{User: user, Links: links}
}

type decoratedUsers struct {
	Links   Links        `json:"_links" xml:"Links>Link"`
	Users   []users.User `json:"users" xml:"Page>User"`
	XMLName struct{}     `json:"-" xml:"Users"`
}

func decorateUsers(values []users.User, page page.Page, canWrite bool, baseURI string) decoratedUsers {
	for i := range values {
		values[i].Password = nil
		values[i].Sessions = nil
	}
	links := Links{}
	links.CollectionCR(fmt.Sprintf("%s/users", baseURI), page, canWrite)
	return decoratedUsers{Users: values, Links: links}
}
