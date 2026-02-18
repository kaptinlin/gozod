package main

import (
	"errors"
	"fmt"

	"github.com/kaptinlin/gozod"
	"github.com/kaptinlin/gozod/locales"
)

func main() {
	// Set the global error language to Chinese.
	gozod.SetConfig(locales.ZhCN())

	// Define a schema to validate user data.
	userSchema := gozod.Object(gozod.ObjectSchema{
		"username": gozod.String().Min(3),
		"email":    gozod.String().Email(),
	}).Strict() // Use Strict() to disallow extra fields.

	// Prepare some invalid data to trigger validation errors.
	invalidData := map[string]any{
		"username": "ab",
		"email":    "invalid-email",
		"extra":    "field",
	}

	// Parse the data. The resulting error will have Chinese messages.
	_, err := userSchema.Parse(invalidData)
	if err == nil {
		return // Should not happen with invalid data.
	}

	if zErr, ok := errors.AsType[*gozod.ZodError](err); ok {
		// PrettifyError provides a human-readable summary of all issues.
		fmt.Println("--- Prettified Error (in Chinese) ---")
		fmt.Println(gozod.PrettifyError(zErr))

		// FlattenError is useful for displaying errors in a form UI.
		fmt.Println("\n--- Flattened Error (in Chinese) ---")
		flat := gozod.FlattenError(zErr)
		fmt.Printf("Form-level errors: %v\n", flat.FormErrors)
		fmt.Printf("Field-level errors: %v\n", flat.FieldErrors)
	}
}
