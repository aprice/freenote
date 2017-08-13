//+build !dev,!debug

package server

import "net/http"

func doDebug(w http.ResponseWriter, r *http.Request) {
	statusResponse(w, http.StatusNotFound)
}
