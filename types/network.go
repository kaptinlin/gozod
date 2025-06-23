package types

import (
	"errors"
	"net"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/regexes"
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
	core.ZodTypeDef
	Type    core.ZodTypeCode // "ipv4"
	Version string           // "v4"
}

// ZodIPv4Internals contains IPv4 validator internal state
type ZodIPv4Internals struct {
	core.ZodTypeInternals
	Def  *ZodIPv4Def
	Isst issues.ZodIssueInvalidType
}

// ZodIPv4 represents an IPv4 address validation schema
type ZodIPv4 struct {
	internals *ZodIPv4Internals
}

// GetInternals returns the internal state of the schema
func (z *ZodIPv4) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// GetZod returns the IPv4-specific internals for framework usage
func (z *ZodIPv4) GetZod() *ZodIPv4Internals {
	return z.internals
}

// Parse validates the input value against the IPv4 schema
func (z *ZodIPv4) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	var parseCtx *core.ParseContext
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	// Use ParsePrimitive fast path with custom validator.
	return engine.ParsePrimitive[string](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeString,
		func(value string, checks []core.ZodCheck, ctx *core.ParseContext) error {
			// Execute any attached validation checks first (none by default)
			if len(checks) > 0 {
				payload := core.NewParsePayload(value)
				engine.RunChecksOnValue(value, checks, payload, ctx)
				if len(payload.GetIssues()) > 0 {
					return issues.NewZodError(issues.ConvertRawIssuesToIssues(payload.GetIssues(), ctx))
				}
			}

			// IPv4 format validation using regex + net.ParseIP
			if !regexes.IPv4.MatchString(value) {
				return issues.NewZodError([]core.ZodIssue{issues.FinalizeIssue(
					issues.CreateInvalidFormatIssue("ipv4", value, nil), ctx, core.GetConfig())})
			}

			ip := net.ParseIP(value)
			if ip == nil || ip.To4() == nil {
				return issues.NewZodError([]core.ZodIssue{issues.FinalizeIssue(
					issues.CreateInvalidFormatIssue("ipv4", value, nil), ctx, core.GetConfig())})
			}
			return nil
		},
		parseCtx,
	)
}

