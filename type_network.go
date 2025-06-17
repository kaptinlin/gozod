package gozod

import (
	"errors"
	"net"
	"strings"

	"github.com/kaptinlin/gozod/regexes"
)

//////////////////////////////////////////
//////////////////////////////////////////
//////////                      //////////
//////////   Network Types      //////////
//////////                      //////////
//////////////////////////////////////////
//////////////////////////////////////////

//////////////////////////////////////////
//////////   IPv4 Type           ////////
//////////////////////////////////////////

// ZodIPv4Def defines the configuration for IPv4 address validation
type ZodIPv4Def struct {
	ZodTypeDef
	Type    string // "ipv4"
	Version string // "v4"
}

// ZodIPv4Internals contains IPv4 validator internal state
type ZodIPv4Internals struct {
	ZodTypeInternals
	Def  *ZodIPv4Def
	Isst ZodIssueInvalidType
}

// ZodIPv4 represents an IPv4 address validation schema
type ZodIPv4 struct {
	internals *ZodIPv4Internals
}

// GetInternals returns the internal state of the schema
func (z *ZodIPv4) GetInternals() *ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// GetZod returns the IPv4-specific internals for framework usage
func (z *ZodIPv4) GetZod() *ZodIPv4Internals {
	return z.internals
}

// Parse validates the input value against the IPv4 schema
func (z *ZodIPv4) Parse(input any, ctx ...*ParseContext) (any, error) {
	var parseCtx *ParseContext
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}
	return Parse[any, any](any(z).(ZodType[any, any]), input, parseCtx)
}

// MustParse validates the input value and panics on failure
func (z *ZodIPv4) MustParse(input any, ctx ...*ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Refine adds custom validation to the IPv4 schema
func (z *ZodIPv4) RefineAny(fn func(any) bool, params ...SchemaParams) ZodType[any, any] {
	var schemaParams SchemaParams
	if len(params) > 0 {
		schemaParams = params[0]
	}

	return NewZodCustom(RefineFn[interface{}](fn), schemaParams)
}

// Check adds modern validation using direct payload access
func (z *ZodIPv4) Check(fn CheckFn) ZodType[any, any] {
	custom := NewZodCustom(fn, SchemaParams{})
	custom.GetInternals().Parse = func(payload *ParsePayload, ctx *ParseContext) *ParsePayload {
		// first execute the original parse
		result, err := z.Parse(payload.Value, ctx)
		if err != nil {
			payload.Issues = append(payload.Issues, ZodRawIssue{
				Code:    "invalid_type",
				Message: err.Error(),
				Input:   payload.Value,
			})
			return payload
		}
		payload.Value = result

		// then execute the check function
		fn(payload)
		return payload
	}
	return any(custom).(ZodType[any, any])
}

// Optional makes the IPv4 address optional
func (z *ZodIPv4) Optional() ZodType[any, any] {
	return any(Optional(z)).(ZodType[any, any])
}

// Nilable makes the IPv4 address nilable
// - Only changes nil handling, not type inference
// - Uses standard Clone + setNilable pattern like type_string.go
func (z *ZodIPv4) Nilable() ZodType[any, any] {
	// use Clone method to create a new instance, avoid manual state copying
	return Clone(z, func(def *ZodTypeDef) {
		// no need to modify def, because Nilable is a runtime flag
	}).(*ZodIPv4).setNilable()
}

// setNilable set the Nilable flag internal method
func (z *ZodIPv4) setNilable() ZodType[any, any] {
	z.internals.Nilable = true
	return z
}

// TransformAny creates a transform with given function
func (z *ZodIPv4) TransformAny(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	transform := NewZodTransform[any, any](fn)

	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: any(transform).(ZodType[any, any]),
		def: ZodTypeDef{Type: "pipe"},
	}
}

// Pipe creates a pipe with another schema
func (z *ZodIPv4) Pipe(schema ZodType[any, any]) ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: schema,
		def: ZodTypeDef{Type: "pipe"},
	}
}

// Prefault adds a prefault value
func (z *ZodIPv4) Prefault(value any) ZodType[any, any] {
	return Prefault[any, any](any(z).(ZodType[any, any]), value)
}

// PrefaultFunc adds a prefault value using a function
func (z *ZodIPv4) PrefaultFunc(fn func() any) ZodType[any, any] {
	return PrefaultFunc[any, any](any(z).(ZodType[any, any]), fn)
}

