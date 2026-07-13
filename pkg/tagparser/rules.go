package tagparser

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/go-json-experiment/json"
)

// ErrUnknownRule indicates that a tag names no supported semantic operation.
var ErrUnknownRule = errors.New("tagparser: unknown rule")

// ErrMissingOperand indicates that a rule omitted its required value.
var ErrMissingOperand = errors.New("tagparser: missing rule operand")

// ErrInvalidArity indicates that a rule received too many operands.
var ErrInvalidArity = errors.New("tagparser: invalid rule arity")

// ErrInvalidOperand indicates that a rule value cannot be parsed.
var ErrInvalidOperand = errors.New("tagparser: invalid rule operand")

// ErrInapplicableRule indicates that a rule has no meaning for the field type.
var ErrInapplicableRule = errors.New("tagparser: rule is not applicable to field type")

// CompileError reports why a field's parsed tag cannot become an executable plan.
type CompileError struct {
	Field string
	Rule  string
	Err   error
}

func (e *CompileError) Error() string {
	return fmt.Sprintf("tagparser: field %s: rule %q: %v", e.Field, e.Rule, e.Err)
}

func (e *CompileError) Unwrap() error {
	return e.Err
}

// RuleOp identifies the semantic operation represented by a tag rule.
type RuleOp string

// FieldFamily identifies the schema family a compiled operation targets.
type FieldFamily string

// FieldFamily values classify the Go type family targeted by a tag rule.
const (
	FieldFamilyUnknown         FieldFamily = "unknown"
	FieldFamilyString          FieldFamily = "string"
	FieldFamilySignedInteger   FieldFamily = "signed_integer"
	FieldFamilyUnsignedInteger FieldFamily = "unsigned_integer"
	FieldFamilyFloat           FieldFamily = "float"
	FieldFamilyBool            FieldFamily = "bool"
	FieldFamilySlice           FieldFamily = "slice"
	FieldFamilyArray           FieldFamily = "array"
	FieldFamilyMap             FieldFamily = "map"
	FieldFamilyStruct          FieldFamily = "struct"
	FieldFamilyTime            FieldFamily = "time"
)

// RuleOp values identify executable tag operations.
const (
	RuleRequired    RuleOp = "required"
	RuleOptional    RuleOp = "optional"
	RuleCoerce      RuleOp = "coerce"
	RuleNilable     RuleOp = "nilable"
	RuleMin         RuleOp = "min"
	RuleMax         RuleOp = "max"
	RuleLength      RuleOp = "length"
	RuleGT          RuleOp = "gt"
	RuleGTE         RuleOp = "gte"
	RuleLT          RuleOp = "lt"
	RuleLTE         RuleOp = "lte"
	RuleRegex       RuleOp = "regex"
	RuleIncludes    RuleOp = "includes"
	RuleStartsWith  RuleOp = "startswith"
	RuleEndsWith    RuleOp = "endswith"
	RuleMultipleOf  RuleOp = "multipleof"
	RuleDefault     RuleOp = "default"
	RulePrefault    RuleOp = "prefault"
	RuleTrim        RuleOp = "trim"
	RuleLowercase   RuleOp = "lowercase"
	RuleUppercase   RuleOp = "uppercase"
	RuleEmail       RuleOp = "email"
	RuleURL         RuleOp = "url"
	RuleUUID        RuleOp = "uuid"
	RuleIPv4        RuleOp = "ipv4"
	RuleIPv6        RuleOp = "ipv6"
	RuleCIDRv4      RuleOp = "cidrv4"
	RuleCIDRv6      RuleOp = "cidrv6"
	RuleCUID        RuleOp = "cuid"
	RuleCUID2       RuleOp = "cuid2"
	RuleJWT         RuleOp = "jwt"
	RuleISODateTime RuleOp = "iso_datetime"
	RuleISODate     RuleOp = "iso_date"
	RuleISOTime     RuleOp = "iso_time"
	RuleISODuration RuleOp = "iso_duration"
	RuleTime        RuleOp = "time"
	RulePositive    RuleOp = "positive"
	RuleNegative    RuleOp = "negative"
	RuleFinite      RuleOp = "finite"
	RuleNonEmpty    RuleOp = "nonempty"
	RuleEnum        RuleOp = "enum"
	RuleLiteral     RuleOp = "literal"
)

