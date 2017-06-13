package rest

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/aprice/freenote/stringset"
)

var supportedResponseTypes = []string{
	"application/json",
	"application/javascript",
	"application/xml",
	"text/xml",
}

var errUnsupportedMediaType = errors.New("Unspported Media Type")

func parseRequest(r *http.Request, payload interface{}) error {
	switch r.Header.Get("Content-Type") {
	case "application/json":
		return json.NewDecoder(r.Body).Decode(payload)
	case "application/xml", "text/xml":
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
	if len(types) == 0 {
		return ""
	}
	typeSet := stringset.New()
	typeSet.Add(types...)
	ah := r.Header.Get("Accept")
	if ah == "" || ah == "*/*" {
		return types[0]
	}
	rtypes := strings.Split(ah, ",")
	var (
		weight, q float64
		parts     []string
		ctype     string
		err       error
	)

	for _, t := range rtypes {
		parts = strings.Split(t, ";")
		q = 1.0
		if len(parts) > 1 && strings.HasPrefix(parts[1], "q=") {
			t = parts[0]
			q, err = strconv.ParseFloat(strings.TrimPrefix(parts[1], "q="), 64)
			if err != nil {
				q = 1.0
			}
		}
		if q < weight {
			continue
		}
		t = strings.TrimSpace(t)
		if t == "*/*" {
			weight = q
			ctype = types[0]
		}
		if typeSet.Contains(t) {
			weight = q
			ctype = t
		}
	}
	return ctype
}
