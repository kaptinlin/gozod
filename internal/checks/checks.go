// Package checks provides validation checks for schema validation
package checks

import (
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/utils"
)

// =============================================================================
// PARAMETER TYPES AND NORMALIZATION
// =============================================================================

// CheckParams represents unified check parameters structure
// Supports both string shorthand and detailed configuration
type CheckParams struct {
	Error string // Custom error message
}

// normalizeCheckParams standardizes check parameters from various input formats
// Supports: string (shorthand) | CheckParams (detailed) | core.SchemaParams (legacy)
func normalizeCheckParams(params ...any) *CheckParams {
	if len(params) == 0 {
		return nil
	}

	param := params[0]

	// Support string parameter shorthand syntax
	if str, ok := param.(string); ok {
		return &CheckParams{Error: str}
	}

	// Support structured CheckParams
	if p, ok := param.(CheckParams); ok {
		return &p
	}

	// Support legacy core.SchemaParams for backward compatibility
	if p, ok := param.(core.SchemaParams); ok {
		if errStr, ok := p.Error.(string); ok {
			return &CheckParams{Error: errStr}
		}
	}

	return nil
}

// applyCheckParams applies normalized parameters to check definition
func applyCheckParams(def *core.ZodCheckDef, params *CheckParams) {
	if params != nil && params.Error != "" {
		errorMap := core.ZodErrorMap(func(issue core.ZodRawIssue) string {
			return params.Error
		})
		def.Error = &errorMap
	}
}

// normalizeParams legacy parameter normalization for backward compatibility
// Deprecated: Use normalizeCheckParams and applyCheckParams instead
func normalizeParams(def *core.ZodCheckDef, params []string) {
	if len(params) > 0 && params[0] != "" {
		// Convert string parameters to error map
		message := params[0]
		errorMap := core.ZodErrorMap(func(issue core.ZodRawIssue) string {
			return message
		})
		def.Error = &errorMap
	}
}

// =============================================================================
// JSON SCHEMA BAG OPERATIONS
// =============================================================================

// setBagProperty sets a property in the schema's bag for JSON Schema generation
// Used to store metadata that will be included in generated JSON Schema
func setBagProperty(schema any, key string, value any) {
	if s, ok := schema.(interface{ GetInternals() *core.ZodTypeInternals }); ok {
		internals := s.GetInternals()
		if internals.Bag == nil {
			internals.Bag = make(map[string]any)
		}
		internals.Bag[key] = value
	}
}

// mergeConstraint merges a constraint into the schema's bag with conflict resolution
// Uses merge function to handle conflicts when the same constraint exists
func mergeConstraint(schema any, key string, value any, merge func(old, new any) any) {
	if s, ok := schema.(interface{ GetInternals() *core.ZodTypeInternals }); ok {
		internals := s.GetInternals()
		if internals.Bag == nil {
			internals.Bag = make(map[string]any)
		}

		if existing, exists := internals.Bag[key]; exists {
			internals.Bag[key] = merge(existing, value)
		} else {
			internals.Bag[key] = value
		}
	}
}

// mergeMinimumConstraint merges minimum constraint, choosing the stricter value
func mergeMinimumConstraint(schema any, value any, inclusive bool) {
	key := "minimum"
	if !inclusive {
		key = "exclusiveMinimum"
	}

	mergeConstraint(schema, key, value, func(old, new any) any {
		// Choose stricter (larger) minimum value
		if utils.CompareValues(new, old) > 0 {
			return new
		}
		return old
	})
}

// mergeMaximumConstraint merges maximum constraint, choosing the stricter value
func mergeMaximumConstraint(schema any, value any, inclusive bool) {
	key := "maximum"
	if !inclusive {
		key = "exclusiveMaximum"
	}

	mergeConstraint(schema, key, value, func(old, new any) any {
		// Choose stricter (smaller) maximum value
		if utils.CompareValues(new, old) < 0 {
			return new
		}
		return old
	})
}
