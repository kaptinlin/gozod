package tagparser

import (
	"slices"
	"strings"
)

// RuleOp identifies the semantic operation represented by a tag rule.
type RuleOp string

// RuleOp values describe shared tag-rule operations.
const (
	RuleStructural  RuleOp = "structural"
	RuleMethod      RuleOp = "method"
	RuleStringCheck RuleOp = "string_check"
	RuleTime        RuleOp = "time"
	RulePositive    RuleOp = "positive"
	RuleNegative    RuleOp = "negative"
	RuleFinite      RuleOp = "finite"
	RuleNonEmpty    RuleOp = "nonempty"
	RuleEnum        RuleOp = "enum"
	RuleLiteral     RuleOp = "literal"
	RuleDefault     RuleOp = "default"
	RulePrefault    RuleOp = "prefault"
	RuleUnsupported RuleOp = "unsupported"
)

// RulePlan is the shared semantic plan for a parsed tag rule.
type RulePlan struct {
	Rule   TagRule
	Op     RuleOp
	Method string
}

// OptionalPlacement describes where an Optional modifier belongs relative to
// field operations when a backend must express optionality as a fluent call.
type OptionalPlacement string

const (
	// OptionalPlacementNone means no generated Optional modifier is needed.
	OptionalPlacementNone OptionalPlacement = "none"
	// OptionalPlacementBeforeOperations places Optional before planned rule operations.
	OptionalPlacementBeforeOperations OptionalPlacement = "before_operations"
	// OptionalPlacementAfterOperations places Optional after planned rule operations.
	OptionalPlacementAfterOperations OptionalPlacement = "after_operations"
)

// FieldPlan is the shared semantic plan for a parsed struct field.
type FieldPlan struct {
	Operations             []RulePlan
	RuntimePointerOptional OptionalPlacement
	GeneratedOptional      OptionalPlacement
}

// CompileFieldPlan converts parsed field facts into semantic operations shared
// by runtime reflection and code generation.
func CompileFieldPlan(field *FieldInfo) FieldPlan {
	if field == nil {
		return FieldPlan{}
	}
	rules := field.ValidationRules()
	operations := make([]RulePlan, 0, len(rules))
	for _, rule := range rules {
		operations = append(operations, CompileRule(rule))
	}
	hasDefaultLike := hasDefaultLikeOperation(operations)

	return FieldPlan{
		Operations:             operations,
		RuntimePointerOptional: optionalPlacement(field.NeedsPointerOptional(), hasDefaultLike),
		GeneratedOptional:      optionalPlacement(field.NeedsGeneratedOptional(), hasDefaultLike),
	}
}

// OperationsExcept returns planned operations except those for the named raw rules.
func (p FieldPlan) OperationsExcept(names ...string) []RulePlan {
	if len(names) == 0 {
		return slices.Clone(p.Operations)
	}
	operations := make([]RulePlan, 0, len(p.Operations))
	for _, operation := range p.Operations {
		if !slices.Contains(names, operation.Rule.Name) {
			operations = append(operations, operation)
		}
	}
	return operations
}

