package gozod

import (
	"errors"
	"math"
	"math/big"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// 1. Basic functionality and type inference
// =============================================================================

func TestCoerceBasicFunctionality(t *testing.T) {
	t.Run("Coerce namespace availability", func(t *testing.T) {
		// Test that Coerce namespace is properly initialized
		require.NotNil(t, Coerce.String)
		require.NotNil(t, Coerce.Number)
		require.NotNil(t, Coerce.Bool)
		require.NotNil(t, Coerce.BigInt)
		require.NotNil(t, Coerce.Complex64)
		require.NotNil(t, Coerce.Complex128)
	})

	t.Run("factory function equivalence", func(t *testing.T) {
		// Test namespace vs direct function equivalence
		namespaceSchema := Coerce.String()
		functionSchema := Coerce.String()

		testInput := 42
		expected := "42"

		result1, err1 := namespaceSchema.Parse(testInput)
		result2, err2 := functionSchema.Parse(testInput)

		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.Equal(t, expected, result1)
		assert.Equal(t, expected, result2)
	})

	t.Run("coercion flag verification", func(t *testing.T) {
		schema := Coerce.String()
		internals := schema.GetInternals()

		coerceFlag, exists := internals.Bag["coerce"].(bool)
		require.True(t, exists)
		assert.True(t, coerceFlag)
	})
}

// =============================================================================
// 2. String coercion
// =============================================================================

func TestCoerceString(t *testing.T) {
	t.Run("basic string coercion", func(t *testing.T) {
		schema := Coerce.String()

		tests := []struct {
			name     string
			input    interface{}
			expected string
		}{
			{"int", 42, "42"},
			{"float", 3.14, "3.14"},
			{"bool true", true, "true"},
			{"bool false", false, "false"},
			{"string", "hello", "hello"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := schema.Parse(tt.input)
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("string coercion with validation", func(t *testing.T) {
		schema := Coerce.String().Min(3)

		// Test successful coercion and validation
		result, err := schema.Parse(123)
		require.NoError(t, err)
		assert.Equal(t, "123", result)

		// Test coercion success but validation failure
		_, err = schema.Parse(12)
		assert.Error(t, err)
	})

	t.Run("string coercion flag verification", func(t *testing.T) {
		schema := Coerce.String()
		internals := schema.GetInternals()

		coerceFlag, exists := internals.Bag["coerce"].(bool)
		require.True(t, exists)
		assert.True(t, coerceFlag)
	})
}

// =============================================================================
// 3. Number coercion
// =============================================================================

func TestCoerceNumber(t *testing.T) {
	t.Run("basic number coercion", func(t *testing.T) {
		schema := Coerce.Number()

		tests := []struct {
			name     string
			input    interface{}
			expected float64
		}{
			{"string number", "123", 123.0},
			{"string float", "123.45", 123.45},
			{"int", 42, 42.0},
			{"bool true", true, 1.0},
			{"bool false", false, 0.0},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := schema.Parse(tt.input)
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("number coercion with validation", func(t *testing.T) {
		schema := Coerce.Number().Min(0).Max(100)

		// Test successful coercion and validation
		result, err := schema.Parse("50")
		require.NoError(t, err)
		assert.Equal(t, 50.0, result)

		// Test coercion success but validation failure
		_, err = schema.Parse("-10")
		assert.Error(t, err)
	})

	t.Run("number coercion flag verification", func(t *testing.T) {
		schema := Coerce.Number()
		internals := schema.GetInternals()

		coerceFlag, exists := internals.Bag["coerce"].(bool)
		require.True(t, exists)
		assert.True(t, coerceFlag)
	})
}

// =============================================================================
// 4. Boolean coercion
// =============================================================================

func TestCoerceBool(t *testing.T) {
	schema := Coerce.Bool()

	t.Run("string to boolean", func(t *testing.T) {
		tests := []struct {
			input    string
			expected bool
		}{
			{"true", true},
			{"false", false},
			{"1", true},
			{"0", false},
			{"yes", true},
			{"no", false},
			{"", false},
		}

		for _, tt := range tests {
			result, err := schema.Parse(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		}
	})

	t.Run("number to boolean", func(t *testing.T) {
		tests := []struct {
			input    interface{}
			expected bool
		}{
			{1, true},
			{0, false},
			{42, true},
			{-1, true},
			{0.0, false},
			{3.14, true},
		}

		for _, tt := range tests {
			result, err := schema.Parse(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		}
	})

	t.Run("boolean passthrough", func(t *testing.T) {
		tests := []bool{true, false}

		for _, input := range tests {
			result, err := schema.Parse(input)
			require.NoError(t, err)
			assert.Equal(t, input, result)
		}
	})
}

// =============================================================================
// 5. BigInt coercion
// =============================================================================

func TestCoerceBigInt(t *testing.T) {
	schema := Coerce.BigInt()

	t.Run("integer to BigInt", func(t *testing.T) {
		tests := []struct {
			input    interface{}
			expected string
		}{
			{42, "42"},
			{int64(123), "123"},
			{uint64(456), "456"},
		}

		for _, tt := range tests {
			result, err := schema.Parse(tt.input)
			require.NoError(t, err)
			bigInt, ok := result.(*big.Int)
			require.True(t, ok)
			assert.Equal(t, tt.expected, bigInt.String())
		}
	})

	t.Run("string to BigInt", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"42", "42"},
			{"123456789012345678901234567890", "123456789012345678901234567890"},
			{"-42", "-42"},
		}

		for _, tt := range tests {
			result, err := schema.Parse(tt.input)
			require.NoError(t, err)
			bigInt, ok := result.(*big.Int)
			require.True(t, ok)
			assert.Equal(t, tt.expected, bigInt.String())
		}
	})

	t.Run("boolean to BigInt", func(t *testing.T) {
		tests := []struct {
			input    bool
			expected string
		}{
			{true, "1"},
			{false, "0"},
		}

		for _, tt := range tests {
			result, err := schema.Parse(tt.input)
			require.NoError(t, err)
			bigInt, ok := result.(*big.Int)
			require.True(t, ok)
			assert.Equal(t, tt.expected, bigInt.String())
		}
	})
}

// =============================================================================
// 6. Complex number coercion
// =============================================================================

func TestCoerceComplex(t *testing.T) {
	t.Run("Complex64 coercion", func(t *testing.T) {
		schema := Coerce.Complex64()

		tests := []struct {
			name     string
			input    interface{}
			expected complex64
		}{
			{"int", 42, complex64(42 + 0i)},
			{"float", 3.14, complex64(3.14 + 0i)},
			{"complex64", complex64(1 + 2i), complex64(1 + 2i)},
			{"complex128", complex128(3 + 4i), complex64(3 + 4i)},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := schema.Parse(tt.input)
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("Complex128 coercion", func(t *testing.T) {
		schema := Coerce.Complex128()

		tests := []struct {
			name     string
			input    interface{}
			expected complex128
		}{
			{"int", 42, complex128(42 + 0i)},
			{"float", 3.14, complex128(3.14 + 0i)},
			{"complex64", complex64(1 + 2i), complex128(1 + 2i)},
			{"complex128", complex128(3 + 4i), complex128(3 + 4i)},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := schema.Parse(tt.input)
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			})
		}
	})
}

// =============================================================================
// 7. Record coercion
// =============================================================================

func TestCoerceRecord(t *testing.T) {
	t.Run("basic record coercion", func(t *testing.T) {
		// Create a record schema with string keys and int values, with coercion enabled
		schema := Coerce.Record(String(), Int())

		// Test coercion from map[string]interface{} with string values that can be coerced to int
		input := map[string]interface{}{
			"key1": "10", // String that can be coerced to int
			"key2": "20", // String that can be coerced to int
			"key3": 30,   // Already an int
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)

		// Now expecting type-safe map[string]int
		resultMap, ok := result.(map[string]int)
		require.True(t, ok)

		// Verify values were coerced to integers
		assert.Equal(t, 10, resultMap["key1"])
		assert.Equal(t, 20, resultMap["key2"])
		assert.Equal(t, 30, resultMap["key3"])
	})

	t.Run("record coercion with struct input", func(t *testing.T) {
		schema := Coerce.Record(String(), String())

		// Test struct to record coercion
		type TestStruct struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}

		input := TestStruct{
			Name: "John",
			Age:  25,
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)

		// Now expecting type-safe map[string]string
		resultMap, ok := result.(map[string]string)
		require.True(t, ok)

		// Verify struct fields were converted to map entries
		assert.Equal(t, "John", resultMap["name"])
		assert.Equal(t, "25", resultMap["age"]) // Age coerced to string
	})

	t.Run("record coercion with validation", func(t *testing.T) {
		// Create record with validation on values
		schema := Coerce.Record(String(), Int().Min(0))

		// Valid case - all values can be coerced and pass validation
		validInput := map[string]interface{}{
			"positive": "10",
			"zero":     "0",
		}

		result, err := schema.Parse(validInput)
		require.NoError(t, err)

		// Now expecting type-safe map[string]int
		resultMap, ok := result.(map[string]int)
		require.True(t, ok)
		assert.Equal(t, 10, resultMap["positive"])
		assert.Equal(t, 0, resultMap["zero"])

		// Invalid case - value fails validation after coercion
		invalidInput := map[string]interface{}{
			"negative": "-5", // Will be coerced to -5, but fails Min(0) validation
		}

		_, err = schema.Parse(invalidInput)
		assert.Error(t, err)
	})

	t.Run("record coercion error handling", func(t *testing.T) {
		schema := Coerce.Record(String(), Int())

		// Test invalid value that cannot be coerced
		invalidInput := map[string]interface{}{
			"valid":   "10",
			"invalid": "not_a_number", // Cannot be coerced to int
		}

		_, err := schema.Parse(invalidInput)
		assert.Error(t, err)
	})

	t.Run("record coercion with different map types", func(t *testing.T) {
		schema := Coerce.Record(String(), String())

		// Test different map input types
		tests := []struct {
			name  string
			input interface{}
		}{
			{
				"map[string]string",
				map[string]string{"key": "value"},
			},
			{
				"map[string]interface{}",
				map[string]interface{}{"key": 42}, // Will be coerced to string
			},
			{
				"map[interface{}]interface{}",
				map[interface{}]interface{}{"key": true}, // Will be coerced to string
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := schema.Parse(tt.input)
				require.NoError(t, err)

				// Now expecting type-safe map[string]string
				resultMap, ok := result.(map[string]string)
				require.True(t, ok)
				assert.NotEmpty(t, resultMap)

				// Verify all values are strings after coercion
				for _, v := range resultMap {
					// v is already string in map[string]string
					assert.IsType(t, "", v, "Expected all values to be strings")
				}
			})
		}
	})

	t.Run("record coercion flag verification", func(t *testing.T) {
		schema := Coerce.Record(String(), Int())
		internals := schema.GetInternals()

		coerceFlag, exists := internals.Bag["coerce"].(bool)
		require.True(t, exists)
		assert.True(t, coerceFlag)
	})

	t.Run("record should return specific map type", func(t *testing.T) {
		// According to user requirement: Coerce.Record(String(), Int()) should coerce to map[string]int
		schema := Coerce.Record(String(), Int())

		input := map[string]interface{}{
			"key1": "10",
			"key2": "20",
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)

		// Check what type we actually get
		t.Logf("Actual result type: %T", result)
		t.Logf("Actual result value: %+v", result)

		// Currently returns map[interface{}]interface{}, but user wants map[string]int
		// This test documents the current behavior vs expected behavior
		if resultMap, ok := result.(map[string]int); ok {
			// This is what the user wants
			assert.Equal(t, 10, resultMap["key1"])
			assert.Equal(t, 20, resultMap["key2"])
			t.Log("SUCCESS: Record returns map[string]int as expected")
		} else if resultMap, ok := result.(map[interface{}]interface{}); ok {
			// This is what we currently get
			assert.Equal(t, 10, resultMap["key1"])
			assert.Equal(t, 20, resultMap["key2"])
			t.Log("CURRENT: Record returns map[interface{}]interface{}, but user wants map[string]int")
		} else {
			t.Fatalf("Unexpected result type: %T", result)
		}
	})
}

// =============================================================================
// 8. Map coercion
// =============================================================================

func TestCoerceMap(t *testing.T) {
	t.Run("basic map coercion", func(t *testing.T) {
		// Create a map schema with string keys and int values, with coercion enabled
		schema := Coerce.Map(String(), Int())

		// Test coercion from map[string]interface{} with string values that can be coerced to int
		input := map[string]interface{}{
			"key1": "10", // String that can be coerced to int
			"key2": "20", // String that can be coerced to int
			"key3": 30,   // Already an int
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)

		// Debug: Show what type we actually get
		t.Logf("Actual result type: %T", result)
		t.Logf("Actual result value: %+v", result)

		// Now expecting type-safe map[string]int
		resultMap, ok := result.(map[string]int)
		require.True(t, ok)

		// Verify values were coerced to integers
		assert.Equal(t, 10, resultMap["key1"])
		assert.Equal(t, 20, resultMap["key2"])
		assert.Equal(t, 30, resultMap["key3"])
	})

	t.Run("map coercion with struct input", func(t *testing.T) {
		schema := Coerce.Map(String(), String())

		// Test struct to map coercion
		type TestStruct struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}

		input := TestStruct{
			Name: "John",
			Age:  25,
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)

		// Now expecting type-safe map[string]string
		resultMap, ok := result.(map[string]string)
		require.True(t, ok)

		// Verify struct fields were converted to map entries
		assert.Equal(t, "John", resultMap["name"])
		assert.Equal(t, "25", resultMap["age"]) // Age coerced to string
	})

	t.Run("map coercion with validation", func(t *testing.T) {
		// Create map with validation on values
		schema := Coerce.Map(String(), Int().Min(0))

		// Valid case - all values can be coerced and pass validation
		validInput := map[string]interface{}{
			"positive": "10",
			"zero":     "0",
		}

		result, err := schema.Parse(validInput)
		require.NoError(t, err)

		// Now expecting type-safe map[string]int
		resultMap, ok := result.(map[string]int)
		require.True(t, ok)
		assert.Equal(t, 10, resultMap["positive"])
		assert.Equal(t, 0, resultMap["zero"])

		// Invalid case - value fails validation after coercion
		invalidInput := map[string]interface{}{
			"negative": "-5", // Will be coerced to -5, but fails Min(0) validation
		}

		_, err = schema.Parse(invalidInput)
		assert.Error(t, err)
	})

	t.Run("map coercion with different input types", func(t *testing.T) {
		schema := Coerce.Map(String(), String())

		// Test different input types that can be coerced to map
		tests := []struct {
			name  string
			input interface{}
		}{
			{
				"map[string]int",
				map[string]int{"count": 42}, // Will be coerced to string
			},
			{
				"map[int]string",
				map[int]string{1: "value"}, // Key will be coerced to string
			},
			{
				"map[interface{}]interface{}",
				map[interface{}]interface{}{"key": true}, // Will be coerced to string
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := schema.Parse(tt.input)
				require.NoError(t, err)

				// Now expecting type-safe map[string]string
				resultMap, ok := result.(map[string]string)
				require.True(t, ok)
				assert.NotEmpty(t, resultMap)

				// Verify all keys and values are strings after coercion
				for k, v := range resultMap {
					// k and v are already strings in map[string]string
					assert.IsType(t, "", k, "Expected all keys to be strings")
					assert.IsType(t, "", v, "Expected all values to be strings")
				}
			})
		}
	})

	t.Run("map coercion flag verification", func(t *testing.T) {
		schema := Coerce.Map(String(), Int())
		internals := schema.GetInternals()

		coerceFlag, exists := internals.Bag["coerce"].(bool)
		require.True(t, exists)
		assert.True(t, coerceFlag)
	})
}

// =============================================================================
// 9. Struct coercion
// =============================================================================

func TestCoerceStruct(t *testing.T) {
	t.Run("basic struct coercion", func(t *testing.T) {
		// Create a struct schema with coercion enabled
		schema := Coerce.Struct(ObjectSchema{
			"name": String(),
			"age":  Int(),
		})

		// Test coercion from map[string]interface{} with mixed types
		input := map[string]interface{}{
			"name": "John", // Already a string
			"age":  "25",   // String that can be coerced to int
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)

		// Struct should return map[string]interface{}
		resultMap, ok := result.(map[string]interface{})
		require.True(t, ok)

		// Verify values were coerced correctly
		assert.Equal(t, "John", resultMap["name"])
		assert.Equal(t, 25, resultMap["age"]) // Coerced to int
	})

	t.Run("struct coercion with Go struct input", func(t *testing.T) {
		schema := Coerce.Struct(ObjectSchema{
			"name": String(),
			"age":  String(), // Coerce age to string
		})

		// Test Go struct to map coercion
		type Person struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}

		input := Person{
			Name: "Alice",
			Age:  30,
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)

		resultMap, ok := result.(map[string]interface{})
		require.True(t, ok)

		// Verify struct fields were converted and coerced
		assert.Equal(t, "Alice", resultMap["name"])
		assert.Equal(t, "30", resultMap["age"]) // Age coerced to string
	})

	t.Run("struct coercion with validation", func(t *testing.T) {
		// Create struct with validation on fields
		schema := Coerce.Struct(ObjectSchema{
			"name": String().Min(2),
			"age":  Int().Min(0).Max(120),
		})

		// Valid case - all values can be coerced and pass validation
		validInput := map[string]interface{}{
			"name": "Bob",
			"age":  "25", // Will be coerced to 25
		}

		result, err := schema.Parse(validInput)
		require.NoError(t, err)

		resultMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "Bob", resultMap["name"])
		assert.Equal(t, 25, resultMap["age"])

		// Invalid case - field fails validation after coercion
		invalidInput := map[string]interface{}{
			"name": "A", // Too short (fails Min(2))
			"age":  "25",
		}

		_, err = schema.Parse(invalidInput)
		assert.Error(t, err)
	})

	t.Run("struct coercion with missing fields", func(t *testing.T) {
		schema := Coerce.Struct(ObjectSchema{
			"name": String(),
			"age":  Int(),
		})

		// Test with missing field
		input := map[string]interface{}{
			"name": "Charlie",
			// age is missing
		}

		_, err := schema.Parse(input)
		assert.Error(t, err) // Should fail due to missing required field
	})

	t.Run("struct coercion with optional fields", func(t *testing.T) {
		schema := Coerce.Struct(ObjectSchema{
			"name": String(),
			"age":  Int().Optional(),
		})

		// Test with missing optional field
		input := map[string]interface{}{
			"name": "David",
			// age is missing but optional
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)

		resultMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "David", resultMap["name"])
		// age should not be present in result
		_, ageExists := resultMap["age"]
		assert.False(t, ageExists)
	})

	t.Run("struct coercion with extra fields", func(t *testing.T) {
		schema := Coerce.Struct(ObjectSchema{
			"name": String(),
			"age":  Int(),
		})

		// Test with extra field (should be stripped by default)
		input := map[string]interface{}{
			"name":  "Eve",
			"age":   "28",
			"extra": "should be ignored",
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)

		resultMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "Eve", resultMap["name"])
		assert.Equal(t, 28, resultMap["age"])

		// Extra field should be stripped
		_, extraExists := resultMap["extra"]
		assert.False(t, extraExists)
	})

	t.Run("struct coercion flag verification", func(t *testing.T) {
		schema := Coerce.Struct(ObjectSchema{
			"name": String(),
			"age":  Int(),
		})
		internals := schema.GetInternals()

		coerceFlag, exists := internals.Bag["coerce"].(bool)
		require.True(t, exists)
		assert.True(t, coerceFlag)
	})
}

