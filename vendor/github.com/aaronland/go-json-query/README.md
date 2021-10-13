# go-json-query

Go package for querying and filter JSON documents using tidwall/gjson-style paths and regular expressions for testing values.

## Documentation

[![Go Reference](https://pkg.go.dev/badge/github.com/aaronland/go-json-query.svg)](https://pkg.go.dev/github.com/aaronland/go-json-query)

## Important

Documentation is incomplete.

## Example

```
import (
	"context"
	"flag"
	"fmt"
	"github.com/aaronland/go-json-query"
	"io"
	"os"
	"strings"
)

func main() {

	var queries query.QueryFlags
	flag.Var(&queries, "query", "One or more {PATH}={REGEXP} parameters for filtering records.")

	valid_modes := strings.Join([]string{query.QUERYSET_MODE_ALL, query.QUERYSET_MODE_ANY}, ", ")
	desc_modes := fmt.Sprintf("Specify how query filtering should be evaluated. Valid modes are: %s", valid_modes)

	query_mode := flag.String("query-mode", query.QUERYSET_MODE_ALL, desc_modes)

	flag.Parse()

	paths := flag.Args()

	qs := &query.QuerySet{
		Queries: queries,
		Mode:    *query_mode,
	}

	ctx := context.Background()

	for _, path := range paths {

		fh, _ := os.Open(path)
		defer fh.Close()

		body, _ := io.ReadAll(fh)

		matches, _ := query.Matches(ctx, qs, body)

		fmt.Printf("%s\t%t\n", path, matches)
	}
}
```

## See also

* https://github.com/tidwall/gjson