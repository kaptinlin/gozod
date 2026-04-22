package validate_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kaptinlin/gozod/pkg/validate"
)

func TestRegexNilPattern(t *testing.T) {
	t.Parallel()

	assert.False(t, validate.Regex("gozod", nil))
}

func TestPropertyNilValidator(t *testing.T) {
	t.Parallel()

	assert.False(t, validate.Property(map[string]any{"name": "gozod"}, "name", nil))
}
