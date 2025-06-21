package issues

import (
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/pkg/mapx"
)

// =============================================================================
// RAW ISSUE CONSTRUCTOR FUNCTIONS
// =============================================================================

// NewRawIssue creates a new raw issue with options pattern
func NewRawIssue(code core.IssueCode, input any, options ...func(*core.ZodRawIssue)) core.ZodRawIssue {
	issue := core.ZodRawIssue{
		Code:  code,
		Input: input,
		Path:  []any{},
		// Properties will be initialized lazily when needed
	}

	for _, option := range options {
		option(&issue)
	}

	return issue
}

// NewRawIssueFromMessage creates a ZodRawIssue with custom message
func NewRawIssueFromMessage(message string, input any, inst any) core.ZodRawIssue {
	return core.ZodRawIssue{
		Code:    core.Custom,
		Message: message,
		Input:   input,
		Inst:    inst,
		Path:    []any{},
		// Properties will be initialized lazily when needed
	}
}

// =============================================================================
// OPTION FUNCTIONS
// =============================================================================

// WithOrigin sets the origin field
func WithOrigin(origin string) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		mapx.Set(issue.Properties, "origin", origin)
	}
}

// WithMessage sets the message field
func WithMessage(message string) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		issue.Message = message
	}
}

// WithMinimum sets the minimum field
func WithMinimum(minimum any) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		mapx.Set(issue.Properties, "minimum", minimum)
	}
}

// WithMaximum sets the maximum field
func WithMaximum(maximum any) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		mapx.Set(issue.Properties, "maximum", maximum)
	}
}

// WithExpected sets the expected field
func WithExpected(expected string) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		mapx.Set(issue.Properties, "expected", expected)
	}
}

// WithReceived sets the received field
func WithReceived(received string) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		mapx.Set(issue.Properties, "received", received)
	}
}

// WithPath sets the path field
func WithPath(path []any) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		issue.Path = path
	}
}

// WithInclusive sets the inclusive field
func WithInclusive(inclusive bool) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		mapx.Set(issue.Properties, "inclusive", inclusive)
	}
}

// WithFormat sets the format field
func WithFormat(format string) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		mapx.Set(issue.Properties, "format", format)
	}
}

// WithContinue sets the continue field
func WithContinue(cont bool) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		issue.Continue = cont
	}
}

// WithPattern sets the pattern field
func WithPattern(pattern string) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		mapx.Set(issue.Properties, "pattern", pattern)
	}
}

// WithPrefix sets the prefix field
func WithPrefix(prefix string) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		mapx.Set(issue.Properties, "prefix", prefix)
	}
}

// WithSuffix sets the suffix field
func WithSuffix(suffix string) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		mapx.Set(issue.Properties, "suffix", suffix)
	}
}

// WithIncludes sets the includes field
func WithIncludes(includes string) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		mapx.Set(issue.Properties, "includes", includes)
	}
}

// WithDivisor sets the divisor field
func WithDivisor(divisor any) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		mapx.Set(issue.Properties, "divisor", divisor)
	}
}

// WithKeys sets the keys field
func WithKeys(keys []string) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		mapx.Set(issue.Properties, "keys", keys)
	}
}

// WithValues sets the values field
func WithValues(values []any) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		mapx.Set(issue.Properties, "values", values)
	}
}

// WithAlgorithm sets the algorithm field
func WithAlgorithm(algorithm string) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		mapx.Set(issue.Properties, "algorithm", algorithm)
	}
}

// WithParams sets the params field
func WithParams(params map[string]any) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		mapx.Set(issue.Properties, "params", params)
	}
}

// WithInst sets the inst field
func WithInst(inst any) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		issue.Inst = inst
	}
}

// WithProperty sets a generic property
func WithProperty(key string, value any) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		mapx.Set(issue.Properties, key, value)
	}
}

// WithProperties merges multiple properties
func WithProperties(properties map[string]any) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		issue.Properties = mapx.Merge(issue.Properties, properties)
	}
}
