// Package reflectx provides reflection utilities for type checking and
// value manipulation.
//
// Key features:
//   - Type checking (IsNil, IsBool, IsString, IsNumeric, etc.)
//   - Pointer dereferencing (Deref, DerefAll)
//   - Value extraction (StringVal, Length, Size)
//   - Type conversion (Convert)
//   - Runtime type detection aligned with Zod's type system
//
// Usage:
//
//	if reflectx.IsNumeric(value) {
//	    // handle numeric value
//	}
//
//	val, ok := reflectx.Deref(maybePtr)
//	pt := reflectx.ParsedType(value)
package reflectx
