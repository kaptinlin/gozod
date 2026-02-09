// Package structx converts between Go structs and map[string]any.
//
// Field names are derived from json struct tags, falling back to the
// Go field name when no tag is present. Fields tagged with json:"-"
// are skipped.
//
// Usage:
//
//	m, err := structx.ToMap(myStruct)
//	result, err := structx.FromMap(m, reflect.TypeOf(MyStruct{}))
package structx
