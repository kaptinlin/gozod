// Package main provides file writing functionality for the gozodgen tool.
// This module handles the generation and writing of Go code files containing
// optimized validation schemas based on struct analysis.
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

// basicTypes is a set of basic Go type names for fast lookup
var basicTypes = map[string]bool{
	"string": true, "int": true, "int8": true, "int16": true, "int32": true, "int64": true,
	"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true,
	"float32": true, "float64": true, "bool": true, "complex64": true, "complex128": true,
}

// basicTypeConstructors maps basic Go types to their GoZod constructor calls
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

// FileWriter handles the generation and writing of Go code files
type FileWriter struct {
	outputDir    string
	packageName  string
	outputSuffix string
	templates    *template.Template
	dryRun       bool
	verbose      bool
}

// NewFileWriter creates a new file writer instance
func NewFileWriter(outputDir, packageName, outputSuffix string, dryRun, verbose bool) (*FileWriter, error) {
	// Load code generation templates
	templates, err := loadTemplates()
	if err != nil {
		return nil, fmt.Errorf("failed to load templates: %w", err)
	}

	return &FileWriter{
		outputDir:    outputDir,
		packageName:  packageName,
		outputSuffix: outputSuffix,
		templates:    templates,
		dryRun:       dryRun,
		verbose:      verbose,
	}, nil
}

// WriteGeneratedCode writes the generated code for a struct to a file
func (w *FileWriter) WriteGeneratedCode(info *GenerationInfo) error {
	// Generate the output file path based on the source file
	outputPath := w.getOutputPathFromSource(info.FilePath, info.Name)

	if w.verbose {
		fmt.Printf("Generating code for struct %s -> %s\n", info.Name, outputPath)
	}

	// Generate the code content
	content, err := w.generateCode(info)
	if err != nil {
		return fmt.Errorf("failed to generate code for %s: %w", info.Name, err)
	}

	// In dry-run mode, just print the content
	if w.dryRun {
		fmt.Printf("=== Generated code for %s ===\n%s\n", info.Name, content)
		return nil
	}

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0750); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write the file
	if err := os.WriteFile(outputPath, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write file %s: %w", outputPath, err)
	}

	if w.verbose {
		fmt.Printf("Successfully generated %s\n", outputPath)
	}

	return nil
}