// MustParse validates the input value and panics on failure
func (z *ZodIPv4) MustParse(input any, ctx ...*core.ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Refine adds custom validation to the IPv4 schema
func (z *ZodIPv4) RefineAny(fn func(any) bool, params ...any) core.ZodType[any, any] {
	return Custom(fn, params...)
}

// Check adds modern validation using direct payload access
func (z *ZodIPv4) Check(fn core.CheckFn) core.ZodType[any, any] {
	custom := Custom(fn, core.SchemaParams{})
	custom.GetInternals().Parse = func(payload *core.ParsePayload, ctx *core.ParseContext) *core.ParsePayload {
		// first execute the original parse
		result, err := z.Parse(payload.GetValue(), ctx)
		if err != nil {
			payload.AddIssue(issues.CreateInvalidTypeWithMsg(core.ZodTypeIPv4, err.Error(), payload.GetValue()))
			return payload
		}
		payload.SetValue(result)

		// then execute the check function
		fn(payload)
		return payload
	}
	return any(custom).(core.ZodType[any, any])
}

// Optional makes the IPv4 address optional
func (z *ZodIPv4) Optional() core.ZodType[any, any] {
	return any(Optional(z)).(core.ZodType[any, any])
}

// Nilable makes the IPv4 address nilable
// - Only changes nil handling, not type inference
// - Uses standard Clone + SetNilable pattern like type_string.go
func (z *ZodIPv4) Nilable() core.ZodType[any, any] {
	// use Clone method to create a new instance, avoid manual state copying
	cloned := engine.Clone(z, func(def *core.ZodTypeDef) {
		// no need to modify def, because Nilable is a runtime flag
	}).(*ZodIPv4)
	cloned.internals.SetNilable()
	return cloned
}

// TransformAny creates a transform with given function
func (z *ZodIPv4) TransformAny(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	transform := Transform[any, any](fn)

	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: any(transform).(core.ZodType[any, any]),
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

// Pipe creates a pipe with another schema
func (z *ZodIPv4) Pipe(schema core.ZodType[any, any]) core.ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: schema,
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

// Prefault adds a prefault value
func (z *ZodIPv4) Prefault(value any) core.ZodType[any, any] {
	return Prefault(any(z).(core.ZodType[any, any]), value)
}

// PrefaultFunc adds a prefault value using a function
func (z *ZodIPv4) PrefaultFunc(fn func() any) core.ZodType[any, any] {
	return PrefaultFunc(any(z).(core.ZodType[any, any]), fn)
}

// createZodIPv4FromDef creates a ZodIPv4 from definition
func createZodIPv4FromDef(def *ZodIPv4Def) *ZodIPv4 {
	internals := &ZodIPv4Internals{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(core.ZodTypeIPv4),
		Def:              def,
		Isst:             issues.ZodIssueInvalidType{Expected: core.ZodTypeIPv4},
	}

	// Simplified parse: delegate to schema.Parse (which already includes IPv4 validation)
	internals.Parse = func(payload *core.ParsePayload, ctx *core.ParseContext) *core.ParsePayload {
		res, err := (&ZodIPv4{internals: internals}).Parse(payload.GetValue(), ctx)
		if err != nil {
			var zErr *issues.ZodError
			if errors.As(err, &zErr) {
				for _, issue := range zErr.Issues {
					rawIssue := issues.ConvertZodIssueToRaw(issue)
					rawIssue.Path = issue.Path
					payload.AddIssue(rawIssue)
				}
			}
			return payload
		}

		payload.SetValue(res)
		return payload
	}

	zodSchema := &ZodIPv4{internals: internals}

	// Set constructor for Clone support
	internals.Constructor = func(def *core.ZodTypeDef) core.ZodType[any, any] {
		ipv4Def := &ZodIPv4Def{
			ZodTypeDef: *def,
			Type:       core.ZodTypeIPv4,
			Version:    "v4",
		}
		return any(createZodIPv4FromDef(ipv4Def)).(core.ZodType[any, any])
	}

	// Initialize the schema
	engine.InitZodType(any(zodSchema).(core.ZodType[any, any]), &def.ZodTypeDef)

	return zodSchema
}

//////////////////////////////////////////
//////////   IPv6 Type           ////////
//////////////////////////////////////////

// ZodIPv6Def defines the configuration for IPv6 address validation
type ZodIPv6Def struct {
	core.ZodTypeDef
	Type    core.ZodTypeCode // "ipv6"
	Version string           // "v6"
}

// ZodIPv6Internals contains IPv6 validator internal state
type ZodIPv6Internals struct {
	core.ZodTypeInternals
	Def  *ZodIPv6Def
	Isst issues.ZodIssueInvalidType
}

// ZodIPv6 represents an IPv6 address validation schema
type ZodIPv6 struct {
	internals *ZodIPv6Internals
}

// GetInternals returns the internal state of the schema
func (z *ZodIPv6) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// GetZod returns the IPv6-specific internals for framework usage
func (z *ZodIPv6) GetZod() *ZodIPv6Internals {
	return z.internals
}

// Parse validates the input value against the IPv6 schema
func (z *ZodIPv6) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	var parseCtx *core.ParseContext
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	// Use ParsePrimitive fast path with custom validator.
	return engine.ParsePrimitive[string](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeString,
		func(value string, checks []core.ZodCheck, ctx *core.ParseContext) error {
			// Execute any attached validation checks first
			if len(checks) > 0 {
				payload := core.NewParsePayload(value)
				engine.RunChecksOnValue(value, checks, payload, ctx)
				if len(payload.GetIssues()) > 0 {
					return issues.NewZodError(issues.ConvertRawIssuesToIssues(payload.GetIssues(), ctx))
				}
			}

			// IPv6 format validation with regex + net.ParseIP (must be IPv6, not IPv4)
			if !regexes.IPv6.MatchString(value) {
				return issues.NewZodError([]core.ZodIssue{issues.FinalizeIssue(
					issues.CreateInvalidFormatIssue("ipv6", value, nil), ctx, core.GetConfig())})
			}

			ip := net.ParseIP(value)
			if ip == nil || ip.To4() != nil { // To4 != nil indicates IPv4
				return issues.NewZodError([]core.ZodIssue{issues.FinalizeIssue(
					issues.CreateInvalidFormatIssue("ipv6", value, nil), ctx, core.GetConfig())})
			}

			return nil
		},
		parseCtx,
	)
}

// MustParse validates the input value and panics on failure
func (z *ZodIPv6) MustParse(input any, ctx ...*core.ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Refine adds custom validation to the IPv6 schema
func (z *ZodIPv6) RefineAny(fn func(any) bool, params ...any) core.ZodType[any, any] {
	return Custom(fn, params...)
}

// Check adds modern validation using direct payload access
func (z *ZodIPv6) Check(fn core.CheckFn) core.ZodType[any, any] {
	custom := Custom(fn, core.SchemaParams{})
	custom.GetInternals().Parse = func(payload *core.ParsePayload, ctx *core.ParseContext) *core.ParsePayload {
		// first execute the original parse
		result, err := z.Parse(payload.GetValue(), ctx)
		if err != nil {
			payload.AddIssue(issues.CreateInvalidTypeWithMsg(core.ZodTypeIPv6, err.Error(), payload.GetValue()))
			return payload
		}
		payload.SetValue(result)

		// then execute the check function
		fn(payload)
		return payload
	}
	return any(custom).(core.ZodType[any, any])
}

// Optional makes the IPv6 address optional
func (z *ZodIPv6) Optional() core.ZodType[any, any] {
	return any(Optional(z)).(core.ZodType[any, any])
}

// Nilable makes the IPv6 address nilable
// - Only changes nil handling, not type inference
// - Uses standard Clone + SetNilable pattern like type_string.go
func (z *ZodIPv6) Nilable() core.ZodType[any, any] {
	// use Clone method to create a new instance, avoid manual state copying
	cloned := engine.Clone(z, func(def *core.ZodTypeDef) {
		// no need to modify def, because Nilable is a runtime flag
	}).(*ZodIPv6)
	cloned.internals.SetNilable()
	return cloned
}

// Transform creates a transform with given function
func (z *ZodIPv6) Transform(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return z.TransformAny(fn)
}

// TransformAny creates a transform with given function
func (z *ZodIPv6) TransformAny(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	transform := Transform[any, any](fn)

	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: any(transform).(core.ZodType[any, any]),
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

// Pipe creates a pipe with another schema
func (z *ZodIPv6) Pipe(schema core.ZodType[any, any]) core.ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: schema,
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

// Prefault adds a prefault value
func (z *ZodIPv6) Prefault(value any) core.ZodType[any, any] {
	return Prefault(any(z).(core.ZodType[any, any]), value)
}

// PrefaultFunc adds a prefault value using a function
func (z *ZodIPv6) PrefaultFunc(fn func() any) core.ZodType[any, any] {
	return PrefaultFunc(any(z).(core.ZodType[any, any]), fn)
}

// createZodIPv6FromDef creates a ZodIPv6 from definition
func createZodIPv6FromDef(def *ZodIPv6Def) *ZodIPv6 {
	internals := &ZodIPv6Internals{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(core.ZodTypeIPv6),
		Def:              def,
		Isst:             issues.ZodIssueInvalidType{Expected: core.ZodTypeIPv6},
	}

	// Simplified parse: delegate to schema.Parse (which already includes IPv6 validation)
	internals.Parse = func(payload *core.ParsePayload, ctx *core.ParseContext) *core.ParsePayload {
		res, err := (&ZodIPv6{internals: internals}).Parse(payload.GetValue(), ctx)
		if err != nil {
			var zErr *issues.ZodError
			if errors.As(err, &zErr) {
				for _, issue := range zErr.Issues {
					rawIssue := issues.ConvertZodIssueToRaw(issue)
					rawIssue.Path = issue.Path
					payload.AddIssue(rawIssue)
				}
			}
			return payload
		}

		payload.SetValue(res)
		return payload
	}

	zodSchema := &ZodIPv6{internals: internals}

	// Set constructor for Clone support
	internals.Constructor = func(def *core.ZodTypeDef) core.ZodType[any, any] {
		ipv6Def := &ZodIPv6Def{
			ZodTypeDef: *def,
			Type:       "ipv6",
			Version:    "v6",
		}
		return any(createZodIPv6FromDef(ipv6Def)).(core.ZodType[any, any])
	}

	// Initialize the schema
	engine.InitZodType(any(zodSchema).(core.ZodType[any, any]), &def.ZodTypeDef)

	return zodSchema
}

//////////////////////////////////////////
//////////   CIDRv4 Type         ////////
//////////////////////////////////////////

// ZodCIDRv4Def defines the configuration for IPv4 CIDR validation
type ZodCIDRv4Def struct {
	core.ZodTypeDef
	Type    core.ZodTypeCode // "cidrv4"
	Version string           // "v4"
}

// ZodCIDRv4Internals contains IPv4 CIDR validator internal state
type ZodCIDRv4Internals struct {
	core.ZodTypeInternals
	Def  *ZodCIDRv4Def
	Isst issues.ZodIssueInvalidType
}

// ZodCIDRv4 represents an IPv4 CIDR validation schema
type ZodCIDRv4 struct {
	internals *ZodCIDRv4Internals
}

// GetInternals returns the internal state of the schema
func (z *ZodCIDRv4) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// GetZod returns the CIDRv4-specific internals for framework usage
func (z *ZodCIDRv4) GetZod() *ZodCIDRv4Internals {
	return z.internals
}

// Parse smart type inference validation and parsing
func (z *ZodCIDRv4) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	// use unified nil handling logic
	if input == nil {
		if !z.internals.Nilable {
			rawIssue := issues.CreateInvalidTypeIssue("string", input)
			finalIssue := issues.FinalizeIssue(rawIssue, nil, core.GetConfig())
			return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
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
				rawIssue := issues.CreateInvalidTypeIssue("string", input)
				finalIssue := issues.FinalizeIssue(rawIssue, nil, core.GetConfig())
				return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
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
		rawIssue := issues.CreateInvalidTypeIssue("string", input)
		finalIssue := issues.FinalizeIssue(rawIssue, nil, core.GetConfig())
		return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
	}
}

// validateCIDRv4 validates IPv4 CIDR format
func (z *ZodCIDRv4) validateCIDRv4(value string) error {
	// Use net.ParseCIDR for accurate CIDR validation
	_, _, err := net.ParseCIDR(value)
	if err != nil {
		rawIssue := issues.CreateInvalidFormatIssue("cidrv4", value, nil)
		finalIssue := issues.FinalizeIssue(rawIssue, nil, core.GetConfig())
		return issues.NewZodError([]core.ZodIssue{finalIssue})
	}
	return nil
}

// MustParse validates the input value and panics on failure
func (z *ZodCIDRv4) MustParse(input any, ctx ...*core.ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Refine adds custom validation to the CIDRv4 schema
func (z *ZodCIDRv4) RefineAny(fn func(any) bool, params ...any) core.ZodType[any, any] {
	return Custom(fn, params...)
}

// Check adds modern validation using direct payload access
func (z *ZodCIDRv4) Check(fn core.CheckFn) core.ZodType[any, any] {
	custom := Custom(fn, core.SchemaParams{})
	custom.GetInternals().Parse = func(payload *core.ParsePayload, ctx *core.ParseContext) *core.ParsePayload {
		// first execute the original parse
		result, err := z.Parse(payload.GetValue(), ctx)
		if err != nil {
			payload.AddIssue(issues.CreateInvalidTypeWithMsg(core.ZodTypeCIDRv4, err.Error(), payload.GetValue()))
			return payload
		}
		payload.SetValue(result)

		// then execute the check function
		fn(payload)
		return payload
	}
	return any(custom).(core.ZodType[any, any])
}

// Optional makes the CIDRv4 optional
func (z *ZodCIDRv4) Optional() core.ZodType[any, any] {
	return any(Optional(z)).(core.ZodType[any, any])
}

// Nilable makes the CIDRv4 nilable
// Nilable makes the CIDRv4 address nilable
// - Only changes nil handling, not type inference
// - Uses standard Clone + SetNilable pattern like type_string.go
func (z *ZodCIDRv4) Nilable() core.ZodType[any, any] {
	// use Clone method to create a new instance, avoid manual state copying
	cloned := engine.Clone(z, func(def *core.ZodTypeDef) {
		// no need to modify def, because Nilable is a runtime flag
	}).(*ZodCIDRv4)
	cloned.internals.SetNilable()
	return cloned
}

// Transform creates a transform with given function
func (z *ZodCIDRv4) Transform(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return z.TransformAny(fn)
}

// TransformAny creates a transform with given function
func (z *ZodCIDRv4) TransformAny(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	transform := Transform[any, any](fn)

	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: any(transform).(core.ZodType[any, any]),
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

// Pipe creates a pipe with another schema
func (z *ZodCIDRv4) Pipe(schema core.ZodType[any, any]) core.ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: schema,
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

// Prefault adds a prefault value
func (z *ZodCIDRv4) Prefault(value any) core.ZodType[any, any] {
	return Prefault(any(z).(core.ZodType[any, any]), value)
}

// PrefaultFunc adds a prefault value using a function
func (z *ZodCIDRv4) PrefaultFunc(fn func() any) core.ZodType[any, any] {
	return PrefaultFunc(any(z).(core.ZodType[any, any]), fn)
}

// createZodCIDRv4FromDef creates a ZodCIDRv4 from definition
func createZodCIDRv4FromDef(def *ZodCIDRv4Def) *ZodCIDRv4 {
	internals := &ZodCIDRv4Internals{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(core.ZodTypeCIDRv4),
		Def:              def,
		Isst:             issues.ZodIssueInvalidType{Expected: core.ZodTypeCIDRv4},
	}

	// Simplified parse: delegate to schema.Parse (which already includes CIDRv4 validation)
	internals.Parse = func(payload *core.ParsePayload, ctx *core.ParseContext) *core.ParsePayload {
		res, err := (&ZodCIDRv4{internals: internals}).Parse(payload.GetValue(), ctx)
		if err != nil {
			var zErr *issues.ZodError
			if errors.As(err, &zErr) {
				for _, issue := range zErr.Issues {
					rawIssue := issues.ConvertZodIssueToRaw(issue)
					rawIssue.Path = issue.Path
					payload.AddIssue(rawIssue)
				}
			}
			return payload
		}

		payload.SetValue(res)
		return payload
	}

	zodSchema := &ZodCIDRv4{internals: internals}

	// Set constructor for Clone support
	internals.Constructor = func(def *core.ZodTypeDef) core.ZodType[any, any] {
		cidrv4Def := &ZodCIDRv4Def{
			ZodTypeDef: *def,
			Type:       "cidrv4",
			Version:    "v4",
		}
		return any(createZodCIDRv4FromDef(cidrv4Def)).(core.ZodType[any, any])
	}

	// Initialize the schema
	engine.InitZodType(any(zodSchema).(core.ZodType[any, any]), &def.ZodTypeDef)

	return zodSchema
}

//////////////////////////////////////////
//////////   CIDRv6 Type         ////////
//////////////////////////////////////////

// ZodCIDRv6Def defines the configuration for IPv6 CIDR validation
type ZodCIDRv6Def struct {
	core.ZodTypeDef
	Type    core.ZodTypeCode // "cidrv6"
	Version string           // "v6"
}

// ZodCIDRv6Internals contains IPv6 CIDR validator internal state
type ZodCIDRv6Internals struct {
	core.ZodTypeInternals
	Def  *ZodCIDRv6Def
	Isst issues.ZodIssueInvalidType
}

// ZodCIDRv6 represents an IPv6 CIDR validation schema
type ZodCIDRv6 struct {
	internals *ZodCIDRv6Internals
}

// GetInternals returns the internal state of the schema
func (z *ZodCIDRv6) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// GetZod returns the CIDRv6-specific internals for framework usage
func (z *ZodCIDRv6) GetZod() *ZodCIDRv6Internals {
	return z.internals
}

// Parse smart type inference validation and parsing
func (z *ZodCIDRv6) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	// use unified nil handling logic
	if input == nil {
		if !z.internals.Nilable {
			rawIssue := issues.CreateInvalidTypeIssue("string", input)
			finalIssue := issues.FinalizeIssue(rawIssue, nil, core.GetConfig())
			return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
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
				rawIssue := issues.CreateInvalidTypeIssue("string", input)
				finalIssue := issues.FinalizeIssue(rawIssue, nil, core.GetConfig())
				return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
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
		rawIssue := issues.CreateInvalidTypeIssue("string", input)
		finalIssue := issues.FinalizeIssue(rawIssue, nil, core.GetConfig())
		return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
	}
}

// validateCIDRv6 validates IPv6 CIDR format
func (z *ZodCIDRv6) validateCIDRv6(value string) error {
	// Use net.ParseCIDR for accurate CIDR validation
	ip, _, err := net.ParseCIDR(value)
	if err != nil {
		rawIssue := issues.CreateInvalidFormatIssue("cidrv6", value, nil)
		finalIssue := issues.FinalizeIssue(rawIssue, nil, core.GetConfig())
		return issues.NewZodError([]core.ZodIssue{finalIssue})
	}
	// Ensure it's actually IPv6
	if ip.To4() != nil {
		rawIssue := issues.CreateInvalidFormatIssue("cidrv6", value, nil)
		finalIssue := issues.FinalizeIssue(rawIssue, nil, core.GetConfig())
		return issues.NewZodError([]core.ZodIssue{finalIssue})
	}
	return nil
}

// MustParse validates the input value and panics on failure
func (z *ZodCIDRv6) MustParse(input any, ctx ...*core.ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Refine adds custom validation to the CIDRv6 schema
func (z *ZodCIDRv6) RefineAny(fn func(any) bool, params ...any) core.ZodType[any, any] {
	return Custom(core.RefineFn[any](fn), params...)
}

// Check adds modern validation using direct payload access
func (z *ZodCIDRv6) Check(fn core.CheckFn) core.ZodType[any, any] {
	custom := Custom(fn, core.SchemaParams{})
	custom.GetInternals().Parse = func(payload *core.ParsePayload, ctx *core.ParseContext) *core.ParsePayload {
		// first execute the original parse
		result, err := z.Parse(payload.GetValue(), ctx)
		if err != nil {
			payload.AddIssue(issues.CreateInvalidTypeWithMsg(core.ZodTypeCIDRv6, err.Error(), payload.GetValue()))
			return payload
		}
		payload.SetValue(result)

		// then execute the check function
		fn(payload)
		return payload
	}
	return any(custom).(core.ZodType[any, any])
}

// Optional makes the CIDRv6 optional
func (z *ZodCIDRv6) Optional() core.ZodType[any, any] {
	return any(Optional(z)).(core.ZodType[any, any])
}

// Nilable makes the CIDRv6 address nilable
// - Only changes nil handling, not type inference
// - Uses standard Clone + SetNilable pattern like type_string.go
func (z *ZodCIDRv6) Nilable() core.ZodType[any, any] {
	// use Clone method to create a new instance, avoid manual state copying
	cloned := engine.Clone(z, func(def *core.ZodTypeDef) {
		// no need to modify def, because Nilable is a runtime flag
	}).(*ZodCIDRv6)
	cloned.internals.SetNilable()
	return cloned
}

// Transform creates a transform with given function
func (z *ZodCIDRv6) Transform(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return z.TransformAny(fn)
}

// TransformAny creates a transform with given function
func (z *ZodCIDRv6) TransformAny(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	transform := Transform[any, any](fn)

	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: any(transform).(core.ZodType[any, any]),
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

// Pipe creates a pipe with another schema
func (z *ZodCIDRv6) Pipe(schema core.ZodType[any, any]) core.ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: schema,
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

// createZodCIDRv6FromDef creates a ZodCIDRv6 from definition
func createZodCIDRv6FromDef(def *ZodCIDRv6Def) *ZodCIDRv6 {
	internals := &ZodCIDRv6Internals{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(core.ZodTypeCIDRv6),
		Def:              def,
		Isst:             issues.ZodIssueInvalidType{Expected: core.ZodTypeCIDRv6},
	}

	// Simplified parse: delegate to schema.Parse (which already includes CIDRv6 validation)
	internals.Parse = func(payload *core.ParsePayload, ctx *core.ParseContext) *core.ParsePayload {
		res, err := (&ZodCIDRv6{internals: internals}).Parse(payload.GetValue(), ctx)
		if err != nil {
			var zErr *issues.ZodError
			if errors.As(err, &zErr) {
				for _, issue := range zErr.Issues {
					rawIssue := issues.ConvertZodIssueToRaw(issue)
					rawIssue.Path = issue.Path
					payload.AddIssue(rawIssue)
				}
			}
			return payload
		}

		payload.SetValue(res)
		return payload
	}

	zodSchema := &ZodCIDRv6{internals: internals}

	// Set constructor for Clone support
	internals.Constructor = func(def *core.ZodTypeDef) core.ZodType[any, any] {
		cidrv6Def := &ZodCIDRv6Def{
			ZodTypeDef: *def,
			Type:       "cidrv6",
			Version:    "v6",
		}
		return any(createZodCIDRv6FromDef(cidrv6Def)).(core.ZodType[any, any])
	}

	// Initialize the schema
	engine.InitZodType(any(zodSchema).(core.ZodType[any, any]), &def.ZodTypeDef)

	return zodSchema
}

//////////////////////////////////////////
//////////   Public Constructors   //////
//////////////////////////////////////////

// IPv4 creates a new IPv4 address schema
func IPv4(params ...any) *ZodIPv4 {
	def := &ZodIPv4Def{
		ZodTypeDef: core.ZodTypeDef{Type: core.ZodTypeIPv4},
		Type:       core.ZodTypeIPv4,
		Version:    "v4",
	}

	schema := createZodIPv4FromDef(def)

	// Apply schema parameters
	if len(params) > 0 {
		if schemaParam, ok := params[0].(core.SchemaParams); ok {
			if schemaParam.Error != nil {
				errorMap := issues.CreateErrorMap(schemaParam.Error)
				if errorMap != nil {
					def.Error = errorMap
					schema.internals.Error = errorMap
				}
			}
		}
	}

	return schema
}

// IPv6 creates a new IPv6 address schema
func IPv6(params ...any) *ZodIPv6 {
	def := &ZodIPv6Def{
		ZodTypeDef: core.ZodTypeDef{Type: core.ZodTypeIPv6},
		Type:       core.ZodTypeIPv6,
		Version:    "v6",
	}

	schema := createZodIPv6FromDef(def)

	// Apply schema parameters
	if len(params) > 0 {
		if schemaParam, ok := params[0].(core.SchemaParams); ok {
			if schemaParam.Error != nil {
				errorMap := issues.CreateErrorMap(schemaParam.Error)
				if errorMap != nil {
					def.Error = errorMap
					schema.internals.Error = errorMap
				}
			}
		}
	}

	return schema
}

// CIDRv4 creates a new IPv4 CIDR schema
func CIDRv4(params ...any) *ZodCIDRv4 {
	def := &ZodCIDRv4Def{
		ZodTypeDef: core.ZodTypeDef{Type: core.ZodTypeCIDRv4},
		Type:       core.ZodTypeCIDRv4,
		Version:    "v4",
	}

	schema := createZodCIDRv4FromDef(def)

	// Apply schema parameters
	if len(params) > 0 {
		if schemaParam, ok := params[0].(core.SchemaParams); ok {
			if schemaParam.Error != nil {
				errorMap := issues.CreateErrorMap(schemaParam.Error)
				if errorMap != nil {
					def.Error = errorMap
					schema.internals.Error = errorMap
				}
			}
		}
	}

	return schema
}

// CIDRv6 creates a new IPv6 CIDR schema
func CIDRv6(params ...any) *ZodCIDRv6 {
	def := &ZodCIDRv6Def{
		ZodTypeDef: core.ZodTypeDef{Type: core.ZodTypeCIDRv6},
		Type:       core.ZodTypeCIDRv6,
		Version:    "v6",
	}

	schema := createZodCIDRv6FromDef(def)

	// Apply schema parameters
	if len(params) > 0 {
		if schemaParam, ok := params[0].(core.SchemaParams); ok {
			if schemaParam.Error != nil {
				errorMap := issues.CreateErrorMap(schemaParam.Error)
				if errorMap != nil {
					def.Error = errorMap
					schema.internals.Error = errorMap
				}
			}
		}
	}

	return schema
}

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodIPv4) Unwrap() core.ZodType[any, any] {
	return any(z).(core.ZodType[any, any])
}

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodIPv6) Unwrap() core.ZodType[any, any] {
	return any(z).(core.ZodType[any, any])
}

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodCIDRv4) Unwrap() core.ZodType[any, any] {
	return any(z).(core.ZodType[any, any])
}

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodCIDRv6) Unwrap() core.ZodType[any, any] {
	return any(z).(core.ZodType[any, any])
}

// ==================== NETWORK TYPES DEFAULT WRAPPER SYSTEM ====================

// ZodIPv4Default is the Default wrapper for IPv4 type
type ZodIPv4Default struct {
	*ZodDefault[*ZodIPv4] // embed specific pointer, allow method promotion
}

// GetInternals implements ZodType interface
func (s ZodIPv4Default) GetInternals() *core.ZodTypeInternals {
	return s.ZodDefault.GetInternals()
}

// Parse implements ZodType interface
func (s ZodIPv4Default) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	return s.ZodDefault.Parse(input, ctx...)
}

// ZodIPv6Default is the Default wrapper for IPv6 type
type ZodIPv6Default struct {
	*ZodDefault[*ZodIPv6] // embed specific pointer, allow method promotion
}

// GetInternals implements ZodType interface
func (s ZodIPv6Default) GetInternals() *core.ZodTypeInternals {
	return s.ZodDefault.GetInternals()
}

// Parse implements ZodType interface
func (s ZodIPv6Default) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	return s.ZodDefault.Parse(input, ctx...)
}

// ZodCIDRv4Default is the Default wrapper for CIDRv4 type
type ZodCIDRv4Default struct {
	*ZodDefault[*ZodCIDRv4] // embed specific pointer, allow method promotion
}

// GetInternals implements ZodType interface
func (s ZodCIDRv4Default) GetInternals() *core.ZodTypeInternals {
	return s.ZodDefault.GetInternals()
}

// Parse implements ZodType interface
func (s ZodCIDRv4Default) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	return s.ZodDefault.Parse(input, ctx...)
}

