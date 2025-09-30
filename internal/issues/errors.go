package issues

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/slicex"
	"github.com/kaptinlin/gozod/pkg/structx"
)

// Pre-compiled regex for path validation and formatting
var nonWordRegex = regexp.MustCompile(`[^\w$]`)

// =============================================================================
// ZOD ERROR CLASS
// =============================================================================

// ZodError represents a validation error with a collection of issues
// Implements TypeScript Zod v4 compatible error structure and behavior
type ZodError struct {
	Type   any        `json:"type"`   // The expected type that failed validation
	Issues []ZodIssue `json:"issues"` // Collection of validation issues
	Zod    struct {   // Internal Zod metadata
		Output any        `json:"output"` // The output value (if any)
		Def    []ZodIssue `json:"def"`    // Issue definitions
	} `json:"_zod"`
	Stack string `json:"stack,omitempty"` // Stack trace for debugging
	Name  string `json:"name"`            // Error name identifier

	// Internal formatter for customizing error messages
	formatter MessageFormatter `json:"-"`
}

// NewZodError creates a new validation error with the given issues
// Uses the default message formatter for error formatting
// Uses Go 1.22+ slices.Clone for efficient issue copying
func NewZodError(issues []ZodIssue) *ZodError {
	// Clone issues slice to avoid mutation of original
	issueCopy := slices.Clone(issues)

	err := &ZodError{
		Type:      any(nil),
		Issues:    issueCopy,
		Name:      "ZodError",
		formatter: defaultFormatter,
	}
	err.Zod.Output = any(nil)
	err.Zod.Def = issueCopy
	return err
}

// NewZodErrorWithFormatter creates a new validation error with a custom formatter
// Allows for localized or customized error message generation
func NewZodErrorWithFormatter(issues []ZodIssue, formatter MessageFormatter) *ZodError {
	err := NewZodError(issues)
	if formatter != nil {
		err.formatter = formatter
	}
	return err
}

// Error implements the error interface using the configured formatter
// Returns a prettified string representation of all validation issues
func (e *ZodError) Error() string {
	if e == nil {
		return ""
	}

	if slicex.IsEmpty(e.Issues) {
		return "Validation failed"
	}

	return PrettifyErrorWithFormatter(e, e.formatter)
}

// GetFormatter returns the current message formatter
func (e *ZodError) GetFormatter() MessageFormatter {
	return e.formatter
}

// SetFormatter sets a new message formatter for this error
// Allows dynamic switching of error message formats
func (e *ZodError) SetFormatter(formatter MessageFormatter) {
	if formatter != nil {
		e.formatter = formatter
	}
}

// IsZodError checks if an error is a ZodError and extracts it
// This function provides similar functionality to errors.As for ZodError specifically
func IsZodError(err error, target **ZodError) bool {
	if err == nil {
		return false
	}

	// Use errors.As to unwrap and check for ZodError
	var zodErr *ZodError
	if errors.As(err, &zodErr) {
		if target != nil {
			*target = zodErr
		}
		return true
	}

	return false
}

// =============================================================================
// ERROR FORMATTING UTILITIES
// =============================================================================

// ZodFormattedError represents a formatted error structure following TypeScript patterns
// Provides hierarchical error representation with field-level error grouping
type ZodFormattedError map[string]any

// FormatError formats a ZodError into a structured error object
// Creates a hierarchical representation matching TypeScript Zod v4 format
func FormatError(error *ZodError) ZodFormattedError {
	return FormatErrorWithMapper(error, func(issue ZodIssue) string {
		// Use existing message if available, otherwise format using formatter
		if issue.Message != "" {
			return issue.Message
		}

		return error.formatter.FormatMessage(core.ZodRawIssue{
			Code:       issue.Code,
			Path:       issue.Path,
			Message:    issue.Message,
			Properties: convertZodIssueToProperties(issue),
		})
	})
}

