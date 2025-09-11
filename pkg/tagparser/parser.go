// Package tagparser provides shared tag parsing functionality for GoZod
// Extracted from types/struct.go to enable reuse by cmd/gozodgen and other components
package tagparser

import (
	"reflect"
	"strings"
)

// TagParser handles gozod tag parsing with configurable tag name
type TagParser struct {
	tagName string // Tag name to parse (default: "gozod")
}

// New creates a new TagParser with default "gozod" tag name
func New() *TagParser {
	return &TagParser{
		tagName: "gozod",
	}
}

// NewWithTagName creates a new TagParser with custom tag name
func NewWithTagName(tagName string) *TagParser {
	return &TagParser{
		tagName: tagName,
	}
}

// FieldInfo represents parsed information about a struct field
type FieldInfo struct {
	Name     string       // Go field name
	Type     reflect.Type // Go field type
	TypeName string       // AST-based type name string for circular reference detection
	JsonName string       // JSON field name (from json tag or field name)
	GozodTag string       // Raw gozod tag value
	Rules    []TagRule    // Parsed validation rules
	Required bool         // Whether field is required (has "required" rule)
	Optional bool         // Whether field should be optional (pointer type without required)
	Nilable  bool         // Whether field has "nilable" rule
}

// TagRule represents a single validation rule parsed from a tag
type TagRule struct {
	Name   string   // Rule name (e.g., "min", "max", "email")
	Params []string // Rule parameters (e.g., ["2"] for "min=2")
}

// ParseStructTags parses all gozod tags in a struct type and returns field information
func (p *TagParser) ParseStructTags(structType reflect.Type) ([]FieldInfo, error) {
	var fields []FieldInfo

	// Handle pointer to struct
	if structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
	}

	// Ensure it's a struct
	if structType.Kind() != reflect.Struct {
		return nil, nil // Not a struct, no fields to parse
	}

	// Iterate through all exported fields
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Skip fields with gozod:"-" tag
		gozodTag := field.Tag.Get(p.tagName)
		if gozodTag == "-" {
			continue
		}

		// Parse field information
		fieldInfo := FieldInfo{
			Name:     field.Name,
			Type:     field.Type,
			JsonName: getJSONFieldName(field),
			GozodTag: gozodTag,
		}

		// Parse validation rules from tag
		if gozodTag != "" {
			rules, err := p.ParseTagString(gozodTag)
			if err != nil {
				return nil, err
			}
			fieldInfo.Rules = rules

			// Check for special rules
			fieldInfo.Required = hasRule(rules, "required")
			fieldInfo.Nilable = hasRule(rules, "nilable")
		}

		// Determine if field should be optional
		fieldInfo.Optional = shouldBeOptional(field, fieldInfo.Required, fieldInfo.Nilable)

		fields = append(fields, fieldInfo)
	}

	return fields, nil
}

// ParseTagString parses a single tag string into validation rules
func (p *TagParser) ParseTagString(tag string) ([]TagRule, error) {
	var rules []TagRule

	if tag == "" {
		return rules, nil
	}

	// Split by comma, handling escaped commas
	parts := parseTagParts(tag)

	for _, part := range parts {
		rule := parseTagRule(strings.TrimSpace(part))
		if rule.Name != "" {
			rules = append(rules, rule)
		}
	}

	return rules, nil
}

