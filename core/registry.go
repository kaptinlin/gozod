package core

import (
	"sync"
)

// Registry provides a lightweight implementation to associate arbitrary,
// strongly-typed metadata with any `ZodSchema` instance. The API is
// intentionally minimal—only the most common CRUD operations are included.
//
// Requirements & Design Notes:
//   - M must be a comparable type (basic types, structs, interfaces, etc.).
//   - The implementation relies on a `map[ZodSchema]M`. This works because
//     every built-in GoZod schema is a pointer and therefore comparable.
//   - The Registry is concurrency-safe; read and write operations are guarded
//     by an RWMutex. If you need advanced features (filtering, cloning, etc.),
//     you can compose them on top of this implementation.
//   - The design mirrors the TypeScript Zod Registry feature but remains a
//     strict subset—no over-engineering or extra capabilities that don't
//     exist in the original TS API (e.g., no named registries).
type Registry[M any] struct {
	mu   sync.RWMutex
	meta map[ZodSchema]M
}

// NewRegistry creates an empty Registry.
func NewRegistry[M any]() *Registry[M] {
	return &Registry[M]{
		meta: make(map[ZodSchema]M),
	}
}

// Add associates a schema with metadata; it overwrites the entry if it already exists.
// Returns the Registry to allow for method chaining.
func (r *Registry[M]) Add(schema ZodSchema, m M) *Registry[M] {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.meta[schema] = m
	return r
}

// Get retrieves the metadata for a schema and reports whether it was found.
func (r *Registry[M]) Get(schema ZodSchema) (M, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	m, ok := r.meta[schema]
	return m, ok
}

// Remove deletes an entry from the registry.
// Returns the Registry to allow for method chaining.
func (r *Registry[M]) Remove(schema ZodSchema) *Registry[M] {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.meta, schema)
	return r
}

// Has checks if a schema is registered in the Registry.
func (r *Registry[M]) Has(schema ZodSchema) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.meta[schema]
	return ok
}

// Range iterates over the schemas and metadata in the registry.
// If the callback function returns false, iteration stops.
func (r *Registry[M]) Range(f func(schema ZodSchema, m M) bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for k, v := range r.meta {
		if !f(k, v) {
			break
		}
	}
}

//------------------------------------------------------------------------------
// Global Registry
//------------------------------------------------------------------------------

// GlobalMeta defines a standard, optional metadata structure that aligns with
// common JSON Schema fields. Users can extend this for their own purposes.
type GlobalMeta struct {
	ID          string `json:"id,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Examples    []any  `json:"examples,omitempty"`
}

// GlobalRegistry is the default, framework-provided global registry for
// convenient, shared metadata collection.
var GlobalRegistry = NewRegistry[GlobalMeta]()
