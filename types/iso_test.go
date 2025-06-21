package types

import (
	"testing"
	"time"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// 1. Basic functionality and type inference
// =============================================================================

func TestISOBasicFunctionality(t *testing.T) {
	t.Run("date constructor", func(t *testing.T) {
		schema := ISO.Date()
		require.NotNil(t, schema)

		result, err := schema.Parse("2023-12-25")
		require.NoError(t, err)
		assert.Equal(t, "2023-12-25", result)
	})

	t.Run("time constructor", func(t *testing.T) {
		schema := ISO.Time()
		require.NotNil(t, schema)

		result, err := schema.Parse("14:30:00")
		require.NoError(t, err)
		assert.Equal(t, "14:30:00", result)
	})

	t.Run("datetime constructor", func(t *testing.T) {
		schema := ISO.DateTime()
		require.NotNil(t, schema)

		result, err := schema.Parse("2023-12-25T14:30:00Z")
		require.NoError(t, err)
		assert.Equal(t, "2023-12-25T14:30:00Z", result)
	})

	t.Run("duration constructor", func(t *testing.T) {
		schema := ISO.Duration()
		require.NotNil(t, schema)

		result, err := schema.Parse("P1Y2M3DT4H5M6S")
		require.NoError(t, err)
		assert.Equal(t, "P1Y2M3DT4H5M6S", result)
	})

	t.Run("smart type inference", func(t *testing.T) {
		schema := ISO.Date()

		// String input returns string
		result, err := schema.Parse("2023-01-01")
		require.NoError(t, err)
		assert.IsType(t, "", result)
		assert.Equal(t, "2023-01-01", result)

		// Pointer input returns same pointer
		dateStr := "2023-01-01"
		result2, err := schema.Parse(&dateStr)
		require.NoError(t, err)
		assert.IsType(t, (*string)(nil), result2)
		assert.Equal(t, &dateStr, result2)
	})

	t.Run("nilable modifier", func(t *testing.T) {
		schema := ISO.Date().Nilable()

		// nil input should succeed, return nil pointer
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Valid input keeps type inference
		result2, err := schema.Parse("2023-01-01")
		require.NoError(t, err)
		assert.Equal(t, "2023-01-01", result2)
	})
}

// =============================================================================
// 2. Coerce (type coercion)
// =============================================================================

func TestISOCoercion(t *testing.T) {
	t.Run("date coercion from time.Time", func(t *testing.T) {
		schema := ISO.Coerce().Date()

		testTime := time.Date(2023, 6, 15, 14, 30, 0, 0, time.UTC)
		result, err := schema.Parse(testTime)
		require.NoError(t, err)
		assert.Equal(t, "2023-06-15", result)

		// Test with *time.Time
		result, err = schema.Parse(&testTime)
		require.NoError(t, err)
		assert.Equal(t, "2023-06-15", result)
	})

	t.Run("date coercion from various formats", func(t *testing.T) {
		schema := ISO.Coerce().Date()

		testCases := []struct {
			input    any
			expected string
		}{
			{"2023-01-01", "2023-01-01"},           // already ISO format
			{"2023-01-01T10:30:00Z", "2023-01-01"}, // ISO datetime
			{"2023-01-01 10:30:00", "2023-01-01"},  // space-separated datetime
			{"01/01/2023", "2023-01-01"},           // US format
		}

		for _, tc := range testCases {
			result, err := schema.Parse(tc.input)
			require.NoError(t, err, "Should coerce %v to %s", tc.input, tc.expected)
			assert.Equal(t, tc.expected, result)
		}
	})

	t.Run("coercion failure fallback", func(t *testing.T) {
		schema := ISO.Coerce().Date()

		invalidInputs := []any{
			123,
			true,
			[]string{"2023-01-01"},
			map[string]string{"date": "2023-01-01"},
		}

		for _, input := range invalidInputs {
			_, err := schema.Parse(input)
			assert.Error(t, err, "Should fail to coerce %v", input)
		}
	})

	t.Run("coerce with validation", func(t *testing.T) {
		schema := ISO.Coerce().Date().Min("2023-01-01")

		// Should coerce and then validate
		testTime := time.Date(2023, 6, 15, 14, 30, 0, 0, time.UTC)
		result, err := schema.Parse(testTime)
		require.NoError(t, err)
		assert.Equal(t, "2023-06-15", result)

		// Should coerce but fail validation
		oldTime := time.Date(2022, 6, 15, 14, 30, 0, 0, time.UTC)
		_, err = schema.Parse(oldTime)
		assert.Error(t, err)
	})
}

// =============================================================================
// 3. Validation methods
// =============================================================================

func TestISOValidations(t *testing.T) {
	t.Run("date validation", func(t *testing.T) {
		schema := ISO.Date()

		validDates := []string{
			"1970-01-01", "2022-01-31", "2022-03-31", "2022-04-30",
			"2022-05-31", "2022-06-30", "2022-07-31", "2022-08-31",
			"2022-09-30", "2022-10-31", "2022-11-30", "2022-12-31",
			"2000-02-29", "2400-02-29", // leap years
		}
		for _, date := range validDates {
			result, err := schema.Parse(date)
			require.NoError(t, err, "Date %s should be valid", date)
			assert.Equal(t, date, result)
		}

		invalidDates := []string{
			"", "foo", "200-01-01", "20000-01-01", "2000-0-01",
			"2000-011-01", "2000-01-0", "2000-01-011", "2000/01/01",
			"01-01-2022", "01/01/2022", "2000-01-01 00:00:00Z",
			"2020-10-14T17:42:29+00:00", "2020-10-14T17:42:29Z",
			"2020-10-14T17:42:29", "2020-10-14T17:42:29.123Z",
			"2000-00-12", "2000-12-00", "2000-01-32", "2000-13-01",
			"2000-21-01", "2000-02-30", "2000-02-31", "2000-04-31",
			"2000-06-31", "2000-09-31", "2000-11-31",
			"2022-02-29", "2100-02-29", "2200-02-29", "2300-02-29", "2500-02-29", // invalid leap years
		}
		for _, date := range invalidDates {
			_, err := schema.Parse(date)
			assert.Error(t, err, "Date %s should be invalid", date)
		}
	})

	t.Run("time validation", func(t *testing.T) {
		schema := ISO.Time()

		validTimes := []string{
			"00:00:00", "23:00:00", "00:59:00", "00:00:59",
			"23:59:59", "09:52:31", "23:59:59.9999999", "00:00",
		}
		for _, time := range validTimes {
			result, err := schema.Parse(time)
			require.NoError(t, err, "Time %s should be valid", time)
			assert.Equal(t, time, result)
		}

		invalidTimes := []string{
			"", "foo", "00:00:00Z", "0:00:00", "00:0:00",
			"00:00:0", "00:00:00.000+00:00", "24:00:00",
			"00:60:00", "00:00:60", "24:60:60",
		}
		for _, time := range invalidTimes {
			_, err := schema.Parse(time)
			assert.Error(t, err, "Time %s should be invalid", time)
		}
	})

	t.Run("datetime validation", func(t *testing.T) {
		schema := ISO.DateTime()

		validDatetimes := []string{
			"1970-01-01T00:00:00.000Z", "2022-10-13T09:52:31.816Z",
			"2022-10-13T09:52:31.8162314Z", "1970-01-01T00:00:00Z",
			"2022-10-13T09:52:31Z",
		}
		for _, datetime := range validDatetimes {
			result, err := schema.Parse(datetime)
			require.NoError(t, err, "Datetime %s should be valid", datetime)
			assert.Equal(t, datetime, result)
		}

		invalidDatetimes := []string{
			"", "foo", "2020-10-14", "T18:45:12.123",
			"2020-10-14T17:42:29+00:00", // This should fail for basic datetime
		}
		for _, datetime := range invalidDatetimes {
			_, err := schema.Parse(datetime)
			assert.Error(t, err, "Datetime %s should be invalid", datetime)
		}
	})

	t.Run("datetime with precision", func(t *testing.T) {
		// Precision -1 (no milliseconds)
		schemaNeg1 := ISO.DateTime(ISODateTimeOptions{Precision: intPtr(-1), Offset: true, Local: true})
		validNeg1 := []string{
			"1970-01-01T00:00Z", "2022-10-13T09:52Z",
			"2022-10-13T09:52+02:00", "2022-10-13T09:52",
		}
		for _, datetime := range validNeg1 {
			result, err := schemaNeg1.Parse(datetime)
			require.NoError(t, err, "Datetime %s should be valid for precision -1", datetime)
			assert.Equal(t, datetime, result)
		}

		invalidNeg1 := []string{
			"tuna", "2022-10-13T09:52+02", "1970-01-01T00:00:00.000Z",
			"1970-01-01T00:00:00.Z", "2022-10-13T09:52:31.816Z",
		}
		for _, datetime := range invalidNeg1 {
			_, err := schemaNeg1.Parse(datetime)
			assert.Error(t, err, "Datetime %s should be invalid for precision -1", datetime)
		}

		// Precision 0
		schema0 := ISO.DateTime(ISODateTimeOptions{Precision: intPtr(0)})
		valid0 := []string{
			"1970-01-01T00:00:00Z", "2022-10-13T09:52:31Z",
		}
		for _, datetime := range valid0 {
			result, err := schema0.Parse(datetime)
			require.NoError(t, err, "Datetime %s should be valid for precision 0", datetime)
			assert.Equal(t, datetime, result)
		}

		invalid0 := []string{
			"tuna", "1970-01-01T00:00:00.000Z", "1970-01-01T00:00:00.Z",
			"2022-10-13T09:52:31.816Z",
		}
		for _, datetime := range invalid0 {
			_, err := schema0.Parse(datetime)
			assert.Error(t, err, "Datetime %s should be invalid for precision 0", datetime)
		}

		// Precision 3
		schema3 := ISO.DateTime(ISODateTimeOptions{Precision: intPtr(3)})
		valid3 := []string{
			"1970-01-01T00:00:00.000Z", "2022-10-13T09:52:31.123Z",
		}
		for _, datetime := range valid3 {
			result, err := schema3.Parse(datetime)
			require.NoError(t, err, "Datetime %s should be valid for precision 3", datetime)
			assert.Equal(t, datetime, result)
		}

		invalid3 := []string{
			"tuna", "1970-01-01T00:00:00.1Z", "1970-01-01T00:00:00.12Z",
			"2022-10-13T09:52:31Z",
		}
		for _, datetime := range invalid3 {
			_, err := schema3.Parse(datetime)
			assert.Error(t, err, "Datetime %s should be invalid for precision 3", datetime)
		}
	})

	t.Run("datetime with offset", func(t *testing.T) {
		schema := ISO.DateTime(ISODateTimeOptions{Offset: true})

		validDatetimes := []string{
			"1970-01-01T00:00:00.000Z", "2022-10-13T09:52:31.816234134Z",
			"1970-01-01T00:00:00Z", "2022-10-13T09:52:31.4Z",
			"2020-10-14T17:42:29+00:00", "2020-10-14T17:42:29+03:15",
		}
		for _, datetime := range validDatetimes {
			result, err := schema.Parse(datetime)
			require.NoError(t, err, "Datetime %s should be valid with offset", datetime)
			assert.Equal(t, datetime, result)
		}

		invalidDatetimes := []string{
			"2020-10-14T17:42:29+0315", "2020-10-14T17:42:29+03",
			"tuna", "2022-10-13T09:52:31.Z",
		}
		for _, datetime := range invalidDatetimes {
			_, err := schema.Parse(datetime)
			assert.Error(t, err, "Datetime %s should be invalid with offset", datetime)
		}
	})

	t.Run("datetime with local option", func(t *testing.T) {
		schema := ISO.DateTime(ISODateTimeOptions{Local: true})

		validDatetimes := []string{
			"1970-01-01T00:00", "1970-01-01T00:00:00",
			"2022-10-13T09:52:31.816", "1970-01-01T00:00:00.000",
		}
		for _, datetime := range validDatetimes {
			result, err := schema.Parse(datetime)
			require.NoError(t, err, "Datetime %s should be valid with local", datetime)
			assert.Equal(t, datetime, result)
		}

		invalidDatetimes := []string{
			"1970-01-01T00", "2022-10-13T09:52:31+00:00",
			"2022-10-13 09:52:31", "2022-10-13T24:52:31",
			"2022-10-13T24:52", "2022-10-13T24:52Z",
		}
		for _, datetime := range invalidDatetimes {
			_, err := schema.Parse(datetime)
			assert.Error(t, err, "Datetime %s should be invalid with local", datetime)
		}
	})

	t.Run("datetime with local and offset", func(t *testing.T) {
		schema := ISO.DateTime(ISODateTimeOptions{Local: true, Offset: true})

		validDatetimes := []string{
			"2022-10-13T12:52:00", "2022-10-13T12:52:00Z",
			"2022-10-13T12:52Z", "2022-10-13T12:52",
			"2022-10-13T12:52+02:00",
		}
		for _, datetime := range validDatetimes {
			result, err := schema.Parse(datetime)
			require.NoError(t, err, "Datetime %s should be valid with local+offset", datetime)
			assert.Equal(t, datetime, result)
		}

		invalidDatetimes := []string{
			"2022-10-13T12:52:00+02",
		}
		for _, datetime := range invalidDatetimes {
			_, err := schema.Parse(datetime)
			assert.Error(t, err, "Datetime %s should be invalid with local+offset", datetime)
		}
	})

	t.Run("duration validation", func(t *testing.T) {
		schema := ISO.Duration()

		validDurations := []string{
			"P3Y6M4DT12H30M5S", "P2Y9M3DT12H31M8.001S",
			"PT0,001S", "PT12H30M5S", "P1Y", "P2MT30M",
			"PT6H", "P5W",
		}
		for _, duration := range validDurations {
			result, err := schema.Parse(duration)
			require.NoError(t, err, "Duration %s should be valid", duration)
			assert.Equal(t, duration, result)
		}

		invalidDurations := []string{
			"foo bar", "", " ", "P", "PT", "P1Y2MT", "T1H",
			"P0.5Y1D", "P0,5Y6M", "P1YT", "P-2M-1D",
			"P-5DT-10H", "P1W2D", "-P1D",
		}
		for _, duration := range invalidDurations {
			_, err := schema.Parse(duration)
			assert.Error(t, err, "Duration %s should be invalid", duration)
		}
	})

	t.Run("time with precision", func(t *testing.T) {
		schema2 := ISO.Time(ISOTimeOptions{Precision: intPtr(2)})

		validTimes := []string{
			"00:00:00.00", "09:52:31.12", "23:59:59.99",
		}
		for _, time := range validTimes {
			result, err := schema2.Parse(time)
			require.NoError(t, err, "Time %s should be valid for precision 2", time)
			assert.Equal(t, time, result)
		}

		invalidTimes := []string{
			"", "foo", "00:00:00", "00:00:00.00Z",
			"00:00:00.0", "00:00:00.000", "00:00:00.00+00:00",
		}
		for _, time := range invalidTimes {
			_, err := schema2.Parse(time)
			assert.Error(t, err, "Time %s should be invalid for precision 2", time)
		}
	})

	t.Run("min max date validation", func(t *testing.T) {
		benchmarkDate := "2022-11-05"
		beforeDate := "2022-11-04"
		afterDate := "2022-11-06"

		minCheck := ISO.Date().Min(benchmarkDate)
		maxCheck := ISO.Date().Max(benchmarkDate)

		// Passing validations
		result, err := minCheck.Parse(benchmarkDate)
		require.NoError(t, err)
		assert.Equal(t, benchmarkDate, result)

		result, err = minCheck.Parse(afterDate)
		require.NoError(t, err)
		assert.Equal(t, afterDate, result)

		result, err = maxCheck.Parse(benchmarkDate)
		require.NoError(t, err)
		assert.Equal(t, benchmarkDate, result)

		result, err = maxCheck.Parse(beforeDate)
		require.NoError(t, err)
		assert.Equal(t, beforeDate, result)

		// Failing validations
		_, err = minCheck.Parse(beforeDate)
		assert.Error(t, err)

		_, err = maxCheck.Parse(afterDate)
		assert.Error(t, err)

		// Test min max getters
		assert.Equal(t, benchmarkDate, minCheck.MinDate())
		newAfterDate := "2022-11-07"
		newMinCheck := minCheck.Min(newAfterDate)
		assert.Equal(t, newAfterDate, newMinCheck.MinDate())

		assert.Equal(t, benchmarkDate, maxCheck.MaxDate())
		newBeforeDate := "2022-11-03"
		newMaxCheck := maxCheck.Max(newBeforeDate)
		assert.Equal(t, newBeforeDate, newMaxCheck.MaxDate())
	})

	t.Run("format validation runs before min/max validation", func(t *testing.T) {
		// This test demonstrates that ISO.Date() includes format validation
		// AND min/max validation, and format validation happens first
		schema := ISO.Date().Min("2023-01-01")

		tests := []struct {
			name          string
			input         string
			expectedError string
			description   string
		}{
			{
				"valid format and valid date",
				"2023-06-15",
				"",
				"Should pass both format and min validation",
			},
			{
				"valid format but too early",
				"2022-12-31",
				"too_small",
				"Should pass format validation but fail min validation",
			},
			{
				"invalid format with slashes",
				"2023/06/15",
				"invalid_format",
				"Should fail format validation before min validation is checked",
			},
			{
				"invalid format but lexicographically after minimum",
				"2023/12/31",
				"invalid_format",
				"Should fail format validation even though '2023/12/31' > '2023-01-01' as string",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := schema.Parse(tt.input)

				if tt.expectedError == "" {
					assert.NoError(t, err, tt.description)
				} else {
					assert.Error(t, err, tt.description)

					var zodErr *issues.ZodError
					if assert.True(t, issues.IsZodError(err, &zodErr)) {
						assert.Len(t, zodErr.Issues, 1)
						assert.Equal(t, tt.expectedError, string(zodErr.Issues[0].Code), tt.description)
					}
				}
			})
		}
	})
}

