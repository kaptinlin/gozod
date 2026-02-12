package engine

import (
	"fmt"
	"maps"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/pkg/mapx"
)

// InitZodType initializes the common fields of a ZodType.
func InitZodType[T core.ZodType[any]](schema T, def *core.ZodTypeDef) {
	i := schema.Internals()

	i.Type = def.Type
	i.Error = def.Error

	if def.Coerce {
		i.SetCoerce(true)
	}

	if i.Bag == nil {
		i.Bag = make(map[string]any)
	}
	mapx.Set(i.Bag, "type", def.Type)

	i.Checks = make([]core.ZodCheck, 0)
	for _, c := range def.Checks {
		i.AddCheck(c)
	}

	if i.Values == nil {
		i.Values = make(map[any]struct{})
	}

	for _, c := range i.Checks {
		if c == nil {
			continue
		}
		if ci := c.Zod(); ci != nil {
			for _, fn := range ci.OnAttach {
				fn(any(schema).(core.ZodType[any]))
			}
		}
	}
}

// NewBaseZodTypeInternals creates a ZodTypeInternals with initialized collections.
func NewBaseZodTypeInternals(typeName core.ZodTypeCode) core.ZodTypeInternals {
	return core.ZodTypeInternals{
		Type:   typeName,
		Checks: make([]core.ZodCheck, 0),
		Values: make(map[any]struct{}),
		Bag:    make(map[string]any),
	}
}

// AddCheck adds a validation check and returns a new schema instance (copy-on-write).
func AddCheck[T interface{ Internals() *core.ZodTypeInternals }](schema T, check core.ZodCheck) core.ZodType[any] {
	src := schema.Internals()

	checks := append(append([]core.ZodCheck(nil), src.Checks...), check)
	def := &core.ZodTypeDef{
		Type:   src.Type,
		Error:  src.Error,
		Checks: checks,
	}

	if src.Constructor == nil {
		panic(fmt.Sprintf("no constructor found for type %T: framework bug", schema))
	}

	dst := src.Constructor(def)
	di := dst.Internals()

	cloned := src.Clone()
	*di = *cloned
	di.Checks = checks

	if c, ok := dst.(core.Cloneable); ok {
		if s, ok := any(schema).(core.Cloneable); ok {
			c.CloneFrom(s)
		}
	}

	if check != nil {
		if ci := check.Zod(); ci != nil {
			for _, fn := range ci.OnAttach {
				fn(dst)
			}
		}
	}

	return dst
}

// Clone creates a new schema instance with optional definition modifications (copy-on-write).
func Clone[T interface{ Internals() *core.ZodTypeInternals }](schema T, modify func(*core.ZodTypeDef)) core.ZodType[any] {
	src := schema.Internals()

	def := &core.ZodTypeDef{
		Type:   src.Type,
		Error:  src.Error,
		Checks: append([]core.ZodCheck(nil), src.Checks...),
	}

	if modify != nil {
		modify(def)
	}

	if src.Constructor == nil {
		panic(fmt.Sprintf("no constructor found for type %T: framework bug", schema))
	}

	dst := src.Constructor(def)
	di := dst.Internals()

	cloned := src.Clone()
	*di = *cloned

	di.Type = def.Type
	di.Error = def.Error
	di.Checks = def.Checks

	if c, ok := dst.(core.Cloneable); ok {
		if s, ok := any(schema).(core.Cloneable); ok {
			c.CloneFrom(s)
		}
	}

	return dst
}

// CopyInternalsState copies all state from src to dst.
func CopyInternalsState(dst, src *core.ZodTypeInternals) {
	if src == nil || dst == nil {
		return
	}
	if cloned := src.Clone(); cloned != nil {
		*dst = *cloned
	}
}

// CreateInternalsWithState creates new internals with state cloned from src.
func CreateInternalsWithState(src *core.ZodTypeInternals, typeName core.ZodTypeCode) *core.ZodTypeInternals {
	if src == nil {
		return &core.ZodTypeInternals{
			Type:   typeName,
			Checks: make([]core.ZodCheck, 0),
			Values: make(map[any]struct{}),
			Bag:    make(map[string]any),
		}
	}
	cloned := src.Clone()
	cloned.Type = typeName
	return cloned
}

// MergeInternalsState merges state from src into dst, preserving dst's identity.
func MergeInternalsState(dst, src *core.ZodTypeInternals) {
	if src == nil || dst == nil {
		return
	}

	// Merge flags
	if src.IsOptional() {
		dst.SetOptional(true)
	}
	if src.IsNilable() {
		dst.SetNilable(true)
	}
	if src.IsCoerce() {
		dst.SetCoerce(true)
	}

	// Merge modifiers
	if src.DefaultValue != nil {
		dst.SetDefaultValue(src.DefaultValue)
	}
	if src.DefaultFunc != nil {
		dst.SetDefaultFunc(src.DefaultFunc)
	}
	if src.PrefaultValue != nil {
		dst.SetPrefaultValue(src.PrefaultValue)
	}
	if src.PrefaultFunc != nil {
		dst.SetPrefaultFunc(src.PrefaultFunc)
	}
	if src.Transform != nil {
		dst.SetTransform(src.Transform)
	}

	for _, c := range src.Checks {
		dst.AddCheck(c)
	}

	if len(src.Values) > 0 {
		if dst.Values == nil {
			dst.Values = make(map[any]struct{})
		}
		maps.Copy(dst.Values, src.Values)
	}

	if len(src.Bag) > 0 {
		if dst.Bag == nil {
			dst.Bag = make(map[string]any)
		}
		dst.Bag = mapx.Merge(dst.Bag, src.Bag)
	}

	if src.Pattern != nil {
		dst.Pattern = src.Pattern
	}
	if src.Constructor != nil {
		dst.Constructor = src.Constructor
	}
}
