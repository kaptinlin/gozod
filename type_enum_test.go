package gozod

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// 1. Basic functionality and type inference
// =============================================================================

func TestEnumBasicFunctionality(t *testing.T) {
	t.Run("enum from values", func(t *testing.T) {
		// new API: more concise and type-safe
		myEnum := Enum("Red", "Green", "Blue")

		// Test enum property access
		options := myEnum.Options()
		assert.Contains(t, options, "Red")
		assert.Contains(t, options, "Green")
		assert.Contains(t, options, "Blue")

		// Test parsing
		result, err := myEnum.Parse("Red")
		require.NoError(t, err)
		assert.Equal(t, "Red", result)
	})

	t.Run("enum from map", func(t *testing.T) {
		fruits := map[string]string{
			"Apple":  "apple",
			"Banana": "banana",
		}
		fruitEnum := EnumMap(fruits)

		// Test parsing values
		result, err := fruitEnum.Parse("apple")
		require.NoError(t, err)
		assert.Equal(t, "apple", result)

		result, err = fruitEnum.Parse("banana")
		require.NoError(t, err)
		assert.Equal(t, "banana", result)

		// Test enum property access
		enumObj := fruitEnum.Enum()
		assert.Equal(t, "apple", enumObj["Apple"])
		assert.Equal(t, "banana", enumObj["Banana"])
	})

	t.Run("enum from slice", func(t *testing.T) {
		foods := []string{"Pasta", "Pizza", "Tacos", "Burgers", "Salad"}
		foodEnum := EnumSlice(foods)

		result, err := foodEnum.Parse("Pasta")
		require.NoError(t, err)
		assert.Equal(t, "Pasta", result)

		_, err = foodEnum.Parse("Cucumbers")
		assert.Error(t, err)
	})

	t.Run("enum from native enum with numeric keys", func(t *testing.T) {
		fruitValues := map[string]int{
			"Apple":  10,
			"Banana": 20,
		}
		fruitEnum := EnumMap(fruitValues)

		// Test parsing numeric values
		result, err := fruitEnum.Parse(10)
		require.NoError(t, err)
		assert.Equal(t, 10, result)

		result, err = fruitEnum.Parse(20)
		require.NoError(t, err)
		assert.Equal(t, 20, result)
	})

	t.Run("go native iota constants", func(t *testing.T) {
		// Go native iota constants
		type Color int
		const (
			Red Color = iota
			Green
			Blue
		)

		colorEnum := Enum(Red, Green, Blue)

		// Test parsing iota values
		result, err := colorEnum.Parse(Red)
		require.NoError(t, err)
		assert.Equal(t, Red, result)

		result, err = colorEnum.Parse(Green)
		require.NoError(t, err)
		assert.Equal(t, Green, result)

		result, err = colorEnum.Parse(Blue)
		require.NoError(t, err)
		assert.Equal(t, Blue, result)

		// Test invalid value
		_, err = colorEnum.Parse(Color(99))
		assert.Error(t, err)
	})

	t.Run("go native string constants", func(t *testing.T) {
		// Go native string constants
		type Status string
		const (
			StatusActive   Status = "active"
			StatusInactive Status = "inactive"
			StatusPending  Status = "pending"
		)

		statusEnum := Enum(StatusActive, StatusInactive, StatusPending)

		// Test parsing string constants
		result, err := statusEnum.Parse(StatusActive)
		require.NoError(t, err)
		assert.Equal(t, StatusActive, result)

		result, err = statusEnum.Parse(StatusInactive)
		require.NoError(t, err)
		assert.Equal(t, StatusInactive, result)

		// Test invalid value
		_, err = statusEnum.Parse(Status("unknown"))
		assert.Error(t, err)
	})

	t.Run("integer enum", func(t *testing.T) {
		// integer enum
		intEnum := Enum(1, 2, 3, 42)

		// Test parsing different values
		result, err := intEnum.Parse(1)
		require.NoError(t, err)
		assert.Equal(t, 1, result)

		result, err = intEnum.Parse(42)
		require.NoError(t, err)
		assert.Equal(t, 42, result)

		// Test invalid value
		_, err = intEnum.Parse(99)
		assert.Error(t, err)
	})

	t.Run("boolean enum", func(t *testing.T) {
		// boolean enum
		boolEnum := Enum(true, false)

		// Test parsing boolean values
		result, err := boolEnum.Parse(true)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		result, err = boolEnum.Parse(false)
		require.NoError(t, err)
		assert.Equal(t, false, result)
	})

	t.Run("smart type inference", func(t *testing.T) {
		schema := EnumMap(map[string]string{
			"ACTIVE":   "active",
			"INACTIVE": "inactive",
		})

		// String input returns string
		result1, err := schema.Parse("active")
		require.NoError(t, err)
		assert.IsType(t, "", result1)
		assert.Equal(t, "active", result1)

		// Pointer input returns same pointer (smart type inference)
		str := "inactive"
		result2, err := schema.Parse(&str)
		require.NoError(t, err)
		assert.IsType(t, (*string)(nil), result2)
		assert.Equal(t, &str, result2)

		// Test with integer enum and pointer
		intSchema := EnumMap(map[string]int{
			"ONE": 1,
			"TWO": 2,
		})

		// Integer input returns integer
		result3, err := intSchema.Parse(1)
		require.NoError(t, err)
		assert.IsType(t, 0, result3)
		assert.Equal(t, 1, result3)

		// Integer pointer input returns same pointer
		intVal := 2
		result4, err := intSchema.Parse(&intVal)
		require.NoError(t, err)
		assert.IsType(t, (*int)(nil), result4)
		assert.Equal(t, &intVal, result4)
	})

	t.Run("nilable modifier", func(t *testing.T) {
		schema := EnumMap(map[string]string{
			"SMALL":  "small",
			"MEDIUM": "medium",
			"LARGE":  "large",
		}).Nilable()

		// nil input should succeed, return nil
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Valid input keeps type inference
		result2, err := schema.Parse("small")
		require.NoError(t, err)
		assert.Equal(t, "small", result2)
		assert.IsType(t, "", result2)
	})

	t.Run("get options", func(t *testing.T) {
		schema := Enum("tuna", "trout")
		options := schema.Options()
		assert.Len(t, options, 2)
		assert.Contains(t, options, "tuna")
		assert.Contains(t, options, "trout")
	})
}

