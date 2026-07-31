// Package main implements the gozodgen code generation tool.
// This tool generates explicit Schema methods for Go structs with gozod tags.
// Package-local generated dependencies call their generated schema methods.
//
// Usage:
//
//	gozodgen [flags] [packages...]
//
// Flags:
//
//	-suffix string     Output file suffix (default: "_gen.go")
//	-package string    Specify package name (default: auto-detect)
//	-tag-name string   Struct tag used for validation rules (default: "gozod")
//	-field-name-tag string
//	                   Struct tag used for field names (default: "json")
//	-method string     Name of the generated method (default: "Schema")
//	-verbose          Verbose output
//	-dry-run          Preview generated code without writing files
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"unicode"
)

var errConfigNil = errors.New("config cannot be nil")

// Defaults for the rule tag, field-name tag, and generated method name.
const (
	defaultRuleTag      = "gozod"
	defaultFieldNameTag = "json"
	defaultMethodName   = "Schema"
)

// Command line flags
var (
	outputSuffix     = flag.String("suffix", "_gen.go", "Output file suffix (e.g., '_schema.go', '_validators.go')")
	packageName      = flag.String("package", "", "Specify package name (default: auto-detect)")
	ruleTagFlag      = flag.String("tag-name", defaultRuleTag, "Struct tag used for validation rules (e.g. gozod, validate)")
	fieldNameTagFlag = flag.String("field-name-tag", defaultFieldNameTag, "Struct tag used for field names (e.g. json, yaml, toml)")
	method           = flag.String("method", defaultMethodName, "Name of the generated method")
	verbose          = flag.Bool("verbose", false, "Verbose output")
	dryRun           = flag.Bool("dry-run", false, "Preview generated code without writing files")
	help             = flag.Bool("help", false, "Show help message")
)

func main() {
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	if !isExportedIdent(*method) {
		log.Fatalf("[ERROR] invalid -method %q: must be a valid exported Go identifier", *method)
	}

	// Get target packages from command line arguments
	packages := flag.Args()
	if len(packages) == 0 {
		// Default to current directory
		packages = []string{"."}
	}

	if *verbose {
		log.Printf("[INFO] Starting gozodgen code generation")
		log.Printf("[INFO] Target packages: %v", packages)
		log.Printf("[INFO] Output suffix: %s", *outputSuffix)
		if *dryRun {
			log.Printf("[INFO] Dry run mode enabled")
		}
	}

	// Create generator with configuration
	config := &GeneratorConfig{
		OutputSuffix: *outputSuffix,
		PackageName:  *packageName,
		RuleTagName:  *ruleTagFlag,
		FieldNameTag: *fieldNameTagFlag,
		MethodName:   *method,
		Verbose:      *verbose,
		DryRun:       *dryRun,
	}

	generator, err := NewCodeGenerator(config)
	if err != nil {
		log.Fatalf("[ERROR] Failed to create generator: %v", err)
	}

	// Process each package
	var totalProcessedPackages int
	for _, pkg := range packages {
		if *verbose {
			log.Printf("[INFO] Processing package: %s", pkg)
		}

		err := generator.ProcessPackage(pkg)
		if err != nil {
			log.Fatalf("[ERROR] Failed to process package %s: %v", pkg, err)
		}

		totalProcessedPackages++
		if *verbose {
			log.Printf("[INFO] Processed package %s", pkg)
		}
	}

	if *verbose {
		if totalProcessedPackages > 0 {
			log.Printf("[INFO] Code generation completed! Processed %d packages total", totalProcessedPackages)
		} else {
			log.Printf("[INFO] No structs found requiring code generation")
		}
	}
}

// showHelp displays the help message.
func showHelp() {
	fmt.Println(`gozodgen - GoZod Code Generation Tool

Generates explicit Schema methods for Go structs with gozod tags.
Package-local generated dependencies call their generated schema methods.

USAGE:
    gozodgen [flags] [packages...]

FLAGS:`)
	flag.PrintDefaults()
	fmt.Println(`
EXAMPLES:
    # Generate for current package
    gozodgen

    # Generate for specific packages
    gozodgen ./models ./api

    # Dry run to preview generated code
    gozodgen -dry-run -verbose

    # Use custom output suffix
    gozodgen -suffix="_schema.go"

	    # Resolve field names from yaml tags
	    gozodgen -field-name-tag=yaml

	    # Read validation rules from validate tags
	    gozodgen -tag-name=validate

	    # Generate a method named Validate instead of Schema
	    gozodgen -method=Validate

DIRECTIVES:
    Add //go:generate gozodgen to your Go files to enable automatic
    code generation when running 'go generate'.

    Example:
        //go:generate gozodgen
        type User struct {
            Name string ` + "`gozod:\"required,min=2\"`" + `
        }

OUTPUT:
    Generated files follow the pattern: <original>_gen.go
    Each generated file contains Schema() methods for structs with
    gozod tags. Imported or otherwise opaque field types may retain an
    explicit runtime-reflection fallback.`)
}

// GeneratorConfig holds configuration for the code generator.
type GeneratorConfig struct {
	OutputSuffix string // File suffix for generated files
	PackageName  string // Override package name
	RuleTagName  string // Struct tag used for validation rules (default "gozod")
	FieldNameTag string // Struct tag used for field names (default "json")
	MethodName   string // Generated method name (default "Schema")
	Verbose      bool   // Enable verbose logging
	DryRun       bool   // Preview mode without writing files
}

// isExportedIdent reports whether s is a valid exported Go identifier
// (begins with an uppercase letter, followed by letters, digits, or underscores).
func isExportedIdent(s string) bool {
	for i, r := range s {
		switch {
		case i == 0 && !unicode.IsUpper(r):
			return false
		case i > 0 && !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_':
			return false
		}
	}
	return s != ""
}

// FieldSchemaInfo contains field schema generation information.
type FieldSchemaInfo struct {
	FieldName  string // Field name
	SchemaCode string // Generated schema code
}

// CodeGenerator represents the main code generation engine.
type CodeGenerator struct {
	config   *GeneratorConfig
	analyzer *StructAnalyzer
	writer   *FileWriter
}

// NewCodeGenerator creates a new code generator instance.
func NewCodeGenerator(config *GeneratorConfig) (*CodeGenerator, error) {
	if config == nil {
		return nil, errConfigNil
	}

	analyzer, err := NewStructAnalyzer()
	if err != nil {
		return nil, fmt.Errorf("create analyzer: %w", err)
	}
	if config.FieldNameTag != "" {
		analyzer.fieldNameTag = config.FieldNameTag
	}
	if config.RuleTagName != "" {
		analyzer.ruleTagName = config.RuleTagName
	}

	writer, err := NewFileWriter("", config.PackageName, config.OutputSuffix, config.DryRun, config.Verbose)
	if err != nil {
		return nil, fmt.Errorf("create writer: %w", err)
	}
	if config.MethodName != "" {
		writer.methodName = config.MethodName
	}
	if config.FieldNameTag != "" {
		writer.fieldNameTag = config.FieldNameTag
	}

	return &CodeGenerator{
		config:   config,
		analyzer: analyzer,
		writer:   writer,
	}, nil
}
