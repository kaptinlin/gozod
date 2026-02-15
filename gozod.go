// Package gozod provides a TypeScript Zod v4-inspired validation library for Go,
// offering strongly-typed, zero-dependency data validation with intelligent type inference.
package gozod

import (
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/internal/utils"
	"github.com/kaptinlin/gozod/jsonschema"
	"github.com/kaptinlin/gozod/types"
)

// =============================================================================
// CORE DEFINITIONS – Basic validation primitives, checks, and configurations
// =============================================================================

// ZodType is a generic alias for core.ZodType for ergonomic use.
type ZodType[T any] = core.ZodType[T]

// Schema and configuration aliases
type (
	SchemaParams = core.SchemaParams
	ObjectSchema = core.ObjectSchema
	StructSchema = core.StructSchema
	ZodConfig    = core.ZodConfig
)

// Global configuration functions
var (
	SetConfig = core.SetConfig
	Config    = core.Config
)

// Validation check aliases
type (
	ZodCheck           = core.ZodCheck
	ZodCheckInternals  = core.ZodCheckInternals
	ZodCheckDef        = core.ZodCheckDef
	ZodCheckFn         = core.ZodCheckFn
	ZodWhenFn          = core.ZodWhenFn
	CheckParams        = core.CheckParams
	CustomParams       = core.CustomParams
	ZodRefineFn[T any] = core.ZodRefineFn[T]
)

// Validation payload and issue code aliases
type (
	ParsePayload = core.ParsePayload
	IssueCode    = core.IssueCode
)

// Issue code constants
const (
	IssueInvalidType      = core.InvalidType
	IssueInvalidValue     = core.InvalidValue
	IssueInvalidFormat    = core.InvalidFormat
	IssueInvalidUnion     = core.InvalidUnion
	IssueInvalidKey       = core.InvalidKey
	IssueInvalidElement   = core.InvalidElement
	IssueTooBig           = core.TooBig
	IssueTooSmall         = core.TooSmall
	IssueNotMultipleOf    = core.NotMultipleOf
	IssueUnrecognizedKeys = core.UnrecognizedKeys
	IssueCustom           = core.Custom
)

// =============================================================================
// TYPE ALIASES – Schema definitions for various data types
// =============================================================================

// -----------------------------------------------------------------------------
// Primitive Types
// -----------------------------------------------------------------------------
type (
	ZodString[T types.StringConstraint]          = types.ZodString[T]
	ZodBool[T types.BoolConstraint]              = types.ZodBool[T]
	ZodInteger[T types.IntegerConstraint, R any] = types.ZodInteger[T, R]
	ZodFloat[T types.FloatConstraint, R any]     = types.ZodFloat[T, R]
	ZodComplex[T types.ComplexConstraint]        = types.ZodComplex[T]
	ZodTime[T types.TimeConstraint]              = types.ZodTime[T]
)

// -----------------------------------------------------------------------------
// String Format Types
// -----------------------------------------------------------------------------
type (
	// ZodEmail validates email addresses.
	ZodEmail[T types.EmailConstraint]      = types.ZodEmail[T]
	ZodEmoji[T types.StringConstraint]     = types.ZodEmoji[T]
	ZodBase64[T types.StringConstraint]    = types.ZodBase64[T]
	ZodBase64URL[T types.StringConstraint] = types.ZodBase64URL[T]
	ZodHex[T types.StringConstraint]       = types.ZodHex[T]

	// ZodIPv4 validates IPv4 addresses.
	ZodIPv4[T types.StringConstraint]     = types.ZodIPv4[T]
	ZodIPv6[T types.StringConstraint]     = types.ZodIPv6[T]
	ZodCIDRv4[T types.StringConstraint]   = types.ZodCIDRv4[T]
	ZodCIDRv6[T types.StringConstraint]   = types.ZodCIDRv6[T]
	ZodURL[T types.StringConstraint]      = types.ZodURL[T]
	ZodHostname[T types.StringConstraint] = types.ZodHostname[T]
	ZodMAC[T types.StringConstraint]      = types.ZodMAC[T]
	ZodE164[T types.StringConstraint]     = types.ZodE164[T]
	URLOptions                            = types.URLOptions

	// ZodIso validates ISO 8601 formatted strings.
	ZodIso[T types.IsoConstraint] = types.ZodIso[T]
	IsoDatetimeOptions            = types.IsoDatetimeOptions
	IsoTimeOptions                = types.IsoTimeOptions

	// ZodCUID validates CUID strings.
	ZodCUID[T types.StringConstraint]   = types.ZodCUID[T]
	ZodCUID2[T types.StringConstraint]  = types.ZodCUID2[T]
	ZodGUID[T types.StringConstraint]   = types.ZodGUID[T]
	ZodULID[T types.StringConstraint]   = types.ZodULID[T]
	ZodXID[T types.StringConstraint]    = types.ZodXID[T]
	ZodKSUID[T types.StringConstraint]  = types.ZodKSUID[T]
	ZodNanoID[T types.StringConstraint] = types.ZodNanoID[T]
	ZodUUID[T types.StringConstraint]   = types.ZodUUID[T]
	ZodJWT[T types.StringConstraint]    = types.ZodJWT[T]
	JWTOptions                          = types.JWTOptions
)

