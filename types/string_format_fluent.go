package types

import "github.com/kaptinlin/gozod/core"

func wrapStringFluent[T StringConstraint, W core.ZodSchema](source core.ZodSchema, base *ZodString[T], wrap func(*ZodString[T]) W) W {
	wrapped := wrap(base)
	core.CopyGlobalMeta(source, wrapped)
	return wrapped
}

func withStringWrapperMeta[T StringConstraint, W core.ZodSchema](source core.ZodSchema, base *ZodString[T], wrap func(*ZodString[T]) W, meta core.GlobalMeta) W {
	clone := wrap(base.withInternals(base.Internals().Clone()))
	existing, ok := core.GlobalRegistry.Get(source)
	if !ok {
		core.GlobalRegistry.Add(clone, meta)
		return clone
	}
	core.GlobalRegistry.Add(clone, mergeGlobalMeta(existing, meta))
	return clone
}

func mergeGlobalMeta(existing, meta core.GlobalMeta) core.GlobalMeta {
	if meta.ID != "" {
		existing.ID = meta.ID
	}
	if meta.Title != "" {
		existing.Title = meta.Title
	}
	if meta.Description != "" {
		existing.Description = meta.Description
	}
	if len(meta.Examples) > 0 {
		existing.Examples = meta.Examples
	}
	return existing
}

// Default sets a fallback value returned when input is nil.
func (z *ZodEmail[T]) Default(v string) *ZodEmail[T] {
	return wrapStringFluent(z, z.ZodString.Default(v), newEmail[T])
}

// DefaultFunc sets a fallback function called when input is nil.
func (z *ZodEmail[T]) DefaultFunc(fn func() string) *ZodEmail[T] {
	return wrapStringFluent(z, z.ZodString.DefaultFunc(fn), newEmail[T])
}

// Prefault sets a fallback value that goes through the full validation pipeline.
func (z *ZodEmail[T]) Prefault(v string) *ZodEmail[T] {
	return wrapStringFluent(z, z.ZodString.Prefault(v), newEmail[T])
}

// PrefaultFunc sets a fallback function that goes through the full validation pipeline.
func (z *ZodEmail[T]) PrefaultFunc(fn func() string) *ZodEmail[T] {
	return wrapStringFluent(z, z.ZodString.PrefaultFunc(fn), newEmail[T])
}

// Meta attaches metadata to this schema via the global registry.
func (z *ZodEmail[T]) Meta(meta core.GlobalMeta) *ZodEmail[T] {
	return withStringWrapperMeta(z, z.ZodString, newEmail[T], meta)
}

// Describe attaches a description to this schema via the global registry.
func (z *ZodEmail[T]) Describe(description string) *ZodEmail[T] {
	return z.Meta(core.GlobalMeta{Description: description})
}

// RefineAny adds custom validation that receives the raw value as any.
func (z *ZodEmail[T]) RefineAny(fn func(any) bool, params ...any) *ZodEmail[T] {
	return wrapStringFluent(z, z.ZodString.RefineAny(fn, params...), newEmail[T])
}

// Default sets a fallback value returned when input is nil.
func (z *ZodGUID[T]) Default(v string) *ZodGUID[T] {
	return wrapStringFluent(z, z.ZodString.Default(v), newGUID[T])
}

// DefaultFunc sets a fallback function called when input is nil.
func (z *ZodGUID[T]) DefaultFunc(fn func() string) *ZodGUID[T] {
	return wrapStringFluent(z, z.ZodString.DefaultFunc(fn), newGUID[T])
}

// Prefault sets a fallback value that goes through the full validation pipeline.
func (z *ZodGUID[T]) Prefault(v string) *ZodGUID[T] {
	return wrapStringFluent(z, z.ZodString.Prefault(v), newGUID[T])
}

// PrefaultFunc sets a fallback function that goes through the full validation pipeline.
func (z *ZodGUID[T]) PrefaultFunc(fn func() string) *ZodGUID[T] {
	return wrapStringFluent(z, z.ZodString.PrefaultFunc(fn), newGUID[T])
}

// Meta attaches metadata to this schema via the global registry.
func (z *ZodGUID[T]) Meta(meta core.GlobalMeta) *ZodGUID[T] {
	return withStringWrapperMeta(z, z.ZodString, newGUID[T], meta)
}

// Describe attaches a description to this schema via the global registry.
func (z *ZodGUID[T]) Describe(description string) *ZodGUID[T] {
	return z.Meta(core.GlobalMeta{Description: description})
}

// RefineAny adds custom validation that receives the raw value as any.
func (z *ZodGUID[T]) RefineAny(fn func(any) bool, params ...any) *ZodGUID[T] {
	return wrapStringFluent(z, z.ZodString.RefineAny(fn, params...), newGUID[T])
}

// Default sets a fallback value returned when input is nil.
func (z *ZodCUID[T]) Default(v string) *ZodCUID[T] {
	return wrapStringFluent(z, z.ZodString.Default(v), newCUID[T])
}