// generateCode generates the Go code for a struct
func (w *FileWriter) generateCode(info *GenerationInfo) (string, error) {
	// Generate field schemas
	fieldSchemas, err := w.generateFieldSchemasForStruct(info.Fields, info.Name)
	if err != nil {
		return "", fmt.Errorf("failed to generate field schemas: %w", err)
	}

	// Determine package name
	packageName := w.packageName
	if packageName == "" {
		packageName = info.Package
	}

	// Prepare template data
	templateData := &TemplateData{
		PackageName:   packageName,
		StructName:    info.Name,
		Fields:        info.Fields,
		FieldSchemas:  fieldSchemas,
		Imports:       w.generateImports(info),
		GeneratedTime: time.Now().Format(time.RFC3339),
	}

	// Execute the main template
	var buf strings.Builder
	if err := w.templates.ExecuteTemplate(&buf, "main", templateData); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// generateImports generates the import statements needed for the generated code
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

// generateFieldSchemasForStruct generates schema code for all fields with struct context
func (w *FileWriter) generateFieldSchemasForStruct(fields []tagparser.FieldInfo, structName string) ([]FieldSchemaInfo, error) {
	fieldSchemas := make([]FieldSchemaInfo, 0, len(fields))

	for _, field := range fields {
		schemaCode, err := w.generateFieldSchemaCodeForStruct(field, structName)
		if err != nil {
			return nil, fmt.Errorf("failed to generate schema for field %s: %w", field.Name, err)
		}

		fieldSchemas = append(fieldSchemas, FieldSchemaInfo{
			FieldName:  field.JsonName,
			SchemaCode: schemaCode,
		})
	}

	return fieldSchemas, nil
}

// generateFieldSchemaCodeForStruct generates GoZod schema code for a single field with struct context
func (w *FileWriter) generateFieldSchemaCodeForStruct(field tagparser.FieldInfo, structName string) (string, error) {
	// Use AST type name for better circular reference detection
	fieldTypeName := field.TypeName
	if fieldTypeName == "" {
		fieldTypeName = field.Type.String() // fallback to reflection
	}
	// Check for special UUID case first
	if w.hasUUIDRule(field.Rules) && w.isStringType(field.Type) {
		var b strings.Builder
		b.WriteString("gozod.Uuid()")

		// Apply other validation rules (excluding UUID)
		for _, rule := range field.Rules {
			if rule.Name != "uuid" {
				validatorCode := w.generateValidatorChain(rule, field.Type)
				if validatorCode != "" {
					b.WriteString(validatorCode)
				}
			}
		}

		// Handle optional fields
		if !field.Required && !w.isPointerType(field.Type) {
			b.WriteString(".Optional()")
		}

		return b.String(), nil
	}

	// Check for special Enum case
	enumRule := w.getEnumRule(field.Rules)
	if enumRule != nil && w.isStringType(field.Type) {
		var values []string
		for _, param := range enumRule.Params {
			values = append(values, fmt.Sprintf(`"%s"`, param))
		}

		var b strings.Builder
		b.WriteString("gozod.Enum(")
		b.WriteString(strings.Join(values, ", "))
		b.WriteByte(')')

		// Apply other validation rules (excluding enum)
		for _, rule := range field.Rules {
			if rule.Name != "enum" {
				validatorCode := w.generateValidatorChain(rule, field.Type)
				if validatorCode != "" {
					b.WriteString(validatorCode)
				}
			}
		}

		// Handle optional fields
		if !field.Required && !w.isPointerType(field.Type) {
			b.WriteString(".Optional()")
		}

		return b.String(), nil
	}

	// Get base constructor for the field type using AST type name for better circular detection
	baseConstructor := w.getBaseConstructorFromTypeName(fieldTypeName, structName)

	var b strings.Builder
	b.WriteString(baseConstructor)

	// Apply validation rules in order
	for _, rule := range field.Rules {
		validatorCode := w.generateValidatorChain(rule, field.Type)
		if validatorCode != "" {
			b.WriteString(validatorCode)
		}
	}

	// Handle optional fields
	// Pointer types are always optional
	// Non-pointer, non-required fields are also optional
	if w.isPointerType(field.Type) || !field.Required {
		b.WriteString(".Optional()")
	}

	return b.String(), nil
}

// generateValidatorChain generates validator method chain for a rule
func (w *FileWriter) generateValidatorChain(rule tagparser.TagRule, fieldType reflect.Type) string {
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
			// Handle default values based on type
			value := rule.Params[0]
			// For complex JSON values, don't split by individual words
			if len(rule.Params) > 1 && !strings.HasPrefix(value, "[") && !strings.HasPrefix(value, "{") {
				// This might be a single parameter that was incorrectly split
				value = strings.Join(rule.Params, " ")
			}
			return w.generateDefaultValue(value, fieldType)
		}
	case "prefault":
		if len(rule.Params) > 0 {
			// Handle prefault values based on type
			value := rule.Params[0]
			// For complex JSON values, don't split by individual words
			if len(rule.Params) > 1 && !strings.HasPrefix(value, "[") && !strings.HasPrefix(value, "{") {
				// This might be a single parameter that was incorrectly split
				value = strings.Join(rule.Params, " ")
			}
			return w.generatePrefaultValue(value, fieldType)
		}
	case "refine":
		if len(rule.Params) > 0 {
			// Generate refine method call with custom function
			return fmt.Sprintf(".Refine(%s)", rule.Params[0])
		}
	case "check":
		if len(rule.Params) > 0 {
			// Generate custom check method call
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

// getBaseConstructorFromTypeName creates constructor from AST type name with circular reference detection
func (w *FileWriter) getBaseConstructorFromTypeName(typeName, structName string) string {
	// Handle pointer types
	if strings.HasPrefix(typeName, "*") {
		baseType := strings.TrimPrefix(typeName, "*")
		if w.isBasicTypeName(baseType) {
			return w.getBasicTypeConstructor(baseType)
		}
		// Check for circular reference
		if structName != "" && baseType == structName {
			return fmt.Sprintf("gozod.Lazy(func() gozod.ZodType[any] { return gozod.FromStruct[%s]() })", baseType)
		}
		return fmt.Sprintf("gozod.FromStruct[%s]()", baseType)
	}

	// Handle slice types
	if strings.HasPrefix(typeName, "[]") {
		elemTypeName := strings.TrimPrefix(typeName, "[]")
		// Remove pointer prefix from element type for comparison
		cleanElemType := strings.TrimPrefix(elemTypeName, "*")

		// Check for circular reference in slice element
		if structName != "" && cleanElemType == structName {
			return fmt.Sprintf("gozod.Slice(gozod.Lazy(func() gozod.ZodType[any] { return gozod.FromStruct[%s]() }))", cleanElemType)
		}

		elemConstructor := w.getBaseConstructorFromTypeName(elemTypeName, structName)
		return fmt.Sprintf("gozod.Slice(%s)", elemConstructor)
	}

	// Handle map types
	if strings.HasPrefix(typeName, "map[") {
		// Extract value type from map[K]V
		if idx := strings.LastIndex(typeName, "]"); idx != -1 && idx < len(typeName)-1 {
			valueTypeName := typeName[idx+1:]
			cleanValueType := strings.TrimPrefix(valueTypeName, "*")

			// Check for circular reference in map value
			if structName != "" && cleanValueType == structName {
				return fmt.Sprintf("gozod.Record(gozod.Lazy(func() gozod.ZodType[any] { return gozod.FromStruct[%s]() }))", cleanValueType)
			}

			valueConstructor := w.getBaseConstructorFromTypeName(valueTypeName, structName)
			return fmt.Sprintf("gozod.Record(%s)", valueConstructor)
		}
		return "gozod.Record(gozod.Any())"
	}

	// Handle basic types
	if w.isBasicTypeName(typeName) {
		return w.getBasicTypeConstructor(typeName)
	}

	// Handle special types
	if typeName == "time.Time" {
		return "gozod.Time()"
	}

	// Handle struct types
	// Check for circular reference
	if structName != "" && typeName == structName {
		return fmt.Sprintf("gozod.Lazy(func() gozod.ZodType[any] { return gozod.FromStruct[%s]() })", typeName)
	}

	// Custom struct type
	if typeName != "unknown" {
		return fmt.Sprintf("gozod.FromStruct[%s]()", typeName)
	}

	return "gozod.Any()"
}

// isBasicTypeName checks if a type name is a basic Go type
func (w *FileWriter) isBasicTypeName(typeName string) bool {
	return basicTypes[typeName]
}

// getBasicTypeConstructor returns the GoZod constructor for basic Go types
func (w *FileWriter) getBasicTypeConstructor(typeName string) string {
	if constructor, exists := basicTypeConstructors[typeName]; exists {
		return constructor
	}
	return "gozod.Any()"
}

// isPointerType checks if a type is a pointer type
func (w *FileWriter) isPointerType(t reflect.Type) bool {
	return t.Kind() == reflect.Ptr
}

// isStringType checks if a type is string or *string
func (w *FileWriter) isStringType(t reflect.Type) bool {
	if t.Kind() == reflect.String {
		return true
	}
	if t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.String {
		return true
	}
	return false
}

// generateDefaultValue generates Go code for default values based on type.
func (w *FileWriter) generateDefaultValue(value string, fieldType reflect.Type) string {
	return w.generateTypedValue("Default", value, fieldType)
}

// generatePrefaultValue generates Go code for prefault values based on type.
func (w *FileWriter) generatePrefaultValue(value string, fieldType reflect.Type) string {
	return w.generateTypedValue("Prefault", value, fieldType)
}

// generateTypedValue generates a method call (.Default or .Prefault) with
// the value formatted according to the field's Go type.
func (w *FileWriter) generateTypedValue(method, value string, fieldType reflect.Type) string {
	switch fieldType.Kind() { //nolint:exhaustive // only special-cased types need distinct formatting
	case reflect.String:
		return fmt.Sprintf(`.%s("%s")`, method, value)
	case reflect.Slice, reflect.Array:
		return generateSliceValue(method, value, fieldType)
	case reflect.Map:
		return generateMapValue(method, value, fieldType)
	case reflect.Ptr:
		return w.generateTypedValue(method, value, fieldType.Elem())
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

	var jsonResult []interface{}
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

	var jsonResult map[string]interface{}
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
			return fmt.Sprintf(".%s(map[string]interface{}{%s})", method, strings.Join(items, ", "))
		default:
			// Fall through to fallback
		}
	}

	return fmt.Sprintf(".%s(%s)", method, value)
}

// hasUUIDRule checks if field has UUID validation rule
func (w *FileWriter) hasUUIDRule(rules []tagparser.TagRule) bool {
	for _, rule := range rules {
		if rule.Name == "uuid" {
			return true
		}
	}
	return false
}

// getEnumRule returns the enum rule if it exists
func (w *FileWriter) getEnumRule(rules []tagparser.TagRule) *tagparser.TagRule {
	for _, rule := range rules {
		if rule.Name == "enum" {
			return &rule
		}
	}
	return nil
}

// getOutputPathFromSource generates output file path based on source file location
func (w *FileWriter) getOutputPathFromSource(sourceFilePath, structName string) string {
	// Get directory from source file
	sourceDir := filepath.Dir(sourceFilePath)

	// Convert struct name to snake_case for file name
	fileName := toSnakeCase(structName) + w.outputSuffix

	return filepath.Join(sourceDir, fileName)
}

// TemplateData contains data passed to code generation templates
type TemplateData struct {
	PackageName   string
	StructName    string
	Fields        []tagparser.FieldInfo
	FieldSchemas  []FieldSchemaInfo
	Imports       []string
	GeneratedTime string
}

// Type definitions are now in main.go

// loadTemplates loads the code generation templates
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

// firstLowerCase converts the first character of a string to lowercase and handles special cases
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

// receiverName generates a valid Go receiver variable name
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

// toSnakeCase converts CamelCase to snake_case properly
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
