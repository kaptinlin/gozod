package gozod

import (
	"github.com/kaptinlin/gozod/core"
	issues "github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/types"
)

// =============================================================================
// TYPE ALIASES – re-export core and internal types at package root
// =============================================================================

// Validation primitives
// Generic ZodType alias for ergonomic use
// Note: Generic type aliases are supported from Go 1.20+
// Users can declare variables such as: var s gozod.ZodType[any,any]
type ZodType[I any, O any] = core.ZodType[I, O]

// Common helper aliases expected by README examples
// SchemaParams allows optional parameters when constructing schemas
// ObjectSchema / StructSchema provide shape definitions for Object & Struct

type (
	SchemaParams = core.SchemaParams
	ObjectSchema = core.ObjectSchema
	StructSchema = core.StructSchema

	// Error related aliases
	ZodError    = issues.ZodError
	ZodIssue    = core.ZodIssue
	ZodRawIssue = core.ZodRawIssue

	// Global configuration
	ZodConfig = core.ZodConfig
)

// =============================================================================
// FUNCTION / VALUE FORWARDERS – expose types package constructors at root
// =============================================================================

// Primitive & literal types
var (
	Any     = types.Any
	Unknown = types.Unknown
	Never   = types.Never
	String  = types.String
	Bool    = types.Bool

	// Integer variants
	Int    = types.Int
	Int8   = types.Int8
	Int16  = types.Int16
	Int32  = types.Int32
	Int64  = types.Int64
	Uint   = types.Uint
	Uint8  = types.Uint8
	Uint16 = types.Uint16
	Uint32 = types.Uint32
	Uint64 = types.Uint64
	Byte   = types.Byte
	Rune   = types.Rune

	// Floating-point & numeric
	Float32 = types.Float32
	Float64 = types.Float64
	Number  = types.Number

	// Big / complex numbers
	BigInt       = types.BigInt
	Int64BigInt  = types.Int64BigInt
	Uint64BigInt = types.Uint64BigInt
	Complex64    = types.Complex64
	Complex128   = types.Complex128

	// Collections & structured data
	Slice        = types.Slice
	Array        = types.Array
	Map          = types.Map
	Record       = types.Record
	Object       = types.Object
	Struct       = types.Struct
	StrictStruct = types.StrictStruct
	LooseStruct  = types.LooseStruct

	// Miscellaneous specialised schemas
	File     = types.File
	Function = types.Function
	Nil      = types.Nil
	Null     = types.Null

	// Combinators & higher-order schemas
	Union              = types.Union
	Intersection       = types.Intersection
	DiscriminatedUnion = types.DiscriminatedUnion
	Lazy               = types.Lazy

	// Utilities
	Check = types.Check
)

// =============================================================================
// COERCE NAMESPACE – expose coercive helpers
// =============================================================================

var Coerce = types.Coerce

// =============================================================================
// ERROR UTILITIES
// =============================================================================

var IsZodError = issues.IsZodError

// =============================================================================
// GLOBAL CONFIGURATION – expose Config helpers
// =============================================================================

var (
	Config    = core.Config
	GetConfig = core.GetConfig
)
