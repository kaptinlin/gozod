package main

import (
	"fmt"

	"github.com/kaptinlin/gozod"
)

func main() {
	fmt.Println("Object validation examples")

	// Define a nested object schema with modifiers.
	addressSchema := gozod.Object(gozod.ObjectSchema{
		"street": gozod.String().Min(3),
		"city":   gozod.String().Min(2),
		"zip":    gozod.String().RegexString(`^\d{5}$`),
	})

	userSchema := gozod.Object(gozod.ObjectSchema{
		"name":    gozod.String().Min(2),
		"age":     gozod.Int().Min(0).Max(150).Optional(), // optional field
		"email":   gozod.String().Email().Optional(),      // optional field
		"address": addressSchema,                          // nested object
	})

	// Valid input
	valid := map[string]any{
		"name": "Alice",
		"address": map[string]any{
			"street": "1 Infinite Loop",
			"city":   "Cupertino",
			"zip":    "95014",
		},
	}
	if v, err := userSchema.Parse(valid); err == nil {
		fmt.Printf("✓ valid user -> %+v\n", v)
	}

	// Missing required nested field (zip).
	invalid := map[string]any{
		"name": "Bob",
		"address": map[string]any{
			"street": "Main St",
			"city":   "NY",
		},
	}
	if _, err := userSchema.Parse(invalid); err != nil {
		fmt.Printf("✗ invalid user -> %v\n", err)
	}
}
