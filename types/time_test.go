package types

import (
	"testing"
	"time"

	"github.com/kaptinlin/gozod/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Basic functionality tests
// =============================================================================

func TestTime_BasicFunctionality(t *testing.T) {
	t.Run("valid time inputs", func(t *testing.T) {
		schema := Time()

		// Test current time
		now := time.Now()
		result, err := schema.Parse(now)
		require.NoError(t, err)
		assert.True(t, result.Equal(now))

		// Test specific time
		specificTime := time.Date(2023, 12, 25, 10, 30, 45, 0, time.UTC)
		result, err = schema.Parse(specificTime)
		require.NoError(t, err)
		assert.True(t, result.Equal(specificTime))

		// Test zero time
		zeroTime := time.Time{}
		result, err = schema.Parse(zeroTime)
		require.NoError(t, err)
		assert.True(t, result.Equal(zeroTime))
	})

	t.Run("invalid type inputs", func(t *testing.T) {
		schema := Time()

		invalidInputs := []any{
			"not a time", 123, 3.14, []time.Time{time.Now()}, nil,
			true, false, map[string]any{}, struct{}{},
		}

		for _, input := range invalidInputs {
			_, err := schema.Parse(input)
			assert.Error(t, err, "Expected error for input: %v", input)
		}
	})

	t.Run("Parse and MustParse methods", func(t *testing.T) {
		schema := Time()
		testTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

		// Test Parse method
		result, err := schema.Parse(testTime)
		require.NoError(t, err)
		assert.True(t, result.Equal(testTime))

		// Test MustParse method
		mustResult := schema.MustParse(testTime)
		assert.True(t, mustResult.Equal(testTime))

		// Test panic on invalid input
		assert.Panics(t, func() {
			schema.MustParse("invalid")
		})
	})

	t.Run("custom error message", func(t *testing.T) {
		customError := "Expected a time value"
		schema := Time(core.SchemaParams{Error: customError})

		require.NotNil(t, schema)
		assert.Equal(t, core.ZodTypeTime, schema.internals.Def.Type)

		_, err := schema.Parse("invalid")
		assert.Error(t, err)
	})
}

// =============================================================================
// Type safety tests
// =============================================================================

func TestTime_TypeSafety(t *testing.T) {
	t.Run("Time returns time.Time type", func(t *testing.T) {
		schema := Time()
		require.NotNil(t, schema)

		testTime := time.Date(2023, 6, 15, 14, 30, 0, 0, time.UTC)
		result, err := schema.Parse(testTime)
		require.NoError(t, err)
		assert.True(t, result.Equal(testTime))
		assert.IsType(t, time.Time{}, result) // Ensure type is time.Time

		// Test with different time
		now := time.Now()
		result, err = schema.Parse(now)
		require.NoError(t, err)
		assert.True(t, result.Equal(now))
		assert.IsType(t, time.Time{}, result)
	})

	t.Run("TimePtr returns *time.Time type", func(t *testing.T) {
		schema := TimePtr()
		require.NotNil(t, schema)

		testTime := time.Date(2023, 6, 15, 14, 30, 0, 0, time.UTC)
		result, err := schema.Parse(testTime)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.Equal(testTime))
		assert.IsType(t, (*time.Time)(nil), result) // Ensure type is *time.Time

		// Test with pointer input
		result, err = schema.Parse(&testTime)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.Equal(testTime))
	})

	t.Run("type inference with assignment", func(t *testing.T) {
		// Type-inference friendly API
		timeSchema := Time()   // time.Time type
		ptrSchema := TimePtr() // *time.Time type

		testTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

		// Test time.Time type
		result1, err1 := timeSchema.Parse(testTime)
		require.NoError(t, err1)
		assert.IsType(t, time.Time{}, result1)
		assert.True(t, result1.Equal(testTime))

		// Test *time.Time type
		result2, err2 := ptrSchema.Parse(testTime)
		require.NoError(t, err2)
		assert.IsType(t, (*time.Time)(nil), result2)
		require.NotNil(t, result2)
		assert.True(t, result2.Equal(testTime))
	})

	t.Run("MustParse type safety", func(t *testing.T) {
		testTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

		// Test time.Time type
		timeSchema := Time()
		result := timeSchema.MustParse(testTime)
		assert.IsType(t, time.Time{}, result)
		assert.True(t, result.Equal(testTime))

		// Test *time.Time type
		ptrSchema := TimePtr()
		ptrResult := ptrSchema.MustParse(testTime)
		assert.IsType(t, (*time.Time)(nil), ptrResult)
		require.NotNil(t, ptrResult)
		assert.True(t, ptrResult.Equal(testTime))
	})
}

// =============================================================================
// Modifier methods tests
// =============================================================================

