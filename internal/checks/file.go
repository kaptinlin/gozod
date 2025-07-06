package checks

import (
	"mime/multipart"
	"os"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
)

// =============================================================================
// FILE SIZE VALIDATION FUNCTIONS
// =============================================================================

// MinFileSize creates a minimum file size validation check
// Supports: MinFileSize(1024, "file too small") or MinFileSize(1024, CheckParams{Error: "minimum size required"})
func MinFileSize(minimum int64, params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "min_file_size"}
	ApplyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			size := getFileSize(payload.GetValue())
			if size < minimum {
				payload.AddIssue(issues.CreateTooSmallIssue(minimum, false, "file", payload.GetValue()))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Set minimum size for JSON Schema
				SetBagProperty(schema, "minSize", minimum)
			},
		},
	}
}

// MaxFileSize creates a maximum file size validation check
// Supports: MaxFileSize(5242880, "file too large") or MaxFileSize(5242880, CheckParams{Error: "maximum size exceeded"})
func MaxFileSize(maximum int64, params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "max_file_size"}
	ApplyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			size := getFileSize(payload.GetValue())
			if size > maximum {
				payload.AddIssue(issues.CreateTooBigIssue(maximum, false, "file", payload.GetValue()))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Set maximum size for JSON Schema
				SetBagProperty(schema, "maxSize", maximum)
			},
		},
	}
}

// FileSize creates an exact file size validation check
// Supports: FileSize(1024, "file must be exactly 1KB") or FileSize(1024, CheckParams{Error: "exact size required"})
func FileSize(expected int64, params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "file_size_equals"}
	ApplyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			size := getFileSize(payload.GetValue())
			if size != expected {
				// Determine if too big or too small based on actual size
				if size > expected {
					payload.AddIssue(issues.CreateTooBigIssue(expected, true, "file", payload.GetValue()))
				} else {
					payload.AddIssue(issues.CreateTooSmallIssue(expected, true, "file", payload.GetValue()))
				}
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Exact size sets both min and max
				SetBagProperty(schema, "minSize", expected)
				SetBagProperty(schema, "maxSize", expected)
				SetBagProperty(schema, "size", expected)
			},
		},
	}
}

// =============================================================================
// MIME TYPE VALIDATION FUNCTIONS
// =============================================================================

// Mime creates a MIME type validation check for file schemas
// Supports: Mime([]string{"text/plain", "application/json"}, "invalid mime")
func Mime(mimeTypes []string, params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "mime_type"}
	ApplyCheckParams(def, checkParams)

	// Convert list to a set for efficient lookup
	allowed := make(map[string]struct{}, len(mimeTypes))
	for _, m := range mimeTypes {
		allowed[m] = struct{}{}
	}

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			var contentType string
			value := payload.GetValue()
			switch f := value.(type) {
			case *multipart.FileHeader:
				contentType = f.Header.Get("Content-Type")
			case multipart.FileHeader:
				contentType = f.Header.Get("Content-Type")
			default:
				// Not a file header, treat as invalid
				addInvalidValueIssue(mimeTypes, payload)
				return
			}

			if _, ok := allowed[contentType]; !ok {
				addInvalidValueIssue(mimeTypes, payload)
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Save mime list to schema bag (for JSON Schema etc.)
				SetBagProperty(schema, "mime", mimeTypes)
			},
		},
	}
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// getFileSize returns file size in bytes (0 if unknown)
func getFileSize(v any) int64 {
	switch f := v.(type) {
	case *multipart.FileHeader:
		return f.Size
	case multipart.FileHeader:
		return f.Size
	case *os.File:
		if stat, err := f.Stat(); err == nil {
			return stat.Size()
		}
	case os.File:
		if stat, err := f.Stat(); err == nil {
			return stat.Size()
		}
	}
	return 0
}

func addInvalidValueIssue(mimeTypes []string, payload *core.ParsePayload) {
	values := make([]any, len(mimeTypes))
	for i, v := range mimeTypes {
		values[i] = v
	}
	payload.AddIssue(issues.CreateInvalidValueIssue(values, payload.GetValue()))
}
