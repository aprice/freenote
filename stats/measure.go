//+build dev debug

package stats

import (
	"expvar"
	"time"
)

var opMap = make(map[string]opStats)

type opStats struct {
	count     *expvar.Int
	avgTime   *expvar.Int
	totalTime time.Duration
	errors    *expvar.Int
}

// Record an operation execution
func Record(op Operation) {
	opChan <- op
}

// Record an operation failing
func RecordError(name string) {
	opChan <- Operation{
		Name:  name,
		Error: true,
	}
}

// Measure an operation. Easiest used as `defer Measure("name")()`.
func Measure(name string) func() {
	start := time.Now()
	return func() {
		opChan <- Operation{
			Name:    name,
			Elapsed: time.Since(start),
		}
	}
}

var opChan = make(chan Operation, 10)

type sentinel struct{}

var nothing = sentinel{}

func Run() func() {
	stop := make(chan sentinel)
	go func() {
		var (
			entry opStats
			ok    bool
			pfx   string
		)
		for {
			select {
			case <-stop:
				return
			case op := <-opChan:
				if entry, ok = opMap[op.Name]; !ok {
					pfx = "op." + op.Name + "."
					entry = opStats{
						expvar.NewInt(pfx + "count"),
						expvar.NewInt(pfx + "average"),
						0,
						expvar.NewInt(pfx + "errors"),
					}
					opMap[op.Name] = entry
				}
				entry.count.Add(1)
				if op.Elapsed > 0 {
					entry.totalTime = entry.totalTime + op.Elapsed
				}
				if op.Error {
					entry.errors.Add(1)
				}
				entry.avgTime.Set(int64(entry.totalTime/time.Millisecond) / entry.count.Value())
			}
		}
	}()
	return func() { stop <- nothing }
}