// -----------------------------------------------------------------------------
// Collection Types
// -----------------------------------------------------------------------------
type (
	ZodArray[T any, R any]      = types.ZodArray[T, R]
	ZodSlice[T any, R any]      = types.ZodSlice[T, R]
	ZodTuple[T any, R any]      = types.ZodTuple[T, R]
	ZodMap[T any, R any]        = types.ZodMap[T, R]
	ZodSet[T comparable, R any] = types.ZodSet[T, R]
	ZodRecord[T any, R any]     = types.ZodRecord[T, R]
	ZodObject[T any, R any]     = types.ZodObject[T, R]
	ZodStruct[T any, R any]     = types.ZodStruct[T, R]
)

// -----------------------------------------------------------------------------
// Composite Types
// -----------------------------------------------------------------------------
type (
	ZodUnion[T any, R any]              = types.ZodUnion[T, R]
	ZodIntersection[T any, R any]       = types.ZodIntersection[T, R]
	ZodDiscriminatedUnion[T any, R any] = types.ZodDiscriminatedUnion[T, R]
)

// -----------------------------------------------------------------------------
// Special Types
// -----------------------------------------------------------------------------
type (
	ZodAny[T any, R any]                        = types.ZodAny[T, R]
	ZodUnknown[T any, R any]                    = types.ZodUnknown[T, R]
	ZodNever[T any, R any]                      = types.ZodNever[T, R]
	ZodNil[T any, R any]                        = types.ZodNil[T, R]
	ZodFile[T any, R any]                       = types.ZodFile[T, R]
	ZodFunction[T types.FunctionConstraint]     = types.ZodFunction[T]
	ZodStringBool[T types.StringBoolConstraint] = types.ZodStringBool[T]
	ZodLazy[T types.LazyConstraint]             = types.ZodLazy[T]
	ZodEnum[T comparable, R any]                = types.ZodEnum[T, R]
	ZodLiteral[T comparable, R any]             = types.ZodLiteral[T, R]
)

// =============================================================================
// CONSTRUCTORS – Functions to create schema instances
// =============================================================================

// -----------------------------------------------------------------------------
// Primitive Type Constructors
// -----------------------------------------------------------------------------

// String constructors
var (
	String       = types.String
	StringPtr    = types.StringPtr
	Email        = types.Email
	EmailPtr     = types.EmailPtr
	Emoji        = types.Emoji
	EmojiPtr     = types.EmojiPtr
	Base64       = types.Base64
	Base64Ptr    = types.Base64Ptr
	Base64URL    = types.Base64URL
	Base64URLPtr = types.Base64URLPtr
	Hex          = types.Hex
	HexPtr       = types.HexPtr
)

// Boolean constructors
var (
	Bool    = types.Bool
	BoolPtr = types.BoolPtr
)

