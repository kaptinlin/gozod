// Package issues provides error creation, formatting, and management for GoZod validation.
package issues

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/utils"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/slicex"
	"github.com/kaptinlin/gozod/pkg/structx"
)

// ZodError represents a validation error with a collection of issues.
type ZodError struct {
	Type   any        `json:"type"`   // The expected type that failed validation
	Issues []ZodIssue `json:"issues"` // Collection of validation issues
	Zod    struct {   // Internal Zod metadata
		Output any        `json:"output"` // The output value (if any)
		Def    []ZodIssue `json:"def"`    // Issue definitions
	} `json:"_zod"`
	Stack string `json:"stack,omitempty"` // Stack trace for debugging
	Name  string `json:"name"`            // Error name identifier

	formatter MessageFormatter `json:"-"`
}

// NewZodError creates a new validation error with the given issues.
func NewZodError(issues []ZodIssue) *ZodError {
	copied := slices.Clone(issues)

	err := &ZodError{
		Type:      nil,
		Issues:    copied,
		Name:      "ZodError",
		formatter: defaultFormatter,
	}
	err.Zod.Output = nil
	err.Zod.Def = copied
	return err
}

// NewZodErrorWithFormatter creates a new validation error with a custom formatter.
func NewZodErrorWithFormatter(issues []ZodIssue, formatter MessageFormatter) *ZodError {
	err := NewZodError(issues)
	if formatter != nil {
		err.formatter = formatter
	}
	return err
}

// Error implements the error interface using the configured formatter.
func (e *ZodError) Error() string {
	if e == nil {
		return ""
	}

	if slicex.IsEmpty(e.Issues) {
		return "Validation failed"
	}

	return PrettifyErrorWithFormatter(e, e.formatter)
}

// Formatter returns the current message formatter.
func (e *ZodError) Formatter() MessageFormatter {
	return e.formatter
}

// SetFormatter sets a new message formatter for this error.
func (e *ZodError) SetFormatter(formatter MessageFormatter) {
	if formatter != nil {
		e.formatter = formatter
	}
}

// IsZodError checks if an error is a ZodError and extracts it.
func IsZodError(err error, target **ZodError) bool {
	if err == nil {
		return false
	}

	zodErr, ok := errors.AsType[*ZodError](err)
	if ok && target != nil {
		*target = zodErr
	}
	return ok
}

// ZodFormattedError represents a formatted error with hierarchical field-level grouping.
type ZodFormattedError map[string]any

// FormatError formats a ZodError into a structured error object.
func FormatError(zodErr *ZodError) ZodFormattedError {
	return FormatErrorWithMapper(zodErr, defaultIssueMapper(zodErr.formatter))
}

// FormatErrorWithMapper formats a ZodError with custom message mapping.
func FormatErrorWithMapper(zodErr *ZodError, mapper func(ZodIssue) string) ZodFormattedError {
	fieldErrors := make(ZodFormattedError)
	fieldErrors["_errors"] = []string{}

	var processError func(*ZodError)
	processError = func(ze *ZodError) {
		for _, issue := range ze.Issues {
			switch issue.Code {
			case core.InvalidUnion:
				if !slicex.IsEmpty(issue.Errors) {
					for _, unionErrors := range issue.Errors {
						if !slicex.IsEmpty(unionErrors) {
							processError(&ZodError{Issues: unionErrors, formatter: ze.formatter})
						}
					}
				}
			case core.InvalidKey, core.InvalidElement:
				if !slicex.IsEmpty(issue.Issues) {
					processError(&ZodError{Issues: issue.Issues, formatter: ze.formatter})
				}
			case core.InvalidType, core.InvalidValue, core.InvalidFormat,
				core.TooBig, core.TooSmall, core.NotMultipleOf,
				core.UnrecognizedKeys, core.Custom, core.InvalidSchema,
				core.InvalidDiscriminator, core.IncompatibleTypes, core.MissingRequired,
				core.TypeConversion, core.NilPointer:
				if slicex.IsEmpty(issue.Path) {
					if errors, ok := fieldErrors["_errors"].([]string); ok {
						fieldErrors["_errors"] = append(errors, mapper(issue))
					}
				} else {
					curr := fieldErrors
					for i, pathEl := range issue.Path {
						key := fmt.Sprintf("%v", pathEl)

						if !mapx.Has(curr, key) {
							curr[key] = ZodFormattedError{"_errors": []string{}}
						}

						currMap, ok := curr[key].(ZodFormattedError)
						if !ok {
							continue
						}

						if i == len(issue.Path)-1 {
							if errors, ok := currMap["_errors"].([]string); ok {
								currMap["_errors"] = append(errors, mapper(issue))
							}
						} else {
							curr = currMap
						}
					}
				}
			}
		}
	}

	processError(zodErr)
	return fieldErrors
}

