package httpfilecache

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
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

	dir, err := ioutil.TempDir("", "httpfilecache")
	if err != nil {
		t.Fatalf("failed to create test directory: %s", err)
	}
	// t.Log(dir)
	defer os.RemoveAll(dir)

	os.Setenv("XDG_CACHE_HOME", dir)
	client := Client()

	// Non-cached result
	resp, err := client.Get(server.URL + "/404")
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 404 {
		t.Fatalf("want: 404, got: %d\n", resp.StatusCode)
	}

	path := URLToFilepath(resp.Request.URL)
	_, err = os.Stat(filepath.Join(dir, "httpfilecache", path))
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

	path = URLToFilepath(resp.Request.URL)
	_, err = os.Stat(filepath.Join(dir, "httpfilecache", path))
	if err != nil {
		t.Fatal(err)
	}
}