// FormatErrorWithMapper formats a ZodError with custom message mapping
// Allows for custom message generation while maintaining structure
func FormatErrorWithMapper(error *ZodError, mapper func(ZodIssue) string) ZodFormattedError {
	fieldErrors := make(ZodFormattedError)
	fieldErrors["_errors"] = []string{}

	var processError func(*ZodError)
	processError = func(error *ZodError) {
		for _, issue := range error.Issues {
			switch issue.Code {
			case core.InvalidUnion:
				// Use slicex to process union errors
				if !slicex.IsEmpty(issue.Errors) {
					for _, unionErrors := range issue.Errors {
						if !slicex.IsEmpty(unionErrors) {
							subError := &ZodError{Issues: unionErrors, formatter: error.formatter}
							processError(subError)
						}
					}
				}
			case core.InvalidKey, core.InvalidElement:
				// Use slicex to check if there are nested issues
				if !slicex.IsEmpty(issue.Issues) {
					subError := &ZodError{Issues: issue.Issues, formatter: error.formatter}
					processError(subError)
				}
			case core.InvalidType, core.InvalidValue, core.InvalidFormat,
				core.TooBig, core.TooSmall, core.NotMultipleOf,
				core.UnrecognizedKeys, core.Custom, core.InvalidSchema,
				core.InvalidDiscriminator, core.IncompatibleTypes, core.MissingRequired,
				core.TypeConversion, core.NilPointer:
				// Handle regular issues with path-based organization
				if slicex.IsEmpty(issue.Path) {
					// Root-level errors go into _errors array
					if errors, ok := fieldErrors["_errors"].([]string); ok {
						fieldErrors["_errors"] = append(errors, mapper(issue))
					}
				} else {
					// Build nested structure following the path using mapx
					curr := fieldErrors
					for i, pathEl := range issue.Path {
						key := fmt.Sprintf("%v", pathEl)
						terminal := i == len(issue.Path)-1

						if !terminal {
							// Use mapx to safely check and create structure
							if !mapx.Has(curr, key) {
								curr[key] = ZodFormattedError{"_errors": []string{}}
							}
							if currMap, ok := curr[key].(ZodFormattedError); ok {
								curr = currMap
							}
						} else {
							// Terminal path element - add the error message
							if !mapx.Has(curr, key) {
								curr[key] = ZodFormattedError{"_errors": []string{}}
							}
							if currMap, ok := curr[key].(ZodFormattedError); ok {
								if errors, ok := currMap["_errors"].([]string); ok {
									currMap["_errors"] = append(errors, mapper(issue))
								}
							}
						}
					}
				}
			}
		}
	}

	processError(error)
	return fieldErrors
}

// =============================================================================
// TREE AND FLAT ERROR STRUCTURES
// =============================================================================

// ZodErrorTree represents a tree-structured error following TypeScript Zod patterns
// Provides a hierarchical view of validation errors for complex data structures
type ZodErrorTree struct {
	Errors     []string                 `json:"errors"`               // Errors at this level
	Properties map[string]*ZodErrorTree `json:"properties,omitempty"` // Object property errors
	Items      [](*ZodErrorTree)        `json:"items,omitempty"`      // Array/slice item errors
}

// FlattenedError represents a flattened error structure for simple form validation
// Separates top-level form errors from field-specific errors
type FlattenedError struct {
	FormErrors  []string            `json:"formErrors"`  // Top-level errors (path is empty)
	FieldErrors map[string][]string `json:"fieldErrors"` // Field-level errors by field name
}

// TreeifyError formats a ZodError into a tree structure
// Provides hierarchical error representation for complex data structures
func TreeifyError(error *ZodError) *ZodErrorTree {
	return TreeifyErrorWithMapper(error, func(issue ZodIssue) string {
		// Use existing message if available, otherwise format using formatter
		if issue.Message != "" {
			return issue.Message
		}

		return error.formatter.FormatMessage(core.ZodRawIssue{
			Code:       issue.Code,
			Path:       issue.Path,
			Message:    issue.Message,
			Properties: convertZodIssueToProperties(issue),
		})
	})
}

// TreeifyErrorWithMapper converts a ZodError into a tree structure with custom message mapping
// Uses precise pre-allocation based on issue count for optimal performance
func TreeifyErrorWithMapper(error *ZodError, mapper func(ZodIssue) string) *ZodErrorTree {
	// Pre-allocate based on issue count to reduce allocations
	issueCount := len(error.Issues)
	tree := &ZodErrorTree{
		Errors:     make([]string, 0, max(issueCount/4, 2)), // Conservative estimate
		Properties: make(map[string]*ZodErrorTree, max(issueCount/2, 4)),
		Items:      make([](*ZodErrorTree), 0, max(issueCount/4, 2)),
	}

	for _, issue := range error.Issues {
		processIssueInTree(issue, tree, mapper)
	}

	return tree
}

