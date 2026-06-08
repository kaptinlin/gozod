package jsonschema

import (
	"maps"
	"math"

	lib "github.com/kaptinlin/jsonschema"

	"github.com/kaptinlin/gozod/core"
)

func cloneBag(internals *core.ZodTypeInternals) map[string]any {
	if internals == nil || len(internals.Bag) == 0 {
		return nil
	}
	return maps.Clone(internals.Bag)
}

func (c *converter) applyStringBag(jsonSchema *lib.Schema, bag map[string]any) {
	if len(bag) == 0 {
		return
	}

	// Process aggregated patterns from the bag
	if patternsRaw, ok := bag["patterns"]; ok {
		if patterns, ok := patternsRaw.([]string); ok && len(patterns) > 0 {
			// Deduplicate patterns to handle schemas that add the same check multiple ways
			uniquePatterns := make(map[string]struct{})
			var result []string
			for _, p := range patterns {
				if _, ok := uniquePatterns[p]; !ok {
					uniquePatterns[p] = struct{}{}
					result = append(result, p)
				}
			}
			patterns = result

			if len(patterns) == 1 {
				jsonSchema.Pattern = new(patterns[0])
			} else {
				jsonSchema.AllOf = make([]*lib.Schema, len(patterns))
				for i, p := range patterns {
					jsonSchema.AllOf[i] = &lib.Schema{
						Pattern: &p,
					}
				}
			}

			// Remove patterns from the local copy so applyBag doesn't re-add them.
			delete(bag, "patterns")
		}
	}

	// Apply other string-related properties from the bag
	if val, ok := bag["format"].(string); ok {
		jsonSchema.Format = &val
	}
	if v, ok := bag["minLength"]; ok {
		if n, ok := toFloat(v); ok {
			jsonSchema.MinLength = &n
		}
	}
	if v, ok := bag["maxLength"]; ok {
		if n, ok := toFloat(v); ok {
			jsonSchema.MaxLength = &n
		}
	}
	if v, ok := bag["contentEncoding"]; ok {
		if ce, ok := v.(string); ok {
			jsonSchema.ContentEncoding = &ce
		}
	}
	if v, ok := bag["contentMediaType"]; ok {
		if cmt, ok := v.(string); ok {
			jsonSchema.ContentMediaType = &cmt
		}
	}
}

// applyBag copies well-known constraint keys from Zod internals.Bag to JSON Schema.
func (c *converter) applyBag(js *lib.Schema, bag map[string]any) {
	// First handle pattern/patterns specially as they may need to merge
	if v, ok := bag["pattern"]; ok {
		if p, ok := v.(string); ok {
			if js.Pattern == nil {
				js.Pattern = &p
			} else {
				if js.AllOf == nil {
					js.AllOf = []*lib.Schema{{Pattern: js.Pattern}}
				}
				js.AllOf = append(js.AllOf, &lib.Schema{Pattern: &p})
				js.Pattern = nil
			}
		}
	}

	if v, ok := bag["patterns"]; ok {
		if patterns, ok := v.([]string); ok && len(patterns) > 0 {
			// Move existing single pattern to allOf
			if js.Pattern != nil {
				if js.AllOf == nil {
					js.AllOf = []*lib.Schema{}
				}
				js.AllOf = append(js.AllOf, &lib.Schema{Pattern: js.Pattern})
				js.Pattern = nil
			}

			if len(patterns) == 1 && js.AllOf == nil {
				js.Pattern = new(patterns[0])
			} else {
				if js.AllOf == nil {
					js.AllOf = []*lib.Schema{}
				}
				for _, p := range patterns {
					js.AllOf = append(js.AllOf, &lib.Schema{Pattern: &p})
				}
			}
		}
	}

	// Table-driven simple setters to minimize reflection and branching.
	for k, v := range bag {
		switch k {
		case "minLength":
			if f, ok := toFloat(v); ok {
				js.MinLength = &f
			}
		case "maxLength":
			if f, ok := toFloat(v); ok {
				js.MaxLength = &f
			}
		case "format":
			if s, ok := v.(string); ok {
				js.Format = &s
			}
		case "contentEncoding":
			if s, ok := v.(string); ok {
				js.ContentEncoding = &s
			}
		case "contentMediaType":
			if s, ok := v.(string); ok {
				js.ContentMediaType = &s
			}
		case "minimum":
			if r, ok := toRat(v); ok {
				js.Minimum = &r
			}
		case "maximum":
			if r, ok := toRat(v); ok {
				js.Maximum = &r
			}
		case "multipleOf":
			if r, ok := toRat(v); ok {
				js.MultipleOf = &r
			}
		case "exclusiveMinimum":
			if r, ok := toRat(v); ok {
				js.ExclusiveMinimum = &r
			}
		case "exclusiveMaximum":
			if r, ok := toRat(v); ok {
				js.ExclusiveMaximum = &r
			}
		case "minItems":
			if f, ok := toFloat(v); ok {
				js.MinItems = &f
			}
		case "maxItems":
			if f, ok := toFloat(v); ok {
				js.MaxItems = &f
			}
		case "minSize":
			if f, ok := toFloat(v); ok {
				js.MinLength = &f
			}
		case "maxSize":
			if f, ok := toFloat(v); ok {
				js.MaxLength = &f
			}
		case "size":
			if f, ok := toFloat(v); ok {
				js.MinLength = &f
				js.MaxLength = &f
			}
		case "mime":
			if mimes, ok := v.([]string); ok && len(mimes) == 1 {
				js.ContentMediaType = new(mimes[0])
			}
		}
	}
}

