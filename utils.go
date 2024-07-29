package prouter

import (
	"encoding/json"
	"io"
	"mime"
	"net/http"
	
	"github.com/pkg/errors"
)

func matchesContentType(contentType, expectedType string) error {
	mimetype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return errors.Wrapf(err, "malformed Content-Type header (%s)", contentType)
	}
	if mimetype != expectedType {
		return errors.Errorf("unsupported Content-Type header (%s): must be '%s'", contentType, expectedType)
	}
	return nil
}

// CheckForJSON makes sure that the request's Content-Type is application/json.
func CheckForJSON(r *http.Request) error {
	ct := r.Header.Get("Content-Type")
	
	// No Content-Type header is ok as long as there's no Body
	if ct == "" && (r.Body == nil || r.ContentLength == 0) {
		return nil
	}
	
	// Otherwise it better be json
	return matchesContentType(ct, "application/json")
}

func ReadJSON(r *http.Request, out interface{}) error {
	err := CheckForJSON(r)
	if err != nil {
		return err
	}
	if r.Body == nil || r.ContentLength == 0 {
		return nil
	}
	
	dec := json.NewDecoder(r.Body)
	err = dec.Decode(out)
	defer r.Body.Close()
	if err != nil {
		if err == io.EOF {
			return errors.New("invalid JSON: got EOF while reading request body")
		}
		return errors.Wrap(err, "invalid JSON")
	}
	
	if dec.More() {
		return errors.New("unexpected content after JSON")
	}
	return nil
}

// WriteJSON writes the value v to the http response stream as json with standard json encoding.
func WriteJSON(w http.ResponseWriter, code int, v interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}
