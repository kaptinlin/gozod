package types_test

import (
	"testing"

	z "github.com/kaptinlin/gozod"
	"github.com/kaptinlin/gozod/core"
)

// TestDescribe tests the Describe method on ZodString
func TestDescribe(t *testing.T) {
	desc := "A valid user ID"
	schema := z.String().Describe(desc)

	// Validation should passes
	if _, err := schema.Parse("user123"); err != nil {
		t.Errorf("Validation failed: %v", err)
	}

	// Verify metadata registration
	// We check if the schema instance is registered in the GlobalRegistry
	// Note: schema returned by Describe is the one registered

	meta, ok := core.GlobalRegistry.Get(schema)
	if !ok {
		t.Fatal("Metadata not found in GlobalRegistry")
	}

	if meta.Description != desc {
		t.Errorf("Description = %q, want %q", meta.Description, desc)
	}
}

// TestMeta tests the Meta method on ZodString
func TestMeta(t *testing.T) {
	metaData := core.GlobalMeta{
		ID:          "meta-test-id",
		Title:       "Meta Test Title",
		Description: "Meta description",
		Examples:    []any{"example1", "example2"},
	}

	schema := z.String().Meta(metaData)

	if _, err := schema.Parse("valid"); err != nil {
		t.Errorf("Parse failed: %v", err)
	}

	registered, ok := core.GlobalRegistry.Get(schema)
	if !ok {
		t.Fatal("Metadata not found")
	}

	if registered.ID != metaData.ID {
		t.Errorf("ID = %q, want %q", registered.ID, metaData.ID)
	}
	if registered.Title != metaData.Title {
		t.Errorf("Title = %q, want %q", registered.Title, metaData.Title)
	}
	if registered.Description != metaData.Description {
		t.Errorf("Description = %q, want %q", registered.Description, metaData.Description)
	}
	if len(registered.Examples) != 2 {
		t.Errorf("Examples len = %d, want 2", len(registered.Examples))
	}
}
