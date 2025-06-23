package types

import (
	"fmt"
	"mime/multipart"
	"net/textproto"
	"os"
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// 1. Basic functionality and type inference
// =============================================================================

func TestFileBasicFunctionality(t *testing.T) {
	t.Run("constructors", func(t *testing.T) {
		schema := File()
		require.NotNil(t, schema)
		assert.Equal(t, core.ZodTypeFile, schema.GetInternals().Type)

		schema2 := File()
		require.NotNil(t, schema2)
		assert.Equal(t, core.ZodTypeFile, schema2.GetInternals().Type)
	})

	t.Run("accepts valid file types", func(t *testing.T) {
		schema := File()

		// Test multipart.FileHeader
		fileHeader := createTestFileHeader("test.txt", 100, "text/plain")
		result, err := schema.Parse(fileHeader)
		require.NoError(t, err)
		assert.Equal(t, fileHeader, result)

		// Test os.File
		tmpFile := createTestOSFile(t, "test content")
		defer func() { _ = os.Remove(tmpFile.Name()) }()
		defer func() { _ = tmpFile.Close() }()

		result, err = schema.Parse(tmpFile)
		require.NoError(t, err)
		assert.Equal(t, tmpFile, result)
	})

	t.Run("rejects invalid types", func(t *testing.T) {
		schema := File()

		invalidInputs := []any{
			"string",
			123,
			[]byte("bytes"),
			map[string]any{"key": "value"},
		}

		for _, input := range invalidInputs {
			_, err := schema.Parse(input)
			assert.Error(t, err)
		}
	})

	t.Run("MustParse", func(t *testing.T) {
		schema := File()
		fileHeader := createTestFileHeader("test.txt", 100, "text/plain")

		result := schema.MustParse(fileHeader)
		assert.Equal(t, fileHeader, result)

		assert.Panics(t, func() {
			schema.MustParse("invalid")
		})
	})
}

// =============================================================================
// 2. Validation methods
// =============================================================================

func TestFileValidation(t *testing.T) {
	t.Run("min size validation", func(t *testing.T) {
		schema := File().Min(100)

		// Valid case
		largeFile := createTestFileHeader("large.txt", 200, "text/plain")
		result, err := schema.Parse(largeFile)
		require.NoError(t, err)
		assert.Equal(t, largeFile, result)

		// Invalid case
		smallFile := createTestFileHeader("small.txt", 50, "text/plain")
		_, err = schema.Parse(smallFile)
		assert.Error(t, err)
	})

	t.Run("max size validation", func(t *testing.T) {
		schema := File().Max(1000)

		// Valid case
		smallFile := createTestFileHeader("small.txt", 500, "text/plain")
		result, err := schema.Parse(smallFile)
		require.NoError(t, err)
		assert.Equal(t, smallFile, result)

		// Invalid case
		largeFile := createTestFileHeader("large.txt", 1500, "text/plain")
		_, err = schema.Parse(largeFile)
		assert.Error(t, err)
	})

	t.Run("mime type validation", func(t *testing.T) {
		schema := File().Mime([]string{"text/plain", "text/html"})

		// Valid case
		textFile := createTestFileHeader("doc.txt", 100, "text/plain")
		result, err := schema.Parse(textFile)
		require.NoError(t, err)
		assert.Equal(t, textFile, result)

		// Invalid case
		imageFile := createTestFileHeader("image.jpg", 100, "image/jpeg")
		_, err = schema.Parse(imageFile)
		assert.Error(t, err)
	})

	t.Run("combined validation", func(t *testing.T) {
		schema := File().Min(100).Max(1000).Mime([]string{"text/plain"})

		// Valid case
		validFile := createTestFileHeader("valid.txt", 500, "text/plain")
		result, err := schema.Parse(validFile)
		require.NoError(t, err)
		assert.Equal(t, validFile, result)

		// Invalid cases
		tooSmall := createTestFileHeader("small.txt", 50, "text/plain")
		_, err = schema.Parse(tooSmall)
		assert.Error(t, err)

		wrongMime := createTestFileHeader("image.jpg", 500, "image/jpeg")
		_, err = schema.Parse(wrongMime)
		assert.Error(t, err)
	})
}

// =============================================================================
// 3. Modifiers and wrappers
// =============================================================================

func TestFileModifiers(t *testing.T) {
	t.Run("nilable wrapper", func(t *testing.T) {
		schema := File().Nilable()

		// Valid file
		fileHeader := createTestFileHeader("test.txt", 100, "text/plain")
		result, err := schema.Parse(fileHeader)
		require.NoError(t, err)
		assert.Equal(t, fileHeader, result)

		// Nil value
		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("optional wrapper", func(t *testing.T) {
		schema := File().Optional()

		// Valid file
		fileHeader := createTestFileHeader("test.txt", 100, "text/plain")
		result, err := schema.Parse(fileHeader)
		require.NoError(t, err)
		assert.Equal(t, fileHeader, result)

		// Nil value
		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

// =============================================================================
// 4. Chaining and method composition
// =============================================================================

func TestFileChaining(t *testing.T) {
	t.Run("method chaining", func(t *testing.T) {
		schema := File().Min(100).Max(1000).Mime([]string{"text/plain"})

		validFile := createTestFileHeader("valid.txt", 500, "text/plain")
		result, err := schema.Parse(validFile)
		require.NoError(t, err)
		assert.Equal(t, validFile, result)
	})

	t.Run("refine chaining", func(t *testing.T) {
		schema := File().Min(100).RefineAny(func(val any) bool {
			if fh, ok := val.(*multipart.FileHeader); ok {
				return fh.Filename != "forbidden.txt"
			}
			return true
		}, core.SchemaParams{Error: "Forbidden filename"})

		// Valid case
		validFile := createTestFileHeader("allowed.txt", 200, "text/plain")
		result, err := schema.Parse(validFile)
		require.NoError(t, err)
		assert.Equal(t, validFile, result)

		// Invalid case
		forbiddenFile := createTestFileHeader("forbidden.txt", 200, "text/plain")
		_, err = schema.Parse(forbiddenFile)
		assert.Error(t, err)
	})
}

// =============================================================================
// 5. Transform/Pipe
// =============================================================================

func TestFileTransform(t *testing.T) {
	t.Run("transform to filename", func(t *testing.T) {
		schema := File().TransformAny(func(val any, ctx *core.RefinementContext) (any, error) {
			switch f := val.(type) {
			case *multipart.FileHeader:
				return f.Filename, nil
			case *os.File:
				return f.Name(), nil
			default:
				return val, nil
			}
		})

		fileHeader := createTestFileHeader("document.txt", 100, "text/plain")
		result, err := schema.Parse(fileHeader)
		require.NoError(t, err)
		assert.Equal(t, "document.txt", result)
	})

	t.Run("pipe composition", func(t *testing.T) {
		pipe := File().Min(100).Pipe(File().Max(1000))

		validFile := createTestFileHeader("valid.txt", 500, "text/plain")
		result, err := pipe.Parse(validFile)
		require.NoError(t, err)
		assert.Equal(t, validFile, result)

		// Should fail min validation
		smallFile := createTestFileHeader("small.txt", 50, "text/plain")
		_, err = pipe.Parse(smallFile)
		assert.Error(t, err)
	})
}

// =============================================================================
// 6. Refine
// =============================================================================

func TestFileRefine(t *testing.T) {
	t.Run("refine with custom validation", func(t *testing.T) {
		schema := File().RefineAny(func(val any) bool {
			if fh, ok := val.(*multipart.FileHeader); ok {
				return fh.Size > 0 && fh.Filename != ""
			}
			return false
		}, core.SchemaParams{Error: "File must have size and filename"})

		// Valid case
		validFile := createTestFileHeader("valid.txt", 100, "text/plain")
		result, err := schema.Parse(validFile)
		require.NoError(t, err)
		assert.Equal(t, validFile, result)

		// Invalid case
		invalidFile := createTestFileHeader("", 0, "text/plain")
		_, err = schema.Parse(invalidFile)
		assert.Error(t, err)
	})
}

// =============================================================================
// 7. Error handling
// =============================================================================

func TestFileErrorHandling(t *testing.T) {
	t.Run("validation errors", func(t *testing.T) {
		schema := File().Min(1000)
		smallFile := createTestFileHeader("small.txt", 100, "text/plain")

		_, err := schema.Parse(smallFile)
		assert.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.Greater(t, len(zodErr.Issues), 0)
	})

	t.Run("custom error messages", func(t *testing.T) {
		schema := File(core.SchemaParams{Error: "core.Custom file error"})

		_, err := schema.Parse("invalid")
		assert.Error(t, err)
		// core.Custom error handling may vary, just ensure we get an error
		assert.NotEmpty(t, err.Error())
	})
}

// =============================================================================
// 8. Edge cases and internals
// =============================================================================

func TestFileEdgeCases(t *testing.T) {
	t.Run("internals access", func(t *testing.T) {
		schema := File()
		internals := schema.GetInternals()

		assert.Equal(t, core.ZodTypeFile, internals.Type)
		assert.Equal(t, core.Version, internals.Version)

		zodInternals := schema.GetZod()
		assert.NotNil(t, zodInternals)
		assert.Equal(t, core.ZodTypeFile, zodInternals.Def.Type)
	})

	t.Run("nil file header", func(t *testing.T) {
		schema := File().Nilable()

		var nilHeader *multipart.FileHeader
		result, err := schema.Parse(nilHeader)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("zero size file", func(t *testing.T) {
		schema := File()
		zeroFile := createTestFileHeader("empty.txt", 0, "text/plain")

		result, err := schema.Parse(zeroFile)
		require.NoError(t, err)
		assert.Equal(t, zeroFile, result)
	})
}

// =============================================================================
// 9. Integration tests
// =============================================================================

func TestFileIntegration(t *testing.T) {
	t.Run("file upload validation", func(t *testing.T) {
		// Simulate file upload validation
		uploadSchema := File().
			Min(1024).                                 // At least 1KB
			Max(5*1024*1024).                          // Max 5MB
			Mime([]string{"image/jpeg", "image/png"}). // Only images
			RefineAny(func(val any) bool {
				if fh, ok := val.(*multipart.FileHeader); ok {
					// Check filename extension
					filename := fh.Filename
					return len(filename) > 4 &&
						(filename[len(filename)-4:] == ".jpg" ||
							filename[len(filename)-4:] == ".png")
				}
				return false
			}, core.SchemaParams{Error: "Invalid file extension"})

		// Valid upload
		validImage := createTestFileHeader("photo.jpg", 2*1024*1024, "image/jpeg")
		result, err := uploadSchema.Parse(validImage)
		require.NoError(t, err)
		assert.Equal(t, validImage, result)

		// Invalid cases
		tooSmall := createTestFileHeader("tiny.jpg", 512, "image/jpeg")
		_, err = uploadSchema.Parse(tooSmall)
		assert.Error(t, err)

		wrongType := createTestFileHeader("doc.txt", 2*1024*1024, "text/plain")
		_, err = uploadSchema.Parse(wrongType)
		assert.Error(t, err)
	})

	t.Run("file processing pipeline", func(t *testing.T) {
		// Transform file to metadata
		metadataSchema := File().TransformAny(func(val any, ctx *core.RefinementContext) (any, error) {
			if fh, ok := val.(*multipart.FileHeader); ok {
				return map[string]any{
					"filename": fh.Filename,
					"size":     fh.Size,
					"type":     fh.Header.Get("Content-Type"),
				}, nil
			}
			return val, nil
		})

		fileHeader := createTestFileHeader("document.pdf", 1024*1024, "application/pdf")
		result, err := metadataSchema.Parse(fileHeader)
		require.NoError(t, err)

		metadata, ok := result.(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "document.pdf", metadata["filename"])
		assert.Equal(t, int64(1024*1024), metadata["size"])
		assert.Equal(t, "application/pdf", metadata["type"])
	})
}

// =============================================================================
// 10. Default and Prefault tests
// =============================================================================

func TestFileDefaultAndPrefault(t *testing.T) {
	t.Run("default value", func(t *testing.T) {
		defaultFile := createTestFileHeader("default.txt", 100, "text/plain")
		schema := File().Default(defaultFile)

		// Valid file should not use default
		validFile := createTestFileHeader("valid.txt", 200, "text/html")
		result, err := schema.Parse(validFile)
		require.NoError(t, err)
		assert.Equal(t, validFile, result)

		// nil input should use default
		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, defaultFile, result)
	})

	t.Run("function-based default", func(t *testing.T) {
		counter := 0
		schema := File().DefaultFunc(func() any {
			counter++
			return createTestFileHeader(fmt.Sprintf("generated-%d.txt", counter), 100, "text/plain")
		})

		// Each nil input generates a new default value
		result1, err1 := schema.Parse(nil)
		require.NoError(t, err1)
		file1, ok1 := result1.(*multipart.FileHeader)
		require.True(t, ok1)
		assert.Equal(t, "generated-1.txt", file1.Filename)

		result2, err2 := schema.Parse(nil)
		require.NoError(t, err2)
		file2, ok2 := result2.(*multipart.FileHeader)
		require.True(t, ok2)
		assert.Equal(t, "generated-2.txt", file2.Filename)

		// Valid input bypasses default generation
		validFile := createTestFileHeader("valid.txt", 200, "text/html")
		result3, err3 := schema.Parse(validFile)
		require.NoError(t, err3)
		assert.Equal(t, validFile, result3)

		// Counter should only increment for nil inputs
		assert.Equal(t, 2, counter)
	})

	t.Run("prefault value", func(t *testing.T) {
		fallbackFile := createTestFileHeader("fallback.txt", 1500, "text/plain")
		schema := File().Min(1000).Prefault(fallbackFile)

		// Valid file should pass through
		validFile := createTestFileHeader("large.txt", 2000, "text/html")
		result, err := schema.Parse(validFile)
		require.NoError(t, err)
		assert.Equal(t, validFile, result)

		// Invalid file should use prefault
		smallFile := createTestFileHeader("small.txt", 100, "text/plain")
		result, err = schema.Parse(smallFile)
		require.NoError(t, err)
		assert.Equal(t, fallbackFile, result)

		// Invalid type should use prefault
		result, err = schema.Parse("not a file")
		require.NoError(t, err)
		assert.Equal(t, fallbackFile, result)
	})

	t.Run("prefault function", func(t *testing.T) {
		counter := 0
		schema := File().Min(1000).PrefaultFunc(func() any {
			counter++
			return createTestFileHeader(fmt.Sprintf("fallback-%d.txt", counter), 1500, "text/plain")
		})

		// Valid file should not call function
		validFile := createTestFileHeader("large.txt", 2000, "text/html")
		result, err := schema.Parse(validFile)
		require.NoError(t, err)
		assert.Equal(t, validFile, result)
		assert.Equal(t, 0, counter)

		// Invalid file should call function
		smallFile := createTestFileHeader("small.txt", 100, "text/plain")
		result, err = schema.Parse(smallFile)
		require.NoError(t, err)
		fallbackFile, ok := result.(*multipart.FileHeader)
		require.True(t, ok)
		assert.Equal(t, "fallback-1.txt", fallbackFile.Filename)
		assert.Equal(t, int64(1500), fallbackFile.Size)
		assert.Equal(t, 1, counter)

		// Another invalid input should increment counter
		result, err = schema.Parse("invalid")
		require.NoError(t, err)
		fallbackFile2, ok := result.(*multipart.FileHeader)
		require.True(t, ok)
		assert.Equal(t, "fallback-2.txt", fallbackFile2.Filename)
		assert.Equal(t, 2, counter)
	})

	t.Run("default vs prefault distinction", func(t *testing.T) {
		defaultFile := createTestFileHeader("default.txt", 100, "text/plain")
		fallbackFile := createTestFileHeader("fallback.txt", 1500, "text/plain")

		defaultSchema := File().Default(defaultFile)
		prefaultSchema := File().Min(1000).Prefault(fallbackFile)

		// For valid input: different behaviors
		validFile := createTestFileHeader("valid.txt", 2000, "text/html")
		result1, err1 := defaultSchema.Parse(validFile)
		require.NoError(t, err1)
		assert.Equal(t, validFile, result1, "Default: valid input passes through")

		result2, err2 := prefaultSchema.Parse(validFile)
		require.NoError(t, err2)
		assert.Equal(t, validFile, result2, "Prefault: valid input passes through")

		// For nil input: different behaviors
		result3, err3 := defaultSchema.Parse(nil)
		require.NoError(t, err3)
		assert.Equal(t, defaultFile, result3, "Default: nil gets default value")

		result4, err4 := prefaultSchema.Parse(nil)
		require.NoError(t, err4)
		assert.Equal(t, fallbackFile, result4, "Prefault: nil fails validation, use fallback")

		// For invalid file size: different behaviors
		smallFile := createTestFileHeader("small.txt", 100, "text/plain")
		result5, err5 := defaultSchema.Parse(smallFile)
		require.NoError(t, err5)
		assert.Equal(t, smallFile, result5, "Default: valid file type passes through")

		result6, err6 := prefaultSchema.Parse(smallFile)
		require.NoError(t, err6)
		assert.Equal(t, fallbackFile, result6, "Prefault: validation fails, use fallback")
	})

	t.Run("default with chaining", func(t *testing.T) {
		defaultFile := createTestFileHeader("default.txt", 2000, "text/plain")
		schema := File().Default(defaultFile).Min(1000)

		// nil input: use default, should pass validation
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, defaultFile, result)

		// Valid file: should pass
		validFile := createTestFileHeader("valid.txt", 1500, "text/html")
		result, err = schema.Parse(validFile)
		require.NoError(t, err)
		assert.Equal(t, validFile, result)

		// Invalid file: should fail validation
		smallFile := createTestFileHeader("small.txt", 500, "text/plain")
		_, err = schema.Parse(smallFile)
		assert.Error(t, err)
	})

	t.Run("default with transform compatibility", func(t *testing.T) {
		defaultFile := createTestFileHeader("default.txt", 1024, "text/plain")
		schema := File().
			Default(defaultFile).
			TransformAny(func(input any, ctx *core.RefinementContext) (any, error) {
				if fh, ok := input.(*multipart.FileHeader); ok {
					return map[string]any{
						"filename": fh.Filename,
						"size":     fh.Size,
						"type":     fh.Header.Get("Content-Type"),
						"sizeKB":   fh.Size / 1024,
					}, nil
				}
				return input, nil
			})

		// Non-nil input: should transform
		validFile := createTestFileHeader("valid.txt", 2048, "text/html")
		result1, err1 := schema.Parse(validFile)
		require.NoError(t, err1)
		result1Map, ok1 := result1.(map[string]any)
		require.True(t, ok1)
		assert.Equal(t, "valid.txt", result1Map["filename"])
		assert.Equal(t, int64(2048), result1Map["size"])
		assert.Equal(t, int64(2), result1Map["sizeKB"])

		// nil input: use default then transform
		result2, err2 := schema.Parse(nil)
		require.NoError(t, err2)
		result2Map, ok2 := result2.(map[string]any)
		require.True(t, ok2)
		assert.Equal(t, "default.txt", result2Map["filename"])
		assert.Equal(t, int64(1024), result2Map["size"])
		assert.Equal(t, int64(1), result2Map["sizeKB"])
	})

	t.Run("os.File support", func(t *testing.T) {
		// Test default with os.File
		tmpFile := createTestOSFile(t, "test content")
		defer func() { _ = os.Remove(tmpFile.Name()) }()
		defer func() { _ = tmpFile.Close() }()

		schema := File().Default(tmpFile)

		// nil input should use default os.File
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, tmpFile, result)

		// Valid multipart.FileHeader should pass through
		fileHeader := createTestFileHeader("test.txt", 100, "text/plain")
		result, err = schema.Parse(fileHeader)
		require.NoError(t, err)
		assert.Equal(t, fileHeader, result)
	})
}

// =============================================================================
// 11. Consistency and boundary tests
// =============================================================================

func TestFileBoundaryConditions(t *testing.T) {
	t.Run("inclusive min size", func(t *testing.T) {
		schema := File().Min(5)

		// Size exactly at the lower boundary should succeed
		exact := createTestFileHeader("exact_min.txt", 5, "text/plain")
		result, err := schema.Parse(exact)
		require.NoError(t, err)
		assert.Equal(t, exact, result)

		// Size below the boundary should fail
		tooSmall := createTestFileHeader("too_small.txt", 4, "text/plain")
		_, err = schema.Parse(tooSmall)
		assert.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.Equal(t, issues.TooSmall, zodErr.Issues[0].Code)
	})

	t.Run("inclusive max size", func(t *testing.T) {
		schema := File().Max(8)

		// Size exactly at the upper boundary should succeed
		exact := createTestFileHeader("exact_max.txt", 8, "text/plain")
		result, err := schema.Parse(exact)
		require.NoError(t, err)
		assert.Equal(t, exact, result)

		// Size above the boundary should fail
		tooLarge := createTestFileHeader("too_large.txt", 9, "text/plain")
		_, err = schema.Parse(tooLarge)
		assert.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.Equal(t, issues.TooBig, zodErr.Issues[0].Code)
	})
}

// =============================================================================
// 12. Exact size validation
// =============================================================================

func TestFileExactSizeValidation(t *testing.T) {
	t.Run("exact size pass and fail", func(t *testing.T) {
		schema := File().Size(100)

		// Matching size should pass
		exact := createTestFileHeader("exact_size.txt", 100, "application/json")
		result, err := schema.Parse(exact)
		require.NoError(t, err)
		assert.Equal(t, exact, result)

		// Non-matching size should fail
		mismatch := createTestFileHeader("mismatch.txt", 120, "application/json")
		_, err = schema.Parse(mismatch)
		assert.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		// Either TooSmall or TooBig are acceptable depending on the direction of the mismatch
		hasSizeIssue := zodErr.Issues[0].Code == issues.TooSmall || zodErr.Issues[0].Code == issues.TooBig
		assert.True(t, hasSizeIssue, "Expected size constraint issue")
	})
}

// =============================================================================
// 13. Multiple MIME type validation
// =============================================================================

func TestFileMimeValidationMultipleTypes(t *testing.T) {
	schema := File().Mime([]string{"text/plain", "application/json"})

	// Valid MIME types
	validPlain := createTestFileHeader("plain.txt", 10, "text/plain")
	_, err := schema.Parse(validPlain)
	require.NoError(t, err)

	validJSON := createTestFileHeader("data.json", 10, "application/json")
	_, err = schema.Parse(validJSON)
	require.NoError(t, err)

	// Invalid MIME type
	invalid := createTestFileHeader("image.jpg", 10, "image/jpeg")
	_, err = schema.Parse(invalid)
	assert.Error(t, err)

	var zodErr *issues.ZodError
	require.True(t, issues.IsZodError(err, &zodErr))
	assert.Equal(t, issues.InvalidValue, zodErr.Issues[0].Code)
}

// =============================================================================
// Test helper functions
// =============================================================================

// createTestFileHeader creates a multipart.FileHeader for testing
func createTestFileHeader(filename string, size int64, contentType string) *multipart.FileHeader {
	header := textproto.MIMEHeader{}
	if contentType != "" {
		header.Set("Content-Type", contentType)
	}

	return &multipart.FileHeader{
		Filename: filename,
		Size:     size,
		Header:   header,
	}
}

// createTestOSFile creates a temporary os.File for testing
func createTestOSFile(t *testing.T, content string) *os.File {
	t.Helper()

	tmpFile, err := os.CreateTemp("", "gozod_test_*.txt")
	require.NoError(t, err)

	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)

	// Reset file pointer to beginning
	_, err = tmpFile.Seek(0, 0)
	require.NoError(t, err)

	return tmpFile
}
