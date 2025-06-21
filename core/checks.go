package core

// =============================================================================
// VALIDATION CHECK INTERFACES
// =============================================================================

// ZodCheck represents validation constraint interface
// All validation checks must implement this interface
type ZodCheck interface {
	GetZod() *ZodCheckInternals
}

// ZodCheckInternals contains check internal state and configuration
// This structure holds all the necessary data for executing a validation check
type ZodCheckInternals struct {
	Def      *ZodCheckDef                     // Check definition with metadata
	Issc     *ZodIssueBase                    // Issues this check might throw
	Check    ZodCheckFn                       // Validation function to execute
	OnAttach []func(schema any)               // Schema attachment callbacks
	When     func(payload *ParsePayload) bool // Conditional execution predicate
}

// GetZod implements ZodCheck interface
// Returns the internal structure itself following TypeScript pattern
func (c *ZodCheckInternals) GetZod() *ZodCheckInternals {
	return c
}

// ZodCheckDef defines a check definition
// Contains the basic configuration for any validation check
type ZodCheckDef struct {
	Check string       // Check type identifier (e.g., "min_length", "regex")
	Error *ZodErrorMap // Custom error mapping function
	Abort bool         // Whether to abort parsing on validation failure
}

// ZodCheckFn defines validation execution function
// This function performs the actual validation logic
type ZodCheckFn func(payload *ParsePayload)

// ZodWhenFn defines conditional function type
// Used to determine when a check should be executed
type ZodWhenFn func(payload *ParsePayload) bool

// =============================================================================
// FUNCTION TYPE DEFINITIONS
// =============================================================================

// CheckFn defines the function signature for validation checks
// Simple boolean validation function
type CheckFn func(payload *ParsePayload)

// RefineFn defines the function signature for refinement validation
// Type-safe refinement function with generic type parameter
type RefineFn[T any] func(value T) bool

// TransformFn defines the function signature for transformation
// Transforms input of type T to output of type R
type TransformFn[T any, R any] func(T, *RefinementContext) (R, error)

// =============================================================================
// CONSTRUCTOR FUNCTION TYPES
// =============================================================================

// ConstructorFn defines the signature for schema constructor functions
// Used to create new instances of schema types
type ConstructorFn[T ZodType[any, any]] func(def *ZodTypeDef) T

// ModifierFn defines the signature for schema modifier functions
// Used to add modifiers like Optional, Nilable, etc.
type ModifierFn[T ZodType[any, any]] func(schema T) ZodType[any, any]