// Integer constructors
var (
	Int      = types.Int
	IntPtr   = types.IntPtr
	Int8     = types.Int8
	Int8Ptr  = types.Int8Ptr
	Int16    = types.Int16
	Int16Ptr = types.Int16Ptr
	Int32    = types.Int32
	Int32Ptr = types.Int32Ptr
	Int64    = types.Int64
	Int64Ptr = types.Int64Ptr
)

// Unsigned integer constructors
var (
	Uint      = types.Uint
	UintPtr   = types.UintPtr
	Uint8     = types.Uint8
	Uint8Ptr  = types.Uint8Ptr
	Uint16    = types.Uint16
	Uint16Ptr = types.Uint16Ptr
	Uint32    = types.Uint32
	Uint32Ptr = types.Uint32Ptr
	Uint64    = types.Uint64
	Uint64Ptr = types.Uint64Ptr
)

// Float constructors
var (
	Float      = types.Float
	FloatPtr   = types.FloatPtr
	Float32    = types.Float32
	Float32Ptr = types.Float32Ptr
	Float64    = types.Float64
	Float64Ptr = types.Float64Ptr
	Number     = types.Number
	NumberPtr  = types.NumberPtr
)

// BigInt constructors
var (
	BigInt    = types.BigInt
	BigIntPtr = types.BigIntPtr
)

// Complex number constructors
var (
	Complex       = types.Complex
	ComplexPtr    = types.ComplexPtr
	Complex64     = types.Complex64
	Complex64Ptr  = types.Complex64Ptr
	Complex128    = types.Complex128
	Complex128Ptr = types.Complex128Ptr
)

// Time constructors
var (
	Time    = types.Time
	TimePtr = types.TimePtr
)

// -----------------------------------------------------------------------------
// String Format Constructors
// -----------------------------------------------------------------------------

// Network type constructors
var (
	IPv4             = types.IPv4
	IPv4Ptr          = types.IPv4Ptr
	IPv6             = types.IPv6
	IPv6Ptr          = types.IPv6Ptr
	CIDRv4           = types.CIDRv4
	CIDRv4Ptr        = types.CIDRv4Ptr
	CIDRv6           = types.CIDRv6
	CIDRv6Ptr        = types.CIDRv6Ptr
	URL              = types.URL
	URLPtr           = types.URLPtr
	Hostname         = types.Hostname
	HostnamePtr      = types.HostnamePtr
	MAC              = types.MAC
	MACPtr           = types.MACPtr
	MACWithDelimiter = types.MACWithDelimiter
	E164             = types.E164
	E164Ptr          = types.E164Ptr
	HTTPURL          = types.HTTPURL
	HTTPURLPtr       = types.HTTPURLPtr
)

// ISO 8601 format constructors
var (
	Iso            = types.Iso
	IsoPtr         = types.IsoPtr
	IsoDateTime    = types.IsoDateTime
	IsoDateTimePtr = types.IsoDateTimePtr
	IsoDate        = types.IsoDate
	IsoDatePtr     = types.IsoDatePtr
	IsoTime        = types.IsoTime
	IsoTimePtr     = types.IsoTimePtr
	IsoDuration    = types.IsoDuration
	IsoDurationPtr = types.IsoDurationPtr
)

// ISO precision constants
var (
	PrecisionMinute      = types.PrecisionMinute
	PrecisionSecond      = types.PrecisionSecond
	PrecisionDecisecond  = types.PrecisionDecisecond
	PrecisionCentisecond = types.PrecisionCentisecond
	PrecisionMillisecond = types.PrecisionMillisecond
	PrecisionMicrosecond = types.PrecisionMicrosecond
	PrecisionNanosecond  = types.PrecisionNanosecond
)