// DefaultFunc sets a fallback function called when input is nil.
func (z *ZodCUID[T]) DefaultFunc(fn func() string) *ZodCUID[T] {
	return wrapStringFluent(z, z.ZodString.DefaultFunc(fn), newCUID[T])
}

// Prefault sets a fallback value that goes through the full validation pipeline.
func (z *ZodCUID[T]) Prefault(v string) *ZodCUID[T] {
	return wrapStringFluent(z, z.ZodString.Prefault(v), newCUID[T])
}

// PrefaultFunc sets a fallback function that goes through the full validation pipeline.
func (z *ZodCUID[T]) PrefaultFunc(fn func() string) *ZodCUID[T] {
	return wrapStringFluent(z, z.ZodString.PrefaultFunc(fn), newCUID[T])
}

// Meta attaches metadata to this schema via the global registry.
func (z *ZodCUID[T]) Meta(meta core.GlobalMeta) *ZodCUID[T] {
	return withStringWrapperMeta(z, z.ZodString, newCUID[T], meta)
}

// Describe attaches a description to this schema via the global registry.
func (z *ZodCUID[T]) Describe(description string) *ZodCUID[T] {
	return z.Meta(core.GlobalMeta{Description: description})
}

// RefineAny adds custom validation that receives the raw value as any.
func (z *ZodCUID[T]) RefineAny(fn func(any) bool, params ...any) *ZodCUID[T] {
	return wrapStringFluent(z, z.ZodString.RefineAny(fn, params...), newCUID[T])
}

// Default sets a fallback value returned when input is nil.
func (z *ZodCUID2[T]) Default(v string) *ZodCUID2[T] {
	return wrapStringFluent(z, z.ZodString.Default(v), newCUID2[T])
}

// DefaultFunc sets a fallback function called when input is nil.
func (z *ZodCUID2[T]) DefaultFunc(fn func() string) *ZodCUID2[T] {
	return wrapStringFluent(z, z.ZodString.DefaultFunc(fn), newCUID2[T])
}

// Prefault sets a fallback value that goes through the full validation pipeline.
func (z *ZodCUID2[T]) Prefault(v string) *ZodCUID2[T] {
	return wrapStringFluent(z, z.ZodString.Prefault(v), newCUID2[T])
}

// PrefaultFunc sets a fallback function that goes through the full validation pipeline.
func (z *ZodCUID2[T]) PrefaultFunc(fn func() string) *ZodCUID2[T] {
	return wrapStringFluent(z, z.ZodString.PrefaultFunc(fn), newCUID2[T])
}

// Meta attaches metadata to this schema via the global registry.
func (z *ZodCUID2[T]) Meta(meta core.GlobalMeta) *ZodCUID2[T] {
	return withStringWrapperMeta(z, z.ZodString, newCUID2[T], meta)
}

// Describe attaches a description to this schema via the global registry.
func (z *ZodCUID2[T]) Describe(description string) *ZodCUID2[T] {
	return z.Meta(core.GlobalMeta{Description: description})
}

// RefineAny adds custom validation that receives the raw value as any.
func (z *ZodCUID2[T]) RefineAny(fn func(any) bool, params ...any) *ZodCUID2[T] {
	return wrapStringFluent(z, z.ZodString.RefineAny(fn, params...), newCUID2[T])
}

// Default sets a fallback value returned when input is nil.
func (z *ZodULID[T]) Default(v string) *ZodULID[T] {
	return wrapStringFluent(z, z.ZodString.Default(v), newULID[T])
}

// DefaultFunc sets a fallback function called when input is nil.
func (z *ZodULID[T]) DefaultFunc(fn func() string) *ZodULID[T] {
	return wrapStringFluent(z, z.ZodString.DefaultFunc(fn), newULID[T])
}

// Prefault sets a fallback value that goes through the full validation pipeline.
func (z *ZodULID[T]) Prefault(v string) *ZodULID[T] {
	return wrapStringFluent(z, z.ZodString.Prefault(v), newULID[T])
}

// PrefaultFunc sets a fallback function that goes through the full validation pipeline.
func (z *ZodULID[T]) PrefaultFunc(fn func() string) *ZodULID[T] {
	return wrapStringFluent(z, z.ZodString.PrefaultFunc(fn), newULID[T])
}

// Meta attaches metadata to this schema via the global registry.
func (z *ZodULID[T]) Meta(meta core.GlobalMeta) *ZodULID[T] {
	return withStringWrapperMeta(z, z.ZodString, newULID[T], meta)
}

// Describe attaches a description to this schema via the global registry.
func (z *ZodULID[T]) Describe(description string) *ZodULID[T] {
	return z.Meta(core.GlobalMeta{Description: description})
}

// RefineAny adds custom validation that receives the raw value as any.
func (z *ZodULID[T]) RefineAny(fn func(any) bool, params ...any) *ZodULID[T] {
	return wrapStringFluent(z, z.ZodString.RefineAny(fn, params...), newULID[T])
}

// Default sets a fallback value returned when input is nil.
func (z *ZodXID[T]) Default(v string) *ZodXID[T] {
	return wrapStringFluent(z, z.ZodString.Default(v), newXID[T])
}

