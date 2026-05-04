package validate_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kaptinlin/gozod/pkg/validate"
)

func TestLengthPredicates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  bool
		want bool
	}{
		{name: "max length accepts shorter string", got: validate.MaxLength("go", 3), want: true},
		{name: "max length rejects longer string", got: validate.MaxLength("gozod", 3), want: false},
		{name: "min length accepts longer slice", got: validate.MinLength([]int{1, 2, 3}, 2), want: true},
		{name: "min length rejects shorter slice", got: validate.MinLength([]int{1}, 2), want: false},
		{name: "exact length accepts array", got: validate.Length([2]int{}, 2), want: true},
		{name: "exact length rejects mismatch", got: validate.Length([2]int{}, 3), want: false},
		{name: "max length rejects unsupported input", got: validate.MaxLength(42, 2), want: false},
		{name: "min length rejects unsupported input", got: validate.MinLength(42, 2), want: false},
		{name: "length rejects unsupported input", got: validate.Length(42, 2), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.got)
		})
	}
}

func TestSizePredicates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  bool
		want bool
	}{
		{name: "max size accepts smaller map", got: validate.MaxSize(map[string]any{"a": 1}, 2), want: true},
		{name: "max size rejects larger map", got: validate.MaxSize(map[string]any{"a": 1, "b": 2}, 1), want: false},
		{name: "min size accepts larger generic map", got: validate.MinSize(map[any]any{"a": 1, "b": 2}, 2), want: true},
		{name: "min size rejects smaller generic map", got: validate.MinSize(map[any]any{"a": 1}, 2), want: false},
		{name: "exact size accepts buffered channel length", got: validate.Size(bufferedChannel(1, 2), 2), want: true},
		{name: "exact size rejects channel mismatch", got: validate.Size(make(chan int, 2), 1), want: false},
		{name: "size falls back to string length", got: validate.Size("abc", 3), want: true},
		{name: "max size rejects unsupported input", got: validate.MaxSize(42, 2), want: false},
		{name: "min size rejects unsupported input", got: validate.MinSize(42, 2), want: false},
		{name: "size rejects unsupported input", got: validate.Size(42, 2), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.got)
		})
	}
}

func bufferedChannel(values ...int) chan int {
	ch := make(chan int, len(values))
	for _, value := range values {
		ch <- value
	}
	return ch
}
