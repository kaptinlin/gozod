package checks

import (
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/internal/utils"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/slicex"
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

// ensureBag ensures the schema's Bag is initialized and returns it
func ensureBag(schema any) map[string]any {
	if s, ok := schema.(interface{ GetInternals() *core.ZodTypeInternals }); ok {
		internals := s.GetInternals()
		if internals.Bag == nil {
			internals.Bag = make(map[string]any)
		}
		return internals.Bag
	}
	return nil
}

// SetBagProperty sets a property in the schema's bag for JSON Schema generation
// Used to store metadata that will be included in generated JSON Schema
func SetBagProperty(schema any, key string, value any) {
	if bag := ensureBag(schema); bag != nil {
		bag[key] = value
	}
}

// mergeConstraint merges a constraint into the schema's bag with conflict resolution
// Uses merge function to handle conflicts when the same constraint exists
func mergeConstraint(schema any, key string, value any, merge func(old, new any) any) {
	if bag := ensureBag(schema); bag != nil {
		if existing, exists := bag[key]; exists {
			bag[key] = merge(existing, value)
		} else {
			bag[key] = value
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

	// Ensure no conflicting opposite constraint remains
	if s, ok := schema.(interface{ GetInternals() *core.ZodTypeInternals }); ok {
		bag := s.GetInternals().Bag
		if bag == nil {
			return
		}
		if inclusive {
			// We now have an inclusive minimum; remove any existing exclusiveMin that is lower or equal
			if ex, exists := bag["exclusiveMinimum"]; exists {
				if utils.CompareValues(value, ex) >= 0 {
					delete(bag, "exclusiveMinimum")
				}
			}
		} else {
			// We now have an exclusive minimum; remove inclusive min if equal or smaller
			if mi, exists := bag["minimum"]; exists {
				if utils.CompareValues(value, mi) >= 0 {
					delete(bag, "minimum")
				}
			}
		}
	}
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

	// Ensure no conflicting opposite constraint remains
	if s, ok := schema.(interface{ GetInternals() *core.ZodTypeInternals }); ok {
		bag := s.GetInternals().Bag
		if bag == nil {
			return
		}
		if inclusive {
			// inclusive maximum => remove exclusiveMaximum if higher or equal
			if ex, exists := bag["exclusiveMaximum"]; exists {
				if utils.CompareValues(value, ex) <= 0 {
					delete(bag, "exclusiveMaximum")
				}
			}
		} else {
			// exclusive maximum => remove maximum if higher or equal
			if ma, exists := bag["maximum"]; exists {
				if utils.CompareValues(value, ma) <= 0 {
					delete(bag, "maximum")
				}
			}
		}
	}
}

// =============================================================================
// CUSTOM VALIDATION CHECKS
// =============================================================================

// ZodCheckCustomDef defines custom validation constraint for user-defined validation logic
type ZodCheckCustomDef struct {
	core.ZodCheckDef
	Type   string         // Custom check type identifier
	Params map[string]any // Additional parameters for custom logic
	Fn     any            // RefineFn or CheckFn function
	FnType string         // "refine" or "check" function type
}

// ZodCheckCustomInternals contains custom check internal state and validation state
type ZodCheckCustomInternals struct {
	core.ZodCheckInternals
	Def  *ZodCheckCustomDef // Custom check definition
	Issc *core.ZodIssueBase // Issue base for error handling
	Bag  map[string]any     // Additional metadata storage
}

// ZodCheckCustom represents custom validation check for user-defined validation logic
// This check executes user-provided refine or check functions with proper error handling
type ZodCheckCustom struct {
	Internals *ZodCheckCustomInternals
}

// GetZod returns the internal check structure for execution
func (z *ZodCheckCustom) GetZod() *core.ZodCheckInternals {
	return &z.Internals.ZodCheckInternals
}

// =============================================================================
// CONSTRUCTOR FUNCTIONS
// =============================================================================

// NewCustom creates a new custom validation check with user-defined validation logic
// Uses unified parameter handling following Zod TypeScript v4 pattern
func NewCustom[T any](fn any, args ...any) *ZodCheckCustom {
	// Use unified parameter handling with CustomParams
	param := utils.GetFirstParam(args...)
	customParams := utils.NormalizeCustomParams(param)

	def := &ZodCheckCustomDef{
		ZodCheckDef: core.ZodCheckDef{
			Check: "custom",
			Error: nil,
			Abort: customParams.Abort, // Use CustomParams.Abort directly
		},
		Fn:     fn,
		Params: make(map[string]any),
	}

	// Determine function type
	switch fn.(type) {
	case core.ZodRefineFn[T], func(T) bool:
		def.FnType = "refine"
	case core.ZodCheckFn, func(*core.ParsePayload):
		def.FnType = "check"
	default:
		def.FnType = "refine" // Default to refine for backward compatibility
	}

	// Apply error configuration from CustomParams
	if customParams.Error != nil {
		if errorMap, ok := utils.ToErrorMap(customParams.Error); ok {
			def.Error = errorMap
		}
	}

	// Handle additional parameters from CustomParams.Params
	if len(customParams.Params) > 0 {
		for k, v := range customParams.Params {
			def.Params[k] = v
		}
	}

	// Store custom path if provided
	if len(customParams.Path) > 0 {
		def.Params["path"] = customParams.Path
	}

	// Create internals with enhanced metadata storage
	internals := &ZodCheckCustomInternals{
		ZodCheckInternals: core.ZodCheckInternals{
			Def:  &def.ZodCheckDef,
			When: customParams.When, // Attach when predicate from CustomParams
		},
		Def:  def,
		Issc: &core.ZodIssueBase{},
		Bag:  make(map[string]any),
	}

	// Set up the validation function
	internals.Check = func(payload *core.ParsePayload) {
		executeCustomCheck(payload, internals)
	}

	return &ZodCheckCustom{Internals: internals}
}

// handleRefineResult processes refine function validation results with strong typing
// Creates appropriate error issues when validation fails
func handleRefineResult(result bool, payload *core.ParsePayload, input any, internals *ZodCheckCustomInternals) {
	// Only process if validation failed
	if result {
		// Validation passed, no error
		return
	}

	// Construct error path from payload path
	payloadPath := payload.GetPath()
	path := make([]any, len(payloadPath))
	copy(path, payloadPath)

	// Override the issue path if a custom `path` parameter is provided
	if customPath, ok := internals.Def.Params["path"]; ok {
		switch p := customPath.(type) {
		case []any:
			path = p
		case []string:
			conv := make([]any, len(p))
			for i, v := range p {
				conv[i] = v
			}
			path = conv
		case string:
			path = []any{p}
		}
	}

	// Determine the error message to use - use mapx for safe access
	var errorMessage string
	if internals.Def.Error != nil {
		// Handle both string and function error messages
		errorMap := *internals.Def.Error

		// Create a temporary issue to get the custom message
		tempIssue := core.ZodRawIssue{
			Code:  core.Custom,
			Input: input,
			Path:  path,
		}
		errorMessage = errorMap(tempIssue)
	}

	// If no custom message, use default
	if errorMessage == "" {
		errorMessage = "Invalid input"
	}

	// Create properties map using mapx for safe operations
	properties := make(map[string]any)
	mapx.Set(properties, "origin", "custom")
	mapx.Set(properties, "continue", !internals.Def.Abort)

	// Add custom parameters if provided using mapx
	if mapx.Count(internals.Def.Params) > 0 {
		mapx.Set(properties, "params", internals.Def.Params)
	}

	// Use low-level CreateCustomIssue directly since ParseContext is not available at this level
	// The high-level error creation pattern will be applied at the engine level
	issue := issues.CreateCustomIssue(errorMessage, properties, input)

	// Set the Input and Inst for downstream processing
	issue.Input = input
	issue.Inst = internals

	// Attach issue with explicit path
	payload.AddIssueWithPath(issue, path)
}

// executeCustomCheck executes custom validation check with strong typing support
// Handles both refine functions (boolean return) and check functions (payload modification)
func executeCustomCheck(payload *core.ParsePayload, internals *ZodCheckCustomInternals) {
	defer func() {
		if r := recover(); r != nil {
			value := payload.GetValue()
			handleRefineResult(false, payload, value, internals)
		}
	}()

	switch internals.Def.FnType {
	case "refine":
		// Execute refine function with type-safe casting
		// Support common concrete signatures first for performance
		switch fn := internals.Def.Fn.(type) {
		case func([]any) bool:
			value := payload.GetValue()
			if arr, ok := value.([]any); ok {
				result := fn(arr)
				handleRefineResult(result, payload, value, internals)
			} else {
				handleRefineResult(false, payload, value, internals)
			}
		case func(string) bool:
			value := payload.GetValue()
			if str, ok := value.(string); ok {
				result := fn(str)
				handleRefineResult(result, payload, value, internals)
			} else {
				// Type mismatch: create validation error instead of panic
				handleRefineResult(false, payload, value, internals)
			}
		case func(map[string]any) bool:
			// Use mapx to safely handle map type checking
			value := payload.GetValue()
			if mapData, ok := value.(map[string]any); ok {
				result := fn(mapData)
				handleRefineResult(result, payload, value, internals)
			} else {
				// Type mismatch: create validation error instead of panic
				handleRefineResult(false, payload, value, internals)
			}
		case func(any) bool:
			// Handle any type - accepts any value
			value := payload.GetValue()
			result := fn(value)
			handleRefineResult(result, payload, value, internals)
		case core.ZodRefineFn[string]:
			value := payload.GetValue()
			if str, ok := value.(string); ok {
				result := fn(str)
				handleRefineResult(result, payload, value, internals)
			} else {
				// Type mismatch: create validation error instead of panic
				handleRefineResult(false, payload, value, internals)
			}
		case core.ZodRefineFn[map[string]any]:
			// Use mapx to safely handle map type checking
			value := payload.GetValue()
			if mapData, ok := value.(map[string]any); ok {
				result := fn(mapData)
				handleRefineResult(result, payload, value, internals)
			} else {
				// Type mismatch: create validation error instead of panic
				handleRefineResult(false, payload, value, internals)
			}
		case core.ZodRefineFn[any]:
			// Handle any type - accepts any value
			value := payload.GetValue()
			result := fn(value)
			handleRefineResult(result, payload, value, internals)
		default:
			// Unknown refine function type: treat as validation failure
			value := payload.GetValue()
			handleRefineResult(false, payload, value, internals)
		}

	case "check":
		// Execute check function with direct payload access
		// The CheckFn has access to payload.AddIssue() method for ctx.issues.push() functionality
		if checkFn, ok := internals.Def.Fn.(core.ZodCheckFn); ok {
			checkFn(payload)
		} else if checkFn, ok := internals.Def.Fn.(func(*core.ParsePayload)); ok {
			checkFn(payload)
		} else {
			// Invalid check function type: create validation error
			value := payload.GetValue()
			handleRefineResult(false, payload, value, internals)
		}

	default:
		// Unknown custom function type: create validation error
		value := payload.GetValue()
		handleRefineResult(false, payload, value, internals)
	}
}

// =============================================================================
// TRANSFORM VALIDATION CHECKS
// =============================================================================

// ZodCheckOverwriteDef defines overwrite validation constraint for value transformation
type ZodCheckOverwriteDef struct {
	core.ZodCheckDef
	Transform func(any) any // Transformation function
}

// ZodCheckOverwriteInternals contains overwrite check internal state
type ZodCheckOverwriteInternals struct {
	core.ZodCheckInternals
	Def *ZodCheckOverwriteDef
}

// ZodCheckOverwrite represents overwrite validation check for value transformation
// This check modifies the payload value using the provided transformation function
type ZodCheckOverwrite struct {
	Internals *ZodCheckOverwriteInternals
}

// GetZod returns the internal check structure for execution
func (c *ZodCheckOverwrite) GetZod() *core.ZodCheckInternals {
	return &c.Internals.ZodCheckInternals
}

// NewZodCheckOverwrite creates a new check that overwrites input with transformed value
func NewZodCheckOverwrite(transform func(any) any, args ...any) *ZodCheckOverwrite {
	// Use unified parameter handling
	param := utils.GetFirstParam(args...)
	normalizedParams := utils.NormalizeParams(param)

	def := &ZodCheckOverwriteDef{
		ZodCheckDef: core.ZodCheckDef{Check: "overwrite"},
		Transform:   transform,
	}

	// Apply schema parameters using the centralized utility
	ApplySchemaParamsToCheck(&def.ZodCheckDef, normalizedParams)

	// Create internals with validation function
	internals := &ZodCheckOverwriteInternals{
		ZodCheckInternals: core.ZodCheckInternals{
			Def: &def.ZodCheckDef,
		},
		Def: def,
	}

	// Set up the transformation function
	internals.Check = func(payload *core.ParsePayload) {
		// Transform the input value
		transformedValue := transform(payload.GetValue())
		// Overwrite the payload value
		payload.SetValue(transformedValue)
	}

	return &ZodCheckOverwrite{Internals: internals}
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// PrefixIssues adds path prefix to validation issues for nested object validation
// This utility function is used by property validation checks
func PrefixIssues(path any, issues []core.ZodRawIssue) []core.ZodRawIssue {
	// Use slicex to safely handle issues slice
	if slicex.IsEmpty(issues) {
		return issues
	}

	for i := range issues {
		// Ensure path exists using slicex
		if issues[i].Path == nil {
			issues[i].Path = make([]any, 0)
		}
		// Prepend the path segment using slicex operations
		newPath := make([]any, len(issues[i].Path)+1)
		newPath[0] = path
		copy(newPath[1:], issues[i].Path)
		issues[i].Path = newPath
	}
	return issues
}

// =============================================================================
// PROPERTY VALIDATION CHECKS
// =============================================================================

// ZodCheckPropertyDef defines property validation constraint for specific object fields
type ZodCheckPropertyDef struct {
	core.ZodCheckDef
	Property string         // Property name to validate
	Schema   core.ZodSchema // Schema to validate property against
}

// ZodCheckPropertyInternals contains property check internal state
type ZodCheckPropertyInternals struct {
	core.ZodCheckInternals
	Def  *ZodCheckPropertyDef // Property check definition
	Issc *core.ZodIssueBase   // Issue base for error handling
}

// ZodCheckProperty represents property validation check for specific object fields
// This check validates that a specific property of an object matches the given schema
type ZodCheckProperty struct {
	Internals *ZodCheckPropertyInternals
}

// GetZod returns the internal check structure for execution
func (z *ZodCheckProperty) GetZod() *core.ZodCheckInternals {
	return &z.Internals.ZodCheckInternals
}

// NewProperty creates a new property validation check
// Validates that input[property] matches the provided schema
func NewProperty(property string, schema core.ZodSchema, args ...any) *ZodCheckProperty {
	// Use unified parameter handling
	param := utils.GetFirstParam(args...)
	normalizedParams := utils.NormalizeParams(param)

	def := &ZodCheckPropertyDef{
		ZodCheckDef: core.ZodCheckDef{Check: "property"},
		Property:    property,
		Schema:      schema,
	}

	// Apply schema parameters using the centralized utility
	ApplySchemaParamsToCheck(&def.ZodCheckDef, normalizedParams)

	// Create internals
	internals := &ZodCheckPropertyInternals{
		ZodCheckInternals: core.ZodCheckInternals{
			Def: &def.ZodCheckDef,
		},
		Def:  def,
		Issc: &core.ZodIssueBase{},
	}

	// Set up the validation function
	internals.Check = func(payload *core.ParsePayload) {
		executePropertyCheck(payload, internals)
	}

	return &ZodCheckProperty{Internals: internals}
}

// executePropertyCheck executes property validation check
func executePropertyCheck(payload *core.ParsePayload, internals *ZodCheckPropertyInternals) {
	input := payload.GetValue()

	// Ensure input is an object
	objMap, ok := input.(map[string]any)
	if !ok {
		// If input is not an object, skip property validation
		return
	}

	// Get the property value
	propertyValue, exists := objMap[internals.Def.Property]
	if !exists {
		// Property doesn't exist, skip validation (let object schema handle required fields)
		return
	}

	// Validate property value using ParseAny - no reflection needed!
	_, parseErr := internals.Def.Schema.ParseAny(propertyValue)

	if parseErr != nil {
		// Validation failed, create an issue with the property path
		path := append(payload.GetPath(), internals.Def.Property)

		// Determine the error message
		var errorMessage string
		if internals.Def.Error != nil {
			errorMap := *internals.Def.Error
			tempIssue := core.ZodRawIssue{
				Code:  core.Custom,
				Input: propertyValue,
				Path:  path,
			}
			errorMessage = errorMap(tempIssue)
		} else {
			errorMessage = parseErr.Error()
		}

		// Create properties map
		properties := make(map[string]any)
		mapx.Set(properties, "origin", "property")
		mapx.Set(properties, "property", internals.Def.Property)
		mapx.Set(properties, "continue", !internals.Def.Abort)

		// Create custom issue
		issue := issues.CreateCustomIssue(errorMessage, properties, propertyValue)
		issue.Input = propertyValue
		issue.Inst = internals

		// Add issue with property path
		payload.AddIssueWithPath(issue, path)
	}
}