// DefaultFunc sets a fallback function called when input is nil.
func (z *ZodXID[T]) DefaultFunc(fn func() string) *ZodXID[T] {
	return wrapStringFluent(z, z.ZodString.DefaultFunc(fn), newXID[T])
}

// Prefault sets a fallback value that goes through the full validation pipeline.
func (z *ZodXID[T]) Prefault(v string) *ZodXID[T] {
	return wrapStringFluent(z, z.ZodString.Prefault(v), newXID[T])
}

// PrefaultFunc sets a fallback function that goes through the full validation pipeline.
func (z *ZodXID[T]) PrefaultFunc(fn func() string) *ZodXID[T] {
	return wrapStringFluent(z, z.ZodString.PrefaultFunc(fn), newXID[T])
}

// Meta attaches metadata to this schema via the global registry.
func (z *ZodXID[T]) Meta(meta core.GlobalMeta) *ZodXID[T] {
	return withStringWrapperMeta(z, z.ZodString, newXID[T], meta)
}

// Describe attaches a description to this schema via the global registry.
func (z *ZodXID[T]) Describe(description string) *ZodXID[T] {
	return z.Meta(core.GlobalMeta{Description: description})
}

// RefineAny adds custom validation that receives the raw value as any.
func (z *ZodXID[T]) RefineAny(fn func(any) bool, params ...any) *ZodXID[T] {
	return wrapStringFluent(z, z.ZodString.RefineAny(fn, params...), newXID[T])
}

// Default sets a fallback value returned when input is nil.
func (z *ZodKSUID[T]) Default(v string) *ZodKSUID[T] {
	return wrapStringFluent(z, z.ZodString.Default(v), newKSUID[T])
}

// DefaultFunc sets a fallback function called when input is nil.
func (z *ZodKSUID[T]) DefaultFunc(fn func() string) *ZodKSUID[T] {
	return wrapStringFluent(z, z.ZodString.DefaultFunc(fn), newKSUID[T])
}

// Prefault sets a fallback value that goes through the full validation pipeline.
func (z *ZodKSUID[T]) Prefault(v string) *ZodKSUID[T] {
	return wrapStringFluent(z, z.ZodString.Prefault(v), newKSUID[T])
}

// PrefaultFunc sets a fallback function that goes through the full validation pipeline.
func (z *ZodKSUID[T]) PrefaultFunc(fn func() string) *ZodKSUID[T] {
	return wrapStringFluent(z, z.ZodString.PrefaultFunc(fn), newKSUID[T])
}

// Meta attaches metadata to this schema via the global registry.
func (z *ZodKSUID[T]) Meta(meta core.GlobalMeta) *ZodKSUID[T] {
	return withStringWrapperMeta(z, z.ZodString, newKSUID[T], meta)
}

// Describe attaches a description to this schema via the global registry.
func (z *ZodKSUID[T]) Describe(description string) *ZodKSUID[T] {
	return z.Meta(core.GlobalMeta{Description: description})
}

// RefineAny adds custom validation that receives the raw value as any.
func (z *ZodKSUID[T]) RefineAny(fn func(any) bool, params ...any) *ZodKSUID[T] {
	return wrapStringFluent(z, z.ZodString.RefineAny(fn, params...), newKSUID[T])
}

// Default sets a fallback value returned when input is nil.
func (z *ZodNanoID[T]) Default(v string) *ZodNanoID[T] {
	return wrapStringFluent(z, z.ZodString.Default(v), newNanoID[T])
}

// DefaultFunc sets a fallback function called when input is nil.
func (z *ZodNanoID[T]) DefaultFunc(fn func() string) *ZodNanoID[T] {
	return wrapStringFluent(z, z.ZodString.DefaultFunc(fn), newNanoID[T])
}

// Prefault sets a fallback value that goes through the full validation pipeline.
func (z *ZodNanoID[T]) Prefault(v string) *ZodNanoID[T] {
	return wrapStringFluent(z, z.ZodString.Prefault(v), newNanoID[T])
}

// PrefaultFunc sets a fallback function that goes through the full validation pipeline.
func (z *ZodNanoID[T]) PrefaultFunc(fn func() string) *ZodNanoID[T] {
	return wrapStringFluent(z, z.ZodString.PrefaultFunc(fn), newNanoID[T])
}

// Meta attaches metadata to this schema via the global registry.
func (z *ZodNanoID[T]) Meta(meta core.GlobalMeta) *ZodNanoID[T] {
	return withStringWrapperMeta(z, z.ZodString, newNanoID[T], meta)
}

// Describe attaches a description to this schema via the global registry.
func (z *ZodNanoID[T]) Describe(description string) *ZodNanoID[T] {
	return z.Meta(core.GlobalMeta{Description: description})
}

// RefineAny adds custom validation that receives the raw value as any.
func (z *ZodNanoID[T]) RefineAny(fn func(any) bool, params ...any) *ZodNanoID[T] {
	return wrapStringFluent(z, z.ZodString.RefineAny(fn, params...), newNanoID[T])
}

