package engine

import (
	"reflect"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/cloneutil"
)

// ----------------------------------------------------------------------------
// Modifier processing
// ----------------------------------------------------------------------------

// processModifiers handles modifier processing for nil input.
// It returns early for non-nil input. The parseCore parameter is accepted
// for API compatibility but unused -- all modifier logic is in processModifiersCore.
func processModifiers[T any](
	input any,
	internals *core.ZodTypeInternals,
	expectedType core.ZodTypeCode,
	_ func(any) (any, error),
	ctx *core.ParseContext,
) (any, bool, error) {
	return processModifiersCore[T](input, internals, expectedType, ctx)
}

// ProcessNilModifiers applies the shared ordered modifier semantics for schemas
// with custom parse loops.
func ProcessNilModifiers[T any](
	input any,
	internals *core.ZodTypeInternals,
	expectedType core.ZodTypeCode,
	ctx *core.ParseContext,
) (any, bool, error) {
	return processModifiersCore[T](input, internals, expectedType, ctx)
}

// processModifiersCore contains the shared modifier logic.
func processModifiersCore[T any](
	input any,
	internals *core.ZodTypeInternals,
	expectedType core.ZodTypeCode,
	ctx *core.ParseContext,
) (any, bool, error) {
	if !isNilInput(input) {
		return nil, false, nil
	}

	if len(internals.Modifiers) > 0 {
		if r, handled, err := processOrderedModifiers[T](internals, ctx); handled || err != nil || r != nil {
			return r, handled, err
		}
	} else {
		if r, handled, err := processLegacyModifiers[T](internals, ctx); handled || err != nil || r != nil {
			return r, handled, err
		}
	}

	return processNilFallback[T](internals, expectedType, ctx)
}

func processOrderedModifiers[T any](
	internals *core.ZodTypeInternals,
	ctx *core.ParseContext,
) (any, bool, error) {
	for i := len(internals.Modifiers) - 1; i >= 0; i-- {
		r, handled, err := processModifier[T](internals.Modifiers[i], internals, ctx)
		if handled || err != nil || r != nil {
			return r, handled, err
		}
	}
	return nil, false, nil
}

func processLegacyModifiers[T any](
	internals *core.ZodTypeInternals,
	ctx *core.ParseContext,
) (any, bool, error) {
	for _, kind := range []core.ZodModifierKind{
		core.ZodModifierDefault,
		core.ZodModifierPrefault,
		core.ZodModifierNonOptional,
		core.ZodModifierOptional,
	} {
		r, handled, err := processLegacyModifier[T](kind, internals, ctx)
		if handled || err != nil || r != nil {
			return r, handled, err
		}
	}
	return nil, false, nil
}

func processModifier[T any](
	modifier core.ZodModifier,
	internals *core.ZodTypeInternals,
	ctx *core.ParseContext,
) (any, bool, error) {
	switch modifier.Kind {
	case core.ZodModifierDefault:
		return processDefaultModifier(modifier, internals, ctx)
	case core.ZodModifierPrefault:
		return processPrefaultModifier(modifier)
	case core.ZodModifierNonOptional:
		if !internals.NonOptional {
			return nil, false, nil
		}
		if reflect.TypeFor[T]().Kind() != reflect.Pointer {
			return nil, true, issues.CreateNonOptionalError(ctx)
		}
	case core.ZodModifierOptional:
		if !internals.Optional && !internals.ExactOptional {
			return nil, false, nil
		}
		return processOptionalModifier(internals, ctx)
	case core.ZodModifierNilable:
		if !internals.Nilable {
			return nil, false, nil
		}
		return processOptionalModifier(internals, ctx)
	}
	return nil, false, nil
}

func processLegacyModifier[T any](
	kind core.ZodModifierKind,
	internals *core.ZodTypeInternals,
	ctx *core.ParseContext,
) (any, bool, error) {
	switch kind {
	case core.ZodModifierDefault:
		return processLegacyDefaultModifier(internals, ctx)
	case core.ZodModifierPrefault:
		return processLegacyPrefaultModifier(internals)
	default:
		return processModifier[T](core.ZodModifier{Kind: kind}, internals, ctx)
	}
}