func TestTime_Modifiers(t *testing.T) {
	t.Run("Optional always returns *time.Time", func(t *testing.T) {
		// From time.Time to *time.Time via Optional
		timeSchema := Time()
		optionalSchema := timeSchema.Optional()

		// Type check: ensure it returns *ZodTime[*time.Time]
		var _ = optionalSchema

		// Functionality test
		testTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		result, err := optionalSchema.Parse(testTime)
		require.NoError(t, err)
		assert.IsType(t, (*time.Time)(nil), result)
		require.NotNil(t, result)
		assert.True(t, result.Equal(testTime))

		// From *time.Time to *time.Time via Optional (maintains type)
		ptrSchema := TimePtr()
		optionalPtrSchema := ptrSchema.Optional()
		var _ = optionalPtrSchema
	})

	t.Run("Nilable always returns *time.Time", func(t *testing.T) {
		timeSchema := Time()
		nilableSchema := timeSchema.Nilable()

		var _ = nilableSchema

		// Test nil handling
		result, err := nilableSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Default preserves current type", func(t *testing.T) {
		defaultTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

		// time.Time maintains time.Time
		timeSchema := Time()
		defaultTimeSchema := timeSchema.Default(defaultTime)
		var _ = defaultTimeSchema

		// *time.Time maintains *time.Time
		ptrSchema := TimePtr()
		defaultPtrSchema := ptrSchema.Default(defaultTime)
		var _ = defaultPtrSchema
	})

	t.Run("Prefault preserves current type", func(t *testing.T) {
		prefaultTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

		// time.Time maintains time.Time
		timeSchema := Time()
		prefaultTimeSchema := timeSchema.Prefault(prefaultTime)
		var _ = prefaultTimeSchema

		// *time.Time maintains *time.Time
		ptrSchema := TimePtr()
		prefaultPtrSchema := ptrSchema.Prefault(prefaultTime)
		var _ = prefaultPtrSchema
	})

	t.Run("Nullish combines optional and nilable", func(t *testing.T) {
		timeSchema := Time()
		nullishSchema := timeSchema.Nullish()

		var _ = nullishSchema

		// Test nil handling
		result, err := nullishSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Test valid time
		testTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		result, err = nullishSchema.Parse(testTime)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.Equal(testTime))
	})
}

// =============================================================================
// Chaining tests
// =============================================================================

func TestTime_Chaining(t *testing.T) {
	defaultTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	t.Run("type evolution through chaining", func(t *testing.T) {
		// Chain with type evolution
		schema := Time(). // *ZodTime[time.Time]
					Default(defaultTime). // *ZodTime[time.Time] (maintains type)
					Optional()            // *ZodTime[*time.Time] (type conversion)

		var _ = schema

		// Test final behavior
		testTime := time.Date(2023, 6, 15, 12, 0, 0, 0, time.UTC)
		result, err := schema.Parse(testTime)
		require.NoError(t, err)
		assert.IsType(t, (*time.Time)(nil), result)
		require.NotNil(t, result)
		assert.True(t, result.Equal(testTime))
	})

	t.Run("complex chaining", func(t *testing.T) {
		schema := TimePtr(). // *ZodTime[*time.Time]
					Nilable().           // *ZodTime[*time.Time] (maintains type)
					Default(defaultTime) // *ZodTime[*time.Time] (maintains type)

		var _ = schema

		testTime := time.Date(2023, 12, 25, 15, 30, 0, 0, time.UTC)
		result, err := schema.Parse(testTime)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.Equal(testTime))
	})

	t.Run("default and prefault chaining", func(t *testing.T) {
		prefaultTime := time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC)

		schema := Time().
			Default(defaultTime).
			Prefault(prefaultTime)

		testTime := time.Date(2023, 6, 15, 10, 0, 0, 0, time.UTC)
		result, err := schema.Parse(testTime)
		require.NoError(t, err)
		assert.True(t, result.Equal(testTime))
	})
}

// =============================================================================
// Default and prefault tests
// =============================================================================

