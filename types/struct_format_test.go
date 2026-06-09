package types

import (
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type formatUser struct {
	UserName string `gozod:"min=3" json:"userName" yaml:"user_name"`
}

// firstIssuePath returns the path of the first issue of a failed parse.
func firstIssuePath(t *testing.T, err error) []any {
	t.Helper()
	require.Error(t, err)
	var zErr *issues.ZodError
	require.True(t, issues.IsZodError(err, &zErr))
	require.NotEmpty(t, zErr.Issues)
	return zErr.Issues[0].Path
}

func TestFromStruct_WithFormat(t *testing.T) {
	t.Run("default uses json field name", func(t *testing.T) {
		schema := FromStruct[formatUser]()
		_, err := schema.Parse(formatUser{UserName: "ab"})
		assert.Equal(t, []any{"userName"}, firstIssuePath(t, err))
	})

	t.Run("WithFormat yaml uses yaml field name", func(t *testing.T) {
		schema := FromStruct[formatUser](WithFormat("yaml"))
		_, err := schema.Parse(formatUser{UserName: "ab"})
		assert.Equal(t, []any{"user_name"}, firstIssuePath(t, err))
	})

	t.Run("format survives copy-on-write modifier", func(t *testing.T) {
		base := FromStruct[formatUser](WithFormat("yaml"))
		derived := base.Describe("a user")
		_, err := derived.Parse(formatUser{UserName: "ab"})
		assert.Equal(t, []any{"user_name"}, firstIssuePath(t, err))
	})
}

func TestObject_WithFormat_CopyOnWrite(t *testing.T) {
	base := Object(core.ObjectSchema{"user_name": String().Min(3)})
	withYAML := base.WithFormat("yaml")

	assert.Equal(t, "", base.internals.Format, "original schema must be unchanged")
	assert.Equal(t, "yaml", withYAML.internals.Format)

	// Format survives a subsequent copy-on-write modifier.
	assert.Equal(t, "yaml", withYAML.Describe("x").internals.Format)
}
