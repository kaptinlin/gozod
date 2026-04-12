package validate

import "github.com/kaptinlin/gozod/pkg/reflectx"

// MaxLength reports whether value's length is at most maximum.
func MaxLength(value any, maximum int) bool {
	l, ok := reflectx.Length(value)
	if !ok {
		return false
	}
	return l <= maximum
}

// MinLength reports whether value's length is at least minimum.
func MinLength(value any, minimum int) bool {
	l, ok := reflectx.Length(value)
	if !ok {
		return false
	}
	return l >= minimum
}

// Length reports whether value's length equals exactly the expected length.
func Length(value any, expected int) bool {
	l, ok := reflectx.Length(value)
	if !ok {
		return false
	}
	return l == expected
}

// MaxSize reports whether the collection's size is at most maximum.
func MaxSize(value any, maximum int) bool {
	size, ok := collectionSize(value)
	if !ok {
		return false
	}
	return size <= maximum
}

// MinSize reports whether the collection's size is at least minimum.
func MinSize(value any, minimum int) bool {
	size, ok := collectionSize(value)
	if !ok {
		return false
	}
	return size >= minimum
}

// Size reports whether the collection's size equals exactly the expected size.
func Size(value any, expected int) bool {
	size, ok := collectionSize(value)
	if !ok {
		return false
	}
	return size == expected
}
