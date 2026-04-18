package gozod

import (
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/internal/utils"
)

type (
	ZodError    = issues.ZodError
	ZodIssue    = core.ZodIssue
	ZodRawIssue = core.ZodRawIssue
)

var IsZodError = issues.IsZodError

type (
	ZodFormattedError = issues.ZodFormattedError
	ZodErrorTree      = issues.ZodErrorTree
	FlattenedError    = issues.FlattenedError
	MessageFormatter  = issues.MessageFormatter
)

var (
	TreeifyError               = issues.TreeifyError
	PrettifyError              = issues.PrettifyError
	FlattenError               = issues.FlattenError
	FormatError                = issues.FormatError
	TreeifyErrorWithMapper     = issues.TreeifyErrorWithMapper
	PrettifyErrorWithFormatter = issues.PrettifyErrorWithFormatter
	FlattenErrorWithMapper     = issues.FlattenErrorWithMapper
	FlattenErrorWithFormatter  = issues.FlattenErrorWithFormatter
	ToDotPath                  = utils.ToDotPath
	FormatErrorPath            = utils.FormatErrorPath
)
