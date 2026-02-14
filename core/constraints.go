package core

// Describable defines schema types that support metadata chaining.
// The self-referential constraint ensures Describe and Meta return
// the concrete schema type for fluent method chaining.
//
// Go 1.26 self-referential generic constraint.
type Describable[S Describable[S]] interface {
	Describe(string) S
	Meta(GlobalMeta) S
	Internals() *ZodTypeInternals
}

// Refineable defines schema types that support custom validation.
// The self-referential constraint ensures RefineAny returns the
// concrete schema type for chaining additional validations.
//
// Go 1.26 self-referential generic constraint.
type Refineable[S Refineable[S]] interface {
	RefineAny(func(any) bool, ...any) S
	Internals() *ZodTypeInternals
}

// DescribeSchema applies a description to any Describable schema,
// returning the same concrete type.
func DescribeSchema[S Describable[S]](schema S, desc string) S {
	return schema.Describe(desc)
}

// MetaSchema applies metadata to any Describable schema,
// returning the same concrete type.
func MetaSchema[S Describable[S]](schema S, meta GlobalMeta) S {
	return schema.Meta(meta)
}

// ApplyRefinements applies multiple validation rules to any Refineable schema.
func ApplyRefinements[S Refineable[S]](schema S, rules []func(any) bool) S {
	for _, rule := range rules {
		schema = schema.RefineAny(rule)
	}
	return schema
}