// createZodIPv4FromDef creates a ZodIPv4 from definition
func createZodIPv4FromDef(def *ZodIPv4Def) *ZodIPv4 {
	internals := &ZodIPv4Internals{
		ZodTypeInternals: newBaseZodTypeInternals("ipv4"),
		Def:              def,
		Isst:             ZodIssueInvalidType{Expected: "IPv4 address"},
	}

	// Set up parse function
	internals.Parse = func(payload *ParsePayload, ctx *ParseContext) *ParsePayload {
		parseCtx := (*ParseContext)(nil)
		if ctx != nil {
			parseCtx = ctx
		}

		// use parseType template, including smart type inference and IPv4 format validation
		typeChecker := func(v any) (string, bool) {
			if str, ok := v.(string); ok {
				return str, true
			}
			return "", false
		}

		coercer := func(v any) (string, bool) {
			// IPv4 does not support coercion, only accepts strings
			return "", false
		}

		validator := func(value string, checks []ZodCheck, ctx *ParseContext) error {
			// run basic checks
			if len(checks) > 0 {
				runChecksOnValue(value, checks, payload, ctx)
				if len(payload.Issues) > 0 {
					return &ZodError{Issues: convertRawIssuesToIssues(payload.Issues, ctx)}
				}
			}

			// IPv4 format validation
			if !regexes.IPv4.MatchString(value) {
				issue := CreateInvalidFormatIssue(value, "ipv4", WithPattern(regexes.IPv4.String()))
				payload.Issues = append(payload.Issues, issue)
				return &ZodError{Issues: convertRawIssuesToIssues(payload.Issues, ctx)}
			}

			// use Go's net package for additional validation
			ip := net.ParseIP(value)
			if ip == nil || ip.To4() == nil {
				issue := CreateInvalidFormatIssue(value, "ipv4", WithPattern(regexes.IPv4.String()))
				payload.Issues = append(payload.Issues, issue)
				return &ZodError{Issues: convertRawIssuesToIssues(payload.Issues, ctx)}
			}

			return nil
		}

		result, err := parseType[string](
			payload.Value,
			&internals.ZodTypeInternals,
			"ipv4",
			typeChecker,
			func(v any) (*string, bool) { ptr, ok := v.(*string); return ptr, ok },
			validator,
			coercer,
			parseCtx,
		)

		if err != nil {
			// convert error to ParsePayload format
			var zodErr *ZodError
			if errors.As(err, &zodErr) {
				for _, issue := range zodErr.Issues {
					rawIssue := ZodRawIssue{
						Code:    issue.Code,
						Input:   issue.Input,
						Path:    issue.Path,
						Message: issue.Message,
					}
					payload.Issues = append(payload.Issues, rawIssue)
				}
			}
			return payload
		}

		payload.Value = result
		return payload
	}

	schema := &ZodIPv4{internals: internals}

	// Set constructor for Clone support
	internals.Constructor = func(def *ZodTypeDef) ZodType[any, any] {
		ipv4Def := &ZodIPv4Def{
			ZodTypeDef: *def,
			Type:       "ipv4",
			Version:    "v4",
		}
		return any(createZodIPv4FromDef(ipv4Def)).(ZodType[any, any])
	}

	// Initialize the schema
	initZodType(any(schema).(ZodType[any, any]), &def.ZodTypeDef)

	return schema
}

//////////////////////////////////////////
//////////   IPv6 Type           ////////
//////////////////////////////////////////

// ZodIPv6Def defines the configuration for IPv6 address validation
type ZodIPv6Def struct {
	ZodTypeDef
	Type    string // "ipv6"
	Version string // "v6"
}

// ZodIPv6Internals contains IPv6 validator internal state
type ZodIPv6Internals struct {
	ZodTypeInternals
	Def  *ZodIPv6Def
	Isst ZodIssueInvalidType
}

// ZodIPv6 represents an IPv6 address validation schema
type ZodIPv6 struct {
	internals *ZodIPv6Internals
}

// GetInternals returns the internal state of the schema
func (z *ZodIPv6) GetInternals() *ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// GetZod returns the IPv6-specific internals for framework usage
func (z *ZodIPv6) GetZod() *ZodIPv6Internals {
	return z.internals
}

// Parse validates the input value against the IPv6 schema
func (z *ZodIPv6) Parse(input any, ctx ...*ParseContext) (any, error) {
	var parseCtx *ParseContext
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}
	return Parse[any, any](any(z).(ZodType[any, any]), input, parseCtx)
}

// MustParse validates the input value and panics on failure
func (z *ZodIPv6) MustParse(input any, ctx ...*ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Refine adds custom validation to the IPv6 schema
func (z *ZodIPv6) RefineAny(fn func(any) bool, params ...SchemaParams) ZodType[any, any] {
	var schemaParams SchemaParams
	if len(params) > 0 {
		schemaParams = params[0]
	}

	return NewZodCustom(RefineFn[interface{}](fn), schemaParams)
}

// Check adds modern validation using direct payload access
func (z *ZodIPv6) Check(fn CheckFn) ZodType[any, any] {
	custom := NewZodCustom(fn, SchemaParams{})
	custom.GetInternals().Parse = func(payload *ParsePayload, ctx *ParseContext) *ParsePayload {
		// first execute the original parse
		result, err := z.Parse(payload.Value, ctx)
		if err != nil {
			payload.Issues = append(payload.Issues, ZodRawIssue{
				Code:    "invalid_type",
				Message: err.Error(),
				Input:   payload.Value,
			})
			return payload
		}
		payload.Value = result

		// then execute the check function
		fn(payload)
		return payload
	}
	return any(custom).(ZodType[any, any])
}

// Optional makes the IPv6 address optional
func (z *ZodIPv6) Optional() ZodType[any, any] {
	return any(Optional(z)).(ZodType[any, any])
}

// Nilable makes the IPv6 address nilable
// - Only changes nil handling, not type inference
// - Uses standard Clone + setNilable pattern like type_string.go
func (z *ZodIPv6) Nilable() ZodType[any, any] {
	// use Clone method to create a new instance, avoid manual state copying
	return Clone(z, func(def *ZodTypeDef) {
		// no need to modify def, because Nilable is a runtime flag
	}).(*ZodIPv6).setNilable()
}

// setNilable set the Nilable flag internal method
func (z *ZodIPv6) setNilable() ZodType[any, any] {
	z.internals.Nilable = true
	return z
}

// Transform creates a transform with given function
func (z *ZodIPv6) Transform(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	return z.TransformAny(fn)
}

// TransformAny creates a transform with given function
func (z *ZodIPv6) TransformAny(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	transform := NewZodTransform[any, any](fn)

	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: any(transform).(ZodType[any, any]),
		def: ZodTypeDef{Type: "pipe"},
	}
}

// Pipe creates a pipe with another schema
func (z *ZodIPv6) Pipe(schema ZodType[any, any]) ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: schema,
		def: ZodTypeDef{Type: "pipe"},
	}
}

// Prefault adds a prefault value
func (z *ZodIPv6) Prefault(value any) ZodType[any, any] {
	return Prefault[any, any](any(z).(ZodType[any, any]), value)
}

// PrefaultFunc adds a prefault value using a function
func (z *ZodIPv6) PrefaultFunc(fn func() any) ZodType[any, any] {
	return PrefaultFunc[any, any](any(z).(ZodType[any, any]), fn)
}