// =============================================================================
// 4. Modifiers and wrappers
// =============================================================================

func TestISOModifiers(t *testing.T) {
	t.Run("nilable modifier", func(t *testing.T) {
		schema := ISO.Date().Nilable()

		// Valid date
		result, err := schema.Parse("2023-01-01")
		require.NoError(t, err)
		assert.Equal(t, "2023-01-01", result)

		// Nil value should be allowed
		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Nil pointer should be allowed
		var nilPtr *string
		result, err = schema.Parse(nilPtr)
		require.NoError(t, err)
		assert.Equal(t, (*string)(nil), result)
	})

	t.Run("optional modifier", func(t *testing.T) {
		schema := Optional(ISO.Time())

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		result, err = schema.Parse("14:30:00")
		require.NoError(t, err)
		assert.Equal(t, "14:30:00", result)
	})

	t.Run("must parse", func(t *testing.T) {
		schema := ISO.Date()

		// Valid input should not panic
		result := schema.MustParse("2023-01-01")
		assert.Equal(t, "2023-01-01", result)

		// Invalid input should panic
		assert.Panics(t, func() {
			schema.MustParse("invalid-date")
		})
	})
}

// =============================================================================
// 5. Chaining and method composition
// =============================================================================

func TestISOChaining(t *testing.T) {
	t.Run("min and max combined", func(t *testing.T) {
		schema := ISO.Date().Min("2023-01-01").Max("2023-12-31")

		// Valid dates (within range)
		validDates := []string{
			"2023-01-01", "2023-06-15", "2023-12-31",
		}
		for _, date := range validDates {
			result, err := schema.Parse(date)
			require.NoError(t, err, "Date %s should be valid (2023-01-01 to 2023-12-31)", date)
			assert.Equal(t, date, result)
		}

		// Invalid dates (outside range)
		invalidDates := []string{
			"2022-12-31", "2024-01-01",
		}
		for _, date := range invalidDates {
			_, err := schema.Parse(date)
			assert.Error(t, err, "Date %s should be invalid (outside 2023-01-01 to 2023-12-31)", date)
		}
	})

	t.Run("chaining with nilable", func(t *testing.T) {
		schema := ISO.Date().Min("2023-01-01").Nilable()

		// Valid date
		result, err := schema.Parse("2023-06-15")
		require.NoError(t, err)
		assert.Equal(t, "2023-06-15", result)

		// Nil should be allowed
		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Invalid date should fail
		_, err = schema.Parse("2022-01-01")
		assert.Error(t, err)
	})
}

