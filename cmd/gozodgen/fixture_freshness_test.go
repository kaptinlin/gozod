package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeneratedFixturesAreFresh(t *testing.T) {
	fixtures := []struct {
		name        string
		directory   string
		packageName string
		sources     []string
		outputs     []string
	}{
		{
			name:        "simple testdata",
			directory:   "testdata",
			packageName: "testdata",
			sources:     []string{"simple_struct.go"},
			outputs:     []string{"user_gen.go"},
		},
		{
			name:        "complex testdata",
			directory:   "testdata",
			packageName: "testdata",
			sources:     []string{"complex_struct.go"},
			outputs:     []string{"category_gen.go", "product_gen.go"},
		},
		{
			name:        "recursive testdata",
			directory:   "testdata",
			packageName: "testdata",
			sources:     []string{"circular_struct.go"},
			outputs:     []string{"company_gen.go", "department_gen.go", "employee_gen.go", "node_gen.go"},
		},
		{
			name:        "validator testdata",
			directory:   "testdata",
			packageName: "testdata",
			sources:     []string{"validators.go"},
			outputs: []string{
				"all_types_struct_gen.go",
				"default_struct_gen.go",
				"float_default_struct_gen.go",
				"float_prefault_struct_gen.go",
				"interface_map_prefault_struct_gen.go",
				"interface_map_struct_gen.go",
				"prefault_struct_gen.go",
				"validator_struct_gen.go",
			},
		},
		{
			name:        "coercion fixture",
			directory:   "coercionfixture",
			packageName: "coercionfixture",
			sources:     []string{"coerce_struct.go"},
			outputs:     []string{"coerce_struct_gen.go"},
		},
		{
			name:        "provider fixture",
			directory:   "providerfixture",
			packageName: "providerfixture",
			sources:     []string{"provider_struct.go"},
			outputs:     []string{"company_gen.go", "department_gen.go", "employee_gen.go", "node_gen.go"},
		},
		{
			name:        "presence fixture",
			directory:   "presencefixture",
			packageName: "presencefixture",
			sources:     []string{"presence.go"},
			outputs:     []string{"presence_gen.go"},
		},
		{
			name:        "explicit any fixture",
			directory:   "anyfixture",
			packageName: "anyfixture",
			sources:     []string{"values.go"},
			outputs:     []string{"values_gen.go"},
		},
		{
			name:        "fixed array fixture",
			directory:   "arrayfixture",
			packageName: "arrayfixture",
			sources:     []string{"arrays.go"},
			outputs:     []string{"arrays_gen.go", "composite_arrays_gen.go"},
		},
		{
			name:        "named scalar fixture",
			directory:   "namedscalarfixture",
			packageName: "namedscalarfixture",
			sources:     []string{"scalars.go"},
			outputs:     []string{"scalars_gen.go"},
		},
	}

	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {
			helper := NewTestHelper(t)
			for _, source := range fixture.sources {
				content, err := os.ReadFile(filepath.Join(fixture.directory, source))
				require.NoError(t, err)
				helper.CreateGoFile(source, string(content))
			}

			config := &GeneratorConfig{OutputSuffix: "_gen.go", PackageName: fixture.packageName}
			generator, err := NewCodeGenerator(config)
			require.NoError(t, err)
			writer, err := NewFileWriter("", fixture.packageName, config.OutputSuffix, false, false)
			require.NoError(t, err)
			generator.writer = writer
			require.NoError(t, generator.ProcessPackage(helper.GetTempDir()))

			wantFiles := append(append([]string{}, fixture.sources...), fixture.outputs...)
			assert.ElementsMatch(t, wantFiles, helper.ListGeneratedFiles())
			for _, output := range fixture.outputs {
				want, err := os.ReadFile(filepath.Join(fixture.directory, output))
				require.NoError(t, err)
				got, err := os.ReadFile(filepath.Join(helper.GetTempDir(), output))
				require.NoError(t, err)
				assert.Equal(t, want, got, output)
			}
		})
	}
}
