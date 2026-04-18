package gozod

import "github.com/kaptinlin/gozod/jsonschema"

type JSONSchemaOptions = jsonschema.Options
type OverrideContext = jsonschema.OverrideContext
type FromJSONSchemaOptions = jsonschema.FromJSONSchemaOptions

var (
	ToJSONSchema                     = jsonschema.ToJSONSchema
	FromJSONSchema                   = jsonschema.FromJSONSchema
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
	ErrMapNoMethods                  = jsonschema.ErrMapNoMethods
	ErrMapKeyNotSchema               = jsonschema.ErrMapKeyNotSchema
	ErrMapValueNotSchema             = jsonschema.ErrMapValueNotSchema
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
