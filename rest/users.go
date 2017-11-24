package rest

import (
	"fmt"

	"github.com/aprice/freenote/page"
	"github.com/aprice/freenote/users"
)

// DecoratedUser represents a user with hypermedia links.
type DecoratedUser struct {
	Links Links `json:"_links" xml:"Links>Link"`
	users.User
	XMLName struct{} `json:"-" xml:"User"`
}

// DecorateUser decorates a User with hypermedia links.
func DecorateUser(user users.User, canReadNotes, canWrite bool, baseURI string) DecoratedUser {
	links := Links{}
	links.RecordRUD(fmt.Sprintf("%s/users/%s", baseURI, user.ID), canWrite)
	if canReadNotes {
		links.Add(Link{
			Rel:    "notes",
			Method: "GET",
			Href:   fmt.Sprintf("%s/users/%s/notes", baseURI, user.ID),
		})
	}
	if canWrite {
		links.Add(Link{
			Rel:    "password",
			Method: "PUT",
			Href:   fmt.Sprintf("%s/users/%s/password", baseURI, user.ID),
		})
		links.Add(Link{
			Rel:    "createnote",
			Method: "POST",
			Href:   fmt.Sprintf("%s/users/%s/notes", baseURI, user.ID),
		})
	}
	user.Password = nil
	user.Sessions = nil
	return DecoratedUser{User: user, Links: links}
}

// DecoratedUsers represents a collection of Users decorated with hypermedia
// links for the collection.
type DecoratedUsers struct {
	Links   Links        `json:"_links" xml:"Links>Link"`
	Users   []users.User `json:"users" xml:"Page>User"`
	XMLName struct{}     `json:"-" xml:"Users"`
}

// DecorateUsers decorates a collection of Users with hypermedia links for the
// collection.
func DecorateUsers(values []users.User, page page.Page, canWrite bool, baseURI string) DecoratedUsers {
	for i := range values {
		values[i].Password = nil
		values[i].Sessions = nil
	}
	links := Links{}
	links.CollectionCR(fmt.Sprintf("%s/users", baseURI), page, canWrite)
	return DecoratedUsers{Users: values, Links: links}
}
