package engine

import (
	"reflect"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/structx"
)

// WithError creates a SchemaParams with a custom error message.
func WithError(message string) core.SchemaParams {
	return core.SchemaParams{Error: message}
}

// WithDescription creates a SchemaParams with a description.
func WithDescription(desc string) core.SchemaParams {
	return core.SchemaParams{Description: desc}
}

// WithAbort creates a SchemaParams with abort-on-error enabled.
func WithAbort() core.SchemaParams {
	return core.SchemaParams{Abort: true}
}

// ProcessSchemaParams merges schema parameters into a configuration map.
func ProcessSchemaParams(params ...core.SchemaParams) map[string]any {
	cfg := make(map[string]any)
	for _, p := range params {
		if m, err := structx.ToMap(p); err == nil {
			cfg = mapx.Merge(cfg, m)
		}
	}
	return cfg
}

// IsValidSchemaType reports whether value implements core.ZodType.
func IsValidSchemaType(value any) bool {
	if value == nil {
		return false
	}
	return reflect.TypeOf(value).Implements(reflect.TypeFor[core.ZodType[any]]())
}
