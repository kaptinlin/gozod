// Package slicex provides utility functions for working with Go slices.
//
// Key features:
//   - Slice conversion (ToAny, ToTyped, ToStrings)
//   - Element extraction (Extract, ExtractArray, ExtractSlice)
//   - Slice merging (Merge, Append, Prepend)
//   - Common utilities (Contains, Unique, Filter, Map, Join)
//
// Usage:
//
//	anySlice := slicex.ToAny(typedSlice)
//	unique := slicex.Unique(slice)
//	filtered := slicex.Filter(slice, predicate)
package slicex
