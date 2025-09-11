package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// BASIC FUNCTIONALITY TESTS
// =============================================================================

func TestISODateTime_Basic(t *testing.T) {
	schema := IsoDateTime()

	t.Run("valid_formats", func(t *testing.T) {
		validCases := []string{
			"2023-12-25T15:30:45Z",
			"2023-12-25T15:30:45.123Z",
			"2023-12-25T15:30:45+08:00",
			"2023-01-01T00:00:00Z",
		}

		for _, tc := range validCases {
			result, err := schema.Parse(tc)
			require.NoError(t, err, "should accept: %s", tc)
			assert.Equal(t, tc, result)
		}
	})

	t.Run("invalid_formats", func(t *testing.T) {
		invalidCases := []string{
			"invalid-datetime",
			"2023-12-25 15:30:45",  // missing T
			"2023-12-25T25:30:45Z", // invalid hour
			"2023-13-25T15:30:45Z", // invalid month
			"",                     // empty
		}

		for _, tc := range invalidCases {
			_, err := schema.Parse(tc)
			assert.Error(t, err, "should reject: %s", tc)
		}
	})

	t.Run("wrong_input_types", func(t *testing.T) {
		wrongTypes := []any{123, true, nil, []string{"2023-12-25T15:30:45Z"}}
		for _, input := range wrongTypes {
			_, err := schema.Parse(input)
			assert.Error(t, err, "should reject: %v", input)
		}
	})
}

func TestISODate_Basic(t *testing.T) {
	schema := IsoDate()

	t.Run("valid_formats", func(t *testing.T) {
		validCases := []string{
			"2023-12-25",
			"2023-01-01",
			"2024-02-29", // leap year
		}

		for _, tc := range validCases {
			result, err := schema.Parse(tc)
			require.NoError(t, err, "should accept: %s", tc)
			assert.Equal(t, tc, result)
		}
	})

	t.Run("invalid_formats", func(t *testing.T) {
		invalidCases := []string{
			"invalid-date",
			"25/12/2023",
			"2023-13-01", // invalid month
			"2023-12-32", // invalid day
			"",           // empty
		}

		for _, tc := range invalidCases {
			_, err := schema.Parse(tc)
			assert.Error(t, err, "should reject: %s", tc)
		}
	})
}

func TestISOTime_Basic(t *testing.T) {
	schema := IsoTime()

	t.Run("valid_formats", func(t *testing.T) {
		validCases := []string{
			"15:30",
			"15:30:45",
			"15:30:45.123",
			"00:00:00",
			"23:59:59",
		}

		for _, tc := range validCases {
			result, err := schema.Parse(tc)
			require.NoError(t, err, "should accept: %s", tc)
			assert.Equal(t, tc, result)
		}
	})

	t.Run("invalid_formats", func(t *testing.T) {
		invalidCases := []string{
			"invalid-time",
			"25:00:00",  // invalid hour
			"12:60:00",  // invalid minute
			"12:30:60",  // invalid second
			"12:30:00Z", // no timezone allowed
			"",          // empty
		}

		for _, tc := range invalidCases {
			_, err := schema.Parse(tc)
			assert.Error(t, err, "should reject: %s", tc)
		}
	})
}

func TestISODuration_Basic(t *testing.T) {
	schema := IsoDuration()

	t.Run("valid_formats", func(t *testing.T) {
		validCases := []string{
			"P1Y2M3DT4H5M6S", // full format
			"P1Y",            // years only
			"P1D",            // days only
			"PT1H",           // hours only
			"PT1M",           // minutes only
			"PT1S",           // seconds only
			"P1W",            // weeks
		}

		for _, tc := range validCases {
			result, err := schema.Parse(tc)
			require.NoError(t, err, "should accept: %s", tc)
			assert.Equal(t, tc, result)
		}
	})

	t.Run("invalid_formats", func(t *testing.T) {
		invalidCases := []string{
			"invalid-duration",
			"P",    // empty
			"PT",   // empty time part
			"1Y2M", // missing P
			"",     // empty
		}

		for _, tc := range invalidCases {
			_, err := schema.Parse(tc)
			assert.Error(t, err, "should reject: %s", tc)
		}
	})
}

// =============================================================================
// OPTIONS AND PRECISION TESTS
// =============================================================================

