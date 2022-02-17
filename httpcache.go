/*
Package httpfilecache wraps a http.RoundTripper and stores all successful GET and
HEAD requests on disk forever. This is mostly for testing to prevent hitting the
same endpoints repeatedly while prototyping or running integration tests.
The cache is located at ~/.cache/httpfilecache (using XDG_CACHE_HOME for ~/.cache
if set). Individual responses must be deleted if a fresher copy is needed.
*/
package httpfilecache

import (
	"bufio"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// URLToFilepath maps URLs to logical file locations in nested folders
// [tld] / [domain.sub1.www] / [file-path-here]
func URLToFilepath(u *url.URL) string {

	var paths []string
	for _, part := range strings.Split(u.Host, ".") {
		paths = append([]string{part}, paths...)
	}
	tld := paths[0]
	host := strings.Join(paths[1:], ".")

	return filepath.Join(tld, host, sanitize(u.Path))
}

// sanitize path replacing multiple forbidden runes with a single "-"
func sanitize(s string) string {
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

// GetCacheDir gets an XDG compliant directory for storing cached responses
// This can be changed by setting os.Setenv("XDG_CACHE_HOME", dir)
func GetCacheDir() string {
	cacheRoot := os.ExpandEnv("$HOME/.cache")
	if xdgCache, ok := os.LookupEnv("XDG_CACHE_HOME"); ok {
		cacheRoot = xdgCache
	}
	return filepath.Join(cacheRoot, "httpfilecache")
}

type roundTripper struct {
	http.RoundTripper
	cacheDir string
}

func (c roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {

	// Mutations (POST, PUT, etc..) are not cacheable
	if req.Method != "GET" && req.Method != "HEAD" {
		return c.RoundTripper.RoundTrip(req)
	}

	path := filepath.Join(c.cacheDir, URLToFilepath(req.URL))

	if file, err := os.Open(path); err == nil {
		return http.ReadResponse(bufio.NewReader(file), req)
	}
	resp, err := c.RoundTripper.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	// Only on success do we save the result
	if resp.StatusCode == 200 {
		body, err := httputil.DumpResponse(resp, true)
		if err == nil {
			if err := os.MkdirAll(filepath.Dir(path), 0770); err != nil {
				return nil, err
			}
			err = ioutil.WriteFile(path, body, 0600)
		}
	}

	return resp, err
}

// Wrap creates a new caching http.RoundTripper that uses the given RoundTripper
// to fetch responses that don't yet exist in the cache.
func Wrap(transport http.RoundTripper) http.RoundTripper {
	return roundTripper{transport, GetCacheDir()}
}

// Client returns an *http.Client that caches responses.
func Client() *http.Client {
	return &http.Client{Transport: Wrap(http.DefaultTransport)}
}
