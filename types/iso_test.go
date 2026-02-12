package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestISODateTime_Basic(t *testing.T) {
	s := IsoDateTime()

	t.Run("valid", func(t *testing.T) {
		valid := []string{
			"2023-12-25T15:30:45Z",
			"2023-12-25T15:30:45.123Z",
			"2023-12-25T15:30:45+08:00",
			"2023-01-01T00:00:00Z",
		}
		for _, v := range valid {
			got, err := s.Parse(v)
			require.NoError(t, err, "Parse(%q)", v)
			assert.Equal(t, v, got)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		invalid := []string{
			"invalid-datetime",
			"2023-12-25 15:30:45",
			"2023-12-25T25:30:45Z",
			"2023-13-25T15:30:45Z",
			"",
		}
		for _, v := range invalid {
			_, err := s.Parse(v)
			assert.Error(t, err, "Parse(%q)", v)
		}
	})

	t.Run("wrong_types", func(t *testing.T) {
		for _, v := range []any{123, true, nil, []string{"a"}} {
			_, err := s.Parse(v)
			assert.Error(t, err, "Parse(%v)", v)
		}
	})
}

func TestISODate_Basic(t *testing.T) {
	s := IsoDate()

	t.Run("valid", func(t *testing.T) {
		for _, v := range []string{"2023-12-25", "2023-01-01", "2024-02-29"} {
			got, err := s.Parse(v)
			require.NoError(t, err, "Parse(%q)", v)
			assert.Equal(t, v, got)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		for _, v := range []string{"invalid-date", "25/12/2023", "2023-13-01", "2023-12-32", ""} {
			_, err := s.Parse(v)
			assert.Error(t, err, "Parse(%q)", v)
		}
	})
}

func TestISOTime_Basic(t *testing.T) {
	s := IsoTime()

	t.Run("valid", func(t *testing.T) {
		for _, v := range []string{"15:30", "15:30:45", "15:30:45.123", "00:00:00", "23:59:59"} {
			got, err := s.Parse(v)
			require.NoError(t, err, "Parse(%q)", v)
			assert.Equal(t, v, got)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		for _, v := range []string{"invalid-time", "25:00:00", "12:60:00", "12:30:60", "12:30:00Z", ""} {
			_, err := s.Parse(v)
			assert.Error(t, err, "Parse(%q)", v)
		}
	})
}

func TestISODuration_Basic(t *testing.T) {
	s := IsoDuration()

	t.Run("valid", func(t *testing.T) {
		for _, v := range []string{"P1Y2M3DT4H5M6S", "P1Y", "P1D", "PT1H", "PT1M", "PT1S", "P1W"} {
			got, err := s.Parse(v)
			require.NoError(t, err, "Parse(%q)", v)
			assert.Equal(t, v, got)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		for _, v := range []string{"invalid-duration", "P", "PT", "1Y2M", ""} {
			_, err := s.Parse(v)
			assert.Error(t, err, "Parse(%q)", v)
		}
	})
}

func TestISODateTime_Options(t *testing.T) {
	t.Run("offset", func(t *testing.T) {
		s := IsoDateTime(IsoDatetimeOptions{Offset: true})

		for _, v := range []string{
			"2020-01-01T06:15:00Z",
			"2020-01-01T06:15:00+02:00",
			"2020-01-01T06:15:00-05:00",
		} {
			got, err := s.Parse(v)
			require.NoError(t, err, "Parse(%q)", v)
			assert.Equal(t, v, got)
		}

		_, err := s.Parse("2020-01-01T06:15:00+02")
		assert.Error(t, err, "incomplete offset")
	})

	t.Run("local", func(t *testing.T) {
		s := IsoDateTime(IsoDatetimeOptions{Local: true})

		for _, v := range []string{
			"2020-01-01T06:15:00Z",
			"2020-01-01T06:15:00",
			"2020-01-01T06:15",
		} {
			got, err := s.Parse(v)
			require.NoError(t, err, "Parse(%q)", v)
			assert.Equal(t, v, got)
		}
	})

	t.Run("precision", func(t *testing.T) {
		// Minute precision
		ms := IsoDateTime(IsoDatetimeOptions{Precision: PrecisionMinute})
		got, err := ms.Parse("2020-01-01T06:15Z")
		require.NoError(t, err)
		assert.Equal(t, "2020-01-01T06:15Z", got)

		_, err = ms.Parse("2020-01-01T06:15:00Z")
		assert.Error(t, err, "should reject seconds for minute precision")

		// Second precision
		ss := IsoDateTime(IsoDatetimeOptions{Precision: PrecisionSecond})
		got, err = ss.Parse("2020-01-01T06:15:00Z")
		require.NoError(t, err)
		assert.Equal(t, "2020-01-01T06:15:00Z", got)

		_, err = ss.Parse("2020-01-01T06:15:00.123Z")
		assert.Error(t, err, "should reject milliseconds for second precision")

		// Millisecond precision
		mss := IsoDateTime(IsoDatetimeOptions{Precision: PrecisionMillisecond})
		got, err = mss.Parse("2020-01-01T06:15:00.123Z")
		require.NoError(t, err)
		assert.Equal(t, "2020-01-01T06:15:00.123Z", got)

		_, err = mss.Parse("2020-01-01T06:15:00Z")
		assert.Error(t, err, "should reject format without milliseconds")
	})
}

func TestISOTime_Options(t *testing.T) {
	t.Run("precision", func(t *testing.T) {
		// Minute precision
		ms := IsoTime(IsoTimeOptions{Precision: PrecisionMinute})
		got, err := ms.Parse("15:30")
		require.NoError(t, err)
		assert.Equal(t, "15:30", got)

		_, err = ms.Parse("15:30:00")
		assert.Error(t, err, "should reject seconds")

		// Second precision
		ss := IsoTime(IsoTimeOptions{Precision: PrecisionSecond})
		got, err = ss.Parse("15:30:00")
		require.NoError(t, err)
		assert.Equal(t, "15:30:00", got)

		_, err = ss.Parse("15:30")
		assert.Error(t, err, "should reject minute-only format")

		// Millisecond precision
		mss := IsoTime(IsoTimeOptions{Precision: PrecisionMillisecond})
		got, err = mss.Parse("15:30:00.123")
		require.NoError(t, err)
		assert.Equal(t, "15:30:00.123", got)

		_, err = mss.Parse("15:30:00")
		assert.Error(t, err, "should reject format without milliseconds")
	})
}

func TestISO_Modifiers(t *testing.T) {
	t.Run("optional", func(t *testing.T) {
		s := IsoDateTime().Optional()

		got, err := s.Parse("2023-12-25T15:30:45Z")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "2023-12-25T15:30:45Z", *got)

		got, err = s.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("default", func(t *testing.T) {
		want := "2023-01-01T00:00:00Z"
		s := IsoDateTime().Default(want)

		got, err := s.Parse("2023-12-25T15:30:45Z")
		require.NoError(t, err)
		assert.Equal(t, "2023-12-25T15:30:45Z", got)

		got, err = s.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("default_func", func(t *testing.T) {
		s := IsoDateTime().DefaultFunc(func() string {
			return "2023-01-01T00:00:00Z"
		})

		got, err := s.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "2023-01-01T00:00:00Z", got)
	})
}

func TestISO_DefaultAndPrefault(t *testing.T) {
	t.Run("default_priority_over_prefault", func(t *testing.T) {
		s := IsoDateTime().Default("2023-01-01T00:00:00Z").Prefault("2023-12-31T23:59:59Z")

		got, err := s.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "2023-01-01T00:00:00Z", got)
	})

	t.Run("default_short_circuit", func(t *testing.T) {
		s := IsoDateTime().Min("2023-06-01T00:00:00Z").Default("2023-01-01T00:00:00Z")

		got, err := s.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "2023-01-01T00:00:00Z", got)
	})

	t.Run("prefault_requires_validation", func(t *testing.T) {
		s := IsoDateTime().Min("2023-06-01T00:00:00Z").Prefault("2023-01-01T00:00:00Z")

		_, err := s.Parse(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Too small")
	})

	t.Run("prefault_only_on_nil", func(t *testing.T) {
		s := IsoDateTime().Min("2023-06-01T00:00:00Z").Prefault("2023-12-31T23:59:59Z")

		_, err := s.Parse("2023-01-01T00:00:00Z")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Too small")
	})

	t.Run("default_func_and_prefault_func", func(t *testing.T) {
		defaultCalled := false
		prefaultCalled := false

		s := IsoDateTime().DefaultFunc(func() string {
			defaultCalled = true
			return "2023-01-01T00:00:00Z"
		}).PrefaultFunc(func() string {
			prefaultCalled = true
			return "2023-12-31T23:59:59Z"
		})

		got, err := s.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "2023-01-01T00:00:00Z", got)
		assert.True(t, defaultCalled)
		assert.False(t, prefaultCalled)
	})

	t.Run("prefault_validation_failure", func(t *testing.T) {
		s := IsoDate().Min("2023-06-01").Prefault("2023-01-01")

		_, err := s.Parse(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Too small")
	})
}

func TestISO_RangeValidation(t *testing.T) {
	t.Run("datetime_range", func(t *testing.T) {
		s := IsoDateTime().Min("2024-01-15T10:00:00Z").Max("2024-01-15T18:00:00Z")

		got, err := s.Parse("2024-01-15T14:30:45Z")
		require.NoError(t, err)
		assert.Equal(t, "2024-01-15T14:30:45Z", got)

		_, err = s.Parse("2024-01-15T10:00:00Z")
		assert.NoError(t, err)

		_, err = s.Parse("2024-01-15T18:00:00Z")
		assert.NoError(t, err)

		_, err = s.Parse("2024-01-15T09:59:59Z")
		assert.Error(t, err, "before minimum")

		_, err = s.Parse("2024-01-15T18:00:01Z")
		assert.Error(t, err, "after maximum")
	})

	t.Run("date_range", func(t *testing.T) {
		s := IsoDate().Min("2022-11-05").Max("2022-11-10")

		_, err := s.Parse("2022-11-07")
		assert.NoError(t, err)

		_, err = s.Parse("2022-11-04")
		assert.Error(t, err, "before minimum")

		_, err = s.Parse("2022-11-11")
		assert.Error(t, err, "after maximum")
	})

	t.Run("time_range", func(t *testing.T) {
		s := IsoTime().Min("09:00:00").Max("17:00:00")

		_, err := s.Parse("12:30:45")
		assert.NoError(t, err)

		_, err = s.Parse("08:59:59")
		assert.Error(t, err, "before minimum")

		_, err = s.Parse("17:00:01")
		assert.Error(t, err, "after maximum")
	})
}

func TestISO_TypeSafety(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		got := IsoDateTime().MustParse("2023-12-25T15:30:45Z")
		assert.Equal(t, "2023-12-25T15:30:45Z", got)
	})

	t.Run("optional_returns_pointer", func(t *testing.T) {
		s := IsoDateTime().Optional()
		got := s.MustParse("2023-12-25T15:30:45Z")
		require.NotNil(t, got)
		assert.Equal(t, "2023-12-25T15:30:45Z", *got)

		got = s.MustParse(nil)
		assert.Nil(t, got)
	})

	t.Run("ptr_schema", func(t *testing.T) {
		got, err := IsoDateTimePtr().Parse("2023-12-25T15:30:45Z")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "2023-12-25T15:30:45Z", *got)
	})
}

func TestISO_Integration(t *testing.T) {
	t.Run("real_world_usage", func(t *testing.T) {
		dt := IsoDateTime()
		d := IsoDate()
		tm := IsoTime()
		dur := IsoDuration()

		got, err := dt.Parse("2023-12-25T09:00:00Z")
		require.NoError(t, err)
		assert.Equal(t, "2023-12-25T09:00:00Z", got)

		got, err = d.Parse("2023-12-25")
		require.NoError(t, err)
		assert.Equal(t, "2023-12-25", got)

		got, err = tm.Parse("09:00:00")
		require.NoError(t, err)
		assert.Equal(t, "09:00:00", got)

		got, err = dur.Parse("PT8H")
		require.NoError(t, err)
		assert.Equal(t, "PT8H", got)
	})

	t.Run("combined_options", func(t *testing.T) {
		s := IsoDateTime(IsoDatetimeOptions{
			Precision: PrecisionMillisecond,
			Offset:    true,
			Local:     true,
		})

		for _, v := range []string{
			"2023-12-25T15:30:45.123Z",
			"2023-12-25T15:30:45.123+08:00",
			"2023-12-25T15:30:45.123",
		} {
			got, err := s.Parse(v)
			require.NoError(t, err, "Parse(%q)", v)
			assert.Equal(t, v, got)
		}

		_, err := s.Parse("2023-12-25T15:30:45Z")
		assert.Error(t, err, "should reject format without milliseconds")
	})
}

func TestISO_ErrorHandling(t *testing.T) {
	t.Run("custom_error_message", func(t *testing.T) {
		s := IsoDateTime("Please provide a valid ISO 8601 datetime format")
		_, err := s.Parse("invalid-datetime")
		assert.Error(t, err)
	})

	t.Run("must_parse_panics", func(t *testing.T) {
		assert.Panics(t, func() {
			IsoDateTime().MustParse("invalid-datetime")
		})
	})

	t.Run("empty_string", func(t *testing.T) {
		tests := []struct {
			name string
			fn   func() (string, error)
		}{
			{"datetime", func() (string, error) { return IsoDateTime().Parse("") }},
			{"date", func() (string, error) { return IsoDate().Parse("") }},
			{"time", func() (string, error) { return IsoTime().Parse("") }},
			{"duration", func() (string, error) { return IsoDuration().Parse("") }},
		}
		for _, tc := range tests {
			_, err := tc.fn()
			assert.Error(t, err, "%s should reject empty string", tc.name)
		}
	})
}
