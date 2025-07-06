package main

import (
	"fmt"
	"time"

	"github.com/kaptinlin/gozod/coerce"
)

func main() {
	fmt.Println("Type coercion examples")

	// Example 1: Coerce various values to string.
	strSchema := coerce.String()
	for _, input := range []any{123, true, 3.14} {
		if v, _ := strSchema.Parse(input); v != "" {
			fmt.Printf("✓ %v (%T) -> %q\n", input, input, v)
		}
	}

	// Example 2: Coerce string/boolean to number (float64).
	numSchema := coerce.Number()
	for _, input := range []any{"42", true} {
		if v, _ := numSchema.Parse(input); v != 0 {
			fmt.Printf("✓ %v (%T) -> %v\n", input, input, v)
		}
	}

	// Example 3: Coerce various inputs to boolean.
	boolSchema := coerce.Bool()
	for _, input := range []any{"true", 1, 0, "false"} {
		if v, _ := boolSchema.Parse(input); true {
			fmt.Printf("✓ %v (%T) -> %v\n", input, input, v)
		}
	}

	// Example 4: Coerce timestamp / RFC3339 string to time.Time.
	timeSchema := coerce.Time()
	inputs := []any{int64(1700000000), "2023-12-25T15:30:00Z"}
	for _, input := range inputs {
		if v, err := timeSchema.Parse(input); err == nil {
			fmt.Printf("✓ %v (%T) -> %s\n", input, input, v.UTC().Format(time.RFC3339))
		}
	}
}
