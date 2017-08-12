package rest

import (
	"testing"
)

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
		rh := &requestHandler{path: test.inPath}
		actual := rh.popSegment()
		if actual != test.seg {
			t.Errorf("popSegment(%q)=%q, expected %q", test.inPath, actual, test.seg)
		}
		if rh.path != test.outPath {
			t.Errorf("popSegment(%q), r.URL.Path=%q, expected %q", test.inPath, rh.path, test.outPath)
		}
	}
}
