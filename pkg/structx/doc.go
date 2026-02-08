// Package structx provides utility functions for working with Go structs.
//
// Key features:
//   - Struct to map conversion (ToMap, Marshal)
//   - Map to struct conversion (FromMap, Unmarshal)
//   - JSON tag support for field naming
//
// Usage:
//
//	m := structx.ToMap(myStruct)
//	result, err := structx.FromMap(m, reflect.TypeOf(MyStruct{}))
package structx
