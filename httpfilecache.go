package httpfilecache

import (
	"bufio"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"time"
)

// Package httpfilecache wraps a http.RoundTripper and stores all successful
// requests on disk returning them until the expires TTL is reached and a new
// request is sent with it's response again stored on disk.

type RoundTripper struct {
	CacheDir  string
	TTL       time.Duration
	Transform RequestTransform
	Methods   []string
	http.RoundTripper
}

func (c RoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {

	if !inStringSlice(c.Methods, req.Method) {
		return c.RoundTripper.RoundTrip(req)
	}

	filenamepath := filepath.Join(c.CacheDir, c.Transform(req))

	// If the file has not expired
	fileInfo, _ := os.Stat(filenamepath)
	if fileInfo != nil && fileInfo.ModTime().Add(c.TTL).After(time.Now()) {
		if file, err := os.Open(filenamepath); err == nil {
			return http.ReadResponse(bufio.NewReader(file), req)
		}
	}

	resp, err := c.RoundTripper.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	// Only on success do we save the result
	if resp.StatusCode == 200 {
		body, err := httputil.DumpResponse(resp, true)
		if err == nil {
			if err = os.MkdirAll(filepath.Dir(filenamepath), 0755); err != nil {
				return nil, err
			}
			err = os.WriteFile(filenamepath, body, 0600)
			if err != nil {
				return nil, err
			}
		}
	}

	return resp, err
}

// NewClient returns an *http.Client that caches responses.
func NewClient(cacheDir string, ttl time.Duration) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()

	return &http.Client{
		Transport: RoundTripper{
			cacheDir,
			ttl,
			URLToFilepath,
			[]string{"GET", "HEAD"}, // mutations (POST, PUT, etc..) are generally not cacheable
			transport,
		},
		Timeout: 60 * time.Second,
	}
}

func inStringSlice(values []string, value string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}