// Unique identifier constructors
var (
	Cuid      = types.Cuid
	CuidPtr   = types.CuidPtr
	Cuid2     = types.Cuid2
	Cuid2Ptr  = types.Cuid2Ptr
	GUID      = types.GUID
	GUIDPtr   = types.GUIDPtr
	Ulid      = types.Ulid
	UlidPtr   = types.UlidPtr
	Xid       = types.Xid
	XidPtr    = types.XidPtr
	Ksuid     = types.Ksuid
	KsuidPtr  = types.KsuidPtr
	Nanoid    = types.Nanoid
	NanoidPtr = types.NanoidPtr
	UUID      = types.UUID
	UUIDPtr   = types.UUIDPtr
	Uuidv4    = types.Uuidv4
	Uuidv4Ptr = types.Uuidv4Ptr
	Uuidv6    = types.Uuidv6
	Uuidv6Ptr = types.Uuidv6Ptr
	Uuidv7    = types.Uuidv7
	Uuidv7Ptr = types.Uuidv7Ptr
	JWT       = types.JWT
	JWTPtr    = types.JWTPtr
)

// -----------------------------------------------------------------------------
// Collection Type Constructors
// -----------------------------------------------------------------------------

var (
	Array    = types.Array
	ArrayPtr = types.ArrayPtr
	Map      = types.Map
	MapPtr   = types.MapPtr
)

// Set creates a set schema with element validation (returns value constraint).
// In Go, sets are represented as map[T]struct{} where T must be comparable.
// TypeScript Zod v4 equivalent: z.set(schema)
//
// Example:
//
//	schema := Set[string](String())
//	result, _ := schema.Parse(map[string]struct{}{"a": {}, "b": {}})
func Set[T comparable](valueSchema any, paramArgs ...any) *ZodSet[T, map[T]struct{}] {
	return types.Set[T](valueSchema, paramArgs...)
}

// SetPtr creates a set schema with pointer constraint (returns pointer constraint).
func SetPtr[T comparable](valueSchema any, paramArgs ...any) *ZodSet[T, *map[T]struct{}] {
	return types.SetPtr[T](valueSchema, paramArgs...)
}

// Tuple creates a tuple schema with fixed positional items
// TypeScript Zod v4 equivalent: z.tuple([...])
//
// Usage:
//
//	tuple := gozod.Tuple(gozod.String(), gozod.Int())
//	result, err := tuple.Parse([]any{"hello", 42})
var (
	Tuple         = types.Tuple
	TupleWithRest = types.TupleWithRest
	TuplePtr      = types.TuplePtr
)

// Record creates a record schema with the specified key schema and value schema.
// Example: Record(String(), Int()) parses map[string]int.
func Record[K any, V any](keySchema any, valueSchema core.ZodType[V], paramArgs ...any) *ZodRecord[map[string]V, map[string]V] {
	return types.RecordTyped[map[string]V, map[string]V](keySchema, valueSchema, paramArgs...)
}

// RecordPtr is the pointer-returning counterpart of Record.
func RecordPtr[K any, V any](keySchema any, valueSchema core.ZodType[V], paramArgs ...any) *ZodRecord[map[string]V, *map[string]V] {
	return types.RecordTyped[map[string]V, *map[string]V](keySchema, valueSchema, paramArgs...)
}

// LooseRecord creates a record schema that passes through non-matching keys unchanged.
// Unlike regular Record which errors on keys that don't match the key schema,
// LooseRecord preserves non-matching keys without validation.
//
// TypeScript Zod v4 equivalent: z.looseRecord(keySchema, valueSchema)
//
// Example:
//
//	// Only validate keys starting with "S_"
//	schema := LooseRecord(String().Regex(`^S_`), String())
//	result, _ := schema.Parse(map[string]any{"S_name": "John", "other": 123})
//	// Result: {"S_name": "John", "other": 123} - "other" key is preserved
var (
	LooseRecord    = types.LooseRecord
	LooseRecordPtr = types.LooseRecordPtr
)

func Slice[T any](elementSchema any, paramArgs ...any) *ZodSlice[T, []T] {
	return types.Slice[T](elementSchema, paramArgs...)
}

func SlicePtr[T any](elementSchema any, paramArgs ...any) *ZodSlice[T, *[]T] {
	return types.SlicePtr[T](elementSchema, paramArgs...)
}

var (
	Object          = types.Object
	ObjectPtr       = types.ObjectPtr
	StrictObject    = types.StrictObject
	StrictObjectPtr = types.StrictObjectPtr
	LooseObject     = types.LooseObject
	LooseObjectPtr  = types.LooseObjectPtr
)