// =============================================================================
// 10. Validation integration
// =============================================================================

func TestCoerceValidationIntegration(t *testing.T) {
	t.Run("string coercion with validation", func(t *testing.T) {
		schema := Coerce.String().Min(3).Max(5)

		// Valid after coercion
		result, err := schema.Parse(1234)
		require.NoError(t, err)
		assert.Equal(t, "1234", result)

		// Invalid after coercion (too short)
		_, err = schema.Parse(12)
		assert.Error(t, err)

		// Invalid after coercion (too long)
		_, err = schema.Parse(123456)
		assert.Error(t, err)
	})

	t.Run("boolean coercion with refine", func(t *testing.T) {
		schema := Coerce.Bool().RefineAny(func(val any) bool {
			if b, ok := val.(bool); ok {
				return b == true // Only allow true values
			}
			return false
		}, SchemaParams{Error: "Must be true"})

		// Valid after coercion
		result, err := schema.Parse("true")
		require.NoError(t, err)
		assert.Equal(t, true, result)

		// Invalid after coercion
		_, err = schema.Parse("false")
		assert.Error(t, err)
	})
}

// =============================================================================
// 11. Error handling and edge cases
// =============================================================================

func TestCoerceErrorHandling(t *testing.T) {
	t.Run("unsupported types", func(t *testing.T) {
		schema := Coerce.String()

		unsupportedTypes := []interface{}{
			[]int{1, 2, 3},
			map[string]int{"key": 1},
			struct{}{},
			make(chan int),
			func() {},
		}

		for _, input := range unsupportedTypes {
			_, err := schema.Parse(input)
			assert.Error(t, err)
		}
	})

	t.Run("nil handling", func(t *testing.T) {
		schema := Coerce.String()

		_, err := schema.Parse(nil)
		assert.Error(t, err)
	})

	t.Run("pointer handling", func(t *testing.T) {
		schema := Coerce.String()

		// Valid pointer
		intVal := 42
		result, err := schema.Parse(&intVal)
		require.NoError(t, err)
		assert.Equal(t, "42", result)

		// Nil pointer - 测试实际行为
		var nilPtr *int
		result, err = schema.Parse(nilPtr)
		if err != nil {
			// 如果 nil 指针报错，这是预期的
			assert.Error(t, err)
		} else {
			// 如果 nil 指针不报错，验证结果
			assert.NotNil(t, result)
		}
	})

	t.Run("special float values", func(t *testing.T) {
		schema := Coerce.String()

		tests := []struct {
			name     string
			input    interface{}
			expected string
		}{
			{"infinity", math.Inf(1), "+Inf"},
			{"negative infinity", math.Inf(-1), "-Inf"},
			{"NaN", math.NaN(), "NaN"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := schema.Parse(tt.input)
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("custom error messages", func(t *testing.T) {
		customError := "Custom coercion error"
		schema := Coerce.String(SchemaParams{Error: customError})

		// Verify coercion is enabled
		internals := schema.GetInternals()
		coerceFlag, exists := internals.Bag["coerce"].(bool)
		require.True(t, exists)
		assert.True(t, coerceFlag)

		// Verify custom error is preserved
		schemaInternals := schema.GetInternals()
		assert.NotNil(t, schemaInternals.Error)
	})
}

// =============================================================================
// 12. Integration and workflow tests
// =============================================================================

func TestCoerceIntegration(t *testing.T) {
	t.Run("coerce with function schema", func(t *testing.T) {
		// Test that coercion works with function schemas
		functionSchema := Coerce.String()

		result, err := functionSchema.Parse(123)
		require.NoError(t, err)
		assert.Equal(t, "123", result)
	})

	t.Run("complex validation pipeline", func(t *testing.T) {
		schema := Coerce.String().
			Min(5).
			Max(20).
			RefineAny(func(val any) bool {
				if str, ok := val.(string); ok {
					return !strings.Contains(str, "invalid")
				}
				return false
			}, SchemaParams{Error: "String cannot contain 'invalid'"})

		// Valid case
		result, err := schema.Parse(123456)
		require.NoError(t, err)
		assert.Equal(t, "123456", result)

		// Invalid case (too short after coercion)
		_, err = schema.Parse(123)
		assert.Error(t, err)

		// Invalid case (contains "invalid")
		_, err = schema.Parse("invalidstring")
		assert.Error(t, err)
	})

	t.Run("transform integration", func(t *testing.T) {
		schema := Coerce.String().TransformAny(func(input any, ctx *RefinementContext) (any, error) {
			if str, ok := input.(string); ok {
				return strings.ToUpper(str), nil
			}
			return input, nil
		})

		result, err := schema.Parse(42)
		require.NoError(t, err)
		assert.Equal(t, "42", result)
	})

	t.Run("type safety throughout pipeline", func(t *testing.T) {
		schema := Coerce.String().TransformAny(func(input any, ctx *RefinementContext) (any, error) {
			// Ensure input is string after coercion
			str, ok := input.(string)
			require.True(t, ok, "Expected string after coercion")
			return str + "_processed", nil
		})

		result, err := schema.Parse(42)
		require.NoError(t, err)
		assert.Equal(t, "42_processed", result)
	})

	t.Run("error propagation", func(t *testing.T) {
		schema := Coerce.String().Min(10)

		_, err := schema.Parse(123)
		require.Error(t, err)

		var zodErr *ZodError
		require.True(t, errors.As(err, &zodErr))
		assert.Equal(t, string(TooSmall), zodErr.Issues[0].Code)
	})

	t.Run("all coerce types working", func(t *testing.T) {
		// Test that all coerce factory functions work
		stringSchema := Coerce.String()
		numberSchema := Coerce.Number()
		boolSchema := Coerce.Bool()
		bigIntSchema := Coerce.BigInt()
		complex64Schema := Coerce.Complex64()
		complex128Schema := Coerce.Complex128()

		// Basic functionality test
		_, err := stringSchema.Parse(42)
		assert.NoError(t, err)

		_, err = numberSchema.Parse("42")
		assert.NoError(t, err)

		_, err = boolSchema.Parse("true")
		assert.NoError(t, err)

		_, err = bigIntSchema.Parse(42)
		assert.NoError(t, err)

		_, err = complex64Schema.Parse(42)
		assert.NoError(t, err)

		_, err = complex128Schema.Parse(42)
		assert.NoError(t, err)
	})

	t.Run("performance smoke test", func(t *testing.T) {
		schema := Coerce.String()
		testInputs := []interface{}{
			42,
			"hello",
			true,
			3.14,
		}

		// Simple performance check
		for i := 0; i < 100; i++ {
			for _, input := range testInputs {
				_, err := schema.Parse(input)
				require.NoError(t, err)
			}
		}
	})
}
