package gozod

import (
	"fmt"
	"regexp"
)

// Version is the current version of the library
var Version = "0.1.0"

// =============================================================================
// ZOD TYPE CONSTANTS
// =============================================================================

const (
	ZodTypeString        = "string"
	ZodTypeNumber        = "number"
	ZodTypeNaN           = "nan"
	ZodTypeInteger       = "integer"
	ZodTypeBigInt        = "bigint"
	ZodTypeBool          = "bool"
	ZodTypeDate          = "date"
	ZodTypeNil           = "nil"
	ZodTypeAny           = "any"
	ZodTypeUnknown       = "unknown"
	ZodTypeNever         = "never"
	ZodTypeVoid          = "void"
	ZodTypeArray         = "array"
	ZodTypeObject        = "object"
	ZodTypeStruct        = "struct"
	ZodTypeStringBool    = "stringbool"
	ZodTypeUnion         = "union"
	ZodTypeDiscriminated = "discriminated_union"
	ZodTypeIntersection  = "intersection"
	ZodTypeTuple         = "tuple"
	ZodTypeRecord        = "record"
	ZodTypeMap           = "map"
	ZodTypeSet           = "set"
	ZodTypeFunction      = "function"
	ZodTypeLazy          = "lazy"
	ZodTypeLiteral       = "literal"
	ZodTypeEnum          = "enum"
	ZodTypeOptional      = "optional"
	ZodTypeNilable       = "nilable"
	ZodTypeDefault       = "default"
	ZodTypePrefault      = "prefault"
	ZodTypePipeline      = "pipeline"
	ZodTypeTransform     = "transform"
	ZodTypePipe          = "pipe"
	ZodTypeCustom        = "custom"
	ZodTypeCheck         = "check"
	ZodTypeRefine        = "refine"
)

// =============================================================================
// CORE INTERFACES
// =============================================================================

// Cloneable defines the ability to copy type-specific state
type Cloneable interface {
	CloneFrom(source any)
}

// Coercible defines the ability to perform type coercion
type Coercible interface {
	Coerce(input interface{}) (output interface{}, success bool)
}

// ZodType defines the common interface for all Zod validators
type ZodType[In, Out any] interface {
	// Core parsing methods
	Parse(input any, ctx ...*ParseContext) (any, error)
	MustParse(input any, ctx ...*ParseContext) any

	// Modifiers
	Nilable() ZodType[any, any]

	// Validation and transformation
	RefineAny(fn func(any) bool, params ...SchemaParams) ZodType[any, any]
	TransformAny(fn func(any, *RefinementContext) (any, error)) ZodType[any, any]
	Pipe(out ZodType[any, any]) ZodType[In, any]

	// Internal access
	GetInternals() *ZodTypeInternals
	Unwrap() ZodType[any, any]
}

// =============================================================================
// SCHEMA DEFINITIONS
// =============================================================================

// ZodTypeDef defines the basic schema definition structure
type ZodTypeDef struct {
	Type   string
	Error  *ZodErrorMap
	Checks []ZodCheck
}

// ZodTypeInternals contains schema internal state
type ZodTypeInternals struct {
	Version string
	Type    string
	Checks  []ZodCheck
	Parse   func(payload *ParsePayload, ctx *ParseContext) *ParsePayload

	// Core validation flags
	Coerce   bool
	Optional bool
	Nilable  bool

	// Optionality fields
	OptIn  string
	OptOut string

	// Constructor and configuration
	Constructor func(def *ZodTypeDef) ZodType[any, any]
	Values      map[interface{}]struct{}
	Pattern     *regexp.Regexp
	Error       *ZodErrorMap
	Bag         map[string]interface{}
}

// SchemaParams contains optional parameters for schema creation
type SchemaParams struct {
	Description   string
	Error         interface{} // string or ZodErrorMap
	Coerce        bool
	Abort         bool
	Path          []string
	Params        map[string]interface{}
	UnionFallback bool
}

// =============================================================================
// UTILITY TYPES
// =============================================================================

