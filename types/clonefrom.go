package types

import "github.com/kaptinlin/gozod/core"

func syncCloneMetadata(source, target core.ZodSchema) {
	if source == nil || target == nil {
		return
	}
	if meta, ok := core.GlobalRegistry.Get(source); ok {
		core.GlobalRegistry.Add(target, meta)
		return
	}
	core.GlobalRegistry.Remove(target)
}

func cloneWithPreservedChecks(source, target core.ZodSchema, clone func()) {
	if source == nil || target == nil || clone == nil {
		return
	}

	originalChecks := target.Internals().Checks
	clone()
	target.Internals().Checks = originalChecks
	syncCloneMetadata(source, target)
}
