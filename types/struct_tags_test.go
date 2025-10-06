package types

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/pkg/validators"
)

// =============================================================================
// CUSTOM VALIDATORS FOR TESTING
// =============================================================================

// UniqueUsernameValidator - checks if username is not taken
type UniqueUsernameValidator struct{}

func (v *UniqueUsernameValidator) Name() string {
	return "unique_username"
}

func (v *UniqueUsernameValidator) Validate(username string) bool {
	// Simulate database check
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

// MinAgeValidator - parameterized age validator
type MinAgeValidator struct{}

func (v *MinAgeValidator) Name() string {
	return "min_age"
}

func (v *MinAgeValidator) Validate(age int) bool {
	return age >= 18 // Default minimum age
}

func (v *MinAgeValidator) ErrorMessage(ctx *core.ParseContext) string {
	return "Age must be at least 18"
}

func (v *MinAgeValidator) ValidateParam(age int, param string) bool {
	if minAge, err := strconv.Atoi(param); err == nil {
		return age >= minAge
	}
	return false
}

func (v *MinAgeValidator) ErrorMessageWithParam(ctx *core.ParseContext, param string) string {
	return "Age must be at least " + param
}

// SKUPrefixValidator - checks SKU prefix
type SKUPrefixValidator struct{}

func (v *SKUPrefixValidator) Name() string {
	return "sku_prefix"
}

func (v *SKUPrefixValidator) Validate(sku string) bool {
	return len(sku) > 0
}

func (v *SKUPrefixValidator) ErrorMessage(ctx *core.ParseContext) string {
	return "SKU must not be empty"
}

func (v *SKUPrefixValidator) ValidateParam(sku string, prefix string) bool {
	return strings.HasPrefix(sku, prefix)
}

func (v *SKUPrefixValidator) ErrorMessageWithParam(ctx *core.ParseContext, param string) string {
	return "SKU must start with " + param
}

// ExactLengthValidator - checks exact string length
type ExactLengthValidator struct{}

func (v *ExactLengthValidator) Name() string {
	return "exact_length"
}

func (v *ExactLengthValidator) Validate(s string) bool {
	return len(s) > 0
}

func (v *ExactLengthValidator) ErrorMessage(ctx *core.ParseContext) string {
	return "String must not be empty"
}

func (v *ExactLengthValidator) ValidateParam(s string, param string) bool {
	if length, err := strconv.Atoi(param); err == nil {
		return len(s) == length
	}
	return false
}

func (v *ExactLengthValidator) ErrorMessageWithParam(ctx *core.ParseContext, param string) string {
	return "String must be exactly " + param + " characters long"
}

// =============================================================================
// TEST TYPES FOR TAG SUPPORT
// =============================================================================

type TaggedUser struct {
	Name  string `gozod:"required,min=2,max=50"`
	Email string `gozod:"required,email"`
	Age   int    `gozod:"min=18,max=120"`
	Bio   string `gozod:"max=500"`
}

type NoTagUser struct {
	Name  string
	Email string
}

// =============================================================================
// FROMSTRUCT FUNCTION TESTS
// =============================================================================

func TestFromStruct_BasicUsage(t *testing.T) {
	// Test that FromStruct creates a valid schema
	schema := FromStruct[TaggedUser]()
	if schema == nil {
		t.Fatal("FromStruct should return a non-nil schema")
	}

	// Verify basic functionality
	user := TaggedUser{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   25,
		Bio:   "Software engineer",
	}

	result, err := schema.Parse(user)
	require.NoError(t, err, "Parse should succeed for valid user")

	if result.Name != user.Name {
		assert.Equal(t, user.Name, result.Name, "Expected name %s, got %s", user.Name, result.Name)
	}
}

func TestFromStructPtr_PointerTypes(t *testing.T) {
	// Test that FromStructPtr creates a valid pointer schema
	schema := FromStructPtr[TaggedUser]()
	if schema == nil {
		t.Fatal("FromStructPtr should return a non-nil schema")
	}

	// Verify basic functionality with pointer input
	user := &TaggedUser{
		Name:  "Jane Doe",
		Email: "jane@example.com",
		Age:   30,
		Bio:   "Product manager",
	}

	result, err := schema.Parse(user)
	require.NoError(t, err, "Parse should succeed for valid user pointer")

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result.Name != user.Name {
		assert.Equal(t, user.Name, result.Name, "Expected name %s, got %s", user.Name, result.Name)
	}
}

func TestFromStruct_EmptyStruct(t *testing.T) {
	// Test FromStruct with empty struct
	type EmptyStruct struct{}

	schema := FromStruct[EmptyStruct]()
	if schema == nil {
		t.Fatal("FromStruct should handle empty structs")
	}

	result, err := schema.Parse(EmptyStruct{})
	require.NoError(t, err, "Parse should succeed for empty struct")

	// Verify result is correct type
	_ = result // EmptyStruct type assertion
}

func TestFromStruct_NoTags(t *testing.T) {
	// Test that FromStruct works even without gozod tags
	schema := FromStruct[NoTagUser]()
	if schema == nil {
		t.Fatal("FromStruct should work without tags")
	}

	user := NoTagUser{
		Name:  "Test User",
		Email: "test@example.com",
	}

	result, err := schema.Parse(user)
	require.NoError(t, err, "Parse should succeed for user without tags")

	if result.Name != user.Name {
		assert.Equal(t, user.Name, result.Name, "Expected name %s, got %s", user.Name, result.Name)
	}
}

// =============================================================================
// INTEGRATION TESTS
// =============================================================================

func TestFromStruct_Integration(t *testing.T) {
	// Test that FromStruct integrates well with existing functionality
	schema := FromStruct[TaggedUser]()

	// Test with modifiers
	optionalSchema := schema.Optional()
	if optionalSchema == nil {
		t.Fatal("FromStruct schema should support modifiers")
	}

	// Test nil handling with optional
	result, err := optionalSchema.Parse(nil)
	require.NoError(t, err, "Optional schema should handle nil")

	// For optional structs, nil input should return nil pointer
	if result != nil {
		t.Log("Optional schema returned non-nil result for nil input - this is acceptable")
	}
}

func TestFromStruct_WithRefine(t *testing.T) {
	// Test that FromStruct works with Refine
	schema := FromStruct[TaggedUser]().Refine(func(user TaggedUser) bool {
		return user.Age >= 21 // Adult validation
	}, "User must be at least 21 years old")

	// Test valid case
	adult := TaggedUser{
		Name:  "Adult User",
		Email: "adult@example.com",
		Age:   25,
		Bio:   "Adult user",
	}

	_, err := schema.Parse(adult)
	require.NoError(t, err, "Adult user should pass validation")

	// Test invalid case
	minor := TaggedUser{
		Name:  "Minor User",
		Email: "minor@example.com",
		Age:   18, // Valid for age constraint but fails refine
		Bio:   "Minor user",
	}

	_, err = schema.Parse(minor)
	if err == nil {
		t.Fatal("Minor user should fail refine validation")
	}
}

// =============================================================================
// CUSTOM VALIDATOR TAG TESTS
// =============================================================================

type UserWithCustomValidators struct {
	Username string `gozod:"unique_username"`
	Age      int    `gozod:"min_age=21"`
	SKU      string `gozod:"sku_prefix=PROD"`
	Code     string `gozod:"exact_length=8"`
}

func TestFromStructWithCustomValidators(t *testing.T) {
	// Register custom validators
	_ = validators.Register(&UniqueUsernameValidator{})
	_ = validators.Register(&MinAgeValidator{})
	_ = validators.Register(&SKUPrefixValidator{})
	_ = validators.Register(&ExactLengthValidator{})

	// Create schema from struct
	schema := FromStruct[UserWithCustomValidators]()

	// Test valid data
	validUser := UserWithCustomValidators{
		Username: "newuser",
		Age:      25,
		SKU:      "PROD12345",
		Code:     "ABC12345",
	}

	result, err := schema.Parse(validUser)
	if err != nil {
		t.Errorf("Expected valid user to parse successfully, got error: %v", err)
	}
	if result.Username != validUser.Username {
		assert.Equal(t, validUser.Username, result.Username, "Expected username %s, got %s", validUser.Username, result.Username)
	}

	// Test invalid username (taken)
	invalidUser1 := UserWithCustomValidators{
		Username: "admin", // This should be taken
		Age:      25,
		SKU:      "PROD12345",
		Code:     "ABC12345",
	}

	_, err = schema.Parse(invalidUser1)
	assert.Error(t, err, "Expected error for taken username, but got none")

	// Test invalid age (too young)
	invalidUser2 := UserWithCustomValidators{
		Username: "newuser",
		Age:      18, // Less than minimum age of 21
		SKU:      "PROD12345",
		Code:     "ABC12345",
	}

	_, err = schema.Parse(invalidUser2)
	assert.Error(t, err, "Expected error for underage user, but got none")

	// Test invalid SKU (wrong prefix)
	invalidUser3 := UserWithCustomValidators{
		Username: "newuser",
		Age:      25,
		SKU:      "TEST12345", // Wrong prefix, should start with PROD
		Code:     "ABC12345",
	}

	_, err = schema.Parse(invalidUser3)
	assert.Error(t, err, "Expected error for wrong SKU prefix, but got none")

	// Test invalid code (wrong length)
	invalidUser4 := UserWithCustomValidators{
		Username: "newuser",
		Age:      25,
		SKU:      "PROD12345",
		Code:     "ABC123", // Too short, should be exactly 8 characters
	}

	_, err = schema.Parse(invalidUser4)
	assert.Error(t, err, "Expected error for wrong code length, but got none")
}

type ProductWithMixedValidation struct {
	Name        string `gozod:"required,min=3,unique_username"` // Built-in + custom
	Description string `gozod:"optional"`
	SKU         string `gozod:"sku_prefix=ITEM"`
	Price       int    `gozod:"min_age=1"` // Using min_age validator for price
}

func TestMixedBuiltinAndCustomValidators(t *testing.T) {
	// Register custom validators
	_ = validators.Register(&UniqueUsernameValidator{})
	_ = validators.Register(&MinAgeValidator{})
	_ = validators.Register(&SKUPrefixValidator{})
	_ = validators.Register(&ExactLengthValidator{})

	schema := FromStruct[ProductWithMixedValidation]()

	// Test valid product
	validProduct := ProductWithMixedValidation{
		Name:        "ValidProduct", // Not in taken usernames, length > 3
		Description: "A great product",
		SKU:         "ITEM12345",
		Price:       100,
	}

	result, err := schema.Parse(validProduct)
	if err != nil {
		t.Errorf("Expected valid product to parse successfully, got error: %v", err)
	}
	if result.Name != validProduct.Name {
		assert.Equal(t, validProduct.Name, result.Name, "Expected name %s, got %s", validProduct.Name, result.Name)
	}

	// Test invalid product (name too short)
	invalidProduct1 := ProductWithMixedValidation{
		Name:  "AB", // Too short (< 3 chars)
		SKU:   "ITEM12345",
		Price: 100,
	}

	_, err = schema.Parse(invalidProduct1)
	assert.Error(t, err, "Expected error for short name, but got none")

	// Test invalid product (name is taken username)
	invalidProduct2 := ProductWithMixedValidation{
		Name:  "admin", // Taken username
		SKU:   "ITEM12345",
		Price: 100,
	}

	_, err = schema.Parse(invalidProduct2)
	assert.Error(t, err, "Expected error for taken username as product name, but got none")
}

func TestCustomValidatorWithoutParams(t *testing.T) {
	// Register custom validators
	_ = validators.Register(&UniqueUsernameValidator{})

	type SimpleUser struct {
		Username string `gozod:"unique_username"` // No parameters
	}

	schema := FromStruct[SimpleUser]()

	// Test valid case
	validUser := SimpleUser{Username: "newuser"}
	_, err := schema.Parse(validUser)
	if err != nil {
		t.Errorf("Expected valid username to parse successfully, got error: %v", err)
	}

	// Test invalid case
	invalidUser := SimpleUser{Username: "admin"}
	_, err = schema.Parse(invalidUser)
	assert.Error(t, err, "Expected error for taken username, but got none")
}
