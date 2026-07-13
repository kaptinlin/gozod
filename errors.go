package gozod

import (
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/internal/utils"
	"github.com/kaptinlin/gozod/types"
)

type (
	ZodError           = issues.ZodError
	ZodIssue           = core.ZodIssue
	ZodRawIssue        = core.ZodRawIssue
	DiscriminatorError = types.DiscriminatorError
)

var (
	ErrNilDiscriminatedUnionOption = types.ErrOptionIsNil
	ErrMissingDiscriminatorValues  = types.ErrNoDiscriminatorValues
	ErrDuplicateDiscriminator      = types.ErrDuplicateDiscriminator
	ErrNoValidDiscriminators       = types.ErrNoValidDiscriminators
)

func IsZodError(err error, target **ZodError) bool {
	return issues.IsZodError(err, target)
}

type (
	ZodFormattedError = issues.ZodFormattedError
	ZodErrorTree      = issues.ZodErrorTree
	FlattenedError    = issues.FlattenedError
	MessageFormatter  = issues.MessageFormatter
)

func TreeifyError(zodErr *ZodError) *ZodErrorTree {
	return issues.TreeifyError(zodErr)
}

func PrettifyError(zodErr *ZodError) string {
	return issues.PrettifyError(zodErr)
}

func FlattenError(zodErr *ZodError) *FlattenedError {
	return issues.FlattenError(zodErr)
}

func FormatError(zodErr *ZodError) ZodFormattedError {
	return issues.FormatError(zodErr)
}

func TreeifyErrorWithMapper(zodErr *ZodError, mapper func(ZodIssue) string) *ZodErrorTree {
	return issues.TreeifyErrorWithMapper(zodErr, mapper)
}

func PrettifyErrorWithFormatter(zodErr *ZodError, formatter MessageFormatter) string {
	return issues.PrettifyErrorWithFormatter(zodErr, formatter)
}

func FlattenErrorWithMapper(zodErr *ZodError, mapper func(ZodIssue) string) *FlattenedError {
	return issues.FlattenErrorWithMapper(zodErr, mapper)
}

func FlattenErrorWithFormatter(zodErr *ZodError, formatter MessageFormatter) *FlattenedError {
	return issues.FlattenErrorWithFormatter(zodErr, formatter)
}

func ToDotPath(path []any) string {
	return utils.ToDotPath(path)
}

func FormatErrorPath(path []any, style string) string {
	return utils.FormatErrorPath(path, style)
}
