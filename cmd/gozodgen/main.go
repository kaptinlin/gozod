// Package main implements the gozodgen code generation tool.
// This tool generates optimized Schema methods for Go structs with gozod tags,
// enabling zero-reflection validation at runtime.
//
// Usage:
//
//	gozodgen [flags] [packages...]
//
// Flags:
//
//	-suffix string     Output file suffix (default: "_gen.go")
//	-package string    Specify package name (default: auto-detect)
//	-tags string       Build tags
//	-verbose          Verbose output
//	-dry-run          Preview generated code without writing files
//	-force            Force regeneration of all files
package main

import (
	"flag"
	"fmt"
	"log"
	"reflect"
	"strings"
)

// Static error variables to comply with err113
var (
	ErrConfigNil = fmt.Errorf("config cannot be nil")
)

// Command line flags
var (
	outputSuffix = flag.String("suffix", "_gen.go", "Output file suffix (e.g., '_schema.go', '_validators.go')")
	packageName  = flag.String("package", "", "Specify package name (default: auto-detect)")
	buildTags    = flag.String("tags", "", "Build tags")
	verbose      = flag.Bool("verbose", false, "Verbose output")
	dryRun       = flag.Bool("dry-run", false, "Preview generated code without writing files")
	force        = flag.Bool("force", false, "Force regeneration of all files")
	help         = flag.Bool("help", false, "Show help message")
)

func main() {
	flag.Parse()

	if *help {
		showHelp()
		return
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
		BuildTags:    parseBuildTags(*buildTags),
		Verbose:      *verbose,
		DryRun:       *dryRun,
		Force:        *force,
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

// showHelp displays the help message
func showHelp() {
	fmt.Println(`gozodgen - GoZod Code Generation Tool

Generates optimized Schema methods for Go structs with gozod tags,
enabling zero-reflection validation at runtime.

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

    # Force regeneration with custom suffix
    gozodgen -force -suffix="_validators.go"

    # Generate with build tags
    gozodgen -tags="integration,test"

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
    gozod tags, providing zero-reflection validation performance.`)
}

// parseBuildTags parses comma-separated build tags
func parseBuildTags(tags string) []string {
	if tags == "" {
		return nil
	}
	parts := strings.Split(strings.TrimSpace(tags), ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// GeneratorConfig holds configuration for the code generator
type GeneratorConfig struct {
	OutputSuffix string   // File suffix for generated files
	PackageName  string   // Override package name
	BuildTags    []string // Build tags to include
	Verbose      bool     // Enable verbose logging
	DryRun       bool     // Preview mode without writing files
	Force        bool     // Force regeneration
}

// ValidatorGenerator generates validator code for fields
type ValidatorGenerator func(fieldType string, params []string) string

// FieldInfo represents information about a struct field for code generation
type FieldInfo struct {
	Name     string       // Field name
	Type     reflect.Type // Field type
	Rules    []string     // Parsed validation rules
	JsonName string       // JSON field name (from json tag)
	GozodTag string       // Raw gozod tag value
}

// ValidatorInfo contains validator generation information
type ValidatorInfo struct {
	FieldName     string // Field name in schema
	ValidatorName string // Validator function name
	Value         string // Generated validator code
	ChainCall     string // Method chain call
}

// FieldSchemaInfo contains field schema generation information
type FieldSchemaInfo struct {
	FieldName  string // Field name
	SchemaCode string // Generated schema code
}

// CodeGenerator represents the main code generation engine
type CodeGenerator struct {
	config       *GeneratorConfig
	analyzer     *StructAnalyzer
	writer       *FileWriter
	typeMap      map[string]string
	validatorMap map[string]ValidatorGenerator
}

// NewCodeGenerator creates a new code generator instance
func NewCodeGenerator(config *GeneratorConfig) (*CodeGenerator, error) {
	if config == nil {
		return nil, ErrConfigNil
	}

	// Create analyzer for parsing Go source files
	analyzer, err := NewStructAnalyzer()
	if err != nil {
		return nil, fmt.Errorf("failed to create analyzer: %w", err)
	}

	// Create file writer for output
	writer, err := NewFileWriter("", config.PackageName, config.OutputSuffix, config.DryRun, config.Verbose)
	if err != nil {
		return nil, fmt.Errorf("failed to create writer: %w", err)
	}

	return &CodeGenerator{
		config:       config,
		analyzer:     analyzer,
		writer:       writer,
		typeMap:      createTypeMapping(),
		validatorMap: createValidatorMapping(),
	}, nil
}
