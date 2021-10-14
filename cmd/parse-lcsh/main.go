// parse-lcsh is a command-line tool to parse the Library of Congress `lcsh.both.ndjson` file and out CSV-encoded
// subject heading ID and (English) label data.
package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/sfomuseum/go-csvdict"
	"github.com/sfomuseum/go-libraryofcongress/walk"
	"github.com/tidwall/gjson"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func main() {

	flag.Parse()

	uris := flag.Args()
	ctx := context.Background()

	w, err := walk.NewNDJSONWalker(ctx, "ndjson://")

	if err != nil {
		log.Fatalf("Failed to create walker, %v", err)
	}

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

	seen := new(sync.Map)
	
	cb_func := walkCallbackFunc(csv_wr, seen)

	err = w.WalkURIs(ctx, cb_func, uris...)

	if err != nil {
		log.Fatalf("Failed to walk LCSH data, %v", err)
	}
}

func walkCallbackFunc(csv_wr *csvdict.Writer, seen *sync.Map) walk.WalkCallbackFunction {

	fn := func(ctx context.Context, body []byte) error {

		rsp := gjson.GetBytes(body, "@graph")

		if !rsp.Exists() {
			return fmt.Errorf("Record is missing @graph property")
		}

		for _, item := range rsp.Array() {

			id_rsp := item.Get("@id")
			id := id_rsp.String()

			if !strings.HasPrefix(id, "http://id.loc.gov/authorities/subjects/") {
				continue
			}

			sh_id := filepath.Base(id)

			label_rsp := item.Get("madsrdf:authoritativeLabel.@value")
			label := label_rsp.String()

			if label == "" {
				continue
			}

			_, loaded := seen.LoadOrStore(sh_id, true)

			if loaded {
				continue
			}

			out := map[string]string{
				"id":    sh_id,
				"label": label,
			}

			err := csv_wr.WriteRow(out)

			if err != nil {
				return fmt.Errorf("Failed to write %s (%s), %v", id, label, err)
			}
		}

		csv_wr.Flush()
		return nil
	}

	return fn
}