func Struct[T any](params ...any) *ZodStruct[T, T] {
	return types.Struct[T](params...)
}

func StructPtr[T any](params ...any) *ZodStruct[T, *T] {
	return types.StructPtr[T](params...)
}

// -----------------------------------------------------------------------------
// Composite Type Constructors
// -----------------------------------------------------------------------------
var (
	Union                 = types.Union
	UnionPtr              = types.UnionPtr
	Xor                   = types.Xor
	XorPtr                = types.XorPtr
	XorOf                 = types.XorOf
	Intersection          = types.Intersection
	IntersectionPtr       = types.IntersectionPtr
	DiscriminatedUnion    = types.DiscriminatedUnion
	DiscriminatedUnionPtr = types.DiscriminatedUnionPtr
)

// -----------------------------------------------------------------------------
// Special Type Constructors
// -----------------------------------------------------------------------------
var (
	Any           = types.Any
	AnyPtr        = types.AnyPtr
	Unknown       = types.Unknown
	UnknownPtr    = types.UnknownPtr
	Never         = types.Never
	NeverPtr      = types.NeverPtr
	Nil           = types.Nil
	NilPtr        = types.NilPtr
	File          = types.File
	FilePtr       = types.FilePtr
	Function      = types.Function
	FunctionPtr   = types.FunctionPtr
	StringBool    = types.StringBool
	StringBoolPtr = types.StringBoolPtr
)

func Literal[T comparable](value T, params ...any) *ZodLiteral[T, T] {
	return types.Literal(value, params...)
}

func LiteralPtr[T comparable](value T, params ...any) *ZodLiteral[T, *T] {
	return types.LiteralPtr(value, params...)
}

func LiteralOf[T comparable](values []T, params ...any) *ZodLiteral[T, T] {
	return types.LiteralOf(values, params...)
}

func LiteralPtrOf[T comparable](values []T, params ...any) *ZodLiteral[T, *T] {
	return types.LiteralPtrOf(values, params...)
}

func Enum[T comparable](values ...T) *ZodEnum[T, T] {
	return types.Enum(values...)
}

func EnumSlice[T comparable](values []T) *ZodEnum[T, T] {
	return types.EnumSlice(values)
}

func EnumMap[T comparable](entries map[string]T, params ...any) *ZodEnum[T, T] {
	return types.EnumMap(entries, params...)
}

func EnumPtr[T comparable](values ...T) *ZodEnum[T, *T] {
	return types.EnumPtr(values...)
}

func EnumSlicePtr[T comparable](values []T) *ZodEnum[T, *T] {
	return types.EnumSlicePtr(values)
}

func EnumMapPtr[T comparable](entries map[string]T, params ...any) *ZodEnum[T, *T] {
	return types.EnumMapPtr(entries, params...)
}

var (
	LazyAny = types.LazyAny
	LazyPtr = types.LazyPtr
)

func Lazy[S types.ZodSchemaType](getter func() S, params ...any) *types.ZodLazyTyped[S] {
	return types.Lazy(getter, params...)
}

// =============================================================================
// ERROR-RELATED EXPORTS
// =============================================================================

// Error type aliases
type (
	ZodError    = issues.ZodError
	ZodIssue    = core.ZodIssue
	ZodRawIssue = core.ZodRawIssue
)

// IsZodError checks whether an error is a ZodError.
var IsZodError = issues.IsZodError

// -----------------------------------------------------------------------------
// Error Formatting Utilities
// -----------------------------------------------------------------------------
// Expose commonly used error formatting types so that callers do not need to
// import the internal/issues package directly.
//
// These are simple type aliases to the corresponding definitions in the
// internal package, allowing external consumers to work with them via the
// primary gozod import path.

type (
	ZodFormattedError = issues.ZodFormattedError
	ZodErrorTree      = issues.ZodErrorTree
	FlattenedError    = issues.FlattenedError
	MessageFormatter  = issues.MessageFormatter
)

