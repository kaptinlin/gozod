// Package namedscalarfixture verifies generated schemas for declared scalar types.
package namedscalarfixture

// UserID is a declared string type.
type UserID string

// Counter is a declared integer type.
type Counter int64

// Enabled is a declared boolean type.
type Enabled bool

// Label is a true alias for string.
type Label = string

// Scalars exercises declared scalar types in composite fields.
type Scalars struct {
	ID           UserID             `json:"id" gozod:"required,min=2"`
	Count        Counter            `json:"count" gozod:"required,positive"`
	Active       Enabled            `json:"active" gozod:"required"`
	Alias        Label              `json:"alias" gozod:"required"`
	Pointer      *UserID            `json:"pointer" gozod:"required"`
	Slice        []UserID           `json:"slice" gozod:"required"`
	Map          map[string]Counter `json:"map" gozod:"required"`
	Array        [2]UserID          `json:"array" gozod:"required"`
	DefaultSlice []UserID           `json:"default_slice" gozod:"default=[\"primary\",\"backup\"]"`
	PrefaultMap  map[string]UserID  `json:"prefault_map" gozod:"prefault={\"primary\":\"user-1\"}"`
}