// ZodCIDRv6Default is the Default wrapper for CIDRv6 type
type ZodCIDRv6Default struct {
	*ZodDefault[*ZodCIDRv6] // embed specific pointer, allow method promotion
}

// GetInternals implements ZodType interface
func (s ZodCIDRv6Default) GetInternals() *core.ZodTypeInternals {
	return s.ZodDefault.GetInternals()
}

// Parse implements ZodType interface
func (s ZodCIDRv6Default) Parse(input any, ctx ...*core.ParseContext) (any, error) {
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
func (s ZodIPv4Default) Optional() core.ZodType[any, any] {
	return Optional(any(s).(core.ZodType[any, any]))
}

// Nilable adds a nilable check to the IPv4 schema, returns ZodType support chain call
func (s ZodIPv4Default) Nilable() core.ZodType[any, any] {
	return Nilable(any(s).(core.ZodType[any, any]))
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
func (s ZodIPv6Default) Optional() core.ZodType[any, any] {
	return Optional(any(s).(core.ZodType[any, any]))
}

// Nilable adds a nilable check to the IPv6 schema, returns ZodType support chain call
func (s ZodIPv6Default) Nilable() core.ZodType[any, any] {
	return Nilable(any(s).(core.ZodType[any, any]))
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
func (s ZodCIDRv4Default) Optional() core.ZodType[any, any] {
	return Optional(any(s).(core.ZodType[any, any]))
}

// Nilable adds a nilable check to the CIDRv4 schema, returns ZodType support chain call
func (s ZodCIDRv4Default) Nilable() core.ZodType[any, any] {
	return Nilable(any(s).(core.ZodType[any, any]))
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
func (s ZodCIDRv6Default) Optional() core.ZodType[any, any] {
	return Optional(any(s).(core.ZodType[any, any]))
}

// Nilable adds a nilable check to the CIDRv6 schema, returns ZodType support chain call
func (s ZodCIDRv6Default) Nilable() core.ZodType[any, any] {
	return Nilable(any(s).(core.ZodType[any, any]))
}
