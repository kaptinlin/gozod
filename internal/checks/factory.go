// Package checks provides validation check factories for the GoZod validation
// library. It implements Zod v4-compatible check creation with JSON Schema
// metadata support.
package checks

import (
	"maps"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/internal/utils"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// Type definitions

// ZodCheckCustomDef defines a custom validation constraint.
type ZodCheckCustomDef struct {
	core.ZodCheckDef
	Type   string         // check type identifier
	Params map[string]any // additional parameters
	Fn     any            // RefineFn or CheckFn
	FnType string         // "refine" or "check"
}

// ZodCheckCustomInternals contains custom check internal state.
type ZodCheckCustomInternals struct {
	core.ZodCheckInternals
	Def  *ZodCheckCustomDef
	Issc *core.ZodIssueBase
	Bag  map[string]any
}

// ZodCheckCustom represents a custom validation check.
type ZodCheckCustom struct {
	Internals *ZodCheckCustomInternals
}

// Zod returns the internal check structure for execution.
func (c *ZodCheckCustom) Zod() *core.ZodCheckInternals {
	return &c.Internals.ZodCheckInternals
}

// ZodCheckOverwriteDef defines a value transformation constraint.
type ZodCheckOverwriteDef struct {
	core.ZodCheckDef
	Transform func(any) any
}

// ZodCheckOverwriteInternals contains overwrite check internal state.
type ZodCheckOverwriteInternals struct {
	core.ZodCheckInternals
	Def *ZodCheckOverwriteDef
}

// ZodCheckOverwrite represents a check that replaces the payload value.
type ZodCheckOverwrite struct {
	Internals *ZodCheckOverwriteInternals
}

// Zod returns the internal check structure for execution.
func (c *ZodCheckOverwrite) Zod() *core.ZodCheckInternals {
	return &c.Internals.ZodCheckInternals
}

// ZodCheckPropertyDef defines a property validation constraint.
type ZodCheckPropertyDef struct {
	core.ZodCheckDef
	Property string
	Schema   core.ZodSchema
}

// ZodCheckPropertyInternals contains property check internal state.
type ZodCheckPropertyInternals struct {
	core.ZodCheckInternals
	Def  *ZodCheckPropertyDef
	Issc *core.ZodIssueBase
}

// ZodCheckProperty validates a specific object property against a schema.
type ZodCheckProperty struct {
	Internals *ZodCheckPropertyInternals
}

// Zod returns the internal check structure for execution.
func (c *ZodCheckProperty) Zod() *core.ZodCheckInternals {
	return &c.Internals.ZodCheckInternals
}

// Parameter normalization

// NormalizeCheckParams standardizes check parameters from various input formats.
// It accepts a string shorthand or core.SchemaParams.
func NormalizeCheckParams(params ...any) *core.CheckParams {
	if len(params) == 0 {
		return nil
	}

	switch p := params[0].(type) {
	case string:
		return &core.CheckParams{Error: p}
	case core.SchemaParams:
		if s, ok := p.Error.(string); ok {
			return &core.CheckParams{Error: s}
		}
	}

	return nil
}

// ApplyCheckParams applies normalized parameters to a check definition.
func ApplyCheckParams(def *core.ZodCheckDef, cp *core.CheckParams) {
	if cp != nil && cp.Error != "" {
		def.Error = new(core.ZodErrorMap(func(_ core.ZodRawIssue) string {
			return cp.Error
		}))
	}
}

// ApplySchemaParamsToCheck applies SchemaParams to a check definition.
func ApplySchemaParamsToCheck(def *core.ZodCheckDef, sp *core.SchemaParams) {
	if sp == nil {
		return
	}
	if sp.Error != nil {
		if em, ok := utils.ToErrorMap(sp.Error); ok {
			def.Error = em
		}
	}
	if sp.Abort {
		def.Abort = true
	}
}

// Constructors

// NewCustom creates a custom validation check with user-defined logic.
func NewCustom[T any](fn any, args ...any) *ZodCheckCustom {
	cp := utils.NormalizeCustomParams(utils.FirstParam(args...))

	def := &ZodCheckCustomDef{
		ZodCheckDef: core.ZodCheckDef{
			Check: "custom",
			Abort: cp.Abort,
		},
		Fn:     fn,
		Params: make(map[string]any),
	}

	switch fn.(type) {
	case core.ZodRefineFn[T], func(T) bool:
		def.FnType = "refine"
	case core.ZodCheckFn, func(*core.ParsePayload):
		def.FnType = "check"
	default:
		def.FnType = "refine"
	}

	if cp.Error != nil {
		if em, ok := utils.ToErrorMap(cp.Error); ok {
			def.Error = em
		}
	}
	if len(cp.Params) > 0 {
		maps.Copy(def.Params, cp.Params)
	}
	if len(cp.Path) > 0 {
		def.Params["path"] = cp.Path
	}

	internals := &ZodCheckCustomInternals{
		ZodCheckInternals: core.ZodCheckInternals{
			Def:  &def.ZodCheckDef,
			When: cp.When,
		},
		Def:  def,
		Issc: &core.ZodIssueBase{},
		Bag:  make(map[string]any),
	}
	internals.Check = func(payload *core.ParsePayload) {
		executeCustomCheck(payload, internals)
	}

	return &ZodCheckCustom{Internals: internals}
}

// NewZodCheckOverwrite creates a check that overwrites input with a transformed value.
func NewZodCheckOverwrite(transform func(any) any, args ...any) *ZodCheckOverwrite {
	sp := utils.NormalizeParams(utils.FirstParam(args...))

	def := &ZodCheckOverwriteDef{
		ZodCheckDef: core.ZodCheckDef{Check: "overwrite"},
		Transform:   transform,
	}
	ApplySchemaParamsToCheck(&def.ZodCheckDef, sp)

	internals := &ZodCheckOverwriteInternals{
		ZodCheckInternals: core.ZodCheckInternals{
			Def: &def.ZodCheckDef,
		},
		Def: def,
	}
	internals.Check = func(payload *core.ParsePayload) {
		payload.SetValue(transform(payload.Value()))
	}

	return &ZodCheckOverwrite{Internals: internals}
}

// NewProperty creates a property validation check that validates
// input[property] against the provided schema.
func NewProperty(property string, schema core.ZodSchema, args ...any) *ZodCheckProperty {
	sp := utils.NormalizeParams(utils.FirstParam(args...))

	def := &ZodCheckPropertyDef{
		ZodCheckDef: core.ZodCheckDef{Check: "property"},
		Property:    property,
		Schema:      schema,
	}
	ApplySchemaParamsToCheck(&def.ZodCheckDef, sp)

	internals := &ZodCheckPropertyInternals{
		ZodCheckInternals: core.ZodCheckInternals{
			Def: &def.ZodCheckDef,
		},
		Def:  def,
		Issc: &core.ZodIssueBase{},
	}
	internals.Check = func(payload *core.ParsePayload) {
		executePropertyCheck(payload, internals)
	}

	return &ZodCheckProperty{Internals: internals}
}

// Bag and constraint utilities

// ensureBag initializes and returns the schema's Bag map.
func ensureBag(schema any) map[string]any {
	s, ok := schema.(interface{ Internals() *core.ZodTypeInternals })
	if !ok {
		return nil
	}
	internals := s.Internals()
	if internals.Bag == nil {
		internals.Bag = make(map[string]any)
	}
	return internals.Bag
}

// SetBagProperty sets a property in the schema's bag for JSON Schema generation.
func SetBagProperty(schema any, key string, value any) {
	if bag := ensureBag(schema); bag != nil {
		bag[key] = value
	}
}

// mergeConstraint merges a constraint into the schema's bag with conflict resolution.
func mergeConstraint(schema any, key string, value any, merge func(old, new any) any) {
	bag := ensureBag(schema)
	if bag == nil {
		return
	}
	existing, ok := bag[key]
	if !ok {
		bag[key] = value
		return
	}
	bag[key] = merge(existing, value)
}

// mergeMinimumConstraint merges minimum constraint, choosing the stricter value.
func mergeMinimumConstraint(schema any, value any, inclusive bool) {
	key := "minimum"
	conflict := "exclusiveMinimum"
	if !inclusive {
		key = "exclusiveMinimum"
		conflict = "minimum"
	}

	mergeConstraint(schema, key, value, func(old, new any) any {
		if utils.CompareValues(new, old) > 0 {
			return new
		}
		return old
	})

	removeConflictingBound(schema, conflict, value, func(cmp int) bool { return cmp >= 0 })
}

// mergeMaximumConstraint merges maximum constraint, choosing the stricter value.
func mergeMaximumConstraint(schema any, value any, inclusive bool) {
	key := "maximum"
	conflict := "exclusiveMaximum"
	if !inclusive {
		key = "exclusiveMaximum"
		conflict = "maximum"
	}

	mergeConstraint(schema, key, value, func(old, new any) any {
		if utils.CompareValues(new, old) < 0 {
			return new
		}
		return old
	})

	removeConflictingBound(schema, conflict, value, func(cmp int) bool { return cmp <= 0 })
}

// removeConflictingBound removes a conflicting bound from the bag when the
// comparison between value and the existing bound satisfies shouldRemove.
func removeConflictingBound(schema any, key string, value any, shouldRemove func(int) bool) {
	bag := ensureBag(schema)
	if bag == nil {
		return
	}
	existing, ok := bag[key]
	if !ok {
		return
	}
	if shouldRemove(utils.CompareValues(value, existing)) {
		delete(bag, key)
	}
}

// Issue path utilities

// PrefixIssues prepends a path segment to all issues for nested validation.
func PrefixIssues(path any, iss []core.ZodRawIssue) []core.ZodRawIssue {
	if slicex.IsEmpty(iss) {
		return iss
	}
	for i := range iss {
		newPath := make([]any, 1+len(iss[i].Path))
		newPath[0] = path
		copy(newPath[1:], iss[i].Path)
		iss[i].Path = newPath
	}
	return iss
}

// Custom check execution

// executeCustomCheck dispatches to the appropriate refine or check function.
func executeCustomCheck(payload *core.ParsePayload, ci *ZodCheckCustomInternals) {
	defer func() {
		if r := recover(); r != nil {
			handleRefineResult(false, payload, payload.Value(), ci)
		}
	}()

	switch ci.Def.FnType {
	case "refine":
		executeRefine(payload, ci)
	case "check":
		executeCheckFn(payload, ci)
	default:
		handleRefineResult(false, payload, payload.Value(), ci)
	}
}

// executeRefine dispatches typed refine functions and reports validation results.
func executeRefine(payload *core.ParsePayload, ci *ZodCheckCustomInternals) {
	v := payload.Value()
	switch fn := ci.Def.Fn.(type) {
	case func([]any) bool:
		arr, ok := v.([]any)
		if !ok {
			handleRefineResult(false, payload, v, ci)
			return
		}
		handleRefineResult(fn(arr), payload, v, ci)
	case func(string) bool:
		s, ok := v.(string)
		if !ok {
			handleRefineResult(false, payload, v, ci)
			return
		}
		handleRefineResult(fn(s), payload, v, ci)
	case func(map[string]any) bool:
		m, ok := v.(map[string]any)
		if !ok {
			handleRefineResult(false, payload, v, ci)
			return
		}
		handleRefineResult(fn(m), payload, v, ci)
	case func(any) bool:
		handleRefineResult(fn(v), payload, v, ci)
	case core.ZodRefineFn[string]:
		s, ok := v.(string)
		if !ok {
			handleRefineResult(false, payload, v, ci)
			return
		}
		handleRefineResult(fn(s), payload, v, ci)
	case core.ZodRefineFn[map[string]any]:
		m, ok := v.(map[string]any)
		if !ok {
			handleRefineResult(false, payload, v, ci)
			return
		}
		handleRefineResult(fn(m), payload, v, ci)
	case core.ZodRefineFn[any]:
		handleRefineResult(fn(v), payload, v, ci)
	default:
		handleRefineResult(false, payload, v, ci)
	}
}

// executeCheckFn runs a check-style function that modifies the payload directly.
func executeCheckFn(payload *core.ParsePayload, ci *ZodCheckCustomInternals) {
	switch fn := ci.Def.Fn.(type) {
	case core.ZodCheckFn:
		fn(payload)
	case func(*core.ParsePayload):
		fn(payload)
	default:
		handleRefineResult(false, payload, payload.Value(), ci)
	}
}

// handleRefineResult creates an error issue when a refine function returns false.
func handleRefineResult(ok bool, payload *core.ParsePayload, input any, ci *ZodCheckCustomInternals) {
	if ok {
		return
	}

	path := resolvePath(payload, ci)

	msg := resolveErrorMessage(ci, input, path)
	if msg == "" {
		msg = "Invalid input"
	}

	props := map[string]any{
		"origin":   "custom",
		"continue": !ci.Def.Abort,
	}
	if len(ci.Def.Params) > 0 {
		props["params"] = ci.Def.Params
	}

	issue := issues.CreateCustomIssue(msg, props, input)
	issue.Input = input
	issue.Inst = ci
	payload.AddIssueWithPath(issue, path)
}

// resolvePath returns the issue path, using a custom path override if configured.
func resolvePath(payload *core.ParsePayload, ci *ZodCheckCustomInternals) []any {
	pp := payload.Path()
	path := make([]any, len(pp))
	copy(path, pp)

	cp, ok := ci.Def.Params["path"]
	if !ok {
		return path
	}
	switch v := cp.(type) {
	case []any:
		return v
	case []string:
		out := make([]any, len(v))
		for i, s := range v {
			out[i] = s
		}
		return out
	case string:
		return []any{v}
	}
	return path
}

// resolveErrorMessage returns the custom error message if configured.
func resolveErrorMessage(ci *ZodCheckCustomInternals, input any, path []any) string {
	if ci.Def.Error == nil {
		return ""
	}
	em := *ci.Def.Error
	return em(core.ZodRawIssue{
		Code:  core.Custom,
		Input: input,
		Path:  path,
	})
}

// Property check execution

// executePropertyCheck validates a specific property of an object.
func executePropertyCheck(payload *core.ParsePayload, pi *ZodCheckPropertyInternals) {
	obj, ok := payload.Value().(map[string]any)
	if !ok {
		return
	}

	val, exists := obj[pi.Def.Property]
	if !exists {
		return
	}

	_, err := pi.Def.Schema.ParseAny(val)
	if err == nil {
		return
	}

	path := append(payload.Path(), pi.Def.Property)

	msg := err.Error()
	if pi.Def.Error != nil {
		em := *pi.Def.Error
		msg = em(core.ZodRawIssue{
			Code:  core.Custom,
			Input: val,
			Path:  path,
		})
	}

	props := map[string]any{
		"origin":   "property",
		"property": pi.Def.Property,
		"continue": !pi.Def.Abort,
	}

	issue := issues.CreateCustomIssue(msg, props, val)
	issue.Input = val
	issue.Inst = pi
	payload.AddIssueWithPath(issue, path)
}
