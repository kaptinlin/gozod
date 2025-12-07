// Package reflectx provides reflection utilities for type checking and value manipulation.
//
// Key features:
//   - Type checking functions (IsNil, IsZero, IsBool, IsString, IsNumeric, etc.)
//   - Pointer dereferencing utilities (Deref, DerefAll)
//   - Value extraction and conversion helpers
//   - Runtime type detection aligned with Zod's type system
//
// Usage:
//
//	if reflectx.IsNumeric(value) {
//	    // handle numeric value
//	}
//
//	derefed, ok := reflectx.Deref(maybePtr)
//	parsedType := reflectx.ParsedType(value)
package reflectx