// RulePlan is the shared semantic plan for a parsed tag rule.
type RulePlan struct {
	Name    string
	Op      RuleOp
	Family  FieldFamily
	Operand any
}

type ruleDefinition struct {
	op       RuleOp
	minArgs  int
	maxArgs  int
	families []FieldFamily
}

var ruleDefinitions = map[string]ruleDefinition{
	"required":     {op: RuleRequired, maxArgs: 0},
	"optional":     {op: RuleOptional, maxArgs: 0},
	"coerce":       {op: RuleCoerce, maxArgs: 0, families: scalarAndTimeFamilies},
	"nilable":      {op: RuleNilable, maxArgs: 0},
	"min":          unaryRule(RuleMin, lengthAndNumericFamilies),
	"max":          unaryRule(RuleMax, lengthAndNumericFamilies),
	"length":       unaryRule(RuleLength, sequenceFamilies),
	"gt":           unaryRule(RuleGT, numericFamilies),
	"gte":          unaryRule(RuleGTE, numericFamilies),
	"lt":           unaryRule(RuleLT, numericFamilies),
	"lte":          unaryRule(RuleLTE, numericFamilies),
	"regex":        unaryRule(RuleRegex, stringFamilies),
	"includes":     unaryRule(RuleIncludes, stringFamilies),
	"startswith":   unaryRule(RuleStartsWith, stringFamilies),
	"endswith":     unaryRule(RuleEndsWith, stringFamilies),
	"multipleof":   unaryRule(RuleMultipleOf, numericFamilies),
	"default":      variadicRule(RuleDefault, valueFamilies),
	"prefault":     variadicRule(RulePrefault, valueFamilies),
	"trim":         noArgRule(RuleTrim, stringFamilies),
	"lowercase":    noArgRule(RuleLowercase, stringFamilies),
	"uppercase":    noArgRule(RuleUppercase, stringFamilies),
	"email":        noArgRule(RuleEmail, stringFamilies),
	"url":          noArgRule(RuleURL, stringFamilies),
	"uuid":         noArgRule(RuleUUID, stringFamilies),
	"ipv4":         noArgRule(RuleIPv4, stringFamilies),
	"ipv6":         noArgRule(RuleIPv6, stringFamilies),
	"cidrv4":       noArgRule(RuleCIDRv4, stringFamilies),
	"cidrv6":       noArgRule(RuleCIDRv6, stringFamilies),
	"cuid":         noArgRule(RuleCUID, stringFamilies),
	"cuid2":        noArgRule(RuleCUID2, stringFamilies),
	"jwt":          noArgRule(RuleJWT, stringFamilies),
	"iso_datetime": noArgRule(RuleISODateTime, stringFamilies),
	"iso_date":     noArgRule(RuleISODate, stringFamilies),
	"iso_time":     noArgRule(RuleISOTime, stringFamilies),
	"iso_duration": noArgRule(RuleISODuration, stringFamilies),
	"time":         noArgRule(RuleTime, []FieldFamily{FieldFamilyTime}),
	"positive":     noArgRule(RulePositive, numericFamilies),
	"negative":     noArgRule(RuleNegative, numericFamilies),
	"finite":       noArgRule(RuleFinite, []FieldFamily{FieldFamilyFloat}),
	"nonempty":     noArgRule(RuleNonEmpty, sequenceFamilies),
	"enum":         variadicRule(RuleEnum, []FieldFamily{FieldFamilyString, FieldFamilySignedInteger}),
	"literal":      unaryRule(RuleLiteral, scalarFamilies),
}

