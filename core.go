package gozod

import "github.com/kaptinlin/gozod/core"

// ZodType is a generic alias for core.ZodType for ergonomic use.
type ZodType[T any] = core.ZodType[T]

// Unwrapper allows wrapper types to expose their underlying value for validation.
type Unwrapper = core.Unwrapper

type (
	SchemaParams = core.SchemaParams
	ObjectSchema = core.ObjectSchema
	StructSchema = core.StructSchema
	ZodConfig    = core.ZodConfig
)

var (
	SetConfig = core.SetConfig
	Config    = core.Config
)

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

type (
	ParsePayload = core.ParsePayload
	IssueCode    = core.IssueCode
)

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