func TestISODateTime_Options(t *testing.T) {
	t.Run("offset_option", func(t *testing.T) {
		schema := IsoDateTime(IsoDatetimeOptions{Offset: true})

		// Should accept timezone offsets
		validCases := []string{
			"2020-01-01T06:15:00Z",
			"2020-01-01T06:15:00+02:00",
			"2020-01-01T06:15:00-05:00",
		}

		for _, tc := range validCases {
			result, err := schema.Parse(tc)
			require.NoError(t, err, "should accept: %s", tc)
			assert.Equal(t, tc, result)
		}

		// Should reject invalid offset formats
		_, err := schema.Parse("2020-01-01T06:15:00+02")
		assert.Error(t, err, "should reject incomplete offset")
	})

	t.Run("local_option", func(t *testing.T) {
		schema := IsoDateTime(IsoDatetimeOptions{Local: true})

		// Should accept local time (no timezone)
		validCases := []string{
			"2020-01-01T06:15:00Z", // still accept Z
			"2020-01-01T06:15:00",  // local time
			"2020-01-01T06:15",     // optional seconds
		}

		for _, tc := range validCases {
			result, err := schema.Parse(tc)
			require.NoError(t, err, "should accept: %s", tc)
			assert.Equal(t, tc, result)
		}
	})

	t.Run("precision_option", func(t *testing.T) {
		// Test minute precision
		minuteSchema := IsoDateTime(IsoDatetimeOptions{Precision: PrecisionMinute})
		result, err := minuteSchema.Parse("2020-01-01T06:15Z")
		require.NoError(t, err)
		assert.Equal(t, "2020-01-01T06:15Z", result)

		_, err = minuteSchema.Parse("2020-01-01T06:15:00Z")
		assert.Error(t, err, "should reject seconds for minute precision")

		// Test second precision
		secondSchema := IsoDateTime(IsoDatetimeOptions{Precision: PrecisionSecond})
		result, err = secondSchema.Parse("2020-01-01T06:15:00Z")
		require.NoError(t, err)
		assert.Equal(t, "2020-01-01T06:15:00Z", result)

		_, err = secondSchema.Parse("2020-01-01T06:15:00.123Z")
		assert.Error(t, err, "should reject milliseconds for second precision")

		// Test millisecond precision
		millisecondSchema := IsoDateTime(IsoDatetimeOptions{Precision: PrecisionMillisecond})
		result, err = millisecondSchema.Parse("2020-01-01T06:15:00.123Z")
		require.NoError(t, err)
		assert.Equal(t, "2020-01-01T06:15:00.123Z", result)

		_, err = millisecondSchema.Parse("2020-01-01T06:15:00Z")
		assert.Error(t, err, "should reject format without milliseconds")
	})
}

func TestISOTime_Options(t *testing.T) {
	t.Run("precision_options", func(t *testing.T) {
		// Test minute precision
		minuteSchema := IsoTime(IsoTimeOptions{Precision: PrecisionMinute})
		result, err := minuteSchema.Parse("15:30")
		require.NoError(t, err)
		assert.Equal(t, "15:30", result)

		_, err = minuteSchema.Parse("15:30:00")
		assert.Error(t, err, "should reject seconds")

		// Test second precision
		secondSchema := IsoTime(IsoTimeOptions{Precision: PrecisionSecond})
		result, err = secondSchema.Parse("15:30:00")
		require.NoError(t, err)
		assert.Equal(t, "15:30:00", result)

		_, err = secondSchema.Parse("15:30")
		assert.Error(t, err, "should reject minute-only format")

		// Test millisecond precision
		millisecondSchema := IsoTime(IsoTimeOptions{Precision: PrecisionMillisecond})
		result, err = millisecondSchema.Parse("15:30:00.123")
		require.NoError(t, err)
		assert.Equal(t, "15:30:00.123", result)

		_, err = millisecondSchema.Parse("15:30:00")
		assert.Error(t, err, "should reject format without milliseconds")
	})
}

// =============================================================================
// MODIFIERS TESTS
// =============================================================================

func TestISO_Modifiers(t *testing.T) {
	t.Run("optional_modifier", func(t *testing.T) {
		schema := IsoDateTime().Optional()

		// Valid datetime
		result, err := schema.Parse("2023-12-25T15:30:45Z")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "2023-12-25T15:30:45Z", *result)

		// nil should be allowed
		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("default_modifier", func(t *testing.T) {
		defaultValue := "2023-01-01T00:00:00Z"
		schema := IsoDateTime().Default(defaultValue)

		// Valid input
		result, err := schema.Parse("2023-12-25T15:30:45Z")
		require.NoError(t, err)
		assert.Equal(t, "2023-12-25T15:30:45Z", result)

		// nil should return default
		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, defaultValue, result)
	})

	t.Run("default_func_modifier", func(t *testing.T) {
		schema := IsoDateTime().DefaultFunc(func() string {
			return "2023-01-01T00:00:00Z"
		})

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "2023-01-01T00:00:00Z", result)
	})
}