// =============================================================================
// 6. Transform/Pipe
// =============================================================================

func TestISOTransform(t *testing.T) {
	t.Run("transform to time.Time", func(t *testing.T) {
		schema := ISO.Date().TransformAny(func(val any, ctx *core.RefinementContext) (any, error) {
			if dateStr, ok := val.(string); ok {
				return time.Parse("2006-01-02", dateStr)
			}
			return nil, assert.AnError
		})

		result, err := schema.Parse("2023-06-15")
		require.NoError(t, err)

		parsedTime, ok := result.(time.Time)
		require.True(t, ok)
		assert.Equal(t, 2023, parsedTime.Year())
		assert.Equal(t, time.June, parsedTime.Month())
		assert.Equal(t, 15, parsedTime.Day())
	})

	t.Run("pipe to other schema", func(t *testing.T) {
		// Parse ISO date, then transform to day of week
		schema := ISO.Date().Pipe(
			String().TransformAny(func(val any, ctx *core.RefinementContext) (any, error) {
				if dateStr, ok := val.(string); ok {
					if date, err := time.Parse("2006-01-02", dateStr); err == nil {
						return date.Weekday().String(), nil
					}
				}
				return nil, assert.AnError
			}),
		)

		result, err := schema.Parse("2023-06-12") // Monday
		require.NoError(t, err)
		assert.Equal(t, "Monday", result)
	})
}

