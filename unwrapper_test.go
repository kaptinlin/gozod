package gozod_test

import (
	"testing"

	"github.com/kaptinlin/gozod"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Optional is a test wrapper type
type Optional[T any] struct {
	Value T
	set   bool
}

func (o Optional[T]) Unwrap() (any, bool) {
	return o.Value, o.set
}

func (o Optional[T]) IsSet() bool {
	return o.set
}

func NewOptional[T any](value T) Optional[T] {
	return Optional[T]{Value: value, set: true}
}

func EmptyOptional[T any]() Optional[T] {
	return Optional[T]{set: false}
}

// TestUnwrapperInterface verifies Optional implements Unwrapper
func TestUnwrapperInterface(t *testing.T) {
	opt := NewOptional(42)
	var _ gozod.Unwrapper = opt
}

// TestBasicIntValidation tests if FromStruct correctly validates int fields
func TestBasicIntValidation(t *testing.T) {
	type Config struct {
		Port int `gozod:"min=1000,max=9999"`
	}

	t.Run("valid value", func(t *testing.T) {
		config := Config{Port: 3000}
		schema := gozod.FromStruct[Config]()
		_, err := schema.Parse(config)
		require.NoError(t, err)
	})

	t.Run("invalid value - too small", func(t *testing.T) {
		config := Config{Port: 100}
		schema := gozod.FromStruct[Config]()
		_, err := schema.Parse(config)
		require.Error(t, err, "expected validation error")
	})
}

// TestFromStructWithOptional tests if FromStruct recognizes Optional fields
func TestFromStructWithOptional(t *testing.T) {
	type Config struct {
		Port Optional[int] `gozod:"min=1000,max=9999"`
	}

	schema := gozod.FromStruct[Config]()
	shape := schema.Shape()

	_, ok := shape["Port"]
	require.True(t, ok, "Port schema not found in shape")
}

// TestStructWithOptionalFields tests struct validation with Optional fields
func TestStructWithOptionalFields(t *testing.T) {
	type Config struct {
		Host string        `gozod:"required"`
		Port Optional[int] `gozod:"min=1000,max=9999"`
	}

	t.Run("optional field not set - skip validation", func(t *testing.T) {
		config := Config{
			Host: "localhost",
			Port: EmptyOptional[int](),
		}

		schema := gozod.FromStruct[Config]()
		result, err := schema.Parse(config)
		require.NoError(t, err)
		assert.Equal(t, "localhost", result.Host)
	})

	t.Run("optional field set with valid value", func(t *testing.T) {
		config := Config{
			Host: "localhost",
			Port: NewOptional(3000),
		}

		schema := gozod.FromStruct[Config]()
		result, err := schema.Parse(config)
		require.NoError(t, err)
		assert.Equal(t, 3000, result.Port.Value)
	})

	t.Run("optional field set with invalid value", func(t *testing.T) {
		config := Config{
			Host: "localhost",
			Port: NewOptional(100), // < 1000
		}

		schema := gozod.FromStruct[Config]()
		_, err := schema.Parse(config)
		require.Error(t, err, "expected validation error")
	})

	t.Run("optional field set with zero value", func(t *testing.T) {
		config := Config{
			Host: "localhost",
			Port: NewOptional(0),
		}

		schema := gozod.FromStruct[Config]()
		_, err := schema.Parse(config)
		require.Error(t, err, "expected validation error for zero value")
	})
}