// =============================================================================
// Default and prefault tests
// =============================================================================

func TestISO_DefaultAndPrefault(t *testing.T) {
	// Test 1: Default has higher priority than Prefault
	t.Run("Default priority over Prefault", func(t *testing.T) {
		schema := IsoDateTime().Default("2023-01-01T00:00:00Z").Prefault("2023-12-31T23:59:59Z")

		// When input is nil, Default should take precedence
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "2023-01-01T00:00:00Z", result)
	})

	// Test 2: Default short-circuit mechanism
	t.Run("Default short-circuit bypasses validation", func(t *testing.T) {
		schema := IsoDateTime().Min("2023-06-01T00:00:00Z").Default("2023-01-01T00:00:00Z") // Default violates Min constraint

		// Default should bypass validation even if it violates constraints
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "2023-01-01T00:00:00Z", result)
	})

	// Test 3: Prefault requires full validation
	t.Run("Prefault requires full validation", func(t *testing.T) {
		schema := IsoDateTime().Min("2023-06-01T00:00:00Z").Prefault("2023-01-01T00:00:00Z") // Prefault violates Min constraint

		// Prefault should fail validation if it violates constraints
		_, err := schema.Parse(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Too small")
	})

	// Test 4: Prefault only triggers on nil input
	t.Run("Prefault only triggers on nil input", func(t *testing.T) {
		schema := IsoDateTime().Min("2023-06-01T00:00:00Z").Prefault("2023-12-31T23:59:59Z")

		// Non-nil input that fails validation should not trigger Prefault
		_, err := schema.Parse("2023-01-01T00:00:00Z") // This input violates Min constraint
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Too small")
	})

	// Test 5: DefaultFunc and PrefaultFunc behavior
	t.Run("DefaultFunc and PrefaultFunc behavior", func(t *testing.T) {
		defaultCalled := false
		prefaultCalled := false

		schema := IsoDateTime().DefaultFunc(func() string {
			defaultCalled = true
			return "2023-01-01T00:00:00Z"
		}).PrefaultFunc(func() string {
			prefaultCalled = true
			return "2023-12-31T23:59:59Z"
		})

		// DefaultFunc should be called and take precedence
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "2023-01-01T00:00:00Z", result)
		assert.True(t, defaultCalled)
		assert.False(t, prefaultCalled) // PrefaultFunc should not be called
	})

	// Test 6: Error handling for Prefault validation failure
	t.Run("Prefault validation failure returns error", func(t *testing.T) {
		schema := IsoDate().Min("2023-06-01").Prefault("2023-01-01") // Prefault violates Min constraint

		// Should return validation error, not attempt fallback
		_, err := schema.Parse(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Too small")
	})
}

// =============================================================================
// RANGE VALIDATION TESTS
// =============================================================================

func TestISO_RangeValidation(t *testing.T) {
	t.Run("datetime_range", func(t *testing.T) {
		schema := IsoDateTime().Min("2024-01-15T10:00:00Z").Max("2024-01-15T18:00:00Z")

		// Should accept values within range
		result, err := schema.Parse("2024-01-15T14:30:45Z")
		require.NoError(t, err)
		assert.Equal(t, "2024-01-15T14:30:45Z", result)

		// Should accept boundary values
		_, err = schema.Parse("2024-01-15T10:00:00Z")
		assert.NoError(t, err)

		_, err = schema.Parse("2024-01-15T18:00:00Z")
		assert.NoError(t, err)

		// Should reject values outside range
		_, err = schema.Parse("2024-01-15T09:59:59Z")
		assert.Error(t, err, "should reject datetime before minimum")

		_, err = schema.Parse("2024-01-15T18:00:01Z")
		assert.Error(t, err, "should reject datetime after maximum")
	})

	t.Run("date_range", func(t *testing.T) {
		schema := IsoDate().Min("2022-11-05").Max("2022-11-10")

		// Should accept dates within range
		_, err := schema.Parse("2022-11-07")
		assert.NoError(t, err)

		// Should reject dates outside range
		_, err = schema.Parse("2022-11-04")
		assert.Error(t, err, "should reject date before minimum")

		_, err = schema.Parse("2022-11-11")
		assert.Error(t, err, "should reject date after maximum")
	})

	t.Run("time_range", func(t *testing.T) {
		schema := IsoTime().Min("09:00:00").Max("17:00:00")

		// Should accept times within range
		_, err := schema.Parse("12:30:45")
		assert.NoError(t, err)

		// Should reject times outside range
		_, err = schema.Parse("08:59:59")
		assert.Error(t, err, "should reject time before minimum")

		_, err = schema.Parse("17:00:01")
		assert.Error(t, err, "should reject time after maximum")
	})
}

