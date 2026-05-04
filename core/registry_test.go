package core

import (
	"testing"

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

func TestCopyGlobalMeta_CopiesRegisteredMetadata(t *testing.T) {
	// GlobalRegistry is process-wide, so this test runs serially.
	from := newRegistrySchema()
	to := newRegistrySchema()
	GlobalRegistry.Remove(from).Remove(to)
	t.Cleanup(func() {
		GlobalRegistry.Remove(from).Remove(to)
	})

	meta := GlobalMeta{ID: "schema-id", Title: "User", Examples: []any{"alice"}}
	GlobalRegistry.Add(from, meta)

	CopyGlobalMeta(from, to)

	got, ok := GlobalRegistry.Get(to)
	require.True(t, ok)
	assert.Equal(t, meta.ID, got.ID)
	assert.Equal(t, meta.Title, got.Title)
	assert.Len(t, got.Examples, 1)

	missing := newRegistrySchema()
	GlobalRegistry.Remove(missing)
	CopyGlobalMeta(nil, missing)
	assert.False(t, GlobalRegistry.Has(missing))
	CopyGlobalMeta(from, nil)
}
