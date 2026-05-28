package locales

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/gozod/core"
)

func TestDefaultLocaleFormattersProduceMessages(t *testing.T) {
	// DefaultLocales is process-wide, so this test runs serially.
	issues := []core.ZodRawIssue{
		{
			Code:       core.InvalidType,
			Input:      123,
			Properties: map[string]any{"expected": "string"},
		},
		{
			Code: core.TooSmall,
			Properties: map[string]any{
				"origin":    "string",
				"minimum":   3,
				"inclusive": true,
			},
		},
		{
			Code: core.TooBig,
			Properties: map[string]any{
				"origin":    "array",
				"maximum":   2,
				"inclusive": false,
			},
		},
		{
			Code: core.InvalidFormat,
			Properties: map[string]any{
				"format": "starts_with",
				"prefix": "go",
			},
		},
		{
			Code:       core.NotMultipleOf,
			Properties: map[string]any{"divisor": 5},
		},
		{
			Code:       core.UnrecognizedKeys,
			Properties: map[string]any{"keys": []string{"extra"}},
		},
	}

	for locale, formatter := range DefaultLocales {
		t.Run(locale, func(t *testing.T) {
			require.NotNil(t, formatter)
			for _, issue := range issues {
				assert.NotEmpty(t, formatter(issue))
			}
		})
	}
}

func TestEnglishPublicFormatterHelpers(t *testing.T) {
	t.Parallel()

	issue := core.ZodRawIssue{
		Code:       core.InvalidFormat,
		Properties: map[string]any{"format": "starts_with", "prefix": "go"},
	}

	config := EN()
	require.NotNil(t, config)
	require.NotNil(t, config.LocaleError)
	assert.NotEmpty(t, config.LocaleError(issue))
	assert.Equal(t, config.LocaleError(issue), FormatMessageEn(issue))
}
