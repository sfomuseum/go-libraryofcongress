package walk

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

type Reader struct {
	io.ReadCloser
	io.ReaderAt
	R io.ReadCloser
	N int64
}

func NewReader(r io.ReadCloser) *Reader {
	return &Reader{R: r}
}

func (r *Reader) ReadAt(p []byte, off int64) (n int, err error) {
	if off < r.N {
		return 0, errors.New("invalid offset")
	}
	diff := off - r.N
	written, err := io.CopyN(ioutil.Discard, r.R, diff)
	r.N += written
	if err != nil {
		return 0, err
	}

	n, err = r.R.Read(p)
	r.N += int64(n)
	return
}

func (r *Reader) Read(p []byte) (n int, err error) {
	return r.R.Read(p)
}

func (r *Reader) Close() error {
	return r.R.Close()
}

// OpenURI opens 'uri' a and returns an `io.Reader` and the size of the file. If 'uri' is
// prefixed with "https://" then the body of the file will be retrieved via an HTTP GET request.
func OpenURI(ctx context.Context, uri string) (*Reader, int64, error) {

	var r io.ReadCloser
	var sz int64

	u, err := url.Parse(uri)

	if err != nil {
		return nil, 0, fmt.Errorf("Failed to parse URI for '%s', %w", uri, err)
	}

	switch u.Scheme {
	case "http", "https":

		rsp, err := http.Get(uri)

		if err != nil {
			return nil, 0, fmt.Errorf("Failed to retrieve %s, %w", uri, err)
		}

		r = rsp.Body
		sz = rsp.ContentLength

	default:

		fh, err := os.Open(uri)

		if err != nil {
			return nil, 0, fmt.Errorf("Failed to open %s, %v", uri, err)
		}

		info, _ := os.Stat(uri)

		r = fh
		sz = info.Size()
	}

	return NewReader(r), sz, nil
}
