package issues

import (
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/pkg/mapx"
)

// NewRawIssue creates a new raw issue with functional options.
func NewRawIssue(code core.IssueCode, input any, options ...func(*core.ZodRawIssue)) core.ZodRawIssue {
	issue := core.ZodRawIssue{
		Code:  code,
		Input: input,
		Path:  []any{},
	}

	for _, option := range options {
		option(&issue)
	}

	return issue
}

// NewRawIssueFromMessage creates a ZodRawIssue with a custom message.
func NewRawIssueFromMessage(message string, input any, inst any) core.ZodRawIssue {
	return core.ZodRawIssue{
		Code:    core.Custom,
		Message: message,
		Input:   input,
		Inst:    inst,
		Path:    []any{},
	}
}

func WithOrigin(origin string) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		issue.Properties["origin"] = origin
	}
}

func WithMessage(message string) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		issue.Message = message
	}
}

func WithMinimum(minimum any) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		issue.Properties["minimum"] = minimum
	}
}

func WithMaximum(maximum any) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		issue.Properties["maximum"] = maximum
	}
}

func WithExpected(expected string) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		issue.Properties["expected"] = expected
	}
}

func WithReceived(received string) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		issue.Properties["received"] = received
	}
}

func WithPath(path []any) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		issue.Path = path
	}
}

func WithInclusive(inclusive bool) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		issue.Properties["inclusive"] = inclusive
	}
}

func WithFormat(format string) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		issue.Properties["format"] = format
	}
}

func WithContinue(cont bool) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		issue.Continue = cont
	}
}

func WithPattern(pattern string) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		issue.Properties["pattern"] = pattern
	}
}

func WithPrefix(prefix string) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		issue.Properties["prefix"] = prefix
	}
}

func WithSuffix(suffix string) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		issue.Properties["suffix"] = suffix
	}
}

func WithIncludes(includes string) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		issue.Properties["includes"] = includes
	}
}

func WithDivisor(divisor any) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		issue.Properties["divisor"] = divisor
	}
}

func WithKeys(keys []string) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		issue.Properties["keys"] = keys
	}
}

func WithValues(values []any) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		issue.Properties["values"] = values
	}
}

func WithAlgorithm(algorithm string) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		issue.Properties["algorithm"] = algorithm
	}
}

func WithParams(params map[string]any) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		issue.Properties["params"] = params
	}
}

func WithInst(inst any) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		issue.Inst = inst
	}
}

// WithProperty sets a single property by key.
func WithProperty(key string, value any) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		issue.Properties[key] = value
	}
}

// WithProperties merges multiple properties into the issue.
func WithProperties(properties map[string]any) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		issue.Properties = mapx.Merge(issue.Properties, properties)
	}
}
