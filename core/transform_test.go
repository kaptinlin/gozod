package core

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testZodType[T any] struct {
	internals *ZodTypeInternals
	parse     func(any, *ParseContext) (T, error)
}

func newTestZodType[T any](typeCode ZodTypeCode, parse func(any, *ParseContext) (T, error)) *testZodType[T] {
	return &testZodType[T]{
		internals: &ZodTypeInternals{Type: typeCode},
		parse:     parse,
	}
}

func (s *testZodType[T]) Parse(input any, ctx ...*ParseContext) (T, error) {
	return s.parse(input, getOrCreateContext(ctx...))
}

func (s *testZodType[T]) MustParse(input any, ctx ...*ParseContext) T {
	result, err := s.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

func (s *testZodType[T]) ParseAny(input any, ctx ...*ParseContext) (any, error) {
	return s.Parse(input, ctx...)
}

func (s *testZodType[T]) Internals() *ZodTypeInternals {
	return s.internals
}

func (s *testZodType[T]) IsOptional() bool {
	return s.internals.IsOptional()
}

func (s *testZodType[T]) IsNilable() bool {
	return s.internals.IsNilable()
}

func TestZodTransform_ParsesAndTransformsValidatedValue(t *testing.T) {
	t.Parallel()

	source := newTestZodType(ZodTypeString, func(input any, ctx *ParseContext) (string, error) {
		value, ok := input.(string)
		if !ok {
			return "", ErrInvalidTransformType
		}
		return value, nil
	})
	source.internals.SetOptional(true)
	source.internals.SetNilable(true)

	var transformSawReportInput bool
	transform := NewZodTransform(source, func(value string, ctx *RefinementContext) (int, error) {
		transformSawReportInput = ctx.ReportInput
		return len(value), nil
	})

	got, err := transform.Parse("gozod", &ParseContext{ReportInput: true})
	require.NoError(t, err)
	assert.Equal(t, 5, got)
	assert.True(t, transformSawReportInput)
	assert.Equal(t, ZodTypeTransform, transform.Internals().Type)
	assert.True(t, transform.IsOptional())
	assert.True(t, transform.IsNilable())
	assert.Same(t, source, transform.Inner())

	anyGot, err := transform.ParseAny("go")
	require.NoError(t, err)
	assert.Equal(t, 2, anyGot)
}

func TestZodTransform_ReturnsRefinementIssues(t *testing.T) {
	t.Parallel()

	source := newTestZodType(ZodTypeString, func(input any, ctx *ParseContext) (string, error) {
		return input.(string), nil
	})
	transform := NewZodTransform(source, func(value string, ctx *RefinementContext) (string, error) {
		ctx.AddIssue(ZodIssue{ZodIssueBase: ZodIssueBase{Message: "invalid"}})
		return value, nil
	})

	_, err := transform.Parse("gozod")
	require.Error(t, err)
	var joined interface{ Unwrap() []error }
	require.ErrorAs(t, err, &joined)
	assert.Len(t, joined.Unwrap(), 1)
}

func TestZodTransform_DefaultValueBypassesTransform(t *testing.T) {
	t.Parallel()

	source := newTestZodType(ZodTypeString, func(input any, ctx *ParseContext) (string, error) {
		if input == nil {
			return "fallback", nil
		}
		return input.(string), nil
	})
	source.internals.SetDefaultValue("fallback")

	called := false
	transform := NewZodTransform(source, func(value string, ctx *RefinementContext) (string, error) {
		called = true
		return value + "!", nil
	})

	got, err := transform.Parse(nil)
	require.NoError(t, err)
	assert.Equal(t, "fallback", got)
	assert.False(t, called)
}

func TestZodPipe_ParsesThroughSourceAndTarget(t *testing.T) {
	t.Parallel()

	source := newTestZodType(ZodTypeString, func(input any, ctx *ParseContext) (string, error) {
		return input.(string), nil
	})
	target := newTestZodType[any](ZodTypeNumber, func(input any, ctx *ParseContext) (any, error) {
		value, ok := input.(int)
		if !ok {
			return nil, ErrInvalidTransformType
		}
		return value, nil
	})

	var pipeSawReportInput bool
	pipe := NewZodPipe(source, target, func(input string, ctx *ParseContext) (int, error) {
		pipeSawReportInput = ctx.ReportInput
		return len(input), nil
	})

	got, err := pipe.Parse("gozod", &ParseContext{ReportInput: true})
	require.NoError(t, err)
	assert.Equal(t, 5, got)
	assert.True(t, pipeSawReportInput)
	assert.Equal(t, ZodTypePipe, pipe.Internals().Type)
	assert.Same(t, source, pipe.Inner())
	assert.Same(t, target, pipe.Output())

	anyGot, err := pipe.ParseAny("go")
	require.NoError(t, err)
	assert.Equal(t, 2, anyGot)
}

func TestMustParsePanicsWithErrorValue(t *testing.T) {
	t.Parallel()

	boom := errors.New("boom")
	source := newTestZodType(ZodTypeString, func(input any, ctx *ParseContext) (string, error) {
		return "", boom
	})
	transform := NewZodTransform(source, func(value string, ctx *RefinementContext) (string, error) {
		return value, nil
	})

	defer func() {
		recovered := recover()
		require.NotNil(t, recovered)
		err, ok := recovered.(error)
		require.True(t, ok)
		require.ErrorIs(t, err, boom)
	}()

	_ = transform.MustParse("gozod")
}
