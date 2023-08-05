package httpfilecache

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCache(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/404":
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		default:
			w.Write([]byte("Hello World"))
		}
	}))

	defer server.Close()

	dir, err := os.MkdirTemp("", "httpfilecache")
	if err != nil {
		t.Fatalf("failed to create test directory: %s", err)
	}
	// t.Log(dir)
	defer os.RemoveAll(dir)

	client := NewClient(dir, time.Second*10)

	// Non-cached result
	resp, err := client.Get(server.URL + "/404")
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 404 {
		t.Fatalf("want: 404, got: %d\n", resp.StatusCode)
	}

	path := URLToFilepath(resp.Request)
	_, err = os.Stat(filepath.Join(dir, path))
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatal(err)
	}

	// Cached result
	resp, err = client.Get(server.URL + "/")
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("want: 200, got: %d\n", resp.StatusCode)
	}

	path = URLToFilepath(resp.Request)
	_, err = os.Stat(filepath.Join(dir, path))
	if err != nil {
		t.Fatal(err)
	}
}

func TestSanitize(t *testing.T) {
	testCases := []struct {
		input  string
		output string
	}{
		{
			input:  "",
			output: "",
		},
		{
			input:  "//",
			output: "-",
		},
		{
			input:  "path?",
			output: "path-",
		},
		{
			input:  "file-path-here?query=here",
			output: "file-path-here-query-here",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.output, func(t *testing.T) {
			if Sanitize(tc.input) != tc.output {
				t.Logf("%q != %q\n", Sanitize(tc.input), tc.output)
				t.Fail()
			}
		})
	}
}
