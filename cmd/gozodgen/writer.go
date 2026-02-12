package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"
	"time"

	"github.com/go-json-experiment/json"

	"github.com/kaptinlin/gozod/pkg/tagparser"
)

// basicTypes is a set of basic Go type names for fast lookup.
var basicTypes = map[string]bool{
	"string": true, "int": true, "int8": true, "int16": true, "int32": true, "int64": true,
	"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true,
	"float32": true, "float64": true, "bool": true, "complex64": true, "complex128": true,
}

// basicTypeConstructors maps basic Go types to their GoZod constructor calls.
var basicTypeConstructors = map[string]string{
	"string":     "gozod.String()",
	"int":        "gozod.Int()",
	"int8":       "gozod.Int8()",
	"int16":      "gozod.Int16()",
	"int32":      "gozod.Int32()",
	"int64":      "gozod.Int64()",
	"uint":       "gozod.Uint()",
	"uint8":      "gozod.Uint8()",
	"uint16":     "gozod.Uint16()",
	"uint32":     "gozod.Uint32()",
	"uint64":     "gozod.Uint64()",
	"float32":    "gozod.Float32()",
	"float64":    "gozod.Float64()",
	"bool":       "gozod.Bool()",
	"complex64":  "gozod.Complex64()",
	"complex128": "gozod.Complex128()",
}

// FileWriter handles the generation and writing of Go code files.
type FileWriter struct {
	outputDir    string
	packageName  string
	outputSuffix string
	templates    *template.Template
	dryRun       bool
	verbose      bool
}

// NewFileWriter creates a new FileWriter instance.
func NewFileWriter(outputDir, packageName, outputSuffix string, dryRun, verbose bool) (*FileWriter, error) {
	tmpl, err := loadTemplates()
	if err != nil {
		return nil, fmt.Errorf("load templates: %w", err)
	}

	return &FileWriter{
		outputDir:    outputDir,
		packageName:  packageName,
		outputSuffix: outputSuffix,
		templates:    tmpl,
		dryRun:       dryRun,
		verbose:      verbose,
	}, nil
}

// WriteGeneratedCode writes the generated code for a struct to a file.
func (w *FileWriter) WriteGeneratedCode(info *GenerationInfo) error {
	outputPath := w.outputPath(info.FilePath, info.Name)

	if w.verbose {
		fmt.Printf("Generating code for struct %s -> %s\n", info.Name, outputPath)
	}

	content, err := w.generateCode(info)
	if err != nil {
		return fmt.Errorf("generate code for %s: %w", info.Name, err)
	}

	if w.dryRun {
		fmt.Printf("=== Generated code for %s ===\n%s\n", info.Name, content)
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0750); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	if err := os.WriteFile(outputPath, []byte(content), 0600); err != nil {
		return fmt.Errorf("write file %s: %w", outputPath, err)
	}

	if w.verbose {
		fmt.Printf("Generated %s\n", outputPath)
	}
	return nil
}