// createZodIPv6FromDef creates a ZodIPv6 from definition
func createZodIPv6FromDef(def *ZodIPv6Def) *ZodIPv6 {
	internals := &ZodIPv6Internals{
		ZodTypeInternals: newBaseZodTypeInternals("ipv6"),
		Def:              def,
		Isst:             ZodIssueInvalidType{Expected: "IPv6 address"},
	}

	// Set up parse function
	internals.Parse = func(payload *ParsePayload, ctx *ParseContext) *ParsePayload {
		parseCtx := (*ParseContext)(nil)
		if ctx != nil {
			parseCtx = ctx
		}

		// use parseType template, including smart type inference and IPv6 format validation
		typeChecker := func(v any) (string, bool) {
			if str, ok := v.(string); ok {
				return str, true
			}
			return "", false
		}

		coercer := func(v any) (string, bool) {
			// IPv6 does not support coercion, only accepts strings
			return "", false
		}

		validator := func(value string, checks []ZodCheck, ctx *ParseContext) error {
			// run basic checks
			if len(checks) > 0 {
				runChecksOnValue(value, checks, payload, ctx)
				if len(payload.Issues) > 0 {
					return &ZodError{Issues: convertRawIssuesToIssues(payload.Issues, ctx)}
				}
			}

			// IPv6 format validation
			if !regexes.IPv6.MatchString(value) {
				issue := CreateInvalidFormatIssue(value, "ipv6", WithPattern(regexes.IPv6.String()))
				payload.Issues = append(payload.Issues, issue)
				return &ZodError{Issues: convertRawIssuesToIssues(payload.Issues, ctx)}
			}

			// use Go's net package for additional validation
			ip := net.ParseIP(value)
			if ip == nil || ip.To4() != nil {
				// ip.To4() != nil means it's an IPv4 address, not IPv6
				issue := CreateInvalidFormatIssue(value, "ipv6", WithPattern(regexes.IPv6.String()))
				payload.Issues = append(payload.Issues, issue)
				return &ZodError{Issues: convertRawIssuesToIssues(payload.Issues, ctx)}
			}

			return nil
		}

		result, err := parseType[string](
			payload.Value,
			&internals.ZodTypeInternals,
			"ipv6",
			typeChecker,
			func(v any) (*string, bool) { ptr, ok := v.(*string); return ptr, ok },
			validator,
			coercer,
			parseCtx,
		)

		if err != nil {
			// convert error to ParsePayload format
			var zodErr *ZodError
			if errors.As(err, &zodErr) {
				for _, issue := range zodErr.Issues {
					rawIssue := ZodRawIssue{
						Code:    issue.Code,
						Input:   issue.Input,
						Path:    issue.Path,
						Message: issue.Message,
					}
					payload.Issues = append(payload.Issues, rawIssue)
				}
			}
			return payload
		}

		payload.Value = result
		return payload
	}

	schema := &ZodIPv6{internals: internals}

	// Set constructor for Clone support
	internals.Constructor = func(def *ZodTypeDef) ZodType[any, any] {
		ipv6Def := &ZodIPv6Def{
			ZodTypeDef: *def,
			Type:       "ipv6",
			Version:    "v6",
		}
		return any(createZodIPv6FromDef(ipv6Def)).(ZodType[any, any])
	}

	// Initialize the schema
	initZodType(any(schema).(ZodType[any, any]), &def.ZodTypeDef)

	return schema
}

//////////////////////////////////////////
//////////   CIDRv4 Type         ////////
//////////////////////////////////////////

// ZodCIDRv4Def defines the configuration for IPv4 CIDR validation
type ZodCIDRv4Def struct {
	ZodTypeDef
	Type    string // "cidrv4"
	Version string // "v4"
}

// ZodCIDRv4Internals contains IPv4 CIDR validator internal state
type ZodCIDRv4Internals struct {
	ZodTypeInternals
	Def  *ZodCIDRv4Def
	Isst ZodIssueInvalidType
}

// ZodCIDRv4 represents an IPv4 CIDR validation schema
type ZodCIDRv4 struct {
	internals *ZodCIDRv4Internals
}

// GetInternals returns the internal state of the schema
func (z *ZodCIDRv4) GetInternals() *ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// GetZod returns the CIDRv4-specific internals for framework usage
func (z *ZodCIDRv4) GetZod() *ZodCIDRv4Internals {
	return z.internals
}

// Parse smart type inference validation and parsing
func (z *ZodCIDRv4) Parse(input any, ctx ...*ParseContext) (any, error) {
	// use unified nil handling logic
	if input == nil {
		if !z.internals.Nilable {
			rawIssue := CreateInvalidTypeIssue(input, "string", "null")
			finalIssue := FinalizeIssue(rawIssue, nil, GetConfig())
			return nil, NewZodError([]ZodIssue{finalIssue})
		}
		return (*string)(nil), nil // return string type nil pointer
	}

	// smart type inference: input type determines output type
	switch v := input.(type) {
	case string:
		// validate CIDR format
		if err := z.validateCIDRv4(v); err != nil {
			return nil, err
		}
		return v, nil // string → string (keep original type)

	case *string:
		if v == nil {
			if !z.internals.Nilable {
				rawIssue := CreateInvalidTypeIssue(input, "string", "null")
				finalIssue := FinalizeIssue(rawIssue, nil, GetConfig())
				return nil, NewZodError([]ZodIssue{finalIssue})
			}
			return (*string)(nil), nil
		}
		// validate CIDR format
		if err := z.validateCIDRv4(*v); err != nil {
			return nil, err
		}
		return v, nil // *string → *string (keep original type)

	default:
		// use unified error creation
		rawIssue := CreateInvalidTypeIssue(input, "string", string(GetParsedType(input)))
		finalIssue := FinalizeIssue(rawIssue, nil, GetConfig())
		return nil, NewZodError([]ZodIssue{finalIssue})
	}
}

