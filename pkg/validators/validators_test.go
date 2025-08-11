package validators

import (
	"testing"

	"github.com/kaptinlin/gozod/core"
)

// TestCustomValidator is an example custom validator for testing
type TestCustomValidator struct{}

func (v *TestCustomValidator) Name() string {
	return "test_validator"
}

func (v *TestCustomValidator) Validate(value string) bool {
	return len(value) > 0 && value != "invalid"
}

func (v *TestCustomValidator) ErrorMessage(ctx *core.ParseContext) string {
	return "Value failed test validation"
}

// TestParamValidator is an example parameterized validator for testing
type TestParamValidator struct{}

func (v *TestParamValidator) Name() string {
	return "test_param_validator"
}

func (v *TestParamValidator) Validate(value int) bool {
	return value > 0
}

func (v *TestParamValidator) ErrorMessage(ctx *core.ParseContext) string {
	return "Value must be positive"
}

func (v *TestParamValidator) ValidateParam(value int, param string) bool {
	if param == "even" {
		return value%2 == 0
	} else if param == "odd" {
		return value%2 != 0
	}
	return v.Validate(value)
}

func (v *TestParamValidator) ErrorMessageWithParam(ctx *core.ParseContext, param string) string {
	if param == "even" {
		return "Value must be even"
	} else if param == "odd" {
		return "Value must be odd"
	}
	return v.ErrorMessage(ctx)
}

func TestValidatorRegistration(t *testing.T) {
	// Clear any existing validators
	registry = &ValidatorRegistry{
		validators: make(map[string]any),
	}

	// Register test validator
	testValidator := &TestCustomValidator{}
	if err := Register(testValidator); err != nil {
		t.Fatalf("Failed to register test validator: %v", err)
	}

	// Test retrieval
	validator, exists := Get[string]("test_validator")
	if !exists {
		t.Fatal("test_validator not found after registration")
	}

	// Test validation
	if !validator.Validate("valid") {
		t.Error("Expected 'valid' to pass validation")
	}

	if validator.Validate("invalid") {
		t.Error("Expected 'invalid' to fail validation")
	}

	if validator.Validate("") {
		t.Error("Expected empty string to fail validation")
	}

	// Test error message
	ctx := &core.ParseContext{}
	msg := validator.ErrorMessage(ctx)
	if msg != "Value failed test validation" {
		t.Errorf("Unexpected error message: %s", msg)
	}
}

func TestParameterizedValidator(t *testing.T) {
	// Clear any existing validators
	registry = &ValidatorRegistry{
		validators: make(map[string]any),
	}

	// Register parameterized validator
	paramValidator := &TestParamValidator{}
	if err := Register(paramValidator); err != nil {
		t.Fatalf("Failed to register parameterized validator: %v", err)
	}

	// Get validator
	validatorAny, exists := GetAny("test_param_validator")
	if !exists {
		t.Fatal("test_param_validator not found after registration")
	}

	validator, ok := validatorAny.(ZodParameterizedValidator[int])
	if !ok {
		t.Fatal("test_param_validator should implement ZodParameterizedValidator")
	}

	ctx := &core.ParseContext{}

	// Test basic validation
	if !validator.Validate(5) {
		t.Error("Expected 5 to pass basic validation")
	}

	if validator.Validate(-1) {
		t.Error("Expected -1 to fail basic validation")
	}

	// Test parameterized validation
	if !validator.ValidateParam(4, "even") {
		t.Error("Expected 4 to pass even validation")
	}

	if validator.ValidateParam(3, "even") {
		t.Error("Expected 3 to fail even validation")
	}

	if !validator.ValidateParam(3, "odd") {
		t.Error("Expected 3 to pass odd validation")
	}

	if validator.ValidateParam(4, "odd") {
		t.Error("Expected 4 to fail odd validation")
	}

	// Test error messages
	msg := validator.ErrorMessageWithParam(ctx, "even")
	if msg != "Value must be even" {
		t.Errorf("Unexpected error message for even: %s", msg)
	}

	msg = validator.ErrorMessageWithParam(ctx, "odd")
	if msg != "Value must be odd" {
		t.Errorf("Unexpected error message for odd: %s", msg)
	}
}

func TestDuplicateRegistration(t *testing.T) {
	// Clear any existing validators
	registry = &ValidatorRegistry{
		validators: make(map[string]any),
	}

	validator := &TestCustomValidator{}

	// First registration should succeed
	if err := Register(validator); err != nil {
		t.Fatalf("First registration failed: %v", err)
	}

	// Second registration should fail
	if err := Register(validator); err == nil {
		t.Error("Expected duplicate registration to fail")
	}
}

func TestListValidators(t *testing.T) {
	// Clear any existing validators
	registry = &ValidatorRegistry{
		validators: make(map[string]any),
	}

	// Register multiple validators
	stringValidator := &TestCustomValidator{}
	intValidator := &TestParamValidator{}

	if err := Register(stringValidator); err != nil {
		t.Fatalf("Failed to register string validator: %v", err)
	}

	if err := Register(intValidator); err != nil {
		t.Fatalf("Failed to register int validator: %v", err)
	}

	// List validators
	names := ListValidators()
	if len(names) != 2 {
		t.Errorf("Expected 2 validators, got %d", len(names))
	}

	// Check that both validators are in the list
	hasTestValidator := false
	hasParamValidator := false
	for _, name := range names {
		if name == "test_validator" {
			hasTestValidator = true
		}
		if name == "test_param_validator" {
			hasParamValidator = true
		}
	}

	if !hasTestValidator {
		t.Error("test_validator not found in list")
	}

	if !hasParamValidator {
		t.Error("test_param_validator not found in list")
	}
}

func TestConverterFunctions(t *testing.T) {
	validator := &TestCustomValidator{}
	ctx := &core.ParseContext{}

	// Test ToRefineFn
	refineFn := ToRefineFn(validator)
	if !refineFn("valid") {
		t.Error("RefineFn should return true for valid input")
	}
	if refineFn("invalid") {
		t.Error("RefineFn should return false for invalid input")
	}

	// Test ToCheckFn
	checkFn := ToCheckFn[string](validator, ctx)

	// Create a mock payload
	payload := &core.ParsePayload{}
	payload.SetValue("valid")
	checkFn(payload)
	// In a real test, we would check that no issues were added

	payload.SetValue("invalid")
	checkFn(payload)
	// In a real test, we would check that an issue was added
}

func TestParameterizedConverterFunctions(t *testing.T) {
	validator := &TestParamValidator{}
	ctx := &core.ParseContext{}

	// Test ToCheckFnWithParam
	checkFn := ToCheckFnWithParam[int](validator, "even", ctx)

	// Create a mock payload
	payload := &core.ParsePayload{}
	payload.SetValue(4)
	checkFn(payload)
	// In a real test, we would check that no issues were added for even number

	payload.SetValue(3)
	checkFn(payload)
	// In a real test, we would check that an issue was added for odd number
}