// =============================================================================
// 7. Refine
// =============================================================================

func TestISORefine(t *testing.T) {
	t.Run("refine with custom validation", func(t *testing.T) {
		// Only allow weekdays (Monday-Friday)
		schema := ISO.Date().RefineAny(func(val any) bool {
			if dateStr, ok := val.(string); ok {
				if date, err := time.Parse("2006-01-02", dateStr); err == nil {
					weekday := date.Weekday()
					return weekday >= time.Monday && weekday <= time.Friday
				}
			}
			return false
		}, core.SchemaParams{
			Error: "Date must be a weekday (Monday-Friday)",
		})

		// Test weekday (should pass)
		result, err := schema.Parse("2023-06-12") // Monday
		require.NoError(t, err)
		assert.Equal(t, "2023-06-12", result)

		// Test weekend (should fail)
		_, err = schema.Parse("2023-06-11") // Sunday
		assert.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.Contains(t, zodErr.Issues[0].Message, "weekday")
	})

	t.Run("refine with min validation", func(t *testing.T) {
		schema := ISO.Date().Min("2023-01-01").RefineAny(func(val any) bool {
			if dateStr, ok := val.(string); ok {
				return dateStr != "2023-02-29" // Reject this specific invalid date
			}
			return true
		}, core.SchemaParams{
			Error: "This specific date is not allowed",
		})

		// Valid date should pass
		result, err := schema.Parse("2023-06-15")
		require.NoError(t, err)
		assert.Equal(t, "2023-06-15", result)

		// Date before min should fail min validation
		_, err = schema.Parse("2022-12-31")
		assert.Error(t, err)

		// Refined date should fail refine validation
		_, err = schema.Parse("2023-02-29")
		assert.Error(t, err)
	})
}

