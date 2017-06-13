package stringset

type sentinel struct{}

var nothing = sentinel{}

type StringSet map[string]sentinel

func New() StringSet {
	return StringSet{}
}

func (ss StringSet) Add(in ...string) {
	for _, str := range in {
		ss[str] = nothing
	}
}

func (ss StringSet) Contains(in string) bool {
	_, ok := ss[in]
	return ok
}

func (ss StringSet) Count() int {
	return len(ss)
}
