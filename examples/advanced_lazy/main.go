package main

import (
	"fmt"

	"github.com/kaptinlin/gozod"
)

// Demonstrates using Lazy to defer construction of a Struct schema.
func main() {
	fmt.Println("Lazy + Struct example")

	type Config struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	}

	// Lazy schema builds Config validation only on first Parse.
	cfgSchema := gozod.Lazy(func() *gozod.ZodStruct[Config, Config] {
		fmt.Println("inner struct schema initialized")
		return gozod.Struct[Config](gozod.StructSchema{
			"host": gozod.String().Min(1),
			"port": gozod.Int().Min(1).Max(65535),
		})
	})

	sample := Config{Host: "localhost", Port: 8080}
	if v, err := cfgSchema.Parse(sample); err != nil {
		fmt.Println("validation failed:", err)
	} else {
		fmt.Printf("✓ parsed config -> %+v\n", v)
	}
}