// parseTagParts splits tag string by commas, handling escapes and JSON structures
func parseTagParts(tag string) []string {
	var parts []string
	var current strings.Builder
	var bracketDepth int
	var braceDepth int
	var inQuotes bool
	var quoteChar rune
	escaped := false

	for i, char := range tag {
		switch char {
		case '\\':
			if i+1 < len(tag) {
				// Handle escape sequences
				current.WriteRune(char)
				escaped = true
			} else {
				current.WriteRune(char)
			}
		case '"', '\'':
			if !escaped {
				if !inQuotes {
					// Start of quoted string
					inQuotes = true
					quoteChar = char
				} else if char == quoteChar {
					// End of quoted string
					inQuotes = false
				}
			}
			current.WriteRune(char)
			escaped = false
		case '[':
			if !inQuotes && !escaped {
				bracketDepth++
			}
			current.WriteRune(char)
			escaped = false
		case ']':
			if !inQuotes && !escaped {
				bracketDepth--
			}
			current.WriteRune(char)
			escaped = false
		case '{':
			if !inQuotes && !escaped {
				braceDepth++
			}
			current.WriteRune(char)
			escaped = false
		case '}':
			if !inQuotes && !escaped {
				braceDepth--
			}
			current.WriteRune(char)
			escaped = false
		case ',':
			if !escaped && !inQuotes && bracketDepth == 0 && braceDepth == 0 {
				// Unescaped comma outside quotes and brackets - end current part
				parts = append(parts, current.String())
				current.Reset()
			} else {
				current.WriteRune(char)
			}
			escaped = false
		default:
			current.WriteRune(char)
			escaped = false
		}
	}

	// Add final part
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// parseTagRule parses a single rule part into TagRule
func parseTagRule(part string) TagRule {
	if part == "" {
		return TagRule{}
	}

	// Check if rule has parameters (contains =)
	if idx := strings.Index(part, "="); idx != -1 {
		name := strings.TrimSpace(part[:idx])
		paramStr := strings.TrimSpace(part[idx+1:])

		// Parse parameters - split by space for multi-value params
		var params []string
		if paramStr != "" {
			// Handle quoted parameters
			switch {
			case strings.HasPrefix(paramStr, "'") && strings.HasSuffix(paramStr, "'"):
				// Single quoted parameter
				unquoted := paramStr[1 : len(paramStr)-1]
				params = []string{unescapeString(unquoted)}
			case strings.Contains(paramStr, " "):
				// Multiple space-separated parameters
				params = strings.Fields(paramStr)
			default:
				// Single parameter
				params = []string{paramStr}
			}
		}

		return TagRule{
			Name:   name,
			Params: params,
		}
	}

	// Rule without parameters
	return TagRule{
		Name:   strings.TrimSpace(part),
		Params: nil,
	}
}

// unescapeString handles escape sequences in tag parameters
func unescapeString(s string) string {
	s = strings.ReplaceAll(s, "\\,", ",")
	s = strings.ReplaceAll(s, "\\n", "\n")
	s = strings.ReplaceAll(s, "\\t", "\t")
	s = strings.ReplaceAll(s, "\\'", "'")
	s = strings.ReplaceAll(s, "\\\\", "\\")
	return s
}

// getJSONFieldName extracts JSON field name from struct field
func getJSONFieldName(field reflect.StructField) string {
	jsonTag := field.Tag.Get("json")
	if jsonTag == "" {
		return field.Name
	}

	// Handle json:"-" (skip field) and json:",omitempty" etc.
	if jsonTag == "-" {
		return field.Name // Use Go field name as fallback
	}

	// Extract name before first comma
	if idx := strings.Index(jsonTag, ","); idx != -1 {
		jsonName := strings.TrimSpace(jsonTag[:idx])
		if jsonName != "" {
			return jsonName
		}
	} else {
		return strings.TrimSpace(jsonTag)
	}

	return field.Name
}

// hasRule checks if a rule with given name exists in rules slice
func hasRule(rules []TagRule, name string) bool {
	for _, rule := range rules {
		if rule.Name == name {
			return true
		}
	}
	return false
}

// shouldBeOptional determines if a field should be optional based on type and rules
func shouldBeOptional(field reflect.StructField, required, nilable bool) bool {
	// If explicitly required, not optional
	if required {
		return false
	}

	// If nilable tag is present, it's optional
	if nilable {
		return true
	}

	// Pointer types are optional by default (unless required)
	if field.Type.Kind() == reflect.Ptr {
		return true
	}

	// Non-pointer types are not optional by default
	return false
}