// =============================================================================
// 8. Error handling
// =============================================================================

func TestISOErrorHandling(t *testing.T) {
	t.Run("custom error messages", func(t *testing.T) {
		dateSchema := ISO.Date(ISODateOptions{
			Error: "Please provide a valid date in YYYY-MM-DD format",
		})
		_, err := dateSchema.Parse("invalid-date")
		require.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.Contains(t, zodErr.Issues[0].Message, "Please provide a valid date in YYYY-MM-DD format")

		timeSchema := ISO.Time(ISOTimeOptions{Error: "core.Custom time validation error"})
		_, err = timeSchema.Parse("invalid-time")
		assert.Error(t, err)

		datetimeSchema := ISO.DateTime(ISODateTimeOptions{Error: "core.Custom datetime validation error"})
		_, err = datetimeSchema.Parse("invalid-datetime")
		assert.Error(t, err)

		durationSchema := ISO.Duration(ISODurationOptions{Error: "core.Custom duration validation error"})
		_, err = durationSchema.Parse("invalid-duration")
		assert.Error(t, err)
	})

	t.Run("error structure for invalid type", func(t *testing.T) {
		schema := ISO.Date()
		_, err := schema.Parse(123)

		require.Error(t, err)
		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))

		assert.Len(t, zodErr.Issues, 1)
		issue := zodErr.Issues[0]
		assert.Equal(t, core.InvalidType, issue.Code)
		assert.Equal(t, "string", issue.Expected)
		assert.Equal(t, "number", issue.Received)
	})

	t.Run("error structure for invalid format", func(t *testing.T) {
		schema := ISO.Date()
		_, err := schema.Parse("invalid-date")

		require.Error(t, err)
		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))

		assert.Len(t, zodErr.Issues, 1)
		issue := zodErr.Issues[0]
		assert.Contains(t, issue.Message, "Invalid ISO date format")
	})

	t.Run("error structure for min validation", func(t *testing.T) {
		schema := ISO.Date().Min("2023-01-01")
		_, err := schema.Parse("2022-12-31")

		require.Error(t, err)
		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))

		assert.Len(t, zodErr.Issues, 1)
		assert.Contains(t, zodErr.Issues[0].Message, "2023-01-01")
	})

	t.Run("custom error for min validation", func(t *testing.T) {
		schema := ISO.Date().Min("1900-01-01", core.SchemaParams{
			Error: "Too old! Date must be after 1900-01-01",
		})

		_, err := schema.Parse("1899-12-31")
		require.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.Contains(t, zodErr.Issues[0].Message, "Too old!")
	})

	t.Run("custom error for max validation", func(t *testing.T) {
		today := time.Now().Format("2006-01-02")
		schema := ISO.Date().Max(today, core.SchemaParams{
			Error: "Too young! Date cannot be in the future",
		})

		// Test with future date
		futureDate := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
		_, err := schema.Parse(futureDate)
		require.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.Contains(t, zodErr.Issues[0].Message, "Too young!")
	})
}

