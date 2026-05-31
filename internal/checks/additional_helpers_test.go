package checks

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/gozod/core"
)

type checkAttachSchema struct {
	internals core.ZodTypeInternals
}

func newCheckAttachSchema(schemaType core.ZodTypeCode) *checkAttachSchema {
	return &checkAttachSchema{internals: core.ZodTypeInternals{Type: schemaType}}
}

func (s *checkAttachSchema) Internals() *core.ZodTypeInternals {
	return &s.internals
}

func (s *checkAttachSchema) ParseAny(input any, _ ...*core.ParseContext) (any, error) {
	return input, nil
}

func (s *checkAttachSchema) IsOptional() bool {
	return s.internals.IsOptional()
}

func (s *checkAttachSchema) IsNilable() bool {
	return s.internals.IsNilable()
}

func applyCheckAttach(t *testing.T, check core.ZodCheck, schema *checkAttachSchema) map[string]any {
	t.Helper()

	internals := check.Zod()
	require.NotNil(t, internals)
	for _, attach := range internals.OnAttach {
		if attach != nil {
			attach(schema)
		}
	}
	return schema.internals.Bag
}

func requireCheckAccepts(t *testing.T, check core.ZodCheck, input any) {
	t.Helper()

	payload := core.NewParsePayload(input)
	executeCheck(check, payload)
	require.Len(t, payload.Issues(), 0)
}

func requireCheckRejects(t *testing.T, check core.ZodCheck, input any, wantCode core.IssueCode) core.ZodRawIssue {
	t.Helper()

	payload := core.NewParsePayload(input)
	executeCheck(check, payload)
	issues := payload.Issues()
	require.Len(t, issues, 1)
	assert.Equal(t, wantCode, issues[0].Code)
	return issues[0]
}
