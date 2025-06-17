package gozod

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Pre-compiled regex for path validation
var nonWordRegex = regexp.MustCompile(`[^\w$]`)

////////////////////////    ERROR CLASS   ////////////////////////

// ZodError represents a validation error with a collection of issues
type ZodError struct {
	Type   interface{} `json:"type"`
	Issues []ZodIssue  `json:"issues"`
	Zod    struct {
		Output interface{} `json:"output"`
		Def    []ZodIssue  `json:"def"`
	} `json:"_zod"`
	Stack string `json:"stack,omitempty"`
	Name  string `json:"name"`
}

// NewZodError creates a new validation error with the given issues
func NewZodError(issues []ZodIssue) *ZodError {
	err := &ZodError{
		Type:   interface{}(nil),
		Issues: issues,
		Name:   "ZodError",
	}
	err.Zod.Output = interface{}(nil)
	err.Zod.Def = issues
	return err
}

// Error implements the error interface
func (e *ZodError) Error() string {
	return PrettifyError(e)
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

///////////////////    ERROR UTILITIES   ////////////////////////

// ZodFormattedError represents a formatted error structure
type ZodFormattedError map[string]interface{}

// FormatError formats a ZodError into a structured error object
func FormatError(error *ZodError) ZodFormattedError {
	return FormatErrorWithMapper(error, func(issue ZodIssue) string {
		return issue.Message
	})
}

// FormatErrorWithMapper formats a ZodError with custom message mapping
func FormatErrorWithMapper(error *ZodError, mapper func(ZodIssue) string) ZodFormattedError {
	fieldErrors := make(ZodFormattedError)
	fieldErrors["_errors"] = []string{}

	var processError func(*ZodError)
	processError = func(error *ZodError) {
		for _, issue := range error.Issues {
			switch issue.Code {
			case "invalid_union":
				for _, unionErrors := range issue.Errors {
					subError := &ZodError{Issues: unionErrors}
					processError(subError)
				}
			case "invalid_key":
				subError := &ZodError{Issues: issue.Issues}
				processError(subError)
			case "invalid_element":
				subError := &ZodError{Issues: issue.Issues}
				processError(subError)
			default:
				if len(issue.Path) == 0 {
					if errors, ok := fieldErrors["_errors"].([]string); ok {
						fieldErrors["_errors"] = append(errors, mapper(issue))
					}
				} else {
					curr := fieldErrors
					for i, pathEl := range issue.Path {
						key := fmt.Sprintf("%v", pathEl)
						terminal := i == len(issue.Path)-1

						if !terminal {
							if curr[key] == nil {
								curr[key] = ZodFormattedError{"_errors": []string{}}
							}
							if currMap, ok := curr[key].(ZodFormattedError); ok {
								curr = currMap
							}
						} else {
							if curr[key] == nil {
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

// ZodErrorTree represents a tree-structured error
type ZodErrorTree struct {
	Errors     []string                 `json:"errors"`
	Properties map[string]*ZodErrorTree `json:"properties,omitempty"`
	Items      [](*ZodErrorTree)        `json:"items,omitempty"`
}

// FlattenedError represents a flattened error structure for simple form validation
type FlattenedError struct {
	FormErrors  []string            `json:"formErrors"`  // Top-level errors (path is empty)
	FieldErrors map[string][]string `json:"fieldErrors"` // Field-level errors by field name
}

// TreeifyError converts a ZodError into a tree structure
func TreeifyError(error *ZodError) *ZodErrorTree {
	return TreeifyErrorWithMapper(error, func(issue ZodIssue) string {
		return issue.Message
	})
}

// TreeifyErrorWithMapper converts a ZodError into a tree structure with custom mapping
func TreeifyErrorWithMapper(error *ZodError, mapper func(ZodIssue) string) *ZodErrorTree {
	result := &ZodErrorTree{
		Errors: []string{},
	}

	var processError func(*ZodError, []interface{})
	processError = func(error *ZodError, path []interface{}) {
		for _, issue := range error.Issues {
			switch issue.Code {
			case "invalid_union":
				for _, unionErrors := range issue.Errors {
					subError := &ZodError{Issues: unionErrors}
					processError(subError, issue.Path)
				}
			case "invalid_key":
				subError := &ZodError{Issues: issue.Issues}
				processError(subError, issue.Path)
			case "invalid_element":
				subError := &ZodError{Issues: issue.Issues}
				processError(subError, issue.Path)
			default:
				fullpath := make([]interface{}, 0, len(path)+len(issue.Path))
				fullpath = append(fullpath, path...)
				fullpath = append(fullpath, issue.Path...)

				// Form-level errors (empty path)
				if len(fullpath) == 0 {
					result.Errors = append(result.Errors, mapper(issue))
					continue
				}

				curr := result
				for i, el := range fullpath {
					terminal := i == len(fullpath)-1
					if key, ok := el.(string); ok {
						if curr.Properties == nil {
							curr.Properties = make(map[string]*ZodErrorTree)
						}
						if curr.Properties[key] == nil {
							curr.Properties[key] = &ZodErrorTree{Errors: []string{}}
						}
						curr = curr.Properties[key]
					} else if idx, ok := el.(int); ok {
						if curr.Items == nil {
							curr.Items = make([]*ZodErrorTree, 0)
						}
						for len(curr.Items) <= idx {
							curr.Items = append(curr.Items, &ZodErrorTree{Errors: []string{}})
						}
						curr = curr.Items[idx]
					}

					if terminal {
						curr.Errors = append(curr.Errors, mapper(issue))
					}
				}
			}
		}
	}

	processError(error, []interface{}{})
	return result
}

// FlattenError converts a ZodError into a flat structure for simple form validation
func FlattenError(error *ZodError) *FlattenedError {
	return FlattenErrorWithMapper(error, func(issue ZodIssue) string {
		return issue.Message
	})
}

// FlattenErrorWithMapper converts a ZodError into a flat structure with custom mapping
func FlattenErrorWithMapper(error *ZodError, mapper func(ZodIssue) string) *FlattenedError {
	result := &FlattenedError{
		FormErrors:  []string{},
		FieldErrors: make(map[string][]string),
	}

	var processError func(*ZodError, []interface{})
	processError = func(error *ZodError, path []interface{}) {
		for _, issue := range error.Issues {
			switch issue.Code {
			case "invalid_union":
				for _, unionErrors := range issue.Errors {
					subError := &ZodError{Issues: unionErrors}
					processError(subError, issue.Path)
				}
			case "invalid_key":
				subError := &ZodError{Issues: issue.Issues}
				processError(subError, issue.Path)
			case "invalid_element":
				subError := &ZodError{Issues: issue.Issues}
				processError(subError, issue.Path)
			default:
				fullpath := make([]interface{}, 0, len(path)+len(issue.Path))
				fullpath = append(fullpath, path...)
				fullpath = append(fullpath, issue.Path...)

				// Form-level errors (empty path)
				if len(fullpath) == 0 {
					result.FormErrors = append(result.FormErrors, mapper(issue))
					continue
				}

				// Field-level errors - use only the first level as the key
				var fieldKey string
				if len(fullpath) > 0 {
					fieldKey = fmt.Sprintf("%v", fullpath[0])
				}

				if fieldKey != "" {
					if result.FieldErrors[fieldKey] == nil {
						result.FieldErrors[fieldKey] = []string{}
					}
					result.FieldErrors[fieldKey] = append(result.FieldErrors[fieldKey], mapper(issue))
				}
			}
		}
	}

	processError(error, []interface{}{})
	return result
}

// ToDotPath converts a path array to dot notation
func ToDotPath(path []interface{}) string {
	var segs []string
	for _, seg := range path {
		switch v := seg.(type) {
		case int:
			segs = append(segs, fmt.Sprintf("[%d]", v))
		case string:
			// Check if string contains non-word characters
			if nonWordRegex.MatchString(v) {
				jsonBytes, _ := json.Marshal(v)
				segs = append(segs, fmt.Sprintf("[%s]", string(jsonBytes)))
			} else {
				if len(segs) > 0 {
					segs = append(segs, ".")
				}
				segs = append(segs, v)
			}
		default:
			// Handle other types as strings
			jsonBytes, _ := json.Marshal(fmt.Sprintf("%v", v))
			segs = append(segs, fmt.Sprintf("[%s]", string(jsonBytes)))
		}
	}
	return strings.Join(segs, "")
}

// PrettifyError formats a ZodError as a human-readable string
func PrettifyError(error *ZodError) string {
	lines := make([]string, 0, len(error.Issues)*2) // Pre-allocate with estimated capacity

	// Sort issues by path length
	issues := make([]ZodIssue, len(error.Issues))
	copy(issues, error.Issues)
	sort.Slice(issues, func(i, j int) bool {
		return len(issues[i].Path) < len(issues[j].Path)
	})

	// Process each issue
	for _, issue := range issues {
		lines = append(lines, fmt.Sprintf("✖ %s", issue.Message))
		if len(issue.Path) > 0 {
			lines = append(lines, fmt.Sprintf("  → at %s", ToDotPath(issue.Path)))
		}
	}

	return strings.Join(lines, "\n")
}