// validateCIDRv4 validates IPv4 CIDR format
func (z *ZodCIDRv4) validateCIDRv4(value string) error {
	// Use net.ParseCIDR for accurate CIDR validation
	_, _, err := net.ParseCIDR(value)
	if err != nil {
		rawIssue := CreateInvalidFormatIssue(value, "cidrv4")
		finalIssue := FinalizeIssue(rawIssue, nil, GetConfig())
		return NewZodError([]ZodIssue{finalIssue})
	}
	return nil
}

// MustParse validates the input value and panics on failure
func (z *ZodCIDRv4) MustParse(input any, ctx ...*ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Refine adds custom validation to the CIDRv4 schema
func (z *ZodCIDRv4) RefineAny(fn func(any) bool, params ...SchemaParams) ZodType[any, any] {
	var schemaParams SchemaParams
	if len(params) > 0 {
		schemaParams = params[0]
	}

	return NewZodCustom(RefineFn[interface{}](fn), schemaParams)
}

// Check adds modern validation using direct payload access
func (z *ZodCIDRv4) Check(fn CheckFn) ZodType[any, any] {
	custom := NewZodCustom(fn, SchemaParams{})
	custom.GetInternals().Parse = func(payload *ParsePayload, ctx *ParseContext) *ParsePayload {
		// first execute the original parse
		result, err := z.Parse(payload.Value, ctx)
		if err != nil {
			payload.Issues = append(payload.Issues, ZodRawIssue{
				Code:    "invalid_type",
				Message: err.Error(),
				Input:   payload.Value,
			})
			return payload
		}
		payload.Value = result

		// then execute the check function
		fn(payload)
		return payload
	}
	return any(custom).(ZodType[any, any])
}

// Optional makes the CIDRv4 optional
func (z *ZodCIDRv4) Optional() ZodType[any, any] {
	return any(Optional(z)).(ZodType[any, any])
}

// Nilable makes the CIDRv4 nilable
// Nilable makes the CIDRv4 address nilable
// - Only changes nil handling, not type inference
// - Uses standard Clone + setNilable pattern like type_string.go
func (z *ZodCIDRv4) Nilable() ZodType[any, any] {
	// use Clone method to create a new instance, avoid manual state copying
	return Clone(z, func(def *ZodTypeDef) {
		// no need to modify def, because Nilable is a runtime flag
	}).(*ZodCIDRv4).setNilable()
}

// setNilable set the Nilable flag internal method
func (z *ZodCIDRv4) setNilable() ZodType[any, any] {
	z.internals.Nilable = true
	return z
}

// Transform creates a transform with given function
func (z *ZodCIDRv4) Transform(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	return z.TransformAny(fn)
}

// TransformAny creates a transform with given function
func (z *ZodCIDRv4) TransformAny(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	transform := NewZodTransform[any, any](fn)

	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: any(transform).(ZodType[any, any]),
		def: ZodTypeDef{Type: "pipe"},
	}
}

// Pipe creates a pipe with another schema
func (z *ZodCIDRv4) Pipe(schema ZodType[any, any]) ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: schema,
		def: ZodTypeDef{Type: "pipe"},
	}
}

// Prefault adds a prefault value
func (z *ZodCIDRv4) Prefault(value any) ZodType[any, any] {
	return Prefault[any, any](any(z).(ZodType[any, any]), value)
}

// PrefaultFunc adds a prefault value using a function
func (z *ZodCIDRv4) PrefaultFunc(fn func() any) ZodType[any, any] {
	return PrefaultFunc[any, any](any(z).(ZodType[any, any]), fn)
}

// createZodCIDRv4FromDef creates a ZodCIDRv4 from definition
func createZodCIDRv4FromDef(def *ZodCIDRv4Def) *ZodCIDRv4 {
	internals := &ZodCIDRv4Internals{
		ZodTypeInternals: newBaseZodTypeInternals("cidrv4"),
		Def:              def,
		Isst:             ZodIssueInvalidType{Expected: "IPv4 CIDR"},
	}

	// Set up parse function
	internals.Parse = func(payload *ParsePayload, ctx *ParseContext) *ParsePayload {
		parseCtx := (*ParseContext)(nil)
		if ctx != nil {
			parseCtx = ctx
		}

		// use parseType template, including smart type inference and CIDRv4 format validation
		typeChecker := func(v any) (string, bool) {
			if str, ok := v.(string); ok {
				return str, true
			}
			return "", false
		}

		coercer := func(v any) (string, bool) {
			// CIDRv4 does not support coercion, only accepts strings
			return "", false
		}

		validator := func(value string, checks []ZodCheck, ctx *ParseContext) error {
			// run basic checks
			if len(checks) > 0 {
				runChecksOnValue(value, checks, payload, ctx)
				if len(payload.Issues) > 0 {
					return &ZodError{Issues: convertRawIssuesToIssues(payload.Issues, ctx)}
				}
			}

			// CIDRv4 format validation
			if !regexes.CIDRv4.MatchString(value) {
				issue := CreateInvalidFormatIssue(value, "cidrv4", WithPattern(regexes.CIDRv4.String()))
				payload.Issues = append(payload.Issues, issue)
				return &ZodError{Issues: convertRawIssuesToIssues(payload.Issues, ctx)}
			}

			// use Go's net package for additional validation
			_, _, err := net.ParseCIDR(value)
			if err != nil {
				issue := CreateInvalidFormatIssue(value, "cidrv4", WithPattern(regexes.CIDRv4.String()))
				payload.Issues = append(payload.Issues, issue)
				return &ZodError{Issues: convertRawIssuesToIssues(payload.Issues, ctx)}
			}

			return nil
		}

		result, err := parseType[string](
			payload.Value,
			&internals.ZodTypeInternals,
			"cidrv4",
			typeChecker,
			func(v any) (*string, bool) { ptr, ok := v.(*string); return ptr, ok },
			validator,
			coercer,
			parseCtx,
		)

		if err != nil {
			// convert error to ParsePayload format
			var zodErr *ZodError
			if errors.As(err, &zodErr) {
				for _, issue := range zodErr.Issues {
					rawIssue := ZodRawIssue{
						Code:    issue.Code,
						Input:   issue.Input,
						Path:    issue.Path,
						Message: issue.Message,
					}
					payload.Issues = append(payload.Issues, rawIssue)
				}
			}
			return payload
		}

		payload.Value = result
		return payload
	}

	schema := &ZodCIDRv4{internals: internals}

	// Set constructor for Clone support
	internals.Constructor = func(def *ZodTypeDef) ZodType[any, any] {
		cidrv4Def := &ZodCIDRv4Def{
			ZodTypeDef: *def,
			Type:       "cidrv4",
			Version:    "v4",
		}
		return any(createZodCIDRv4FromDef(cidrv4Def)).(ZodType[any, any])
	}

	// Initialize the schema
	initZodType(any(schema).(ZodType[any, any]), &def.ZodTypeDef)

	return schema
}

