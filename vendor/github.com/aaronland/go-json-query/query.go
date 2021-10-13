package query

import (
	"context"
	"github.com/tidwall/gjson"
	_ "log"
	"regexp"
)

// QUERYSET_MODE_ANY is a flag to signal that only one match in a QuerySet needs to be successful.
const QUERYSET_MODE_ANY string = "ANY"

// QUERYSET_MODE_ALL is a flag to signal that only all matches in a QuerySet needs to be successful.
const QUERYSET_MODE_ALL string = "ALL"

// QuerySet is a struct containing one or more Query instances and flags for how the results of those queries should be interpreted.
type QuerySet struct {
	// A set of Query instances
	Queries []*Query
	// A string flag representing how query results should be interpreted.
	Mode string
}

// Query is an atomic query to perform against a JSON document.
type Query struct {
	// A valid tidwall/gjson query path.
	Path string
	// A valid regular expression.
	Match *regexp.Regexp
}

// Matches compares the set of queries in 'qs' against a JSON record ('body') and returns true or false depending on whether or not some or all of those queries are matched successfully.
func Matches(ctx context.Context, qs *QuerySet, body []byte) (bool, error) {

	select {
	case <-ctx.Done():
		return false, nil
	default:
		// pass
	}

	queries := qs.Queries
	mode := qs.Mode

	tests := len(queries)
	matches := 0

	for _, q := range queries {

		rsp := gjson.GetBytes(body, q.Path)

		if !rsp.Exists() {

			if mode == QUERYSET_MODE_ALL {
				break
			}
		}

		for _, r := range rsp.Array() {

			if q.Match.MatchString(r.String()) {

				matches += 1

				if mode == QUERYSET_MODE_ANY {
					break
				}
			}
		}

		if mode == QUERYSET_MODE_ANY && matches > 0 {
			break
		}

	}

	if mode == QUERYSET_MODE_ALL {

		if matches < tests {
			return false, nil
		}
	}

	if matches == 0 {
		return false, nil
	}

	return true, nil
}
