// Package mapx provides type-safe utility functions for working with Go maps.
//
// The core generic functions [ValueOf] and [ValueOr] eliminate
// boilerplate for typed map access. Convenience wrappers ([String],
// [Bool], etc.) are provided for common types.
//
// Usage:
//
//	name, ok := mapx.ValueOf[string](m, "name")
//	port := mapx.ValueOr(m, "port", 8080)
//	merged := mapx.Merge(a, b)
package mapx