// generateCode generates the Go code for a struct.
func (w *FileWriter) generateCode(info *GenerationInfo) (string, error) {
	fieldSchemas, err := w.generateFieldSchemas(info.Fields, info.Name)
	if err != nil {
		return "", fmt.Errorf("generate field schemas: %w", err)
	}

	pkgName := w.packageName
	if pkgName == "" {
		pkgName = info.Package
	}

	data := &TemplateData{
		PackageName:   pkgName,
		StructName:    info.Name,
		Fields:        info.Fields,
		FieldSchemas:  fieldSchemas,
		Imports:       w.generateImports(info),
		GeneratedTime: time.Now().Format(time.RFC3339),
	}

	var buf strings.Builder
	if err := w.templates.ExecuteTemplate(&buf, "main", data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	return buf.String(), nil
}

// generateImports generates the import statements needed for the generated code.
func (w *FileWriter) generateImports(info *GenerationInfo) []string {
	imports := make(map[string]bool)

	// Always include gozod core
	imports["github.com/kaptinlin/gozod"] = true

	// Check if any transformations are used (strings package needed)
	needsStrings := false
	needsRegexp := false
	needsCore := false

	for _, field := range info.Fields {
		// Check for time.Time fields
		if strings.Contains(field.Type.String(), "time.Time") {
			imports["time"] = true
		}

		// Note: Lazy evaluation now uses gozod.ZodType instead of core.ZodType
		// so we don't need to import core for lazy evaluation

		// Check for validators that need specific imports
		for _, rule := range field.Rules {
			switch rule.Name {
			case "trim", "lowercase", "uppercase":
				needsStrings = true
			case "regex":
				needsRegexp = true
			case "uuid":
				// UUID validation doesn't need additional imports
			case "url":
				// Only URL validation needs net/url, not email
				imports["net/url"] = true
			case "ipv4", "ipv6":
				// IP validation might need net package
				imports["net"] = true
			case "refine", "check":
				// These might need core types if using complex validation
				needsCore = true
			}
		}
	}

	if needsStrings {
		imports["strings"] = true
	}
	if needsRegexp {
		imports["regexp"] = true
	}
	if needsCore {
		imports["github.com/kaptinlin/gozod/core"] = true
	}

	// Convert map to sorted slice
	result := make([]string, 0, len(imports))
	for imp := range imports {
		result = append(result, imp)
	}

	return result
}

// generateFieldSchemas generates schema code for all fields.
func (w *FileWriter) generateFieldSchemas(fields []tagparser.FieldInfo, structName string) ([]FieldSchemaInfo, error) {
	schemas := make([]FieldSchemaInfo, 0, len(fields))
	for _, field := range fields {
		code, err := w.generateFieldSchemaCode(field, structName)
		if err != nil {
			return nil, fmt.Errorf("generate schema for field %s: %w", field.Name, err)
		}
		schemas = append(schemas, FieldSchemaInfo{
			FieldName:  field.JSONName,
			SchemaCode: code,
		})
	}
	return schemas, nil
}

// generateFieldSchemaCode generates GoZod schema code for a single field.
func (w *FileWriter) generateFieldSchemaCode(field tagparser.FieldInfo, structName string) (string, error) {
	typeName := field.TypeName
	if typeName == "" {
		typeName = field.Type.String()
	}

	// UUID special case
	if hasUUIDRule(field.Rules) && isStringType(field.Type) {
		var b strings.Builder
		b.WriteString("gozod.Uuid()")
		for _, rule := range field.Rules {
			if rule.Name != "uuid" {
				if code := generateValidatorChain(rule, field.Type); code != "" {
					b.WriteString(code)
				}
			}
		}
		if !field.Required && !isPointerType(field.Type) {
			b.WriteString(".Optional()")
		}
		return b.String(), nil
	}

	// Enum special case
	if rule := findEnumRule(field.Rules); rule != nil && isStringType(field.Type) {
		values := make([]string, 0, len(rule.Params))
		for _, param := range rule.Params {
			values = append(values, fmt.Sprintf(`"%s"`, param))
		}
		var b strings.Builder
		b.WriteString("gozod.Enum(")
		b.WriteString(strings.Join(values, ", "))
		b.WriteByte(')')
		for _, r := range field.Rules {
			if r.Name != "enum" {
				if code := generateValidatorChain(r, field.Type); code != "" {
					b.WriteString(code)
				}
			}
		}
		if !field.Required && !isPointerType(field.Type) {
			b.WriteString(".Optional()")
		}
		return b.String(), nil
	}

	// General case
	var b strings.Builder
	b.WriteString(baseConstructor(typeName, structName))
	for _, rule := range field.Rules {
		if code := generateValidatorChain(rule, field.Type); code != "" {
			b.WriteString(code)
		}
	}
	if isPointerType(field.Type) || !field.Required {
		b.WriteString(".Optional()")
	}
	return b.String(), nil
}

// generateValidatorChain returns the validator method chain for a rule.
func generateValidatorChain(rule tagparser.TagRule, fieldType reflect.Type) string {
	switch rule.Name {
	case "required":
		// Required is handled by the optional logic
		return ""
	case "min":
		if len(rule.Params) > 0 {
			return fmt.Sprintf(".Min(%s)", rule.Params[0])
		}
	case "max":
		if len(rule.Params) > 0 {
			return fmt.Sprintf(".Max(%s)", rule.Params[0])
		}
	case "email":
		return ".Email()"
	case "url":
		return ".URL()"
	case "uuid":
		// UUID fields should use the UUID constructor instead of String + constraint
		return ""
	case "ipv4":
		return ".IPv4()"
	case "ipv6":
		return ".IPv6()"
	case "gt":
		if len(rule.Params) > 0 {
			return fmt.Sprintf(".Gt(%s)", rule.Params[0])
		}
	case "gte":
		if len(rule.Params) > 0 {
			return fmt.Sprintf(".Gte(%s)", rule.Params[0])
		}
	case "lt":
		if len(rule.Params) > 0 {
			return fmt.Sprintf(".Lt(%s)", rule.Params[0])
		}
	case "lte":
		if len(rule.Params) > 0 {
			return fmt.Sprintf(".Lte(%s)", rule.Params[0])
		}
	case "enum":
		// Enum should use the Enum constructor, not a method on String
		return ""
	case "regex":
		if len(rule.Params) > 0 {
			// Escape the regex pattern for double quotes
			escapedPattern := strings.ReplaceAll(rule.Params[0], "\\", "\\\\")
			escapedPattern = strings.ReplaceAll(escapedPattern, "\"", "\\\"")
			return fmt.Sprintf(".Regex(regexp.MustCompile(\"%s\"))", escapedPattern)
		}
	case "default":
		if len(rule.Params) > 0 {
			value := rule.Params[0]
			if len(rule.Params) > 1 && !strings.HasPrefix(value, "[") && !strings.HasPrefix(value, "{") {
				value = strings.Join(rule.Params, " ")
			}
			return generateDefaultValue(value, fieldType)
		}
	case "prefault":
		if len(rule.Params) > 0 {
			value := rule.Params[0]
			if len(rule.Params) > 1 && !strings.HasPrefix(value, "[") && !strings.HasPrefix(value, "{") {
				value = strings.Join(rule.Params, " ")
			}
			return generatePrefaultValue(value, fieldType)
		}
	case "refine":
		if len(rule.Params) > 0 {
			return fmt.Sprintf(".Refine(%s)", rule.Params[0])
		}
	case "check":
		if len(rule.Params) > 0 {
			return fmt.Sprintf(".Check(%s)", rule.Params[0])
		}
	case "trim":
		return ".Trim()"
	case "lowercase":
		return ".ToLowerCase()"
	case "uppercase":
		return ".ToUpperCase()"
	case "nilable":
		return ".Nilable()"
	case "time":
		// Time validation is handled by the Time() constructor
		return ""
	}
	return ""
}

// baseConstructor returns the GoZod constructor for a type name with circular reference detection.
func baseConstructor(typeName, structName string) string {
	if base, ok := strings.CutPrefix(typeName, "*"); ok {
		if basicTypes[base] {
			return basicTypeConstructor(base)
		}
		if structName != "" && base == structName {
			return fmt.Sprintf("gozod.Lazy(func() gozod.ZodType[any] { return gozod.FromStruct[%s]() })", base)
		}
		return fmt.Sprintf("gozod.FromStruct[%s]()", base)
	}

	if elem, ok := strings.CutPrefix(typeName, "[]"); ok {
		clean := strings.TrimPrefix(elem, "*")
		if structName != "" && clean == structName {
			return fmt.Sprintf("gozod.Slice(gozod.Lazy(func() gozod.ZodType[any] { return gozod.FromStruct[%s]() }))", clean)
		}
		return fmt.Sprintf("gozod.Slice(%s)", baseConstructor(elem, structName))
	}

	if strings.HasPrefix(typeName, "map[") {
		if idx := strings.LastIndex(typeName, "]"); idx != -1 && idx < len(typeName)-1 {
			valType := typeName[idx+1:]
			clean := strings.TrimPrefix(valType, "*")
			if structName != "" && clean == structName {
				return fmt.Sprintf("gozod.Record(gozod.Lazy(func() gozod.ZodType[any] { return gozod.FromStruct[%s]() }))", clean)
			}
			return fmt.Sprintf("gozod.Record(%s)", baseConstructor(valType, structName))
		}
		return "gozod.Record(gozod.Any())"
	}

	if basicTypes[typeName] {
		return basicTypeConstructor(typeName)
	}
	if typeName == "time.Time" {
		return "gozod.Time()"
	}
	if structName != "" && typeName == structName {
		return fmt.Sprintf("gozod.Lazy(func() gozod.ZodType[any] { return gozod.FromStruct[%s]() })", typeName)
	}
	if typeName != "unknown" {
		return fmt.Sprintf("gozod.FromStruct[%s]()", typeName)
	}
	return "gozod.Any()"
}

// basicTypeConstructor returns the GoZod constructor for a basic Go type.
func basicTypeConstructor(typeName string) string {
	if c, ok := basicTypeConstructors[typeName]; ok {
		return c
	}
	return "gozod.Any()"
}

// isPointerType reports whether t is a pointer type.
func isPointerType(t reflect.Type) bool {
	return t.Kind() == reflect.Pointer
}

// isStringType reports whether t is string or *string.
func isStringType(t reflect.Type) bool {
	return t.Kind() == reflect.String ||
		(t.Kind() == reflect.Pointer && t.Elem().Kind() == reflect.String)
}

// generateDefaultValue returns a .Default(...) call formatted for the field type.
func generateDefaultValue(value string, fieldType reflect.Type) string {
	return generateTypedValue("Default", value, fieldType)
}

// generatePrefaultValue returns a .Prefault(...) call formatted for the field type.
func generatePrefaultValue(value string, fieldType reflect.Type) string {
	return generateTypedValue("Prefault", value, fieldType)
}

// generateTypedValue returns a method call (.Default or .Prefault) with
// the value formatted according to the field's Go type.
func generateTypedValue(method, value string, fieldType reflect.Type) string {
	switch fieldType.Kind() { //nolint:exhaustive // only special-cased types need distinct formatting
	case reflect.String:
		return fmt.Sprintf(`.%s("%s")`, method, value)
	case reflect.Slice, reflect.Array:
		return generateSliceValue(method, value, fieldType)
	case reflect.Map:
		return generateMapValue(method, value, fieldType)
	case reflect.Pointer:
		return generateTypedValue(method, value, fieldType.Elem())
	default:
		return fmt.Sprintf(".%s(%s)", method, value)
	}
}

// generateSliceValue generates a Go slice literal for .Default() or .Prefault() calls.
func generateSliceValue(method, value string, fieldType reflect.Type) string {
	value = strings.TrimSpace(value)
	if !strings.HasPrefix(value, "[") || !strings.HasSuffix(value, "]") {
		return fmt.Sprintf(".%s(%s)", method, value)
	}

	var jsonResult []any
	if err := json.Unmarshal([]byte(value), &jsonResult); err == nil {
		elemType := fieldType.Elem()
		switch elemType.Kind() { //nolint:exhaustive // only common JSON-decodable types need special handling
		case reflect.String:
			items := make([]string, 0, len(jsonResult))
			for _, item := range jsonResult {
				if str, ok := item.(string); ok {
					items = append(items, fmt.Sprintf(`"%s"`, str))
				}
			}
			return fmt.Sprintf(".%s([]string{%s})", method, strings.Join(items, ", "))
		case reflect.Int:
			items := make([]string, 0, len(jsonResult))
			for _, item := range jsonResult {
				if num, ok := item.(float64); ok {
					items = append(items, fmt.Sprintf("%d", int(num)))
				}
			}
			return fmt.Sprintf(".%s([]int{%s})", method, strings.Join(items, ", "))
		case reflect.Float64:
			items := make([]string, 0, len(jsonResult))
			for _, item := range jsonResult {
				if num, ok := item.(float64); ok {
					items = append(items, fmt.Sprintf("%g", num))
				}
			}
			return fmt.Sprintf(".%s([]float64{%s})", method, strings.Join(items, ", "))
		case reflect.Bool:
			items := make([]string, 0, len(jsonResult))
			for _, item := range jsonResult {
				if b, ok := item.(bool); ok {
					items = append(items, fmt.Sprintf("%t", b))
				}
			}
			return fmt.Sprintf(".%s([]bool{%s})", method, strings.Join(items, ", "))
		default:
			// Fall through to fallback
		}
	}

	return fmt.Sprintf(".%s(%s)", method, value)
}

// generateMapValue generates a Go map literal for .Default() or .Prefault() calls.
func generateMapValue(method, value string, fieldType reflect.Type) string {
	value = strings.TrimSpace(value)
	if !strings.HasPrefix(value, "{") || !strings.HasSuffix(value, "}") {
		return fmt.Sprintf(".%s(%s)", method, value)
	}

	var jsonResult map[string]any
	if err := json.Unmarshal([]byte(value), &jsonResult); err == nil {
		switch fieldType.Elem().Kind() { //nolint:exhaustive // only string and interface map values are JSON-decodable
		case reflect.String:
			items := make([]string, 0, len(jsonResult))
			for k, v := range jsonResult {
				if str, ok := v.(string); ok {
					items = append(items, fmt.Sprintf(`"%s": "%s"`, k, str))
				}
			}
			return fmt.Sprintf(".%s(map[string]string{%s})", method, strings.Join(items, ", "))
		case reflect.Interface:
			items := make([]string, 0, len(jsonResult))
			for k, v := range jsonResult {
				switch val := v.(type) {
				case string:
					items = append(items, fmt.Sprintf(`"%s": "%s"`, k, val))
				case float64:
					items = append(items, fmt.Sprintf(`"%s": %g`, k, val))
				case bool:
					items = append(items, fmt.Sprintf(`"%s": %t`, k, val))
				default:
					items = append(items, fmt.Sprintf(`"%s": %v`, k, val))
				}
			}
			return fmt.Sprintf(".%s(map[string]any{%s})", method, strings.Join(items, ", "))
		default:
			// Fall through to fallback
		}
	}

	return fmt.Sprintf(".%s(%s)", method, value)
}

// hasUUIDRule reports whether rules contain a UUID validation rule.
func hasUUIDRule(rules []tagparser.TagRule) bool {
	for _, rule := range rules {
		if rule.Name == "uuid" {
			return true
		}
	}
	return false
}

// findEnumRule returns the enum rule if present, or nil.
func findEnumRule(rules []tagparser.TagRule) *tagparser.TagRule {
	for _, rule := range rules {
		if rule.Name == "enum" {
			return &rule
		}
	}
	return nil
}

// outputPath generates the output file path based on source file location.
func (w *FileWriter) outputPath(sourceFilePath, structName string) string {
	dir := filepath.Dir(sourceFilePath)
	return filepath.Join(dir, toSnakeCase(structName)+w.outputSuffix)
}

// TemplateData contains data passed to code generation templates.
type TemplateData struct {
	PackageName   string
	StructName    string
	Fields        []tagparser.FieldInfo
	FieldSchemas  []FieldSchemaInfo
	Imports       []string
	GeneratedTime string
}

// loadTemplates loads the code generation templates.
func loadTemplates() (*template.Template, error) {
	// Define the main template for generated code
	mainTemplate := `// Code generated by gozodgen. DO NOT EDIT.
// Generated at: {{.GeneratedTime}}

package {{.PackageName}}

import (
{{- range .Imports}}
	"{{.}}"
{{- end}}
)

// Schema returns a pre-built gozod schema for {{.StructName}}
// This generated function provides zero-reflection validation with optimal performance
func ({{.StructName | receiverName}} {{.StructName}}) Schema() *gozod.ZodStruct[{{.StructName}}, {{.StructName}}] {
	return gozod.Struct[{{.StructName}}](gozod.StructSchema{
{{- range .FieldSchemas}}
		"{{.FieldName}}": {{.SchemaCode}},
{{- end}}
	})
}
`

	// Create template with custom functions
	tmpl := template.New("main").Funcs(template.FuncMap{
		"firstLower":   firstLowerCase,
		"receiverName": receiverName,
	})

	// Parse the main template
	tmpl, err := tmpl.Parse(mainTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse main template: %w", err)
	}

	return tmpl, nil
}

// firstLowerCase converts the first character of a string to lowercase and handles special cases.
func firstLowerCase(s string) string {
	if s == "" {
		return s
	}

	// Handle generic types like APIResponse[T any]
	if strings.Contains(s, "[") {
		s = strings.Split(s, "[")[0]
	}

	// Convert to lowercase and handle acronyms
	if len(s) == 1 {
		return strings.ToLower(s)
	}

	// Handle acronyms like "APIResponse" -> "apiResponse"
	if len(s) > 1 && s[1] >= 'A' && s[1] <= 'Z' {
		// Find where the acronym ends
		i := 0
		for i < len(s) && s[i] >= 'A' && s[i] <= 'Z' {
			i++
		}
		if i > 1 && i < len(s) {
			// Keep all but last letter of acronym uppercase, then continue
			return strings.ToLower(s[:i-1]) + strings.ToLower(string(s[i-1])) + s[i:]
		}
	}

	// Standard case: just lowercase first character
	return strings.ToLower(s[:1]) + s[1:]
}

// receiverName generates a valid Go receiver variable name.
func receiverName(s string) string {
	if s == "" {
		return "x"
	}

	// Handle generic types like APIResponse[T any]
	if strings.Contains(s, "[") {
		s = strings.Split(s, "[")[0]
	}

	// Convert to a valid receiver name by taking first letter of each word
	var result strings.Builder

	// Special handling for acronyms and camel case
	prevUpper := false
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i == 0 || !prevUpper {
				result.WriteRune(r + ('a' - 'A')) // Convert to lowercase
			}
			prevUpper = true
		} else {
			prevUpper = false
			if i == 0 {
				result.WriteRune(r)
			}
		}
	}

	name := result.String()
	if name == "" {
		return "x"
	}

	// Handle reserved words or make sure it's a valid identifier
	if name == "type" || name == "interface" || name == "struct" {
		return name + "Val"
	}

	return name
}

// toSnakeCase converts CamelCase to snake_case properly.
func toSnakeCase(s string) string {
	if s == "" {
		return s
	}

	var result strings.Builder
	result.Grow(len(s) * 2) // Pre-allocate with enough space

	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			// Add underscore before uppercase letters (except the first character)
			if i > 0 {
				result.WriteRune('_')
			}
			// Convert to lowercase
			result.WriteRune(r - 'A' + 'a')
		} else {
			// Keep lowercase letters and other characters as-is
			result.WriteRune(r)
		}
	}

	return result.String()
}
