package rest

import "testing"

func TestFirstSegment(t *testing.T) {
	testCases := []struct {
		in  string
		out string
	}{
		{"/", "/"},
		{"/users", "/users"},
		{"/notes/abcd1234", "/notes"},
	}

	for _, test := range testCases {
		actual := firstSegment(test.in)
		if actual != test.out {
			t.Errorf("firstSegment(%q)=%q, expected %q", test.in, actual, test.out)
		}
	}
}
