// Package coerce provides type coercion utilities for converting values
// between different Go types with proper error handling and overflow detection.
//
// Key features:
//   - Safe type conversions with detailed error messages
//   - Support for primitive types, big.Int, complex numbers, and time.Time
//   - Generic To[T] function for unified coercion API
//   - Overflow detection for numeric conversions
//
// Usage:
//
//	val, err := coerce.To[int64]("123")
//	if err != nil {
//	    // handle error
//	}
//
//	result, err := coerce.ToBool("true")  // true, nil
//	result, err := coerce.ToString(123)  // "123", nil
package coerce
