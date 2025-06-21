package core

// =============================================================================
// TYPE ALIASES FOR CONVENIENCE
// =============================================================================

// Schema is a convenient alias for ZodType[any, any]
// Provides a shorter name for the most common schema type
type Schema = ZodType[any, any]

// ObjectSchema defines object shape for structured validation
// Maps field names to their corresponding validation schemas
type ObjectSchema = map[string]ZodType[any, any]

// StructSchema defines struct shape for validation
// Maps field names to their corresponding validation schemas
type StructSchema = map[string]ZodType[any, any]
