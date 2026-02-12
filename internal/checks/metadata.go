package checks

import (
	"github.com/kaptinlin/gozod/core"
)

// ZodCheckDescribe is a no-op check that attaches a description to the schema's
// metadata in the global registry when the check is attached.
//
// TypeScript Zod v4 equivalent: z.describe(description)
// Usage: schema.Check(checks.Describe("User email address"))
type ZodCheckDescribe struct {
	internals *core.ZodCheckInternals
}

// GetZod returns the internal check structure for execution
func (c *ZodCheckDescribe) GetZod() *core.ZodCheckInternals {
	return c.internals
}

// Describe creates a check that registers a description in the global registry.
// This is a no-op validation check that only attaches metadata when the check
// is added to a schema.
//
// TypeScript Zod v4 equivalent: z.describe(description)
//
// Example:
//
//	schema := gozod.String().Check(gozod.Describe("User email"))
//	// => globalRegistry.get(schema).description === "User email"
func Describe(description string) core.ZodCheck {
	internals := &core.ZodCheckInternals{
		Def: &core.ZodCheckDef{
			Check: "describe",
		},
		// No-op check function - describe doesn't validate anything
		Check: func(payload *core.ParsePayload) {},
		// OnAttach callback registers the description in the global registry
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

// ZodCheckMeta is a no-op check that attaches arbitrary metadata to the schema's
// metadata in the global registry when the check is attached.
//
// TypeScript Zod v4 equivalent: z.meta(metadata)
// Usage: schema.Check(checks.Meta(core.GlobalMeta{Title: "Age", Description: "User's age"}))
type ZodCheckMeta struct {
	internals *core.ZodCheckInternals
}

// GetZod returns the internal check structure for execution
func (c *ZodCheckMeta) GetZod() *core.ZodCheckInternals {
	return c.internals
}

// Meta creates a check that registers metadata in the global registry.
// This is a no-op validation check that only attaches metadata when the check
// is added to a schema.
//
// TypeScript Zod v4 equivalent: z.meta(metadata)
//
// Example:
//
//	schema := gozod.Number().Check(gozod.Meta(gozod.GlobalMeta{
//	    Title: "Age",
//	    Description: "User's age",
//	}))
//	// => globalRegistry.get(schema).title === "Age"
//	// => globalRegistry.get(schema).description === "User's age"
func Meta(metadata core.GlobalMeta) core.ZodCheck {
	internals := &core.ZodCheckInternals{
		Def: &core.ZodCheckDef{
			Check: "meta",
		},
		// No-op check function - meta doesn't validate anything
		Check: func(payload *core.ParsePayload) {},
		// OnAttach callback merges the metadata into the global registry
		OnAttach: []func(any){
			func(schema any) {
				if s, ok := schema.(core.ZodSchema); ok {
					existing, _ := core.GlobalRegistry.Get(s)
					// Override with new metadata if set, preserving existing values
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
				}
			},
		},
	}

	return &ZodCheckMeta{internals: internals}
}
