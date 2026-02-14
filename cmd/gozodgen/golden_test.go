package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestGoldenFiles tests that generated code matches expected golden files
func TestGoldenFiles(t *testing.T) {
	tests := []struct {
		name         string
		sourceFile   string
		expectedFile string
		updateGolden bool // Set to true to update golden files during development
	}{
		{
			name:         "simple struct generation",
			sourceFile:   "simple_struct.go",
			expectedFile: "expected_simple_struct_gen.go",
			updateGolden: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check if source file exists
			sourceFilePath := filepath.Join("testdata", tt.sourceFile)
			if _, err := os.Stat(sourceFilePath); os.IsNotExist(err) {
				t.Skipf("Source file %s not found", sourceFilePath)
				return
			}

			// Check if expected file exists
			expectedFilePath := filepath.Join("testdata", tt.expectedFile)
			if _, err := os.Stat(expectedFilePath); os.IsNotExist(err) && !tt.updateGolden {
				t.Skipf("Expected file %s not found (set updateGolden to true to create it)", expectedFilePath)
				return
			}

			helper := NewTestHelper(t)

			// Read and copy source file to temp directory
			sourceContent, err := os.ReadFile(sourceFilePath)
			require.NoError(t, err, "Failed to read source file")
			helper.CreateGoFile(tt.sourceFile, string(sourceContent))

			// Generate code
			config := &GeneratorConfig{
				OutputSuffix: "_gen.go",
				PackageName:  "testdata",
				Verbose:      false,
				DryRun:       false,
			}

			generator, err := NewCodeGenerator(config)
			require.NoError(t, err, "Failed to create generator")

			writer, err := NewFileWriter(helper.GetTempDir(), config.PackageName, config.OutputSuffix, config.DryRun, config.Verbose)
			require.NoError(t, err, "Failed to create writer")
			generator.writer = writer

			err = generator.ProcessPackage(helper.GetTempDir())
			require.NoError(t, err, "Failed to generate code")

			// Find generated file
			generatedFiles := helper.ListGeneratedFiles()
			var generatedFile string
			for _, file := range generatedFiles {
				if strings.HasSuffix(file, "_gen.go") {
					generatedFile = file
					break
				}
			}

			if generatedFile == "" {
				t.Fatal("No generated file found")
			}

			generatedContent := helper.ReadGeneratedFile(generatedFile)

			// Normalize generated content (remove timestamps and other variable parts)
			normalizedGenerated := normalizeGeneratedCode(generatedContent)

			if tt.updateGolden {
				// Update golden file
				err := os.WriteFile(expectedFilePath, []byte(normalizedGenerated), 0600)
				require.NoError(t, err, "Failed to update golden file")
				t.Logf("Updated golden file: %s", expectedFilePath)
				return
			}

			// Compare with expected content
			expectedContent, err := os.ReadFile(expectedFilePath)
			require.NoError(t, err, "Failed to read expected file")

			normalizedExpected := normalizeGeneratedCode(string(expectedContent))

			if normalizedGenerated != normalizedExpected {
				t.Errorf("Generated code doesn't match golden file.\n\nGenerated:\n%s\n\nExpected:\n%s\n\nDiff: run 'diff -u %s <generated_file>' to see differences",
					normalizedGenerated, normalizedExpected, expectedFilePath)

				// For debugging, write actual generated content to a temp file
				tempFile := filepath.Join(helper.GetTempDir(), "actual_"+tt.expectedFile)
				if writeErr := os.WriteFile(tempFile, []byte(normalizedGenerated), 0600); writeErr != nil {
					t.Logf("Warning: could not write debug file: %v", writeErr)
				}
				t.Logf("Actual generated content written to: %s", tempFile)
			}
		})
	}
}

// normalizeGeneratedCode removes variable parts from generated code for comparison
func normalizeGeneratedCode(content string) string {
	lines := strings.Split(content, "\n")
	normalized := make([]string, 0, len(lines))

	for _, line := range lines {
		// Skip timestamp lines
		if strings.Contains(line, "Generated at:") {
			continue
		}
		// Skip empty lines at the end
		if strings.TrimSpace(line) == "" && len(normalized) > 0 && strings.TrimSpace(normalized[len(normalized)-1]) == "" {
			continue
		}
		normalized = append(normalized, line)
	}

	// Remove trailing empty lines
	for len(normalized) > 0 && strings.TrimSpace(normalized[len(normalized)-1]) == "" {
		normalized = normalized[:len(normalized)-1]
	}

	return strings.Join(normalized, "\n")
}