// ZodErrorTree represents a tree-structured validation error.
type ZodErrorTree struct {
	Errors     []string                 `json:"errors"`               // Errors at this level
	Properties map[string]*ZodErrorTree `json:"properties,omitempty"` // Object property errors
	Items      []*ZodErrorTree          `json:"items,omitempty"`      // Array/slice item errors
}

// FlattenedError separates top-level form errors from field-specific errors.
type FlattenedError struct {
	FormErrors  []string            `json:"formErrors"`  // Top-level errors (path is empty)
	FieldErrors map[string][]string `json:"fieldErrors"` // Field-level errors by field name
}

// TreeifyError formats a ZodError into a tree structure.
func TreeifyError(zodErr *ZodError) *ZodErrorTree {
	return TreeifyErrorWithMapper(zodErr, defaultIssueMapper(zodErr.formatter))
}

// TreeifyErrorWithMapper converts a ZodError into a tree structure with custom message mapping.
func TreeifyErrorWithMapper(zodErr *ZodError, mapper func(ZodIssue) string) *ZodErrorTree {
	issueCount := len(zodErr.Issues)
	tree := &ZodErrorTree{
		Errors:     make([]string, 0, max(issueCount/4, 2)),
		Properties: make(map[string]*ZodErrorTree, max(issueCount/2, 4)),
		Items:      make([]*ZodErrorTree, 0, max(issueCount/4, 2)),
	}

	for _, issue := range zodErr.Issues {
		processIssueInTree(issue, tree, mapper)
	}

	return tree
}

// processIssueInTree processes an issue within a specific tree node.
func processIssueInTree(issue ZodIssue, tree *ZodErrorTree, mapper func(ZodIssue) string) {
	if slicex.IsEmpty(issue.Path) {
		tree.Errors = append(tree.Errors, mapper(issue))
		return
	}

	current := tree
	for i, pathElement := range issue.Path {
		isLast := i == len(issue.Path)-1

		switch element := pathElement.(type) {
		case string:
			if current.Properties == nil {
				current.Properties = make(map[string]*ZodErrorTree)
			}
			if current.Properties[element] == nil {
				current.Properties[element] = &ZodErrorTree{
					Errors:     []string{},
					Properties: make(map[string]*ZodErrorTree),
					Items:      []*ZodErrorTree{},
				}
			}
			current = current.Properties[element]

		case int:
			for len(current.Items) <= element {
				current.Items = append(current.Items, &ZodErrorTree{
					Errors:     []string{},
					Properties: make(map[string]*ZodErrorTree),
					Items:      []*ZodErrorTree{},
				})
			}
			current = current.Items[element]
		}

		if isLast {
			current.Errors = append(current.Errors, mapper(issue))
		}
	}
}

// FlattenError flattens a ZodError into form and field errors.
func FlattenError(zodErr *ZodError) *FlattenedError {
	return FlattenErrorWithMapper(zodErr, defaultIssueMapper(zodErr.formatter))
}

// FlattenErrorWithMapper flattens a ZodError into form and field errors with custom message mapping.
func FlattenErrorWithMapper(zodErr *ZodError, mapper func(ZodIssue) string) *FlattenedError {
	issueCount := len(zodErr.Issues)
	flattened := &FlattenedError{
		FormErrors:  make([]string, 0, max(issueCount/4, 2)),
		FieldErrors: make(map[string][]string, max(issueCount/2, 4)),
	}

	for _, issue := range zodErr.Issues {
		message := mapper(issue)

		if slicex.IsEmpty(issue.Path) {
			flattened.FormErrors = append(flattened.FormErrors, message)
		} else {
			fieldPath := fmt.Sprintf("%v", issue.Path[0])

			if flattened.FieldErrors[fieldPath] == nil {
				flattened.FieldErrors[fieldPath] = make([]string, 0, 2)
			}
			flattened.FieldErrors[fieldPath] = append(flattened.FieldErrors[fieldPath], message)
		}
	}

	return flattened
}

