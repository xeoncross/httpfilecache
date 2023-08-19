package httpfilecache

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"testing"
	"time"
)

// var tempTestDir string = "tmptest"

func TestReplayCachedRequests(t *testing.T) {

	tempTestDir, err := os.MkdirTemp("", "example")
	if err != nil {
		t.Fatal(err)
	}

	number := 3
	_, err = createTempFiles(tempTestDir, number)
	if err != nil {
		t.Fatal(err)
	}

	// fmt.Printf("Looking for files in %s\n", tempTestDir)

	ctx, cancelFn := context.WithTimeout(context.Background(), time.Second*1)
	defer cancelFn()
	ch := ReplayCachedRequests(ctx, tempTestDir)

	for response := range ch {
		if response.Error != nil {
			t.Fatal(response.Error)
		}
		_, err := io.ReadAll(response.Response.Body)
		if err != nil {
			t.Fatal(err)
		}

		// fmt.Printf("response: %s\n", string(b))
		number--
	}

	if number != 0 {
		t.Errorf("failed to replay all cached responses: %d", number)
	}

	if err := os.RemoveAll(tempTestDir); err != nil {
		t.Fatal(err)
	}

}

func createTempFiles(dir string, number int) ([]string, error) {

	files := []string{}

	for i := 0; i < number; i++ {
		f, err := os.CreateTemp(dir, "example_request")
		if err != nil {
			return nil, err
		}

		resp := &http.Response{
			StatusCode: 200,
			Request:    &http.Request{},
			Body:       io.NopCloser(strings.NewReader(fmt.Sprintf("content for %d", i))),
		}

		body, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", f.Name(), err)
		}

		if _, err := f.Write([]byte(body)); err != nil {
			return nil, fmt.Errorf("%s: %w", f.Name(), err)
		}
		if err := f.Close(); err != nil {
			return nil, fmt.Errorf("%s: %w", f.Name(), err)
		}

		files = append(files, f.Name())
	}

	return files, nil
}
