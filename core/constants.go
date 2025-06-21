package core

// =============================================================================
// LIBRARY VERSION
// =============================================================================

// Version is the current version of the library
// Used for compatibility checking and feature detection
var Version = "0.1.0"

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
)

// =============================================================================
// ZOD TYPE CONSTANTS
// =============================================================================

// ZodType constants define the type identifiers for all schema types
// These are used internally to identify and categorize schema types
const (
	// Primitive types
	ZodTypeString  = "string"  // String validation schema
	ZodTypeNumber  = "number"  // Generic number validation
	ZodTypeNaN     = "nan"     // NaN value validation
	ZodTypeInteger = "integer" // Integer validation
	ZodTypeBigInt  = "bigint"  // Big integer validation
	ZodTypeBool    = "bool"    // Boolean validation
	ZodTypeDate    = "date"    // Date validation
	ZodTypeNil     = "nil"     // Nil/null validation

	// Special types
	ZodTypeAny     = "any"     // Accept any value
	ZodTypeUnknown = "unknown" // Unknown type (safer any)
	ZodTypeNever   = "never"   // Never accepts values
	ZodTypeVoid    = "void"    // Void return type

	// Collection types
	ZodTypeArray  = "array"  // Fixed-length array
	ZodTypeSlice  = "slice"  // Dynamic array/slice
	ZodTypeObject = "object" // Object with known shape
	ZodTypeStruct = "struct" // Go struct validation
	ZodTypeRecord = "record" // Key-value record
	ZodTypeMap    = "map"    // Go map validation
	ZodTypeSet    = "set"    // Set validation
	ZodTypeTuple  = "tuple"  // Fixed-length tuple

	// Composite types
	ZodTypeUnion         = "union"               // Union of multiple types
	ZodTypeDiscriminated = "discriminated_union" // Discriminated union
	ZodTypeIntersection  = "intersection"        // Intersection of types

	// Special string types
	ZodTypeStringBool = "stringbool" // String representation of boolean

	// Function and lazy types
	ZodTypeFunction = "function" // Function validation
	ZodTypeLazy     = "lazy"     // Lazy evaluation schema

	// Value types
	ZodTypeLiteral = "literal" // Literal value validation
	ZodTypeEnum    = "enum"    // Enumeration validation

	// Modifier types
	ZodTypeOptional = "optional" // Optional field modifier
	ZodTypeNilable  = "nilable"  // Nilable field modifier
	ZodTypeDefault  = "default"  // Default value wrapper
	ZodTypePrefault = "prefault" // Fallback value wrapper

	// Processing types
	ZodTypePipeline  = "pipeline"  // Processing pipeline
	ZodTypeTransform = "transform" // Value transformation
	ZodTypePipe      = "pipe"      // Schema piping
	ZodTypeCustom    = "custom"    // Custom validation
	ZodTypeCheck     = "check"     // Validation check
	ZodTypeRefine    = "refine"    // Refinement validation

	// Network and format types
	ZodTypeIPv4   = "ipv4"   // IPv4 address validation
	ZodTypeIPv6   = "ipv6"   // IPv6 address validation
	ZodTypeCIDRv4 = "cidrv4" // IPv4 CIDR validation
	ZodTypeCIDRv6 = "cidrv6" // IPv6 CIDR validation
	ZodTypeEmail  = "email"  // Email address validation
	ZodTypeURL    = "url"    // URL validation

	// File and binary types
	ZodTypeFile = "file" // File validation

	// Numeric subtypes
	ZodTypeFloat32    = "float32"    // 32-bit float
	ZodTypeFloat64    = "float64"    // 64-bit float
	ZodTypeInt8       = "int8"       // 8-bit signed integer
	ZodTypeInt16      = "int16"      // 16-bit signed integer
	ZodTypeInt32      = "int32"      // 32-bit signed integer
	ZodTypeInt64      = "int64"      // 64-bit signed integer
	ZodTypeUint8      = "uint8"      // 8-bit unsigned integer
	ZodTypeUint16     = "uint16"     // 16-bit unsigned integer
	ZodTypeUint32     = "uint32"     // 32-bit unsigned integer
	ZodTypeUint64     = "uint64"     // 64-bit unsigned integer
	ZodTypeUintptr    = "uintptr"    // Pointer-sized unsigned integer
	ZodTypeComplex64  = "complex64"  // 64-bit complex number
	ZodTypeComplex128 = "complex128" // 128-bit complex number
)

// =============================================================================
// PARSED TYPE CONSTANTS
// =============================================================================

// ParsedType represents the type of parsed data values
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
	ParsedTypeMap      ParsedType = "map"      // Map data type
	ParsedTypeNaN      ParsedType = "nan"      // Not-a-Number data type
	ParsedTypeNil      ParsedType = "nil"      // Nil/null data type
	ParsedTypeComplex  ParsedType = "complex"  // Complex number data type
	ParsedTypeStruct   ParsedType = "struct"   // Go struct data type
	ParsedTypeEnum     ParsedType = "enum"     // Enumeration data type
	ParsedTypeUnknown  ParsedType = "unknown"  // Unknown data type
)

// =============================================================================
// NUMERIC FORMAT RANGES
// =============================================================================

// NUMBER_FORMAT_RANGES defines numeric format validation ranges
// Used to validate that numbers fit within specific format constraints
var NUMBER_FORMAT_RANGES = map[string][2]float64{
	"safeint": {-9007199254740991, 9007199254740991},             // JavaScript safe integer range
	"int32":   {-2147483648, 2147483647},                         // 32-bit signed integer range
	"uint32":  {0, 4294967295},                                   // 32-bit unsigned integer range
	"float32": {-3.4028234663852886e38, 3.4028234663852886e38},   // 32-bit float range
	"float64": {-1.7976931348623157e308, 1.7976931348623157e308}, // 64-bit float range
}

// BIGINT_FORMAT_RANGES defines big integer format validation ranges
// Used to validate that big integers fit within specific format constraints
var BIGINT_FORMAT_RANGES = map[string][2]int64{
	"int64":  {-9223372036854775808, 9223372036854775807}, // 64-bit signed integer range
	"uint64": {0, 9223372036854775807},                    // 64-bit unsigned integer range (limited to int64 max)
}