// Default sets a fallback value returned when input is nil.
func (z *ZodUUID[T]) Default(v string) *ZodUUID[T] {
	return wrapStringFluent(z, z.ZodString.Default(v), newUUID[T])
}

// DefaultFunc sets a fallback function called when input is nil.
func (z *ZodUUID[T]) DefaultFunc(fn func() string) *ZodUUID[T] {
	return wrapStringFluent(z, z.ZodString.DefaultFunc(fn), newUUID[T])
}

// Prefault sets a fallback value that goes through the full validation pipeline.
func (z *ZodUUID[T]) Prefault(v string) *ZodUUID[T] {
	return wrapStringFluent(z, z.ZodString.Prefault(v), newUUID[T])
}

// PrefaultFunc sets a fallback function that goes through the full validation pipeline.
func (z *ZodUUID[T]) PrefaultFunc(fn func() string) *ZodUUID[T] {
	return wrapStringFluent(z, z.ZodString.PrefaultFunc(fn), newUUID[T])
}

// Meta attaches metadata to this schema via the global registry.
func (z *ZodUUID[T]) Meta(meta core.GlobalMeta) *ZodUUID[T] {
	return withStringWrapperMeta(z, z.ZodString, newUUID[T], meta)
}

// Describe attaches a description to this schema via the global registry.
func (z *ZodUUID[T]) Describe(description string) *ZodUUID[T] {
	return z.Meta(core.GlobalMeta{Description: description})
}

// RefineAny adds custom validation that receives the raw value as any.
func (z *ZodUUID[T]) RefineAny(fn func(any) bool, params ...any) *ZodUUID[T] {
	return wrapStringFluent(z, z.ZodString.RefineAny(fn, params...), newUUID[T])
}

// Default sets a fallback value returned when input is nil.
func (z *ZodIPv4[T]) Default(v string) *ZodIPv4[T] {
	return wrapStringFluent(z, z.ZodString.Default(v), newIPv4[T])
}

// DefaultFunc sets a fallback function called when input is nil.
func (z *ZodIPv4[T]) DefaultFunc(fn func() string) *ZodIPv4[T] {
	return wrapStringFluent(z, z.ZodString.DefaultFunc(fn), newIPv4[T])
}

// Prefault sets a fallback value that goes through the full validation pipeline.
func (z *ZodIPv4[T]) Prefault(v string) *ZodIPv4[T] {
	return wrapStringFluent(z, z.ZodString.Prefault(v), newIPv4[T])
}

// PrefaultFunc sets a fallback function that goes through the full validation pipeline.
func (z *ZodIPv4[T]) PrefaultFunc(fn func() string) *ZodIPv4[T] {
	return wrapStringFluent(z, z.ZodString.PrefaultFunc(fn), newIPv4[T])
}

// Meta attaches metadata to this schema via the global registry.
func (z *ZodIPv4[T]) Meta(meta core.GlobalMeta) *ZodIPv4[T] {
	return withStringWrapperMeta(z, z.ZodString, newIPv4[T], meta)
}

// Describe attaches a description to this schema via the global registry.
func (z *ZodIPv4[T]) Describe(description string) *ZodIPv4[T] {
	return z.Meta(core.GlobalMeta{Description: description})
}

// RefineAny adds custom validation that receives the raw value as any.
func (z *ZodIPv4[T]) RefineAny(fn func(any) bool, params ...any) *ZodIPv4[T] {
	return wrapStringFluent(z, z.ZodString.RefineAny(fn, params...), newIPv4[T])
}

// Default sets a fallback value returned when input is nil.
func (z *ZodIPv6[T]) Default(v string) *ZodIPv6[T] {
	return wrapStringFluent(z, z.ZodString.Default(v), newIPv6[T])
}

// DefaultFunc sets a fallback function called when input is nil.
func (z *ZodIPv6[T]) DefaultFunc(fn func() string) *ZodIPv6[T] {
	return wrapStringFluent(z, z.ZodString.DefaultFunc(fn), newIPv6[T])
}

// Prefault sets a fallback value that goes through the full validation pipeline.
func (z *ZodIPv6[T]) Prefault(v string) *ZodIPv6[T] {
	return wrapStringFluent(z, z.ZodString.Prefault(v), newIPv6[T])
}

// PrefaultFunc sets a fallback function that goes through the full validation pipeline.
func (z *ZodIPv6[T]) PrefaultFunc(fn func() string) *ZodIPv6[T] {
	return wrapStringFluent(z, z.ZodString.PrefaultFunc(fn), newIPv6[T])
}

// Meta attaches metadata to this schema via the global registry.
func (z *ZodIPv6[T]) Meta(meta core.GlobalMeta) *ZodIPv6[T] {
	return withStringWrapperMeta(z, z.ZodString, newIPv6[T], meta)
}

// Describe attaches a description to this schema via the global registry.
func (z *ZodIPv6[T]) Describe(description string) *ZodIPv6[T] {
	return z.Meta(core.GlobalMeta{Description: description})
}

