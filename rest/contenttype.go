package rest

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"net/http"
	"strings"
)

var supportedResponseTypes = []string{"application/json", "application/javascript", "application/xml", "text/xml", "text/html"}

var errUnsupportedMediaType = errors.New("Unspported Media Type")

func parseRequest(r *http.Request, payload interface{}) error {
	switch r.Header.Get("Content-Type") {
	case "application/json":
		return json.NewDecoder(r.Body).Decode(payload)
	case "application/xml":
		return xml.NewDecoder(r.Body).Decode(payload)
	default:
		return errUnsupportedMediaType
	}
}

func sendResponse(w http.ResponseWriter, r *http.Request, payload interface{}, status int) {
	ctype := negotiateType(supportedResponseTypes, r)

	switch ctype {
	case "application/json":
		body, err := json.Marshal(payload)
		if err != nil {
			http.Error(w, "Error Writing JSON: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Type", ctype)
		w.WriteHeader(status)
		w.Write(body)
	case "application/javascript":
		callback := r.URL.Query().Get("cb")
		if callback == "" {
			callback = "JSONP"
		}
		body, err := json.Marshal(payload)
		if err != nil {
			http.Error(w, "Error Writing JSON: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Type", ctype)
		w.WriteHeader(status)
		w.Write([]byte(callback))
		w.Write([]byte("("))
		w.Write(body)
		w.Write([]byte(");"))
	case "application/xml", "text/xml":
		body, err := xml.Marshal(payload)
		if err != nil {
			http.Error(w, "Error Writing XML: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Type", ctype)
		w.WriteHeader(status)
		w.Write(body)
	//TODO: case "text/html":
	//TODO: execute simple HTML template based on payload type
	default:
		http.Error(w, "Not Acceptable", http.StatusNotAcceptable)
	}
}

func negotiateType(types []string, r *http.Request) string {
	ah := r.Header.Get("Accept")
	if ah == "" || ah == "*/*" {
		return types[0]
	}
	rtypes := strings.Split(ah, ",")
	for _, t := range rtypes {
		if semi := strings.Index(t, ";"); semi > 0 {
			t = t[0:semi]
		}
		t = strings.TrimSpace(t)
		for _, s := range types {
			if t == s {
				return s
			}
		}
	}
	return ""
}
