package httpfilecache

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

type Response struct {
	Response *http.Response
	Error    error
}

// ReplayCachedRequests emits all cached responses in the given directory tree
func ReplayCachedRequests(ctx context.Context, dir string) chan Response {
	ch := make(chan Response)

	go func() {
		defer close(ch)
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Deadline exceeded?
			if ctx.Err() != nil {
				return ctx.Err()
			}

			if info.IsDir() {
				return nil
			}

			if file, err := os.Open(path); err == nil {
				res, err := http.ReadResponse(bufio.NewReader(file), nil)
				if err != nil {
					return fmt.Errorf("%s: %w", path, err)
				}
				ch <- Response{Response: res}
			}

			return nil
		})

		if err != nil {
			ch <- Response{Error: err}
			// Don't tell them about this?
			// if !errors.Is(err, context.DeadlineExceeded) {
			// 	ch <- Response{Error: err}
			// }
		}
	}()

	return ch
}
