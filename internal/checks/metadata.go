package checks

import (
	"github.com/kaptinlin/gozod/core"
)

// ZodCheckDescribe is a no-op check that attaches a description to the
// schema's metadata in the global registry.
type ZodCheckDescribe struct {
	internals *core.ZodCheckInternals
}

// GetZod returns the internal check structure for execution.
func (c *ZodCheckDescribe) GetZod() *core.ZodCheckInternals {
	return c.internals
}

// Describe creates a no-op check that registers a description in the
// global registry when attached to a schema.
//
// Example:
//
//	schema := gozod.String().Check(gozod.Describe("User email"))
func Describe(description string) core.ZodCheck {
	internals := &core.ZodCheckInternals{
		Def:   &core.ZodCheckDef{Check: "describe"},
		Check: func(_ *core.ParsePayload) {},
		OnAttach: []func(any){
			func(schema any) {
				if s, ok := schema.(core.ZodSchema); ok {
					existing, _ := core.GlobalRegistry.Get(s)
					existing.Description = description
					core.GlobalRegistry.Add(s, existing)
				}
			},
		},
	}
	return &ZodCheckDescribe{internals: internals}
}

// ZodCheckMeta is a no-op check that attaches arbitrary metadata to the
// schema's metadata in the global registry.
type ZodCheckMeta struct {
	internals *core.ZodCheckInternals
}

// GetZod returns the internal check structure for execution.
func (c *ZodCheckMeta) GetZod() *core.ZodCheckInternals {
	return c.internals
}

// Meta creates a no-op check that merges metadata into the global
// registry when attached to a schema.
//
// Example:
//
//	schema := gozod.Number().Check(gozod.Meta(gozod.GlobalMeta{
//	    Title: "Age",
//	    Description: "User's age",
//	}))
func Meta(metadata core.GlobalMeta) core.ZodCheck {
	internals := &core.ZodCheckInternals{
		Def:   &core.ZodCheckDef{Check: "meta"},
		Check: func(_ *core.ParsePayload) {},
		OnAttach: []func(any){
			func(schema any) {
				s, ok := schema.(core.ZodSchema)
				if !ok {
					return
				}
				existing, _ := core.GlobalRegistry.Get(s)
				if metadata.ID != "" {
					existing.ID = metadata.ID
				}
				if metadata.Title != "" {
					existing.Title = metadata.Title
				}
				if metadata.Description != "" {
					existing.Description = metadata.Description
				}
				if len(metadata.Examples) > 0 {
					existing.Examples = metadata.Examples
				}
				core.GlobalRegistry.Add(s, existing)
			},
		},
	}
	return &ZodCheckMeta{internals: internals}
}
