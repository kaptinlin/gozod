package checks

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/internal/utils"
	"github.com/kaptinlin/gozod/pkg/validate"
)

// executeCheck executes the check logic, respecting conditional execution.
func executeCheck(check core.ZodCheck, payload *core.ParsePayload) {
	internals := check.Zod()
	if internals == nil {
		return
	}
	if internals.When != nil && !internals.When(payload) {
		return
	}
	if internals.Check != nil {
		internals.Check(payload)
	}
}

func TestDirectCheckCreation(t *testing.T) {
	def := &core.ZodCheckDef{
		Check: "test_check",
	}

	check := &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if payload.Value() == nil {
				return
			}
			if str, ok := payload.Value().(string); ok && str == "invalid" {
				payload.AddIssue(issues.CreateCustomIssue("test error", nil, payload.Value()))
			}
		},
	}

	internals := check.Zod()
	require.NotNil(t, internals, "Zod returned nil")
	assert.Equal(t, "test_check", internals.Def.Check)
}

func TestCheckExecution(t *testing.T) {
	def := &core.ZodCheckDef{
		Check: "length_check",
	}

	check := &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.MaxLength(payload.Value(), 5) {
				origin := utils.OriginFromValue(payload.Value())
				payload.AddIssue(issues.CreateTooBigIssue(5, true, origin, payload.Value()))
			}
		},
	}

	payload := core.NewParsePayload("test")

	executeCheck(check, payload)

	if len(payload.Issues()) != 0 {
		t.Errorf("Expected no issues for valid input, got %d", len(payload.Issues()))
	}

	payload = core.NewParsePayload("this is too long")

	executeCheck(check, payload)

	if len(payload.Issues()) != 1 {
		t.Errorf("Expected 1 issue for invalid input, got %d", len(payload.Issues()))
	}
}

func TestConditionalExecution(t *testing.T) {
	def := &core.ZodCheckDef{
		Check: "conditional_check",
	}

	check := &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			payload.AddIssue(issues.CreateCustomIssue("should not execute", nil, payload.Value()))
		},
		When: func(payload *core.ParsePayload) bool {
			_, ok := payload.Value().(string)
			return ok
		},
	}

	payload := core.NewParsePayload(42)

	executeCheck(check, payload)

	if len(payload.Issues()) != 0 {
		t.Errorf("Expected no issues for skipped check, got %d", len(payload.Issues()))
	}

	payload = core.NewParsePayload("test")

	executeCheck(check, payload)

	if len(payload.Issues()) != 1 {
		t.Errorf("Expected 1 issue for executed check, got %d", len(payload.Issues()))
	}
}

func BenchmarkDirectCheckCreation(b *testing.B) {
	def := &core.ZodCheckDef{
		Check: "benchmark_check",
	}

	b.ResetTimer()
	for b.Loop() {
		check := &core.ZodCheckInternals{
			Def:   def,
			Check: func(payload *core.ParsePayload) {},
		}
		_ = check
	}
}

func BenchmarkCheckExecution(b *testing.B) {
	def := &core.ZodCheckDef{
		Check: "benchmark_check",
	}

	check := &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			validate.MaxLength(payload.Value(), 10)
		},
	}

	b.ResetTimer()
	for b.Loop() {
		newPayload := core.NewParsePayload("test")
		executeCheck(check, newPayload)
	}
}

type propertyTestSchema struct {
	parse func(any, *core.ParseContext) (any, error)
}

func (s propertyTestSchema) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	var parseCtx *core.ParseContext
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}
	return s.parse(input, parseCtx)
}

func (s propertyTestSchema) Internals() *core.ZodTypeInternals {
	return &core.ZodTypeInternals{Type: core.ZodTypeAny}
}

func (s propertyTestSchema) IsOptional() bool {
	return false
}

func (s propertyTestSchema) IsNilable() bool {
	return false
}

func TestNewProperty_ValidatesPresentMapProperty(t *testing.T) {
	schema := propertyTestSchema{
		parse: func(input any, ctx *core.ParseContext) (any, error) {
			age, ok := input.(int)
			if !ok || age < 18 {
				return nil, errors.New("age must be at least 18")
			}
			return age, nil
		},
	}
	check := NewProperty("age", schema)

	valid := core.NewParsePayload(map[string]any{"age": 21})
	executeCheck(check, valid)
	assert.Empty(t, valid.Issues())

	missing := core.NewParsePayload(map[string]any{"name": "gopher"})
	executeCheck(check, missing)
	assert.Empty(t, missing.Issues())

	nonMap := core.NewParsePayload("not an object")
	executeCheck(check, nonMap)
	assert.Empty(t, nonMap.Issues())

	invalid := core.NewParsePayloadWithPath(map[string]any{"age": 17}, []any{"user"})
	executeCheck(check, invalid)
	issues := invalid.Issues()
	require.Len(t, issues, 1)
	assert.Equal(t, core.Custom, issues[0].Code)
	assert.Equal(t, "age must be at least 18", issues[0].Message)
	require.Len(t, issues[0].Path, 2)
	assert.Equal(t, "user", issues[0].Path[0])
	assert.Equal(t, "age", issues[0].Path[1])
	assert.Equal(t, "property", issues[0].Properties["origin"])
	assert.Equal(t, "age", issues[0].Properties["property"])
}
