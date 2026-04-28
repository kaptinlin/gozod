// Package tagparser provides shared tag parsing functionality
// for GoZod. Extracted from types/struct.go to enable reuse by
// cmd/gozodgen and other components.
package tagparser

import (
	"errors"
	"reflect"
	"slices"
	"strings"
)

const defaultTagName = "gozod"

// unescaper handles escape sequences in tag parameters
// using a single-pass replacer.
var unescaper = strings.NewReplacer(
	`\,`, ",",
	`\n`, "\n",
	`\t`, "\t",
	`\'`, "'",
	`\\`, `\`,
)

// TagRule represents a single validation rule parsed from a tag.
type TagRule struct {
	Name   string   // e.g., "min", "max", "email"
	Params []string // e.g., ["2"] for "min=2"
}

// JSONField represents the resolved JSON contract for a struct field.
type JSONField struct {
	Name string
	Skip bool
}

// ErrTypeMustBeStruct indicates that the input type is neither a
// struct nor a pointer to struct.
var ErrTypeMustBeStruct = errors.New("tagparser: type must be a struct or pointer to struct")

// FieldInfo represents parsed information about a struct field.
type FieldInfo struct {
	Name     string       // Go field name
	Type     reflect.Type // field type
	TypeName string       // AST type name for circular reference detection
	JSONName string       // from json tag, or Go field name
	GoZodTag string       // raw gozod tag value
	Rules    []TagRule    // parsed tag rules before helper-level filtering
	Required bool         // whether the parsed rules include "required"
	Optional bool         // derived optionality for parsed-tag callers
	Nilable  bool         // whether the parsed rules include "nilable"
}

// TagParser handles gozod tag parsing with configurable tag name.
type TagParser struct {
	tagName string
}

// New creates a [TagParser] with the default "gozod" tag name.
func New() *TagParser {
	return &TagParser{tagName: defaultTagName}
}

// NewWithTagName creates a [TagParser] with a custom tag name.
func NewWithTagName(name string) *TagParser {
	if strings.TrimSpace(name) == "" {
		name = defaultTagName
	}
	return &TagParser{tagName: name}
}

// ParseStructTags parses all gozod tags in a struct type
// and returns [FieldInfo] for each exported field.
//
// Non-struct input returns an empty slice and no error. Use
// [ParseStructTagsStrict] when non-struct input should fail.
func (p *TagParser) ParseStructTags(typ reflect.Type) ([]FieldInfo, error) {
	structType, ok := normalizedStructType(typ)
	if !ok {
		return []FieldInfo{}, nil
	}

	return p.parseStructFields(structType)
}

// ParseStructTagsStrict parses all gozod tags in a struct type
// and returns [ErrTypeMustBeStruct] for non-struct input.
func (p *TagParser) ParseStructTagsStrict(typ reflect.Type) ([]FieldInfo, error) {
	structType, ok := normalizedStructType(typ)
	if !ok {
		return nil, ErrTypeMustBeStruct
	}

	return p.parseStructFields(structType)
}

func (p *TagParser) parseStructFields(typ reflect.Type) ([]FieldInfo, error) {
	fields := make([]FieldInfo, 0, typ.NumField())

	for f := range typ.Fields() {
		if !f.IsExported() {
			continue
		}

		tag := f.Tag.Get(p.tagName)
		if tag == "-" {
			continue
		}

		jsonField := JSONFieldName(f)
		if jsonField.Skip {
			continue
		}

		info := FieldInfo{
			Name:     f.Name,
			Type:     f.Type,
			JSONName: jsonField.Name,
			GoZodTag: tag,
		}

		if tag != "" {
			rules, err := p.ParseTagString(tag)
			if err != nil {
				return nil, err
			}
			info.Rules = rules
			info.Required = hasRule(rules, "required")
			info.Nilable = hasRule(rules, "nilable")
		}

		info.Optional = isOptional(f, info.Required, info.Nilable)
		fields = append(fields, info)
	}

	return fields, nil
}

func normalizedStructType(typ reflect.Type) (reflect.Type, bool) {
	if typ == nil {
		return nil, false
	}
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	return typ, typ.Kind() == reflect.Struct
}

// ParseTagString parses a single tag string into [TagRule] values.
func (p *TagParser) ParseTagString(tag string) ([]TagRule, error) {
	if tag == "" {
		return []TagRule{}, nil
	}

	parts := splitParts(tag)
	rules := make([]TagRule, 0, len(parts))

	for _, part := range parts {
		if rule := parseRule(strings.TrimSpace(part)); rule.Name != "" {
			rules = append(rules, rule)
		}
	}

	return rules, nil
}

// splitParts splits a tag string by commas, respecting escapes,
// quotes, brackets, and braces.
func splitParts(tag string) []string {
	parts := make([]string, 0, strings.Count(tag, ",")+1)
	var buf strings.Builder
	var brackets, braces int
	var quoted bool
	var quote rune
	escaped := false

	for i, ch := range tag {
		if escaped {
			buf.WriteRune(ch)
			escaped = false
			continue
		}

		switch ch {
		case '\\':
			buf.WriteRune(ch)
			escaped = i+1 < len(tag)
		case '"', '\'':
			if !quoted {
				quoted = true
				quote = ch
			} else if ch == quote {
				quoted = false
			}
			buf.WriteRune(ch)
		case '[':
			if !quoted {
				brackets++
			}
			buf.WriteRune(ch)
		case ']':
			if !quoted {
				brackets--
			}
			buf.WriteRune(ch)
		case '{':
			if !quoted {
				braces++
			}
			buf.WriteRune(ch)
		case '}':
			if !quoted {
				braces--
			}
			buf.WriteRune(ch)
		case ',':
			if !quoted && brackets == 0 && braces == 0 {
				parts = append(parts, buf.String())
				buf.Reset()
			} else {
				buf.WriteRune(ch)
			}
		default:
			buf.WriteRune(ch)
		}
	}

	if buf.Len() > 0 {
		parts = append(parts, buf.String())
	}

	return parts
}

// parseRule parses a single rule string (e.g. "min=2")
// into a [TagRule].
func parseRule(part string) TagRule {
	if part == "" {
		return TagRule{}
	}

	name, raw, ok := strings.Cut(part, "=")
	if !ok {
		return TagRule{Name: strings.TrimSpace(part)}
	}

	name = strings.TrimSpace(name)
	raw = strings.TrimSpace(raw)

	var params []string
	if raw != "" {
		if quoted, ok := strings.CutPrefix(raw, "'"); ok {
			if quoted, ok := strings.CutSuffix(quoted, "'"); ok {
				params = []string{unescaper.Replace(quoted)}
			} else if strings.Contains(raw, " ") {
				params = strings.Fields(raw)
			} else {
				params = []string{raw}
			}
		} else if strings.Contains(raw, " ") {
			params = strings.Fields(raw)
		} else {
			params = []string{raw}
		}
	}

	return TagRule{Name: name, Params: params}
}

// JSONFieldName resolves the JSON field name and skip marker for a struct field.
func JSONFieldName(f reflect.StructField) JSONField {
	tag := f.Tag.Get("json")
	if tag == "" {
		return JSONField{Name: f.Name}
	}
	if tag == "-" {
		return JSONField{Skip: true}
	}

	name, _, _ := strings.Cut(tag, ",")
	name = strings.TrimSpace(name)
	if name == "" {
		name = f.Name
	}
	return JSONField{Name: name}
}

// hasRule reports whether a rule with the given name exists.
func hasRule(rules []TagRule, name string) bool {
	return slices.ContainsFunc(rules, func(r TagRule) bool {
		return r.Name == name
	})
}

// isOptional reports whether a field should be treated as optional
// based on its type and parsed rules.
func isOptional(f reflect.StructField, required, nilable bool) bool {
	if required {
		return false
	}
	return nilable || f.Type.Kind() == reflect.Pointer
}
