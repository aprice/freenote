package stringset

type sentinel struct{}

var nothing = sentinel{}

// StringSet maintains a set of unique strings.
type StringSet map[string]sentinel

// New creates a new StringSet.
func New() StringSet {
	return StringSet{}
}

// Add a string or strings to the set.
func (ss StringSet) Add(in ...string) {
	for _, str := range in {
		ss[str] = nothing
	}
}

// Contains returns true if this set contains the given string.
func (ss StringSet) Contains(in string) bool {
	_, ok := ss[in]
	return ok
}

// Count returns the number of unique strings in this set.
func (ss StringSet) Count() int {
	return len(ss)
}