// FlattenErrorWithFormatter flattens a ZodError with a custom formatter.
func FlattenErrorWithFormatter(zodErr *ZodError, formatter MessageFormatter) *FlattenedError {
	return FlattenErrorWithMapper(zodErr, defaultIssueMapper(formatter))
}

// ToDotPath converts a path array to dot notation string.
func ToDotPath(path []any) string {
	return utils.ToDotPath(path)
}

// PrettifyError formats a ZodError into a readable string using its formatter.
func PrettifyError(zodErr *ZodError) string {
	return PrettifyErrorWithFormatter(zodErr, zodErr.formatter)
}

// PrettifyErrorWithFormatter formats a ZodError into a readable string with custom formatter.
func PrettifyErrorWithFormatter(zodErr *ZodError, formatter MessageFormatter) string {
	if zodErr == nil || len(zodErr.Issues) == 0 {
		return "Validation failed"
	}

	var builder strings.Builder
	builder.Grow(len(zodErr.Issues) * 50)

	for i, issue := range zodErr.Issues {
		if i > 0 {
			builder.WriteString("; ")
		}

		message := issue.Message
		if message == "" && formatter != nil {
			message = formatter.FormatMessage(core.ZodRawIssue{
				Code:       issue.Code,
				Path:       issue.Path,
				Message:    issue.Message,
				Properties: convertZodIssueToProperties(issue),
			})
		}

		if len(issue.Path) > 0 {
			pathStr := ToDotPath(issue.Path)
			builder.WriteString(pathStr)
			builder.WriteString(": ")
		}
		builder.WriteString(message)
	}

	return builder.String()
}

// defaultIssueMapper returns a mapper function that formats an issue using
// the given formatter, preferring the issue's own message when available.
// This eliminates the repeated closure pattern across FormatError, TreeifyError,
// FlattenError, and FlattenErrorWithFormatter.
func defaultIssueMapper(formatter MessageFormatter) func(ZodIssue) string {
	return func(issue ZodIssue) string {
		if issue.Message != "" {
			return issue.Message
		}
		return formatter.FormatMessage(core.ZodRawIssue{
			Code:       issue.Code,
			Path:       issue.Path,
			Message:    issue.Message,
			Properties: convertZodIssueToProperties(issue),
		})
	}
}

// excludedPropertyKeys contains the ZodIssue struct keys that should not
// appear in the properties map passed to formatters.
var excludedPropertyKeys = map[string]struct{}{
	"Code": {}, "Path": {}, "Message": {}, "Continue": {},
	"Inst": {}, "Issues": {}, "Errors": {},
}

// convertZodIssueToProperties converts a ZodIssue to properties map for the formatter.
func convertZodIssueToProperties(issue ZodIssue) map[string]any {
	if properties, err := structx.ToMap(issue); err == nil {
		result := mapx.Copy(properties)
		maps.DeleteFunc(result, func(k string, _ any) bool {
			_, excluded := excludedPropertyKeys[k]
			return excluded
		})
		return result
	}

	// Fallback: create properties map manually
	properties := make(map[string]any)

	if issue.Expected != "" {
		properties["expected"] = issue.Expected
	}
	if issue.Received != "" {
		properties["received"] = issue.Received
	}
	if issue.Minimum != nil {
		properties["minimum"] = issue.Minimum
	}
	if issue.Maximum != nil {
		properties["maximum"] = issue.Maximum
	}
	if issue.Format != "" {
		properties["format"] = issue.Format
	}
	if issue.Pattern != "" {
		properties["pattern"] = issue.Pattern
	}
	if issue.Prefix != "" {
		properties["startsWith"] = issue.Prefix
	}
	if issue.Suffix != "" {
		properties["endsWith"] = issue.Suffix
	}
	if issue.Includes != "" {
		properties["includes"] = issue.Includes
	}
	if issue.Algorithm != "" {
		properties["algorithm"] = issue.Algorithm
	}
	if issue.Divisor != nil {
		properties["divisor"] = issue.Divisor
	}
	if len(issue.Keys) > 0 {
		properties["keys"] = issue.Keys
	}
	if len(issue.Values) > 0 {
		properties["values"] = issue.Values
	}
	if issue.Origin != "" {
		properties["origin"] = issue.Origin
	}
	if issue.Key != nil {
		properties["key"] = issue.Key
	}
	if len(issue.Params) > 0 {
		maps.Copy(properties, issue.Params)
	}

	properties["inclusive"] = issue.Inclusive

	if issue.Message != "" {
		properties["originalMessage"] = issue.Message
	}

	return properties
}
