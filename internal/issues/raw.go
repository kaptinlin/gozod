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

// WithOrigin sets the origin property.
func WithOrigin(origin string) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		issue.Properties["origin"] = origin
	}
}

// WithMessage sets the message field.
func WithMessage(message string) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		issue.Message = message
	}
}

// WithMinimum sets the minimum property.
func WithMinimum(minimum any) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		issue.Properties["minimum"] = minimum
	}
}

// WithMaximum sets the maximum property.
func WithMaximum(maximum any) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		issue.Properties["maximum"] = maximum
	}
}

// WithExpected sets the expected property.
func WithExpected(expected string) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		issue.Properties["expected"] = expected
	}
}

// WithReceived sets the received property.
func WithReceived(received string) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		issue.Properties["received"] = received
	}
}

// WithPath sets the path field.
func WithPath(path []any) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		issue.Path = path
	}
}

// WithInclusive sets the inclusive property.
func WithInclusive(inclusive bool) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		issue.Properties["inclusive"] = inclusive
	}
}

// WithFormat sets the format property.
func WithFormat(format string) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		issue.Properties["format"] = format
	}
}

// WithContinue sets the continue field.
func WithContinue(cont bool) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		issue.Continue = cont
	}
}

// WithPattern sets the pattern property.
func WithPattern(pattern string) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		issue.Properties["pattern"] = pattern
	}
}

// WithPrefix sets the prefix property.
func WithPrefix(prefix string) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		issue.Properties["prefix"] = prefix
	}
}

// WithSuffix sets the suffix property.
func WithSuffix(suffix string) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		issue.Properties["suffix"] = suffix
	}
}

// WithIncludes sets the includes property.
func WithIncludes(includes string) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		issue.Properties["includes"] = includes
	}
}

// WithDivisor sets the divisor property.
func WithDivisor(divisor any) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		issue.Properties["divisor"] = divisor
	}
}

// WithKeys sets the keys property.
func WithKeys(keys []string) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		issue.Properties["keys"] = keys
	}
}

// WithValues sets the values property.
func WithValues(values []any) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		issue.Properties["values"] = values
	}
}

// WithAlgorithm sets the algorithm property.
func WithAlgorithm(algorithm string) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		issue.Properties["algorithm"] = algorithm
	}
}

// WithParams sets the params property.
func WithParams(params map[string]any) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		issue.Properties["params"] = params
	}
}

// WithInst sets the inst field.
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
