package main

import (
	"fmt"

	"github.com/kaptinlin/gozod"
)

type User struct {
	Name  string `validate:"required,min=2,max=50"`
	Email string `validate:"required,email"`
	Age   int    `validate:"min=18,max=120"`
}

func main() {
	// Use custom tag name "validate" instead of default "gozod"
	schema := gozod.FromStruct[User](gozod.WithTagName("validate"))

	// Valid user
	user := User{
		Name:  "Alice Smith",
		Email: "alice@example.com",
		Age:   28,
	}

	result, err := schema.Parse(user)
	if err != nil {
		fmt.Printf("Validation error: %v\n", err)
		return
	}

	fmt.Printf("✓ Valid user: %+v\n", result)

	// Invalid user
	invalidUser := User{
		Name:  "A",
		Email: "invalid-email",
		Age:   15,
	}

	_, err = schema.Parse(invalidUser)
	if err != nil {
		fmt.Printf("✗ Validation failed: %v\n", err)
	}
}
