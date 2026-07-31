// Package main implements the gozodgen code generation tool.
// This file contains the core code generation logic that integrates with
// the existing pkg/tags infrastructure to generate optimized Schema methods.
package main

import (
	"fmt"
)

// ProcessPackage analyzes and generates code for all structs in the specified package.
func (g *CodeGenerator) ProcessPackage(packagePath string) error {
	if g.config.Verbose {
		fmt.Printf("Processing package: %s\n", packagePath)
	}

	structInfos, err := g.analyzer.AnalyzePackage(packagePath)
	if err != nil {
		return fmt.Errorf("analyze package %s: %w", packagePath, err)
	}

	if len(structInfos) == 0 {
		if g.config.Verbose {
			fmt.Printf("No structs found requiring code generation in package: %s\n", packagePath)
		}
		return nil
	}
	g.writer.providers = newGeneratedProviderPlan(structInfos)

	rendered := make([]renderedFile, 0, len(structInfos))
	for _, structInfo := range structInfos {
		result, err := g.writer.renderGeneratedCode(structInfo)
		if err != nil {
			return fmt.Errorf("generate code for struct %s: %w", structInfo.Name, err)
		}
		rendered = append(rendered, result)
	}

	for i, result := range rendered {
		if err := g.writer.publishGeneratedCode(result); err != nil {
			return fmt.Errorf("generate code for struct %s: %w", structInfos[i].Name, err)
		}
		if g.config.Verbose {
			fmt.Printf("Generated code for struct: %s\n", structInfos[i].Name)
		}
	}

	return nil
}