//////////////////////////////////////////
//////////   CIDRv6 Type         ////////
//////////////////////////////////////////

// ZodCIDRv6Def defines the configuration for IPv6 CIDR validation
type ZodCIDRv6Def struct {
	ZodTypeDef
	Type    string // "cidrv6"
	Version string // "v6"
}

// ZodCIDRv6Internals contains IPv6 CIDR validator internal state
type ZodCIDRv6Internals struct {
	ZodTypeInternals
	Def  *ZodCIDRv6Def
	Isst ZodIssueInvalidType
}

// ZodCIDRv6 represents an IPv6 CIDR validation schema
type ZodCIDRv6 struct {
	internals *ZodCIDRv6Internals
}

// GetInternals returns the internal state of the schema
func (z *ZodCIDRv6) GetInternals() *ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// GetZod returns the CIDRv6-specific internals for framework usage
func (z *ZodCIDRv6) GetZod() *ZodCIDRv6Internals {
	return z.internals
}

// Parse smart type inference validation and parsing
func (z *ZodCIDRv6) Parse(input any, ctx ...*ParseContext) (any, error) {
	// use unified nil handling logic
	if input == nil {
		if !z.internals.Nilable {
			rawIssue := CreateInvalidTypeIssue(input, "string", "null")
			finalIssue := FinalizeIssue(rawIssue, nil, GetConfig())
			return nil, NewZodError([]ZodIssue{finalIssue})
		}
		return (*string)(nil), nil // return string type nil pointer
	}

	// smart type inference: input type determines output type
	switch v := input.(type) {
	case string:
		// validate CIDR format
		if err := z.validateCIDRv6(v); err != nil {
			return nil, err
		}
		return v, nil // string → string (keep original type)

	case *string:
		if v == nil {
			if !z.internals.Nilable {
				rawIssue := CreateInvalidTypeIssue(input, "string", "null")
				finalIssue := FinalizeIssue(rawIssue, nil, GetConfig())
				return nil, NewZodError([]ZodIssue{finalIssue})
			}
			return (*string)(nil), nil
		}
		// validate CIDR format
		if err := z.validateCIDRv6(*v); err != nil {
			return nil, err
		}
		return v, nil // *string → *string (keep original type)

	default:
		// use unified error creation
		rawIssue := CreateInvalidTypeIssue(input, "string", string(GetParsedType(input)))
		finalIssue := FinalizeIssue(rawIssue, nil, GetConfig())
		return nil, NewZodError([]ZodIssue{finalIssue})
	}
}

// validateCIDRv6 validates IPv6 CIDR format
func (z *ZodCIDRv6) validateCIDRv6(value string) error {
	// Use net.ParseCIDR for accurate CIDR validation
	ip, _, err := net.ParseCIDR(value)
	if err != nil {
		rawIssue := CreateInvalidFormatIssue(value, "cidrv6")
		finalIssue := FinalizeIssue(rawIssue, nil, GetConfig())
		return NewZodError([]ZodIssue{finalIssue})
	}
	// Ensure it's actually IPv6
	if ip.To4() != nil {
		rawIssue := CreateInvalidFormatIssue(value, "cidrv6")
		finalIssue := FinalizeIssue(rawIssue, nil, GetConfig())
		return NewZodError([]ZodIssue{finalIssue})
	}
	return nil
}

// MustParse validates the input value and panics on failure
func (z *ZodCIDRv6) MustParse(input any, ctx ...*ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Refine adds custom validation to the CIDRv6 schema
func (z *ZodCIDRv6) RefineAny(fn func(any) bool, params ...SchemaParams) ZodType[any, any] {
	var schemaParams SchemaParams
	if len(params) > 0 {
		schemaParams = params[0]
	}

	return NewZodCustom(RefineFn[interface{}](fn), schemaParams)
}

// Check adds modern validation using direct payload access
func (z *ZodCIDRv6) Check(fn CheckFn) ZodType[any, any] {
	custom := NewZodCustom(fn, SchemaParams{})
	custom.GetInternals().Parse = func(payload *ParsePayload, ctx *ParseContext) *ParsePayload {
		// first execute the original parse
		result, err := z.Parse(payload.Value, ctx)
		if err != nil {
			payload.Issues = append(payload.Issues, ZodRawIssue{
				Code:    "invalid_type",
				Message: err.Error(),
				Input:   payload.Value,
			})
			return payload
		}
		payload.Value = result

		// then execute the check function
		fn(payload)
		return payload
	}
	return any(custom).(ZodType[any, any])
}

// Optional makes the CIDRv6 optional
func (z *ZodCIDRv6) Optional() ZodType[any, any] {
	return any(Optional(z)).(ZodType[any, any])
}

// Nilable makes the CIDRv6 address nilable
// - Only changes nil handling, not type inference
// - Uses standard Clone + setNilable pattern like type_string.go
func (z *ZodCIDRv6) Nilable() ZodType[any, any] {
	// use Clone method to create a new instance, avoid manual state copying
	return Clone(z, func(def *ZodTypeDef) {
		// no need to modify def, because Nilable is a runtime flag
	}).(*ZodCIDRv6).setNilable()
}

// setNilable set the Nilable flag internal method
func (z *ZodCIDRv6) setNilable() ZodType[any, any] {
	z.internals.Nilable = true
	return z
}

// Transform creates a transform with given function
func (z *ZodCIDRv6) Transform(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	return z.TransformAny(fn)
}

// TransformAny creates a transform with given function
func (z *ZodCIDRv6) TransformAny(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	transform := NewZodTransform[any, any](fn)

	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: any(transform).(ZodType[any, any]),
		def: ZodTypeDef{Type: "pipe"},
	}
}

