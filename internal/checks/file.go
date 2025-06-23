package checks

import (
	"mime/multipart"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
)

// Mime creates a MIME type validation check for file schemas
// Supports: Mime([]string{"text/plain", "application/json"}, "invalid mime")
func Mime(mimeTypes []string, params ...any) core.ZodCheck {
	checkParams := normalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "mime_type"}
	applyCheckParams(def, checkParams)

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
				setBagProperty(schema, "mime", mimeTypes)
			},
		},
	}
}

func addInvalidValueIssue(mimeTypes []string, payload *core.ParsePayload) {
	values := make([]any, len(mimeTypes))
	for i, v := range mimeTypes {
		values[i] = v
	}
	payload.AddIssue(issues.CreateInvalidValueIssue(values, payload.GetValue()))
}
