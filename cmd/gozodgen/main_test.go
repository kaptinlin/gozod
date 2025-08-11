package main

import (
	"flag"
	"os"
	"testing"
)

func TestParseBuildTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "single tag",
			input:    "integration",
			expected: []string{"integration"},
		},
		{
			name:     "multiple tags",
			input:    "integration,test,dev",
			expected: []string{"integration", "test", "dev"},
		},
		{
			name:     "tags with spaces",
			input:    " integration , test , dev ",
			expected: []string{"integration", "test", "dev"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseBuildTags(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d tags, got %d", len(tt.expected), len(result))
				return
			}

			for i, tag := range result {
				if tag != tt.expected[i] {
					t.Errorf("Expected tag %s, got %s", tt.expected[i], tag)
				}
			}
		})
	}
}

func TestGeneratorConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *GeneratorConfig
		expectValid bool
	}{
		{
			name: "valid config",
			config: &GeneratorConfig{
				OutputSuffix: "_gen.go",
				PackageName:  "main",
				BuildTags:    []string{"integration"},
				Verbose:      true,
				DryRun:       false,
				Force:        true,
			},
			expectValid: true,
		},
		{
			name: "minimal config",
			config: &GeneratorConfig{
				OutputSuffix: "_schema.go",
			},
			expectValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that we can create a generator with the config
			generator, err := NewCodeGenerator(tt.config)
			if tt.expectValid && err != nil {
				t.Errorf("Expected valid config, but got error: %v", err)
			}
			if tt.expectValid && generator == nil {
				t.Error("Expected non-nil generator")
			}
		})
	}
}

// Mock main function behavior for testing
func runMainWithArgs(args []string) (outputSuffix, packageName string, verbose, dryRun bool, packages []string) {
	// Reset flags
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Re-define flags for testing
	outputSuffixFlag := flag.String("suffix", "_gen.go", "Output file suffix")
	packageNameFlag := flag.String("package", "", "Specify package name")
	verboseFlag := flag.Bool("verbose", false, "Verbose output")
	dryRunFlag := flag.Bool("dry-run", false, "Preview mode")

	// Set test args and parse
	os.Args = append([]string{"gozodgen"}, args...)
	flag.Parse()

	// Get results and return them directly
	return *outputSuffixFlag, *packageNameFlag, *verboseFlag, *dryRunFlag, flag.Args()
}

func TestMainCommandLineParsing(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		expectedSuffix   string
		expectedPackage  string
		expectedVerbose  bool
		expectedDryRun   bool
		expectedPackages []string
	}{
		{
			name:             "default arguments",
			args:             []string{},
			expectedSuffix:   "_gen.go",
			expectedPackage:  "",
			expectedVerbose:  false,
			expectedDryRun:   false,
			expectedPackages: []string{},
		},
		{
			name:             "custom suffix",
			args:             []string{"-suffix", "_schema.go"},
			expectedSuffix:   "_schema.go",
			expectedPackage:  "",
			expectedVerbose:  false,
			expectedDryRun:   false,
			expectedPackages: []string{},
		},
		{
			name:             "verbose and dry run",
			args:             []string{"-verbose", "-dry-run"},
			expectedSuffix:   "_gen.go",
			expectedPackage:  "",
			expectedVerbose:  true,
			expectedDryRun:   true,
			expectedPackages: []string{},
		},
		{
			name:             "package name and target packages",
			args:             []string{"-package", "models", "./models", "./api"},
			expectedSuffix:   "_gen.go",
			expectedPackage:  "models",
			expectedVerbose:  false,
			expectedDryRun:   false,
			expectedPackages: []string{"./models", "./api"},
		},
		{
			name:             "all options",
			args:             []string{"-suffix", "_validators.go", "-package", "api", "-verbose", "-dry-run", "./src"},
			expectedSuffix:   "_validators.go",
			expectedPackage:  "api",
			expectedVerbose:  true,
			expectedDryRun:   true,
			expectedPackages: []string{"./src"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suffix, pkg, verbose, dryRun, packages := runMainWithArgs(tt.args)

			if suffix != tt.expectedSuffix {
				t.Errorf("Expected suffix %s, got %s", tt.expectedSuffix, suffix)
			}
			if pkg != tt.expectedPackage {
				t.Errorf("Expected package %s, got %s", tt.expectedPackage, pkg)
			}
			if verbose != tt.expectedVerbose {
				t.Errorf("Expected verbose %t, got %t", tt.expectedVerbose, verbose)
			}
			if dryRun != tt.expectedDryRun {
				t.Errorf("Expected dryRun %t, got %t", tt.expectedDryRun, dryRun)
			}

			if len(packages) != len(tt.expectedPackages) {
				t.Errorf("Expected %d packages, got %d", len(tt.expectedPackages), len(packages))
				return
			}

			for i, pkg := range packages {
				if pkg != tt.expectedPackages[i] {
					t.Errorf("Package %d: expected %s, got %s", i, tt.expectedPackages[i], pkg)
				}
			}
		})
	}
}