// CompileRule converts a parsed tag rule into a shared semantic plan.
func CompileRule(rule TagRule) RulePlan {
	plan := RulePlan{Rule: rule, Op: RuleUnsupported}
	switch rule.Name {
	case "required", "optional", "coerce":
		plan.Op = RuleStructural
	case "nilable":
		plan.Op = RuleMethod
		plan.Method = "Nilable"
	case "min":
		plan.Op = RuleMethod
		plan.Method = "Min"
	case "max":
		plan.Op = RuleMethod
		plan.Method = "Max"
	case "length":
		plan.Op = RuleMethod
		plan.Method = "Length"
	case "gt":
		plan.Op = RuleMethod
		plan.Method = "Gt"
	case "gte":
		plan.Op = RuleMethod
		plan.Method = "Gte"
	case "lt":
		plan.Op = RuleMethod
		plan.Method = "Lt"
	case "lte":
		plan.Op = RuleMethod
		plan.Method = "Lte"
	case "regex":
		plan.Op = RuleMethod
		plan.Method = "Regex"
	case "includes":
		plan.Op = RuleMethod
		plan.Method = "Includes"
	case "startswith":
		plan.Op = RuleMethod
		plan.Method = "StartsWith"
	case "endswith":
		plan.Op = RuleMethod
		plan.Method = "EndsWith"
	case "multipleof":
		plan.Op = RuleMethod
		plan.Method = "MultipleOf"
	case "default":
		plan.Op = RuleDefault
		plan.Method = "Default"
	case "prefault":
		plan.Op = RulePrefault
		plan.Method = "Prefault"
	case "refine":
		plan.Op = RuleMethod
		plan.Method = "Refine"
	case "check":
		plan.Op = RuleMethod
		plan.Method = "Check"
	case "trim":
		plan.Op = RuleMethod
		plan.Method = "Trim"
	case "lowercase":
		plan.Op = RuleMethod
		plan.Method = "ToLowerCase"
	case "uppercase":
		plan.Op = RuleMethod
		plan.Method = "ToUpperCase"
	case "email":
		plan.Op = RuleStringCheck
		plan.Method = "Email"
	case "url":
		plan.Op = RuleStringCheck
		plan.Method = "URL"
	case "uuid":
		plan.Op = RuleStringCheck
		plan.Method = "UUID"
	case "ipv4":
		plan.Op = RuleStringCheck
		plan.Method = "IPv4"
	case "ipv6":
		plan.Op = RuleStringCheck
		plan.Method = "IPv6"
	case "cidrv4":
		plan.Op = RuleStringCheck
		plan.Method = "CIDRv4"
	case "cidrv6":
		plan.Op = RuleStringCheck
		plan.Method = "CIDRv6"
	case "cuid":
		plan.Op = RuleStringCheck
		plan.Method = "CUID"
	case "cuid2":
		plan.Op = RuleStringCheck
		plan.Method = "CUID2"
	case "jwt":
		plan.Op = RuleStringCheck
		plan.Method = "JWT"
	case "iso_datetime":
		plan.Op = RuleStringCheck
		plan.Method = "IsoDateTime"
	case "iso_date":
		plan.Op = RuleStringCheck
		plan.Method = "IsoDate"
	case "iso_time":
		plan.Op = RuleStringCheck
		plan.Method = "IsoTime"
	case "iso_duration":
		plan.Op = RuleStringCheck
		plan.Method = "IsoDuration"
	case "time":
		plan.Op = RuleTime
	case "positive":
		plan.Op = RulePositive
	case "negative":
		plan.Op = RuleNegative
	case "finite":
		plan.Op = RuleFinite
	case "nonempty":
		plan.Op = RuleNonEmpty
	case "enum":
		plan.Op = RuleEnum
	case "literal":
		plan.Op = RuleLiteral
	}
	return plan
}

func hasDefaultLikeOperation(operations []RulePlan) bool {
	for _, operation := range operations {
		if operation.Op == RuleDefault || operation.Op == RulePrefault {
			return true
		}
	}
	return false
}

func optionalPlacement(needsOptional, hasDefaultLike bool) OptionalPlacement {
	if !needsOptional {
		return OptionalPlacementNone
	}
	if hasDefaultLike {
		return OptionalPlacementBeforeOperations
	}
	return OptionalPlacementAfterOperations
}

// FirstParam returns the first parameter for rules that take one argument.
func (p RulePlan) FirstParam() (string, bool) {
	if len(p.Rule.Params) == 0 {
		return "", false
	}
	return p.Rule.Params[0], true
}

// JoinedValue returns a single value for default-like rules.
func (p RulePlan) JoinedValue() (string, bool) {
	if len(p.Rule.Params) == 0 {
		return "", false
	}
	value := p.Rule.Params[0]
	if len(p.Rule.Params) > 1 && !strings.HasPrefix(value, "[") && !strings.HasPrefix(value, "{") {
		value = strings.Join(p.Rule.Params, " ")
	}
	return value, true
}
