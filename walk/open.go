package walk

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/jeffallen/seekinghttp"
)

// OpenURI opens 'uri' a and returns an `io.Reader` and the size of the file. If 'uri' is
// prefixed with "https://" then the body of the file will be retrieved via an HTTP GET request.
func OpenURI(ctx context.Context, uri string) (WalkReader, int64, error) {

	var r WalkReader
	var sz int64

	u, err := url.Parse(uri)

	if err != nil {
		return nil, 0, fmt.Errorf("Failed to parse URI for '%s', %w", uri, err)
	}

	switch u.Scheme {
	case "http", "https":

		req := seekinghttp.New(uri)

		s, err := req.Size()

		if err != nil {
			return nil, 0, err
		}

		r = &RemoteWalkReader{remote: req}
		sz = s

	default:

		fh, err := os.Open(uri)

		if err != nil {
			return nil, 0, fmt.Errorf("Failed to open %s, %v", uri, err)
		}

		info, _ := os.Stat(uri)

		r = &LocalWalkReader{local: fh}
		sz = info.Size()
	}

	return r, sz, nil
}
