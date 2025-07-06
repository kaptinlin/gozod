package main

import (
	"fmt"

	"github.com/kaptinlin/gozod"
)

func main() {
	fmt.Println("Intersection examples")

	// Combine length constraints: 3 <= len <= 10.
	schema := gozod.Intersection(
		gozod.String().Min(3),
		gozod.String().Max(10),
	)

	for _, s := range []string{"hello", "hi", "this is very long"} {
		if v, err := schema.Parse(s); err != nil {
			fmt.Printf("✗ %s -> %v\n", s, err)
		} else {
			fmt.Printf("✓ %s -> %s\n", s, v)
		}
	}
}
