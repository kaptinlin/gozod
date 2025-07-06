package engine

import (
	"reflect"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/structx"
)

// =============================================================================
// SCHEMA PARAMETER HELPERS
// =============================================================================

// WithError creates a SchemaParams with custom error message
// Convenient helper for creating schema with custom error handling
func WithError(message string) core.SchemaParams {
	return core.SchemaParams{Error: message}
}

// WithDescription creates a SchemaParams with description
// Convenient helper for adding documentation to schemas
func WithDescription(description string) core.SchemaParams {
	return core.SchemaParams{Description: description}
}

// WithAbort creates a SchemaParams with abort on error enabled
// Convenient helper for configuring early termination on validation failure
func WithAbort() core.SchemaParams {
	return core.SchemaParams{Abort: true}
}

// ProcessSchemaParams processes schema parameters and returns a configuration map
// Uses structx to handle parameter conversion and mapx for merging
func ProcessSchemaParams(params ...core.SchemaParams) map[string]any {
	config := make(map[string]any)

	for _, param := range params {
		// Use structx to handle parameter conversion
		if paramMap, err := structx.ToMap(param); err == nil {
			config = mapx.Merge(config, paramMap)
		}
	}

	return config
}

// ApplySchemaParamsMultiple applies multiple schema parameters to a type definition
// This is an enhanced version that can handle multiple parameters at once
func ApplySchemaParamsMultiple(def *core.ZodTypeDef, params ...core.SchemaParams) {
	for _, param := range params {
		// Use the central params package for parameter application
		// Note: We need to import internal/params but avoid circular dependencies
		// For now, this function is a placeholder that could be enhanced later
		_ = def
		_ = param
	}
}

// IsValidSchemaType checks if a value implements the core ZodType interface
// Useful for runtime type checking and validation
func IsValidSchemaType(value any) bool {
	if value == nil {
		return false
	}

	// Use reflection to check if the type implements ZodType interface
	schemaType := reflect.TypeOf((*core.ZodType[any])(nil)).Elem()
	return reflect.TypeOf(value).Implements(schemaType)
}