func processNilFallback[T any](
	internals *core.ZodTypeInternals,
	expectedType core.ZodTypeCode,
	ctx *core.ParseContext,
) (any, bool, error) {
	if reflect.TypeFor[T]().Kind() == reflect.Pointer {
		return processOptionalModifier(internals, ctx)
	}
	if expectedType == core.ZodTypeNil {
		return processOptionalModifier(internals, ctx)
	}
	if expectedType == core.ZodTypeAny || expectedType == core.ZodTypeUnknown {
		return nil, true, nil
	}
	return nil, true, issues.CreateInvalidTypeError(expectedType, nil, ctx)
}

func processDefaultModifier(
	modifier core.ZodModifier,
	internals *core.ZodTypeInternals,
	ctx *core.ParseContext,
) (any, bool, error) {
	v, ok := resolveDefaultModifier(modifier)
	if !ok {
		return nil, false, nil
	}
	if hasOverwriteCheck(internals.Checks) {
		r, err := ApplyChecks(v, internals.Checks, ctx)
		return r, true, err
	}
	return v, true, nil
}

func processPrefaultModifier(modifier core.ZodModifier) (any, bool, error) {
	if modifier.Func != nil {
		return cloneutil.Clone(modifier.Func()), false, nil
	}
	if modifier.HasValue {
		return cloneutil.Clone(modifier.Value), false, nil
	}
	return nil, false, nil
}

func processLegacyDefaultModifier(internals *core.ZodTypeInternals, ctx *core.ParseContext) (any, bool, error) {
	v := resolveDefault(internals)
	if v == nil {
		return nil, false, nil
	}
	if hasOverwriteCheck(internals.Checks) {
		r, err := ApplyChecks(v, internals.Checks, ctx)
		return r, true, err
	}
	return v, true, nil
}

func processLegacyPrefaultModifier(internals *core.ZodTypeInternals) (any, bool, error) {
	if internals.PrefaultValue != nil {
		return cloneutil.Clone(internals.PrefaultValue), false, nil
	}
	if internals.PrefaultFunc != nil {
		return cloneutil.Clone(internals.PrefaultFunc()), false, nil
	}
	return nil, false, nil
}

func resolveDefaultModifier(modifier core.ZodModifier) (any, bool) {
	if modifier.Func != nil {
		return cloneutil.Clone(modifier.Func()), true
	}
	if modifier.HasValue {
		return cloneutil.Clone(modifier.Value), true
	}
	return nil, false
}

func processOptionalModifier(internals *core.ZodTypeInternals, ctx *core.ParseContext) (any, bool, error) {
	if nc := filterNilChecks(internals.Checks); len(nc) > 0 {
		r, err := ApplyChecks[any](nil, nc, ctx)
		return r, true, err
	}
	return nil, true, nil
}

// ----------------------------------------------------------------------------
// Transform and utility helpers
// ----------------------------------------------------------------------------

// applyTransformIfPresent applies the transform function if set.
func applyTransformIfPresent(
	result any,
	internals *core.ZodTypeInternals,
	ctx *core.ParseContext,
) (any, error) {
	if internals.Transform == nil {
		return result, nil
	}
	refinementCtx := core.NewRefinementContext(ctx, result)
	transformed, err := internals.Transform(result, refinementCtx)
	if err != nil {
		return nil, err
	}
	if ctxErr := refinementCtx.Err(); ctxErr != nil {
		return nil, ctxErr
	}
	return transformed, nil
}

// filterNilChecks returns only checks applicable to nil values
// (overwrite, refine, and custom checks).
func filterNilChecks(checks []core.ZodCheck) []core.ZodCheck {
	var out []core.ZodCheck
	for _, c := range checks {
		if c == nil {
			continue
		}
		ci := c.Zod()
		if ci == nil || ci.Def == nil {
			continue
		}
		switch ci.Def.Check {
		case "overwrite", "refine", "custom":
			out = append(out, c)
		}
	}
	return out
}

// resolveDefault returns the default value for the schema, cloning it if
// it is a DefaultValue, or invoking DefaultFunc. Returns nil when neither is set.
func resolveDefault(internals *core.ZodTypeInternals) any {
	if internals.DefaultValue != nil {
		return cloneutil.Clone(internals.DefaultValue)
	}
	if internals.DefaultFunc != nil {
		return cloneutil.Clone(internals.DefaultFunc())
	}
	return nil
}
