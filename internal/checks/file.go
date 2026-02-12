package checks

import (
	"mime/multipart"
	"os"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
)

// MinFileSize creates a minimum file size validation check.
func MinFileSize(minimum int64, params ...any) core.ZodCheck {
	cp := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "min_file_size"}
	ApplyCheckParams(def, cp)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			size := getFileSize(payload.Value())
			if size < minimum {
				payload.AddIssue(issues.CreateTooSmallIssue(minimum, false, "file", payload.Value()))
			}
		},
		OnAttach: []func(any){
			func(schema any) { SetBagProperty(schema, "minSize", minimum) },
		},
	}
}

// MaxFileSize creates a maximum file size validation check.
func MaxFileSize(maximum int64, params ...any) core.ZodCheck {
	cp := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "max_file_size"}
	ApplyCheckParams(def, cp)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			size := getFileSize(payload.Value())
			if size > maximum {
				payload.AddIssue(issues.CreateTooBigIssue(maximum, false, "file", payload.Value()))
			}
		},
		OnAttach: []func(any){
			func(schema any) { SetBagProperty(schema, "maxSize", maximum) },
		},
	}
}

// FileSize creates an exact file size validation check.
func FileSize(expected int64, params ...any) core.ZodCheck {
	cp := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "file_size_equals"}
	ApplyCheckParams(def, cp)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			size := getFileSize(payload.Value())
			if size != expected {
				if size > expected {
					payload.AddIssue(issues.CreateTooBigIssue(expected, true, "file", payload.Value()))
				} else {
					payload.AddIssue(issues.CreateTooSmallIssue(expected, true, "file", payload.Value()))
				}
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				SetBagProperty(schema, "minSize", expected)
				SetBagProperty(schema, "maxSize", expected)
				SetBagProperty(schema, "size", expected)
			},
		},
	}
}

// Mime creates a MIME type validation check for file schemas.
func Mime(mimeTypes []string, params ...any) core.ZodCheck {
	cp := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "mime_type"}
	ApplyCheckParams(def, cp)

	// Convert list to a set for efficient lookup
	allowed := make(map[string]struct{}, len(mimeTypes))
	for _, m := range mimeTypes {
		allowed[m] = struct{}{}
	}

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			var contentType string
			value := payload.Value()
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
			func(schema any) { SetBagProperty(schema, "mime", mimeTypes) },
		},
	}
}

// getFileSize returns file size in bytes, or 0 if unknown.
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
	payload.AddIssue(issues.CreateInvalidValueIssue(values, payload.Value()))
}
