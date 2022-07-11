package walk

import (
	"archive/zip"
	"context"
	"fmt"
	jsonl_walk "github.com/aaronland/go-jsonl/walk"
	"io"
	"net/url"
	"path/filepath"
	"strconv"
)

// type NDJSONWalker implements the `Walker` interface for NDJSON files.
type NDJSONWalker struct {
	Walker
	// workers is the maximum number of simultaneous workers for processing NDJSON files
	workers int
}

func init() {
	ctx := context.Background()
	RegisterWalker(ctx, "ndjson", NewNDJSONWalker)
}

// NewNDJSONWalker creates a new instance that implements the `Walker` interface for NDJSON files configured
// by 'uri' which is expected to take the form of:
//
//	ndjson://?{PARAMETERS}
//
// Where {PARAMETERS} may be:
// * `?workers=` The number of maximum simultaneous workers for processing NDJSON records. Default is 100.
func NewNDJSONWalker(ctx context.Context, uri string) (Walker, error) {

	max_workers := 100

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	q := u.Query()

	str_workers := q.Get("workers")

	if str_workers != "" {

		w, err := strconv.Atoi(str_workers)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse 'workers' parameter, %w", err)
		}

		max_workers = w
	}

	w := &NDJSONWalker{
		workers: max_workers,
	}

	return w, nil
}

// WalkURIs() processes 'uris' dispatching each record to 'cb'. 'uris' is expected to be a list of compressed ('.zip')
// or uncompressed files on disk.
func (w *NDJSONWalker) WalkURIs(ctx context.Context, cb WalkCallbackFunction, uris ...string) error {

	for _, uri := range uris {

		select {
		case <-ctx.Done():
			return nil
		default:
			// pass
		}

		ext := filepath.Ext(uri)

		var err error

		switch ext {
		case ".zip":
			err = w.WalkZipFile(ctx, cb, uri)
		default:
			err = w.WalkFile(ctx, cb, uri)
		}

		if err != nil {
			return fmt.Errorf("Failed to walk %s, %w", uri, err)
		}
	}

	return nil
}

// WalkFile() processes 'uri' dispatch each record to 'cb'.
func (w *NDJSONWalker) WalkFile(ctx context.Context, cb WalkCallbackFunction, uri string) error {

	r, _, err := OpenURI(ctx, uri)

	if err != nil {
		return fmt.Errorf("Failed to open %s, %v", uri, err)
	}

	defer r.Close()

	err = w.WalkReader(ctx, cb, r)

	if err != nil {
		return fmt.Errorf("Failed to walk %s, %v", uri, err)
	}

	return nil
}

// WalkZipFile() decompresses 'uri' and processes each file (contained in the zip archive) dispatching each record to 'cb'.
func (w *NDJSONWalker) WalkZipFile(ctx context.Context, cb WalkCallbackFunction, uri string) error {

	r, sz, err := OpenURI(ctx, uri)

	if err != nil {
		return fmt.Errorf("Failed to open %s, %v", uri, err)
	}

	defer r.Close()

	zr, err := zip.NewReader(r, sz)

	if err != nil {
		return fmt.Errorf("Failed to create zip reader for %s, %v", uri, err)
	}

	for _, f := range zr.File {

		zip_fh, err := f.Open()

		if err != nil {
			return fmt.Errorf("Failed to open %s, %v", f.Name, err)
		}

		defer zip_fh.Close()

		err = w.WalkReader(ctx, cb, zip_fh)

		if err != nil {
			return fmt.Errorf("Failed to walk %s, %v", uri, err)
		}
	}

	return nil
}

// WalkReader() processes each record in 'r' (which is expected to a line-separate JSON document) and dispatches each record to 'cb'.
func (w *NDJSONWalker) WalkReader(ctx context.Context, cb WalkCallbackFunction, r io.Reader) error {

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var walk_err error

	record_ch := make(chan *jsonl_walk.WalkRecord)
	error_ch := make(chan *jsonl_walk.WalkError)
	done_ch := make(chan bool)

	go func() {

		for {
			select {
			case <-ctx.Done():
				done_ch <- true
				return
			case <-done_ch:
				return
			case err := <-error_ch:
				walk_err = err
				done_ch <- true
				return
			case r := <-record_ch:

				err := cb(ctx, r.Body)

				if err != nil {
					error_ch <- &jsonl_walk.WalkError{
						Path:       r.Path,
						LineNumber: r.LineNumber,
						Err:        fmt.Errorf("Failed to index feature, %w", err),
					}
				}
			}
		}
	}()

	walk_opts := &jsonl_walk.WalkOptions{
		RecordChannel: record_ch,
		ErrorChannel:  error_ch,
		DoneChannel:   done_ch,
		Workers:       w.workers,
	}

	jsonl_walk.WalkReader(ctx, walk_opts, r)

	if walk_err != nil && !jsonl_walk.IsEOFError(walk_err) {
		return fmt.Errorf("Failed to walk document, %v", walk_err)
	}

	return nil
}
