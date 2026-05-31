package checks

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/gozod/core"
)

func TestMetadataChecks_RegisterDescriptionAndMeta(t *testing.T) {
	schema := newCheckAttachSchema(core.ZodTypeString)

	describe := Describe("User email")
	require.Equal(t, "describe", describe.Zod().Def.Check)
	executeCheck(describe, core.NewParsePayload("unused"))
	applyCheckAttach(t, describe, schema)

	meta, ok := core.GlobalRegistry.Get(schema)
	require.True(t, ok)
	assert.Equal(t, "User email", meta.Description)

	metadata := core.GlobalMeta{
		ID:          "Email",
		Title:       "Email address",
		Description: "Primary contact address",
		Examples:    []any{"ada@example.com"},
	}
	metaCheck := Meta(metadata)
	require.Equal(t, "meta", metaCheck.Zod().Def.Check)
	applyCheckAttach(t, metaCheck, schema)

	meta, ok = core.GlobalRegistry.Get(schema)
	require.True(t, ok)
	assert.Equal(t, metadata.ID, meta.ID)
	assert.Equal(t, metadata.Title, meta.Title)
	assert.Equal(t, metadata.Description, meta.Description)
	if diff := cmp.Diff(metadata.Examples, meta.Examples); diff != "" {
		t.Errorf("Meta() examples mismatch (-want +got):\n%s", diff)
	}
}

func TestMetadataChecks_IgnoreNonSchemaTargets(t *testing.T) {
	check := Describe("ignored")
	for _, attach := range check.Zod().OnAttach {
		attach(struct{}{})
	}
}
