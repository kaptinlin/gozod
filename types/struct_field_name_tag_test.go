package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
)

type fieldNameTagUser struct {
	UserName string `gozod:"min=3" json:"userName" yaml:"user_name"`
}

type fieldNameTagProfile struct {
	UserName string `gozod:"min=3" json:"userName" yaml:"user_name"`
}

type fieldNameTagAccount struct {
	Profile fieldNameTagProfile `gozod:"required" json:"profile" yaml:"profile"`
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

func TestFromStruct_WithFieldNameTag(t *testing.T) {
	t.Run("default uses json field name", func(t *testing.T) {
		schema := MustFromStruct[fieldNameTagUser]()
		_, err := schema.Parse(fieldNameTagUser{UserName: "ab"})
		assert.Equal(t, []any{"userName"}, firstIssuePath(t, err))
	})

	t.Run("WithFieldNameTag yaml uses yaml field name", func(t *testing.T) {
		schema := MustFromStruct[fieldNameTagUser](WithFieldNameTag("yaml"))
		_, err := schema.Parse(fieldNameTagUser{UserName: "ab"})
		assert.Equal(t, []any{"user_name"}, firstIssuePath(t, err))
	})

	t.Run("field-name tag survives copy-on-write modifier", func(t *testing.T) {
		base := MustFromStruct[fieldNameTagUser](WithFieldNameTag("yaml"))
		derived := base.Describe("a user")
		_, err := derived.Parse(fieldNameTagUser{UserName: "ab"})
		assert.Equal(t, []any{"user_name"}, firstIssuePath(t, err))
	})

	t.Run("nested structs use selected field-name tag", func(t *testing.T) {
		schema := MustFromStruct[fieldNameTagAccount](WithFieldNameTag("yaml"))
		_, err := schema.Parse(fieldNameTagAccount{Profile: fieldNameTagProfile{UserName: "ab"}})
		assert.Equal(t, []any{"profile", "user_name"}, firstIssuePath(t, err))
	})
}

func TestValidateStruct_WithFieldNameTag(t *testing.T) {
	_, err := ValidateStruct(fieldNameTagUser{UserName: "ab"}, WithFieldNameTag("yaml"))
	assert.Equal(t, []any{"user_name"}, firstIssuePath(t, err))
}

func TestObject_WithFieldNameTag_CopyOnWrite(t *testing.T) {
	base := Object(core.ObjectSchema{
		"user_name": String().Min(3),
		"age":       Int(),
	})
	withYAML := base.WithFieldNameTag("yaml")

	assert.Equal(t, "", base.internals.FieldNameTag, "original schema must be unchanged")
	assert.Equal(t, "yaml", withYAML.internals.FieldNameTag)

	// FieldNameTag survives a subsequent copy-on-write modifier.
	assert.Equal(t, "yaml", withYAML.Describe("x").internals.FieldNameTag)

	picked, err := withYAML.Pick([]string{"user_name"})
	require.NoError(t, err)
	assert.Equal(t, "yaml", picked.internals.FieldNameTag)

	omitted, err := withYAML.Omit([]string{"age"})
	require.NoError(t, err)
	assert.Equal(t, "yaml", omitted.internals.FieldNameTag)

	extended, err := withYAML.Extend(core.ObjectSchema{"active": Bool()})
	require.NoError(t, err)
	assert.Equal(t, "yaml", extended.internals.FieldNameTag)

	assert.Equal(t, "yaml", withYAML.SafeExtend(core.ObjectSchema{"active": Bool()}).internals.FieldNameTag)
	assert.Equal(t, "yaml", withYAML.Merge(Object(core.ObjectSchema{"active": Bool()})).internals.FieldNameTag)
}