// RefineAny adds custom validation that receives the raw value as any.
func (z *ZodIPv6[T]) RefineAny(fn func(any) bool, params ...any) *ZodIPv6[T] {
	return wrapStringFluent(z, z.ZodString.RefineAny(fn, params...), newIPv6[T])
}

// Default sets a fallback value returned when input is nil.
func (z *ZodCIDRv4[T]) Default(v string) *ZodCIDRv4[T] {
	return wrapStringFluent(z, z.ZodString.Default(v), newCIDRv4[T])
}

// DefaultFunc sets a fallback function called when input is nil.
func (z *ZodCIDRv4[T]) DefaultFunc(fn func() string) *ZodCIDRv4[T] {
	return wrapStringFluent(z, z.ZodString.DefaultFunc(fn), newCIDRv4[T])
}

// Prefault sets a fallback value that goes through the full validation pipeline.
func (z *ZodCIDRv4[T]) Prefault(v string) *ZodCIDRv4[T] {
	return wrapStringFluent(z, z.ZodString.Prefault(v), newCIDRv4[T])
}

// PrefaultFunc sets a fallback function that goes through the full validation pipeline.
func (z *ZodCIDRv4[T]) PrefaultFunc(fn func() string) *ZodCIDRv4[T] {
	return wrapStringFluent(z, z.ZodString.PrefaultFunc(fn), newCIDRv4[T])
}

// Meta attaches metadata to this schema via the global registry.
func (z *ZodCIDRv4[T]) Meta(meta core.GlobalMeta) *ZodCIDRv4[T] {
	return withStringWrapperMeta(z, z.ZodString, newCIDRv4[T], meta)
}

// Describe attaches a description to this schema via the global registry.
func (z *ZodCIDRv4[T]) Describe(description string) *ZodCIDRv4[T] {
	return z.Meta(core.GlobalMeta{Description: description})
}

// RefineAny adds custom validation that receives the raw value as any.
func (z *ZodCIDRv4[T]) RefineAny(fn func(any) bool, params ...any) *ZodCIDRv4[T] {
	return wrapStringFluent(z, z.ZodString.RefineAny(fn, params...), newCIDRv4[T])
}

// Default sets a fallback value returned when input is nil.
func (z *ZodCIDRv6[T]) Default(v string) *ZodCIDRv6[T] {
	return wrapStringFluent(z, z.ZodString.Default(v), newCIDRv6[T])
}

// DefaultFunc sets a fallback function called when input is nil.
func (z *ZodCIDRv6[T]) DefaultFunc(fn func() string) *ZodCIDRv6[T] {
	return wrapStringFluent(z, z.ZodString.DefaultFunc(fn), newCIDRv6[T])
}

// Prefault sets a fallback value that goes through the full validation pipeline.
func (z *ZodCIDRv6[T]) Prefault(v string) *ZodCIDRv6[T] {
	return wrapStringFluent(z, z.ZodString.Prefault(v), newCIDRv6[T])
}

// PrefaultFunc sets a fallback function that goes through the full validation pipeline.
func (z *ZodCIDRv6[T]) PrefaultFunc(fn func() string) *ZodCIDRv6[T] {
	return wrapStringFluent(z, z.ZodString.PrefaultFunc(fn), newCIDRv6[T])
}

// Meta attaches metadata to this schema via the global registry.
func (z *ZodCIDRv6[T]) Meta(meta core.GlobalMeta) *ZodCIDRv6[T] {
	return withStringWrapperMeta(z, z.ZodString, newCIDRv6[T], meta)
}

// Describe attaches a description to this schema via the global registry.
func (z *ZodCIDRv6[T]) Describe(description string) *ZodCIDRv6[T] {
	return z.Meta(core.GlobalMeta{Description: description})
}

// RefineAny adds custom validation that receives the raw value as any.
func (z *ZodCIDRv6[T]) RefineAny(fn func(any) bool, params ...any) *ZodCIDRv6[T] {
	return wrapStringFluent(z, z.ZodString.RefineAny(fn, params...), newCIDRv6[T])
}

// Default sets a fallback value returned when input is nil.
func (z *ZodURL[T]) Default(v string) *ZodURL[T] {
	return wrapStringFluent(z, z.ZodString.Default(v), newURL[T])
}

// DefaultFunc sets a fallback function called when input is nil.
func (z *ZodURL[T]) DefaultFunc(fn func() string) *ZodURL[T] {
	return wrapStringFluent(z, z.ZodString.DefaultFunc(fn), newURL[T])
}

// Prefault sets a fallback value that goes through the full validation pipeline.
func (z *ZodURL[T]) Prefault(v string) *ZodURL[T] {
	return wrapStringFluent(z, z.ZodString.Prefault(v), newURL[T])
}

// PrefaultFunc sets a fallback function that goes through the full validation pipeline.
func (z *ZodURL[T]) PrefaultFunc(fn func() string) *ZodURL[T] {
	return wrapStringFluent(z, z.ZodString.PrefaultFunc(fn), newURL[T])
}

