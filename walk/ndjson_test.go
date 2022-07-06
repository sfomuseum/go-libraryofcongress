package walk

import (
	"context"
	"path/filepath"
	"testing"
	"sync/atomic"
)

func TestNDJSONWalker(t *testing.T) {

	ctx := context.Background()

	paths := map[string]int32{
		"../fixtures/lcsh.sample.ndjson": int32(3),
		"../fixtures/lcsh.sample.ndjson.zip": int32(2),
	}

	for rel_path, expected_count := range paths {

		abs_path, err := filepath.Abs(rel_path)

		if err != nil {
			t.Fatalf("Failed to derive absolute path for %s, %v", rel_path, err)
		}

		w, err := NewWalker(ctx, "ndjson://")

		if err != nil {
			t.Fatalf("Failed to create new walker, %v", err)
		}

		count := int32(0)

		cb := func(ctx context.Context, body []byte) error {
			atomic.AddInt32(&count, 1)
			return nil
		}

		err = w.WalkURIs(ctx, cb, abs_path)

		if err != nil {
			t.Fatalf("Failed to walk %s, %v", abs_path, err)
		}

		if count != expected_count {
			t.Fatalf("Unexpected count for %s: %d", abs_path, count)
		}
	}
}
