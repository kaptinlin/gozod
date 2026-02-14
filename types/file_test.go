package types

import (
	"mime/multipart"
	"os"
	"strings"
	"testing"

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
		tmpFile, err := os.CreateTemp("", "strict_*.txt")
		require.NoError(t, err)
		defer func() {
			_ = tmpFile.Close()
			_ = os.Remove(tmpFile.Name())
		}()

		schema := File()
		result, err := schema.StrictParse(tmpFile)
		require.NoError(t, err)
		assert.Equal(t, tmpFile, result)
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

		schema1 := File().Default(defaultFile).Prefault(prefaultFile)
		result1, err1 := schema1.Parse(nil)
		require.NoError(t, err1)
		file1, ok := result1.(*multipart.FileHeader)
		if !ok {
			t.Fatalf("got %T, want *multipart.FileHeader", result1)
		}
		assert.Equal(t, "default.txt", file1.Filename)

		schema2 := FilePtr().Default(defaultFile).Prefault(prefaultFile)
		result2, err2 := schema2.Parse(nil)
		require.NoError(t, err2)
		require.NotNil(t, result2)
		file2, ok := (*result2).(*multipart.FileHeader)
		if !ok {
			t.Fatalf("got %T, want *multipart.FileHeader", *result2)
		}
		assert.Equal(t, "default.txt", file2.Filename)
	})

	t.Run("Default short-circuits validation", func(t *testing.T) {
		defaultFile := &multipart.FileHeader{
			Filename: "small.txt",
			Header:   make(map[string][]string),
			Size:     100,
		}

		schema1 := File().Min(1000).Default(defaultFile)
		result1, err1 := schema1.Parse(nil)
		require.NoError(t, err1)
		file1, ok := result1.(*multipart.FileHeader)
		if !ok {
			t.Fatalf("got %T, want *multipart.FileHeader", result1)
		}
		assert.Equal(t, int64(100), file1.Size)

		schema2 := FilePtr().Min(1000).Default(defaultFile)
		result2, err2 := schema2.Parse(nil)
		require.NoError(t, err2)
		require.NotNil(t, result2)
		file2, ok := (*result2).(*multipart.FileHeader)
		if !ok {
			t.Fatalf("got %T, want *multipart.FileHeader", *result2)
		}
		assert.Equal(t, int64(100), file2.Size)
	})

	t.Run("Prefault goes through full validation", func(t *testing.T) {
		prefaultFile := &multipart.FileHeader{
			Filename: "large.txt",
			Header:   make(map[string][]string),
			Size:     2000,
		}
		schema1 := File().Min(1000).Prefault(prefaultFile)
		result1, err1 := schema1.Parse(nil)
		require.NoError(t, err1)
		file1, ok := result1.(*multipart.FileHeader)
		if !ok {
			t.Fatalf("got %T, want *multipart.FileHeader", result1)
		}
		assert.Equal(t, int64(2000), file1.Size)

		schema2 := FilePtr().Min(1000).Prefault(prefaultFile)
		result2, err2 := schema2.Parse(nil)
		require.NoError(t, err2)
		require.NotNil(t, result2)
		file2, ok := (*result2).(*multipart.FileHeader)
		if !ok {
			t.Fatalf("got %T, want *multipart.FileHeader", *result2)
		}
		assert.Equal(t, int64(2000), file2.Size)
	})

	t.Run("Prefault only triggered by nil input", func(t *testing.T) {
		prefaultFile := &multipart.FileHeader{
			Filename: "prefault.txt",
			Header:   make(map[string][]string),
			Size:     1024,
		}
		schema := File().Prefault(prefaultFile)

		validFile := &multipart.FileHeader{
			Filename: "valid.txt",
			Header:   make(map[string][]string),
			Size:     2048,
		}
		result, err := schema.Parse(validFile)
		require.NoError(t, err)
		file, ok := result.(*multipart.FileHeader)
		if !ok {
			t.Fatalf("got %T, want *multipart.FileHeader", result)
		}
		assert.Equal(t, "valid.txt", file.Filename)

		_, err = schema.Parse("invalid_input")
		require.Error(t, err)
	})

	t.Run("DefaultFunc and PrefaultFunc behavior", func(t *testing.T) {
		defaultCalled := false
		prefaultCalled := false

		defaultFn := func() any {
			defaultCalled = true
			return &multipart.FileHeader{
				Filename: "default_func.txt",
				Header:   make(map[string][]string),
				Size:     1024,
			}
		}

		prefaultFn := func() any {
			prefaultCalled = true
			return &multipart.FileHeader{
				Filename: "prefault_func.txt",
				Header:   make(map[string][]string),
				Size:     2048,
			}
		}

		schema1 := File().DefaultFunc(defaultFn).PrefaultFunc(prefaultFn)
		result1, err1 := schema1.Parse(nil)
		require.NoError(t, err1)
		assert.True(t, defaultCalled)
		assert.False(t, prefaultCalled)
		file1, ok := result1.(*multipart.FileHeader)
		if !ok {
			t.Fatalf("got %T, want *multipart.FileHeader", result1)
		}
		assert.Equal(t, "default_func.txt", file1.Filename)

		defaultCalled = false
		prefaultCalled = false

		schema2 := File().PrefaultFunc(prefaultFn)
		result2, err2 := schema2.Parse(nil)
		require.NoError(t, err2)
		assert.True(t, prefaultCalled)
		file2, ok := result2.(*multipart.FileHeader)
		if !ok {
			t.Fatalf("got %T, want *multipart.FileHeader", result2)
		}
		assert.Equal(t, "prefault_func.txt", file2.Filename)
	})

	t.Run("Prefault validation failure", func(t *testing.T) {
		prefaultFile := &multipart.FileHeader{
			Filename: "small.txt",
			Header:   make(map[string][]string),
			Size:     100,
		}

		schema := File().Min(1000).Prefault(prefaultFile)
		_, err := schema.Parse(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "File size must be at least 1000 bytes")
	})

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
		file, ok := (*result).(*multipart.FileHeader)
		if !ok {
			t.Fatalf("got %T, want *multipart.FileHeader", *result)
		}
		assert.Equal(t, "default_ptr.txt", file.Filename)
	})
}