// processIssueInTree processes an issue within a specific tree node
func processIssueInTree(issue ZodIssue, tree *ZodErrorTree, mapper func(ZodIssue) string) {
	// Use slicex to check if path is empty
	if slicex.IsEmpty(issue.Path) {
		tree.Errors = append(tree.Errors, mapper(issue))
		return
	}

	// Process path using modern Go patterns
	current := tree
	for i, pathElement := range issue.Path {
		isLast := i == len(issue.Path)-1

		switch element := pathElement.(type) {
		case string:
			// Object property access
			if current.Properties == nil {
				current.Properties = make(map[string]*ZodErrorTree)
			}
			if current.Properties[element] == nil {
				current.Properties[element] = &ZodErrorTree{
					Errors:     []string{},
					Properties: make(map[string]*ZodErrorTree),
					Items:      [](*ZodErrorTree){},
				}
			}
			current = current.Properties[element]

		case int:
			// Array/slice index access - use slicex for safer operations
			for len(current.Items) <= element {
				current.Items = append(current.Items, &ZodErrorTree{
					Errors:     []string{},
					Properties: make(map[string]*ZodErrorTree),
					Items:      [](*ZodErrorTree){},
				})
			}
			current = current.Items[element]
		}

		if isLast {
			current.Errors = append(current.Errors, mapper(issue))
		}
	}
}

// FlattenError flattens a ZodError into form and field errors
// Separates top-level form errors from field-specific errors
func FlattenError(error *ZodError) *FlattenedError {
	return FlattenErrorWithMapper(error, func(issue ZodIssue) string {
		// Use existing message if available, otherwise format using formatter
		if issue.Message != "" {
			return issue.Message
		}

		return error.formatter.FormatMessage(core.ZodRawIssue{
			Code:       issue.Code,
			Path:       issue.Path,
			Message:    issue.Message,
			Properties: convertZodIssueToProperties(issue),
		})
	})
}

// FlattenErrorWithMapper flattens a ZodError into form and field errors with custom message mapping
// Uses precise pre-allocation and efficient batch processing
func FlattenErrorWithMapper(error *ZodError, mapper func(ZodIssue) string) *FlattenedError {
	issueCount := len(error.Issues)
	flattened := &FlattenedError{
		FormErrors:  make([]string, 0, max(issueCount/4, 2)),         // Conservative estimate
		FieldErrors: make(map[string][]string, max(issueCount/2, 4)), // More fields expected
	}

	for _, issue := range error.Issues {
		message := mapper(issue)

		if slicex.IsEmpty(issue.Path) {
			// Top-level error
			flattened.FormErrors = append(flattened.FormErrors, message)
		} else {
			// Field-level error - use only the first level of the path as key
			var fieldPath string
			if len(issue.Path) > 0 {
				fieldPath = fmt.Sprintf("%v", issue.Path[0])
			}

			if flattened.FieldErrors[fieldPath] == nil {
				// Pre-allocate slice for this field
				flattened.FieldErrors[fieldPath] = make([]string, 0, 2)
			}
			flattened.FieldErrors[fieldPath] = append(flattened.FieldErrors[fieldPath], message)
		}
	}

	return flattened
}

// FlattenErrorWithFormatter flattens a ZodError into form and field errors with custom formatter
func FlattenErrorWithFormatter(error *ZodError, formatter MessageFormatter) *FlattenedError {
	return FlattenErrorWithMapper(error, func(issue ZodIssue) string {
		// Use existing message if available, otherwise format using custom formatter
		if issue.Message != "" {
			return issue.Message
		}

		return formatter.FormatMessage(core.ZodRawIssue{
			Code:       issue.Code,
			Path:       issue.Path,
			Message:    issue.Message,
			Properties: convertZodIssueToProperties(issue),
		})
	})
}

// =============================================================================
// PATH AND STRING UTILITIES
// =============================================================================

