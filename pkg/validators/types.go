package validators

import (
	"github.com/kaptinlin/gozod/core"
)

// ZodValidator represents a basic validator interface for type T
type ZodValidator[T any] interface {
	// Validate performs basic validation on the value
	Validate(value T) bool

	// Name returns the validator's unique name
	Name() string

	// ErrorMessage returns a basic error message for validation failures
	ErrorMessage(ctx *core.ParseContext) string
}

// ZodParameterizedValidator extends ZodValidator to support parameters
type ZodParameterizedValidator[T any] interface {
	ZodValidator[T]

	// ValidateParam performs validation with a parameter
	ValidateParam(value T, param string) bool

	// ErrorMessageWithParam returns an error message with parameter context
	ErrorMessageWithParam(ctx *core.ParseContext, param string) string
}

// ZodDetailedValidator extends ZodValidator for complex validation scenarios
type ZodDetailedValidator[T any] interface {
	ZodValidator[T]

	// ValidateDetail performs detailed validation with access to ParsePayload
	ValidateDetail(value T, payload *core.ParsePayload)
}
