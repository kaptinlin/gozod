package engine

import (
	"errors"

	"github.com/kaptinlin/gozod/pkg/coerce"
)

// =============================================================================
// PARSER AND SCHEMA ERRORS
// =============================================================================

var (
	// Schema validation errors
	ErrSchemaNoInternals    = errors.New("schema has no internals")
	ErrSchemaNoParseFunc    = errors.New("schema has no parse function")
	ErrParseFuncReturnedNil = errors.New("parse function returned nil")

	// Type assertion and conversion errors - integrated with coerce package
	ErrTypeAssertionFailed = coerce.ErrUnsupported
	ErrCannotConvertType   = coerce.ErrInvalidFormat
)
