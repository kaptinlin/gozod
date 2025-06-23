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

// ApplySchemaParams applies schema parameters to a type definition
// This function processes SchemaParams and updates the type definition accordingly
func ApplySchemaParams(def *core.ZodTypeDef, params ...core.SchemaParams) {
	// Process parameters using the new ProcessSchemaParams function
	config := ProcessSchemaParams(params...)

	// Apply custom error if provided
	if errorMsg := mapx.GetAnyDefault(config, "Error", nil); errorMsg != nil {
		// param.Error should be a function or string that can be converted to ZodErrorMap
		// For now, we'll handle string messages
		if msgStr, ok := errorMsg.(string); ok {
			errorMap := core.ZodErrorMap(func(issue core.ZodRawIssue) string {
				return msgStr
			})
			def.Error = &errorMap
		}
	}

	// Note: Other parameters like Description, Coerce, etc. are handled
	// in type-specific implementations since they affect different parts
	// of the schema differently
}

// IsValidSchemaType checks if a value implements the core ZodType interface
// Useful for runtime type checking and validation
func IsValidSchemaType(value any) bool {
	if value == nil {
		return false
	}

	// Use reflection to check if the type implements ZodType interface
	schemaType := reflect.TypeOf((*core.ZodType[any, any])(nil)).Elem()
	return reflect.TypeOf(value).Implements(schemaType)
}
