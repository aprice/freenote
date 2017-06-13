package rest

import (
	"net/http/httptest"
	"testing"
)

func TestNegotiateType(t *testing.T) {
	cases := []struct {
		accept string
		ctype  string
	}{
		{"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8", "application/xml"},
		{"application/json", "application/json"},
		{"*/*", "application/json"},
	}

	for _, test := range cases {
		r := httptest.NewRequest("GET", "/", nil)
		r.Header.Set("Accept", test.accept)
		got := negotiateType(supportedResponseTypes, r)
		if got != test.ctype {
			t.Errorf("For %q, got %q, expected %q", test.accept, got, test.ctype)
		}
	}
}
