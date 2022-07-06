package walk

import (
	"context"
	"testing"
)

func TestRegisterWalker(t *testing.T) {

	ctx := context.Background()

	err := RegisterWalker(ctx, "ndjson", NewNDJSONWalker)

	if err == nil {
		t.Fatalf("Expected ndjson walker to be registered.")
	}
}

func TestWalker(t *testing.T) {

	ctx := context.Background()

	_, err := NewWalker(ctx, "ndjson://")

	if err != nil {
		t.Fatalf("Failed to create ndjson walker, %v", err)
	}
}
