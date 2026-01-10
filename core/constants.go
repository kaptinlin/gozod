package core

// =============================================================================
// ISSUE CODE CONSTANTS
// =============================================================================

// IssueCode represents validation issue types
// These codes categorize different types of validation failures
type IssueCode string

const (
	// Type validation issues
	InvalidType    IssueCode = "invalid_type"    // Input type doesn't match expected type
	InvalidValue   IssueCode = "invalid_value"   // Input value not in allowed set
	InvalidFormat  IssueCode = "invalid_format"  // String doesn't match expected format
	InvalidUnion   IssueCode = "invalid_union"   // None of union alternatives match
	InvalidKey     IssueCode = "invalid_key"     // Object key validation failed
	InvalidElement IssueCode = "invalid_element" // Array element validation failed

	// Range validation issues
	TooBig        IssueCode = "too_big"         // Value exceeds maximum limit
	TooSmall      IssueCode = "too_small"       // Value below minimum limit
	NotMultipleOf IssueCode = "not_multiple_of" // Value not multiple of divisor

	// Structure validation issues
	UnrecognizedKeys IssueCode = "unrecognized_keys" // Object has unknown keys

	// Custom validation issues
	Custom IssueCode = "custom" // Custom validation failure

	// Schema validation issues
	InvalidSchema IssueCode = "invalid_schema" // Schema definition is invalid

	// Discriminator validation issues
	InvalidDiscriminator IssueCode = "invalid_discriminator" // Discriminator field missing/invalid

	// Intersection validation issues
	IncompatibleTypes IssueCode = "incompatible_types" // Types cannot be merged (intersection)

	// New validation issues
	MissingRequired IssueCode = "missing_required" // Required field is missing
	TypeConversion  IssueCode = "type_conversion"  // Type conversion failed
	NilPointer      IssueCode = "nil_pointer"      // Nil pointer encountered

)

// =============================================================================
// ZOD TYPE CONSTANTS
// =============================================================================

// ZodTypeCode represents a type-safe wrapper for schema type identifiers
// This provides compile-time type safety and better IDE support
type ZodTypeCode string

