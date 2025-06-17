package gozod

import (
	"errors"
)

// Core parsing errors
var (
	ErrSchemaNoInternals     = errors.New("schema has no internals")
	ErrSchemaNoParseFunction = errors.New("schema has no parse function")
	ErrParseReturnedNil      = errors.New("parse function returned nil result")
	ErrTypeAssertionFailed   = errors.New("type assertion failed")
)

// Transform errors
var (
	ErrTransformNilString  = errors.New("cannot transform nil *string")
	ErrTransformNilBool    = errors.New("cannot transform nil *bool")
	ErrTransformNilBigInt  = errors.New("cannot transform nil *big.Int")
	ErrTransformNilComplex = errors.New("cannot transform nil complex number")
	ErrTransformNilArray   = errors.New("cannot transform nil *[]interface{}")
	ErrTransformNilSlice   = errors.New("cannot transform nil slice")
	ErrTransformNilMap     = errors.New("cannot transform nil map")
	ErrTransformNilObject  = errors.New("cannot transform nil object")
	ErrTransformNilStruct  = errors.New("cannot transform nil struct")
	ErrTransformNilRecord  = errors.New("cannot transform nil record")
	ErrTransformNilUnion   = errors.New("cannot transform nil union")
	ErrTransformNilLiteral = errors.New("cannot transform nil literal")
	ErrTransformNilGeneric = errors.New("cannot transform nil *T")
)

// Type validation errors
var (
	ErrExpectedString   = errors.New("expected string or *string")
	ErrExpectedBool     = errors.New("expected bool")
	ErrExpectedArray    = errors.New("expected []interface{} or *[]interface{}")
	ErrExpectedSlice    = errors.New("expected []interface{} or *[]interface{}")
	ErrExpectedMap      = errors.New("expected map type")
	ErrExpectedFunction = errors.New("expected function")
	ErrExpectedFile     = errors.New("expected file type")
	ErrExpectedStruct   = errors.New("expected struct or map[string]interface{}")
	ErrExpectedRecord   = errors.New("expected map[interface{}]interface{} or compatible map type")
	ErrExpectedLiteral  = errors.New("expected literal value or pointer to literal value")
	ErrExpectedNumeric  = errors.New("expected numeric type")
	ErrExpectedBigInt   = errors.New("expected big.Int")
	ErrExpectedComplex  = errors.New("expected complex number")
)

// Function signature errors
var (
	ErrFunctionNil                = errors.New("implement() must be called with a function, got nil")
	ErrFunctionInvalid            = errors.New("implement() must be called with a function")
	ErrFunctionSignatureMismatch  = errors.New("function signature mismatch")
	ErrFunctionParameterMismatch  = errors.New("function parameter type mismatch")
	ErrFunctionReturnTypeMismatch = errors.New("function return type mismatch")
	ErrTransformFunctionSignature = errors.New("Transform function must have signature func(mapType, *RefinementContext) (any, error)")
	ErrTransformFunctionParameter = errors.New("Transform function second parameter must be *RefinementContext")
	ErrTransformFunctionReturn    = errors.New("Transform function must return (any, error)")
)

// Conversion errors
var (
	ErrCannotConvertType                = errors.New("cannot convert type")
	ErrCannotConvertObject              = errors.New("cannot convert object to map")
	ErrCannotAssignNilToNonPointer      = errors.New("cannot assign nil to non-pointer field")
	ErrCannotMergeDifferentTypes        = errors.New("cannot merge different types")
	ErrCannotMergeIncompatibleValues    = errors.New("cannot merge incompatible values")
	ErrExpectedMapsForMerging           = errors.New("expected maps for merging")
	ErrCannotMergeSlicesDifferentLength = errors.New("cannot merge slices of different lengths")
)

// Field operation errors
var (
	ErrFailedToSetField     = errors.New("failed to set field")
	ErrMergeConflictAtKey   = errors.New("merge conflict at key")
	ErrMergeConflictAtIndex = errors.New("merge conflict at index")
)

// Transform specific errors
var (
	ErrEmptyString                = errors.New("empty string")
	ErrTransformationFailed       = errors.New("transformation failed")
	ErrTransformFailed            = errors.New("transform failed")
	ErrExpectedStringInput        = errors.New("expected string input")
	ErrCleanedStringTooShort      = errors.New("cleaned string too short")
	ErrExpectedStringType         = errors.New("expected string")
	ErrContainsInvalidWord        = errors.New("contains invalid word")
	ErrNegativeNumbersNotAllowed  = errors.New("negative numbers not allowed in transform")
	ErrNumberTooLargeForFactorial = errors.New("number too large for factorial")
	ErrInvalidNumber              = errors.New("invalid number")
	ErrTransformError             = errors.New("transform error")
	ErrCannotTransformNilMap      = errors.New("cannot transform nil map")
	ErrExpectedMapType            = errors.New("expected map type")
)

// Type conversion errors
var (
	ErrInvalidTypeForTransform       = errors.New("invalid type for transform")
	ErrUnexpectedStringResultType    = errors.New("unexpected string result type")
	ErrFailedToConvertMapToRequired  = errors.New("failed to convert map to required type")
	ErrUnsupportedTransformParameter = errors.New("unsupported transform function parameter type")
	ErrInvalidComplexFormat          = errors.New("invalid complex number format")
)