func TestTime_DefaultAndPrefault(t *testing.T) {
	defaultTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	prefaultTime := time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC)

	// Test 1: Default has higher priority than Prefault
	t.Run("Default has higher priority than Prefault", func(t *testing.T) {
		// Time type
		schema1 := Time().Default(defaultTime).Prefault(prefaultTime)
		result1, err1 := schema1.Parse(nil)
		require.NoError(t, err1)
		assert.True(t, result1.Equal(defaultTime)) // Should be default, not prefault

		// TimePtr type
		schema2 := TimePtr().Default(defaultTime).Prefault(prefaultTime)
		result2, err2 := schema2.Parse(nil)
		require.NoError(t, err2)
		require.NotNil(t, result2)
		assert.True(t, result2.Equal(defaultTime)) // Should be default, not prefault
	})

	// Test 2: Default short-circuits validation
	t.Run("Default short-circuits validation", func(t *testing.T) {
		// Default value violates time constraint but should still work
		minTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		oldTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC) // Before min
		schema1 := Time().Refine(func(t time.Time) bool {
			return t.After(minTime) || t.Equal(minTime) // Only allow times >= 2024
		}, "Time must be after 2024").Default(oldTime)
		result1, err1 := schema1.Parse(nil)
		require.NoError(t, err1)
		assert.True(t, result1.Equal(oldTime)) // Default bypasses validation

		// Default value violates refinement but should still work
		schema2 := Time().Refine(func(t time.Time) bool {
			return false // Always fail refinement
		}, "Should never pass").Default(defaultTime)
		result2, err2 := schema2.Parse(nil)
		require.NoError(t, err2)
		assert.True(t, result2.Equal(defaultTime)) // Default bypasses validation
	})

	// Test 3: Prefault goes through full validation
	t.Run("Prefault goes through full validation", func(t *testing.T) {
		// Prefault value passes validation
		minTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		validTime := time.Date(2023, 6, 15, 12, 0, 0, 0, time.UTC) // After min
		schema1 := Time().Refine(func(t time.Time) bool {
			return t.After(minTime) || t.Equal(minTime) // Only allow times >= 2023
		}, "Time must be after 2023").Prefault(validTime)
		result1, err1 := schema1.Parse(nil)
		require.NoError(t, err1)
		assert.True(t, result1.Equal(validTime))

		// Prefault value passes refinement
		schema2 := Time().Refine(func(t time.Time) bool {
			return t.Year() == 2023 // Only allow 2023 times
		}, "Must be in 2023").Prefault(prefaultTime)
		result2, err2 := schema2.Parse(nil)
		require.NoError(t, err2)
		assert.True(t, result2.Equal(prefaultTime))
	})

	// Test 4: Prefault only triggered by nil input
	t.Run("Prefault only triggered by nil input", func(t *testing.T) {
		schema := Time().Prefault(prefaultTime)

		// Valid input should override prefault
		testTime := time.Date(2023, 6, 15, 12, 0, 0, 0, time.UTC)
		result, err := schema.Parse(testTime)
		require.NoError(t, err)
		assert.True(t, result.Equal(testTime)) // Should be input, not prefault

		// Invalid input should NOT trigger prefault (should return error)
		_, err = schema.Parse("invalid-time")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid input: expected time, received string")
	})

	// Test 5: DefaultFunc and PrefaultFunc behavior
	t.Run("DefaultFunc and PrefaultFunc behavior", func(t *testing.T) {
		defaultCalled := false
		prefaultCalled := false

		defaultFunc := func() time.Time {
			defaultCalled = true
			return defaultTime
		}

		prefaultFunc := func() time.Time {
			prefaultCalled = true
			return prefaultTime
		}

		// Test DefaultFunc priority over PrefaultFunc
		schema1 := Time().DefaultFunc(defaultFunc).PrefaultFunc(prefaultFunc)
		result1, err1 := schema1.Parse(nil)
		require.NoError(t, err1)
		assert.True(t, defaultCalled)
		assert.False(t, prefaultCalled) // Should not be called due to default priority
		assert.True(t, result1.Equal(defaultTime))

		// Reset flags
		defaultCalled = false
		prefaultCalled = false

		// Test PrefaultFunc alone
		schema2 := Time().PrefaultFunc(prefaultFunc)
		result2, err2 := schema2.Parse(nil)
		require.NoError(t, err2)
		assert.True(t, prefaultCalled)
		assert.True(t, result2.Equal(prefaultTime))
	})

	// Test 6: Prefault validation failure
	t.Run("Prefault validation failure", func(t *testing.T) {
		// Prefault value violates time constraint
		minTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		oldTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC) // Before min
		schema1 := Time().Refine(func(t time.Time) bool {
			return t.After(minTime) || t.Equal(minTime) // Only allow times >= 2024
		}, "Time must be after 2024").Prefault(oldTime)
		_, err1 := schema1.Parse(nil)
		require.Error(t, err1) // Prefault should fail validation
		assert.Contains(t, err1.Error(), "Time must be after 2024")

		// Prefault value violates refinement
		schema2 := Time().Refine(func(t time.Time) bool {
			return t.Year() == 2024 // Only allow 2024 times
		}, "Must be in 2024").Prefault(prefaultTime) // 2023 time
		_, err2 := schema2.Parse(nil)
		require.Error(t, err2) // Prefault should fail validation
		assert.Contains(t, err2.Error(), "Must be in 2024")
	})

	// Test 7: Type preservation in Default and Prefault
	t.Run("Type preservation in Default and Prefault", func(t *testing.T) {
		// time.Time maintains time.Time
		timeSchema := Time()
		defaultTimeSchema := timeSchema.Default(defaultTime)
		var _ = defaultTimeSchema

		prefaultTimeSchema := timeSchema.Prefault(prefaultTime)
		var _ = prefaultTimeSchema

		// *time.Time maintains *time.Time
		ptrSchema := TimePtr()
		defaultPtrSchema := ptrSchema.Default(defaultTime)
		var _ = defaultPtrSchema

		prefaultPtrSchema := ptrSchema.Prefault(prefaultTime)
		var _ = prefaultPtrSchema
	})
}

// =============================================================================
// Refine tests
// =============================================================================

func TestTime_Refine(t *testing.T) {
	t.Run("refine validate", func(t *testing.T) {
		// Only accept times after 2023
		schema := Time().Refine(func(t time.Time) bool {
			return t.Year() >= 2023
		})

		validTime := time.Date(2023, 6, 15, 12, 0, 0, 0, time.UTC)
		result, err := schema.Parse(validTime)
		require.NoError(t, err)
		assert.True(t, result.Equal(validTime))

		invalidTime := time.Date(2022, 6, 15, 12, 0, 0, 0, time.UTC)
		_, err = schema.Parse(invalidTime)
		assert.Error(t, err)
	})

	t.Run("refine with custom error message", func(t *testing.T) {
		errorMessage := "Time must be after 2023"
		schema := TimePtr().Refine(func(t *time.Time) bool {
			return t != nil && t.Year() >= 2023
		}, core.SchemaParams{Error: errorMessage})

		validTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		result, err := schema.Parse(validTime)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.Equal(validTime))

		invalidTime := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
		_, err = schema.Parse(invalidTime)
		assert.Error(t, err)
	})

	t.Run("refine pointer allows nil", func(t *testing.T) {
		schema := TimePtr().Nilable().Refine(func(t *time.Time) bool {
			// Accept nil or times after 2023
			return t == nil || t.Year() >= 2023
		})

		// Expect nil to be accepted
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Valid time should pass
		validTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		result, err = schema.Parse(validTime)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.Equal(validTime))

		// Invalid time should fail
		invalidTime := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
		_, err = schema.Parse(invalidTime)
		assert.Error(t, err)
	})

	t.Run("refine with business logic", func(t *testing.T) {
		// Only accept weekdays
		schema := Time().Refine(func(t time.Time) bool {
			weekday := t.Weekday()
			return weekday != time.Saturday && weekday != time.Sunday
		})

		// Monday should pass
		monday := time.Date(2023, 6, 12, 12, 0, 0, 0, time.UTC) // Monday
		result, err := schema.Parse(monday)
		require.NoError(t, err)
		assert.True(t, result.Equal(monday))

		// Saturday should fail
		saturday := time.Date(2023, 6, 10, 12, 0, 0, 0, time.UTC) // Saturday
		_, err = schema.Parse(saturday)
		assert.Error(t, err)
	})
}

