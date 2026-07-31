package compilefail

import "github.com/kaptinlin/gozod"

func invalidRegistryCall() {
	registry := gozod.NewRegistry[gozod.GlobalMeta]()
	_, _ = gozod.ToJSONSchema(registry)
}
