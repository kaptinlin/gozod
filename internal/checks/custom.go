package checks

import (
	"reflect"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/internal/utils"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

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
	// Use unified parameter handling
	param := utils.GetFirstParam(args...)
	normalizedParams := utils.NormalizeParams(param)

	def := &ZodCheckCustomDef{
		ZodCheckDef: core.ZodCheckDef{Check: "custom"},
		Fn:          fn,
		Params:      make(map[string]any),
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

	// Apply schema parameters using the centralized utility
	ApplySchemaParamsToCheck(&def.ZodCheckDef, normalizedParams)

	// Handle additional parameters from SchemaParams.Params
	if normalizedParams != nil && len(normalizedParams.Params) > 0 {
		for k, v := range normalizedParams.Params {
			def.Params[k] = v
		}
	}

	// Create internals with enhanced metadata storage
	internals := &ZodCheckCustomInternals{
		ZodCheckInternals: core.ZodCheckInternals{
			Def: &def.ZodCheckDef,
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

	// Use CreateCustomIssue directly
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
			// Fallback: attempt to invoke arbitrary func via reflection if it
			// matches signature func(T) bool where T is assignable from the
			// actual payload.GetValue() type. This provides support for additional
			// composite types without enumerating each one.
			fv := reflect.ValueOf(internals.Def.Fn)
			if fv.Kind() == reflect.Func && fv.Type().NumIn() == 1 && fv.Type().NumOut() == 1 && fv.Type().Out(0).Kind() == reflect.Bool {
				// Ensure the input parameter is compatible.
				argType := fv.Type().In(0)
				value := payload.GetValue()
				val := reflect.ValueOf(value)
				if val.IsValid() && val.Type().AssignableTo(argType) {
					resultVals := fv.Call([]reflect.Value{val})
					result := resultVals[0].Bool()
					handleRefineResult(result, payload, value, internals)
					break
				}
			}
			// Unknown refine function type or incompatible value: treat as validation failure
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
