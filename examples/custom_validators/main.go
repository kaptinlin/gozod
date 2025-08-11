package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/kaptinlin/gozod"
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/pkg/validators"
	"github.com/kaptinlin/gozod/types"
)

// ===== Custom Validator Implementations =====

// UniqueUsernameValidator - Basic string validator
type UniqueUsernameValidator struct{}

func (v *UniqueUsernameValidator) Name() string {
	return "unique_username"
}

func (v *UniqueUsernameValidator) Validate(username string) bool {
	// Simulate database check - in production, check against real database
	takenUsernames := map[string]bool{
		"admin": true,
		"root":  true,
		"user":  true,
	}
	return !takenUsernames[strings.ToLower(username)]
}

func (v *UniqueUsernameValidator) ErrorMessage(ctx *core.ParseContext) string {
	return "Username is already taken"
}

// MinValueValidator - Parameterized int validator
type MinValueValidator struct{}

func (v *MinValueValidator) Name() string {
	return "min_value"
}

func (v *MinValueValidator) Validate(value int) bool {
	return value >= 0 // Default minimum
}

func (v *MinValueValidator) ErrorMessage(ctx *core.ParseContext) string {
	return "Value must be non-negative"
}

func (v *MinValueValidator) ValidateParam(value int, param string) bool {
	if minValue, err := strconv.Atoi(param); err == nil {
		return value >= minValue
	}
	return false
}

func (v *MinValueValidator) ErrorMessageWithParam(ctx *core.ParseContext, param string) string {
	return fmt.Sprintf("Value must be at least %s", param)
}

// PrefixValidator - Parameterized string validator
type PrefixValidator struct{}

func (v *PrefixValidator) Name() string {
	return "prefix"
}

func (v *PrefixValidator) Validate(s string) bool {
	return len(s) > 0
}

func (v *PrefixValidator) ErrorMessage(ctx *core.ParseContext) string {
	return "Value must not be empty"
}

func (v *PrefixValidator) ValidateParam(s string, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

func (v *PrefixValidator) ErrorMessageWithParam(ctx *core.ParseContext, param string) string {
	return fmt.Sprintf("Value must start with '%s'", param)
}

// ===== Register Validators =====

func registerCustomValidators() {
	// Register string validators
	if err := validators.Register(&UniqueUsernameValidator{}); err != nil {
		log.Fatalf("Failed to register UniqueUsernameValidator: %v", err)
	}
	if err := validators.Register(&PrefixValidator{}); err != nil {
		log.Fatalf("Failed to register PrefixValidator: %v", err)
	}

	// Register int validators
	if err := validators.Register(&MinValueValidator{}); err != nil {
		log.Fatalf("Failed to register MinValueValidator: %v", err)
	}
}

// ===== Example Structs =====

type User struct {
	Username string `gozod:"required,unique_username"`
	Email    string `gozod:"required,email"`
	Age      int    `gozod:"required,min_value=18"`
}

type Product struct {
	SKU   string `gozod:"required,prefix=PROD"`
	Name  string `gozod:"required,min=3,max=100"`
	Price int    `gozod:"required,min_value=1"`
}

// ===== Main Demo =====

func main() {
	// Register validators at startup
	registerCustomValidators()

	fmt.Println("=== Custom Validators Example ===")

	// Demo 1: User validation with custom validators
	demoUserValidation()

	// Demo 2: Product validation with parameterized validators
	demoProductValidation()

	// Demo 3: Programmatic usage without struct tags
	demoProgrammaticValidation()
}

func demoUserValidation() {
	fmt.Println("1. User Validation")
	fmt.Println("------------------")

	schema := types.FromStruct[User]()

	// Valid user
	validUser := User{
		Username: "johndoe",
		Email:    "john@example.com",
		Age:      25,
	}

	if result, err := schema.Parse(validUser); err == nil {
		fmt.Printf("✅ Valid user: %s (age: %d)\n", result.Username, result.Age)
	} else {
		fmt.Printf("❌ Validation failed: %v\n", err)
	}

	// Invalid: taken username
	invalidUser := User{
		Username: "admin",
		Email:    "admin@example.com",
		Age:      25,
	}

	if _, err := schema.Parse(invalidUser); err != nil {
		fmt.Printf("✅ Correctly rejected taken username: %v\n", err)
	}

	// Invalid: underage
	youngUser := User{
		Username: "younguser",
		Email:    "young@example.com",
		Age:      16,
	}

	if _, err := schema.Parse(youngUser); err != nil {
		fmt.Printf("✅ Correctly rejected underage user: %v\n", err)
	}

	fmt.Println()
}

func demoProductValidation() {
	fmt.Println("2. Product Validation")
	fmt.Println("---------------------")

	schema := types.FromStruct[Product]()

	// Valid product
	validProduct := Product{
		SKU:   "PROD-12345",
		Name:  "Awesome Widget",
		Price: 99,
	}

	if result, err := schema.Parse(validProduct); err == nil {
		fmt.Printf("✅ Valid product: %s ($%d)\n", result.Name, result.Price)
	} else {
		fmt.Printf("❌ Validation failed: %v\n", err)
	}

	// Invalid: wrong prefix
	invalidProduct := Product{
		SKU:   "ITEM-12345",
		Name:  "Wrong SKU Product",
		Price: 50,
	}

	if _, err := schema.Parse(invalidProduct); err != nil {
		fmt.Printf("✅ Correctly rejected wrong SKU prefix: %v\n", err)
	}

	// Invalid: zero price
	freeProduct := Product{
		SKU:   "PROD-FREE",
		Name:  "Free Product",
		Price: 0,
	}

	if _, err := schema.Parse(freeProduct); err != nil {
		fmt.Printf("✅ Correctly rejected zero price: %v\n", err)
	}

	fmt.Println()
}

func demoProgrammaticValidation() {
	fmt.Println("3. Programmatic Validation")
	fmt.Println("--------------------------")

	// Get validator from registry
	usernameValidator, exists := validators.Get[string]("unique_username")
	if !exists {
		log.Fatal("unique_username validator not found")
	}

	// Create schema with custom validator
	schema := gozod.String().
		Min(3).
		Max(20).
		Refine(validators.ToRefineFn(usernameValidator))

	// Test various usernames
	testUsernames := []string{"newuser", "admin", "validname"}

	for _, username := range testUsernames {
		if _, err := schema.Parse(username); err != nil {
			fmt.Printf("❌ '%s' rejected: %v\n", username, err)
		} else {
			fmt.Printf("✅ '%s' accepted\n", username)
		}
	}

	fmt.Println()

	// Demonstrate parameterized validator
	minValueValidator, _ := validators.GetAny("min_value")
	if paramValidator, ok := minValueValidator.(validators.ZodParameterizedValidator[int]); ok {
		values := []int{-5, 0, 10, 20}
		minParam := "10"

		fmt.Printf("Testing values against minimum of %s:\n", minParam)
		for _, val := range values {
			if paramValidator.ValidateParam(val, minParam) {
				fmt.Printf("✅ %d >= %s\n", val, minParam)
			} else {
				fmt.Printf("❌ %d < %s\n", val, minParam)
			}
		}
	}
}