// =============================================================================
// 2. Coerce (type coercion)
// =============================================================================

func TestEnumCoercion(t *testing.T) {
	t.Run("basic coercion behavior", func(t *testing.T) {
		// use new API, but need to support coercion parameter
		schema := Enum(1, 2, 3)

		// Direct matches should work
		result, err := schema.Parse(1)
		require.NoError(t, err)
		assert.Equal(t, 1, result)

		result, err = schema.Parse(2)
		require.NoError(t, err)
		assert.Equal(t, 2, result)
	})

	t.Run("without coercion", func(t *testing.T) {
		schema := Enum(42) // No coercion

		result, err := schema.Parse(42)
		require.NoError(t, err)
		assert.Equal(t, 42, result)

		_, err = schema.Parse("42")
		assert.Error(t, err)
	})
}

// =============================================================================
// 3. Validation methods
// =============================================================================

func TestEnumValidations(t *testing.T) {
	t.Run("extract method", func(t *testing.T) {
		foods := map[string]string{
			"PASTA":   "Pasta",
			"PIZZA":   "Pizza",
			"TACOS":   "Tacos",
			"BURGERS": "Burgers",
			"SALAD":   "Salad",
		}
		foodEnum := EnumMap(foods)
		italianEnum := foodEnum.Extract([]string{"PASTA", "PIZZA"})

		// Should parse extracted values
		result, err := italianEnum.Parse("Pasta")
		require.NoError(t, err)
		assert.Equal(t, "Pasta", result)

		// Should reject non-extracted values
		_, err = italianEnum.Parse("Tacos")
		assert.Error(t, err)
	})

	t.Run("exclude method", func(t *testing.T) {
		foods := map[string]string{
			"PASTA":   "Pasta",
			"PIZZA":   "Pizza",
			"TACOS":   "Tacos",
			"BURGERS": "Burgers",
			"SALAD":   "Salad",
		}
		foodEnum := EnumMap(foods)
		unhealthyEnum := foodEnum.Exclude([]string{"SALAD"})

		// Should parse non-excluded values
		result, err := unhealthyEnum.Parse("Pasta")
		require.NoError(t, err)
		assert.Equal(t, "Pasta", result)

		// Should reject excluded values
		_, err = unhealthyEnum.Parse("Salad")
		assert.Error(t, err)

		// Test empty enum after excluding all
		emptyEnum := foodEnum.Exclude([]string{"PASTA", "PIZZA", "TACOS", "BURGERS", "SALAD"})
		_, err = emptyEnum.Parse("Pasta")
		assert.Error(t, err)
		assert.Len(t, emptyEnum.Enum(), 0)
		assert.Len(t, emptyEnum.Options(), 0)
	})

	t.Run("enum property access", func(t *testing.T) {
		schema := EnumMap(map[string]string{
			"RED":   "red",
			"GREEN": "green",
			"BLUE":  "blue",
		})

		enumObj := schema.Enum()
		assert.Equal(t, "red", enumObj["RED"])
		assert.Equal(t, "green", enumObj["GREEN"])
		assert.Equal(t, "blue", enumObj["BLUE"])
	})

	t.Run("options property access", func(t *testing.T) {
		schema := EnumMap(map[string]string{
			"TUNA":  "tuna",
			"TROUT": "trout",
		})

		options := schema.Options()
		assert.Len(t, options, 2)
		assert.Contains(t, options, "tuna")
		assert.Contains(t, options, "trout")
	})
}

