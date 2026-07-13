package gozod

import (
	lib "github.com/kaptinlin/jsonschema"

	"github.com/kaptinlin/gozod/jsonschema"
)

type JSONSchemaOptions = jsonschema.Options
type OverrideContext = jsonschema.OverrideContext
type FromJSONSchemaOptions = jsonschema.FromJSONSchemaOptions
type JSONSchemaImportError = jsonschema.ImportError
type JSONSchemaImportLossError = jsonschema.ImportLossError
type JSONSchemaUnrepresentableMode = jsonschema.UnrepresentableMode
type JSONSchemaCyclesMode = jsonschema.CyclesMode
type JSONSchemaReusedMode = jsonschema.ReusedMode
type JSONSchemaTargetMode = jsonschema.TargetMode
type JSONSchemaIOMode = jsonschema.IOMode

// ToJSONSchema converts a GoZod schema or registry into JSON Schema.
func ToJSONSchema(input any, opts ...JSONSchemaOptions) (*lib.Schema, error) {
	return jsonschema.ToJSONSchema(input, opts...)
}

// FromJSONSchema converts JSON Schema into a GoZod schema.
func FromJSONSchema(schema *lib.Schema, opts ...FromJSONSchemaOptions) (ZodSchema, error) {
	return jsonschema.FromJSONSchema(schema, opts...)
}

// FromJSONSchemaLossy converts a JSON Schema and reports omitted validation semantics.
func FromJSONSchemaLossy(
	schema *lib.Schema,
	opts ...FromJSONSchemaOptions,
) (ZodSchema, []JSONSchemaImportLossError, error) {
	return jsonschema.FromJSONSchemaLossy(schema, opts...)
}

const (
	JSONSchemaUnrepresentableThrow = jsonschema.UnrepresentableThrow
	JSONSchemaUnrepresentableAny   = jsonschema.UnrepresentableAny
	JSONSchemaCyclesRef            = jsonschema.CyclesRef
	JSONSchemaCyclesThrow          = jsonschema.CyclesThrow
	JSONSchemaReusedInline         = jsonschema.ReusedInline
	JSONSchemaReusedRef            = jsonschema.ReusedRef
	JSONSchemaTargetDraft202012    = jsonschema.TargetDraft202012
	JSONSchemaIOOutput             = jsonschema.IOOutput
	JSONSchemaIOInput              = jsonschema.IOInput
)

var (
	ErrUnsupportedInputType          = jsonschema.ErrUnsupportedInputType
	ErrCircularReference             = jsonschema.ErrCircularReference
	ErrUnrepresentableType           = jsonschema.ErrUnrepresentableType
	ErrSchemaNotObjectOrStruct       = jsonschema.ErrSchemaNotObjectOrStruct
	ErrSliceElementNotSchema         = jsonschema.ErrSliceElementNotSchema
	ErrArrayItemNotSchema            = jsonschema.ErrArrayItemNotSchema
	ErrUnhandledArrayLike            = jsonschema.ErrUnhandledArrayLike
	ErrUnionInvalid                  = jsonschema.ErrUnionInvalid
	ErrUnionNoMembers                = jsonschema.ErrUnionNoMembers
	ErrIntersectionInvalid           = jsonschema.ErrIntersectionInvalid
	ErrInvalidEnumSchema             = jsonschema.ErrInvalidEnumSchema
	ErrEnumExtractValues             = jsonschema.ErrEnumExtractValues
	ErrLiteralNoValuesMethod         = jsonschema.ErrLiteralNoValuesMethod
	ErrLiteralUnexpectedReturnValues = jsonschema.ErrLiteralUnexpectedReturnValues
	ErrExpectedDiscriminatedUnion    = jsonschema.ErrExpectedDiscriminatedUnion
	ErrExpectedRecord                = jsonschema.ErrExpectedRecord
	ErrRecordValueNotSchema          = jsonschema.ErrRecordValueNotSchema
	ErrInvalidRegistrySchemaID       = jsonschema.ErrInvalidRegistrySchemaID
	ErrMapNoMethods                  = jsonschema.ErrMapNoMethods
	ErrMapKeyNotSchema               = jsonschema.ErrMapKeyNotSchema
	ErrMapValueNotSchema             = jsonschema.ErrMapValueNotSchema
	ErrInvalidJSONSchemaOption       = jsonschema.ErrInvalidJSONSchemaOption
	ErrUnsupportedJSONSchemaTarget   = jsonschema.ErrUnsupportedJSONSchemaTarget
	ErrUnsupportedJSONSchemaType     = jsonschema.ErrUnsupportedJSONSchemaType
	ErrUnsupportedJSONSchemaKeyword  = jsonschema.ErrUnsupportedJSONSchemaKeyword
	ErrInvalidJSONSchema             = jsonschema.ErrInvalidJSONSchema
	ErrJSONSchemaCircularRef         = jsonschema.ErrJSONSchemaCircularRef
	ErrJSONSchemaPatternCompile      = jsonschema.ErrJSONSchemaPatternCompile
	ErrJSONSchemaIfThenElse          = jsonschema.ErrJSONSchemaIfThenElse
	ErrJSONSchemaPatternProperties   = jsonschema.ErrJSONSchemaPatternProperties
	ErrJSONSchemaDynamicRef          = jsonschema.ErrJSONSchemaDynamicRef
	ErrJSONSchemaUnevaluatedProps    = jsonschema.ErrJSONSchemaUnevaluatedProps
	ErrJSONSchemaUnevaluatedItems    = jsonschema.ErrJSONSchemaUnevaluatedItems
	ErrJSONSchemaDependentSchemas    = jsonschema.ErrJSONSchemaDependentSchemas
	ErrJSONSchemaPropertyNames       = jsonschema.ErrJSONSchemaPropertyNames
	ErrJSONSchemaContains            = jsonschema.ErrJSONSchemaContains
)
