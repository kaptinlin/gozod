// Package core contains the low-level building blocks shared by every schema
// implementation.  This file groups **validation check** primitives: the
// interfaces, configuration objects and function signatures required to plug
// additional constraints into any Zod schema.
//
// Naming convention
//
//	ZodCheck*   – general check abstractions (interface / internals / def)
//	ZodCheckFn  – validation function executed at runtime
//	ZodWhenFn   – predicate that decides whether the check should run
//	ZodRefineFn – type-safe helper for simple boolean refinements
//
// A *check* is therefore a self-contained unit composed of metadata +
// execution function.  Higher-level packages (`internal/checks`, individual
// type packages, …) create concrete checks by embedding / configuring these
// primitives.
package core

// =============================================================================
// VALIDATION CHECK INTERFACE & INTERNALS
// =============================================================================

// ZodCheck represents the interface for any validation constraint.
// All checks (e.g., min length, regex) must implement this interface to be
// attached to a schema.
type ZodCheck interface {
	GetZod() *ZodCheckInternals
}

// ZodCheckInternals contains the internal state and logic for a check.
// This structure holds all the necessary data for executing a validation check
type ZodCheckInternals struct {
	Def      *ZodCheckDef                     // The check's definition and metadata.
	Issc     *ZodIssueBase                    // A template for issues this check might create.
	Check    ZodCheckFn                       // The validation function to execute.
	OnAttach []func(schema any)               // Callbacks executed when the check is attached to a schema.
	When     func(payload *ParsePayload) bool // A predicate to conditionally run the check.
}

// GetZod implements the ZodCheck interface.
// Returns the internal structure itself following TypeScript pattern
func (c *ZodCheckInternals) GetZod() *ZodCheckInternals {
	return c
}

// =============================================================================
// CHECK DEFINITION & PARAMETERS
// =============================================================================

// ZodCheckDef defines the static configuration for a validation check.
// Contains the basic configuration for any validation check
type ZodCheckDef struct {
	Check string       // The unique identifier for the check (e.g., "min_length").
	Error *ZodErrorMap // An optional custom error mapping function.
	Abort bool         // If true, parsing aborts if this check fails.
}

// CheckParams defines the parameters for attaching a validation check.
// This is a simplified version of SchemaParams for check-specific configuration.
type CheckParams struct {
	Error string // Custom error message.
}

// =============================================================================
// CHECK FUNCTION SIGNATURES
// =============================================================================

// ZodCheckFn defines the signature for a validation execution function.
// This function performs the actual validation logic on a ParsePayload.
type ZodCheckFn func(payload *ParsePayload)

// ZodWhenFn defines the signature for a conditional predicate function.
// It is used to determine if a check should be executed based on the payload.
type ZodWhenFn func(payload *ParsePayload) bool

// ZodRefineFn defines the signature for a simple, type-safe refinement.
// It's a helper for checks that only need to return a boolean result.
type ZodRefineFn[T any] func(value T) bool