// Meta attaches metadata to this schema via the global registry.
func (z *ZodURL[T]) Meta(meta core.GlobalMeta) *ZodURL[T] {
	return withStringWrapperMeta(z, z.ZodString, newURL[T], meta)
}

// Describe attaches a description to this schema via the global registry.
func (z *ZodURL[T]) Describe(description string) *ZodURL[T] {
	return z.Meta(core.GlobalMeta{Description: description})
}

// RefineAny adds custom validation that receives the raw value as any.
func (z *ZodURL[T]) RefineAny(fn func(any) bool, params ...any) *ZodURL[T] {
	return wrapStringFluent(z, z.ZodString.RefineAny(fn, params...), newURL[T])
}

// Default sets a fallback value returned when input is nil.
func (z *ZodHostname[T]) Default(v string) *ZodHostname[T] {
	return wrapStringFluent(z, z.ZodString.Default(v), newHostname[T])
}

// DefaultFunc sets a fallback function called when input is nil.
func (z *ZodHostname[T]) DefaultFunc(fn func() string) *ZodHostname[T] {
	return wrapStringFluent(z, z.ZodString.DefaultFunc(fn), newHostname[T])
}

// Prefault sets a fallback value that goes through the full validation pipeline.
func (z *ZodHostname[T]) Prefault(v string) *ZodHostname[T] {
	return wrapStringFluent(z, z.ZodString.Prefault(v), newHostname[T])
}

// PrefaultFunc sets a fallback function that goes through the full validation pipeline.
func (z *ZodHostname[T]) PrefaultFunc(fn func() string) *ZodHostname[T] {
	return wrapStringFluent(z, z.ZodString.PrefaultFunc(fn), newHostname[T])
}

// Meta attaches metadata to this schema via the global registry.
func (z *ZodHostname[T]) Meta(meta core.GlobalMeta) *ZodHostname[T] {
	return withStringWrapperMeta(z, z.ZodString, newHostname[T], meta)
}

// Describe attaches a description to this schema via the global registry.
func (z *ZodHostname[T]) Describe(description string) *ZodHostname[T] {
	return z.Meta(core.GlobalMeta{Description: description})
}

// RefineAny adds custom validation that receives the raw value as any.
func (z *ZodHostname[T]) RefineAny(fn func(any) bool, params ...any) *ZodHostname[T] {
	return wrapStringFluent(z, z.ZodString.RefineAny(fn, params...), newHostname[T])
}

// Default sets a fallback value returned when input is nil.
func (z *ZodMAC[T]) Default(v string) *ZodMAC[T] {
	return wrapStringFluent(z, z.ZodString.Default(v), newMAC[T])
}

// DefaultFunc sets a fallback function called when input is nil.
func (z *ZodMAC[T]) DefaultFunc(fn func() string) *ZodMAC[T] {
	return wrapStringFluent(z, z.ZodString.DefaultFunc(fn), newMAC[T])
}

// Prefault sets a fallback value that goes through the full validation pipeline.
func (z *ZodMAC[T]) Prefault(v string) *ZodMAC[T] {
	return wrapStringFluent(z, z.ZodString.Prefault(v), newMAC[T])
}

// PrefaultFunc sets a fallback function that goes through the full validation pipeline.
func (z *ZodMAC[T]) PrefaultFunc(fn func() string) *ZodMAC[T] {
	return wrapStringFluent(z, z.ZodString.PrefaultFunc(fn), newMAC[T])
}

// Meta attaches metadata to this schema via the global registry.
func (z *ZodMAC[T]) Meta(meta core.GlobalMeta) *ZodMAC[T] {
	return withStringWrapperMeta(z, z.ZodString, newMAC[T], meta)
}

// Describe attaches a description to this schema via the global registry.
func (z *ZodMAC[T]) Describe(description string) *ZodMAC[T] {
	return z.Meta(core.GlobalMeta{Description: description})
}

// RefineAny adds custom validation that receives the raw value as any.
func (z *ZodMAC[T]) RefineAny(fn func(any) bool, params ...any) *ZodMAC[T] {
	return wrapStringFluent(z, z.ZodString.RefineAny(fn, params...), newMAC[T])
}

// Default sets a fallback value returned when input is nil.
func (z *ZodE164[T]) Default(v string) *ZodE164[T] {
	return wrapStringFluent(z, z.ZodString.Default(v), newE164[T])
}

// DefaultFunc sets a fallback function called when input is nil.
func (z *ZodE164[T]) DefaultFunc(fn func() string) *ZodE164[T] {
	return wrapStringFluent(z, z.ZodString.DefaultFunc(fn), newE164[T])
}

// Prefault sets a fallback value that goes through the full validation pipeline.
func (z *ZodE164[T]) Prefault(v string) *ZodE164[T] {
	return wrapStringFluent(z, z.ZodString.Prefault(v), newE164[T])
}

// PrefaultFunc sets a fallback function that goes through the full validation pipeline.
func (z *ZodE164[T]) PrefaultFunc(fn func() string) *ZodE164[T] {
	return wrapStringFluent(z, z.ZodString.PrefaultFunc(fn), newE164[T])
}

