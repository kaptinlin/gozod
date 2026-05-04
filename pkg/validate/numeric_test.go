package validate_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kaptinlin/gozod/pkg/validate"
)

func TestNumericComparisons(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  bool
		want bool
	}{
		{name: "less than", got: validate.Lt(2, 3), want: true},
		{name: "less than rejects equal", got: validate.Lt(3, 3), want: false},
		{name: "less than or equal accepts equal", got: validate.Lte(3, 3), want: true},
		{name: "greater than", got: validate.Gt(4, 3), want: true},
		{name: "greater than rejects equal", got: validate.Gt(3, 3), want: false},
		{name: "greater than or equal accepts equal", got: validate.Gte(3, 3), want: true},
		{name: "non numeric value", got: validate.Gt("4", 3), want: false},
		{name: "non numeric limit", got: validate.Lt(2, "3"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.got)
		})
	}
}

func TestNumericSignPredicates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  bool
		want bool
	}{
		{name: "positive", got: validate.Positive(1), want: true},
		{name: "positive rejects zero", got: validate.Positive(0), want: false},
		{name: "negative", got: validate.Negative(-1.5), want: true},
		{name: "negative rejects zero", got: validate.Negative(0), want: false},
		{name: "non positive accepts zero", got: validate.NonPositive(0), want: true},
		{name: "non positive rejects positive", got: validate.NonPositive(1), want: false},
		{name: "non negative accepts zero", got: validate.NonNegative(0), want: true},
		{name: "non negative rejects negative", got: validate.NonNegative(-1), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.got)
		})
	}
}

func TestMultipleOf(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   any
		divisor any
		want    bool
	}{
		{name: "integer multiple", value: 10, divisor: 5, want: true},
		{name: "integer non multiple", value: 7, divisor: 3, want: false},
		{name: "negative multiple", value: -12, divisor: 4, want: true},
		{name: "zero divisor", value: 10, divisor: 0, want: false},
		{name: "non numeric value", value: "10", divisor: 5, want: false},
		{name: "non numeric divisor", value: 10, divisor: "5", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, validate.MultipleOf(tt.value, tt.divisor))
		})
	}
}
