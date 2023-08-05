package httpfilecache

import (
	"net/http"
	"path/filepath"
	"strings"
	"unicode"
)

// RequestTransform to map the request to the filesystem
type RequestTransform func(req *http.Request) string

// URLToFilepath maps URLs to logical file locations in nested folders.
// www.example.com/file-path-here?query=here
// becomes
// com/example/www/file-path-here-query-here
func URLToFilepath(req *http.Request) string {

	var paths []string
	for _, part := range strings.Split(req.URL.Host, ".") {
		paths = append([]string{part}, paths...)
	}

	paths = append(paths, Sanitize(req.URL.RequestURI()))
	return filepath.Join(paths...)
}

// Sanitize path replacing multiple forbidden runes with a single "-"
func Sanitize(s string) string {
	var seen bool
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '.' {
			seen = false
			return r
		}
		if seen {
			return -1
		}
		seen = true
		return '-'
	}, s)
}