// Meta attaches metadata to this schema via the global registry.
func (z *ZodE164[T]) Meta(meta core.GlobalMeta) *ZodE164[T] {
	return withStringWrapperMeta(z, z.ZodString, newE164[T], meta)
}

// Describe attaches a description to this schema via the global registry.
func (z *ZodE164[T]) Describe(description string) *ZodE164[T] {
	return z.Meta(core.GlobalMeta{Description: description})
}

// RefineAny adds custom validation that receives the raw value as any.
func (z *ZodE164[T]) RefineAny(fn func(any) bool, params ...any) *ZodE164[T] {
	return wrapStringFluent(z, z.ZodString.RefineAny(fn, params...), newE164[T])
}

// Meta attaches metadata to this schema via the global registry.
func (z *ZodIso[T]) Meta(meta core.GlobalMeta) *ZodIso[T] {
	return withStringWrapperMeta(z, z.ZodString, newIso[T], meta)
}

// Describe attaches a description to this schema via the global registry.
func (z *ZodIso[T]) Describe(description string) *ZodIso[T] {
	return z.Meta(core.GlobalMeta{Description: description})
}

// RefineAny adds custom validation that receives the raw value as any.
func (z *ZodIso[T]) RefineAny(fn func(any) bool, params ...any) *ZodIso[T] {
	return wrapStringFluent(z, z.ZodString.RefineAny(fn, params...), newIso[T])
}

// Default sets a fallback value returned when input is nil.
func (z *ZodEmoji[T]) Default(v string) *ZodEmoji[T] {
	return wrapStringFluent(z, z.ZodString.Default(v), newEmoji[T])
}

// DefaultFunc sets a fallback function called when input is nil.
func (z *ZodEmoji[T]) DefaultFunc(fn func() string) *ZodEmoji[T] {
	return wrapStringFluent(z, z.ZodString.DefaultFunc(fn), newEmoji[T])
}

// Prefault sets a fallback value that goes through the full validation pipeline.
func (z *ZodEmoji[T]) Prefault(v string) *ZodEmoji[T] {
	return wrapStringFluent(z, z.ZodString.Prefault(v), newEmoji[T])
}

// PrefaultFunc sets a fallback function that goes through the full validation pipeline.
func (z *ZodEmoji[T]) PrefaultFunc(fn func() string) *ZodEmoji[T] {
	return wrapStringFluent(z, z.ZodString.PrefaultFunc(fn), newEmoji[T])
}

// Meta attaches metadata to this schema via the global registry.
func (z *ZodEmoji[T]) Meta(meta core.GlobalMeta) *ZodEmoji[T] {
	return withStringWrapperMeta(z, z.ZodString, newEmoji[T], meta)
}

// Describe attaches a description to this schema via the global registry.
func (z *ZodEmoji[T]) Describe(description string) *ZodEmoji[T] {
	return z.Meta(core.GlobalMeta{Description: description})
}

// RefineAny adds custom validation that receives the raw value as any.
func (z *ZodEmoji[T]) RefineAny(fn func(any) bool, params ...any) *ZodEmoji[T] {
	return wrapStringFluent(z, z.ZodString.RefineAny(fn, params...), newEmoji[T])
}

// Default sets a fallback value returned when input is nil.
func (z *ZodJWT[T]) Default(v string) *ZodJWT[T] {
	return wrapStringFluent(z, z.ZodString.Default(v), newJWT[T])
}

// DefaultFunc sets a fallback function called when input is nil.
func (z *ZodJWT[T]) DefaultFunc(fn func() string) *ZodJWT[T] {
	return wrapStringFluent(z, z.ZodString.DefaultFunc(fn), newJWT[T])
}

// Prefault sets a fallback value that goes through the full validation pipeline.
func (z *ZodJWT[T]) Prefault(v string) *ZodJWT[T] {
	return wrapStringFluent(z, z.ZodString.Prefault(v), newJWT[T])
}

// PrefaultFunc sets a fallback function that goes through the full validation pipeline.
func (z *ZodJWT[T]) PrefaultFunc(fn func() string) *ZodJWT[T] {
	return wrapStringFluent(z, z.ZodString.PrefaultFunc(fn), newJWT[T])
}

// Meta attaches metadata to this schema via the global registry.
func (z *ZodJWT[T]) Meta(meta core.GlobalMeta) *ZodJWT[T] {
	return withStringWrapperMeta(z, z.ZodString, newJWT[T], meta)
}

// Describe attaches a description to this schema via the global registry.
func (z *ZodJWT[T]) Describe(description string) *ZodJWT[T] {
	return z.Meta(core.GlobalMeta{Description: description})
}

// RefineAny adds custom validation that receives the raw value as any.
func (z *ZodJWT[T]) RefineAny(fn func(any) bool, params ...any) *ZodJWT[T] {
	return wrapStringFluent(z, z.ZodString.RefineAny(fn, params...), newJWT[T])
}

// Default sets a fallback value returned when input is nil.
func (z *ZodBase64[T]) Default(v string) *ZodBase64[T] {
	return wrapStringFluent(z, z.ZodString.Default(v), newBase64[T])
}

