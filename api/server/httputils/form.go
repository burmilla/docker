package httputils

import (
	"errors"
	"net/http"
	"path/filepath"
	"strings"
)

// BoolValue transforms a form value in different formats into a boolean type.
func BoolValue(r *http.Request, k string) bool {
	s := strings.ToLower(strings.TrimSpace(r.FormValue(k)))
	return !(s == "" || s == "0" || s == "no" || s == "false" || s == "none")
}

// BoolValueOrDefault returns the default bool passed if the query param is
// missing, otherwise it's just a proxy to boolValue above.
func BoolValueOrDefault(r *http.Request, k string, d bool) bool {
	if _, ok := r.Form[k]; !ok {
		return d
	}
	return BoolValue(r, k)
}

// ArchiveOptions stores archive information for different operations.
type ArchiveOptions struct {
	Name string
	Path string
}

// ArchiveFormValues parses form values and turns them into ArchiveOptions.
// It fails if the archive name and path are not in the request.
func ArchiveFormValues(r *http.Request, vars map[string]string) (ArchiveOptions, error) {
	if err := ParseForm(r); err != nil {
		return ArchiveOptions{}, err
	}

	name := vars["name"]
	path := filepath.FromSlash(r.Form.Get("path"))

	switch {
	case name == "":
		return ArchiveOptions{}, errors.New("bad parameter: 'name' cannot be empty")
	case path == "":
		return ArchiveOptions{}, errors.New("bad parameter: 'path' cannot be empty")
	}

	return ArchiveOptions{name, path}, nil
}
