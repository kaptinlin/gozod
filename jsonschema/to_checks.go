package jsonschema

import (
	lib "github.com/kaptinlin/jsonschema"

	"github.com/kaptinlin/gozod/core"
)

func (c *converter) applyCheckProjections(js *lib.Schema, internals *core.ZodTypeInternals) {
	if js == nil || internals == nil || len(internals.Checks) == 0 {
		return
	}
	for _, check := range internals.Checks {
		if check == nil || check.Zod() == nil || check.Zod().Def == nil {
			continue
		}
		c.applyCheckProjection(js, check.Zod().Def)
	}
}

func (c *converter) applyCheckProjection(js *lib.Schema, def *core.ZodCheckDef) {
	switch def.Check {
	case "min_length":
		if n, ok := paramFloat(def, "minimum"); ok {
			setLowerLengthBound(js, n)
		}
	case "max_length":
		if n, ok := paramFloat(def, "maximum"); ok {
			setUpperLengthBound(js, n)
		}
	case "length_equals":
		if n, ok := paramFloat(def, "exact"); ok {
			setExactLengthBound(js, n)
		}
	case "length_range":
		if n, ok := paramFloat(def, "minimum"); ok {
			setLowerLengthBound(js, n)
		}
		if n, ok := paramFloat(def, "maximum"); ok {
			setUpperLengthBound(js, n)
		}
	case "min_size":
		if n, ok := paramFloat(def, "minimum"); ok {
			setLowerSizeBound(js, n)
		}
	case "max_size":
		if n, ok := paramFloat(def, "maximum"); ok {
			setUpperSizeBound(js, n)
		}
	case "size_equals":
		if n, ok := paramFloat(def, "exact"); ok {
			setExactSizeBound(js, n)
		}
	case "size_range":
		if n, ok := paramFloat(def, "minimum"); ok {
			setLowerSizeBound(js, n)
		}
		if n, ok := paramFloat(def, "maximum"); ok {
			setUpperSizeBound(js, n)
		}
	}
}

func paramFloat(def *core.ZodCheckDef, name string) (float64, bool) {
	if def == nil || len(def.Params) == 0 {
		return 0, false
	}
	return toFloat(def.Params[name])
}

func setLowerLengthBound(js *lib.Schema, n float64) {
	if usesArrayLength(js) {
		if js.MinItems == nil || n > *js.MinItems {
			js.MinItems = &n
		}
		return
	}
	if js.MinLength == nil || n > *js.MinLength {
		js.MinLength = &n
	}
}

func setUpperLengthBound(js *lib.Schema, n float64) {
	if usesArrayLength(js) {
		if js.MaxItems == nil || n < *js.MaxItems {
			js.MaxItems = &n
		}
		return
	}
	if js.MaxLength == nil || n < *js.MaxLength {
		js.MaxLength = &n
	}
}

func setExactLengthBound(js *lib.Schema, n float64) {
	if usesArrayLength(js) {
		js.MinItems = &n
		js.MaxItems = &n
		return
	}
	js.MinLength = &n
	js.MaxLength = &n
}

func setLowerSizeBound(js *lib.Schema, n float64) {
	switch {
	case usesArrayLength(js):
		if js.MinItems == nil || n > *js.MinItems {
			js.MinItems = &n
		}
	case usesObjectSize(js):
		if js.MinProperties == nil || n > *js.MinProperties {
			js.MinProperties = &n
		}
	default:
		if js.MinLength == nil || n > *js.MinLength {
			js.MinLength = &n
		}
	}
}

func setUpperSizeBound(js *lib.Schema, n float64) {
	switch {
	case usesArrayLength(js):
		if js.MaxItems == nil || n < *js.MaxItems {
			js.MaxItems = &n
		}
	case usesObjectSize(js):
		if js.MaxProperties == nil || n < *js.MaxProperties {
			js.MaxProperties = &n
		}
	default:
		if js.MaxLength == nil || n < *js.MaxLength {
			js.MaxLength = &n
		}
	}
}

func setExactSizeBound(js *lib.Schema, n float64) {
	switch {
	case usesArrayLength(js):
		js.MinItems = &n
		js.MaxItems = &n
	case usesObjectSize(js):
		js.MinProperties = &n
		js.MaxProperties = &n
	default:
		js.MinLength = &n
		js.MaxLength = &n
	}
}

func usesArrayLength(js *lib.Schema) bool {
	for _, typ := range js.Type {
		if typ == "array" {
			return true
		}
	}
	return false
}

func usesObjectSize(js *lib.Schema) bool {
	for _, typ := range js.Type {
		if typ == "object" {
			return true
		}
	}
	return false
}
