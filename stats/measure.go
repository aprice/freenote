//+build dev debug

package stats

import (
	"encoding/json"
	"expvar"
	"strings"
	"sync"
	"time"
)

var opMap = make(map[string]*opStats)
var mu = sync.Mutex{}

type opStats struct {
	Count     int32
	AvgTime   time.Duration
	TotalTime time.Duration
	Errors    int32
}

func (s opStats) String() string {
	mu.Lock()
	defer mu.Unlock()
	json, _ := json.Marshal(&s)
	return string(json)
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
func Measure(names ...string) func() {
	name := strings.ToLower(strings.Join(names, "_"))
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
			entry *opStats
			ok    bool
		)
		for {
			select {
			case <-stop:
				return
			case op := <-opChan:
				mu.Lock()
				if entry, ok = opMap[op.Name]; !ok {
					name := "op_" + op.Name
					entry = new(opStats)
					opMap[op.Name] = entry
					expvar.Publish(name, entry)
				}
				if op.Elapsed > 0 {
					entry.Count++
					entry.TotalTime += op.Elapsed
					entry.AvgTime = time.Duration(entry.TotalTime.Nanoseconds() / int64(entry.Count))
				}
				if op.Error {
					entry.Errors++
				}
				mu.Unlock()
			}
		}
	}()
	return func() { stop <- nothing }
}
