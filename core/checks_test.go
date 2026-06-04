package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAttachCheck_RunsAttachHooks(t *testing.T) {
	t.Parallel()

	schema := newRegistrySchema()
	check := &ZodCheckInternals{
		OnAttach: []func(any){
			func(schema any) {
				schema.(ZodSchema).Internals().Bag = map[string]any{"first": true}
			},
			nil,
			func(schema any) {
				schema.(ZodSchema).Internals().Bag["second"] = true
			},
		},
	}

	AttachCheck(schema, check)

	assert.Equal(t, map[string]any{"first": true, "second": true}, schema.Internals().Bag)
}

func TestAttachChecks_RunsSchemaChecks(t *testing.T) {
	t.Parallel()

	schema := newRegistrySchema()
	schema.Internals().Checks = []ZodCheck{
		nil,
		&ZodCheckInternals{
			OnAttach: []func(any){
				func(schema any) {
					internals := schema.(ZodSchema).Internals()
					internals.Bag = map[string]any{"attached": 1}
				},
			},
		},
		internalsCheck{},
	}

	AttachChecks(schema)

	assert.Equal(t, map[string]any{"attached": 1}, schema.Internals().Bag)
}

func TestAttachCheck_IgnoresMissingInputs(t *testing.T) {
	t.Parallel()

	schema := newRegistrySchema()
	AttachCheck(nil, &ZodCheckInternals{})
	AttachCheck(schema, nil)
	AttachCheck(schema, internalsCheck{})
	AttachChecks(nil)

	assert.Nil(t, schema.Internals().Bag)
}

func TestNewZodCheckDef_ClonesParams(t *testing.T) {
	t.Parallel()

	params := map[string]any{"minimum": 2}
	def := NewZodCheckDef("min_length", params)
	params["minimum"] = 99

	assert.Equal(t, "min_length", def.Check)
	assert.Equal(t, 2, def.Params["minimum"])
}
