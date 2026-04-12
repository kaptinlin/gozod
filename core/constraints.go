package core

// Describable defines the fluent metadata contract.
//
// This is intentionally separate from ZodSchema/ZodType because metadata
// chaining is not required by dynamic runtime code paths. The self-referential
// constraint ensures Describe and Meta return the concrete schema type.
//
// Go 1.26 self-referential generic constraint.
type Describable[S Describable[S]] interface {
	Describe(string) S
	Meta(GlobalMeta) S
	Internals() *ZodTypeInternals
}

// Refineable defines the fluent refinement contract.
//
// This stays separate from the minimal runtime interfaces so callers can write
// helpers against refinement-capable schemas without implying every runtime
// dependency must also know about fluent validation APIs.
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
