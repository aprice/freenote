//+build dev debug

package rest

import (
	"expvar"
	"net/http"
	"net/http/pprof"
	"strings"
)

func doDebug(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/debug/vars") {
		expvar.Handler().ServeHTTP(w, r)
	} else if r.URL.Path == "/debug/pprof/" {
		pprof.Index(w, r)
	} else if strings.HasPrefix(r.URL.Path, "/debug/pprof/") {
		pprof.Handler(r.URL.Path[13:]).ServeHTTP(w, r)
	}
}
