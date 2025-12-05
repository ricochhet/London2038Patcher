package dlutil

import (
	"context"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/ricochhet/london2038patcher/pkg/errutil"
)

type HTTPClient struct {
	*http.Client
}

// NewHTTPClient returns a HttpClient struct.
func NewHTTPClient(timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		Client: &http.Client{
			Timeout: timeout,
		},
	}
}

// Download downloads a file from a URL into the specified path.
func (c *HTTPClient) Download(ctx context.Context, path, url string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return errutil.WithFrame(err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return errutil.WithFrame(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errutil.WithFramef("HTTP error: %s", resp.Status)
	}

	tmp := path + ".tmp"

	out, err := os.Create(tmp)
	if err != nil {
		return errutil.WithFrame(err)
	}

	if _, err := io.Copy(out, resp.Body); err != nil {
		out.Close()
		return errutil.WithFrame(err)
	}

	if err := out.Close(); err != nil {
		return errutil.WithFrame(err)
	}

	return os.Rename(tmp, path)
}