// TestCreateGoldenFiles is a utility test that can be run to create golden files
// Run with: go test -run TestCreateGoldenFiles -args -create-golden
func TestCreateGoldenFiles(t *testing.T) {
	// Only run this if explicitly requested
	if !hasFlag("-create-golden") {
		t.Skip("Skipping golden file creation (use -create-golden flag to create them)")
	}

	testCases := []string{
		"simple_struct.go",
		"complex_struct.go",
	}

	for _, testCase := range testCases {
		t.Run(testCase, func(t *testing.T) {
			sourceFilePath := filepath.Join("testdata", testCase)
			if _, err := os.Stat(sourceFilePath); os.IsNotExist(err) {
				t.Skipf("Source file %s not found", sourceFilePath)
				return
			}

			helper := NewTestHelper(t)

			// Read source file
			sourceContent, err := os.ReadFile(sourceFilePath)
			require.NoError(t, err, "Failed to read source file")
			helper.CreateGoFile(testCase, string(sourceContent))

			// Generate code
			config := &GeneratorConfig{
				OutputSuffix: "_gen.go",
				PackageName:  "testdata",
				Verbose:      false,
				DryRun:       false,
			}

			generator, err := NewCodeGenerator(config)
			require.NoError(t, err, "Failed to create generator")

			writer, err := NewFileWriter(helper.GetTempDir(), config.PackageName, config.OutputSuffix, config.DryRun, config.Verbose)
			require.NoError(t, err, "Failed to create writer")
			generator.writer = writer

			err = generator.ProcessPackage(helper.GetTempDir())
			require.NoError(t, err, "Failed to generate code")

			// Find generated file
			generatedFiles := helper.ListGeneratedFiles()
			var generatedFile string
			for _, file := range generatedFiles {
				if strings.HasSuffix(file, "_gen.go") {
					generatedFile = file
					break
				}
			}

			if generatedFile == "" {
				t.Fatal("No generated file found")
			}

			generatedContent := helper.ReadGeneratedFile(generatedFile)
			normalizedContent := normalizeGeneratedCode(generatedContent)

			// Create golden file
			baseName := strings.TrimSuffix(testCase, ".go")
			goldenFile := filepath.Join("testdata", "expected_"+baseName+"_gen.go")

			err = os.WriteFile(goldenFile, []byte(normalizedContent), 0600)
			require.NoError(t, err, "Failed to create golden file")

			t.Logf("Created golden file: %s", goldenFile)
		})
	}
}

// hasFlag checks if a specific flag is present in os.Args
func hasFlag(flag string) bool {
	return slices.Contains(os.Args, flag)
}

// TestRegenerateGoldenFiles can be used to update all golden files
func TestRegenerateGoldenFiles(t *testing.T) {
	if !hasFlag("-update-golden") {
		t.Skip("Skipping golden file update (use -update-golden flag to update them)")
	}

	// Update simple struct golden file
	t.Run("update_simple_struct", func(t *testing.T) {
		test := struct {
			name         string
			sourceFile   string
			expectedFile string
			updateGolden bool
		}{
			name:         "simple struct generation",
			sourceFile:   "simple_struct.go",
			expectedFile: "expected_simple_struct_gen.go",
			updateGolden: true,
		}

		// Run the golden file test with update enabled
		sourceFilePath := filepath.Join("testdata", test.sourceFile)
		if _, err := os.Stat(sourceFilePath); os.IsNotExist(err) {
			t.Skipf("Source file %s not found", sourceFilePath)
			return
		}

		helper := NewTestHelper(t)

		sourceContent, err := os.ReadFile(sourceFilePath)
		require.NoError(t, err, "Failed to read source file")
		helper.CreateGoFile(test.sourceFile, string(sourceContent))

		config := &GeneratorConfig{
			OutputSuffix: "_gen.go",
			PackageName:  "testdata",
			Verbose:      false,
			DryRun:       false,
		}

		generator, err := NewCodeGenerator(config)
		require.NoError(t, err, "Failed to create generator")

		writer, err := NewFileWriter(helper.GetTempDir(), config.PackageName, config.OutputSuffix, config.DryRun, config.Verbose)
		require.NoError(t, err, "Failed to create writer")
		generator.writer = writer

		err = generator.ProcessPackage(helper.GetTempDir())
		require.NoError(t, err, "Failed to generate code")

		generatedFiles := helper.ListGeneratedFiles()
		var generatedFile string
		for _, file := range generatedFiles {
			if strings.HasSuffix(file, "_gen.go") {
				generatedFile = file
				break
			}
		}

		if generatedFile == "" {
			t.Fatal("No generated file found")
		}

		generatedContent := helper.ReadGeneratedFile(generatedFile)
		normalizedGenerated := normalizeGeneratedCode(generatedContent)

		expectedFilePath := filepath.Join("testdata", test.expectedFile)
		err = os.WriteFile(expectedFilePath, []byte(normalizedGenerated), 0600)
		require.NoError(t, err, "Failed to update golden file")
		t.Logf("Updated golden file: %s", expectedFilePath)
	})
}
