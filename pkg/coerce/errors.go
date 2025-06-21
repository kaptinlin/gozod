package coerce

import (
	"errors"
	"fmt"
)

// =============================================================================
// COERCION ERRORS
// =============================================================================

// Core coercion errors - precise, concise, and readable
var (
	// Type compatibility errors
	ErrUnsupported = errors.New("conversion not supported")
	ErrNilPointer  = errors.New("nil pointer")

	// Format and parsing errors
	ErrInvalidFormat = errors.New("invalid format")
	ErrEmptyInput    = errors.New("empty input")

	// Numeric conversion errors
	ErrOverflow     = errors.New("value overflow")
	ErrNegativeUint = errors.New("negative to unsigned")
	ErrNotWhole     = errors.New("not whole number")
)

// =============================================================================
// ERROR CONSTRUCTORS - Contextual error creation
// =============================================================================

// NewUnsupportedError creates a detailed unsupported conversion error
func NewUnsupportedError(from, to string) error {
	return fmt.Errorf("cannot convert %s to %s: %w", from, to, ErrUnsupported)
}

// NewFormatError creates a detailed format error with the problematic value
func NewFormatError(value, targetType string) error {
	return fmt.Errorf("cannot parse %q as %s: %w", value, targetType, ErrInvalidFormat)
}

// NewOverflowError creates a detailed overflow error with the problematic value
func NewOverflowError(value any, targetType string) error {
	return fmt.Errorf("value %v overflows %s: %w", value, targetType, ErrOverflow)
}

// NewEmptyInputError creates a detailed empty input error for specific type
func NewEmptyInputError(targetType string) error {
	return fmt.Errorf("empty input cannot convert to %s: %w", targetType, ErrEmptyInput)
}

// NewNegativeUintError creates a detailed negative to unsigned conversion error
func NewNegativeUintError(value any, targetType string) error {
	return fmt.Errorf("negative value %v cannot convert to %s: %w", value, targetType, ErrNegativeUint)
}

// NewNotWholeError creates a detailed non-whole number error
func NewNotWholeError(value any) error {
	return fmt.Errorf("value %v is not a whole number: %w", value, ErrNotWhole)
}
