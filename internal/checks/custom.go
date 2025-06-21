package checks

import (
	"reflect"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
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
	Path   []any          // Error path for nested validation
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

// RefineFn defines function signature for refine checks with generic type support
type RefineFn[T any] func(T) bool

// CheckFn defines function signature for check functions with direct payload access
type CheckFn func(*core.ParsePayload)

// =============================================================================
// CONSTRUCTOR FUNCTIONS
// =============================================================================

// NewCustom creates a new custom validation check with user-defined validation logic
// Supports both legacy core.SchemaParams and new unified parameter handling
func NewCustom[T any](fn any, params ...any) *ZodCheckCustom {
	def := &ZodCheckCustomDef{
		ZodCheckDef: core.ZodCheckDef{Check: "custom"},
		Fn:          fn,
		Params:      make(map[string]any),
	}

	// Determine function type
	switch fn.(type) {
	case RefineFn[T], func(T) bool:
		def.FnType = "refine"
	case CheckFn, func(*core.ParsePayload):
		def.FnType = "check"
	default:
		def.FnType = "refine" // Default to refine for backward compatibility
	}

	// Handle parameter normalization - support both old and new styles
	var schemaParams []core.SchemaParams

	for _, param := range params {
		switch p := param.(type) {
		case string:
			// String shorthand for error message
			schemaParams = append(schemaParams, core.SchemaParams{Error: p})
		case CheckParams:
			// New unified parameter format
			schemaParams = append(schemaParams, core.SchemaParams{Error: p.Error})
		case core.SchemaParams:
			// Legacy format - use as-is
			schemaParams = append(schemaParams, p)
		}
	}

	// Apply schema parameters using existing logic
	core.ApplySchemaParams(&def.ZodCheckDef, schemaParams...)

	// Handle custom parameters that ApplySchemaParams doesn't process
	for _, param := range schemaParams {
		// Handle Path parameter
		if len(param.Path) > 0 {
			def.Path = make([]any, len(param.Path))
			for i, p := range param.Path {
				def.Path[i] = p
			}
		}

		// Handle Params parameter
		if len(param.Params) > 0 {
			for k, v := range param.Params {
				def.Params[k] = v
			}
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

	// Construct error path by combining payload path with custom path - use slicex for safe operations
	path := make([]any, len(payload.Path))
	copy(path, payload.Path)

	// Append custom path if provided using slicex
	if !slicex.IsEmpty(internals.Def.Path) {
		path = append(path, internals.Def.Path...)
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

	// Use CreateCustomIssue directly
	issue := issues.CreateCustomIssue(errorMessage, properties, input)

	// Set the path directly on the issue (not in properties)
	issue.Path = path

	// Set the Input field manually since CreateCustomIssue doesn't set it
	issue.Input = input

	// Set Inst field for schema-level error mapping access
	issue.Inst = internals

	payload.Issues = append(payload.Issues, issue)
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
			if arr, ok := payload.Value.([]any); ok {
				result := fn(arr)
				handleRefineResult(result, payload, payload.Value, internals)
			} else {
				handleRefineResult(false, payload, payload.Value, internals)
			}
		case func(string) bool:
			if str, ok := payload.Value.(string); ok {
				result := fn(str)
				handleRefineResult(result, payload, payload.Value, internals)
			} else {
				// Type mismatch: create validation error instead of panic
				handleRefineResult(false, payload, payload.Value, internals)
			}
		case func(map[string]any) bool:
			// Use mapx to safely handle map type checking
			if mapData, ok := payload.Value.(map[string]any); ok {
				result := fn(mapData)
				handleRefineResult(result, payload, payload.Value, internals)
			} else {
				// Type mismatch: create validation error instead of panic
				handleRefineResult(false, payload, payload.Value, internals)
			}
		case func(any) bool:
			// Handle any type - accepts any value
			result := fn(payload.Value)
			handleRefineResult(result, payload, payload.Value, internals)
		case RefineFn[string]:
			if str, ok := payload.Value.(string); ok {
				result := fn(str)
				handleRefineResult(result, payload, payload.Value, internals)
			} else {
				// Type mismatch: create validation error instead of panic
				handleRefineResult(false, payload, payload.Value, internals)
			}
		case RefineFn[map[string]any]:
			// Use mapx to safely handle map type checking
			if mapData, ok := payload.Value.(map[string]any); ok {
				result := fn(mapData)
				handleRefineResult(result, payload, payload.Value, internals)
			} else {
				// Type mismatch: create validation error instead of panic
				handleRefineResult(false, payload, payload.Value, internals)
			}
		case RefineFn[any]:
			// Handle any type - accepts any value
			result := fn(payload.Value)
			handleRefineResult(result, payload, payload.Value, internals)
		default:
			// Fallback: attempt to invoke arbitrary func via reflection if it
			// matches signature func(T) bool where T is assignable from the
			// actual payload.Value type. This provides support for additional
			// composite types without enumerating each one.
			fv := reflect.ValueOf(internals.Def.Fn)
			if fv.Kind() == reflect.Func && fv.Type().NumIn() == 1 && fv.Type().NumOut() == 1 && fv.Type().Out(0).Kind() == reflect.Bool {
				// Ensure the input parameter is compatible.
				argType := fv.Type().In(0)
				val := reflect.ValueOf(payload.Value)
				if val.IsValid() && val.Type().AssignableTo(argType) {
					resultVals := fv.Call([]reflect.Value{val})
					result := resultVals[0].Bool()
					handleRefineResult(result, payload, payload.Value, internals)
					break
				}
			}
			// Unknown refine function type or incompatible value: treat as validation failure
			handleRefineResult(false, payload, payload.Value, internals)
		}

	case "check":
		// Execute check function with direct payload access
		// The CheckFn has access to payload.AddIssue() method for ctx.issues.push() functionality
		if checkFn, ok := internals.Def.Fn.(CheckFn); ok {
			checkFn(payload)
		} else if checkFn, ok := internals.Def.Fn.(func(*core.ParsePayload)); ok {
			checkFn(payload)
		} else {
			// Invalid check function type: create validation error
			handleRefineResult(false, payload, payload.Value, internals)
		}

	default:
		// Unknown custom function type: create validation error
		handleRefineResult(false, payload, payload.Value, internals)
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

// NewZodCheckOverwrite creates a new overwrite transformation check
// Parameters:
//   - transform: Function to transform the input value
//   - params: Optional schema parameters for error customization
func NewZodCheckOverwrite(transform func(any) any, params ...core.SchemaParams) *ZodCheckOverwrite {
	def := &ZodCheckOverwriteDef{
		ZodCheckDef: core.ZodCheckDef{Check: "overwrite"},
		Transform:   transform,
	}

	// Apply schema parameters for error customization
	core.ApplySchemaParams(&def.ZodCheckDef, params...)

	internals := &ZodCheckOverwriteInternals{
		ZodCheckInternals: core.ZodCheckInternals{
			Def: &def.ZodCheckDef,
		},
		Def: def,
	}

	// Set up the transformation function
	internals.Check = func(payload *core.ParsePayload) {
		// Apply transformation to payload value
		payload.Value = def.Transform(payload.Value)
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
