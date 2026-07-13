package types

import "github.com/kaptinlin/gozod/core"

func finalizeClone(target core.ZodSchema) {
	core.AttachChecks(target)
}

func cloneWithPreservedChecks(source, target core.ZodSchema, clone func()) {
	if source == nil || target == nil || clone == nil {
		return
	}

	originalChecks := target.Internals().Checks
	clone()
	target.Internals().Checks = originalChecks
}
