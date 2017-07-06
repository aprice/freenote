package client

import (
	"io"
	"io/ioutil"
	"net/http"
)

// CleanupResponse handles emptying & closing the request body, for use with
// defer after making a request. Can be used without error checking first.
func CleanupResponse(res *http.Response) {
	// nolint: gas
	if res != nil && res.Body != nil {
		io.Copy(ioutil.Discard, res.Body)
		res.Body.Close()
	}
}
