//+build !dev,!debug

package rest

import "net/http"

func doDebug(w http.ResponseWriter, r *http.Request) {
	statusResponse(w, http.StatusNotFound)
}