func TestTime_RefineAny(t *testing.T) {
	t.Run("refineAny time schema", func(t *testing.T) {
		// Only accept times after 2023 via RefineAny on Time() schema
		schema := Time().RefineAny(func(v any) bool {
			if t, ok := v.(time.Time); ok {
				return t.Year() >= 2023
			}
			return false
		})

		// Valid time passes
		validTime := time.Date(2023, 6, 15, 12, 0, 0, 0, time.UTC)
		result, err := schema.Parse(validTime)
		require.NoError(t, err)
		assert.True(t, result.Equal(validTime))

		// Invalid time fails
		invalidTime := time.Date(2022, 6, 15, 12, 0, 0, 0, time.UTC)
		_, err = schema.Parse(invalidTime)
		assert.Error(t, err)
	})

	t.Run("refineAny pointer schema", func(t *testing.T) {
		// TimePtr().RefineAny sees underlying time.Time value
		schema := TimePtr().RefineAny(func(v any) bool {
			if t, ok := v.(time.Time); ok {
				return t.Year() >= 2023 // accept only times after 2023
			}
			return false
		})

		validTime := time.Date(2023, 6, 15, 12, 0, 0, 0, time.UTC)
		result, err := schema.Parse(validTime)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.Equal(validTime))

		invalidTime := time.Date(2022, 6, 15, 12, 0, 0, 0, time.UTC)
		_, err = schema.Parse(invalidTime)
		assert.Error(t, err)
	})

	t.Run("refineAny with complex validation", func(t *testing.T) {
		// Complex validation: time must be in the future and during business hours
		now := time.Now()
		schema := Time().RefineAny(func(v any) bool {
			if t, ok := v.(time.Time); ok {
				// Must be in the future
				if t.Before(now) {
					return false
				}
				// Must be during business hours (9 AM to 6 PM)
				hour := t.Hour()
				return hour >= 9 && hour < 18
			}
			return false
		})

		// Create a specific future time during business hours (10 AM tomorrow in UTC)
		tomorrow := now.Add(time.Hour * 24)
		futureBusinessTime := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 10, 0, 0, 0, time.UTC)
		result, err := schema.Parse(futureBusinessTime)
		require.NoError(t, err)
		assert.True(t, result.Equal(futureBusinessTime))

		// Create a specific future time outside business hours (8 PM tomorrow in UTC)
		futureNightTime := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 20, 0, 0, 0, time.UTC)
		_, err = schema.Parse(futureNightTime)
		assert.Error(t, err)
	})
}

// =============================================================================
// Coercion tests
// =============================================================================

func TestTime_Coercion(t *testing.T) {
	t.Run("string coercion", func(t *testing.T) {
		schema := CoercedTime()

		// Test RFC3339 format
		timeStr := "2023-06-15T12:30:45Z"
		expected, _ := time.Parse(time.RFC3339, timeStr)
		result, err := schema.Parse(timeStr)
		require.NoError(t, err, "Should coerce RFC3339 string to time")
		assert.True(t, result.Equal(expected))

		// Test date only format
		dateStr := "2023-06-15"
		expected, _ = time.Parse("2006-01-02", dateStr)
		result, err = schema.Parse(dateStr)
		require.NoError(t, err, "Should coerce date string to time")
		assert.True(t, result.Equal(expected))

		// Test datetime format
		datetimeStr := "2023-06-15 12:30:45"
		expected, _ = time.Parse("2006-01-02 15:04:05", datetimeStr)
		result, err = schema.Parse(datetimeStr)
		require.NoError(t, err, "Should coerce datetime string to time")
		assert.True(t, result.Equal(expected))
	})

	t.Run("unix timestamp coercion", func(t *testing.T) {
		schema := CoercedTime()

		testCases := []struct {
			input    any
			name     string
			expected time.Time
		}{
			{int64(1686832245), "int64 unix timestamp", time.Unix(1686832245, 0)},
			{int(1686832245), "int unix timestamp", time.Unix(1686832245, 0)},
			{float64(1686832245), "float64 unix timestamp", time.Unix(1686832245, 0)},
			{float32(1686832245), "float32 unix timestamp", time.Unix(int64(float32(1686832245)), 0)}, // Account for float32 precision
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result, err := schema.Parse(tc.input)
				require.NoError(t, err)
				assert.True(t, result.Equal(tc.expected))
			})
		}
	})

	t.Run("time.Time passthrough", func(t *testing.T) {
		schema := CoercedTime()

		// time.Time should pass through unchanged
		now := time.Now()
		result, err := schema.Parse(now)
		require.NoError(t, err)
		assert.True(t, result.Equal(now))

		// *time.Time should also work
		result, err = schema.Parse(&now)
		require.NoError(t, err)
		assert.True(t, result.Equal(now))
	})

	t.Run("invalid coercion inputs", func(t *testing.T) {
		schema := CoercedTime()

		// Inputs that cannot be coerced
		invalidInputs := []any{
			"not a time", "invalid-date", "2023-13-01", // Invalid date strings
			[]time.Time{time.Now()}, map[string]any{}, struct{}{},
			true, false,
		}

		for _, input := range invalidInputs {
			_, err := schema.Parse(input)
			assert.Error(t, err, "Expected error for input: %v", input)
		}
	})

	t.Run("coerced time with modifiers", func(t *testing.T) {
		schema := CoercedTime().Optional()

		// Should handle nil
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Should coerce string
		timeStr := "2023-06-15T12:30:45Z"
		expected, _ := time.Parse(time.RFC3339, timeStr)
		result, err = schema.Parse(timeStr)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.Equal(expected))
	})
}

// =============================================================================
// Error handling tests
// =============================================================================