var (
	stringFamilies           = []FieldFamily{FieldFamilyString}
	numericFamilies          = []FieldFamily{FieldFamilySignedInteger, FieldFamilyUnsignedInteger, FieldFamilyFloat}
	scalarFamilies           = []FieldFamily{FieldFamilyString, FieldFamilySignedInteger, FieldFamilyUnsignedInteger, FieldFamilyFloat, FieldFamilyBool}
	scalarAndTimeFamilies    = append(slices.Clone(scalarFamilies), FieldFamilyTime)
	sequenceFamilies         = []FieldFamily{FieldFamilyString, FieldFamilySlice, FieldFamilyArray, FieldFamilyMap}
	lengthAndNumericFamilies = append(slices.Clone(sequenceFamilies), numericFamilies...)
	valueFamilies            = []FieldFamily{FieldFamilyString, FieldFamilySignedInteger, FieldFamilyUnsignedInteger, FieldFamilyFloat, FieldFamilyBool, FieldFamilySlice, FieldFamilyArray, FieldFamilyMap, FieldFamilyTime}
)

func noArgRule(op RuleOp, families []FieldFamily) ruleDefinition {
	return ruleDefinition{op: op, maxArgs: 0, families: families}
}

func unaryRule(op RuleOp, families []FieldFamily) ruleDefinition {
	return ruleDefinition{op: op, minArgs: 1, maxArgs: 1, families: families}
}

