package core

// IssueCode represents validation issue types.
// These codes categorize different types of validation failures.
type IssueCode string

const (
	// Type validation issues
	InvalidType    IssueCode = "invalid_type"
	InvalidValue   IssueCode = "invalid_value"
	InvalidFormat  IssueCode = "invalid_format"
	InvalidUnion   IssueCode = "invalid_union"
	InvalidKey     IssueCode = "invalid_key"
	InvalidElement IssueCode = "invalid_element"

	// Range validation issues
	TooBig        IssueCode = "too_big"
	TooSmall      IssueCode = "too_small"
	NotMultipleOf IssueCode = "not_multiple_of"

	// Structure validation issues
	UnrecognizedKeys IssueCode = "unrecognized_keys"

	// Custom validation issues
	Custom IssueCode = "custom"

	// Schema validation issues
	InvalidSchema IssueCode = "invalid_schema"

	// Discriminator validation issues
	InvalidDiscriminator IssueCode = "invalid_discriminator"

	// Intersection validation issues
	IncompatibleTypes IssueCode = "incompatible_types"

	// New validation issues
	MissingRequired IssueCode = "missing_required"
	TypeConversion  IssueCode = "type_conversion"
	NilPointer      IssueCode = "nil_pointer"
)

// ZodTypeCode represents a type-safe wrapper for schema type identifiers.
// This provides compile-time type safety and better IDE support.
type ZodTypeCode string

// ZodType constants define the type identifiers for all schema types.
// These are used internally to identify and categorize schema types.
const (
	// Primitive types
	ZodTypeString  ZodTypeCode = "string"
	ZodTypeNumber  ZodTypeCode = "number"
	ZodTypeNaN     ZodTypeCode = "nan"
	ZodTypeInteger ZodTypeCode = "integer"
	ZodTypeBigInt  ZodTypeCode = "bigint"
	ZodTypeBool    ZodTypeCode = "bool"
	ZodTypeDate    ZodTypeCode = "date"
	ZodTypeNil     ZodTypeCode = "nil"

	// Special types
	ZodTypeAny     ZodTypeCode = "any"
	ZodTypeUnknown ZodTypeCode = "unknown"
	ZodTypeNever   ZodTypeCode = "never"

	// Collection types
	ZodTypeArray  ZodTypeCode = "array"
	ZodTypeSlice  ZodTypeCode = "slice"
	ZodTypeTuple  ZodTypeCode = "tuple"
	ZodTypeObject ZodTypeCode = "object"
	ZodTypeStruct ZodTypeCode = "struct"
	ZodTypeRecord ZodTypeCode = "record"
	ZodTypeMap    ZodTypeCode = "map"
	ZodTypeSet    ZodTypeCode = "set"

	// Composite types
	ZodTypeUnion         ZodTypeCode = "union"
	ZodTypeXor           ZodTypeCode = "xor"
	ZodTypeDiscriminated ZodTypeCode = "discriminated_union"
	ZodTypeIntersection  ZodTypeCode = "intersection"

	// Special string types
	ZodTypeStringBool ZodTypeCode = "stringbool"

	// Function and lazy types
	ZodTypeFunction ZodTypeCode = "function"
	ZodTypeLazy     ZodTypeCode = "lazy"

	// Value types
	ZodTypeLiteral ZodTypeCode = "literal"
	ZodTypeEnum    ZodTypeCode = "enum"

	// Modifier types
	ZodTypeOptional ZodTypeCode = "optional"
	ZodTypeNilable  ZodTypeCode = "nilable"
	ZodTypeDefault  ZodTypeCode = "default"
	ZodTypePrefault ZodTypeCode = "prefault"

	// Processing types
	ZodTypePipeline  ZodTypeCode = "pipeline"
	ZodTypeTransform ZodTypeCode = "transform"
	ZodTypePipe      ZodTypeCode = "pipe"
	ZodTypeCustom    ZodTypeCode = "custom"
	ZodTypeCheck     ZodTypeCode = "check"
	ZodTypeRefine    ZodTypeCode = "refine"

	// Network and format types
	ZodTypeIPv4     ZodTypeCode = "ipv4"
	ZodTypeIPv6     ZodTypeCode = "ipv6"
	ZodTypeCIDRv4   ZodTypeCode = "cidrv4"
	ZodTypeCIDRv6   ZodTypeCode = "cidrv6"
	ZodTypeEmail    ZodTypeCode = "email"
	ZodTypeURL      ZodTypeCode = "url"
	ZodTypeHostname ZodTypeCode = "hostname"
	ZodTypeMAC      ZodTypeCode = "mac"
	ZodTypeE164     ZodTypeCode = "e164"

	// Time types
	ZodTypeTime ZodTypeCode = "time"

	// ISO 8601 format validation types
	ZodTypeIso         ZodTypeCode = "iso"
	ZodTypeISODateTime ZodTypeCode = "iso_datetime"
	ZodTypeISODate     ZodTypeCode = "iso_date"
	ZodTypeISOTime     ZodTypeCode = "iso_time"
	ZodTypeISODuration ZodTypeCode = "iso_duration"

	// File and binary types
	ZodTypeFile ZodTypeCode = "file"

	// Numeric subtypes
	ZodTypeFloat32     ZodTypeCode = "float32"
	ZodTypeFloat64     ZodTypeCode = "float64"
	ZodTypeFloat       ZodTypeCode = "float"
	ZodTypeInt         ZodTypeCode = "int"
	ZodTypeInt8        ZodTypeCode = "int8"
	ZodTypeInt16       ZodTypeCode = "int16"
	ZodTypeInt32       ZodTypeCode = "int32"
	ZodTypeInt64       ZodTypeCode = "int64"
	ZodTypeUint        ZodTypeCode = "uint"
	ZodTypeUint8       ZodTypeCode = "uint8"
	ZodTypeUint16      ZodTypeCode = "uint16"
	ZodTypeUint32      ZodTypeCode = "uint32"
	ZodTypeUint64      ZodTypeCode = "uint64"
	ZodTypeUintptr     ZodTypeCode = "uintptr"
	ZodTypeComplex64   ZodTypeCode = "complex64"
	ZodTypeComplex128  ZodTypeCode = "complex128"
	ZodTypeNonOptional ZodTypeCode = "nonoptional"
)

// ParsedType represents the type of parsed data values at runtime.
// This corresponds to Zod v4's ParsedTypes.
// See: .reference/zod/packages/zod/src/v4/core/util.ts:66-82
// These are used during runtime type detection and validation.
type ParsedType string

const (
	ParsedTypeString   ParsedType = "string"
	ParsedTypeNumber   ParsedType = "number"
	ParsedTypeBigint   ParsedType = "bigint"
	ParsedTypeBool     ParsedType = "bool"
	ParsedTypeFloat    ParsedType = "float"
	ParsedTypeObject   ParsedType = "object"
	ParsedTypeFunction ParsedType = "function"
	ParsedTypeFile     ParsedType = "file"
	ParsedTypeDate     ParsedType = "date"
	ParsedTypeArray    ParsedType = "array"
	ParsedTypeSlice    ParsedType = "slice"
	ParsedTypeTuple    ParsedType = "tuple"
	ParsedTypeMap      ParsedType = "map"
	ParsedTypeSet      ParsedType = "set"
	ParsedTypeNaN      ParsedType = "nan"
	ParsedTypeNil      ParsedType = "nil"
	ParsedTypeComplex  ParsedType = "complex"
	ParsedTypeStruct   ParsedType = "struct"
	ParsedTypeEnum     ParsedType = "enum"
	ParsedTypeUnknown  ParsedType = "unknown"
)