// =============================================================================
// TYPE SAFETY TESTS
// =============================================================================

func TestISO_TypeSafety(t *testing.T) {
	t.Run("basic_schema_returns_string", func(t *testing.T) {
		schema := IsoDateTime()
		result := schema.MustParse("2023-12-25T15:30:45Z")
		assert.Equal(t, "2023-12-25T15:30:45Z", result)
	})

	t.Run("optional_schema_returns_pointer", func(t *testing.T) {
		schema := IsoDateTime().Optional()
		result := schema.MustParse("2023-12-25T15:30:45Z")
		require.NotNil(t, result)
		assert.Equal(t, "2023-12-25T15:30:45Z", *result)

		nilResult := schema.MustParse(nil)
		assert.Nil(t, nilResult)
	})

	t.Run("ptr_schema", func(t *testing.T) {
		schema := IsoDateTimePtr()
		result := schema.MustParse("2023-12-25T15:30:45Z")
		require.NotNil(t, result)
		assert.Equal(t, "2023-12-25T15:30:45Z", *result)
	})
}

// =============================================================================
// INTEGRATION TESTS
// =============================================================================

func TestISO_Integration(t *testing.T) {
	t.Run("real_world_usage", func(t *testing.T) {
		// Define schemas for event system
		datetimeSchema := IsoDateTime()
		dateSchema := IsoDate()
		timeSchema := IsoTime()
		durationSchema := IsoDuration()

		// Validate event data
		startDateTime, err := datetimeSchema.Parse("2023-12-25T09:00:00Z")
		require.NoError(t, err)
		assert.Equal(t, "2023-12-25T09:00:00Z", startDateTime)

		date, err := dateSchema.Parse("2023-12-25")
		require.NoError(t, err)
		assert.Equal(t, "2023-12-25", date)

		time, err := timeSchema.Parse("09:00:00")
		require.NoError(t, err)
		assert.Equal(t, "09:00:00", time)

		duration, err := durationSchema.Parse("PT8H")
		require.NoError(t, err)
		assert.Equal(t, "PT8H", duration)
	})

	t.Run("combined_options", func(t *testing.T) {
		// Test combining precision with other options
		schema := IsoDateTime(IsoDatetimeOptions{
			Precision: PrecisionMillisecond,
			Offset:    true,
			Local:     true,
		})

		validCases := []string{
			"2023-12-25T15:30:45.123Z",      // UTC with milliseconds
			"2023-12-25T15:30:45.123+08:00", // Offset with milliseconds
			"2023-12-25T15:30:45.123",       // Local with milliseconds
		}

		for _, tc := range validCases {
			result, err := schema.Parse(tc)
			require.NoError(t, err, "should accept: %s", tc)
			assert.Equal(t, tc, result)
		}

		// Should still enforce precision requirements
		_, err := schema.Parse("2023-12-25T15:30:45Z")
		assert.Error(t, err, "should reject format without milliseconds")
	})
}

// =============================================================================
// ERROR HANDLING TESTS
// =============================================================================

func TestISO_ErrorHandling(t *testing.T) {
	t.Run("custom_error_messages", func(t *testing.T) {
		schema := IsoDateTime("Please provide a valid ISO 8601 datetime format")
		_, err := schema.Parse("invalid-datetime")
		assert.Error(t, err)
	})

	t.Run("must_parse_panics", func(t *testing.T) {
		schema := IsoDateTime()
		assert.Panics(t, func() {
			schema.MustParse("invalid-datetime")
		})
	})

	t.Run("empty_string_handling", func(t *testing.T) {
		testCases := []struct {
			name   string
			schema func() (string, error)
		}{
			{"datetime", func() (string, error) { return IsoDateTime().Parse("") }},
			{"date", func() (string, error) { return IsoDate().Parse("") }},
			{"time", func() (string, error) { return IsoTime().Parse("") }},
			{"duration", func() (string, error) { return IsoDuration().Parse("") }},
		}

		for _, tc := range testCases {
			_, err := tc.schema()
			assert.Error(t, err, "%s should reject empty string", tc.name)
		}
	})
}
