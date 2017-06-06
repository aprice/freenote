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

func (p *Page) FromQueryString(u *url.URL, sortFields []string) {
	var (
		i   int
		err error
	)
	if i, err = strconv.Atoi(u.Query().Get("start")); err == nil {
		p.Start = i
	}

	if i, err = strconv.Atoi(u.Query().Get("length")); err == nil && i >= 0 {
		p.Length = i
	}
	p.SortBy = sortFields[0]
	sort := u.Query().Get("sort")
	for _, field := range sortFields {
		if strings.EqualFold(sort, field) {
			p.SortBy = field
			break
		}
	}
	if strings.HasPrefix(u.Query().Get("order"), "d") {
		p.SortDescending = true
	} else if strings.HasPrefix(u.Query().Get("order"), "a") {
		p.SortDescending = false
	}
}

func FromQueryString(u *url.URL, sortFields []string, defaultDescending bool) Page {
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
	} else if strings.HasPrefix(u.Query().Get("order"), "a") {
		p.SortDescending = false
	} else {
		p.SortDescending = defaultDescending
	}
	return p
}
