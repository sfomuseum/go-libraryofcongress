// parse-lcnaf is a command-line tool to parse the Library of Congress `lcnaf.both.ndjson` (or `lcnaf.both.ndjson.zip`)
// file and output CSV-encoded subject heading ID and (English) label data.
package main

// Please reconcile this code with cmd/parse-lcsh. Most of it is identical.

import (
	"archive/zip"
	"context"
	"flag"
	"fmt"
	"github.com/aaronland/go-jsonl/walk"
	"github.com/sfomuseum/go-csvdict"
	"github.com/sfomuseum/go-libraryofcongress"
	"github.com/tidwall/gjson"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// This is used to prevent duplicate entries

var catalog *libraryofcongress.Catalog

func main() {

	flag.Parse()

	uris := flag.Args()
	ctx := context.Background()

	c, err := libraryofcongress.NewCatalog(ctx, "tmp://")

	if err != nil {
		log.Fatalf("Failed to create catalog, %v", err)
	}

	catalog = c
	defer catalog.Close(ctx)

	writers := []io.Writer{
		os.Stdout,
	}

	mw := io.MultiWriter(writers...)

	fieldnames := []string{
		"id",
		"label",
	}

	csv_wr, err := csvdict.NewWriter(mw, fieldnames)

	if err != nil {
		log.Fatalf("Failed to create CSV writer, %v", err)
	}

	csv_wr.WriteHeader()

	for _, uri := range uris {

		ext := filepath.Ext(uri)

		switch ext {
		case ".zip":
			err = walkZipFile(ctx, uri, csv_wr)
		default:
			err = walkFile(ctx, uri, csv_wr)
		}
	}

}

func walkFile(ctx context.Context, uri string, csv_wr *csvdict.Writer) error {

	fh, err := os.Open(uri)

	if err != nil {
		fmt.Errorf("Failed to open %s, %v", uri, err)
	}

	defer fh.Close()

	err = walkReader(ctx, fh, csv_wr)

	if err != nil {
		return fmt.Errorf("Failed to walk %s, %v", uri, err)
	}

	return nil
}

func walkZipFile(ctx context.Context, uri string, csv_wr *csvdict.Writer) error {

	fh, err := os.Open(uri)

	if err != nil {
		fmt.Errorf("Failed to open %s, %v", uri, err)
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

		err = walkReader(ctx, zip_fh, csv_wr)

		if err != nil {
			return fmt.Errorf("Failed to walk %s, %v", uri, err)
		}
	}

	return nil
}

func walkReader(ctx context.Context, r io.Reader, csv_wr *csvdict.Writer) error {

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var walk_err error

	record_ch := make(chan *walk.WalkRecord)
	error_ch := make(chan *walk.WalkError)
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

				err := parseRecord(ctx, csv_wr, r.Body)

				if err != nil {
					error_ch <- &walk.WalkError{
						Path:       r.Path,
						LineNumber: r.LineNumber,
						Err:        fmt.Errorf("Failed to index feature, %w", err),
					}
				}
			}
		}
	}()

	walk_opts := &walk.WalkOptions{
		RecordChannel: record_ch,
		ErrorChannel:  error_ch,
		Workers:       100,
	}

	walk.WalkReader(ctx, walk_opts, r)

	<-done_ch

	if walk_err != nil && !walk.IsEOFError(walk_err) {
		return fmt.Errorf("Failed to walk document, %v", walk_err)
	}

	return nil
}

func parseRecord(ctx context.Context, csv_wr *csvdict.Writer, body []byte) error {

	rsp := gjson.GetBytes(body, "@graph")

	if !rsp.Exists() {
		return fmt.Errorf("Record is missing @graph property")
	}

	for _, item := range rsp.Array() {

		id_rsp := item.Get("@id")
		id := id_rsp.String()

		if !strings.HasPrefix(id, "http://id.loc.gov/authorities/names/") {
			continue
		}

		sh_id := filepath.Base(id)

		label_rsp := item.Get("madsrdf:authoritativeLabel")
		label := label_rsp.String()

		if label == "" {
			continue
		}

		exists, err := catalog.ExistsOrStore(ctx, sh_id)

		if err != nil {
			return fmt.Errorf("Failed to determine whether %s exists, %w", sh_id, err)
		}

		if exists {
			continue
		}

		out := map[string]string{
			"id":    sh_id,
			"label": label,
		}

		err = csv_wr.WriteRow(out)

		if err != nil {
			return fmt.Errorf("Failed to write %s (%s), %v", id, label, err)
		}
	}

	csv_wr.Flush()
	return nil
}
