package walk

import (
	"archive/zip"
	"context"
	"fmt"
	jsonl_walk "github.com/aaronland/go-jsonl/walk"
	"io"
	"os"
	"path/filepath"
)

type NDJSONWalker struct {
	Walker
}

func NewNDJSONWalker(ctx context.Context, uri string) (Walker, error) {
	w := &NDJSONWalker{}
	return w, nil
}

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

func (w *NDJSONWalker) WalkFile(ctx context.Context, cb WalkCallbackFunction, uri string) error {

	fh, err := os.Open(uri)

	if err != nil {
		return fmt.Errorf("Failed to open %s, %v", uri, err)
	}

	defer fh.Close()

	err = w.WalkReader(ctx, cb, fh)

	if err != nil {
		return fmt.Errorf("Failed to walk %s, %v", uri, err)
	}

	return nil
}

func (w *NDJSONWalker) WalkZipFile(ctx context.Context, cb WalkCallbackFunction, uri string) error {

	fh, err := os.Open(uri)

	if err != nil {
		return fmt.Errorf("Failed to open %s, %v", uri, err)
	}

	defer fh.Close()

	info, _ := os.Stat(uri)

	r, err := zip.NewReader(fh, info.Size())

	if err != nil {
		return fmt.Errorf("Failed to create zip reader for %s, %v", uri, err)
	}

	for _, f := range r.File {

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
			case err := <-error_ch:
				walk_err = err
				done_ch <- true
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
		Workers:       100,
	}

	jsonl_walk.WalkReader(ctx, walk_opts, r)

	<-done_ch

	if walk_err != nil && !jsonl_walk.IsEOFError(walk_err) {
		return fmt.Errorf("Failed to walk document, %v", walk_err)
	}

	return nil
}
