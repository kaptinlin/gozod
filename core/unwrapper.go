package core

// Unwrapper allows wrapper types to expose their underlying value for validation.
// This enables validation of types like Optional[T], sql.NullString, etc.
//
// Example:
//
//	type Optional[T any] struct {
//	    Value T
//	    set   bool
//	}
//
//	func (o Optional[T]) Unwrap() (any, bool) {
//	    return o.Value, o.set
//	}
type Unwrapper interface {
	// Unwrap returns the underlying value and whether it is present.
	// If ok is false, the value should not be validated.
	Unwrap() (value any, ok bool)
}
