//+build !dev,!debug

package stats

// Record an operation execution
func Record(op Operation) {
}

// Record an operation failing
func RecordError(name string) {
}

// Measure an operation. Easiest used as `defer Measure("name")()`.
func Measure(names ...string) func() {
	return func() {}
}

func Run() func() {
	return func() {}
}