// =============================================================================
// 9. Edge and mutual exclusion cases
// =============================================================================

func TestISOEdgeCases(t *testing.T) {
	t.Run("leap year dates", func(t *testing.T) {
		schema := ISO.Date()

		// Valid leap year date
		result, err := schema.Parse("2024-02-29")
		require.NoError(t, err)
		assert.Equal(t, "2024-02-29", result)

		// Invalid leap year date (2023 is not a leap year)
		_, err = schema.Parse("2023-02-29")
		assert.Error(t, err)
	})

	t.Run("boundary dates", func(t *testing.T) {
		schema := ISO.Date().Min("2023-01-01").Max("2023-01-01")

		// Exactly on boundary (should pass)
		result, err := schema.Parse("2023-01-01")
		require.NoError(t, err)
		assert.Equal(t, "2023-01-01", result)

		// Just outside boundaries (should fail)
		_, err = schema.Parse("2022-12-31")
		assert.Error(t, err)

		_, err = schema.Parse("2023-01-02")
		assert.Error(t, err)
	})

	t.Run("empty string handling", func(t *testing.T) {
		schemas := []core.ZodType[any, any]{
			ISO.Date(), ISO.Time(), ISO.DateTime(), ISO.Duration(),
		}

		for _, schema := range schemas {
			_, err := schema.Parse("")
			assert.Error(t, err, "Empty string should be invalid")
		}
	})

	t.Run("whitespace handling", func(t *testing.T) {
		schema := ISO.Date()

		// Leading/trailing whitespace should fail
		_, err := schema.Parse(" 2023-01-01 ")
		assert.Error(t, err)

		_, err = schema.Parse("2023-01-01\n")
		assert.Error(t, err)
	})

	t.Run("non-string input types", func(t *testing.T) {
		schemas := []core.ZodType[any, any]{
			ISO.Date(), ISO.Time(), ISO.DateTime(), ISO.Duration(),
		}

		invalidInputs := []any{
			123, true, []string{"2023-12-25"},
			map[string]any{"date": "2023-12-25"}, nil,
		}

		for _, schema := range schemas {
			for _, input := range invalidInputs {
				_, err := schema.Parse(input)
				assert.Error(t, err, "Input %v should be rejected", input)
			}
		}
	})

	t.Run("datetime offset normalization", func(t *testing.T) {
		schema := ISO.DateTime(ISODateTimeOptions{Offset: true})

		// These should fail - invalid offset formats
		invalidOffsets := []string{
			"2020-10-14T17:42:29+02",
			"2020-10-14T17:42:29+0200",
		}

		for _, datetime := range invalidOffsets {
			_, err := schema.Parse(datetime)
			assert.Error(t, err, "Invalid offset format %s should fail", datetime)
		}

		// This should pass - valid offset format
		result, err := schema.Parse("2020-10-14T17:42:29+02:00")
		require.NoError(t, err)
		assert.Equal(t, "2020-10-14T17:42:29+02:00", result)
	})
}

