package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type registrySchema struct {
	internals *ZodTypeInternals
}

func newRegistrySchema() *registrySchema {
	return &registrySchema{internals: &ZodTypeInternals{}}
}

func (s *registrySchema) ParseAny(input any, ctx ...*ParseContext) (any, error) {
	return input, nil
}

func (s *registrySchema) Internals() *ZodTypeInternals {
	return s.internals
}

func (s *registrySchema) IsOptional() bool {
	return s.internals.IsOptional()
}

func (s *registrySchema) IsNilable() bool {
	return s.internals.IsNilable()
}

func TestRegistry_CRUDAndEarlyRangeStop(t *testing.T) {
	t.Parallel()

	first := newRegistrySchema()
	second := newRegistrySchema()
	registry := NewRegistry[string]()

	assert.Same(t, registry, registry.Add(first, "first"))
	registry.Add(second, "second")

	got, ok := registry.Get(first)
	require.True(t, ok)
	assert.Equal(t, "first", got)
	assert.True(t, registry.Has(second))

	seen := 0
	registry.Range(func(schema ZodSchema, meta string) bool {
		seen++
		return false
	})
	assert.Equal(t, 1, seen)

	assert.Same(t, registry, registry.Remove(first))
	assert.False(t, registry.Has(first))
}

func TestRegistry_RangeCallbackCanMutateRegistry(t *testing.T) {
	t.Parallel()

	first := newRegistrySchema()
	added := newRegistrySchema()
	registry := NewRegistry[string]().Add(first, "first")
	done := make(chan struct{})
	var gotAdded, nestedRange bool

	go func() {
		defer close(done)
		registry.Range(func(schema ZodSchema, _ string) bool {
			registry.Remove(schema).Add(added, "added")
			_, gotAdded = registry.Get(added)
			registry.Range(func(ZodSchema, string) bool {
				nestedRange = true
				return false
			})
			return false
		})
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Range callback deadlocked while mutating the registry")
	}

	assert.False(t, registry.Has(first))
	assert.True(t, registry.Has(added))
	assert.True(t, gotAdded)
	assert.True(t, nestedRange)
}

func TestRegistry_RangeUsesEntrySnapshot(t *testing.T) {
	t.Parallel()

	first := newRegistrySchema()
	second := newRegistrySchema()
	added := newRegistrySchema()
	registry := NewRegistry[string]().
		Add(first, "first").
		Add(second, "second")
	seen := make(map[ZodSchema]string)
	mutated := false

	registry.Range(func(schema ZodSchema, meta string) bool {
		seen[schema] = meta
		if !mutated {
			mutated = true
			registry.Add(first, "changed-first").
				Add(second, "changed-second").
				Add(added, "added")
		}
		return true
	})

	assert.Equal(t, map[ZodSchema]string{
		first:  "first",
		second: "second",
	}, seen)
}

func TestRegistry_MetadataRemainsCallerOwned(t *testing.T) {
	schema := newRegistrySchema()
	registry := NewRegistry[GlobalMeta]().Add(schema, GlobalMeta{Title: "Caller metadata"})

	meta, ok := registry.Get(schema)
	require.True(t, ok)
	assert.Equal(t, "Caller metadata", meta.Title)
	assert.Equal(t, GlobalMeta{}, schema.Internals().Metadata())
}