// =============================================================================
// 4. Modifiers and wrappers
// =============================================================================

func TestEnumModifiers(t *testing.T) {
	schema := EnumMap(map[string]string{
		"RED":   "red",
		"GREEN": "green",
		"BLUE":  "blue",
	})

	t.Run("optional wrapper", func(t *testing.T) {
		optionalSchema := schema.Optional()

		result, err := optionalSchema.Parse("red")
		require.NoError(t, err)
		assert.Equal(t, "red", result)

		result, err = optionalSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("nilable wrapper", func(t *testing.T) {
		nilableSchema := schema.Nilable()

		result, err := nilableSchema.Parse("green")
		require.NoError(t, err)
		assert.Equal(t, "green", result)

		result, err = nilableSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("nullish wrapper", func(t *testing.T) {
		nullishSchema := schema.Nullish()

		result, err := nullishSchema.Parse("blue")
		require.NoError(t, err)
		assert.Equal(t, "blue", result)

		result, err = nullishSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("MustParse method", func(t *testing.T) {
		// Should not panic for valid values
		result := schema.MustParse("red")
		assert.Equal(t, "red", result)

		// Should panic for invalid values
		assert.Panics(t, func() {
			schema.MustParse("yellow")
		})
	})
}

// =============================================================================
// 5. Chaining and method composition
// =============================================================================

func TestEnumChaining(t *testing.T) {
	t.Run("extract and exclude chaining", func(t *testing.T) {
		foods := map[string]string{
			"PASTA":   "Pasta",
			"PIZZA":   "Pizza",
			"TACOS":   "Tacos",
			"BURGERS": "Burgers",
			"SALAD":   "Salad",
		}
		schema := EnumMap(foods)

		// Chain extract with refine
		italian := schema.Extract([]string{"PASTA", "PIZZA"}).Refine(func(val string) bool {
			// Only allow Pizza
			return val == "Pizza"
		})

		result, err := italian.Parse("Pizza")
		require.NoError(t, err)
		assert.Equal(t, "Pizza", result)

		_, err = italian.Parse("Pasta")
		assert.Error(t, err)
	})

	t.Run("enum property inheritance", func(t *testing.T) {
		foods := map[string]string{
			"PASTA": "Pasta",
			"PIZZA": "Pizza",
			"SALAD": "Salad",
		}
		schema := EnumMap(foods)
		extracted := schema.Extract([]string{"PASTA", "PIZZA"})

		// Extracted enum should have correct properties
		enumObj := extracted.Enum()
		assert.Len(t, enumObj, 2)
		assert.Equal(t, "Pasta", enumObj["PASTA"])
		assert.Equal(t, "Pizza", enumObj["PIZZA"])
		assert.NotContains(t, enumObj, "SALAD")

		options := extracted.Options()
		assert.Len(t, options, 2)
		assert.Contains(t, options, "Pasta")
		assert.Contains(t, options, "Pizza")
		assert.NotContains(t, options, "Salad")
	})

	t.Run("error map inheritance", func(t *testing.T) {
		foods := map[string]string{
			"PASTA":   "Pasta",
			"PIZZA":   "Pizza",
			"TACOS":   "Tacos",
			"BURGERS": "Burgers",
			"SALAD":   "Salad",
		}
		foodEnum := EnumMap(foods)
		italianEnum := foodEnum.Extract([]string{"PASTA", "PIZZA"})

		// Both should have similar error messages
		_, foodsErr := foodEnum.Parse("Cucumbers")
		_, italianErr := italianEnum.Parse("Tacos")

		assert.Error(t, foodsErr)
		assert.Error(t, italianErr)

		// Test exclude with custom error
		unhealthyEnum := foodEnum.Exclude([]string{"SALAD"})
		_, unhealthyErr := unhealthyEnum.Parse("Salad")
		assert.Error(t, unhealthyErr)
	})
}

// =============================================================================
// 6. Transform/Pipe
// =============================================================================

func TestEnumTransform(t *testing.T) {
	schema := EnumMap(map[string]string{
		"RED":   "red",
		"GREEN": "green",
		"BLUE":  "blue",
	})

	t.Run("transform values", func(t *testing.T) {
		transformSchema := schema.TransformAny(func(val any, ctx *RefinementContext) (any, error) {
			if str, ok := val.(string); ok {
				return "#" + str, nil
			}
			return val, nil
		})

		result, err := transformSchema.Parse("red")
		require.NoError(t, err)
		assert.Equal(t, "#red", result)
	})

	t.Run("pipe to string", func(t *testing.T) {
		pipeSchema := schema.Pipe(String().Min(3))

		result, err := pipeSchema.Parse("red")
		require.NoError(t, err)
		assert.Equal(t, "red", result)

		// Should fail if piped value doesn't meet requirements
		pipeSchema2 := schema.Pipe(String().Min(10))
		_, err = pipeSchema2.Parse("red")
		assert.Error(t, err)
	})

	t.Run("transform with extract", func(t *testing.T) {
		extracted := schema.Extract([]string{"RED", "BLUE"})
		transformed := extracted.TransformAny(func(val any, ctx *RefinementContext) (any, error) {
			if str, ok := val.(string); ok {
				return map[string]string{"color": str}, nil
			}
			return val, nil
		})

		result, err := transformed.Parse("red")
		require.NoError(t, err)
		resultMap, ok := result.(map[string]string)
		require.True(t, ok)
		assert.Equal(t, "red", resultMap["color"])
	})
}

// =============================================================================
// 7. Refine
// =============================================================================

func TestEnumRefine(t *testing.T) {
	t.Run("basic refine", func(t *testing.T) {
		schema := EnumMap(map[string]int{
			"LOW":    1,
			"MEDIUM": 2,
			"HIGH":   3,
		}).Refine(func(val int) bool {
			// Don't allow LOW priority
			return val != 1
		})

		// Valid cases
		result, err := schema.Parse(2)
		require.NoError(t, err)
		assert.Equal(t, 2, result)

		result, err = schema.Parse(3)
		require.NoError(t, err)
		assert.Equal(t, 3, result)

		// Invalid case
		_, err = schema.Parse(1)
		assert.Error(t, err)
	})

	t.Run("refine with custom error", func(t *testing.T) {
		schema := EnumMap(map[string]string{
			"ADMIN": "admin",
			"USER":  "user",
		}).Refine(func(val string) bool {
			return val == "admin"
		}, SchemaParams{
			Error: "Only admin role allowed",
		})

		result, err := schema.Parse("admin")
		require.NoError(t, err)
		assert.Equal(t, "admin", result)

		_, err = schema.Parse("user")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Only admin role allowed")
	})
}

// =============================================================================
// 8. Error handling
// =============================================================================

func TestEnumErrorHandling(t *testing.T) {
	t.Run("issue metadata", func(t *testing.T) {
		schema := EnumSlice([]string{"Red", "Green", "Blue"})
		_, err := schema.Parse("Yellow")

		assert.Error(t, err)
		var zodErr *ZodError
		require.True(t, IsZodError(err, &zodErr))
		assert.Len(t, zodErr.Issues, 1)
		assert.Equal(t, "invalid_value", zodErr.Issues[0].Code)
	})

	t.Run("enum error message, invalid enum element", func(t *testing.T) {
		schema := EnumSlice([]string{"Tuna", "Trout"})
		_, err := schema.Parse("Salmon")

		assert.Error(t, err)
		var zodErr *ZodError
		require.True(t, IsZodError(err, &zodErr))
		assert.Len(t, zodErr.Issues, 1)
		assert.Equal(t, "invalid_value", zodErr.Issues[0].Code)
	})

	t.Run("enum error message, invalid type", func(t *testing.T) {
		schema := EnumSlice([]string{"Tuna", "Trout"})
		_, err := schema.Parse(12)

		assert.Error(t, err)
		var zodErr *ZodError
		require.True(t, IsZodError(err, &zodErr))
		assert.Len(t, zodErr.Issues, 1)
		assert.Equal(t, "invalid_value", zodErr.Issues[0].Code)
	})

	t.Run("enum with message returns the custom error message", func(t *testing.T) {
		schema := Enum("apple", "banana")

		// Test with invalid string
		_, err := schema.Parse("berries")
		assert.Error(t, err)
		var zodErr *ZodError
		require.True(t, IsZodError(err, &zodErr))
		assert.NotEmpty(t, zodErr.Issues[0].Message)

		// Test with undefined/nil
		_, err = schema.Parse(nil)
		assert.Error(t, err)
		require.True(t, IsZodError(err, &zodErr))
		assert.NotEmpty(t, zodErr.Issues[0].Message)

		// Test with valid value
		result, err := schema.Parse("banana")
		require.NoError(t, err)
		assert.Equal(t, "banana", result)

		// Test with null
		_, err = schema.Parse(nil)
		assert.Error(t, err)
		require.True(t, IsZodError(err, &zodErr))
		assert.NotEmpty(t, zodErr.Issues[0].Message)
	})

	t.Run("error structure", func(t *testing.T) {
		schema := EnumMap(map[string]string{
			"TUNA":  "Tuna",
			"TROUT": "Trout",
		})

		_, err := schema.Parse("Salmon")
		assert.Error(t, err)

		var zodErr *ZodError
		require.True(t, IsZodError(err, &zodErr))
		assert.Len(t, zodErr.Issues, 1)
		assert.Equal(t, "invalid_value", zodErr.Issues[0].Code)
	})
}

// =============================================================================
// 9. Edge and mutual exclusion cases
// =============================================================================

func TestEnumEdgeCases(t *testing.T) {
	t.Run("empty enum", func(t *testing.T) {
		schema := EnumMap(map[string]string{})

		require.NotNil(t, schema)
		enumInternals := schema.GetZod()
		assert.Equal(t, 0, len(enumInternals.Values))

		// Any value should fail
		_, err := schema.Parse("anything")
		assert.Error(t, err)
	})

	t.Run("case sensitivity", func(t *testing.T) {
		schema := EnumMap(map[string]string{"UPPER": "lower"})

		result, err := schema.Parse("lower")
		require.NoError(t, err)
		assert.Equal(t, "lower", result)

		_, err = schema.Parse("UPPER") // Key, not value
		assert.Error(t, err)

		_, err = schema.Parse("Lower") // Different case
		assert.Error(t, err)
	})

	t.Run("numeric enum filtering", func(t *testing.T) {
		schema := EnumMap(map[string]int{
			"0":        0,
			"1":        1,
			"2":        2,
			"Active":   0,
			"Inactive": 1,
			"Pending":  2,
		})

		// Should accept numeric values
		for _, value := range []int{0, 1, 2} {
			result, err := schema.Parse(value)
			require.NoError(t, err)
			assert.Equal(t, value, result)
		}
	})

	t.Run("enum with diagonal keys", func(t *testing.T) {
		schema := EnumMap(map[string]any{
			"A": 1,
			"B": "A",
		})

		result, err := schema.Parse("A")
		require.NoError(t, err)
		assert.Equal(t, "A", result)

		result, err = schema.Parse(1)
		require.NoError(t, err)
		assert.Equal(t, 1, result)
	})

	t.Run("extract with invalid key", func(t *testing.T) {
		schema := EnumMap(map[string]string{
			"RED":   "red",
			"GREEN": "green",
		})

		// Should panic when extracting non-existent key
		assert.Panics(t, func() {
			schema.Extract([]string{"PURPLE"})
		})
	})

	t.Run("exclude with invalid key", func(t *testing.T) {
		schema := EnumMap(map[string]string{
			"RED":   "red",
			"GREEN": "green",
		})

		// Should panic when excluding non-existent key
		assert.Panics(t, func() {
			schema.Exclude([]string{"PURPLE"})
		})
	})

	t.Run("exclude all values", func(t *testing.T) {
		foods := map[string]string{
			"PASTA": "Pasta",
			"PIZZA": "Pizza",
		}
		schema := EnumMap(foods)
		empty := schema.Exclude([]string{"PASTA", "PIZZA"})

		// Empty enum should reject all values
		_, err := empty.Parse("Pasta")
		assert.Error(t, err)

		_, err = empty.Parse("Pizza")
		assert.Error(t, err)

		// Should have empty enum and options
		assert.Len(t, empty.Enum(), 0)
		assert.Len(t, empty.Options(), 0)
	})
}

// =============================================================================
// 10. Default and Prefault tests
// =============================================================================

func TestEnumDefaultAndPrefault(t *testing.T) {
	schema := EnumMap(map[string]string{
		"RED":   "red",
		"GREEN": "green",
		"BLUE":  "blue",
	})

	t.Run("default values", func(t *testing.T) {
		defaultSchema := schema.Default("red")

		result, err := defaultSchema.Parse("blue")
		require.NoError(t, err)
		assert.Equal(t, "blue", result)

		result, err = defaultSchema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "red", result)
	})

	t.Run("default function", func(t *testing.T) {
		counter := 0
		defaultSchema := schema.DefaultFunc(func() string {
			counter++
			return "green"
		})

		// Each nil input should call the function
		result, err := defaultSchema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "green", result)
		assert.Equal(t, 1, counter)

		// Valid input should not call function
		result, err = defaultSchema.Parse("blue")
		require.NoError(t, err)
		assert.Equal(t, "blue", result)
		assert.Equal(t, 1, counter) // Counter unchanged
	})

	t.Run("prefault values", func(t *testing.T) {
		prefaultSchema := schema.Prefault("red")

		// Valid value
		result, err := prefaultSchema.Parse("green")
		require.NoError(t, err)
		assert.Equal(t, "green", result)

		// Invalid value should use prefault
		result, err = prefaultSchema.Parse("invalid")
		require.NoError(t, err)
		assert.Equal(t, "red", result)
	})

	t.Run("prefault function", func(t *testing.T) {
		counter := 0
		prefaultSchema := schema.PrefaultFunc(func() string {
			counter++
			return "blue"
		})

		// Valid value should not call function
		result, err := prefaultSchema.Parse("green")
		require.NoError(t, err)
		assert.Equal(t, "green", result)
		assert.Equal(t, 0, counter)

		// Invalid value should call function
		result, err = prefaultSchema.Parse("invalid")
		require.NoError(t, err)
		assert.Equal(t, "blue", result)
		assert.Equal(t, 1, counter)
	})

	t.Run("default with chain methods", func(t *testing.T) {
		defaultSchema := EnumMap(map[string]string{
			"SMALL":  "small",
			"MEDIUM": "medium",
			"LARGE":  "large",
		}).Default("medium")

		// Test Enum() method
		enumObj := defaultSchema.Enum()
		expected := map[string]string{
			"SMALL":  "small",
			"MEDIUM": "medium",
			"LARGE":  "large",
		}
		assert.Equal(t, expected, enumObj)

		// Test Options() method
		options := defaultSchema.Options()
		assert.Len(t, options, 3)
		assert.Contains(t, options, "small")
		assert.Contains(t, options, "medium")
		assert.Contains(t, options, "large")

		// Test Extract() method
		extracted := defaultSchema.Extract([]string{"SMALL", "LARGE"})

		// Default value should still work
		result, err := extracted.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "medium", result)

		// Extracted values should work
		result, err = extracted.Parse("small")
		require.NoError(t, err)
		assert.Equal(t, "small", result)

		// Non-extracted values should fail
		_, err = extracted.Parse("medium")
		assert.Error(t, err)

		// Test Exclude() method
		excluded := defaultSchema.Exclude([]string{"LARGE"})

		// Default value should still work
		result, err = excluded.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "medium", result)

		// Non-excluded values should work
		result, err = excluded.Parse("small")
		require.NoError(t, err)
		assert.Equal(t, "small", result)

		// Excluded values should fail
		_, err = excluded.Parse("large")
		assert.Error(t, err)
	})

	t.Run("prefault with chain methods", func(t *testing.T) {
		prefaultSchema := EnumMap(map[string]int{
			"LOW":    1,
			"MEDIUM": 2,
			"HIGH":   3,
		}).Prefault(2)

		// Test Enum() method
		enumObj := prefaultSchema.Enum()
		expected := map[string]int{
			"LOW":    1,
			"MEDIUM": 2,
			"HIGH":   3,
		}
		assert.Equal(t, expected, enumObj)

		// Test Options() method
		options := prefaultSchema.Options()
		assert.Len(t, options, 3)
		assert.Contains(t, options, 1)
		assert.Contains(t, options, 2)
		assert.Contains(t, options, 3)

		// Test Extract() method
		extracted := prefaultSchema.Extract([]string{"LOW", "HIGH"})

		// Extracted values should work
		result, err := extracted.Parse(1)
		require.NoError(t, err)
		assert.Equal(t, 1, result)

		result, err = extracted.Parse(3)
		require.NoError(t, err)
		assert.Equal(t, 3, result)

		// Non-extracted values should use prefault (not error)
		result, err = extracted.Parse(2)
		require.NoError(t, err)
		assert.Equal(t, 2, result) // Should return prefault value 2

		// Test Exclude() method
		excluded := prefaultSchema.Exclude([]string{"HIGH"})

		// Non-excluded values should work
		result, err = excluded.Parse(1)
		require.NoError(t, err)
		assert.Equal(t, 1, result)

		result, err = excluded.Parse(2)
		require.NoError(t, err)
		assert.Equal(t, 2, result)

		// Excluded values should use prefault (not error)
		result, err = excluded.Parse(3)
		require.NoError(t, err)
		assert.Equal(t, 2, result) // Should return prefault value 2
	})
}
