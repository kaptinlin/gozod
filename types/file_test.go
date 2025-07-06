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
		var _ *ZodFile[any, any] = schema

		require.NotNil(t, schema)
	})

	t.Run("FilePtr factory", func(t *testing.T) {
		schema := FilePtr()
		var _ *ZodFile[any, *any] = schema

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
	t.Run("Optional behavior", func(t *testing.T) {
		schema := File()
		optionalSchema := schema.Optional()

		// Type check: ensure it returns *ZodFile[any, *any]
		var _ *ZodFile[any, *any] = optionalSchema

		require.NotNil(t, optionalSchema)
	})

	t.Run("Nilable behavior", func(t *testing.T) {
		schema := File()
		nilableSchema := schema.Nilable()

		var _ *ZodFile[any, *any] = nilableSchema

		require.NotNil(t, nilableSchema)
	})

	t.Run("Default preserves file type", func(t *testing.T) {
		schema := File()
		defaultSchema := schema.Default("default_value")

		var _ *ZodFile[any, any] = defaultSchema

		require.NotNil(t, defaultSchema)
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

// Note: Full file testing with actual os.File or multipart.FileHeader would require
// more complex setup. This basic test ensures the dual generic parameter architecture works.

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
			}
			err = os.Remove(tempFile.Name())
			if err != nil {
				// Handle or log the error if necessary
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
					result := any(newFile)
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
