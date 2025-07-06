// Package checks provides validation checks for schema validation
package checks

import (
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/utils"
)

// =============================================================================
// PARAMETER NORMALIZATION
// =============================================================================

// NormalizeCheckParams standardizes check parameters from various input formats
// Supports: string (shorthand) | core.SchemaParams (detailed)
func NormalizeCheckParams(params ...any) *core.CheckParams {
	if len(params) == 0 {
		return nil
	}

	param := params[0]

	// Support string parameter shorthand syntax
	if str, ok := param.(string); ok {
		return &core.CheckParams{Error: str}
	}

	// Support structured SchemaParams
	if p, ok := param.(core.SchemaParams); ok {
		if errStr, ok := p.Error.(string); ok {
			return &core.CheckParams{Error: errStr}
		}
	}

	return nil
}

// ApplyCheckParams applies normalized parameters to check definition
func ApplyCheckParams(def *core.ZodCheckDef, params *core.CheckParams) {
	if params != nil && params.Error != "" {
		errorMap := core.ZodErrorMap(func(issue core.ZodRawIssue) string {
			return params.Error
		})
		def.Error = &errorMap
	}
}

// ApplySchemaParamsToCheck applies SchemaParams to a check definition
// Used for validation checks that support error and abort configuration
func ApplySchemaParamsToCheck(def *core.ZodCheckDef, params *core.SchemaParams) {
	if params == nil {
		return
	}

	// Apply error configuration
	if params.Error != nil {
		if err, ok := utils.ToErrorMap(params.Error); ok {
			def.Error = err
		}
	}

	// Apply abort configuration
	if params.Abort {
		def.Abort = true
	}
}

// =============================================================================
// JSON SCHEMA BAG OPERATIONS
// =============================================================================

// SetBagProperty sets a property in the schema's bag for JSON Schema generation
// Used to store metadata that will be included in generated JSON Schema
func SetBagProperty(schema any, key string, value any) {
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
