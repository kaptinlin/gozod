// Package mapx provides type-safe utility functions for working with Go maps.
//
// The core generic functions [ValueOf] and [ValueOrDefault] eliminate
// boilerplate for typed map access. Convenience wrappers (GetString,
// GetBool, etc.) are provided for common types.
//
// Usage:
//
//	name, ok := mapx.ValueOf[string](m, "name")
//	port := mapx.ValueOrDefault(m, "port", 8080)
//	merged := mapx.Merge(map1, map2)
package mapx