// Pipe creates a pipe with another schema
func (z *ZodCIDRv6) Pipe(schema ZodType[any, any]) ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: schema,
		def: ZodTypeDef{Type: "pipe"},
	}
}

// createZodCIDRv6FromDef creates a ZodCIDRv6 from definition
func createZodCIDRv6FromDef(def *ZodCIDRv6Def) *ZodCIDRv6 {
	internals := &ZodCIDRv6Internals{
		ZodTypeInternals: newBaseZodTypeInternals("cidrv6"),
		Def:              def,
		Isst:             ZodIssueInvalidType{Expected: "IPv6 CIDR"},
	}

	// Set up parse function
	internals.Parse = func(payload *ParsePayload, ctx *ParseContext) *ParsePayload {
		parseCtx := (*ParseContext)(nil)
		if ctx != nil {
			parseCtx = ctx
		}

		// use parseType template, including smart type inference and CIDRv6 format validation
		typeChecker := func(v any) (string, bool) {
			if str, ok := v.(string); ok {
				return str, true
			}
			return "", false
		}

		coercer := func(v any) (string, bool) {
			// CIDRv6 does not support coercion, only accepts strings
			return "", false
		}

		validator := func(value string, checks []ZodCheck, ctx *ParseContext) error {
			// run basic checks
			if len(checks) > 0 {
				runChecksOnValue(value, checks, payload, ctx)
				if len(payload.Issues) > 0 {
					return &ZodError{Issues: convertRawIssuesToIssues(payload.Issues, ctx)}
				}
			}

			// split address and prefix
			parts := strings.Split(value, "/")
			if len(parts) != 2 {
				issue := CreateInvalidFormatIssue(value, "cidrv6", WithPattern(regexes.CIDRv6.String()))
				payload.Issues = append(payload.Issues, issue)
				return &ZodError{Issues: convertRawIssuesToIssues(payload.Issues, ctx)}
			}

			address := parts[0]
			prefix := parts[1]

			// validate prefix length (must be 0-128)
			prefixNum := 0
			if _, network, err := net.ParseCIDR(value); err != nil || network == nil {
				issue := CreateInvalidFormatIssue(value, "cidrv6", WithPattern(regexes.CIDRv6.String()))
				payload.Issues = append(payload.Issues, issue)
				return &ZodError{Issues: convertRawIssuesToIssues(payload.Issues, ctx)}
			}

			// manually parse prefix length for range check
			for _, char := range prefix {
				if char < '0' || char > '9' {
					issue := CreateInvalidFormatIssue(value, "cidrv6", WithPattern(regexes.CIDRv6.String()))
					payload.Issues = append(payload.Issues, issue)
					return &ZodError{Issues: convertRawIssuesToIssues(payload.Issues, ctx)}
				}
				prefixNum = prefixNum*10 + int(char-'0')
			}

			if prefixNum < 0 || prefixNum > 128 {
				issue := CreateInvalidFormatIssue(value, "cidrv6", WithPattern(regexes.CIDRv6.String()))
				payload.Issues = append(payload.Issues, issue)
				return &ZodError{Issues: convertRawIssuesToIssues(payload.Issues, ctx)}
			}

			// validate IPv6 address part
			ip := net.ParseIP(address)
			if ip == nil || ip.To4() != nil {
				// ip.To4() != nil means it's an IPv4 address, not IPv6
				issue := CreateInvalidFormatIssue(value, "cidrv6", WithPattern(regexes.CIDRv6.String()))
				payload.Issues = append(payload.Issues, issue)
				return &ZodError{Issues: convertRawIssuesToIssues(payload.Issues, ctx)}
			}

			return nil
		}

		result, err := parseType[string](
			payload.Value,
			&internals.ZodTypeInternals,
			"cidrv6",
			typeChecker,
			func(v any) (*string, bool) { ptr, ok := v.(*string); return ptr, ok },
			validator,
			coercer,
			parseCtx,
		)

		if err != nil {
			// convert error to ParsePayload format
			var zodErr *ZodError
			if errors.As(err, &zodErr) {
				for _, issue := range zodErr.Issues {
					rawIssue := ZodRawIssue{
						Code:    issue.Code,
						Input:   issue.Input,
						Path:    issue.Path,
						Message: issue.Message,
					}
					payload.Issues = append(payload.Issues, rawIssue)
				}
			}
			return payload
		}

		payload.Value = result
		return payload
	}

	schema := &ZodCIDRv6{internals: internals}

	// Set constructor for Clone support
	internals.Constructor = func(def *ZodTypeDef) ZodType[any, any] {
		cidrv6Def := &ZodCIDRv6Def{
			ZodTypeDef: *def,
			Type:       "cidrv6",
			Version:    "v6",
		}
		return any(createZodCIDRv6FromDef(cidrv6Def)).(ZodType[any, any])
	}

	// Initialize the schema
	initZodType(any(schema).(ZodType[any, any]), &def.ZodTypeDef)

	return schema
}

