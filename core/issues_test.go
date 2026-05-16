package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestZodIssue_AccessorsExposeTypedProperties(t *testing.T) {
	t.Parallel()

	issue := ZodIssue{
		ZodIssueBase: ZodIssueBase{
			Code:    InvalidType,
			Message: "expected string",
			Path:    []any{"name"},
		},
		Expected: ZodTypeString,
		Received: ZodTypeNumber,
		Minimum:  2,
		Maximum:  5,
		Format:   "email",
		Divisor:  2,
	}

	minValue, ok := issue.MinValue()
	assert.True(t, ok)
	assert.Equal(t, 2, minValue)

	maxValue, ok := issue.MaxValue()
	assert.True(t, ok)
	assert.Equal(t, 5, maxValue)

	expected, ok := issue.ExpectedType()
	assert.True(t, ok)
	assert.Equal(t, ZodTypeString, expected)

	received, ok := issue.ReceivedType()
	assert.True(t, ok)
	assert.Equal(t, ZodTypeNumber, received)

	assert.Equal(t, "expected string", issue.Error())
	assert.Contains(t, issue.String(), "invalid_type")
}

func TestZodIssue_AccessorsRespectIssueCode(t *testing.T) {
	t.Parallel()

	issue := ZodIssue{ZodIssueBase: ZodIssueBase{Code: Custom}}

	_, ok := issue.MinValue()
	assert.False(t, ok)
	_, ok = issue.MaxValue()
	assert.False(t, ok)
	_, ok = issue.ExpectedType()
	assert.False(t, ok)
	_, ok = issue.ReceivedType()
	assert.False(t, ok)
	_, ok = issue.FormatName()
	assert.False(t, ok)
	_, ok = issue.DivisorValue()
	assert.False(t, ok)
}

func TestZodIssue_FormatAndDivisorAccessorsRequireMatchingCode(t *testing.T) {
	t.Parallel()

	formatIssue := ZodIssue{ZodIssueBase: ZodIssueBase{Code: InvalidFormat}, Format: "email"}
	format, ok := formatIssue.FormatName()
	assert.True(t, ok)
	assert.Equal(t, "email", format)

	multipleIssue := ZodIssue{ZodIssueBase: ZodIssueBase{Code: NotMultipleOf}, Divisor: 3}
	divisor, ok := multipleIssue.DivisorValue()
	assert.True(t, ok)
	assert.Equal(t, 3, divisor)
}

func TestZodRawIssue_PropertyAccessors(t *testing.T) {
	t.Parallel()

	issue := ZodRawIssue{Properties: map[string]any{
		"expected":  ZodTypeString,
		"received":  string(ZodTypeNumber),
		"origin":    "string",
		"format":    "email",
		"pattern":   "^[a-z]+$",
		"prefix":    "go",
		"suffix":    "zod",
		"includes":  "oz",
		"minimum":   2,
		"maximum":   10,
		"inclusive": true,
		"divisor":   2,
		"keys":      []string{"extra"},
		"values":    []any{"a", "b"},
	}}

	assert.Equal(t, ZodTypeString, issue.Expected())
	assert.Equal(t, ZodTypeNumber, issue.Received())
	assert.Equal(t, "string", issue.Origin())
	assert.Equal(t, "email", issue.Format())
	assert.Equal(t, "^[a-z]+$", issue.Pattern())
	assert.Equal(t, "go", issue.Prefix())
	assert.Equal(t, "zod", issue.Suffix())
	assert.Equal(t, "oz", issue.Includes())
	assert.Equal(t, 2, issue.Minimum())
	assert.Equal(t, 10, issue.Maximum())
	assert.True(t, issue.Inclusive())
	assert.Equal(t, 2, issue.Divisor())
	assert.Len(t, issue.Keys(), 1)
	assert.Len(t, issue.Values(), 2)
}

func TestZodRawIssue_PropertyAccessorsReturnZeroValues(t *testing.T) {
	t.Parallel()

	issue := ZodRawIssue{Properties: map[string]any{
		"expected": 123,
		"keys":     []any{"wrong"},
	}}

	assert.Empty(t, issue.Expected())
	assert.Empty(t, issue.Received())
	assert.Empty(t, issue.Origin())
	assert.Empty(t, issue.Format())
	assert.Empty(t, issue.Pattern())
	assert.Empty(t, issue.Prefix())
	assert.Empty(t, issue.Suffix())
	assert.Empty(t, issue.Includes())
	assert.Nil(t, issue.Minimum())
	assert.Nil(t, issue.Maximum())
	assert.False(t, issue.Inclusive())
	assert.Nil(t, issue.Divisor())
	assert.Nil(t, issue.Keys())
	assert.Nil(t, issue.Values())
}