// =============================================================================
// 10. Default and Prefault tests
// =============================================================================

func TestISODefaultAndPrefault(t *testing.T) {
	t.Run("integration with other types", func(t *testing.T) {
		// ISO in object schema
		schema := Object(core.ObjectSchema{
			"date":     ISO.Date(),
			"time":     ISO.Time(),
			"datetime": ISO.DateTime(),
			"duration": ISO.Duration(),
		})

		validData := map[string]any{
			"date":     "2023-12-25",
			"time":     "14:30:00",
			"datetime": "2023-12-25T14:30:00Z",
			"duration": "P1Y2M3DT4H5M6S",
		}

		result, err := schema.Parse(validData)
		require.NoError(t, err)
		assert.Equal(t, validData, result)
	})

	t.Run("ISO in array schema", func(t *testing.T) {
		schema := Slice(ISO.Date())

		validDates := []any{"2023-01-01", "2023-12-31", "2024-02-29"}
		result, err := schema.Parse(validDates)
		require.NoError(t, err)
		assert.Equal(t, validDates, result)
	})

	t.Run("ISO with union", func(t *testing.T) {
		schema := Union([]core.ZodType[any, any]{ISO.Date(), ISO.Time()})

		result, err := schema.Parse("2023-12-25")
		require.NoError(t, err)
		assert.Equal(t, "2023-12-25", result)

		result, err = schema.Parse("14:30:00")
		require.NoError(t, err)
		assert.Equal(t, "14:30:00", result)
	})

	t.Run("ISO with optional", func(t *testing.T) {
		schema := Optional(ISO.Date())

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		result, err = schema.Parse("2023-12-25")
		require.NoError(t, err)
		assert.Equal(t, "2023-12-25", result)
	})
}

// Helper function for creating int pointers
func intPtr(i int) *int {
	return &i
}
