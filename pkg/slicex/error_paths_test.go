package slicex

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

func TestToTyped_ReturnsElementConversionError(t *testing.T) {
	t.Parallel()

	_, err := ToTyped[int]([]any{"not an int"})
	require.ErrorIs(t, err, ErrCannotConvertElement)
}

func TestMutationFunctions_ReturnSentinelErrorsForInvalidInputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		operation func() error
		want      error
	}{
		{
			name: "merge invalid first input",
			operation: func() error {
				_, err := Merge(42, []int{1})
				return err
			},
			want: ErrCannotConvertFirst,
		},
		{
			name: "merge invalid second input",
			operation: func() error {
				_, err := Merge([]int{1}, 42)
				return err
			},
			want: ErrCannotConvertSecond,
		},
		{
			name: "append invalid input",
			operation: func() error {
				_, err := Append(42, 1)
				return err
			},
			want: ErrCannotConvert,
		},
		{
			name: "prepend invalid input",
			operation: func() error {
				_, err := Prepend(42, 1)
				return err
			},
			want: ErrCannotConvert,
		},
		{
			name: "reverse invalid input",
			operation: func() error {
				_, err := Reverse(42)
				return err
			},
			want: ErrCannotConvert,
		},
		{
			name: "unique invalid input",
			operation: func() error {
				_, err := Unique(42)
				return err
			},
			want: ErrCannotConvert,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.operation()
			require.ErrorIs(t, err, tt.want)
			require.ErrorIs(t, err, ErrNotCollection)
		})
	}
}

func TestUnique_RemovesStructurallyEqualNonComparableValues(t *testing.T) {
	t.Parallel()

	input := []any{
		map[string]any{"name": "a"},
		map[string]any{"name": "a"},
		[]any{1, "b"},
		[]any{1, "b"},
		[]any{"b", 1},
	}

	got, err := Unique(input)
	require.NoError(t, err)

	want := []any{
		map[string]any{"name": "a"},
		[]any{1, "b"},
		[]any{"b", 1},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Unique() mismatch (-want +got):\n%s", diff)
	}
}
