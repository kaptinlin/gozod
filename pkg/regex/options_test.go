package regex

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTimeOptions(t *testing.T) {
	t.Parallel()

	zeroPrecision := 0
	twoDigits := 2
	minutePrecision := -1

	tests := []struct {
		name      string
		precision *int
		input     string
		want      bool
	}{
		{name: "default accepts minutes", input: "12:30", want: true},
		{name: "default accepts fractional seconds", input: "12:30:45.123", want: true},
		{name: "second precision requires seconds", precision: &zeroPrecision, input: "12:30", want: false},
		{name: "second precision accepts seconds", precision: &zeroPrecision, input: "12:30:45", want: true},
		{name: "fraction precision accepts exact digits", precision: &twoDigits, input: "12:30:45.12", want: true},
		{name: "fraction precision rejects different digits", precision: &twoDigits, input: "12:30:45.1", want: false},
		{name: "minute precision accepts minutes", precision: &minutePrecision, input: "12:30", want: true},
		{name: "minute precision rejects seconds", precision: &minutePrecision, input: "12:30:45", want: false},
		{name: "rejects invalid hour", input: "25:00", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			re := Time(TimeOptions{Precision: tt.precision})
			assert.Equal(t, tt.want, re.MatchString(tt.input))
		})
	}
}

func TestDatetimeOptions(t *testing.T) {
	t.Parallel()

	zeroPrecision := 0
	minutePrecision := -1

	tests := []struct {
		name    string
		options DatetimeOptions
		input   string
		want    bool
	}{
		{name: "default requires zulu timezone", input: "2024-02-29T12:30:45Z", want: true},
		{name: "default rejects local datetime", input: "2024-02-29T12:30:45", want: false},
		{name: "default rejects numeric offset", input: "2024-02-29T12:30:45+05:30", want: false},
		{name: "local accepts missing timezone", options: DatetimeOptions{Local: true}, input: "2024-02-29T12:30:45", want: true},
		{name: "offset accepts numeric timezone", options: DatetimeOptions{Offset: true}, input: "2024-02-29T12:30:45+05:30", want: true},
		{name: "second precision accepts seconds", options: DatetimeOptions{Precision: &zeroPrecision}, input: "2024-02-29T12:30:45Z", want: true},
		{name: "second precision rejects fractional seconds", options: DatetimeOptions{Precision: &zeroPrecision}, input: "2024-02-29T12:30:45.1Z", want: false},
		{name: "minute precision accepts minutes", options: DatetimeOptions{Precision: &minutePrecision}, input: "2024-02-29T12:30Z", want: true},
		{name: "minute precision rejects seconds", options: DatetimeOptions{Precision: &minutePrecision}, input: "2024-02-29T12:30:45Z", want: false},
		{name: "rejects invalid date", input: "2023-02-29T12:30:45Z", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			re := Datetime(tt.options)
			assert.Equal(t, tt.want, re.MatchString(tt.input))
		})
	}
}

func TestUUIDForVersion(t *testing.T) {
	t.Parallel()

	assert.True(t, UUIDForVersion(4).MatchString("550e8400-e29b-41d4-a716-446655440000"))
	assert.False(t, UUIDForVersion(4).MatchString("550e8400-e29b-71d4-a716-446655440000"))
	assert.Same(t, UUID, UUIDForVersion(0))
	assert.Same(t, UUID, UUIDForVersion(9))
}

func TestMACUsesDefaultDelimiterForEmptyOption(t *testing.T) {
	t.Parallel()

	assert.Same(t, MAC(":"), MAC(""))
	assert.True(t, MAC("").MatchString("00:1A:2B:3C:4D:5E"))
	assert.False(t, MAC("").MatchString("00-1A-2B-3C-4D-5E"))
}