// Re-export TypeScript Zod v4 compatible error formatting functions so that
// callers can perform common error transformations without depending on the
// internal package path. These functions match TypeScript Zod's API patterns:
// z.treeifyError(), z.prettifyError(), z.flattenError()
var (
	TreeifyError  = issues.TreeifyError
	PrettifyError = issues.PrettifyError
	FlattenError  = issues.FlattenError
	FormatError   = issues.FormatError
)

// Advanced error formatting functions with custom mappers - these provide
// additional flexibility while maintaining TypeScript Zod v4 compatibility
var (
	TreeifyErrorWithMapper     = issues.TreeifyErrorWithMapper
	PrettifyErrorWithFormatter = issues.PrettifyErrorWithFormatter
	FlattenErrorWithMapper     = issues.FlattenErrorWithMapper
	FlattenErrorWithFormatter  = issues.FlattenErrorWithFormatter
)

// -----------------------------------------------------------------------------
// Path Utilities
// -----------------------------------------------------------------------------
// Error path formatting utilities for displaying validation error paths
// in user-friendly formats

var (
	ToDotPath       = utils.ToDotPath
	FormatErrorPath = utils.FormatErrorPath
)

// -----------------------------------------------------------------------------
// REGISTRY API RE-EXPORTS
// -----------------------------------------------------------------------------

// Registry provides a lightweight, type-safe store for attaching metadata to
// any Schema. This is an alias to core.Registry to make the API available via
// the primary gozod package.
//
// Example:
//
//	fieldReg := gozod.NewRegistry[FieldMeta]()
//	fieldReg.Add(nameSchema, FieldMeta{Title: "User Name"})
//
// See docs/metadata.md for usage patterns and best practices.
type Registry[M any] = core.Registry[M]

// GlobalMeta mirrors common JSON-Schema keys and serves as a convenient default
// metadata structure. Alias to core.GlobalMeta so callers don't need to import
// the core package.
type GlobalMeta = core.GlobalMeta

// NewRegistry creates an empty Registry. It's a thin wrapper around
// core.NewRegistry to expose the constructor at the root package level.
func NewRegistry[M any]() *Registry[M] {
	return core.NewRegistry[M]()
}

// GlobalRegistry is the framework-provided, process-wide registry instance. Use
// it to store shared metadata that should be accessible throughout your
// application.
var GlobalRegistry = core.GlobalRegistry

// =============================================================================
// STRUCT TAG SUPPORT
// =============================================================================

// FromStruct creates a ZodStruct schema from struct tags
// This provides convenient tag-based validation for Go structs
//
// Example:
//
//	type User struct {
//	    Name  string `gozod:"required,min=2,max=50"`
//	    Email string `gozod:"required,email"`
//	}
//
//	schema := gozod.FromStruct[User]()
func FromStruct[T any]() *types.ZodStruct[T, T] {
	return types.FromStruct[T]()
}

// FromStructPtr creates a ZodStruct schema for pointer types from struct tags
// This is useful for handling optional/nullable struct inputs
func FromStructPtr[T any]() *types.ZodStruct[T, *T] {
	return types.FromStructPtr[T]()
}

// =============================================================================
// UTILITY FUNCTIONS
// =============================================================================

// Apply integrates external functions into schema chains.
// It takes a schema and a function that receives the schema and returns
// a potentially different schema type. This enables modular composition
// of common validation patterns.
//
// TypeScript Zod v4 equivalent: schema.apply(fn)
// In Go, this is a standalone generic function due to language constraints.
//
// Example:
//
//	// Define a reusable modifier
//	func addCommonStringChecks[T types.StringConstraint](s *gozod.ZodString[T]) *gozod.ZodString[T] {
//	    return s.Min(1).Max(100).Trim()
//	}
//
//	// Apply it to a schema
//	schema := gozod.Apply(gozod.String(), addCommonStringChecks)
//
//	// Chain with other modifiers
//	optionalSchema := gozod.Apply(gozod.String(), addCommonStringChecks).Optional()
func Apply[S any, R any](schema S, fn func(S) R) R {
	return fn(schema)
}

