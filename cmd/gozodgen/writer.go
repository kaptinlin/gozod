package main

import (
	"errors"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"text/template"

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
	methodName   string
	fieldNameTag string
	providers    *generatedProviderPlan
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
		methodName:   defaultMethodName,
		fieldNameTag: defaultFieldNameTag,
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
	if err := writeFileAtomic(outputPath, []byte(content), 0o600); err != nil {
		return fmt.Errorf("write file %s: %w", outputPath, err)
	}

	if w.verbose {
		fmt.Printf("Generated %s\n", outputPath)
	}
	return nil
}

func writeFileAtomic(path string, content []byte, mode os.FileMode) error {
	dir := filepath.Dir(path)
	temp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	defer func() { _ = os.Remove(tempPath) }()

	if err := temp.Chmod(mode); err != nil {
		return errors.Join(err, temp.Close())
	}
	if _, err := temp.Write(content); err != nil {
		return errors.Join(err, temp.Close())
	}
	if err := temp.Sync(); err != nil {
		return errors.Join(err, temp.Close())
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return os.Rename(tempPath, path)
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
		PackageName:      pkgName,
		StructName:       info.Name,
		MethodName:       w.methodName,
		FieldNameTagCall: w.fieldNameTagCall(),
		Fields:           info.Fields,
		FieldSchemas:     fieldSchemas,
		Imports:          w.generateImports(info),
	}

	var buf strings.Builder
	if err := w.templates.ExecuteTemplate(&buf, "main", data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	formatted, err := format.Source([]byte(buf.String()))
	if err != nil {
		return "", fmt.Errorf("format generated code: %w", err)
	}
	return string(formatted), nil
}

// generateImports generates the import statements needed for the generated code.
func (w *FileWriter) generateImports(info *GenerationInfo) []string {
	imports := make(map[string]bool)

	// Always include gozod core
	imports["github.com/kaptinlin/gozod"] = true

	for _, field := range info.Fields {
		if field.HasCoerceRule() {
			imports["github.com/kaptinlin/gozod/coerce"] = true
		}
		for _, imp := range field.RequiredImports() {
			imports[imp] = true
		}
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
	for i := range fields {
		field := &fields[i]
		code, err := w.generateFieldSchemaCode(field, structName)
		if err != nil {
			return nil, fmt.Errorf("generate schema for field %s: %w", field.Name, err)
		}
		schemas = append(schemas, FieldSchemaInfo{
			FieldName:  field.FieldKey,
			SchemaCode: code,
		})
	}
	return schemas, nil
}

// generateFieldSchemaCode generates GoZod schema code for a single field.
func (w *FileWriter) generateFieldSchemaCode(field *tagparser.FieldInfo, structName string) (string, error) {
	if field == nil || field.Type == nil {
		name, typeName := "<nil>", "<nil>"
		if field != nil {
			name, typeName = field.Name, field.EffectiveTypeName()
		}
		return "", fmt.Errorf("field %s has unknown type %q", name, typeName)
	}
	typeName := field.EffectiveTypeName()
	fieldPlan, err := tagparser.CompileFieldPlan(field)
	if err != nil {
		return "", err
	}

	// String-format constructors mirror runtime tag application: the first
	// format rule replaces the base string schema, then later modifiers apply.
	_, coerces := compiledOperand(fieldPlan, tagparser.RuleCoerce)
	if constructor, consumedRule := stringFormatConstructor(field, fieldPlan); constructor != "" && !coerces {
		var b strings.Builder
		b.WriteString(constructor)
		if fieldPlan.GeneratedOptional == tagparser.OptionalPlacementBeforeOperations {
			b.WriteString(".Optional()")
		}
		for _, operation := range fieldPlan.OperationsExcept(consumedRule) {
			code, err := renderValidatorChain(operation, field.Type)
			if err != nil {
				return "", fmt.Errorf("render %s for field %s: %w", operation.Name, field.Name, err)
			}
			b.WriteString(code)
		}
		if fieldPlan.GeneratedOptional == tagparser.OptionalPlacementAfterOperations {
			b.WriteString(".Optional()")
		}
		return b.String(), nil
	}

	// Enum special case
	if operand, ok := compiledOperand(fieldPlan, tagparser.RuleEnum); ok && isStringFieldType(field.Type) {
		typed, ok := operand.([]any)
		if !ok {
			return "", fmt.Errorf("render enum operand for field %s", field.Name)
		}
		values := make([]string, 0, len(typed))
		for _, value := range typed {
			param, ok := value.(string)
			if !ok {
				return "", fmt.Errorf("render enum operand for field %s", field.Name)
			}
			values = append(values, fmt.Sprintf("%q", param))
		}
		var b strings.Builder
		b.WriteString("gozod.Enum(")
		b.WriteString(strings.Join(values, ", "))
		b.WriteByte(')')
		if fieldPlan.GeneratedOptional == tagparser.OptionalPlacementBeforeOperations {
			b.WriteString(".Optional()")
		}
		for _, operation := range fieldPlan.OperationsExcept("enum") {
			code, err := renderValidatorChain(operation, field.Type)
			if err != nil {
				return "", fmt.Errorf("render %s for field %s: %w", operation.Name, field.Name, err)
			}
			b.WriteString(code)
		}
		if fieldPlan.GeneratedOptional == tagparser.OptionalPlacementAfterOperations {
			b.WriteString(".Optional()")
		}
		return b.String(), nil
	}

	// General case
	var b strings.Builder
	var base string
	if _, coerce := compiledOperand(fieldPlan, tagparser.RuleCoerce); coerce {
		base, err = coercionConstructor(typeName)
	} else {
		base, err = w.baseConstructor(typeName, structName)
	}
	if err != nil {
		return "", fmt.Errorf("render type %s for field %s: %w", typeName, field.Name, err)
	}
	b.WriteString(base)
	if fieldPlan.GeneratedOptional == tagparser.OptionalPlacementBeforeOperations {
		b.WriteString(".Optional()")
	}
	for _, operation := range fieldPlan.Operations {
		code, err := renderValidatorChain(operation, field.Type)
		if err != nil {
			return "", fmt.Errorf("render %s for field %s: %w", operation.Name, field.Name, err)
		}
		b.WriteString(code)
	}
	if fieldPlan.GeneratedOptional == tagparser.OptionalPlacementAfterOperations {
		b.WriteString(".Optional()")
	}
	return b.String(), nil
}

func compiledOperand(plan tagparser.FieldPlan, op tagparser.RuleOp) (any, bool) {
	for _, operation := range plan.Operations {
		if operation.Op == op {
			return operation.Operand, true
		}
	}
	return nil, false
}

func stringFormatConstructor(field *tagparser.FieldInfo, fieldPlan tagparser.FieldPlan) (string, string) {
	if field == nil || !isStringFieldType(field.Type) {
		return "", ""
	}
	for _, operation := range fieldPlan.Operations {
		if constructor := stringFormatConstructorName(operation.Op); constructor != "" {
			return fmt.Sprintf("gozod.%s()", constructor), operation.Name
		}
	}
	return "", ""
}

func isStringFieldType(fieldType reflect.Type) bool {
	if fieldType == nil {
		return false
	}
	if fieldType.Kind() == reflect.Pointer {
		fieldType = fieldType.Elem()
	}
	return fieldType.Kind() == reflect.String
}

// renderValidatorChain returns the validator method chain for a rule.
func renderValidatorChain(plan tagparser.RulePlan, fieldType reflect.Type) (string, error) {
	switch plan.Op {
	case tagparser.RuleDefault, tagparser.RulePrefault:
		method := "Default"
		if plan.Op == tagparser.RulePrefault {
			method = "Prefault"
		}
		return generateCompiledValue(method, plan.Operand, fieldType), nil
	case tagparser.RuleRegex:
		if pattern, ok := plan.Operand.(*regexp.Regexp); ok {
			return fmt.Sprintf(".Regex(regexp.MustCompile(%q))", pattern.String()), nil
		}
		return "", fmt.Errorf("regex operand has type %T", plan.Operand)
	case tagparser.RuleMin, tagparser.RuleMax, tagparser.RuleLength,
		tagparser.RuleGT, tagparser.RuleGTE, tagparser.RuleLT, tagparser.RuleLTE,
		tagparser.RuleMultipleOf:
		if !isNumericOperand(plan.Operand) {
			return "", fmt.Errorf("%s operand has type %T", plan.Name, plan.Operand)
		}
		return fmt.Sprintf(".%s(%s)", methodForOperation(plan.Op), formatCompiledOperand(plan.Operand)), nil
	case tagparser.RuleIncludes, tagparser.RuleStartsWith, tagparser.RuleEndsWith:
		value, ok := plan.Operand.(string)
		if !ok {
			return "", fmt.Errorf("%s operand has type %T", plan.Name, plan.Operand)
		}
		return fmt.Sprintf(".%s(%q)", methodForOperation(plan.Op), value), nil
	case tagparser.RuleNilable, tagparser.RuleTrim, tagparser.RuleLowercase,
		tagparser.RuleUppercase, tagparser.RulePositive, tagparser.RuleNegative,
		tagparser.RuleFinite, tagparser.RuleNonEmpty:
		return fmt.Sprintf(".%s()", methodForOperation(plan.Op)), nil
	case tagparser.RuleEmail, tagparser.RuleURL, tagparser.RuleUUID,
		tagparser.RuleIPv4, tagparser.RuleIPv6, tagparser.RuleCIDRv4,
		tagparser.RuleCIDRv6, tagparser.RuleCUID, tagparser.RuleCUID2,
		tagparser.RuleJWT, tagparser.RuleISODateTime, tagparser.RuleISODate,
		tagparser.RuleISOTime, tagparser.RuleISODuration:
		return fmt.Sprintf(".%s()", stringFormatConstructorName(plan.Op)), nil
	case tagparser.RuleRequired, tagparser.RuleOptional, tagparser.RuleCoerce, tagparser.RuleTime,
		tagparser.RuleEnum, tagparser.RuleLiteral:
		return "", nil
	default:
		return "", fmt.Errorf("unsupported operation %q", plan.Op)
	}
}

func isNumericOperand(value any) bool {
	switch value.(type) {
	case int, int64, uint64, float64:
		return true
	default:
		return false
	}
}

func formatCompiledOperand(value any) string {
	if value, ok := value.(float64); ok {
		literal := strconv.FormatFloat(value, 'f', -1, 64)
		if !strings.ContainsAny(literal, ".eE") {
			literal += ".0"
		}
		return literal
	}
	return fmt.Sprint(value)
}

func methodForOperation(op tagparser.RuleOp) string {
	return map[tagparser.RuleOp]string{
		tagparser.RuleNilable: "Nilable", tagparser.RuleMin: "Min", tagparser.RuleMax: "Max",
		tagparser.RuleLength: "Length", tagparser.RuleGT: "Gt", tagparser.RuleGTE: "Gte",
		tagparser.RuleLT: "Lt", tagparser.RuleLTE: "Lte", tagparser.RuleIncludes: "Includes",
		tagparser.RuleStartsWith: "StartsWith", tagparser.RuleEndsWith: "EndsWith",
		tagparser.RuleMultipleOf: "MultipleOf", tagparser.RuleTrim: "Trim",
		tagparser.RuleLowercase: "ToLowerCase", tagparser.RuleUppercase: "ToUpperCase",
		tagparser.RulePositive: "Positive", tagparser.RuleNegative: "Negative",
		tagparser.RuleFinite: "Finite", tagparser.RuleNonEmpty: "NonEmpty",
	}[op]
}

func stringFormatConstructorName(op tagparser.RuleOp) string {
	return map[tagparser.RuleOp]string{
		tagparser.RuleEmail: "Email", tagparser.RuleURL: "URL", tagparser.RuleUUID: "UUID",
		tagparser.RuleIPv4: "IPv4", tagparser.RuleIPv6: "IPv6", tagparser.RuleCIDRv4: "CIDRv4",
		tagparser.RuleCIDRv6: "CIDRv6", tagparser.RuleCUID: "CUID", tagparser.RuleCUID2: "CUID2",
		tagparser.RuleJWT: "JWT", tagparser.RuleISODateTime: "IsoDateTime",
		tagparser.RuleISODate: "IsoDate", tagparser.RuleISOTime: "IsoTime",
		tagparser.RuleISODuration: "IsoDuration",
	}[op]
}

func generateCompiledValue(method string, value any, _ reflect.Type) string {
	if value == nil {
		return fmt.Sprintf(".%s(nil)", method)
	}
	return fmt.Sprintf(".%s(%#v)", method, value)
}

// baseConstructor returns the GoZod constructor for a type name with circular reference detection.
func baseConstructor(typeName, structName, fieldNameTag string) (string, error) {
	return baseConstructorWithProviders(typeName, structName, fieldNameTag, defaultMethodName, nil)
}

func (w *FileWriter) baseConstructor(typeName, structName string) (string, error) {
	return baseConstructorWithProviders(typeName, structName, w.fieldNameTag, w.methodName, w.providers)
}

func baseConstructorWithProviders(
	typeName string,
	structName string,
	fieldNameTag string,
	methodName string,
	providers *generatedProviderPlan,
) (string, error) {
	if base, ok := strings.CutPrefix(typeName, "*"); ok {
		if basicTypes[base] {
			return basicTypeConstructor(base), nil
		}
		return namedTypeConstructor(base, structName, fieldNameTag, methodName, providers), nil
	}

	if elem, ok := strings.CutPrefix(typeName, "[]"); ok {
		inner, err := baseConstructorWithProviders(elem, structName, fieldNameTag, methodName, providers)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("gozod.Slice[%s](%s)", elem, inner), nil
	}

	if strings.HasPrefix(typeName, "map[") {
		if idx := strings.LastIndex(typeName, "]"); idx != -1 && idx < len(typeName)-1 {
			keyType := typeName[len("map["):idx]
			valType := typeName[idx+1:]
			inner, err := baseConstructorWithProviders(valType, structName, fieldNameTag, methodName, providers)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("gozod.Record[%s, %s](gozod.String(), %s)", keyType, valType, inner), nil
		}
		return "", fmt.Errorf("malformed map type %q", typeName)
	}

	if basicTypes[typeName] {
		return basicTypeConstructor(typeName), nil
	}
	if typeName == "time.Time" {
		return "gozod.Time()", nil
	}
	if typeName == "unknown" || typeName == "any" || typeName == "interface{}" || typeName == "interface {}" {
		return "gozod.Any()", nil
	}
	if typeName == "" {
		return "", fmt.Errorf("unknown type %q", typeName)
	}
	return namedTypeConstructor(typeName, structName, fieldNameTag, methodName, providers), nil
}

func namedTypeConstructor(typeName, structName, fieldNameTag, methodName string, providers *generatedProviderPlan) string {
	if providers.has(typeName) {
		provider := fmt.Sprintf("%s{}.%s()", typeName, methodName)
		if providers.isCyclic(structName, typeName) {
			return lazyStructConstructor(typeName, provider)
		}
		return provider
	}
	fallback := fromStructConstructor(typeName, fieldNameTag)
	if structName != "" && typeName == structName {
		return lazyStructConstructor(typeName, fallback)
	}
	return fallback
}

func lazyStructConstructor(typeName, provider string) string {
	return fmt.Sprintf(
		"gozod.LazyTyped[*%[1]s](func() any { return %[2]s })",
		typeName,
		provider,
	)
}

func fromStructConstructor(typeName, fieldNameTag string) string {
	if fieldNameTag == "" || fieldNameTag == defaultFieldNameTag {
		return fmt.Sprintf("gozod.MustFromStruct[%s]()", typeName)
	}
	return fmt.Sprintf("gozod.MustFromStruct[%s](gozod.WithFieldNameTag(%q))", typeName, fieldNameTag)
}

func coercionConstructor(typeName string) (string, error) {
	pointer := strings.HasPrefix(typeName, "*")
	base := strings.TrimPrefix(typeName, "*")
	constructor, ok := map[string]string{
		"string": "String", "bool": "Bool",
		"int": "Int", "int8": "Int8", "int16": "Int16", "int32": "Int32", "int64": "Int64",
		"uint": "Uint", "uint8": "Uint8", "uint16": "Uint16", "uint32": "Uint32", "uint64": "Uint64",
		"float32": "Float32", "float64": "Float64", "time.Time": "Time",
	}[base]
	if !ok {
		return "", fmt.Errorf("coercion is not supported for %q", typeName)
	}
	if pointer {
		constructor += "Ptr"
	}
	return "coerce." + constructor + "()", nil
}

func (w *FileWriter) fieldNameTagCall() string {
	if w.fieldNameTag == "" || w.fieldNameTag == defaultFieldNameTag {
		return ""
	}
	return fmt.Sprintf(".WithFieldNameTag(%q)", w.fieldNameTag)
}

// basicTypeConstructor returns the GoZod constructor for a basic Go type.
func basicTypeConstructor(typeName string) string {
	if c, ok := basicTypeConstructors[typeName]; ok {
		return c
	}
	if typeName == "unknown" || typeName == "any" || typeName == "interface{}" || typeName == "interface {}" {
		return "gozod.Any()"
	}
	return ""
}

// outputPath generates the output file path based on source file location.
func (w *FileWriter) outputPath(sourceFilePath, structName string) string {
	dir := filepath.Dir(sourceFilePath)
	return filepath.Join(dir, toSnakeCase(structName)+w.outputSuffix)
}

// TemplateData contains data passed to code generation templates.
type TemplateData struct {
	PackageName      string
	StructName       string
	MethodName       string
	FieldNameTagCall string
	Fields           []tagparser.FieldInfo
	FieldSchemas     []FieldSchemaInfo
	Imports          []string
}

// loadTemplates loads the code generation templates.
func loadTemplates() (*template.Template, error) {
	// Define the main template for generated code
	mainTemplate := `// Code generated by gozodgen. DO NOT EDIT.

package {{.PackageName}}

import (
{{- range .Imports}}
	"{{.}}"
{{- end}}
)

// {{.MethodName}} returns a generated gozod schema for {{.StructName}}.
// Package-local generated dependencies call their generated schema methods.
func ({{.StructName | receiverName}} {{.StructName}}) {{.MethodName}}() *gozod.ZodStruct[{{.StructName}}, {{.StructName}}] {
	return gozod.Struct[{{.StructName}}](gozod.StructSchema{
{{- range .FieldSchemas}}
		"{{.FieldName}}": {{.SchemaCode}},
{{- end}}
		}){{.FieldNameTagCall}}
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

func baseTypeName(s string) string {
	name, _, _ := strings.Cut(s, "[")
	return name
}

// firstLowerCase converts the first character of a string to lowercase and handles special cases.
func firstLowerCase(s string) string {
	if s == "" {
		return s
	}

	s = baseTypeName(s)

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

	s = baseTypeName(s)

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
	result.Grow(len(s) * 2)

	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				result.WriteRune('_')
			}
			result.WriteRune(r - 'A' + 'a')
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}