// Integration test that runs the full pipeline
func TestMainIntegration(t *testing.T) {
	helper := NewTestHelper(t)

	// Create a test Go file
	content := `package main

type TestUser struct {
	ID    string ` + "`json:\"id\" gozod:\"required,uuid\"`" + `
	Name  string ` + "`json:\"name\" gozod:\"required,min=2,max=50\"`" + `
	Email string ` + "`json:\"email\" gozod:\"required,email\"`" + `
}

type TestProduct struct {
	ID    string  ` + "`json:\"id\" gozod:\"required,uuid\"`" + `
	Name  string  ` + "`json:\"name\" gozod:\"required,min=1\"`" + `
	Price float64 ` + "`json:\"price\" gozod:\"required,gt=0\"`" + `
}`

	helper.CreateGoFile("models.go", content)

	// Test configuration
	config := &GeneratorConfig{
		OutputSuffix: "_gen.go",
		PackageName:  "main",
		Verbose:      false,
		DryRun:       false,
		Force:        false,
	}

	// Create and run generator
	generator, err := NewCodeGenerator(config)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	// Update writer to use temp directory
	writer, err := NewFileWriter(helper.GetTempDir(), config.PackageName, config.OutputSuffix, config.DryRun, config.Verbose)
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}
	generator.writer = writer

	// Process package
	err = generator.ProcessPackage(helper.GetTempDir())
	if err != nil {
		t.Fatalf("Failed to process package: %v", err)
	}

	// Verify results - check both generated files
	helper.AssertFileExists("test_user_gen.go")
	helper.AssertFileExists("test_product_gen.go")

	userContent := helper.ReadGeneratedFile("test_user_gen.go")
	helper.AssertValidGoCode(userContent)

	productContent := helper.ReadGeneratedFile("test_product_gen.go")
	helper.AssertValidGoCode(productContent)

	// Check that both structs have Schema methods
	helper.AssertCodeContains(userContent,
		"func (tu TestUser) Schema() *gozod.ZodStruct[TestUser, TestUser]",
		`"id": gozod.Uuid()`,
		`"email": gozod.String().Email()`,
	)

	helper.AssertCodeContains(productContent,
		"func (tp TestProduct) Schema() *gozod.ZodStruct[TestProduct, TestProduct]",
		`"price": gozod.Float64().Gt(0)`,
	)

	// Check imports
	helper.AssertImportsCorrect(userContent,
		[]string{"github.com/kaptinlin/gozod"},
		[]string{"github.com/kaptinlin/gozod/core"},
	)

	helper.AssertImportsCorrect(productContent,
		[]string{"github.com/kaptinlin/gozod"},
		[]string{"github.com/kaptinlin/gozod/core"},
	)

	t.Logf("Integration test successful - generated valid code")
}

func TestFieldInfoAndValidatorInfo(t *testing.T) {
	// Test data structures used in main
	fieldInfo := FieldInfo{
		Name:     "TestField",
		Type:     nil, // Would be set to actual type in real usage
		Rules:    []string{"required", "min=2"},
		JsonName: "test_field",
		GozodTag: "required,min=2",
	}

	if fieldInfo.Name != "TestField" {
		t.Errorf("Expected Name to be TestField, got %s", fieldInfo.Name)
	}

	validatorInfo := ValidatorInfo{
		FieldName:     "TestField",
		ValidatorName: "min",
		Value:         "2",
		ChainCall:     ".Min(2)",
	}

	if validatorInfo.ChainCall != ".Min(2)" {
		t.Errorf("Expected ChainCall to be .Min(2), got %s", validatorInfo.ChainCall)
	}

	fieldSchemaInfo := FieldSchemaInfo{
		FieldName:  "TestField",
		SchemaCode: "gozod.String().Min(2)",
	}

	if fieldSchemaInfo.SchemaCode != "gozod.String().Min(2)" {
		t.Errorf("Expected SchemaCode to be gozod.String().Min(2), got %s", fieldSchemaInfo.SchemaCode)
	}
}
