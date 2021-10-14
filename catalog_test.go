package libraryofcongress

import (
	"context"
	"testing"
)

func TestCatalog(t *testing.T) {

	ctx := context.Background()

	c, err := NewCatalog(ctx, "tmp://")

	if err != nil {
		t.Fatalf("Failed to create catalog, %v", err)
	}

	keys := []string{
		"test",
		"test2",
	}

	for _, k := range keys {

		err = c.Store(ctx, k)

		if err != nil {
			t.Fatalf("Failed to store '%s', %v", k, err)
		}

		exists, err := c.Exists(ctx, k)

		if err != nil {
			t.Fatalf("Failed to determine whether '%s' exists, %v", k, err)
		}

		if !exists {
			t.Fatalf("Expected '%s' to exist", k)
		}

	}

	for _, k := range keys {

		exists, err := c.ExistsOrStore(ctx, k)

		if err != nil {
			t.Fatalf("Failed to determine whether '%s' exists (or store), %v", k, err)
		}

		if !exists {
			t.Fatalf("Expected '%s' to exist", k)
		}
	}

	err = c.Close(ctx)

	if err != nil {
		t.Fatalf("Failed to close catalog, %v", err)
	}

}
