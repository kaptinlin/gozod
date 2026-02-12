package core

// ZodCheck represents the interface for any validation constraint.
type ZodCheck interface {
	Zod() *ZodCheckInternals
}

// ZodCheckInternals contains the internal state and logic for a check.
type ZodCheckInternals struct {
	Def      *ZodCheckDef
	Issue    *ZodIssueBase
	Check    ZodCheckFn
	OnAttach []func(schema any)
	When     func(payload *ParsePayload) bool
}

// Zod implements ZodCheck.
func (c *ZodCheckInternals) Zod() *ZodCheckInternals {
	return c
}

// ZodCheckDef defines the static configuration for a validation check.
type ZodCheckDef struct {
	Check string
	Error *ZodErrorMap
	Abort bool
}

// CheckParams defines parameters for attaching a validation check.
type CheckParams struct {
	Error string
}

// CustomParams represents parameters for custom validation checks.
type CustomParams struct {
	Error  any            `json:"error,omitempty"`
	Abort  bool           `json:"abort,omitempty"`
	Path   []any          `json:"path,omitempty"`
	When   ZodWhenFn      `json:"-"`
	Params map[string]any `json:"params,omitempty"`
}

// ZodCustomParams is a type alias for CustomParams.
type ZodCustomParams = CustomParams

// ZodCheckFn defines the signature for a validation function.
type ZodCheckFn func(payload *ParsePayload)

// ZodWhenFn defines the signature for a conditional predicate.
type ZodWhenFn func(payload *ParsePayload) bool

// ZodRefineFn defines the signature for a type-safe refinement.
type ZodRefineFn[T any] func(value T) bool