// DefaultFunc sets a fallback function called when input is nil.
func (z *ZodBase64[T]) DefaultFunc(fn func() string) *ZodBase64[T] {
	return wrapStringFluent(z, z.ZodString.DefaultFunc(fn), newBase64[T])
}

// Prefault sets a fallback value that goes through the full validation pipeline.
func (z *ZodBase64[T]) Prefault(v string) *ZodBase64[T] {
	return wrapStringFluent(z, z.ZodString.Prefault(v), newBase64[T])
}

// PrefaultFunc sets a fallback function that goes through the full validation pipeline.
func (z *ZodBase64[T]) PrefaultFunc(fn func() string) *ZodBase64[T] {
	return wrapStringFluent(z, z.ZodString.PrefaultFunc(fn), newBase64[T])
}

// Meta attaches metadata to this schema via the global registry.
func (z *ZodBase64[T]) Meta(meta core.GlobalMeta) *ZodBase64[T] {
	return withStringWrapperMeta(z, z.ZodString, newBase64[T], meta)
}

// Describe attaches a description to this schema via the global registry.
func (z *ZodBase64[T]) Describe(description string) *ZodBase64[T] {
	return z.Meta(core.GlobalMeta{Description: description})
}

// RefineAny adds custom validation that receives the raw value as any.
func (z *ZodBase64[T]) RefineAny(fn func(any) bool, params ...any) *ZodBase64[T] {
	return wrapStringFluent(z, z.ZodString.RefineAny(fn, params...), newBase64[T])
}

// Default sets a fallback value returned when input is nil.
func (z *ZodBase64URL[T]) Default(v string) *ZodBase64URL[T] {
	return wrapStringFluent(z, z.ZodString.Default(v), newBase64URL[T])
}

// DefaultFunc sets a fallback function called when input is nil.
func (z *ZodBase64URL[T]) DefaultFunc(fn func() string) *ZodBase64URL[T] {
	return wrapStringFluent(z, z.ZodString.DefaultFunc(fn), newBase64URL[T])
}

// Prefault sets a fallback value that goes through the full validation pipeline.
func (z *ZodBase64URL[T]) Prefault(v string) *ZodBase64URL[T] {
	return wrapStringFluent(z, z.ZodString.Prefault(v), newBase64URL[T])
}

// PrefaultFunc sets a fallback function that goes through the full validation pipeline.
func (z *ZodBase64URL[T]) PrefaultFunc(fn func() string) *ZodBase64URL[T] {
	return wrapStringFluent(z, z.ZodString.PrefaultFunc(fn), newBase64URL[T])
}

// Meta attaches metadata to this schema via the global registry.
func (z *ZodBase64URL[T]) Meta(meta core.GlobalMeta) *ZodBase64URL[T] {
	return withStringWrapperMeta(z, z.ZodString, newBase64URL[T], meta)
}

// Describe attaches a description to this schema via the global registry.
func (z *ZodBase64URL[T]) Describe(description string) *ZodBase64URL[T] {
	return z.Meta(core.GlobalMeta{Description: description})
}

// RefineAny adds custom validation that receives the raw value as any.
func (z *ZodBase64URL[T]) RefineAny(fn func(any) bool, params ...any) *ZodBase64URL[T] {
	return wrapStringFluent(z, z.ZodString.RefineAny(fn, params...), newBase64URL[T])
}

// Default sets a fallback value returned when input is nil.
func (z *ZodHex[T]) Default(v string) *ZodHex[T] {
	return wrapStringFluent(z, z.ZodString.Default(v), newHex[T])
}

// DefaultFunc sets a fallback function called when input is nil.
func (z *ZodHex[T]) DefaultFunc(fn func() string) *ZodHex[T] {
	return wrapStringFluent(z, z.ZodString.DefaultFunc(fn), newHex[T])
}

// Prefault sets a fallback value that goes through the full validation pipeline.
func (z *ZodHex[T]) Prefault(v string) *ZodHex[T] {
	return wrapStringFluent(z, z.ZodString.Prefault(v), newHex[T])
}

// PrefaultFunc sets a fallback function that goes through the full validation pipeline.
func (z *ZodHex[T]) PrefaultFunc(fn func() string) *ZodHex[T] {
	return wrapStringFluent(z, z.ZodString.PrefaultFunc(fn), newHex[T])
}

// Meta attaches metadata to this schema via the global registry.
func (z *ZodHex[T]) Meta(meta core.GlobalMeta) *ZodHex[T] {
	return withStringWrapperMeta(z, z.ZodString, newHex[T], meta)
}

// Describe attaches a description to this schema via the global registry.
func (z *ZodHex[T]) Describe(description string) *ZodHex[T] {
	return z.Meta(core.GlobalMeta{Description: description})
}

// RefineAny adds custom validation that receives the raw value as any.
func (z *ZodHex[T]) RefineAny(fn func(any) bool, params ...any) *ZodHex[T] {
	return wrapStringFluent(z, z.ZodString.RefineAny(fn, params...), newHex[T])
}