func TestTime_ErrorHandling(t *testing.T) {
	t.Run("invalid type error", func(t *testing.T) {
		schema := Time()

		_, err := schema.Parse("not a time")
		assert.Error(t, err)
	})

	t.Run("custom error message", func(t *testing.T) {
		customError := "Expected a time value"
		schema := TimePtr(core.SchemaParams{Error: customError})

		_, err := schema.Parse("not a time")
		assert.Error(t, err)
	})

	t.Run("nil handling matrix", func(t *testing.T) {
		testCases := []struct {
			name         string
			schema       func() any
			expectError  bool
			expectedType any
		}{
			{"Time() rejects nil", func() any { return Time() }, true, nil},
			{"TimePtr() rejects nil", func() any { return TimePtr() }, true, nil},
			{"Time().Nilable() accepts nil", func() any { return Time().Nilable() }, false, (*time.Time)(nil)},
			{"TimePtr().Nilable() accepts nil", func() any { return TimePtr().Nilable() }, false, (*time.Time)(nil)},
			{"Time().Optional() accepts nil", func() any { return Time().Optional() }, false, (*time.Time)(nil)},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				schema := tc.schema()

				// Type assertion to get Parse method - this is type-specific
				if timeSchema, ok := schema.(*ZodTime[time.Time]); ok {
					result, err := timeSchema.Parse(nil)
					if tc.expectError {
						assert.Error(t, err)
					} else {
						require.NoError(t, err)
						assert.IsType(t, tc.expectedType, result)
					}
				} else if ptrSchema, ok := schema.(*ZodTime[*time.Time]); ok {
					result, err := ptrSchema.Parse(nil)
					if tc.expectError {
						assert.Error(t, err)
					} else {
						require.NoError(t, err)
						assert.IsType(t, tc.expectedType, result)
					}
				}
			})
		}
	})

	t.Run("coercion error handling", func(t *testing.T) {
		schema := CoercedTime()

		// Test invalid date strings
		invalidDates := []string{
			"2023-13-01", // Invalid month
			"2023-02-30", // Invalid day
			"not-a-date",
			"",
		}

		for _, invalidDate := range invalidDates {
			_, err := schema.Parse(invalidDate)
			assert.Error(t, err, "Expected error for invalid date: %s", invalidDate)
		}
	})
}

// =============================================================================
// Edge case tests
// =============================================================================

func TestTime_EdgeCases(t *testing.T) {
	t.Run("nil handling with *time.Time", func(t *testing.T) {
		schema := TimePtr().Nilable()

		// Test nil input
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Test valid time
		testTime := time.Date(2023, 6, 15, 12, 0, 0, 0, time.UTC)
		result, err = schema.Parse(testTime)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.Equal(testTime))
	})

	t.Run("empty context", func(t *testing.T) {
		schema := Time()
		testTime := time.Date(2023, 6, 15, 12, 0, 0, 0, time.UTC)

		// Parse with empty context slice
		result, err := schema.Parse(testTime)
		require.NoError(t, err)
		assert.True(t, result.Equal(testTime))
	})

	t.Run("performance critical paths", func(t *testing.T) {
		schema := Time()

		// Test that fast paths work correctly
		t.Run("direct time.Time input fast path", func(t *testing.T) {
			testTime := time.Date(2023, 6, 15, 12, 0, 0, 0, time.UTC)
			result, err := schema.Parse(testTime)
			require.NoError(t, err)
			assert.True(t, result.Equal(testTime))
		})

		t.Run("zero time fast path", func(t *testing.T) {
			zeroTime := time.Time{}
			result, err := schema.Parse(zeroTime)
			require.NoError(t, err)
			assert.True(t, result.Equal(zeroTime))
		})
	})

	t.Run("memory efficiency verification", func(t *testing.T) {
		// Create multiple schemas to verify shared state
		defaultTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		testTime := time.Date(2023, 6, 15, 12, 0, 0, 0, time.UTC)

		schema1 := Time()
		schema2 := schema1.Default(defaultTime)
		schema3 := schema2.Optional()

		// All should work independently
		result1, err1 := schema1.Parse(testTime)
		require.NoError(t, err1)
		assert.True(t, result1.Equal(testTime))

		result2, err2 := schema2.Parse(testTime)
		require.NoError(t, err2)
		assert.True(t, result2.Equal(testTime))

		result3, err3 := schema3.Parse(testTime)
		require.NoError(t, err3)
		assert.NotNil(t, result3)
		assert.True(t, result3.Equal(testTime))
	})

	t.Run("transform and pipe integration", func(t *testing.T) {
		schema := Time()
		testTime := time.Date(2023, 6, 15, 12, 30, 45, 0, time.UTC)

		// Test Transform
		transform := schema.Transform(func(t time.Time, ctx *core.RefinementContext) (any, error) {
			return t.Format("2006-01-02"), nil
		})

		result, err := transform.Parse(testTime)
		require.NoError(t, err)
		assert.Equal(t, "2023-06-15", result)

		// Test extractTime helper function
		timeVal := testTime
		ptrVal := &timeVal

		extracted1 := extractTime[time.Time](timeVal)
		assert.True(t, extracted1.Equal(testTime))

		extracted2 := extractTime[*time.Time](ptrVal)
		assert.True(t, extracted2.Equal(testTime))

		nilPtr := (*time.Time)(nil)
		extracted3 := extractTime[*time.Time](nilPtr)
		assert.True(t, extracted3.IsZero())
	})

	t.Run("timezone handling", func(t *testing.T) {
		schema := Time()

		// Test different timezones
		utcTime := time.Date(2023, 6, 15, 12, 0, 0, 0, time.UTC)
		localTime := utcTime.In(time.Local)

		result1, err := schema.Parse(utcTime)
		require.NoError(t, err)
		assert.True(t, result1.Equal(utcTime))

		result2, err := schema.Parse(localTime)
		require.NoError(t, err)
		assert.True(t, result2.Equal(localTime))
	})

	t.Run("precision handling", func(t *testing.T) {
		schema := Time()

		// Test nanosecond precision
		preciseTime := time.Date(2023, 6, 15, 12, 30, 45, 123456789, time.UTC)
		result, err := schema.Parse(preciseTime)
		require.NoError(t, err)
		assert.True(t, result.Equal(preciseTime))
		assert.Equal(t, preciseTime.Nanosecond(), result.Nanosecond())
	})
}

