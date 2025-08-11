// Package main implements the gozodgen code generation tool.
// This file contains the core code generation logic that integrates with
// the existing pkg/tags infrastructure to generate optimized Schema methods.
package main

import (
	"fmt"
	"strings"
)

// Types are now defined in adapter.go

// NewCodeGeneratorFromConfig creates a new CodeGenerator instance with the specified configuration
func NewCodeGeneratorFromConfig(config *GeneratorConfig) *CodeGenerator {
	analyzer, _ := NewStructAnalyzer()
	writer, _ := NewFileWriter("", "", config.OutputSuffix, config.DryRun, config.Verbose)

	generator := &CodeGenerator{
		analyzer:     analyzer,
		writer:       writer,
		config:       config,
		typeMap:      createTypeMapping(),
		validatorMap: createValidatorMapping(),
	}

	return generator
}

// ProcessPackage analyzes and generates code for all structs in the specified package
func (g *CodeGenerator) ProcessPackage(packagePath string) error {
	if g.config.Verbose {
		fmt.Printf("Processing package: %s\n", packagePath)
	}

	// Analyze the package to find structs that need code generation
	structInfos, err := g.analyzer.AnalyzePackage(packagePath)
	if err != nil {
		return fmt.Errorf("failed to analyze package %s: %w", packagePath, err)
	}

	if len(structInfos) == 0 {
		if g.config.Verbose {
			fmt.Printf("No structs found requiring code generation in package: %s\n", packagePath)
		}
		return nil
	}

	// Generate code for each struct individually
	for _, structInfo := range structInfos {
		if err := g.generateStructFile(structInfo); err != nil {
			return fmt.Errorf("failed to generate code for struct %s: %w", structInfo.Name, err)
		}

		if g.config.Verbose {
			fmt.Printf("Generated code for struct: %s\n", structInfo.Name)
		}
	}

	return nil
}

// generateStructFile generates code for a single struct
func (g *CodeGenerator) generateStructFile(structInfo *GenerationInfo) error {
	// Write the generated code to file
	return g.writer.WriteGeneratedCode(structInfo)
}

// createTypeMapping creates the mapping from Go types to GoZod constructors
func createTypeMapping() map[string]string {
	return map[string]string{
		"string":    "gozod.String()",
		"int":       "gozod.Int()",
		"int8":      "gozod.Int8()",
		"int16":     "gozod.Int16()",
		"int32":     "gozod.Int32()",
		"int64":     "gozod.Int64()",
		"uint":      "gozod.Uint()",
		"uint8":     "gozod.Uint8()",
		"uint16":    "gozod.Uint16()",
		"uint32":    "gozod.Uint32()",
		"uint64":    "gozod.Uint64()",
		"float32":   "gozod.Float32()",
		"float64":   "gozod.Float64()",
		"bool":      "gozod.Bool()",
		"time.Time": "gozod.Time()",
	}
}

// createValidatorMapping creates the mapping from validator names to code generators
func createValidatorMapping() map[string]ValidatorGenerator {
	return map[string]ValidatorGenerator{
		"required": func(fieldType string, params []string) string {
			// Required is handled by the type system and optional logic
			return ""
		},
		"min": func(fieldType string, params []string) string {
			if len(params) == 0 {
				return ""
			}
			return fmt.Sprintf(".Min(%s)", params[0])
		},
		"max": func(fieldType string, params []string) string {
			if len(params) == 0 {
				return ""
			}
			return fmt.Sprintf(".Max(%s)", params[0])
		},
		"email": func(fieldType string, params []string) string {
			return ".Email()"
		},
		"url": func(fieldType string, params []string) string {
			return ".URL()"
		},
		"uuid": func(fieldType string, params []string) string {
			return ".UUID()"
		},
		"ipv4": func(fieldType string, params []string) string {
			return ".IPv4()"
		},
		"gt": func(fieldType string, params []string) string {
			if len(params) == 0 {
				return ""
			}
			return fmt.Sprintf(".Gt(%s)", params[0])
		},
		"gte": func(fieldType string, params []string) string {
			if len(params) == 0 {
				return ""
			}
			return fmt.Sprintf(".Gte(%s)", params[0])
		},
		"lt": func(fieldType string, params []string) string {
			if len(params) == 0 {
				return ""
			}
			return fmt.Sprintf(".Lt(%s)", params[0])
		},
		"lte": func(fieldType string, params []string) string {
			if len(params) == 0 {
				return ""
			}
			return fmt.Sprintf(".Lte(%s)", params[0])
		},
		"enum": func(fieldType string, params []string) string {
			if len(params) == 0 {
				return ""
			}
			var values []string
			for _, param := range params {
				values = append(values, fmt.Sprintf(`"%s"`, param))
			}
			return fmt.Sprintf(".Enum(%s)", strings.Join(values, ", "))
		},
		"regex": func(fieldType string, params []string) string {
			if len(params) == 0 {
				return ""
			}
			return fmt.Sprintf(".Regex(`%s`)", params[0])
		},
		"time": func(fieldType string, params []string) string {
			// Time validation is handled by the Time() constructor
			return ""
		},
		"default": func(fieldType string, params []string) string {
			if len(params) == 0 {
				return ""
			}
			// Handle different types of default values
			value := params[0]
			if strings.Contains(fieldType, "string") {
				return fmt.Sprintf(`.Default("%s")`, value)
			}
			return fmt.Sprintf(".Default(%s)", value)
		},
		"prefault": func(fieldType string, params []string) string {
			if len(params) == 0 {
				return ""
			}
			// Handle different types of prefault values
			value := params[0]
			if strings.Contains(fieldType, "string") {
				return fmt.Sprintf(`.Prefault("%s")`, value)
			}
			return fmt.Sprintf(".Prefault(%s)", value)
		},
		"trim": func(fieldType string, params []string) string {
			return ".Transform(gozod.Trim())"
		},
		"lowercase": func(fieldType string, params []string) string {
			return ".Transform(gozod.Lowercase())"
		},
		"uppercase": func(fieldType string, params []string) string {
			return ".Transform(gozod.Uppercase())"
		},
		"nilable": func(fieldType string, params []string) string {
			// Nilable is handled by the optional logic
			return ""
		},
		"refine": func(fieldType string, params []string) string {
			if len(params) == 0 {
				return ""
			}
			// Generate refine method call with function parameter
			return fmt.Sprintf(".Refine(%s)", params[0])
		},
		"check": func(fieldType string, params []string) string {
			if len(params) == 0 {
				return ""
			}
			// Generate custom check method call
			return fmt.Sprintf(".Check(%s)", params[0])
		},
		"lazy": func(fieldType string, params []string) string {
			// Lazy evaluation is handled at the constructor level
			return ""
		},
		"custom": func(fieldType string, params []string) string {
			if len(params) == 0 {
				return ""
			}
			return fmt.Sprintf(".Custom(%s)", params[0])
		},
	}
}
