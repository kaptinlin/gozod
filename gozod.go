package gozod

import (
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/types"
)

// =============================================================================
// CORE DEFINITIONS – Basic validation primitives, checks, and configurations
// =============================================================================

// Generic ZodType alias for ergonomic use
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
	Config    = core.Config
	GetConfig = core.GetConfig
)

// Validation check aliases
type (
	ZodCheck           = core.ZodCheck
	ZodCheckInternals  = core.ZodCheckInternals
	ZodCheckDef        = core.ZodCheckDef
	ZodCheckFn         = core.ZodCheckFn
	ZodWhenFn          = core.ZodWhenFn
	CheckParams        = core.CheckParams
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
	// Standard formats
	ZodEmail[T types.EmailConstraint]      = types.ZodEmail[T]
	ZodEmoji[T types.StringConstraint]     = types.ZodEmoji[T]
	ZodBase64[T types.StringConstraint]    = types.ZodBase64[T]
	ZodBase64URL[T types.StringConstraint] = types.ZodBase64URL[T]

	// Network formats
	ZodIPv4[T types.NetworkConstraint]   = types.ZodIPv4[T]
	ZodIPv6[T types.NetworkConstraint]   = types.ZodIPv6[T]
	ZodCIDRv4[T types.NetworkConstraint] = types.ZodCIDRv4[T]
	ZodCIDRv6[T types.NetworkConstraint] = types.ZodCIDRv6[T]
	ZodURL[T types.NetworkConstraint]    = types.ZodURL[T]
	URLOptions                           = types.URLOptions

	// ISO 8601 formats
	ZodIso[T types.IsoConstraint] = types.ZodIso[T]
	IsoDatetimeOptions            = types.IsoDatetimeOptions
	IsoTimeOptions                = types.IsoTimeOptions

	// Unique identifier formats
	ZodCUID[T types.StringConstraint]   = types.ZodCUID[T]
	ZodCUID2[T types.StringConstraint]  = types.ZodCUID2[T]
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
	ZodArray[T any, R any]  = types.ZodArray[T, R]
	ZodSlice[T any, R any]  = types.ZodSlice[T, R]
	ZodMap[T any, R any]    = types.ZodMap[T, R]
	ZodRecord[T any, R any] = types.ZodRecord[T, R]
	ZodObject[T any, R any] = types.ZodObject[T, R]
	ZodStruct[T any, R any] = types.ZodStruct[T, R]
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
	ZodNil[T any]                               = types.ZodNil[T]
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
	IPv4      = types.IPv4
	IPv4Ptr   = types.IPv4Ptr
	IPv6      = types.IPv6
	IPv6Ptr   = types.IPv6Ptr
	CIDRv4    = types.CIDRv4
	CIDRv4Ptr = types.CIDRv4Ptr
	CIDRv6    = types.CIDRv6
	CIDRv6Ptr = types.CIDRv6Ptr
	URL       = types.URL
	URLPtr    = types.URLPtr
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
	Ulid      = types.Ulid
	UlidPtr   = types.UlidPtr
	Xid       = types.Xid
	XidPtr    = types.XidPtr
	Ksuid     = types.Ksuid
	KsuidPtr  = types.KsuidPtr
	Nanoid    = types.Nanoid
	NanoidPtr = types.NanoidPtr
	Uuid      = types.Uuid
	UuidPtr   = types.UuidPtr
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

// Record creates a record schema whose value type is inferred from the provided valueSchema.
// Example: Record(Int()) returns a schema that parses map[string]int.
func Record[V any](valueSchema core.ZodType[V], paramArgs ...any) *ZodRecord[map[string]V, map[string]V] {
	return types.RecordTyped[map[string]V, map[string]V](valueSchema, paramArgs...)
}

// RecordPtr is the pointer-returning counterpart of Record.
func RecordPtr[V any](valueSchema core.ZodType[V], paramArgs ...any) *ZodRecord[map[string]V, *map[string]V] {
	return types.RecordTyped[map[string]V, *map[string]V](valueSchema, paramArgs...)
}

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
	return types.LiteralTyped[T](values, params...)
}

func LiteralPtrOf[T comparable](values []T, params ...any) *ZodLiteral[T, *T] {
	return types.LiteralTyped[T](values, params...).Nilable()
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

// Error utility function
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

// Re-export frequently used helper functions that operate on *ZodError so that
// callers can perform common error transformations without depending on the
// internal package path.
var (
	PrettifyError              = issues.PrettifyError
	PrettifyErrorWithFormatter = issues.PrettifyErrorWithFormatter
	FlattenErrorWithFormatter  = issues.FlattenErrorWithFormatter
	FlattenError               = issues.FlattenError
	FormatError                = issues.FormatError
	TreeifyError               = issues.TreeifyError
)
