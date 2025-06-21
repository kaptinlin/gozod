package core

// =============================================================================
// CORE TYPE INTERFACES
// =============================================================================

// ZodType defines the common interface for all Zod validators
// This is the foundational interface that all schema types must implement
type ZodType[In, Out any] interface {
	// Core parsing methods - these handle validation and type conversion
	Parse(input any, ctx ...*ParseContext) (any, error)
	MustParse(input any, ctx ...*ParseContext) any

	// Modifiers - these return modified versions of the schema
	Nilable() ZodType[any, any]

	// Validation and transformation - these add behavior to schemas
	RefineAny(fn func(any) bool, params ...any) ZodType[any, any]
	TransformAny(fn func(any, *RefinementContext) (any, error)) ZodType[any, any]
	Pipe(out ZodType[any, any]) ZodType[In, any]

	// Internal access - these provide access to internal state
	GetInternals() *ZodTypeInternals
	Unwrap() ZodType[any, any]
}

// Cloneable defines the ability to copy type-specific state
// Types implementing this can preserve their internal state during cloning
type Cloneable interface {
	CloneFrom(source any)
}

// Coercible defines type coercion capability
// Types implementing this can convert input values to their target type
type Coercible interface {
	Coerce(input any) (output any, success bool)
}

// =============================================================================
// SPECIALIZED INTERFACES
// =============================================================================

// ObjectSchemaInterface represents an object schema interface
// Provides access to the shape definition of object schemas
type ObjectSchemaInterface interface {
	ZodType[any, any]                    // Inherits basic schema functionality
	Shape() map[string]ZodType[any, any] // Returns the object's field definitions
}

// ArraySchemaInterface represents an array schema interface
// Provides access to element validation for array schemas
type ArraySchemaInterface interface {
	ZodType[any, any]           // Inherits basic schema functionality
	Element() ZodType[any, any] // Returns the element validation schema
}
