package validate_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kaptinlin/gozod/pkg/validate"
)

func TestISODateTime(t *testing.T) {
	t.Parallel()

	precision3 := 3
	minutePrecision := -1

	tests := []struct {
		name  string
		value any
		got   bool
		want  bool
	}{
		{name: "rfc3339", value: "2024-02-29T12:30:45Z", got: validate.ISODateTime("2024-02-29T12:30:45Z"), want: true},
		{name: "rejects local time by default", value: "2024-02-29T12:30:45", got: validate.ISODateTime("2024-02-29T12:30:45"), want: false},
		{name: "rejects non string", value: 123, got: validate.ISODateTime(123), want: false},
		{name: "local option accepts missing timezone", value: "2024-02-29T12:30:45", got: validate.ISODateTimeWithOptions("2024-02-29T12:30:45", validate.ISODateTimeOptions{Local: true}), want: true},
		{name: "offset option accepts numeric offset", value: "2024-02-29T12:30:45+05:30", got: validate.ISODateTimeWithOptions("2024-02-29T12:30:45+05:30", validate.ISODateTimeOptions{Offset: true}), want: true},
		{name: "precision option accepts exact fractional digits", value: "2024-02-29T12:30:45.123Z", got: validate.ISODateTimeWithOptions("2024-02-29T12:30:45.123Z", validate.ISODateTimeOptions{Precision: &precision3}), want: true},
		{name: "precision option rejects different fractional digits", value: "2024-02-29T12:30:45.12Z", got: validate.ISODateTimeWithOptions("2024-02-29T12:30:45.12Z", validate.ISODateTimeOptions{Precision: &precision3}), want: false},
		{name: "minute precision rejects seconds", value: "2024-02-29T12:30:45Z", got: validate.ISODateTimeWithOptions("2024-02-29T12:30:45Z", validate.ISODateTimeOptions{Precision: &minutePrecision}), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.got)
		})
	}
}

func TestISODateAndTime(t *testing.T) {
	t.Parallel()

	secondPrecision := 0
	fractionPrecision := 2
	minutePrecision := -1

	tests := []struct {
		name string
		got  bool
		want bool
	}{
		{name: "valid leap date", got: validate.ISODate("2024-02-29"), want: true},
		{name: "invalid leap date", got: validate.ISODate("2023-02-29"), want: false},
		{name: "date rejects non string", got: validate.ISODate(123), want: false},
		{name: "default time accepts fractional seconds", got: validate.ISOTime("12:30:45.123"), want: true},
		{name: "default time rejects invalid hour", got: validate.ISOTime("25:00"), want: false},
		{name: "time rejects non string", got: validate.ISOTime(123), want: false},
		{name: "second precision accepts seconds", got: validate.ISOTimeWithOptions("12:30:45", validate.ISOTimeOptions{Precision: &secondPrecision}), want: true},
		{name: "second precision rejects fractional seconds", got: validate.ISOTimeWithOptions("12:30:45.1", validate.ISOTimeOptions{Precision: &secondPrecision}), want: false},
		{name: "fraction precision accepts exact digits", got: validate.ISOTimeWithOptions("12:30:45.12", validate.ISOTimeOptions{Precision: &fractionPrecision}), want: true},
		{name: "minute precision rejects seconds", got: validate.ISOTimeWithOptions("12:30:45", validate.ISOTimeOptions{Precision: &minutePrecision}), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.got)
		})
	}
}

func TestISODuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value any
		want  bool
	}{
		{name: "week duration", value: "P1W", want: true},
		{name: "date time duration", value: "P1Y2M3DT4H5M6S", want: true},
		{name: "rejects empty period", value: "P", want: false},
		{name: "rejects empty time", value: "PT", want: false},
		{name: "rejects time marker without time units", value: "P1YT", want: false},
		{name: "rejects non string", value: 123, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, validate.ISODuration(tt.value))
		})
	}
}

func TestJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value any
		want  bool
	}{
		{name: "object string", value: `{"ok": true}`, want: true},
		{name: "byte slice", value: []byte(`[1, 2, 3]`), want: true},
		{name: "numeric value coerces to json number", value: 42, want: true},
		{name: "invalid json string", value: `{"ok":`, want: false},
		{name: "unsupported input", value: map[string]any{"ok": true}, want: false},
		{name: "nil input", value: nil, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, validate.JSON(tt.value))
		})
	}
}

func TestProperty(t *testing.T) {
	t.Parallel()

	obj := map[string]any{"count": 3}
	isPositive := func(v any) bool { return validate.Positive(v) }

	tests := []struct {
		name      string
		obj       any
		key       string
		validator func(any) bool
		want      bool
	}{
		{name: "property passes validator", obj: obj, key: "count", validator: isPositive, want: true},
		{name: "property fails validator", obj: obj, key: "count", validator: func(any) bool { return false }, want: false},
		{name: "missing property", obj: obj, key: "missing", validator: isPositive, want: false},
		{name: "nil validator", obj: obj, key: "count", validator: nil, want: false},
		{name: "non map object", obj: struct{}{}, key: "count", validator: isPositive, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, validate.Property(tt.obj, tt.key, tt.validator))
		})
	}
}
