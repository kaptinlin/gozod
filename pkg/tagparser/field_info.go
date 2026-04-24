package tagparser

import (
	"reflect"
	"slices"
	"strings"
)

// HasRules reports whether the field has any parsed tag rules.
func (f FieldInfo) HasRules() bool {
	return len(f.Rules) > 0
}

// HasRule reports whether a parsed rule with the given name exists.
func (f FieldInfo) HasRule(name string) bool {
	return hasRule(f.Rules, name)
}

// IsPointerType reports whether the field type is a pointer.
func (f FieldInfo) IsPointerType() bool {
	return f.Type != nil && f.Type.Kind() == reflect.Pointer
}

// HasCoerceRule reports whether the field enables coercion.
func (f FieldInfo) HasCoerceRule() bool {
	return f.HasRule("coerce")
}

// EffectiveTypeName returns the best available type name for codegen.
func (f FieldInfo) EffectiveTypeName() string {
	if f.TypeName != "" {
		return f.TypeName
	}
	if f.Type != nil {
		return f.Type.String()
	}
	return ""
}

// HasSchemaSpec reports whether the field has tag-derived schema semantics.
// Runtime/codegen consumers should prefer helper methods on [FieldInfo]
// instead of open-coding logic from the raw fields.
func (f FieldInfo) HasSchemaSpec() bool {
	return f.Required || f.HasRules()
}

// UsesTimeImport reports whether codegen should import time for this field.
func (f FieldInfo) UsesTimeImport() bool {
	return f.Type != nil && strings.Contains(f.Type.String(), "time.Time")
}

// EnumRule returns the enum rule if present.
func (f FieldInfo) EnumRule() *TagRule {
	for _, rule := range f.Rules {
		if rule.Name == "enum" {
			return new(rule)
		}
	}
	return nil
}

// RulesExcept returns all parsed rules except those whose names are excluded.
func (f FieldInfo) RulesExcept(names ...string) []TagRule {
	if len(names) == 0 {
		return append([]TagRule(nil), f.Rules...)
	}
	filtered := make([]TagRule, 0, len(f.Rules))
	for _, rule := range f.Rules {
		if !slices.Contains(names, rule.Name) {
			filtered = append(filtered, rule)
		}
	}
	return filtered
}

// ValidationRules returns tag rules that should become schema modifiers
// or checks at runtime/codegen. Structural rules are handled elsewhere.
func (f FieldInfo) ValidationRules() []TagRule {
	return f.RulesExcept("required", "optional", "coerce")
}

// ValidationRulesExcept returns validation rules except those explicitly excluded.
func (f FieldInfo) ValidationRulesExcept(names ...string) []TagRule {
	excluded := append([]string{"required", "optional", "coerce"}, names...)
	return f.RulesExcept(excluded...)
}

// RequiredImports returns the non-gozod imports needed to generate this field.
func (f FieldInfo) RequiredImports() []string {
	imports := make([]string, 0, 5)
	add := func(path string) {
		if path == "" || slices.Contains(imports, path) {
			return
		}
		imports = append(imports, path)
	}

	if f.UsesTimeImport() {
		add("time")
	}
	if f.NeedsStringsImport() {
		add("strings")
	}
	if f.NeedsRegexpImport() {
		add("regexp")
	}
	if f.NeedsNetURLImport() {
		add("net/url")
	}
	if f.NeedsNetImport() {
		add("net")
	}
	if f.NeedsCoreImport() {
		add("github.com/kaptinlin/gozod/core")
	}

	return imports
}

// NeedsStringsImport reports whether codegen should import strings.
func (f FieldInfo) NeedsStringsImport() bool {
	return f.HasRule("trim") || f.HasRule("lowercase") || f.HasRule("uppercase")
}

// NeedsRegexpImport reports whether codegen should import regexp.
func (f FieldInfo) NeedsRegexpImport() bool {
	return f.HasRule("regex")
}

// NeedsNetURLImport reports whether codegen should import net/url.
func (f FieldInfo) NeedsNetURLImport() bool {
	return f.HasRule("url")
}

// NeedsNetImport reports whether codegen should import net.
func (f FieldInfo) NeedsNetImport() bool {
	return f.HasRule("ipv4") || f.HasRule("ipv6")
}

// NeedsCoreImport reports whether codegen should import core.
func (f FieldInfo) NeedsCoreImport() bool {
	return f.HasRule("refine") || f.HasRule("check")
}

// IsUUIDStringField reports whether this field should use the UUID constructor.
func (f FieldInfo) IsUUIDStringField() bool {
	return f.Type != nil && f.HasRule("uuid") && f.Type.Kind() == reflect.String
}

// IsEnumStringField reports whether this field should use the Enum constructor.
func (f FieldInfo) IsEnumStringField() bool {
	return f.Type != nil && f.EnumRule() != nil && f.Type.Kind() == reflect.String
}

// NeedsOptionalModifier reports whether generated schema code should
// append .Optional() for this field.
func (f FieldInfo) NeedsOptionalModifier() bool {
	return !f.Required && !f.IsPointerType()
}

// NeedsGeneratedOptional reports whether generated schema code should
// append .Optional() after applying all rules.
func (f FieldInfo) NeedsGeneratedOptional() bool {
	return f.IsPointerType() || !f.Required
}

// NeedsPointerNilable reports whether a pointer-backed schema should
// become nilable during struct-schema derivation.
func (f FieldInfo) NeedsPointerNilable() bool {
	return f.IsPointerType() && !f.Required
}

// NeedsPointerOptional reports whether pointer field rules should add
// Optional() at the end of parsed tag application.
func (f FieldInfo) NeedsPointerOptional() bool {
	return f.IsPointerType() && !f.Required
}
