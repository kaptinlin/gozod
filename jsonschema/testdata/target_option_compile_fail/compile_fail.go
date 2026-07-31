package compilefail

import "github.com/kaptinlin/gozod"

func invalidTargetOption() {
	_, _ = gozod.ToJSONSchema(gozod.String(), gozod.JSONSchemaOptions{
		Target: gozod.JSONSchemaTargetDraft202012,
	})
}
