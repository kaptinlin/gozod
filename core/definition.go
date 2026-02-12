package core

//
// This file groups all primitives related to the **definition** of a schema.
// These types are used during the construction phase to configure a schema's
// behavior, add validation checks, or define its shape (for objects/structs).

// ObjectSchema defines the shape of an object for validation.
// It maps field names to their corresponding validation schemas.
type ObjectSchema = map[string]ZodSchema

// StructSchema defines the shape of a struct for validation.
// It maps field names to their corresponding validation schemas.
type StructSchema = map[string]ZodSchema

// ZodTypeDef represents the base configuration for creating any schema type.
// It's a blueprint used by type constructors.
type ZodTypeDef struct {
	Type     ZodTypeCode  // Type name using type-safe constants
	Coerce   bool         // Enable coercion
	Required bool         // Field is required
	Error    *ZodErrorMap // Custom error handler
	Checks   []ZodCheck   // Validation checks
}

// SchemaParams contains optional parameters for schema creation and check attachment.
// It provides a flexible way to configure schemas with error messages, descriptions,
// and other extensible options.
type SchemaParams struct {
	Error       any            // Error message or error map (string or ZodErrorMap)
	Description string         // Human-readable description
	Abort       bool           // Abort on first validation failure
	Params      map[string]any // Additional extensible parameters
}

// WithError creates a SchemaParams with a custom error message.
func WithError(message string) SchemaParams {
	return SchemaParams{Error: message}
}

// WithDescription creates a SchemaParams with a description.
func WithDescription(description string) SchemaParams {
	return SchemaParams{Description: description}
}

// WithAbort creates a SchemaParams with abort-on-error enabled.
func WithAbort() SchemaParams {
	return SchemaParams{Abort: true}
}