//////////////////////////////////////////
//////////   Public Constructors   //////
//////////////////////////////////////////

// NewZodIPv4 creates a new IPv4 address schema
func NewZodIPv4(params ...SchemaParams) *ZodIPv4 {
	def := &ZodIPv4Def{
		ZodTypeDef: ZodTypeDef{Type: "ipv4"},
		Type:       "ipv4",
		Version:    "v4",
	}

	schema := createZodIPv4FromDef(def)

	// Apply schema parameters
	if len(params) > 0 {
		param := params[0]
		if param.Error != nil {
			errorMap := createErrorMap(param.Error)
			if errorMap != nil {
				def.Error = errorMap
				schema.internals.Error = errorMap
			}
		}
	}

	return schema
}

// NewZodIPv6 creates a new IPv6 address schema
func NewZodIPv6(params ...SchemaParams) *ZodIPv6 {
	def := &ZodIPv6Def{
		ZodTypeDef: ZodTypeDef{Type: "ipv6"},
		Type:       "ipv6",
		Version:    "v6",
	}

	schema := createZodIPv6FromDef(def)

	// Apply schema parameters
	if len(params) > 0 {
		param := params[0]
		if param.Error != nil {
			errorMap := createErrorMap(param.Error)
			if errorMap != nil {
				def.Error = errorMap
				schema.internals.Error = errorMap
			}
		}
	}

	return schema
}

// NewZodCIDRv4 creates a new IPv4 CIDR schema
func NewZodCIDRv4(params ...SchemaParams) *ZodCIDRv4 {
	def := &ZodCIDRv4Def{
		ZodTypeDef: ZodTypeDef{Type: "cidrv4"},
		Type:       "cidrv4",
		Version:    "v4",
	}

	schema := createZodCIDRv4FromDef(def)

	// Apply schema parameters
	if len(params) > 0 {
		param := params[0]
		if param.Error != nil {
			errorMap := createErrorMap(param.Error)
			if errorMap != nil {
				def.Error = errorMap
				schema.internals.Error = errorMap
			}
		}
	}

	return schema
}

// NewZodCIDRv6 creates a new IPv6 CIDR schema
func NewZodCIDRv6(params ...SchemaParams) *ZodCIDRv6 {
	def := &ZodCIDRv6Def{
		ZodTypeDef: ZodTypeDef{Type: "cidrv6"},
		Type:       "cidrv6",
		Version:    "v6",
	}

	schema := createZodCIDRv6FromDef(def)

	// Apply schema parameters
	if len(params) > 0 {
		param := params[0]
		if param.Error != nil {
			errorMap := createErrorMap(param.Error)
			if errorMap != nil {
				def.Error = errorMap
				schema.internals.Error = errorMap
			}
		}
	}

	return schema
}

//////////////////////////////////////////
//////////   Package-Level API    ///////
//////////////////////////////////////////

// IPv4 creates a new IPv4 address schema (main constructor)
func IPv4(params ...SchemaParams) *ZodIPv4 {
	return NewZodIPv4(params...)
}

// IPv6 creates a new IPv6 address schema (main constructor)
func IPv6(params ...SchemaParams) *ZodIPv6 {
	return NewZodIPv6(params...)
}

// CIDRv4 creates a new IPv4 CIDR schema (main constructor)
func CIDRv4(params ...SchemaParams) *ZodCIDRv4 {
	return NewZodCIDRv4(params...)
}

// CIDRv6 creates a new IPv6 CIDR schema (main constructor)
func CIDRv6(params ...SchemaParams) *ZodCIDRv6 {
	return NewZodCIDRv6(params...)
}

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodIPv4) Unwrap() ZodType[any, any] {
	return any(z).(ZodType[any, any])
}

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodIPv6) Unwrap() ZodType[any, any] {
	return any(z).(ZodType[any, any])
}

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodCIDRv4) Unwrap() ZodType[any, any] {
	return any(z).(ZodType[any, any])
}

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodCIDRv6) Unwrap() ZodType[any, any] {
	return any(z).(ZodType[any, any])
}

// ==================== NETWORK TYPES DEFAULT WRAPPER SYSTEM ====================

// ZodIPv4Default is the Default wrapper for IPv4 type
type ZodIPv4Default struct {
	*ZodDefault[*ZodIPv4] // embed specific pointer, allow method promotion
}

// GetInternals implements ZodType interface
func (s ZodIPv4Default) GetInternals() *ZodTypeInternals {
	return s.ZodDefault.GetInternals()
}

// Parse implements ZodType interface
func (s ZodIPv4Default) Parse(input any, ctx ...*ParseContext) (any, error) {
	return s.ZodDefault.Parse(input, ctx...)
}

// ZodIPv6Default is the Default wrapper for IPv6 type
type ZodIPv6Default struct {
	*ZodDefault[*ZodIPv6] // embed specific pointer, allow method promotion
}

// GetInternals implements ZodType interface
func (s ZodIPv6Default) GetInternals() *ZodTypeInternals {
	return s.ZodDefault.GetInternals()
}

// Parse implements ZodType interface
func (s ZodIPv6Default) Parse(input any, ctx ...*ParseContext) (any, error) {
	return s.ZodDefault.Parse(input, ctx...)
}

// ZodCIDRv4Default is the Default wrapper for CIDRv4 type
type ZodCIDRv4Default struct {
	*ZodDefault[*ZodCIDRv4] // embed specific pointer, allow method promotion
}

// GetInternals implements ZodType interface
func (s ZodCIDRv4Default) GetInternals() *ZodTypeInternals {
	return s.ZodDefault.GetInternals()
}

// Parse implements ZodType interface
func (s ZodCIDRv4Default) Parse(input any, ctx ...*ParseContext) (any, error) {
	return s.ZodDefault.Parse(input, ctx...)
}

