package checks

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/pkg/regex"
	"github.com/kaptinlin/gozod/pkg/validate"
)

func TestFirstPartyCheckDefs_RecordSemanticParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		check core.ZodCheck
		want  map[string]any
	}{
		{
			name:  "min length",
			check: MinLength(3),
			want:  map[string]any{"minimum": 3},
		},
		{
			name:  "exclusive number bound",
			check: Lt(10),
			want:  map[string]any{"maximum": 10, "inclusive": false},
		},
		{
			name:  "regex pattern",
			check: Regex(regexp.MustCompile("^abc$")),
			want:  map[string]any{"pattern": "^abc$"},
		},
		{
			name:  "format projection",
			check: Email(),
			want:  map[string]any{"format": "email", "pattern": regex.Email.String()},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			def := tt.check.Zod().Def
			require.NotNil(t, def)
			assert.Subset(t, def.Params, tt.want)
		})
	}
}

func TestOptionCheckDefs_RecordRuntimeProjectionParams(t *testing.T) {
	t.Parallel()

	precision := 0
	check := ISODateTimeWithOptions(validate.ISODateTimeOptions{
		Precision: &precision,
		Offset:    true,
		Local:     true,
	})

	params := check.Zod().Def.Params
	assert.Equal(t, "iso_datetime", params["format"])
	assert.Equal(t, regex.Datetime(regex.DatetimeOptions{
		Precision: &precision,
		Offset:    true,
		Local:     true,
	}).String(), params["pattern"])
	assert.Equal(t, 0, params["precision"])
	assert.Equal(t, true, params["offset"])
	assert.Equal(t, true, params["local"])

	bag := applyCheckAttach(t, check, newCheckAttachSchema(core.ZodTypeString))
	patterns, ok := bag["patterns"].([]string)
	require.True(t, ok)
	assert.Contains(t, patterns, params["pattern"])
}

func TestFileCheckDef_ClonesMimeTypes(t *testing.T) {
	t.Parallel()

	mimeTypes := []string{"image/png"}
	check := Mime(mimeTypes)
	mimeTypes[0] = "text/plain"

	assert.Equal(t, []string{"image/png"}, check.Zod().Def.Params["mime"])
}