// =============================================================================
// METADATA CHECK FACTORIES
// TypeScript Zod v4 compatible check factories for schema metadata
// =============================================================================

// Describe creates a check that registers a description in the global registry.
// This is a no-op validation check that only attaches metadata when the check
// is added to a schema.
//
// TypeScript Zod v4 equivalent: z.describe(description)
//
// Example:
//
//	schema := gozod.String().Check(gozod.Describe("User email address"))
//	meta, _ := gozod.GlobalRegistry.Get(schema)
//	// meta.Description == "User email address"
var Describe = checks.Describe

// Meta creates a check that registers metadata in the global registry.
// This is a no-op validation check that only attaches metadata when the check
// is added to a schema.
//
// TypeScript Zod v4 equivalent: z.meta(metadata)
//
// Example:
//
//	schema := gozod.Number().Check(gozod.Meta(gozod.GlobalMeta{
//	    Title: "Age",
//	    Description: "User's age in years",
//	}))
//	meta, _ := gozod.GlobalRegistry.Get(schema)
//	// meta.Title == "Age"
//	// meta.Description == "User's age in years"
var Meta = checks.Meta

// =============================================================================
// JSON SCHEMA CONVERSION
// =============================================================================

// JSONSchemaOptions configures the ToJSONSchema conversion.
// Re-exported from jsonschema subpackage for convenience.
type JSONSchemaOptions = jsonschema.Options

// OverrideContext provides context for the Override function in JSON Schema conversion.
// Re-exported from jsonschema subpackage for convenience.
type OverrideContext = jsonschema.OverrideContext

// FromJSONSchemaOptions configures the FromJSONSchema conversion.
// Re-exported from jsonschema subpackage for convenience.
type FromJSONSchemaOptions = jsonschema.FromJSONSchemaOptions

// ToJSONSchema converts a GoZod schema or registry into a JSON Schema instance.
//
// TypeScript Zod v4 equivalent: zodToJsonSchema(schema)
//
// Example:
//
//	schema := gozod.Object(gozod.ObjectSchema{
//	    "name": gozod.String().Min(1),
//	    "age":  gozod.Int().Min(0),
//	})
//	jsonSchema, err := gozod.ToJSONSchema(schema)
var ToJSONSchema = jsonschema.ToJSONSchema

// FromJSONSchema converts a kaptinlin/jsonschema Schema to a GoZod schema.
//
// Example:
//
//	jsonSchema := &lib.Schema{Type: []string{"string"}}
//	zodSchema, err := gozod.FromJSONSchema(jsonSchema)
var FromJSONSchema = jsonschema.FromJSONSchema

// JSON Schema conversion error variables re-exported for convenience.
var (
	// ToJSONSchema errors
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

	// FromJSONSchema errors
	ErrUnsupportedJSONSchemaType    = jsonschema.ErrUnsupportedJSONSchemaType
	ErrUnsupportedJSONSchemaKeyword = jsonschema.ErrUnsupportedJSONSchemaKeyword
	ErrInvalidJSONSchema            = jsonschema.ErrInvalidJSONSchema
	ErrJSONSchemaCircularRef        = jsonschema.ErrJSONSchemaCircularRef
	ErrJSONSchemaPatternCompile     = jsonschema.ErrJSONSchemaPatternCompile
	ErrJSONSchemaIfThenElse         = jsonschema.ErrJSONSchemaIfThenElse
	ErrJSONSchemaPatternProperties  = jsonschema.ErrJSONSchemaPatternProperties
	ErrJSONSchemaDynamicRef         = jsonschema.ErrJSONSchemaDynamicRef
	ErrJSONSchemaUnevaluatedProps   = jsonschema.ErrJSONSchemaUnevaluatedProps
	ErrJSONSchemaUnevaluatedItems   = jsonschema.ErrJSONSchemaUnevaluatedItems
	ErrJSONSchemaDependentSchemas   = jsonschema.ErrJSONSchemaDependentSchemas
	ErrJSONSchemaPropertyNames      = jsonschema.ErrJSONSchemaPropertyNames
	ErrJSONSchemaContains           = jsonschema.ErrJSONSchemaContains
)
