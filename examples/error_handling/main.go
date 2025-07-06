package main

import (
	"errors"
	"fmt"

	"github.com/kaptinlin/gozod"
)

func main() {
	fmt.Println("Error handling examples")

	// Schema: string must be at least 5 chars and a valid email.
	schema := gozod.String().Min(5).Email()

	// Invalid input.
	_, err := schema.Parse("hi")
	if err == nil {
		return // should not happen
	}

	// 1) Check error type using IsZodError helper.
	var zErr *gozod.ZodError
	if gozod.IsZodError(err, &zErr) {
		fmt.Println("--- Issues slice ---")
		for _, is := range zErr.Issues {
			fmt.Printf("path=%v code=%s msg=%s\n", is.Path, is.Code, is.Message)
		}

		// 2) Pretty string.
		fmt.Println("--- PrettifyError ---")
		fmt.Println(gozod.PrettifyError(zErr))

		// 3) Flattened errors for forms.
		flat := gozod.FlattenError(zErr)
		fmt.Println("--- FlattenError.fieldErrors ---")
		fmt.Printf("%#v\n", flat.FieldErrors)
	}

	// Demonstrate standard errors.As as alternative.
	if errors.As(err, &zErr) {
		fmt.Println("Extracted via errors.As ->", zErr.Error())
	}
}
