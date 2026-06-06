// Package tagparser parses struct-tag validation rules into a
// reusable representation shared by GoZod runtime code and codegen.
//
// Callers should prefer [FieldInfo] helper methods such as
// [FieldInfo.ValidationRules], [FieldInfo.NeedsGeneratedOptional], and
// [FieldInfo.RequiredImports] instead of re-deriving schema-generation
// semantics from raw fields like Rules, Required, or Optional.
//
// Use [TagParser.ParseStructTags] for compatibility-oriented callers
// that prefer empty results on non-struct input, or
// [TagParser.ParseStructTagsStrict] when non-struct input should be
// treated as a contract error.
//
// A parser is configured with two tag names: the rule tag (default "gozod")
// and the format tag that supplies field names (default "json"). Use
// [NewWithTags] to select both, or [FieldName] to resolve a single field's
// name from a chosen format tag such as "yaml" or "toml".
package tagparser
