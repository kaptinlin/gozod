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

	// WithFormat selects which struct tag supplies field names. These
	// appear in validation error paths and JSON Schema output. Handy when the
	// struct is decoded from YAML/TOML instead of JSON.
	type Account struct {
		UserName string `gozod:"min=3" json:"userName" yaml:"user_name"`
	}

	yamlSchema := gozod.FromStruct[Account](gozod.WithFormat("yaml"))
	if _, err := yamlSchema.Parse(Account{UserName: "ab"}); err != nil {
		// Error path reads "user_name" (yaml) rather than "userName" (json).
		fmt.Printf("✗ yaml-named field failed: %v\n", err)
	}
}
