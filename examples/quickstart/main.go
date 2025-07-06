package main

import (
	"fmt"

	"github.com/kaptinlin/gozod"
)

func main() {
	fmt.Println("GoZod Quickstart")

	// 1) Validate a simple string (length 2..50).
	nameSchema := gozod.String().Min(2).Max(50)
	if v, err := nameSchema.Parse("Alice"); err == nil {
		fmt.Printf("✓ name: %s\n", v)
	}

	// 2) Validate structured data (map-based object).
	userSchema := gozod.Object(gozod.ObjectSchema{
		"name": gozod.String().Min(2),
		"age":  gozod.Int().Min(0).Max(120),
	})
	user := map[string]any{"name": "Bob", "age": 30}
	if v, err := userSchema.Parse(user); err == nil {
		fmt.Printf("✓ user: %+v\n", v)
	}

	// 3) Demonstrate validation failure.
	if _, err := nameSchema.Parse("A"); err != nil {
		fmt.Printf("✗ invalid name -> %v\n", err)
	}
}
