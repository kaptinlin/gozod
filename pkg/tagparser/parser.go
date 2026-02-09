// Package tagparser provides shared tag parsing functionality for GoZod.
// Extracted from types/struct.go to enable reuse by cmd/gozodgen and other components.
package tagparser

import (
	"reflect"
	"strings"
)

// unescaper handles escape sequences in tag parameters using a single-pass replacer.
var unescaper = strings.NewReplacer(
	`\,`, ",",
	`\n`, "\n",
	`\t`, "\t",
	`\'`, "'",
	`\\`, `\`,
)

// TagParser handles gozod tag parsing with configurable tag name.
type TagParser struct {
	tagName string
}

// New creates a TagParser with the default "gozod" tag name.
func New() *TagParser {
	return &TagParser{tagName: "gozod"}
}

// NewWithTagName creates a TagParser with a custom tag name.
func NewWithTagName(tagName string) *TagParser {
	return &TagParser{tagName: tagName}
}

// FieldInfo represents parsed information about a struct field.
type FieldInfo struct {
	Name     string       // Go field name
	Type     reflect.Type // Go field type
	TypeName string       // AST-based type name for circular reference detection
	JsonName string       // JSON field name (from json tag or field name)
	GozodTag string       // Raw gozod tag value
	Rules    []TagRule    // Parsed validation rules
	Required bool         // Whether field has "required" rule
	Optional bool         // Whether field is optional (pointer without required)
	Nilable  bool         // Whether field has "nilable" rule
}

// TagRule represents a single validation rule parsed from a tag.
type TagRule struct {
	Name   string   // Rule name (e.g., "min", "max", "email")
	Params []string // Rule parameters (e.g., ["2"] for "min=2")
}

// ParseStructTags parses all gozod tags in a struct type and returns field information.
func (p *TagParser) ParseStructTags(structType reflect.Type) ([]FieldInfo, error) {
	// Handle pointer to struct
	if structType.Kind() == reflect.Pointer {
		structType = structType.Elem()
	}

	if structType.Kind() != reflect.Struct {
		return nil, nil
	}

	fields := make([]FieldInfo, 0, structType.NumField())

	for i := range structType.NumField() {
		field := structType.Field(i)

		if !field.IsExported() {
			continue
		}

		tag := field.Tag.Get(p.tagName)
		if tag == "-" {
			continue
		}

		info := FieldInfo{
			Name:     field.Name,
			Type:     field.Type,
			JsonName: jsonFieldName(field),
			GozodTag: tag,
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

		info.Optional = isOptional(field, info.Required, info.Nilable)
		fields = append(fields, info)
	}

	return fields, nil
}

// ParseTagString parses a single tag string into validation rules.
func (p *TagParser) ParseTagString(tag string) ([]TagRule, error) {
	if tag == "" {
		return []TagRule{}, nil
	}

	parts := splitTagParts(tag)
	rules := make([]TagRule, 0, len(parts))

	for _, part := range parts {
		if rule := parseRule(strings.TrimSpace(part)); rule.Name != "" {
			rules = append(rules, rule)
		}
	}

	return rules, nil
}

// splitTagParts splits a tag string by commas, respecting escapes,
// quotes, brackets, and braces.
func splitTagParts(tag string) []string {
	parts := make([]string, 0, strings.Count(tag, ",")+1)
	var current strings.Builder
	var bracketDepth, braceDepth int
	var inQuotes bool
	var quoteChar rune
	escaped := false

	for i, ch := range tag {
		switch ch {
		case '\\':
			if i+1 < len(tag) {
				current.WriteRune(ch)
				escaped = true
			} else {
				current.WriteRune(ch)
			}
		case '"', '\'':
			if !escaped {
				if !inQuotes {
					inQuotes = true
					quoteChar = ch
				} else if ch == quoteChar {
					inQuotes = false
				}
			}
			current.WriteRune(ch)
			escaped = false
		case '[':
			if !inQuotes && !escaped {
				bracketDepth++
			}
			current.WriteRune(ch)
			escaped = false
		case ']':
			if !inQuotes && !escaped {
				bracketDepth--
			}
			current.WriteRune(ch)
			escaped = false
		case '{':
			if !inQuotes && !escaped {
				braceDepth++
			}
			current.WriteRune(ch)
			escaped = false
		case '}':
			if !inQuotes && !escaped {
				braceDepth--
			}
			current.WriteRune(ch)
			escaped = false
		case ',':
			if !escaped && !inQuotes && bracketDepth == 0 && braceDepth == 0 {
				parts = append(parts, current.String())
				current.Reset()
			} else {
				current.WriteRune(ch)
			}
			escaped = false
		default:
			current.WriteRune(ch)
			escaped = false
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// parseRule parses a single rule string (e.g., "min=2") into a TagRule.
func parseRule(part string) TagRule {
	if part == "" {
		return TagRule{}
	}

	name, paramStr, hasParams := strings.Cut(part, "=")
	if !hasParams {
		return TagRule{Name: strings.TrimSpace(part)}
	}

	name = strings.TrimSpace(name)
	paramStr = strings.TrimSpace(paramStr)

	var params []string
	if paramStr != "" {
		switch {
		case strings.HasPrefix(paramStr, "'") && strings.HasSuffix(paramStr, "'"):
			params = []string{unescaper.Replace(paramStr[1 : len(paramStr)-1])}
		case strings.Contains(paramStr, " "):
			params = strings.Fields(paramStr)
		default:
			params = []string{paramStr}
		}
	}

	return TagRule{Name: name, Params: params}
}

// jsonFieldName extracts the JSON field name from a struct field's json tag.
// Falls back to the Go field name when no usable json name is present.
func jsonFieldName(field reflect.StructField) string {
	jsonTag := field.Tag.Get("json")
	if jsonTag == "" || jsonTag == "-" {
		return field.Name
	}

	name, _, _ := strings.Cut(jsonTag, ",")
	if name = strings.TrimSpace(name); name != "" {
		return name
	}

	return field.Name
}

// hasRule reports whether a rule with the given name exists in rules.
func hasRule(rules []TagRule, name string) bool {
	for _, r := range rules {
		if r.Name == name {
			return true
		}
	}
	return false
}

// isOptional reports whether a field should be treated as optional
// based on its type and parsed rules.
func isOptional(field reflect.StructField, required, nilable bool) bool {
	if required {
		return false
	}
	return nilable || field.Type.Kind() == reflect.Pointer
}