// =============================================================================
// OVERWRITE TESTS
// =============================================================================

func TestFile_Overwrite(t *testing.T) {
	t.Run("basic file transformation", func(t *testing.T) {
		schema := File().
			Overwrite(func(f any) any {
				if fh, ok := f.(*multipart.FileHeader); ok {
					return &multipart.FileHeader{
						Filename: "transformed_" + fh.Filename,
						Header:   fh.Header,
						Size:     fh.Size * 2,
					}
				}
				return f
			})

		input := &multipart.FileHeader{
			Filename: "test.txt",
			Header:   make(map[string][]string),
			Size:     1024,
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)
		got, ok := result.(*multipart.FileHeader)
		if !ok {
			t.Fatalf("got %T, want *multipart.FileHeader", result)
		}
		assert.Equal(t, "transformed_test.txt", got.Filename)
		assert.Equal(t, int64(2048), got.Size)
	})

	t.Run("file name normalization", func(t *testing.T) {
		schema := File().
			Overwrite(func(f any) any {
				if fh, ok := f.(*multipart.FileHeader); ok {
					name := strings.ToLower(fh.Filename)
					name = strings.ReplaceAll(name, " ", "_")
					return &multipart.FileHeader{
						Filename: name,
						Header:   fh.Header,
						Size:     fh.Size,
					}
				}
				return f
			})

		input := &multipart.FileHeader{
			Filename: "My Document File.PDF",
			Header:   make(map[string][]string),
			Size:     2048,
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)
		got, ok := result.(*multipart.FileHeader)
		if !ok {
			t.Fatalf("got %T, want *multipart.FileHeader", result)
		}
		assert.Equal(t, "my_document_file.pdf", got.Filename)
	})

	t.Run("os.File transformation", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "test_*.txt")
		require.NoError(t, err)
		defer func() {
			_ = tmpFile.Close()
			_ = os.Remove(tmpFile.Name())
		}()

		schema := File().
			Overwrite(func(f any) any {
				return f
			})

		result, err := schema.Parse(tmpFile)
		require.NoError(t, err)
		assert.Equal(t, tmpFile, result)
	})

	t.Run("pointer type handling", func(t *testing.T) {
		schema := FilePtr().
			Overwrite(func(f *any) *any {
				if f == nil {
					return nil
				}
				if fh, ok := (*f).(*multipart.FileHeader); ok {
					newFH := &multipart.FileHeader{
						Filename: "2024_" + fh.Filename,
						Header:   fh.Header,
						Size:     fh.Size,
					}
					inner := any(newFH)
					return new(any(&inner))
				}
				return f
			})

		input := &multipart.FileHeader{
			Filename: "document.txt",
			Header:   make(map[string][]string),
			Size:     512,
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)
		require.NotNil(t, result)

		anyPtr, ok := (*result).(*any)
		if !ok {
			t.Fatalf("got %T, want *any", *result)
		}
		got, ok := (*anyPtr).(*multipart.FileHeader)
		if !ok {
			t.Fatalf("got %T, want *multipart.FileHeader", *anyPtr)
		}
		assert.Equal(t, "2024_document.txt", got.Filename)
	})

	t.Run("chaining with size validation", func(t *testing.T) {
		schema := File().
			Overwrite(func(f any) any {
				if fh, ok := f.(*multipart.FileHeader); ok {
					if fh.Size < 1000 {
						return &multipart.FileHeader{
							Filename: fh.Filename,
							Header:   fh.Header,
							Size:     1000,
						}
					}
				}
				return f
			})

		input := &multipart.FileHeader{
			Filename: "small.txt",
			Header:   make(map[string][]string),
			Size:     100,
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)
		got, ok := result.(*multipart.FileHeader)
		if !ok {
			t.Fatalf("got %T, want *multipart.FileHeader", result)
		}
		assert.Equal(t, int64(1000), got.Size)
	})

	t.Run("type preservation", func(t *testing.T) {
		schema := File().
			Overwrite(func(f any) any { return f })

		input := &multipart.FileHeader{
			Filename: "test.txt",
			Header:   make(map[string][]string),
			Size:     1024,
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, input, result)
	})

	t.Run("error handling preservation", func(t *testing.T) {
		schema := File().
			Overwrite(func(f any) any { return f })

		_, err := schema.Parse("not a file")
		assert.Error(t, err)

		_, err = schema.Parse(123)
		assert.Error(t, err)
	})

	t.Run("file extension transformation", func(t *testing.T) {
		schema := File().
			Overwrite(func(f any) any {
				if fh, ok := f.(*multipart.FileHeader); ok {
					name := fh.Filename
					if !strings.HasSuffix(strings.ToLower(name), ".txt") {
						name += ".txt"
					}
					return &multipart.FileHeader{
						Filename: name,
						Header:   fh.Header,
						Size:     fh.Size,
					}
				}
				return f
			})

		input := &multipart.FileHeader{
			Filename: "document",
			Header:   make(map[string][]string),
			Size:     1024,
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)
		got, ok := result.(*multipart.FileHeader)
		if !ok {
			t.Fatalf("got %T, want *multipart.FileHeader", result)
		}
		assert.Equal(t, "document.txt", got.Filename)
	})

	t.Run("multiple transformations", func(t *testing.T) {
		schema := File().
			Overwrite(func(f any) any {
				if fh, ok := f.(*multipart.FileHeader); ok {
					return &multipart.FileHeader{
						Filename: strings.ToLower(fh.Filename),
						Header:   fh.Header,
						Size:     fh.Size,
					}
				}
				return f
			}).
			Overwrite(func(f any) any {
				if fh, ok := f.(*multipart.FileHeader); ok {
					return &multipart.FileHeader{
						Filename: "processed_" + fh.Filename,
						Header:   fh.Header,
						Size:     fh.Size,
					}
				}
				return f
			})

		input := &multipart.FileHeader{
			Filename: "MyFile.TXT",
			Header:   make(map[string][]string),
			Size:     1024,
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)
		got, ok := result.(*multipart.FileHeader)
		if !ok {
			t.Fatalf("got %T, want *multipart.FileHeader", result)
		}
		assert.Equal(t, "processed_myfile.txt", got.Filename)
	})
}

