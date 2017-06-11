package rest

import (
	"net/http/httptest"
	"testing"
)

func TestFirstSegment(t *testing.T) {
	testCases := []struct {
		in  string
		out string
	}{
		{"/", ""},
		{"/users", "users"},
		{"/notes/abcd1234", "notes"},
	}

	for _, test := range testCases {
		actual := firstSegment(test.in)
		if actual != test.out {
			t.Errorf("firstSegment(%q)=%q, expected %q", test.in, actual, test.out)
		}
	}
}

func TestPopSegment(t *testing.T) {
	testCases := []struct {
		inPath  string
		seg     string
		outPath string
	}{
		{"/", "", "/"},
		{"/users", "users", "/"},
		{"/notes/abcd1234", "notes", "/abcd1234"},
	}

	for _, test := range testCases {
		r := httptest.NewRequest("GET", test.inPath, nil)
		actual := popSegment(r)
		if actual != test.seg {
			t.Errorf("popSegment(%q)=%q, expected %q", test.inPath, actual, test.seg)
		}
		if r.URL.Path != test.outPath {
			t.Errorf("popSegment(%q), r.URL.Path=%q, expected %q", test.inPath, r.URL.Path, test.outPath)
		}
	}
}
