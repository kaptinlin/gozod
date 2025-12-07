package validators

import (
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	switch param {
	case "even":
		return value%2 == 0
	case "odd":
		return value%2 != 0
	default:
		return v.Validate(value)
	}
}

func (v *TestParamValidator) ErrorMessageWithParam(ctx *core.ParseContext, param string) string {
	switch param {
	case "even":
		return "Value must be even"
	case "odd":
		return "Value must be odd"
	default:
		return v.ErrorMessage(ctx)
	}
}

func TestValidatorRegistration(t *testing.T) {
	// Clear any existing validators
	registry = &ValidatorRegistry{
		validators: make(map[string]any),
	}

	// Register test validator
	testValidator := &TestCustomValidator{}
	err := Register(testValidator)
	require.NoError(t, err)

	// Test retrieval
	validator, exists := Get[string]("test_validator")
	require.True(t, exists, "test_validator not found after registration")

	// Test validation
	assert.True(t, validator.Validate("valid"))
	assert.False(t, validator.Validate("invalid"))
	assert.False(t, validator.Validate(""))

	// Test error message
	ctx := &core.ParseContext{}
	msg := validator.ErrorMessage(ctx)
	assert.Equal(t, "Value failed test validation", msg)
}

func TestParameterizedValidator(t *testing.T) {
	// Clear any existing validators
	registry = &ValidatorRegistry{
		validators: make(map[string]any),
	}

	// Register parameterized validator
	paramValidator := &TestParamValidator{}
	err := Register(paramValidator)
	require.NoError(t, err)

	// Get validator
	validatorAny, exists := GetAny("test_param_validator")
	require.True(t, exists, "test_param_validator not found after registration")

	validator, ok := validatorAny.(ZodParameterizedValidator[int])
	require.True(t, ok, "test_param_validator should implement ZodParameterizedValidator")

	ctx := &core.ParseContext{}

	// Test basic validation
	assert.True(t, validator.Validate(5))
	assert.False(t, validator.Validate(-1))

	// Test parameterized validation
	assert.True(t, validator.ValidateParam(4, "even"))
	assert.False(t, validator.ValidateParam(3, "even"))
	assert.True(t, validator.ValidateParam(3, "odd"))
	assert.False(t, validator.ValidateParam(4, "odd"))

	// Test error messages
	assert.Equal(t, "Value must be even", validator.ErrorMessageWithParam(ctx, "even"))
	assert.Equal(t, "Value must be odd", validator.ErrorMessageWithParam(ctx, "odd"))
}

func TestDuplicateRegistration(t *testing.T) {
	// Clear any existing validators
	registry = &ValidatorRegistry{
		validators: make(map[string]any),
	}

	validator := &TestCustomValidator{}

	// First registration should succeed
	err := Register(validator)
	require.NoError(t, err)

	// Second registration should fail
	err = Register(validator)
	assert.Error(t, err)
}

func TestListValidators(t *testing.T) {
	// Clear any existing validators
	registry = &ValidatorRegistry{
		validators: make(map[string]any),
	}

	// Register multiple validators
	stringValidator := &TestCustomValidator{}
	intValidator := &TestParamValidator{}

	err := Register(stringValidator)
	require.NoError(t, err)

	err = Register(intValidator)
	require.NoError(t, err)

	// List validators
	names := ListValidators()
	assert.Len(t, names, 2)

	// Check that both validators are in the list
	assert.Contains(t, names, "test_validator")
	assert.Contains(t, names, "test_param_validator")
}

func TestConverterFunctions(t *testing.T) {
	validator := &TestCustomValidator{}
	ctx := &core.ParseContext{}

	// Test ToRefineFn
	refineFn := ToRefineFn(validator)
	assert.True(t, refineFn("valid"))
	assert.False(t, refineFn("invalid"))

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

func TestUnregister(t *testing.T) {
	// Clear any existing validators
	registry = &ValidatorRegistry{
		validators: make(map[string]any),
	}

	validator := &TestCustomValidator{}
	err := Register(validator)
	require.NoError(t, err)

	// Verify it exists
	_, exists := Get[string]("test_validator")
	assert.True(t, exists)

	// Unregister
	Unregister("test_validator")

	// Verify it no longer exists
	_, exists = Get[string]("test_validator")
	assert.False(t, exists)
}

func TestClear(t *testing.T) {
	// Clear any existing validators
	registry = &ValidatorRegistry{
		validators: make(map[string]any),
	}

	// Register multiple validators
	err := Register(&TestCustomValidator{})
	require.NoError(t, err)
	err = Register(&TestParamValidator{})
	require.NoError(t, err)

	// Verify they exist
	names := ListValidators()
	assert.Len(t, names, 2)

	// Clear all
	Clear()

	// Verify all gone
	names = ListValidators()
	assert.Empty(t, names)
}
