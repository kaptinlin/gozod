package main

import (
	"fmt"

	"github.com/kaptinlin/gozod"
)

// Demonstrates fundamental string & number validation helpers.
func main() {
	fmt.Println("Primitive type examples")

	// String: length 3‒10, must contain "go".
	strSchema := gozod.String().Min(3).Max(10).Includes("go")
	for _, s := range []string{"gozod", "g"} {
		if v, err := strSchema.Parse(s); err != nil {
			fmt.Printf("✗ %q -> %v\n", s, err)
		} else {
			fmt.Printf("✓ %q -> %s\n", s, v)
		}
	}

	// Number: 0‒100 inclusive & multiple of 2.
	numSchema := gozod.Number().Gte(0).Lte(100).MultipleOf(2)
	for _, n := range []float64{42, 3} {
		if v, err := numSchema.Parse(n); err != nil {
			fmt.Printf("✗ %v -> %v\n", n, err)
		} else {
			fmt.Printf("✓ %v -> %v\n", n, v)
		}
	}
}
