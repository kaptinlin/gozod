package engine

import (
	"fmt"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// InitZodType initializes the common fields of a ZodType.
func InitZodType[T core.ZodType[any]](schema T, def *core.ZodTypeDef) {
	internals := schema.GetInternals()

	internals.Type = def.Type
	internals.Error = def.Error

	if def.Coerce {
		internals.SetCoerce(true)
	}

	if internals.Bag == nil {
		internals.Bag = make(map[string]any)
	}
	mapx.Set(internals.Bag, "type", def.Type)

	// Use slicex to safely copy checks
	if !slicex.IsEmpty(def.Checks) {
		internals.Checks = make([]core.ZodCheck, 0)
		for _, check := range def.Checks {
			internals.AddCheck(check)
		}
	} else {
		internals.Checks = make([]core.ZodCheck, 0)
	}

	if internals.Values == nil {
		internals.Values = make(map[any]struct{})
	}

	for _, check := range internals.Checks {
		if check == nil {
			continue
		}
		if ci := check.GetZod(); ci != nil {
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
func AddCheck[T interface{ GetInternals() *core.ZodTypeInternals }](schema T, check core.ZodCheck) core.ZodType[any] {
	internals := schema.GetInternals()

	newChecks := append(append([]core.ZodCheck(nil), internals.Checks...), check)
	newDef := &core.ZodTypeDef{
		Type:   internals.Type,
		Error:  internals.Error,
		Checks: newChecks,
	}

	if internals.Constructor == nil {
		panic(fmt.Sprintf("Internal error: no constructor found for type %T. This indicates a framework bug - please report this issue.", schema))
	}

	newSchema := internals.Constructor(newDef)
	newInternals := newSchema.GetInternals()

	clonedInternals := internals.Clone()
	*newInternals = *clonedInternals
	newInternals.Checks = newChecks

	// Use Cloneable interface to copy type-specific state
	if cloneable, ok := newSchema.(core.Cloneable); ok {
		if source, ok := any(schema).(core.Cloneable); ok {
			cloneable.CloneFrom(source)
		}
	}

	// Execute OnAttach callbacks for the new check
	if check != nil {
		if ci := check.GetZod(); ci != nil {
			for _, fn := range ci.OnAttach {
				fn(newSchema)
			}
		}
	}

	return newSchema
}

// Clone creates a new schema instance with optional definition modifications (copy-on-write).
func Clone[T interface{ GetInternals() *core.ZodTypeInternals }](schema T, modifyDef func(*core.ZodTypeDef)) core.ZodType[any] {
	internals := schema.GetInternals()

	newDef := &core.ZodTypeDef{
		Type:   internals.Type,
		Error:  internals.Error,
		Checks: append([]core.ZodCheck(nil), internals.Checks...),
	}

	if modifyDef != nil {
		modifyDef(newDef)
	}

	if internals.Constructor == nil {
		panic(fmt.Sprintf("Internal error: no constructor found for type %T. This indicates a framework bug - please report this issue.", schema))
	}

	newSchema := internals.Constructor(newDef)
	newInternals := newSchema.GetInternals()

	clonedInternals := internals.Clone()
	*newInternals = *clonedInternals

	newInternals.Type = newDef.Type
	newInternals.Error = newDef.Error
	newInternals.Checks = newDef.Checks

	// Use Cloneable interface to copy type-specific state
	if cloneable, ok := newSchema.(core.Cloneable); ok {
		if sourceAny := any(schema); sourceAny != nil {
			if sourceCloneable, ok := sourceAny.(core.Cloneable); ok {
				cloneable.CloneFrom(sourceCloneable)
			}
		}
	}

	return newSchema
}

// CopyInternalsState copies all state from source to target internals.
func CopyInternalsState(target *core.ZodTypeInternals, source *core.ZodTypeInternals) {
	if source == nil || target == nil {
		return
	}

	// Use core's Clone() method and copy the result
	cloned := source.Clone()
	if cloned != nil {
		*target = *cloned
	}
}

// CreateInternalsWithState creates new internals with state cloned from source.
func CreateInternalsWithState(source *core.ZodTypeInternals, typeName core.ZodTypeCode) *core.ZodTypeInternals {
	if source == nil {
		return &core.ZodTypeInternals{
			Type:   typeName,
			Checks: make([]core.ZodCheck, 0),
			Values: make(map[any]struct{}),
			Bag:    make(map[string]any),
		}
	}

	// Use core's Clone() method and modify the type
	cloned := source.Clone()
	cloned.Type = typeName
	return cloned
}

// MergeInternalsState merges state from source into target, preserving target's identity.
func MergeInternalsState(target *core.ZodTypeInternals, source *core.ZodTypeInternals) {
	if source == nil || target == nil {
		return
	}

	// Use core convenience methods for flag merging
	if source.IsOptional() {
		target.SetOptional(true)
	}
	if source.IsNilable() {
		target.SetNilable(true)
	}
	if source.IsCoerce() {
		target.SetCoerce(true)
	}

	// Use core convenience methods for modifier merging
	if source.DefaultValue != nil {
		target.SetDefaultValue(source.DefaultValue)
	}
	if source.DefaultFunc != nil {
		target.SetDefaultFunc(source.DefaultFunc)
	}
	if source.PrefaultValue != nil {
		target.SetPrefaultValue(source.PrefaultValue)
	}
	if source.PrefaultFunc != nil {
		target.SetPrefaultFunc(source.PrefaultFunc)
	}
	if source.Transform != nil {
		target.SetTransform(source.Transform)
	}

	// Merge checks using core's AddCheck method
	for _, check := range source.Checks {
		target.AddCheck(check)
	}

	// Merge Values map
	if len(source.Values) > 0 {
		if target.Values == nil {
			target.Values = make(map[any]struct{})
		}
		for k, v := range source.Values {
			target.Values[k] = v
		}
	}

	// Merge Bag using mapx
	if len(source.Bag) > 0 {
		if target.Bag == nil {
			target.Bag = make(map[string]any)
		}
		target.Bag = mapx.Merge(target.Bag, source.Bag)
	}

	// Preserve other important fields
	if source.Pattern != nil {
		target.Pattern = source.Pattern
	}
	if source.Constructor != nil {
		target.Constructor = source.Constructor
	}
}
