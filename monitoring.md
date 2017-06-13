# Monitoring & Profiling

`expvar` and `pprof` routes are enabled if either the `dev` or `debug` build tags are set.

## expvar
Example expvarmon:

`expvarmon -ports=9990 -vars="goroutines,mem:memstats.Alloc,mem:memstats.Sys,mem:memstats.HeapInuse,mem:memstats.HeapIdle,mem:memstats.StackInuse,duration:memstats.PauseNs,duration:memstats.PauseTotalNs,memstats.NumGC,duration:op_req_notes_get.AvgTime,duration:op_req_note_get.AvgTime,duration:op_req_session_get.AvgTime,"`

You can add monitors for different request handlers in the form `op_req_HANDLER_METHOD` where `HANDLER` is one of `user`, `users`, `note`, `notes`, or `session`, and `METHOD` is the HTTP request verb. Each tracks `Count` (number of requests), `Errors` (failed requests), `TotalTime` (total execution time) and `AvgTime` (`TotalTime / Count`).

## pprof

Profiling is available at the normal route, `/debug/pprof`.

## Logging

Log messages are output using the standard logging library, and sent to standard output.
