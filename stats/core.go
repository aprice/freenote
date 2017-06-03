package stats

import (
	"expvar"
	"runtime"
	"time"
)

func init() {
	expvar.Publish("goroutines", expvar.Func(func() interface{} {
		return runtime.NumGoroutine()
	}))
}

type Operation struct {
	Name    string
	Elapsed time.Duration
	Error   bool
}