// =============================================================================
// Additional helper tests
// =============================================================================

func TestTime_HelperFunctions(t *testing.T) {
	t.Run("extractTime helper", func(t *testing.T) {
		testTime := time.Date(2023, 6, 15, 12, 0, 0, 0, time.UTC)

		// Test with time.Time
		extracted := extractTime[time.Time](testTime)
		assert.True(t, extracted.Equal(testTime))

		// Test with *time.Time
		ptrTime := &testTime
		extracted = extractTime[*time.Time](ptrTime)
		assert.True(t, extracted.Equal(testTime))

		// Test with nil pointer
		var nilPtr *time.Time
		extracted = extractTime[*time.Time](nilPtr)
		assert.True(t, extracted.IsZero())
	})

	t.Run("WrapFn Transform functionality", func(t *testing.T) {
		schema := Time()
		testTime := time.Date(2023, 6, 15, 12, 0, 0, 0, time.UTC)

		// Test Transform with WrapFn pattern
		transform := schema.Transform(func(t time.Time, ctx *core.RefinementContext) (any, error) {
			return t.Format("2006-01-02 15:04:05"), nil
		})

		result, err := transform.Parse(testTime)
		require.NoError(t, err)
		assert.Equal(t, "2023-06-15 12:00:00", result)

		// Test Transform with different output type
		timestampTransform := schema.Transform(func(t time.Time, ctx *core.RefinementContext) (any, error) {
			return t.Unix(), nil
		})

		result, err = timestampTransform.Parse(testTime)
		require.NoError(t, err)
		assert.IsType(t, int64(0), result)
		assert.Equal(t, testTime.Unix(), result)
	})
}

// =============================================================================
// Overwrite functionality tests
// =============================================================================