// ObjectSchema defines object shape for structured validation
type ObjectSchema = map[string]ZodType[any, any]

// Schema is a convenient alias for ZodType[any, any]
type Schema = ZodType[any, any]

// =============================================================================
// CHECK AND REFINEMENT FUNCTIONS
// =============================================================================

// CheckFn defines the function signature for validation checks
type CheckFn func(payload *ParsePayload)

// CheckContext provides context for validation checks
type CheckContext struct {
	Value interface{}
	Path  []interface{}
}

// GetContext returns a CheckContext for the payload
func (payload *ParsePayload) GetContext() *CheckContext {
	return &CheckContext{
		Value: payload.Value,
		Path:  payload.Path,
	}
}

// RefineFn defines the function signature for refinement validation
type RefineFn[T any] func(value T) bool

// RefinementContext provides context for validation and transformation
type RefinementContext struct {
	*ParseContext
	Value    any
	AddIssue func(issue ZodIssue)
}

// TransformFn defines the function signature for transformation
type TransformFn[T any, R any] func(T, *RefinementContext) (R, error)

// =============================================================================
// CONSTRUCTOR FUNCTIONS
// =============================================================================

// NewZodTypeDef creates a new ZodTypeDef with the given type
func NewZodTypeDef(typeName string) *ZodTypeDef {
	return &ZodTypeDef{
		Type:   typeName,
		Checks: make([]ZodCheck, 0),
	}
}

// NewZodTypeInternals creates a new ZodTypeInternals with the given definition
func NewZodTypeInternals(def *ZodTypeDef) *ZodTypeInternals {
	return &ZodTypeInternals{
		Version: Version,
		Type:    def.Type,
		Checks:  make([]ZodCheck, len(def.Checks)),
		Values:  make(map[interface{}]struct{}),
		Error:   def.Error,
		Bag:     make(map[string]interface{}),
	}
}

// =============================================================================
// SCHEMA PARAMETER HELPERS
// =============================================================================

// WithError creates a SchemaParams with custom error message
func WithError(message string) SchemaParams {
	return SchemaParams{Error: message}
}

// WithCoercion creates a SchemaParams with coercion enabled
func WithCoercion() SchemaParams {
	return SchemaParams{Coerce: true}
}

// =============================================================================
// GENERIC OPERATIONS
// =============================================================================

// AddCheck adds a validation check to any ZodType and returns new instance
func AddCheck[T interface{ GetInternals() *ZodTypeInternals }](schema T, check ZodCheck) ZodType[any, any] {
	internals := schema.GetInternals()

	// Create new type definition with updated checks
	newDef := &ZodTypeDef{
		Type:   internals.Type,
		Error:  internals.Error,
		Checks: append(make([]ZodCheck, len(internals.Checks)), internals.Checks...),
	}
	newDef.Checks = append(newDef.Checks, check)

	// Use existing constructor
	if internals.Constructor != nil {
		newSchema := internals.Constructor(newDef)
		newInternals := newSchema.GetInternals()

		// Preserve important state flags
		newInternals.Nilable = internals.Nilable
		newInternals.Optional = internals.Optional
		newInternals.Coerce = internals.Coerce

		// Preserve pattern state
		if internals.Pattern != nil {
			newInternals.Pattern = internals.Pattern
		}

		// Preserve values state
		if len(internals.Values) > 0 {
			newInternals.Values = make(map[interface{}]struct{})
			for k, v := range internals.Values {
				newInternals.Values[k] = v
			}
		}

		// Preserve bag state
		if len(internals.Bag) > 0 {
			if newInternals.Bag == nil {
				newInternals.Bag = make(map[string]interface{})
			}
			for k, v := range internals.Bag {
				newInternals.Bag[k] = v
			}
		}

		// Use Cloneable interface to copy type-specific state
		if cloneable, ok := newSchema.(Cloneable); ok {
			if sourceAny := any(schema); sourceAny != nil {
				if sourceCloneable, ok := sourceAny.(Cloneable); ok {
					cloneable.CloneFrom(sourceCloneable)
				}
			}
		}

		// Execute onattach callbacks
		if check != nil {
			if checkInternals := check.GetZod(); checkInternals != nil {
				for _, fn := range checkInternals.OnAttach {
					fn(newSchema)
				}
			}
		}

		return newSchema
	}

	panic(fmt.Sprintf("No constructor found for type: %T", schema))
}