// =============================================================================
// Check and With tests
// =============================================================================

func TestFile_Check(t *testing.T) {
	t.Run("check pushes custom issue", func(t *testing.T) {
		schema := File().Check(
			func(v any, p *core.ParsePayload) {
				if fh, ok := v.(*multipart.FileHeader); ok {
					if fh.Size > 1000 {
						p.AddIssue(core.ZodRawIssue{
							Code:    "custom",
							Message: "file too large",
						})
					}
				}
			},
		)

		large := &multipart.FileHeader{
			Filename: "big.txt",
			Header:   make(map[string][]string),
			Size:     2000,
		}
		_, err := schema.Parse(large)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "file too large")

		small := &multipart.FileHeader{
			Filename: "small.txt",
			Header:   make(map[string][]string),
			Size:     500,
		}
		result, err := schema.Parse(small)
		require.NoError(t, err)
		got, ok := result.(*multipart.FileHeader)
		if !ok {
			t.Fatalf("got %T, want *multipart.FileHeader", result)
		}
		assert.Equal(t, "small.txt", got.Filename)
	})

	t.Run("With is alias for Check", func(t *testing.T) {
		schema := File().With(
			func(v any, p *core.ParsePayload) {
				if fh, ok := v.(*multipart.FileHeader); ok {
					if fh.Filename == "" {
						p.AddIssue(core.ZodRawIssue{
							Code:    "custom",
							Message: "filename required",
						})
					}
				}
			},
		)

		noName := &multipart.FileHeader{
			Header: make(map[string][]string),
			Size:   100,
		}
		_, err := schema.Parse(noName)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "filename required")

		named := &multipart.FileHeader{
			Filename: "test.txt",
			Header:   make(map[string][]string),
			Size:     100,
		}
		result, err := schema.Parse(named)
		require.NoError(t, err)
		got, ok := result.(*multipart.FileHeader)
		if !ok {
			t.Fatalf("got %T, want *multipart.FileHeader", result)
		}
		assert.Equal(t, "test.txt", got.Filename)
	})
}
