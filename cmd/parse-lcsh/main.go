// parse-lcsh is a command-line tool to parse the Library of Congress `lcsh.both.ndjson` file and out CSV-encoded
// subject heading ID and (English) label data. It can also be configured to include broader concepts for each heading
// as well as Wikidata and Worldcat concordances.
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

	broader := flag.Bool("include-broader", false, "If present, include a comma-separated list of `skos:broader` pointers associated with each subject heading")

	wikidata := flag.Bool("include-wikidata", false, "If present, include a Wikidata pointer associated with each subject heading")

	worldcat := flag.Bool("include-worldcat", false, "If present, include a Worldcat pointer associated with each subject heading")

	concordances := flag.Bool("include-concordances", false, "If true will enable the -include-wikidata and -include-worldcat flags")

	all := flag.Bool("include-all", false, "If true will enable all the other -include-* flags")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "parse-lcsh is a command-line tool to parse the Library of Congress `lcsh.both.ndjson` file and out CSV-encoded subject heading ID and (English) label data. It can also be configured to include broader concepts for each heading as well as Wikidata and Worldcat concordances.\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n\t %s [options] lcsh.both.ndjson\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Valid options are:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *concordances {
		*wikidata = true
		*worldcat = true
	}

	if *all {
		*broader = true
		*wikidata = true
		*worldcat = true
	}

	uris := flag.Args()
	ctx := context.Background()

	w, err := walk.NewWalker(ctx, "ndjson://")

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

	if *broader {
		fieldnames = append(fieldnames, "broader")
	}

	if *wikidata {
		fieldnames = append(fieldnames, "wikidata_id")
	}

	if *worldcat {
		fieldnames = append(fieldnames, "worldcat_id")
	}

	csv_wr, err := csvdict.NewWriter(mw, fieldnames)

	if err != nil {
		log.Fatalf("Failed to create CSV writer, %v", err)
	}

	csv_wr.WriteHeader()

	seen := new(sync.Map)

	cb_func := walkCallbackFunc(csv_wr, seen, fieldnames)

	err = w.WalkURIs(ctx, cb_func, uris...)

	if err != nil {
		log.Fatalf("Failed to walk LCSH data, %v", err)
	}
}

func walkCallbackFunc(csv_wr *csvdict.Writer, seen *sync.Map, fieldnames []string) walk.WalkCallbackFunction {

	capture := make(map[string]bool)

	for _, k := range fieldnames {
		capture[k] = true
	}

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

			_, capture_broader := capture["broader"]
			_, capture_wikidata := capture["wikidata_id"]
			_, capture_worldcat := capture["worldcat_id"]

			capture_concordances := false

			if capture_wikidata {
				capture_concordances = true
				out["wikidata_id"] = ""
			}

			if capture_worldcat {
				capture_concordances = true
				out["worldcat_id"] = ""
			}

			if capture_concordances {

				external_rsp := item.Get("madsrdf:hasCloseExternalAuthority")

				if external_rsp.Exists() {

					for _, e := range external_rsp.Array() {

						id := e.Get("@id").String()

						if strings.HasPrefix(id, "http://www.wikidata.org/entity/") && capture_wikidata {
							out["wikidata_id"] = filepath.Base(id)
						}

						if strings.HasPrefix(id, "http://id.worldcat.org/fast/") && capture_worldcat {
							out["worldcat_id"] = filepath.Base(id)
						}
					}
				}

			}

			if capture_broader {

				out["broader"] = ""

				broader_rsp := item.Get("skos:broader")

				if broader_rsp.Exists() {

					others := make([]string, 0)

					for _, b := range broader_rsp.Array() {

						other := b.Get("@id").String()

						if !strings.HasPrefix(other, "http://id.loc.gov/authorities/subjects/") {
							continue
						}

						sh_other := filepath.Base(other)
						others = append(others, sh_other)
					}

					out["broader"] = strings.Join(others, ",")
				}
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