// ToDotPath converts a path array to dot notation string
func ToDotPath(path []any) string {
	if len(path) == 0 {
		return ""
	}

	// Use strings.Builder to construct path string efficiently
	var builder strings.Builder
	builder.Grow(len(path) * 8) // Rough estimate

	for i, element := range path {
		if i > 0 {
			builder.WriteByte('.')
		}

		switch el := element.(type) {
		case string:
			if nonWordRegex.MatchString(el) {
				builder.WriteByte('[')
				builder.WriteByte('"')
				builder.WriteString(el)
				builder.WriteByte('"')
				builder.WriteByte(']')
			} else {
				builder.WriteString(el)
			}
		case int:
			builder.WriteByte('[')
			builder.WriteString(strconv.Itoa(el))
			builder.WriteByte(']')
		default:
			builder.WriteByte('[')
			builder.WriteString(fmt.Sprintf("%v", el))
			builder.WriteByte(']')
		}
	}

	return builder.String()
}

// PrettifyError formats a ZodError into a readable string using its formatter
func PrettifyError(error *ZodError) string {
	return PrettifyErrorWithFormatter(error, error.formatter)
}

// PrettifyErrorWithFormatter formats a ZodError into a readable string with custom formatter
func PrettifyErrorWithFormatter(error *ZodError, formatter MessageFormatter) string {
	if error == nil || len(error.Issues) == 0 {
		return "Validation failed"
	}

	// Use strings.Builder to avoid repeated string allocations
	var builder strings.Builder
	// Pre-allocate capacity based on estimated message lengths
	builder.Grow(len(error.Issues) * 50) // Rough estimate

	for i, issue := range error.Issues {
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

		// Format with path if present
		if len(issue.Path) > 0 {
			pathStr := ToDotPath(issue.Path)
			builder.WriteString(pathStr)
			builder.WriteString(": ")
		}
		builder.WriteString(message)
	}

	return builder.String()
}

// =============================================================================
// UTILITY FUNCTIONS
// =============================================================================

// convertZodIssueToProperties converts a ZodIssue to properties map for formatter
// Ensures all issue data is available to the message formatter
func convertZodIssueToProperties(issue ZodIssue) map[string]any {
	// Use structx to extract issue properties
	if properties, err := structx.ToMap(issue); err == nil {
		// Use mapx to safely handle the properties
		result := mapx.Copy(properties)

		// Remove fields that are handled separately
		delete(result, "Code")
		delete(result, "Path")
		delete(result, "Message")
		delete(result, "Continue")
		delete(result, "Inst")
		delete(result, "Issues")
		delete(result, "Errors")

		return result
	}

	// Fallback: create properties map manually
	properties := make(map[string]any)

	// Use mapx to safely set properties
	if issue.Expected != "" {
		mapx.Set(properties, "expected", issue.Expected)
	}
	if issue.Received != "" {
		mapx.Set(properties, "received", issue.Received)
	}
	if issue.Minimum != nil {
		mapx.Set(properties, "minimum", issue.Minimum)
	}
	if issue.Maximum != nil {
		mapx.Set(properties, "maximum", issue.Maximum)
	}
	if issue.Format != "" {
		mapx.Set(properties, "format", issue.Format)
	}
	if issue.Pattern != "" {
		mapx.Set(properties, "pattern", issue.Pattern)
	}
	if issue.Prefix != "" {
		mapx.Set(properties, "startsWith", issue.Prefix)
	}
	if issue.Suffix != "" {
		mapx.Set(properties, "endsWith", issue.Suffix)
	}
	if issue.Includes != "" {
		mapx.Set(properties, "includes", issue.Includes)
	}
	if issue.Algorithm != "" {
		mapx.Set(properties, "algorithm", issue.Algorithm)
	}
	if issue.Divisor != nil {
		mapx.Set(properties, "divisor", issue.Divisor)
	}
	if !slicex.IsEmpty(issue.Keys) {
		mapx.Set(properties, "keys", issue.Keys)
	}
	if !slicex.IsEmpty(issue.Values) {
		mapx.Set(properties, "values", issue.Values)
	}
	if issue.Origin != "" {
		mapx.Set(properties, "origin", issue.Origin)
	}
	if issue.Key != nil {
		mapx.Set(properties, "key", issue.Key)
	}
	if len(issue.Params) > 0 {
		for k, v := range issue.Params {
			mapx.Set(properties, k, v)
		}
	}

	// Add inclusive flag
	mapx.Set(properties, "inclusive", issue.Inclusive)

	// Add issue-specific properties that might be useful for formatting
	if issue.Message != "" {
		mapx.Set(properties, "originalMessage", issue.Message)
	}

	return properties
}