// Clone creates a new instance of any ZodType with optional definition modifications
func Clone[T interface{ GetInternals() *ZodTypeInternals }](schema T, modifyDef func(*ZodTypeDef)) ZodType[any, any] {
	internals := schema.GetInternals()

	// Create new type definition
	newDef := &ZodTypeDef{
		Type:   internals.Type,
		Error:  internals.Error,
		Checks: append(make([]ZodCheck, len(internals.Checks)), internals.Checks...),
	}

	// Apply modifications if provided
	if modifyDef != nil {
		modifyDef(newDef)
	}

	// Use existing constructor
	if internals.Constructor != nil {
		newSchema := internals.Constructor(newDef)
		newInternals := newSchema.GetInternals()

		// Preserve important state flags
		newInternals.Nilable = internals.Nilable
		newInternals.Optional = internals.Optional
		newInternals.Coerce = internals.Coerce

		// Preserve pattern state
		if internals.Pattern != nil {
			newInternals.Pattern = internals.Pattern
		}

		// Preserve values state
		if len(internals.Values) > 0 {
			newInternals.Values = make(map[interface{}]struct{})
			for k, v := range internals.Values {
				newInternals.Values[k] = v
			}
		}

		// Use Cloneable interface to copy type-specific state
		if cloneable, ok := newSchema.(Cloneable); ok {
			if sourceAny := any(schema); sourceAny != nil {
				if sourceCloneable, ok := sourceAny.(Cloneable); ok {
					cloneable.CloneFrom(sourceCloneable)
				}
			}
		}

		return newSchema
	}

	panic(fmt.Sprintf("No constructor found for type: %T", schema))
}

// =============================================================================
// SCHEMA INITIALIZATION AND VALIDATION
// =============================================================================

// initZodType initializes the common fields of a ZodType
func initZodType[T ZodType[any, any]](schema T, def *ZodTypeDef) {
	internals := schema.GetInternals()

	// Initialize base internals
	internals.Version = Version
	internals.Type = def.Type
	internals.Error = def.Error

	// Initialize checks from definition
	if len(def.Checks) > 0 {
		internals.Checks = make([]ZodCheck, len(def.Checks))
		copy(internals.Checks, def.Checks)
	} else {
		internals.Checks = make([]ZodCheck, 0)
	}

	// Initialize Values map if not already done
	if internals.Values == nil {
		internals.Values = make(map[interface{}]struct{})
	}

	// Run onattach callbacks for all checks
	for _, check := range internals.Checks {
		if check != nil {
			if checkInternals := check.GetZod(); checkInternals != nil {
				for _, fn := range checkInternals.OnAttach {
					fn(any(schema).(ZodType[any, any]))
				}
			}
		}
	}
}

// runChecks executes all checks on a payload synchronously
func runChecks(payload *ParsePayload, checks []ZodCheck, ctx *ParseContext) *ParsePayload {
	isAborted := aborted(*payload, 0)

	for _, check := range checks {
		if check != nil {
			if checkInternals := check.GetZod(); checkInternals != nil {
				// Check conditional execution
				if checkInternals.When != nil && !checkInternals.When(payload) {
					continue
				}

				// Skip if already aborted
				if isAborted {
					continue
				}

				currLen := len(payload.Issues)

				// Execute check
				checkInternals.Check(payload)

				// Check if new issues were added and should abort
				if len(payload.Issues) > currLen {
					if checkInternals.Def.Abort {
						isAborted = true
					}
				}
			}
		}
	}

	return payload
}

// aborted checks if parsing should be aborted
func aborted(x ParsePayload, startIndex int) bool {
	for i := startIndex; i < len(x.Issues); i++ {
		if !x.Issues[i].Continue {
			return true
		}
	}
	return false
}