func TestTime_Overwrite(t *testing.T) {
	t.Run("basic time transformations", func(t *testing.T) {
		// Test UTC conversion
		utcSchema := Time().Overwrite(func(t time.Time) time.Time {
			return t.UTC()
		})

		// Test with a time in a different timezone
		location, err := time.LoadLocation("Asia/Shanghai")
		require.NoError(t, err)

		shanghaiTime := time.Date(2023, 12, 25, 15, 30, 0, 0, location)
		result, err := utcSchema.Parse(shanghaiTime)
		require.NoError(t, err)

		// Should be converted to UTC
		assert.Equal(t, time.UTC, result.Location())
		assert.Equal(t, 7, result.Hour()) // 15:30 Shanghai = 07:30 UTC
		assert.Equal(t, 30, result.Minute())
	})

	t.Run("date truncation", func(t *testing.T) {
		// Test truncation to date part (remove time components)
		dateOnlySchema := Time().Overwrite(func(t time.Time) time.Time {
			return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		})

		inputTime := time.Date(2023, 6, 15, 14, 30, 45, 123456789, time.UTC)
		result, err := dateOnlySchema.Parse(inputTime)
		require.NoError(t, err)

		assert.Equal(t, 2023, result.Year())
		assert.Equal(t, time.June, result.Month())
		assert.Equal(t, 15, result.Day())
		assert.Equal(t, 0, result.Hour())
		assert.Equal(t, 0, result.Minute())
		assert.Equal(t, 0, result.Second())
		assert.Equal(t, 0, result.Nanosecond())
	})

	t.Run("time precision truncation", func(t *testing.T) {
		// Test truncation to hour precision
		hourSchema := Time().Overwrite(func(t time.Time) time.Time {
			return t.Truncate(time.Hour)
		})

		inputTime := time.Date(2023, 6, 15, 14, 30, 45, 123456789, time.UTC)
		result, err := hourSchema.Parse(inputTime)
		require.NoError(t, err)

		assert.Equal(t, 14, result.Hour())
		assert.Equal(t, 0, result.Minute())
		assert.Equal(t, 0, result.Second())
		assert.Equal(t, 0, result.Nanosecond())

		// Test truncation to minute precision
		minuteSchema := Time().Overwrite(func(t time.Time) time.Time {
			return t.Truncate(time.Minute)
		})

		result, err = minuteSchema.Parse(inputTime)
		require.NoError(t, err)

		assert.Equal(t, 14, result.Hour())
		assert.Equal(t, 30, result.Minute())
		assert.Equal(t, 0, result.Second())
		assert.Equal(t, 0, result.Nanosecond())
	})

	t.Run("timezone conversion", func(t *testing.T) {
		// Test forcing timezone to Asia/Shanghai
		shanghaiSchema := Time().Overwrite(func(t time.Time) time.Time {
			location, _ := time.LoadLocation("Asia/Shanghai")
			return t.In(location)
		})

		utcTime := time.Date(2023, 6, 15, 8, 0, 0, 0, time.UTC)
		result, err := shanghaiSchema.Parse(utcTime)
		require.NoError(t, err)

		assert.Equal(t, "Asia/Shanghai", result.Location().String())
		assert.Equal(t, 16, result.Hour()) // 8:00 UTC = 16:00 Shanghai
	})

	t.Run("business logic transformations", func(t *testing.T) {
		// Test normalization to Monday 9:00 AM of the same week
		workWeekSchema := Time().Overwrite(func(t time.Time) time.Time {
			weekday := int(t.Weekday())
			if weekday == 0 { // Sunday
				weekday = 7
			}
			monday := t.AddDate(0, 0, 1-weekday)
			return time.Date(monday.Year(), monday.Month(), monday.Day(), 9, 0, 0, 0, t.Location())
		})

		// Test with a Wednesday
		wednesday := time.Date(2023, 6, 14, 15, 30, 0, 0, time.UTC) // Wednesday
		result, err := workWeekSchema.Parse(wednesday)
		require.NoError(t, err)

		// Should be Monday of the same week at 9:00 AM
		assert.Equal(t, time.Monday, result.Weekday())
		assert.Equal(t, 12, result.Day()) // Monday 2023-06-12
		assert.Equal(t, 9, result.Hour())
		assert.Equal(t, 0, result.Minute())

		// Test with a Sunday
		sunday := time.Date(2023, 6, 18, 20, 0, 0, 0, time.UTC) // Sunday
		result, err = workWeekSchema.Parse(sunday)
		require.NoError(t, err)

		assert.Equal(t, time.Monday, result.Weekday())
		assert.Equal(t, 12, result.Day()) // Monday 2023-06-12 (previous Monday)
		assert.Equal(t, 9, result.Hour())
	})

	t.Run("business hours enforcement", func(t *testing.T) {
		// Test business hours enforcement (9 AM - 6 PM)
		businessHoursSchema := Time().Overwrite(func(t time.Time) time.Time {
			hour := t.Hour()
			if hour < 9 {
				// Before 9:00, adjust to 9:00
				return time.Date(t.Year(), t.Month(), t.Day(), 9, 0, 0, 0, t.Location())
			} else if hour >= 18 {
				// After 18:00, adjust to next day 9:00
				nextDay := t.AddDate(0, 0, 1)
				return time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), 9, 0, 0, 0, t.Location())
			}
			return t
		})

		// Test early morning (should be adjusted to 9:00)
		earlyTime := time.Date(2023, 6, 15, 7, 30, 0, 0, time.UTC)
		result, err := businessHoursSchema.Parse(earlyTime)
		require.NoError(t, err)

		assert.Equal(t, 9, result.Hour())
		assert.Equal(t, 0, result.Minute())
		assert.Equal(t, 15, result.Day()) // Same day

		// Test evening (should be adjusted to next day 9:00)
		eveningTime := time.Date(2023, 6, 15, 20, 30, 0, 0, time.UTC)
		result, err = businessHoursSchema.Parse(eveningTime)
		require.NoError(t, err)

		assert.Equal(t, 9, result.Hour())
		assert.Equal(t, 0, result.Minute())
		assert.Equal(t, 16, result.Day()) // Next day

		// Test during business hours (should remain unchanged)
		businessTime := time.Date(2023, 6, 15, 14, 30, 0, 0, time.UTC)
		result, err = businessHoursSchema.Parse(businessTime)
		require.NoError(t, err)

		assert.Equal(t, 14, result.Hour())
		assert.Equal(t, 30, result.Minute())
		assert.Equal(t, 15, result.Day()) // Same day
	})

	t.Run("chaining with other validations", func(t *testing.T) {
		// Test chaining Overwrite with Refine using fixed time references
		// to avoid time-dependent test failures

		// Use a fixed reference time for predictable testing
		referenceTime := time.Date(2023, 6, 15, 10, 0, 0, 0, time.UTC) // 10:00 AM UTC

		// Schema that normalizes to business hours and validates against reference time
		businessHoursSchema := Time().
			Overwrite(func(t time.Time) time.Time {
				// Normalize to business hours
				hour := t.Hour()
				if hour < 9 || hour >= 18 {
					return time.Date(t.Year(), t.Month(), t.Day(), 9, 0, 0, 0, t.Location())
				}
				return t
			}).
			Refine(func(t time.Time) bool {
				// Validate that time is after our reference time
				return t.After(referenceTime.Add(-2 * time.Hour))
			}, "Time must be within acceptable range")

		// Test case 1: Time during business hours (should pass through unchanged)
		businessTime := time.Date(2023, 6, 15, 14, 30, 0, 0, time.UTC) // 2:30 PM
		result, err := businessHoursSchema.Parse(businessTime)
		require.NoError(t, err)
		assert.Equal(t, 14, result.Hour()) // Should remain 2:30 PM
		assert.Equal(t, 30, result.Minute())

		// Test case 2: Early morning time (should be adjusted to 9:00 AM and pass validation)
		earlyTime := time.Date(2023, 6, 15, 7, 0, 0, 0, time.UTC) // 7:00 AM
		result, err = businessHoursSchema.Parse(earlyTime)
		require.NoError(t, err)
		assert.Equal(t, 9, result.Hour()) // Should be adjusted to 9:00 AM
		assert.Equal(t, 0, result.Minute())

		// Test case 3: Test validation failure with a time that would fail the Refine check
		oldTime := time.Date(2023, 6, 10, 7, 0, 0, 0, time.UTC) // Much older time
		_, err = businessHoursSchema.Parse(oldTime)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Time must be within acceptable range")
	})

	t.Run("multiple overwrite calls", func(t *testing.T) {
		// Test chaining multiple Overwrite calls
		multiTransformSchema := Time().
			Overwrite(func(t time.Time) time.Time {
				return t.UTC() // First: convert to UTC
			}).
			Overwrite(func(t time.Time) time.Time {
				return t.Truncate(time.Hour) // Second: truncate to hour
			}).
			Overwrite(func(t time.Time) time.Time {
				// Third: adjust to business hours
				hour := t.Hour()
				if hour < 9 {
					// Before 9:00, adjust to 9:00 same day
					return time.Date(t.Year(), t.Month(), t.Day(), 9, 0, 0, 0, t.Location())
				} else if hour >= 18 {
					// After 18:00, adjust to next day 9:00
					nextDay := t.AddDate(0, 0, 1)
					return time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), 9, 0, 0, 0, t.Location())
				}
				return t
			})

		// Test with a time that needs all transformations
		location, err := time.LoadLocation("Asia/Shanghai")
		require.NoError(t, err)

		inputTime := time.Date(2023, 6, 15, 7, 30, 45, 123456789, location)
		result, err := multiTransformSchema.Parse(inputTime)
		require.NoError(t, err)

		// Should be: Shanghai -> UTC (23:30 prev day) -> truncate to hour (23:00) -> business hours (9:00 next day)
		assert.Equal(t, time.UTC, result.Location())
		assert.Equal(t, 9, result.Hour())       // Adjusted to business hours
		assert.Equal(t, 0, result.Minute())     // Truncated
		assert.Equal(t, 0, result.Second())     // Truncated
		assert.Equal(t, 0, result.Nanosecond()) // Truncated
		assert.Equal(t, 15, result.Day())       // Should be the next day (original day in Shanghai time)
	})

	t.Run("type preservation", func(t *testing.T) {
		// Test that Overwrite preserves the original type
		timeSchema := Time()
		overwriteSchema := timeSchema.Overwrite(func(t time.Time) time.Time {
			return t.UTC()
		})

		// Both should have the same type
		testTime := time.Date(2023, 6, 15, 14, 30, 0, 0, time.UTC)

		result1, err1 := timeSchema.Parse(testTime)
		require.NoError(t, err1)

		result2, err2 := overwriteSchema.Parse(testTime)
		require.NoError(t, err2)

		// Both results should be of type time.Time
		assert.IsType(t, time.Time{}, result1)
		assert.IsType(t, time.Time{}, result2)
	})

	t.Run("pointer type handling", func(t *testing.T) {
		// Pointer Overwrite should now work and preserve pointer identity
		ptrSchema := TimePtr().Nilable().Overwrite(func(t *time.Time) *time.Time {
			if t == nil {
				return nil
			}
			return new(t.UTC())
		})

		// Test with non-nil value
		testTime := time.Date(2023, 6, 15, 14, 30, 0, 0, time.Local)
		result, err := ptrSchema.Parse(&testTime)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, time.UTC, result.Location())

		// Test with nil value
		result, err = ptrSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("coerced time overwrite", func(t *testing.T) {
		// Test with coerced time schema
		coercedSchema := CoercedTime().Overwrite(func(t time.Time) time.Time {
			// Always set to noon UTC
			return time.Date(t.Year(), t.Month(), t.Day(), 12, 0, 0, 0, time.UTC)
		})

		// Test with string input that can be coerced
		result, err := coercedSchema.Parse("2023-06-15T14:30:00Z")
		require.NoError(t, err)

		assert.Equal(t, 2023, result.Year())
		assert.Equal(t, time.June, result.Month())
		assert.Equal(t, 15, result.Day())
		assert.Equal(t, 12, result.Hour()) // Should be noon
		assert.Equal(t, 0, result.Minute())
		assert.Equal(t, time.UTC, result.Location())
	})

	t.Run("error handling", func(t *testing.T) {
		// Test that invalid inputs still produce errors
		schema := Time().Overwrite(func(t time.Time) time.Time {
			return t.UTC()
		})

		// Invalid input should still cause an error
		_, err := schema.Parse("not a time")
		assert.Error(t, err)

		_, err = schema.Parse(12345)
		assert.Error(t, err)

		_, err = schema.Parse(nil)
		assert.Error(t, err)
	})
}

