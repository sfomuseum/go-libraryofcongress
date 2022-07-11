package walk

import (
	"context"
	"path/filepath"
	"testing"
)

func TestOpenURI(t *testing.T) {

	ctx := context.Background()

	local_tests := map[string]int64{
		"../fixtures/lcsh.sample.ndjson":     16318,
		"../fixtures/lcsh.sample.ndjson.zip": 3118,
	}

	remote_tests := map[string]int64{
		"https://id.loc.gov/download/lcsh.both.ndjson.zip": 323177552,
	}

	for rel_path, expected_sz := range local_tests {

		abs_path, err := filepath.Abs(rel_path)

		if err != nil {
			t.Fatalf("Failed to derive absolute path for '%s', %v", rel_path, err)
		}

		r, sz, err := OpenURI(ctx, abs_path)

		if err != nil {
			t.Fatalf("Failed to open '%s', %v", abs_path, err)
		}

		defer r.Close()

		if sz != expected_sz {
			t.Fatalf("Unexpected file size for %s: %d", abs_path, sz)
		}
	}

	for uri, expected_sz := range remote_tests {

		r, sz, err := OpenURI(ctx, uri)

		if err != nil {
			t.Fatalf("Failed to open '%s', %v", uri, err)
		}

		defer r.Close()

		if sz != expected_sz {
			t.Fatalf("Unexpected file size for %s: %d", uri, sz)
		}
	}

}
