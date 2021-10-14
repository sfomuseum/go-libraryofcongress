// parse-lcnaf is a command-line tool to parse the Library of Congress `lcnaf.both.ndjson` (or `lcnaf.both.ndjson.zip`)
// file and output CSV-encoded subject heading ID and (English) label data.
package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/sfomuseum/go-csvdict"
	"github.com/sfomuseum/go-libraryofcongress"
	"github.com/sfomuseum/go-libraryofcongress/walk"
	"github.com/tidwall/gjson"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {

	flag.Parse()

	uris := flag.Args()
	ctx := context.Background()

	w, err := walk.NewWalker(ctx, "ndjson://")

	if err != nil {
		log.Fatalf("Failed to create walker, %v", err)
	}

	catalog, err := libraryofcongress.NewCatalog(ctx, "tmp://")

	if err != nil {
		log.Fatalf("Failed to create catalog, %v", err)
	}

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

	cb_func := walkCallbackFunc(csv_wr, catalog)

	err = w.WalkURIs(ctx, cb_func, uris...)

	if err != nil {
		log.Fatalf("Failed to walk LCSH data, %v", err)
	}
}

func walkCallbackFunc(csv_wr *csvdict.Writer, catalog *libraryofcongress.Catalog) walk.WalkCallbackFunction {

	fn := func(ctx context.Context, body []byte) error {

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

	return fn
}