func variadicRule(op RuleOp, families []FieldFamily) ruleDefinition {
	return ruleDefinition{op: op, minArgs: 1, maxArgs: -1, families: families}
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
func CompileFieldPlan(field *FieldInfo) (FieldPlan, error) {
	if field == nil {
		return FieldPlan{}, nil
	}
	rules := field.Rules
	operations := make([]RulePlan, 0, len(rules))
	for _, rule := range rules {
		definition, ok := ruleDefinitions[rule.Name]
		if !ok {
			return FieldPlan{}, &CompileError{
				Field: field.Name,
				Rule:  rawRule(rule),
				Err:   ErrUnknownRule,
			}
		}
		if len(rule.Params) < definition.minArgs {
			return FieldPlan{}, &CompileError{
				Field: field.Name,
				Rule:  rawRule(rule),
				Err:   ErrMissingOperand,
			}
		}
		if definition.maxArgs >= 0 && len(rule.Params) > definition.maxArgs {
			return FieldPlan{}, &CompileError{
				Field: field.Name,
				Rule:  rawRule(rule),
				Err:   ErrInvalidArity,
			}
		}
		family := fieldFamily(field.Type)
		if len(definition.families) > 0 && !slices.Contains(definition.families, family) {
			return FieldPlan{}, &CompileError{
				Field: field.Name,
				Rule:  rawRule(rule),
				Err:   fmt.Errorf("%w: %v", ErrInapplicableRule, field.Type),
			}
		}
		operation := RulePlan{Name: rule.Name, Op: definition.op, Family: family}
		operand, err := compileOperand(rule, operation.Family, field.Type)
		if err != nil {
			return FieldPlan{}, &CompileError{
				Field: field.Name,
				Rule:  rawRule(rule),
				Err:   fmt.Errorf("%w: %w", ErrInvalidOperand, err),
			}
		}
		operation.Operand = operand
		operations = append(operations, operation)
	}
	hasDefaultLike := hasDefaultLikeOperation(operations)

	return FieldPlan{
		Operations:             operations,
		RuntimePointerOptional: optionalPlacement(field.NeedsPointerOptional(), hasDefaultLike),
		GeneratedOptional:      optionalPlacement(field.NeedsGeneratedOptional(), hasDefaultLike),
	}, nil
}

func fieldFamily(fieldType reflect.Type) FieldFamily {
	if fieldType == nil {
		return FieldFamilyUnknown
	}
	for fieldType.Kind() == reflect.Pointer {
		fieldType = fieldType.Elem()
	}
	if fieldType.PkgPath() == "time" && fieldType.Name() == "Time" {
		return FieldFamilyTime
	}
	switch fieldType.Kind() {
	case reflect.String:
		return FieldFamilyString
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return FieldFamilySignedInteger
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return FieldFamilyUnsignedInteger
	case reflect.Float32, reflect.Float64:
		return FieldFamilyFloat
	case reflect.Bool:
		return FieldFamilyBool
	case reflect.Slice:
		return FieldFamilySlice
	case reflect.Array:
		return FieldFamilyArray
	case reflect.Map:
		return FieldFamilyMap
	case reflect.Struct:
		return FieldFamilyStruct
	default:
		return FieldFamilyUnknown
	}
}

func compileOperand(rule TagRule, family FieldFamily, fieldType reflect.Type) (any, error) {
	if len(rule.Params) == 0 {
		return nil, nil
	}
	value := rule.Params[0]
	switch rule.Name {
	case "gt", "gte", "lt", "lte", "multipleof", "min", "max":
		switch family {
		case FieldFamilySignedInteger:
			return strconv.ParseInt(value, 10, typeBits(fieldType))
		case FieldFamilyUnsignedInteger:
			return strconv.ParseUint(value, 10, typeBits(fieldType))
		case FieldFamilyFloat:
			return strconv.ParseFloat(value, typeBits(fieldType))
		default:
			return strconv.Atoi(value)
		}
	case "length":
		return strconv.Atoi(value)
	case "regex":
		return regexp.Compile(value)
	case "default", "prefault":
		return compileScalarValue(strings.Join(rule.Params, " "), family, fieldType)
	case "literal":
		return compileScalarValue(value, family, fieldType)
	case "enum":
		values := make([]any, len(rule.Params))
		for i, param := range rule.Params {
			compiled, err := compileScalarValue(param, family, fieldType)
			if err != nil {
				return nil, err
			}
			values[i] = compiled
		}
		return values, nil
	default:
		return value, nil
	}
}

func compileScalarValue(value string, family FieldFamily, fieldType reflect.Type) (any, error) {
	switch family {
	case FieldFamilyString:
		return value, nil
	case FieldFamilyBool:
		return strconv.ParseBool(value)
	case FieldFamilySignedInteger:
		return strconv.ParseInt(value, 10, typeBits(fieldType))
	case FieldFamilyUnsignedInteger:
		return strconv.ParseUint(value, 10, typeBits(fieldType))
	case FieldFamilyFloat:
		return strconv.ParseFloat(value, typeBits(fieldType))
	case FieldFamilySlice, FieldFamilyArray, FieldFamilyMap:
		for fieldType.Kind() == reflect.Pointer {
			fieldType = fieldType.Elem()
		}
		target := reflect.New(fieldType)
		if err := json.Unmarshal([]byte(value), target.Interface()); err != nil {
			return nil, err
		}
		return target.Elem().Interface(), nil
	default:
		return value, nil
	}
}

func typeBits(fieldType reflect.Type) int {
	for fieldType != nil && fieldType.Kind() == reflect.Pointer {
		fieldType = fieldType.Elem()
	}
	if fieldType == nil {
		return 64
	}
	return fieldType.Bits()
}

func rawRule(rule TagRule) string {
	if len(rule.Params) == 0 {
		return rule.Name
	}
	return rule.Name + "=" + strings.Join(rule.Params, " ")
}

// OperationsExcept returns planned operations except those for the named raw rules.
func (p FieldPlan) OperationsExcept(names ...string) []RulePlan {
	if len(names) == 0 {
		return slices.Clone(p.Operations)
	}
	operations := make([]RulePlan, 0, len(p.Operations))
	for _, operation := range p.Operations {
		if !slices.Contains(names, operation.Name) {
			operations = append(operations, operation)
		}
	}
	return operations
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
