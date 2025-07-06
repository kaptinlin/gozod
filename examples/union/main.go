package main

import (
	"fmt"

	"github.com/kaptinlin/gozod"
)

func main() {
	fmt.Println("Union type example")

	// Union schema: accepts either string or int.
	schema := gozod.Union([]any{gozod.String(), gozod.Int()})

	for _, input := range []any{"hello", 42, true} {
		if v, err := schema.Parse(input); err != nil {
			fmt.Printf("✗ %v (%T) -> %v\n", input, input, err)
		} else {
			fmt.Printf("✓ %v (%T) -> %v\n", input, input, v)
		}
	}
}