// ZodCIDRv6Default is the Default wrapper for CIDRv6 type
type ZodCIDRv6Default struct {
	*ZodDefault[*ZodCIDRv6] // embed specific pointer, allow method promotion
}

// GetInternals implements ZodType interface
func (s ZodCIDRv6Default) GetInternals() *ZodTypeInternals {
	return s.ZodDefault.GetInternals()
}

// Parse implements ZodType interface
func (s ZodCIDRv6Default) Parse(input any, ctx ...*ParseContext) (any, error) {
	return s.ZodDefault.Parse(input, ctx...)
}

// ==================== IPv4 DEFAULT METHOD IMPLEMENTATION ====================

// Default adds a default value to the IPv4 schema, returns ZodIPv4Default support chain call
func (z *ZodIPv4) Default(value any) ZodIPv4Default {
	return ZodIPv4Default{
		&ZodDefault[*ZodIPv4]{
			innerType:    z,
			defaultValue: value,
			isFunction:   false,
		},
	}
}

// DefaultFunc adds a default function to the IPv4 schema, returns ZodIPv4Default support chain call
func (z *ZodIPv4) DefaultFunc(fn func() any) ZodIPv4Default {
	genericFn := func() any { return fn() }
	return ZodIPv4Default{
		&ZodDefault[*ZodIPv4]{
			innerType:   z,
			defaultFunc: genericFn,
			isFunction:  true,
		},
	}
}

// Optional adds an optional check to the IPv4 schema, returns ZodType support chain call
func (s ZodIPv4Default) Optional() ZodType[any, any] {
	return Optional(any(s).(ZodType[any, any]))
}

// Nilable adds a nilable check to the IPv4 schema, returns ZodType support chain call
func (s ZodIPv4Default) Nilable() ZodType[any, any] {
	return Nilable(any(s).(ZodType[any, any]))
}

// ==================== IPv6 DEFAULT METHOD IMPLEMENTATION ====================

// Default adds a default value to the IPv6 schema, returns ZodIPv6Default support chain call
func (z *ZodIPv6) Default(value any) ZodIPv6Default {
	return ZodIPv6Default{
		&ZodDefault[*ZodIPv6]{
			innerType:    z,
			defaultValue: value,
			isFunction:   false,
		},
	}
}

// DefaultFunc adds a default function to the IPv6 schema, returns ZodIPv6Default support chain call
func (z *ZodIPv6) DefaultFunc(fn func() any) ZodIPv6Default {
	genericFn := func() any { return fn() }
	return ZodIPv6Default{
		&ZodDefault[*ZodIPv6]{
			innerType:   z,
			defaultFunc: genericFn,
			isFunction:  true,
		},
	}
}

// Optional adds an optional check to the IPv6 schema, returns ZodType support chain call
func (s ZodIPv6Default) Optional() ZodType[any, any] {
	return Optional(any(s).(ZodType[any, any]))
}

// Nilable adds a nilable check to the IPv6 schema, returns ZodType support chain call
func (s ZodIPv6Default) Nilable() ZodType[any, any] {
	return Nilable(any(s).(ZodType[any, any]))
}

// ==================== CIDRv4 DEFAULT METHOD IMPLEMENTATION ====================

// Default adds a default value to the CIDRv4 schema, returns ZodCIDRv4Default support chain call
func (z *ZodCIDRv4) Default(value any) ZodCIDRv4Default {
	return ZodCIDRv4Default{
		&ZodDefault[*ZodCIDRv4]{
			innerType:    z,
			defaultValue: value,
			isFunction:   false,
		},
	}
}

// DefaultFunc adda  fdufiulothe CIDRv4to the CIDRv4 schema,  schemsrZodeturnsDefault Zuppodt chaCI callRv4Default support chain call
func (z *ZodCIDRv4) DefaultFunc(fn func() any) ZodCIDRv4Default {
	genericFn := func() any { return fn() }
	return ZodCIDRv4Default{
		&ZodDefault[*ZodCIDRv4]{
			innerType:   z,
			defaultFunc: genericFn,
			isFunction:  true,
		},
	}
}

// Optional adds an optional check to the CIDRv4 schema, returns ZodType support chain call
func (s ZodCIDRv4Default) Optional() ZodType[any, any] {
	return Optional(any(s).(ZodType[any, any]))
}

// Nilable adds a nilable check to the CIDRv4 schema, returns ZodType support chain call
func (s ZodCIDRv4Default) Nilable() ZodType[any, any] {
	return Nilable(any(s).(ZodType[any, any]))
}

// ==================== CIDRv6 DEFAULT METHOD IMPLEMENTATION ====================

// Default adds a default value to the CIDRv6 schema, returns ZodCIDRv6Default support chain call
func (z *ZodCIDRv6) Default(value any) ZodCIDRv6Default {
	return ZodCIDRv6Default{
		&ZodDefault[*ZodCIDRv6]{
			innerType:    z,
			defaultValue: value,
			isFunction:   false,
		},
	}
}

// DefaultFunc adds a default function to the CIDRv6 schema, returns ZodCIDRv6Default support chain call
func (z *ZodCIDRv6) DefaultFunc(fn func() any) ZodCIDRv6Default {
	genericFn := func() any { return fn() }
	return ZodCIDRv6Default{
		&ZodDefault[*ZodCIDRv6]{
			innerType:   z,
			defaultFunc: genericFn,
			isFunction:  true,
		},
	}
}

// Optional adds an optional check to the CIDRv6 schema, returns ZodType support chain call
func (s ZodCIDRv6Default) Optional() ZodType[any, any] {
	return Optional(any(s).(ZodType[any, any]))
}

// Nilable adds a nilable check to the CIDRv6 schema, returns ZodType support chain call
func (s ZodCIDRv6Default) Nilable() ZodType[any, any] {
	return Nilable(any(s).(ZodType[any, any]))
}
