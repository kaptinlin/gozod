// Package structx converts between Go structs and map[string]any.
//
// Every function takes the struct tag used for field names as its first
// argument (e.g. "json", "yaml", "toml"). Field names fall back to the Go
// field name when no such tag is present, and fields tagged with `tag:"-"`
// are skipped.
//
// Usage:
//
//	m, err := structx.ToMap("json", myStruct)
//	result, err := structx.FromMap("yaml", m, reflect.TypeOf(MyStruct{}))
package structx