func TestTime_Check(t *testing.T) {
	t.Run("check with payload access", func(t *testing.T) {
		schema := Time().Check(func(v time.Time, payload *core.ParsePayload) {
			if v.Year() < 2023 {
				payload.AddIssue(core.ZodRawIssue{Message: "year must be >= 2023"})
			}
		})

		valid := time.Date(2023, 6, 15, 12, 0, 0, 0, time.UTC)
		result, err := schema.Parse(valid)
		require.NoError(t, err)
		assert.True(t, result.Equal(valid))

		invalid := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
		_, err = schema.Parse(invalid)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "year must be >= 2023")
	})

	t.Run("check pointer type", func(t *testing.T) {
		schema := TimePtr().Check(func(v *time.Time, payload *core.ParsePayload) {
			if v != nil && v.Weekday() == time.Sunday {
				payload.AddIssue(core.ZodRawIssue{Message: "no sundays"})
			}
		})

		monday := time.Date(2023, 6, 12, 12, 0, 0, 0, time.UTC)
		result, err := schema.Parse(monday)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.Equal(monday))

		sunday := time.Date(2023, 6, 18, 12, 0, 0, 0, time.UTC)
		_, err = schema.Parse(sunday)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no sundays")
	})
}

func TestTime_With(t *testing.T) {
	t.Run("with is alias for check", func(t *testing.T) {
		schema := Time().With(func(v time.Time, payload *core.ParsePayload) {
			if v.Hour() < 9 || v.Hour() >= 18 {
				payload.AddIssue(core.ZodRawIssue{Message: "business hours only"})
			}
		})

		valid := time.Date(2023, 6, 15, 10, 0, 0, 0, time.UTC)
		result, err := schema.Parse(valid)
		require.NoError(t, err)
		assert.True(t, result.Equal(valid))

		invalid := time.Date(2023, 6, 15, 20, 0, 0, 0, time.UTC)
		_, err = schema.Parse(invalid)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "business hours only")
	})
}

func TestTime_NonOptional(t *testing.T) {
	t.Run("removes optional flag", func(t *testing.T) {
		optional := Time().Optional()
		assert.True(t, optional.IsOptional())

		required := optional.NonOptional()
		assert.False(t, required.IsOptional())

		valid := time.Date(2023, 6, 15, 12, 0, 0, 0, time.UTC)
		result, err := required.Parse(valid)
		require.NoError(t, err)
		assert.True(t, result.Equal(valid))

		_, err = required.Parse(nil)
		assert.Error(t, err)
	})

	t.Run("returns time.Time constraint", func(t *testing.T) {
		schema := TimePtr().Optional().NonOptional()

		valid := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		result, err := schema.Parse(valid)
		require.NoError(t, err)
		assert.IsType(t, time.Time{}, result)
		assert.True(t, result.Equal(valid))
	})
}