// ZodType constants define the type identifiers for all schema types
// These are used internally to identify and categorize schema types
const (
	// Primitive types
	ZodTypeString  ZodTypeCode = "string"  // String validation schema
	ZodTypeNumber  ZodTypeCode = "number"  // Generic number validation
	ZodTypeNaN     ZodTypeCode = "nan"     // NaN value validation
	ZodTypeInteger ZodTypeCode = "integer" // Integer validation
	ZodTypeBigInt  ZodTypeCode = "bigint"  // Big integer validation
	ZodTypeBool    ZodTypeCode = "bool"    // Boolean validation
	ZodTypeDate    ZodTypeCode = "date"    // Date validation
	ZodTypeNil     ZodTypeCode = "nil"     // Nil/null validation

	// Special types
	ZodTypeAny     ZodTypeCode = "any"     // Accept any value
	ZodTypeUnknown ZodTypeCode = "unknown" // Unknown type (safer any)
	ZodTypeNever   ZodTypeCode = "never"   // Never accepts value

	// Collection types
	ZodTypeArray  ZodTypeCode = "array"  // Fixed-length array
	ZodTypeSlice  ZodTypeCode = "slice"  // Dynamic array/slice
	ZodTypeTuple  ZodTypeCode = "tuple"  // Fixed-length tuple with heterogeneous types
	ZodTypeObject ZodTypeCode = "object" // Object with known shape
	ZodTypeStruct ZodTypeCode = "struct" // Go struct validation
	ZodTypeRecord ZodTypeCode = "record" // Key-value record
	ZodTypeMap    ZodTypeCode = "map"    // Go map validation

	// Composite types
	ZodTypeUnion         ZodTypeCode = "union"               // Union of multiple types (anyOf)
	ZodTypeXor           ZodTypeCode = "xor"                 // Exclusive union (oneOf) - exactly one must match
	ZodTypeDiscriminated ZodTypeCode = "discriminated_union" // Discriminated union
	ZodTypeIntersection  ZodTypeCode = "intersection"        // Intersection of types

	// Special string types
	ZodTypeStringBool ZodTypeCode = "stringbool" // String representation of boolean

	// Function and lazy types
	ZodTypeFunction ZodTypeCode = "function" // Function validation
	ZodTypeLazy     ZodTypeCode = "lazy"     // Lazy evaluation schema

	// Value types
	ZodTypeLiteral ZodTypeCode = "literal" // Literal value validation
	ZodTypeEnum    ZodTypeCode = "enum"    // Enumeration validation

	// Modifier types
	ZodTypeOptional ZodTypeCode = "optional" // Optional field modifier
	ZodTypeNilable  ZodTypeCode = "nilable"  // Nilable field modifier
	ZodTypeDefault  ZodTypeCode = "default"  // Default value wrapper
	ZodTypePrefault ZodTypeCode = "prefault" // Fallback value wrapper

	// Processing types
	ZodTypePipeline  ZodTypeCode = "pipeline"  // Processing pipeline
	ZodTypeTransform ZodTypeCode = "transform" // Value transformation
	ZodTypePipe      ZodTypeCode = "pipe"      // Schema piping
	ZodTypeCustom    ZodTypeCode = "custom"    // Custom validation
	ZodTypeCheck     ZodTypeCode = "check"     // Validation check
	ZodTypeRefine    ZodTypeCode = "refine"    // Refinement validation

	// Network and format types
	ZodTypeIPv4     ZodTypeCode = "ipv4"     // IPv4 address validation
	ZodTypeIPv6     ZodTypeCode = "ipv6"     // IPv6 address validation
	ZodTypeCIDRv4   ZodTypeCode = "cidrv4"   // IPv4 CIDR validation
	ZodTypeCIDRv6   ZodTypeCode = "cidrv6"   // IPv6 CIDR validation
	ZodTypeEmail    ZodTypeCode = "email"    // Email address validation
	ZodTypeURL      ZodTypeCode = "url"      // URL validation
	ZodTypeHostname ZodTypeCode = "hostname" // DNS hostname validation
	ZodTypeMAC      ZodTypeCode = "mac"      // MAC address validation
	ZodTypeE164     ZodTypeCode = "e164"     // E.164 phone number validation

	// Time types
	ZodTypeTime ZodTypeCode = "time" // Go time.Time validation

	// ISO 8601 format validation types
	ZodTypeIso         ZodTypeCode = "iso"          // ISO 8601 format validation
	ZodTypeISODateTime ZodTypeCode = "iso_datetime" // ISO 8601 datetime validation
	ZodTypeISODate     ZodTypeCode = "iso_date"     // ISO 8601 date validation
	ZodTypeISOTime     ZodTypeCode = "iso_time"     // ISO 8601 time validation
	ZodTypeISODuration ZodTypeCode = "iso_duration" // ISO 8601 duration validation

	// File and binary types
	ZodTypeFile ZodTypeCode = "file" // File validation

	// Numeric subtypes
	ZodTypeFloat32     ZodTypeCode = "float32"     // 32-bit float
	ZodTypeFloat64     ZodTypeCode = "float64"     // 64-bit float
	ZodTypeFloat       ZodTypeCode = "float"       // Flexible float type (accepts float32, float64)
	ZodTypeInt         ZodTypeCode = "int"         // Flexible integer type (accepts all integer types)
	ZodTypeInt8        ZodTypeCode = "int8"        // 8-bit signed integer
	ZodTypeInt16       ZodTypeCode = "int16"       // 16-bit signed integer
	ZodTypeInt32       ZodTypeCode = "int32"       // 32-bit signed integer
	ZodTypeInt64       ZodTypeCode = "int64"       // 64-bit signed integer
	ZodTypeUint        ZodTypeCode = "uint"        // Platform-dependent unsigned integer
	ZodTypeUint8       ZodTypeCode = "uint8"       // 8-bit unsigned integer
	ZodTypeUint16      ZodTypeCode = "uint16"      // 16-bit unsigned integer
	ZodTypeUint32      ZodTypeCode = "uint32"      // 32-bit unsigned integer
	ZodTypeUint64      ZodTypeCode = "uint64"      // 64-bit unsigned integer
	ZodTypeUintptr     ZodTypeCode = "uintptr"     // Pointer-sized unsigned integer
	ZodTypeComplex64   ZodTypeCode = "complex64"   // 64-bit complex number
	ZodTypeComplex128  ZodTypeCode = "complex128"  // 128-bit complex number
	ZodTypeNonOptional ZodTypeCode = "nonoptional" // Special type identifier for non-optional fields
)

// =============================================================================
// PARSED TYPE CONSTANTS
// =============================================================================

// ParsedType represents the type of parsed data values at runtime
// This corresponds to Zod v4's ParsedTypes
// See: .reference/zod/packages/zod/src/v4/core/util.ts:66-82
// These are used during runtime type detection and validation
type ParsedType string

const (
	ParsedTypeString   ParsedType = "string"   // String data type
	ParsedTypeNumber   ParsedType = "number"   // Numeric data type (integers)
	ParsedTypeBigint   ParsedType = "bigint"   // Big integer data type
	ParsedTypeBool     ParsedType = "bool"     // Boolean data type
	ParsedTypeFloat    ParsedType = "float"    // Floating-point data type
	ParsedTypeObject   ParsedType = "object"   // Object/struct data type
	ParsedTypeFunction ParsedType = "function" // Function data type
	ParsedTypeFile     ParsedType = "file"     // File data type
	ParsedTypeDate     ParsedType = "date"     // Date/time data type
	ParsedTypeArray    ParsedType = "array"    // Fixed-size array data type
	ParsedTypeSlice    ParsedType = "slice"    // Dynamic slice data type
	ParsedTypeTuple    ParsedType = "tuple"    // Fixed-length tuple data type
	ParsedTypeMap      ParsedType = "map"      // Map data type
	ParsedTypeNaN      ParsedType = "nan"      // Not-a-Number data type
	ParsedTypeNil      ParsedType = "nil"      // Nil/null data type
	ParsedTypeComplex  ParsedType = "complex"  // Complex number data type
	ParsedTypeStruct   ParsedType = "struct"   // Go struct data type
	ParsedTypeEnum     ParsedType = "enum"     // Enumeration data type
	ParsedTypeUnknown  ParsedType = "unknown"  // Unknown data type
)