// toFloat converts numeric types to float64.
func toFloat(v any) (float64, bool) {
	switch x := v.(type) {
	case int:
		return float64(x), true
	case int32:
		return float64(x), true
	case int64:
		return float64(x), true
	case uint:
		return float64(x), true
	case uint32:
		return float64(x), true
	case uint64:
		return float64(x), true
	case float32:
		return float64(x), true
	case float64:
		return x, true
	default:
		return 0, false
	}
}

// toRat converts numeric values to jsonschema.Rat (alias for big.Rat wrapper).
func toRat(v any) (lib.Rat, bool) {
	if f, ok := toFloat(v); ok {
		return *lib.NewRat(f), true
	}
	return lib.Rat{}, false
}

// convertFile handles ZodFile -> JSON Schema file representation.
func (c *converter) convertFile(bag map[string]any) (*lib.Schema, error) {
	s := &lib.Schema{
		Type:            []string{"string"},
		Format:          new("binary"),
		ContentEncoding: new("binary"),
	}

	c.applyBag(s, bag)

	// Handle multiple MIME types
	if mimes, ok := bag["mime"].([]string); ok && len(mimes) > 1 {
		anyOf := make([]*lib.Schema, len(mimes))
		for i, mime := range mimes {
			itemSchema := &lib.Schema{
				Type:             []string{"string"},
				Format:           new("binary"),
				ContentEncoding:  new("binary"),
				ContentMediaType: new(mime),
			}
			// Apply min/max/size constraints using helper conversion
			if v, ok := bag["size"]; ok {
				if f, ok := toFloat(v); ok {
					maxLength := f
					itemSchema.MinLength = &f
					itemSchema.MaxLength = &maxLength
				}
			} else {
				if v, ok := bag["minSize"]; ok {
					if f, ok := toFloat(v); ok {
						itemSchema.MinLength = &f
					}
				}
				if v, ok := bag["maxSize"]; ok {
					if f, ok := toFloat(v); ok {
						itemSchema.MaxLength = &f
					}
				}
			}
			anyOf[i] = itemSchema
		}
		s.AnyOf = anyOf
		// Clear top-level properties that are now in anyOf
		s.Type = nil
		s.Format = nil
		s.ContentEncoding = nil
		s.ContentMediaType = nil
		s.MinLength = nil
		s.MaxLength = nil

		// Remove size-related entries from the local copy so the final bag pass
		// does not reapply them to the top-level schema.
		delete(bag, "minSize")
		delete(bag, "maxSize")
		delete(bag, "size")
	}

	return s, nil
}

// numericRangeDefaults maps ZodTypeCode to its inclusive minimum and maximum.
var numericRangeDefaults = map[core.ZodTypeCode][2]float64{
	// Go platform dependent int range (assuming 64-bit build).
	core.ZodTypeInt:     {float64(math.MinInt), float64(math.MaxInt)},
	core.ZodTypeInteger: {float64(math.MinInt), float64(math.MaxInt)},

	// Signed integers
	core.ZodTypeInt8:  {-128, 127},
	core.ZodTypeInt16: {-32768, 32767},
	core.ZodTypeInt32: {-2147483648, 2147483647},
	// Using float64 to hold these large ints – precision loss is acceptable for JSON-Schema ranges
	core.ZodTypeInt64: {float64(math.MinInt64), float64(math.MaxInt64)}, // Int64 full range

	// Unsigned integers
	core.ZodTypeUint:   {0, float64(math.MaxUint)},
	core.ZodTypeUint8:  {0, 255},
	core.ZodTypeUint16: {0, 65535},
	core.ZodTypeUint32: {0, 4294967295},
	core.ZodTypeUint64: {0, 1.844674407371e19}, // approximate math.MaxUint64 as float64

	// Floats
	core.ZodTypeFloat32: {-math.MaxFloat32, math.MaxFloat32},
	core.ZodTypeFloat64: {-math.MaxFloat64, math.MaxFloat64},
}

// applyNumericRangeDefaults populates default numeric range constraints for a given schema based on its type.
func (c *converter) applyNumericRangeDefaults(zodType core.ZodTypeCode, js *lib.Schema, internals *core.ZodTypeInternals) {
	// Apply only at top-level (depth==1)
	if c.depth != 1 {
		return
	}

	// Do not override explicit constraints from bag.
	if internals != nil && internals.Bag != nil {
		for _, key := range []string{"minimum", "exclusiveMinimum", "maximum", "exclusiveMaximum"} {
			if _, exists := internals.Bag[key]; exists {
				return
			}
		}
	}

	rng, ok := numericRangeDefaults[zodType]
	if !ok {
		return
	}

	// Only set if not already set.
	if js.Minimum == nil {
		js.Minimum = lib.NewRat(rng[0])
	}
	if js.Maximum == nil {
		js.Maximum = lib.NewRat(rng[1])
	}
}
