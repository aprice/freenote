package page

import (
	"net/url"
	"testing"
)

func TestFromQueryString(t *testing.T) {
	t.Run("all", func(tt *testing.T) {
		u, err := url.Parse("http://localhost/test?start=20&length=5&sort=field&order=desc")
		if err != nil {
			tt.Error(err)
		}
		p := FromQueryString(u, []string{"field"}, false)
		expected := Page{
			Start:          20,
			Length:         5,
			SortBy:         "field",
			SortDescending: true,
		}
		if p != expected {
			tt.Errorf("URL: %s\n\tGot %#v\n\tExpected %#v", u, p, expected)
		}
	})
	t.Run("none", func(tt *testing.T) {
		u, err := url.Parse("http://localhost/test")
		if err != nil {
			tt.Error(err)
		}
		p := FromQueryString(u, []string{"field"}, false)
		expected := Page{
			Start:          0,
			Length:         10,
			SortBy:         "field",
			SortDescending: false,
		}
		if p != expected {
			tt.Errorf("URL: %s\n\tGot %#v\n\tExpected %#v", u, p, expected)
		}
	})
}
