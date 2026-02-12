package types

import (
	"os"
	"strings"
	"testing"

	"mime/multipart"

	"github.com/kaptinlin/gozod/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFile_BasicFunctionality(t *testing.T) {
	t.Run("File factory", func(t *testing.T) {
		schema := File()
		require.NotNil(t, schema)
	})

	t.Run("FilePtr factory", func(t *testing.T) {
		schema := FilePtr()
		require.NotNil(t, schema)
	})

	t.Run("FileTyped factory", func(t *testing.T) {
		schema := FileTyped[any, any]()
		require.NotNil(t, schema)
	})

	t.Run("File with params", func(t *testing.T) {
		schema := File(core.SchemaParams{
			Error: "Custom error",
		})
		require.NotNil(t, schema)
	})
}

func TestFile_Modifiers(t *testing.T) {
	t.Run("Optional accepts nil", func(t *testing.T) {
		schema := File().Optional()
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Nilable accepts nil", func(t *testing.T) {
		schema := File().Nilable()
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Nullish accepts nil", func(t *testing.T) {
		schema := File().Nullish()
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("ExactOptional", func(t *testing.T) {
		schema := File().ExactOptional()
		require.NotNil(t, schema)
	})

	t.Run("Default preserves file type", func(t *testing.T) {
		schema := File().Default("default_value")
		require.NotNil(t, schema)
	})
}

func TestFile_ValidationMethods(t *testing.T) {
	t.Run("Min size validation", func(t *testing.T) {
		schema := File().Min(1024)

		require.NotNil(t, schema)
	})

	t.Run("Max size validation", func(t *testing.T) {
		schema := File().Max(2048)

		require.NotNil(t, schema)
	})

	t.Run("Exact size validation", func(t *testing.T) {
		schema := File().Size(1024)

		require.NotNil(t, schema)
	})

	t.Run("MIME type validation", func(t *testing.T) {
		schema := File().Mime([]string{"image/jpeg", "image/png"})

		require.NotNil(t, schema)
	})
}

func TestFile_Refine(t *testing.T) {
	t.Run("basic refinement", func(t *testing.T) {
		schema := File().Refine(func(v any) bool {
			// Accept any file for testing
			return true
		})

		require.NotNil(t, schema)
	})

	t.Run("refineAny validation", func(t *testing.T) {
		schema := File().RefineAny(func(v any) bool {
			// Accept any file for testing
			return true
		})

		require.NotNil(t, schema)
	})
}

func TestFile_TypeSafety(t *testing.T) {
	t.Run("type checking", func(t *testing.T) {
		schema := File()

		// Test with invalid input (should fail type check)
		_, err := schema.Parse("not a file")
		assert.Error(t, err, "Expected error for non-file input")

		// Test with nil (should fail unless optional)
		_, err = schema.Parse(nil)
		assert.Error(t, err, "Expected error for nil input on non-optional schema")
	})

	t.Run("optional nil handling", func(t *testing.T) {
		schema := File().Optional()

		// Test with nil (should succeed for optional)
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestFile_Composition(t *testing.T) {
	t.Run("And creates intersection", func(t *testing.T) {
		schema := File()
		intersection := schema.And(String())
		require.NotNil(t, intersection)
	})

	t.Run("Or creates union", func(t *testing.T) {
		schema := File()
		union := schema.Or(String())
		require.NotNil(t, union)
	})
}

func TestFile_Describe(t *testing.T) {
	t.Run("sets description", func(t *testing.T) {
		schema := File().Describe("upload file")
		require.NotNil(t, schema)

		meta, ok := core.GlobalRegistry.Get(schema)
		require.True(t, ok)
		assert.Equal(t, "upload file", meta.Description)
	})
}

func TestFile_Meta(t *testing.T) {
	t.Run("stores metadata", func(t *testing.T) {
		schema := File().Meta(core.GlobalMeta{
			Description: "file upload",
		})
		require.NotNil(t, schema)

		meta, ok := core.GlobalRegistry.Get(schema)
		require.True(t, ok)
		assert.Equal(t, "file upload", meta.Description)
	})
}

func TestFile_StrictParse(t *testing.T) {
	t.Run("accepts valid os.File", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "strict_*.txt")
		require.NoError(t, err)
		defer func() {
			_ = tempFile.Close()
			_ = os.Remove(tempFile.Name())
		}()

		schema := File()
		result, err := schema.StrictParse(tempFile)
		require.NoError(t, err)
		assert.Equal(t, tempFile, result)
	})

	t.Run("accepts multipart file header", func(t *testing.T) {
		header := &multipart.FileHeader{
			Filename: "test.txt",
			Header:   make(map[string][]string),
			Size:     1024,
		}

		schema := File()
		result, err := schema.StrictParse(header)
		require.NoError(t, err)
		assert.Equal(t, header, result)
	})

	t.Run("MustParse panics on nil non-optional", func(t *testing.T) {
		schema := File()
		assert.Panics(t, func() {
			schema.MustParse(nil)
		})
	})
}

// =============================================================================
// Default and prefault tests
// =============================================================================

func TestFile_DefaultAndPrefault(t *testing.T) {
	// Test 1: Default has higher priority than Prefault
	t.Run("Default has higher priority than Prefault", func(t *testing.T) {
		defaultFile := &multipart.FileHeader{
			Filename: "default.txt",
			Header:   make(map[string][]string),
			Size:     1024,
		}
		prefaultFile := &multipart.FileHeader{
			Filename: "prefault.txt",
			Header:   make(map[string][]string),
			Size:     2048,
		}

		// File type
		schema1 := File().Default(defaultFile).Prefault(prefaultFile)
		result1, err1 := schema1.Parse(nil)
		require.NoError(t, err1)
		if file, ok := result1.(*multipart.FileHeader); ok {
			assert.Equal(t, "default.txt", file.Filename) // Should be default, not prefault
		} else {
			t.Fatal("Expected multipart.FileHeader")
		}

		// FilePtr type
		schema2 := FilePtr().Default(defaultFile).Prefault(prefaultFile)
		result2, err2 := schema2.Parse(nil)
		require.NoError(t, err2)
		require.NotNil(t, result2)
		if file, ok := (*result2).(*multipart.FileHeader); ok {
			assert.Equal(t, "default.txt", file.Filename) // Should be default, not prefault
		} else {
			t.Fatal("Expected multipart.FileHeader")
		}
	})

	// Test 2: Default short-circuits validation
	t.Run("Default short-circuits validation", func(t *testing.T) {
		// Create a default file that would fail size validation
		defaultFile := &multipart.FileHeader{
			Filename: "small.txt",
			Header:   make(map[string][]string),
			Size:     100, // Too small for Min(1000)
		}

		// File type - default value violates Min(1000) but should still work
		schema1 := File().Min(1000).Default(defaultFile)
		result1, err1 := schema1.Parse(nil)
		require.NoError(t, err1)
		if file, ok := result1.(*multipart.FileHeader); ok {
			assert.Equal(t, int64(100), file.Size) // Default bypasses validation
		} else {
			t.Fatal("Expected multipart.FileHeader")
		}

		// FilePtr type - default value violates Min(1000) but should still work
		schema2 := FilePtr().Min(1000).Default(defaultFile)
		result2, err2 := schema2.Parse(nil)
		require.NoError(t, err2)
		require.NotNil(t, result2)
		if file, ok := (*result2).(*multipart.FileHeader); ok {
			assert.Equal(t, int64(100), file.Size) // Default bypasses validation
		} else {
			t.Fatal("Expected multipart.FileHeader")
		}
	})

	// Test 3: Prefault goes through full validation
	t.Run("Prefault goes through full validation", func(t *testing.T) {
		// File type - prefault value passes validation
		prefaultFile := &multipart.FileHeader{
			Filename: "large.txt",
			Header:   make(map[string][]string),
			Size:     2000, // Passes Min(1000)
		}
		schema1 := File().Min(1000).Prefault(prefaultFile)
		result1, err1 := schema1.Parse(nil)
		require.NoError(t, err1)
		if file, ok := result1.(*multipart.FileHeader); ok {
			assert.Equal(t, int64(2000), file.Size)
		} else {
			t.Fatal("Expected multipart.FileHeader")
		}

		// FilePtr type - prefault value passes validation
		schema2 := FilePtr().Min(1000).Prefault(prefaultFile)
		result2, err2 := schema2.Parse(nil)
		require.NoError(t, err2)
		require.NotNil(t, result2)
		if file, ok := (*result2).(*multipart.FileHeader); ok {
			assert.Equal(t, int64(2000), file.Size)
		} else {
			t.Fatal("Expected multipart.FileHeader")
		}
	})

	// Test 4: Prefault only triggered by nil input
	t.Run("Prefault only triggered by nil input", func(t *testing.T) {
		prefaultFile := &multipart.FileHeader{
			Filename: "prefault.txt",
			Header:   make(map[string][]string),
			Size:     1024,
		}
		schema := File().Prefault(prefaultFile)

		// Valid input should override prefault
		validFile := &multipart.FileHeader{
			Filename: "valid.txt",
			Header:   make(map[string][]string),
			Size:     2048,
		}
		result, err := schema.Parse(validFile)
		require.NoError(t, err)
		if file, ok := result.(*multipart.FileHeader); ok {
			assert.Equal(t, "valid.txt", file.Filename) // Should be input, not prefault
		} else {
			t.Fatal("Expected multipart.FileHeader")
		}

		// Invalid input should NOT trigger prefault (should return error)
		_, err = schema.Parse("invalid_input")
		require.Error(t, err)
	})

	// Test 5: DefaultFunc and PrefaultFunc behavior
	t.Run("DefaultFunc and PrefaultFunc behavior", func(t *testing.T) {
		defaultCalled := false
		prefaultCalled := false

		defaultFunc := func() any {
			defaultCalled = true
			return &multipart.FileHeader{
				Filename: "default_func.txt",
				Header:   make(map[string][]string),
				Size:     1024,
			}
		}

		prefaultFunc := func() any {
			prefaultCalled = true
			return &multipart.FileHeader{
				Filename: "prefault_func.txt",
				Header:   make(map[string][]string),
				Size:     2048,
			}
		}

		// Test DefaultFunc priority over PrefaultFunc
		schema1 := File().DefaultFunc(defaultFunc).PrefaultFunc(prefaultFunc)
		result1, err1 := schema1.Parse(nil)
		require.NoError(t, err1)
		assert.True(t, defaultCalled)
		assert.False(t, prefaultCalled) // Should not be called due to default priority
		if file, ok := result1.(*multipart.FileHeader); ok {
			assert.Equal(t, "default_func.txt", file.Filename)
		} else {
			t.Fatal("Expected multipart.FileHeader")
		}

		// Reset flags
		defaultCalled = false
		prefaultCalled = false

		// Test PrefaultFunc alone
		schema2 := File().PrefaultFunc(prefaultFunc)
		result2, err2 := schema2.Parse(nil)
		require.NoError(t, err2)
		assert.True(t, prefaultCalled)
		if file, ok := result2.(*multipart.FileHeader); ok {
			assert.Equal(t, "prefault_func.txt", file.Filename)
		} else {
			t.Fatal("Expected multipart.FileHeader")
		}
	})

	// Test 6: Prefault validation failure
	t.Run("Prefault validation failure", func(t *testing.T) {
		// Create a prefault file that fails validation
		prefaultFile := &multipart.FileHeader{
			Filename: "small.txt",
			Header:   make(map[string][]string),
			Size:     100, // Too small for Min(1000)
		}

		schema := File().Min(1000).Prefault(prefaultFile)
		_, err := schema.Parse(nil)
		require.Error(t, err) // Prefault should fail validation
		assert.Contains(t, err.Error(), "File size must be at least 1000 bytes")
	})

	// Test 7: FilePtr with Default and Prefault
	t.Run("FilePtr with Default and Prefault", func(t *testing.T) {
		defaultFile := &multipart.FileHeader{
			Filename: "default_ptr.txt",
			Header:   make(map[string][]string),
			Size:     1024,
		}
		prefaultFile := &multipart.FileHeader{
			Filename: "prefault_ptr.txt",
			Header:   make(map[string][]string),
			Size:     2048,
		}

		schema := FilePtr().Default(defaultFile).Prefault(prefaultFile)
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		if file, ok := (*result).(*multipart.FileHeader); ok {
			assert.Equal(t, "default_ptr.txt", file.Filename) // Should be default
		} else {
			t.Fatal("Expected multipart.FileHeader")
		}
	})
}

// =============================================================================
// OVERWRITE TESTS
// =============================================================================

func TestFile_Overwrite(t *testing.T) {
	t.Run("basic file transformation", func(t *testing.T) {
		schema := File().
			Overwrite(func(file any) any {
				// Transform file size validation
				if multipartFile, ok := file.(*multipart.FileHeader); ok {
					// Create a new file header with modified size for testing
					newFile := &multipart.FileHeader{
						Filename: "transformed_" + multipartFile.Filename,
						Header:   multipartFile.Header,
						Size:     multipartFile.Size * 2, // Double the size
					}
					return newFile
				}
				return file
			})

		// Create a test multipart file
		originalFile := &multipart.FileHeader{
			Filename: "test.txt",
			Header:   make(map[string][]string),
			Size:     1024,
		}

		result, err := schema.Parse(originalFile)
		require.NoError(t, err)

		if transformedFile, ok := result.(*multipart.FileHeader); ok {
			assert.Equal(t, "transformed_test.txt", transformedFile.Filename)
			assert.Equal(t, int64(2048), transformedFile.Size)
		} else {
			t.Fatal("Expected transformed multipart file")
		}
	})

	t.Run("file name normalization", func(t *testing.T) {
		schema := File().
			Overwrite(func(file any) any {
				if multipartFile, ok := file.(*multipart.FileHeader); ok {
					// Normalize filename: lowercase and replace spaces with underscores
					normalizedName := strings.ToLower(multipartFile.Filename)
					normalizedName = strings.ReplaceAll(normalizedName, " ", "_")

					newFile := &multipart.FileHeader{
						Filename: normalizedName,
						Header:   multipartFile.Header,
						Size:     multipartFile.Size,
					}
					return newFile
				}
				return file
			})

		originalFile := &multipart.FileHeader{
			Filename: "My Document File.PDF",
			Header:   make(map[string][]string),
			Size:     2048,
		}

		result, err := schema.Parse(originalFile)
		require.NoError(t, err)

		if transformedFile, ok := result.(*multipart.FileHeader); ok {
			assert.Equal(t, "my_document_file.pdf", transformedFile.Filename)
		} else {
			t.Fatal("Expected transformed file")
		}
	})

	t.Run("os.File transformation", func(t *testing.T) {
		// Create a temporary file for testing
		tempFile, err := os.CreateTemp("", "test_*.txt")
		require.NoError(t, err)
		defer func() {
			err := tempFile.Close()
			if err != nil {
				// Handle or log the error if necessary
				t.Logf("Failed to close temp file: %v", err)
			}
			err = os.Remove(tempFile.Name())
			if err != nil {
				// Handle or log the error if necessary
				t.Logf("Failed to remove temp file: %v", err)
			}
		}()

		schema := File().
			Overwrite(func(file any) any {
				// For os.File, we can't easily transform it, so return as-is
				// In practice, you might want to wrap it or change permissions
				return file
			})

		result, err := schema.Parse(tempFile)
		require.NoError(t, err)

		// Should preserve the file
		assert.Equal(t, tempFile, result)
	})

	t.Run("pointer type handling", func(t *testing.T) {
		schema := FilePtr().
			Overwrite(func(file *any) *any {
				if file == nil {
					return nil
				}

				if multipartFile, ok := (*file).(*multipart.FileHeader); ok {
					// Add timestamp prefix to filename
					newFile := &multipart.FileHeader{
						Filename: "2024_" + multipartFile.Filename,
						Header:   multipartFile.Header,
						Size:     multipartFile.Size,
					}
					// Create nested pointer structure: *any -> *any -> *multipart.FileHeader
					innerAny := any(newFile)
					result := any(&innerAny)
					return &result
				}
				return file
			})

		// Use the actual file type, not wrapped in any
		originalFile := &multipart.FileHeader{
			Filename: "document.txt",
			Header:   make(map[string][]string),
			Size:     512,
		}

		result, err := schema.Parse(originalFile)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Handle the nested pointer structure: *any -> *any -> *multipart.FileHeader
		if anyPtr, ok := (*result).(*any); ok {
			if transformedFile, ok := (*anyPtr).(*multipart.FileHeader); ok {
				assert.Equal(t, "2024_document.txt", transformedFile.Filename)
			} else {
				t.Fatalf("Expected transformed file, got: %T", *anyPtr)
			}
		} else {
			t.Fatalf("Expected *any, got: %T", *result)
		}
	})

	t.Run("chaining with size validation", func(t *testing.T) {
		schema := File().
			Overwrite(func(file any) any {
				// Ensure minimum file size by padding metadata
				if multipartFile, ok := file.(*multipart.FileHeader); ok {
					minSize := int64(1000)
					if multipartFile.Size < minSize {
						newFile := &multipart.FileHeader{
							Filename: multipartFile.Filename,
							Header:   multipartFile.Header,
							Size:     minSize, // Enforce minimum size
						}
						return newFile
					}
				}
				return file
			})

		smallFile := &multipart.FileHeader{
			Filename: "small.txt",
			Header:   make(map[string][]string),
			Size:     100, // Below minimum
		}

		result, err := schema.Parse(smallFile)
		require.NoError(t, err)

		if transformedFile, ok := result.(*multipart.FileHeader); ok {
			assert.Equal(t, int64(1000), transformedFile.Size) // Should be padded to minimum
		} else {
			t.Fatal("Expected transformed file")
		}
	})

	t.Run("type preservation", func(t *testing.T) {
		schema := File().
			Overwrite(func(file any) any {
				return file // Identity transformation
			})

		testFile := &multipart.FileHeader{
			Filename: "test.txt",
			Header:   make(map[string][]string),
			Size:     1024,
		}

		result, err := schema.Parse(testFile)
		require.NoError(t, err)
		assert.Equal(t, testFile, result)
	})

	t.Run("error handling preservation", func(t *testing.T) {
		schema := File().
			Overwrite(func(file any) any {
				return file
			})

		// Invalid input should still fail validation
		_, err := schema.Parse("not a file")
		assert.Error(t, err)

		_, err = schema.Parse(123)
		assert.Error(t, err)
	})

	t.Run("file extension transformation", func(t *testing.T) {
		schema := File().
			Overwrite(func(file any) any {
				if multipartFile, ok := file.(*multipart.FileHeader); ok {
					// Ensure .txt extension
					filename := multipartFile.Filename
					if !strings.HasSuffix(strings.ToLower(filename), ".txt") {
						filename += ".txt"
					}

					newFile := &multipart.FileHeader{
						Filename: filename,
						Header:   multipartFile.Header,
						Size:     multipartFile.Size,
					}
					return newFile
				}
				return file
			})

		fileWithoutExt := &multipart.FileHeader{
			Filename: "document",
			Header:   make(map[string][]string),
			Size:     1024,
		}

		result, err := schema.Parse(fileWithoutExt)
		require.NoError(t, err)

		if transformedFile, ok := result.(*multipart.FileHeader); ok {
			assert.Equal(t, "document.txt", transformedFile.Filename)
		} else {
			t.Fatal("Expected transformed file")
		}
	})

	t.Run("multiple transformations", func(t *testing.T) {
		schema := File().
			Overwrite(func(file any) any {
				// First transformation: normalize filename
				if multipartFile, ok := file.(*multipart.FileHeader); ok {
					newFile := &multipart.FileHeader{
						Filename: strings.ToLower(multipartFile.Filename),
						Header:   multipartFile.Header,
						Size:     multipartFile.Size,
					}
					return newFile
				}
				return file
			}).
			Overwrite(func(file any) any {
				// Second transformation: add prefix
				if multipartFile, ok := file.(*multipart.FileHeader); ok {
					newFile := &multipart.FileHeader{
						Filename: "processed_" + multipartFile.Filename,
						Header:   multipartFile.Header,
						Size:     multipartFile.Size,
					}
					return newFile
				}
				return file
			})

		originalFile := &multipart.FileHeader{
			Filename: "MyFile.TXT",
			Header:   make(map[string][]string),
			Size:     1024,
		}

		result, err := schema.Parse(originalFile)
		require.NoError(t, err)

		if transformedFile, ok := result.(*multipart.FileHeader); ok {
			// Should be lowercase with prefix
			assert.Equal(t, "processed_myfile.txt", transformedFile.Filename)
		} else {
			t.Fatal("Expected transformed file")
		}
	})
}
