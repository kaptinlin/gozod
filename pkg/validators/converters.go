package validators

import (
	"fmt"

	"github.com/kaptinlin/gozod/core"
)

// ToRefineFn converts a ZodValidator to a ZodRefineFn for use with GoZod schemas
func ToRefineFn[T any](validator ZodValidator[T]) core.ZodRefineFn[T] {
	return func(value T) bool {
		return validator.Validate(value)
	}
}

// ToCheckFn converts a ZodValidator to a ZodCheckFn for use with GoZod schemas
func ToCheckFn[T any](validator ZodValidator[T], ctx *core.ParseContext) core.ZodCheckFn {
	return func(payload *core.ParsePayload) {
		value, ok := payload.GetValue().(T)
		if !ok {
			payload.AddIssueWithCode(core.InvalidType, "Type mismatch in validator")
			return
		}

		// Check if it's a detailed validator first (supports multi-language)
		if detailedValidator, ok := validator.(ZodDetailedValidator[T]); ok {
			detailedValidator.ValidateDetail(value, payload)
			return
		}

		// Fall back to simple validator
		if !validator.Validate(value) {
			errorMsg := validator.ErrorMessage(ctx)
			if errorMsg == "" {
				errorMsg = fmt.Sprintf("Validation failed for %s", validator.Name())
			}
			payload.AddIssueWithCode(core.Custom, errorMsg)
		}
	}
}

// ToCheckFnWithParam converts a ZodParameterizedValidator to a ZodCheckFn with parameter
func ToCheckFnWithParam[T any](validator ZodParameterizedValidator[T], param string, ctx *core.ParseContext) core.ZodCheckFn {
	return func(payload *core.ParsePayload) {
		value, ok := payload.GetValue().(T)
		if !ok {
			payload.AddIssueWithCode(core.InvalidType, "Type mismatch in parameterized validator")
			return
		}

		if !validator.ValidateParam(value, param) {
			errorMsg := validator.ErrorMessageWithParam(ctx, param)
			if errorMsg == "" {
				errorMsg = fmt.Sprintf("Validation failed for %s with param %s", validator.Name(), param)
			}
			payload.AddIssueWithCode(core.Custom, errorMsg)
		}
	}
}
