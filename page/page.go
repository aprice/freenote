package page

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type Page struct {
	Start          int    `json:"start"`
	Length         int    `json:"legth"`
	HasMore        bool   `json:"-"`
	SortDescending bool   `json:"-"`
	SortBy         string `json:"-"`
}

func (p Page) QueryString() string {
	var direction string
	if p.SortDescending {
		direction = "desc"
	} else {
		direction = "asc"
	}
	return fmt.Sprintf("start=%d&length=%d&sort=%s&order=%s",
		p.Start, p.Length,
		url.QueryEscape(p.SortBy),
		url.QueryEscape(direction))
}

func FromQueryString(u *url.URL, sortFields []string) Page {
	p := Page{}
	i, err := strconv.Atoi(u.Query().Get("start"))
	if err == nil {
		p.Start = i
	}
	i, err = strconv.Atoi(u.Query().Get("length"))
	if err == nil {
		p.Length = i
	}
	if p.Length <= 0 {
		p.Length = 10
	}
	p.SortBy = sortFields[0]
	sort := u.Query().Get("sort")
	for _, field := range sortFields {
		if sort == field {
			p.SortBy = field
			break
		}
	}
	if strings.HasPrefix(u.Query().Get("order"), "d") {
		p.SortDescending = true
	}
	return p
}
