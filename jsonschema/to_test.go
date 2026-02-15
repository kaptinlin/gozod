package jsonschema

import (
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/go-json-experiment/json"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertJSONEquals(t *testing.T, expected string, actualJSON string) {
	t.Helper()

	var expectedVal, actualVal any

	err := json.Unmarshal([]byte(expected), &expectedVal)
	require.NoError(t, err, "Failed to unmarshal expected JSON")

	err = json.Unmarshal([]byte(actualJSON), &actualVal)
	require.NoError(t, err, "Failed to unmarshal actual JSON")

	if !isSubset(expectedVal, actualVal) {
		assert.Equal(t, expectedVal, actualVal)
	}
}

// isSubset recursively verifies that exp is a subset of act (i.e., all keys/values in exp are present in act).
func isSubset(exp, act any) bool {
	switch e := exp.(type) {
	case map[string]any:
		a, ok := act.(map[string]any)
		if !ok {
			return false
		}
		for k, v := range e {
			av, exists := a[k]
			if !exists {
				return false
			}
			if !isSubset(v, av) {
				return false
			}
		}
		return true
	case []any:
		a, ok := act.([]any)
		if !ok || len(e) != len(a) {
			return false
		}
		for i := range e {
			if !isSubset(e[i], a[i]) {
				return false
			}
		}
		return true
	default:
		return reflect.DeepEqual(exp, act)
	}
}

// =============================================================================
// PRIMITIVE TYPES
// =============================================================================

func TestToJSONSchema_PrimitiveTypes(t *testing.T) {
	testCases := []struct {
		name     string
		schema   core.ZodSchema
		expected string
	}{
		{
			name:     "String",
			schema:   types.String(),
			expected: `{"type":"string"}`,
		},
		{
			name:     "Number",
			schema:   types.Float(),
			expected: `{"type":"number"}`,
		},
		{
			name:     "Boolean",
			schema:   types.Bool(),
			expected: `{"type":"boolean"}`,
		},
		{
			name:     "Null",
			schema:   types.Nil(),
			expected: `{"type":"null"}`,
		},
		{
			name:     "Any",
			schema:   types.Any(),
			expected: `{}`,
		},
		{
			name:     "Unknown",
			schema:   types.Unknown(),
			expected: `{}`,
		},
		{
			name:     "Never",
			schema:   types.Never(),
			expected: `{"not":true}`,
		},
		{
			name:     "Integer",
			schema:   types.Int(),
			expected: `{"type":"integer","minimum":-9.223372036854776e+18,"maximum":9.223372036854776e+18}`,
		},
		{
			name:     "Int8",
			schema:   types.Int8(),
			expected: `{"type":"integer","minimum":-128,"maximum":127}`,
		},
		{
			name:     "Int16",
			schema:   types.Int16(),
			expected: `{"type":"integer","minimum":-32768,"maximum":32767}`,
		},
		{
			name:     "Int32",
			schema:   types.Int32(),
			expected: `{"type":"integer","minimum":-2147483648,"maximum":2147483647}`,
		},
		{
			name:     "Int64",
			schema:   types.Int64(),
			expected: `{"type":"integer","minimum":-9.223372036854776e+18,"maximum":9.223372036854776e+18}`,
		},
		{
			name:     "Uint",
			schema:   types.Uint(),
			expected: `{"type":"integer","minimum":0,"maximum":1.8446744073709552e+19}`,
		},
		{
			name:     "Uint8",
			schema:   types.Uint8(),
			expected: `{"type":"integer","minimum":0,"maximum":255}`,
		},
		{
			name:     "Uint16",
			schema:   types.Uint16(),
			expected: `{"type":"integer","minimum":0,"maximum":65535}`,
		},
		{
			name:     "Uint32",
			schema:   types.Uint32(),
			expected: `{"type":"integer","minimum":0,"maximum":4294967295}`,
		},
		{
			name:     "Uint64",
			schema:   types.Uint64(),
			expected: `{"type":"integer","minimum":0,"maximum":1.844674407371e+19}`,
		},
		{
			name:     "Float32",
			schema:   types.Float32(),
			expected: `{"type":"number","minimum":-3.4028234663852886e+38,"maximum":3.4028234663852886e+38}`,
		},
		{
			name:     "Float64",
			schema:   types.Float64(),
			expected: `{"type":"number","minimum":-1.7976931348623157e+308,"maximum":1.7976931348623157e+308}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonSchema, err := ToJSONSchema(tc.schema)
			assert.NoError(t, err)
			jsonSchemaBytes, err := json.Marshal(jsonSchema)
			assert.NoError(t, err)
			assertJSONEquals(t, tc.expected, string(jsonSchemaBytes))
		})
	}
}

// =============================================================================
// STRING FORMATS
// =============================================================================

func TestToJSONSchema_StringFormats(t *testing.T) {
	testCases := []struct {
		name     string
		schema   core.ZodSchema
		expected string
	}{
		{
			name:     "Email",
			schema:   types.Email(),
			expected: `{"type":"string", "format":"email", "pattern":"^[A-Za-z0-9_'+\\-]+([A-Za-z0-9_'+\\-]*\\.[A-Za-z0-9_'+\\-]+)*@[A-Za-z0-9]([A-Za-z0-9\\-]*[A-Za-z0-9])?(\\.[A-Za-z0-9]([A-Za-z0-9\\-]*[A-Za-z0-9])?)*\\.[A-Za-z]{2,}$"}`,
		},
		{
			name:     "UUID",
			schema:   types.UUID(),
			expected: `{"type":"string","format":"uuid","pattern":"^(?:[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-8][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}|00000000-0000-0000-0000-000000000000)$"}`,
		},
		{
			name:     "UUIDv4",
			schema:   types.Uuidv4(),
			expected: `{"type":"string","format":"uuid","pattern":"^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-4[0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$"}`,
		},
		{
			name:     "UUIDv6",
			schema:   types.Uuidv6(),
			expected: `{"type":"string","format":"uuid","pattern":"^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-6[0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$"}`,
		},
		{
			name:     "UUIDv7",
			schema:   types.Uuidv7(),
			expected: `{"type":"string","format":"uuid","pattern":"^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-7[0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$"}`,
		},
		{
			name:     "URL",
			schema:   types.URL(),
			expected: `{"type":"string","format":"uri","pattern":"^[a-zA-Z][a-zA-Z0-9+.-]*://[^\\s/$.?#].[^\\s]*$"}`,
		},
		{
			name:     "Base64",
			schema:   types.Base64(),
			expected: `{"type":"string","format":"base64","contentEncoding":"base64","pattern":"^$|^(?:[0-9a-zA-Z+/]{4})*(?:(?:[0-9a-zA-Z+/]{2}==)|(?:[0-9a-zA-Z+/]{3}=))?$"}`,
		},
		{
			name:     "Base64URL",
			schema:   types.Base64URL(),
			expected: `{"type":"string","format":"base64url","contentEncoding":"base64url","pattern":"^[A-Za-z0-9_-]*={0,2}$"}`,
		},
		{
			name:     "CUID",
			schema:   types.Cuid(),
			expected: `{"type":"string","format":"cuid","pattern":"^[cC][^\\s-]{8,}$"}`,
		},
		{
			name:     "CUID2",
			schema:   types.Cuid2(),
			expected: `{"type":"string","format":"cuid2","pattern":"^[0-9a-z]+$"}`,
		},
		{
			name:     "ULID",
			schema:   types.Ulid(),
			expected: `{"type":"string","format":"ulid","pattern":"^[0-9A-HJKMNP-TV-Za-hjkmnp-tv-z]{26}$"}`,
		},
		{
			name:     "XID",
			schema:   types.Xid(),
			expected: `{"type":"string","format":"xid","pattern":"^[0-9a-vA-V]{20}$"}`,
		},
		{
			name:     "KSUID",
			schema:   types.Ksuid(),
			expected: `{"type":"string","format":"ksuid","pattern":"^[A-Za-z0-9]{27}$"}`,
		},
		{
			name:     "NanoID",
			schema:   types.Nanoid(),
			expected: `{"type":"string","format":"nanoid","pattern":"^[a-zA-Z0-9_-]{21}$"}`,
		},
		{
			name:     "JWT",
			schema:   types.JWT(),
			expected: `{"type":"string","format":"jwt"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonSchema, err := ToJSONSchema(tc.schema)
			assert.NoError(t, err)
			jsonSchemaBytes, err := json.Marshal(jsonSchema)
			assert.NoError(t, err)
			assertJSONEquals(t, tc.expected, string(jsonSchemaBytes))
		})
	}
}

// =============================================================================
// NETWORK FORMATS
// =============================================================================

func TestToJSONSchema_NetworkFormats(t *testing.T) {
	testCases := []struct {
		name     string
		schema   core.ZodSchema
		expected string
	}{
		{
			name:     "IPv4",
			schema:   types.IPv4(),
			expected: `{"type":"string","format":"ipv4","pattern":"^(?:(?:25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9][0-9]|[0-9])\\.){3}(?:25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9][0-9]|[0-9])$"}`,
		},
		{
			name:     "IPv6",
			schema:   types.IPv6(),
			expected: `{"type":"string","format":"ipv6","pattern":"^(([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))$"}`,
		},
		{
			name:     "CIDRv4",
			schema:   types.CIDRv4(),
			expected: `{"type":"string","format":"cidrv4","pattern":"^((25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9][0-9]|[0-9])\\.){3}(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9][0-9]|[0-9])\\/([0-9]|[1-2][0-9]|3[0-2])$"}`,
		},
		{
			name:     "CIDRv6",
			schema:   types.CIDRv6(),
			expected: `{"type":"string","format":"cidrv6","pattern":"^(([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))\\/(12[0-8]|1[01][0-9]|[1-9]?[0-9])$"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonSchema, err := ToJSONSchema(tc.schema)
			assert.NoError(t, err)
			jsonSchemaBytes, err := json.Marshal(jsonSchema)
			assert.NoError(t, err)
			assertJSONEquals(t, tc.expected, string(jsonSchemaBytes))
		})
	}
}

// =============================================================================
// ISO 8601 FORMATS
// =============================================================================

func TestToJSONSchema_ISOFormats(t *testing.T) {
	testCases := []struct {
		name     string
		schema   core.ZodSchema
		expected string
	}{
		{
			name:     "ISO DateTime",
			schema:   types.IsoDateTime(),
			expected: `{"type":"string","format":"iso_datetime","pattern":"^(?:(?:\\d\\d[2468][048]|\\d\\d[13579][26]|\\d\\d0[48]|[02468][048]00|[13579][26]00)-02-29|\\d{4}-(?:(?:0[13578]|1[02])-(?:0[1-9]|[12]\\d|3[01])|(?:0[469]|11)-(?:0[1-9]|[12]\\d|30)|(?:02)-(?:0[1-9]|1\\d|2[0-8])))T(?:(?:[01]\\d|2[0-3]):[0-5]\\d(?::[0-5]\\d(?:\\.\\d+)?)?(?:Z|[+-](?:[01]\\d|2[0-3]):[0-5]\\d))$"}`,
		},
		{
			name:     "ISO Date",
			schema:   types.IsoDate(),
			expected: `{"type":"string","format":"iso_date","pattern":"^(?:(?:\\d\\d[2468][048]|\\d\\d[13579][26]|\\d\\d0[48]|[02468][048]00|[13579][26]00)-02-29|\\d{4}-(?:(?:0[13578]|1[02])-(?:0[1-9]|[12]\\d|3[01])|(?:0[469]|11)-(?:0[1-9]|[12]\\d|30)|(?:02)-(?:0[1-9]|1\\d|2[0-8])))$"}`,
		},
		{
			name:     "ISO Time",
			schema:   types.IsoTime(),
			expected: `{"type":"string","format":"iso_time","pattern":"^(?:[01]\\d|2[0-3]):[0-5]\\d(?::[0-5]\\d(?:\\.\\d+)?)?$"}`,
		},
		{
			name:     "ISO Duration",
			schema:   types.IsoDuration(),
			expected: `{"type":"string","format":"iso_duration","pattern":"^P(?:(\\d+W)|(\\d+Y)?(\\d+M)?(\\d+D)?(?:T(\\d+H)?(\\d+M)?(\\d+(?:[.,]\\d+)?S)?)?)$"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonSchema, err := ToJSONSchema(tc.schema)
			assert.NoError(t, err)
			jsonSchemaBytes, err := json.Marshal(jsonSchema)
			assert.NoError(t, err)
			assertJSONEquals(t, tc.expected, string(jsonSchemaBytes))
		})
	}
}

// =============================================================================
// FILE TYPES
// =============================================================================

func TestToJSONSchema_FileTypes(t *testing.T) {
	testCases := []struct {
		name     string
		schema   core.ZodSchema
		expected string
	}{
		{
			name:     "File",
			schema:   types.File(),
			expected: `{"type":"string","format":"binary","contentEncoding":"binary"}`,
		},
		{
			name:     "File with Mime and Size",
			schema:   types.File().Mime([]string{"image/png"}).Min(1000).Max(10000),
			expected: `{"type":"string","format":"binary","contentEncoding":"binary","contentMediaType":"image/png","minLength":1000,"maxLength":10000}`,
		},
		{
			name:   "File with multiple Mime types",
			schema: types.File().Mime([]string{"image/png", "image/jpeg"}).Min(1000).Max(10000),
			expected: `{
				"anyOf": [
					{
						"type": "string",
						"format": "binary",
						"contentEncoding": "binary",
						"contentMediaType": "image/png",
						"minLength": 1000,
						"maxLength": 10000
					},
					{
						"type": "string",
						"format": "binary",
						"contentEncoding": "binary",
						"contentMediaType": "image/jpeg",
						"minLength": 1000,
						"maxLength": 10000
					}
				]
			}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonSchema, err := ToJSONSchema(tc.schema)
			assert.NoError(t, err)
			jsonSchemaBytes, err := json.Marshal(jsonSchema)
			assert.NoError(t, err)
			assertJSONEquals(t, tc.expected, string(jsonSchemaBytes))
		})
	}
}

// =============================================================================
// UNSUPPORTED TYPES
// =============================================================================

func TestToJSONSchema_UnsupportedTypes(t *testing.T) {
	unsupported := []struct {
		name   string
		schema core.ZodSchema
	}{
		{"BigInt", types.BigInt()},
		{"BigIntPtr", types.BigIntPtr()},
		{"Complex", types.Complex()},
		{"ComplexPtr", types.ComplexPtr()},
		{"Complex64", types.Complex64()},
		{"Complex64Ptr", types.Complex64Ptr()},
		{"Complex128", types.Complex128()},
		{"Complex128Ptr", types.Complex128Ptr()},
		{"Function", types.Function()},
		{"FunctionPtr", types.FunctionPtr()},
		{
			"Transform",
			types.String().Transform(func(s string, ctx *core.RefinementContext) (any, error) {
				return len(s), nil
			}),
		},
	}

	for _, u := range unsupported {
		t.Run(u.name, func(t *testing.T) {
			_, err := ToJSONSchema(u.schema)
			assert.Error(t, err)
		})
	}
}

func TestToJSONSchema_SupportedTypes(t *testing.T) {
	supported := []struct {
		name     string
		schema   core.ZodSchema
		expected string
	}{
		{"Time", types.Time(), `{"type":"string", "format":"time"}`},
		{"TimePtr", types.TimePtr(), `{"type":"string", "format":"time"}`},
		{"Map", types.Map(types.String(), types.Int()), `{"type":"object", "additionalProperties":{"type":"integer"}}`},
		{"MapPtr", types.MapPtr(types.String(), types.Int()), `{"type":"object", "additionalProperties":{"type":"integer"}}`},
	}

	for _, u := range supported {
		t.Run(u.name, func(t *testing.T) {
			js, err := ToJSONSchema(u.schema)
			assert.NoError(t, err)
			jsonSchemaBytes, err := json.Marshal(js)
			assert.NoError(t, err)
			assertJSONEquals(t, u.expected, string(jsonSchemaBytes))
		})
	}
}

// =============================================================================
// NUMBER CONSTRAINTS
// =============================================================================

func TestToJSONSchema_NumberConstraints(t *testing.T) {
	cases := []struct {
		name     string
		schema   core.ZodSchema
		expected string
	}{
		// Basic Float constraints (matching TypeScript z.number())
		{"MinMax", types.Float().Min(5).Max(10), `{"type":"number","minimum":5,"maximum":10}`},
		{"GtGt", types.Float().Gt(5).Gt(10), `{"type":"number","exclusiveMinimum":10}`},
		{"GtGte", types.Float().Gt(5).Gte(10), `{"type":"number","minimum":10}`},
		{"LtLt", types.Float().Lt(5).Lt(3), `{"type":"number","exclusiveMaximum":3}`},
		{"LtLtLte", types.Float().Lt(5).Lt(3).Lte(2), `{"type":"number","maximum":2}`},
		{"LtLte", types.Float().Lt(5).Lte(3), `{"type":"number","maximum":3}`},
		{"GtLt", types.Float().Gt(5).Lt(10), `{"type":"number","exclusiveMinimum":5,"exclusiveMaximum":10}`},
		{"GteLte", types.Float().Gte(5).Lte(10), `{"type":"number","minimum":5,"maximum":10}`},
		{"Positive", types.Float().Positive(), `{"type":"number","exclusiveMinimum":0}`},
		{"Negative", types.Float().Negative(), `{"type":"number","exclusiveMaximum":0}`},
		{"NonPositive", types.Float().NonPositive(), `{"type":"number","maximum":0}`},
		{"NonNegative", types.Float().NonNegative(), `{"type":"number","minimum":0}`},

		// Integer constraints (matching TypeScript z.int())
		{"IntegerMinMax", types.Int().Min(5).Max(10), `{"type":"integer","minimum":5,"maximum":10}`},
		{"IntegerGtGt", types.Int().Gt(5).Gt(10), `{"type":"integer","exclusiveMinimum":10}`},
		{"IntegerGtGte", types.Int().Gt(5).Gte(10), `{"type":"integer","minimum":10}`},
		{"IntegerLtLt", types.Int().Lt(5).Lt(3), `{"type":"integer","exclusiveMaximum":3}`},
		{"IntegerLtLtLte", types.Int().Lt(5).Lt(3).Lte(2), `{"type":"integer","maximum":2}`},
		{"IntegerLtLte", types.Int().Lt(5).Lte(3), `{"type":"integer","maximum":3}`},
		{"IntegerGtLt", types.Int().Gt(5).Lt(10), `{"type":"integer","exclusiveMinimum":5,"exclusiveMaximum":10}`},
		{"IntegerGteLte", types.Int().Gte(5).Lte(10), `{"type":"integer","minimum":5,"maximum":10}`},
		{"IntegerPositive", types.Int().Positive(), `{"type":"integer","exclusiveMinimum":0}`},
		{"IntegerNegative", types.Int().Negative(), `{"type":"integer","exclusiveMaximum":0}`},
		{"IntegerNonPositive", types.Int().NonPositive(), `{"type":"integer","maximum":0}`},
		{"IntegerNonNegative", types.Int().NonNegative(), `{"type":"integer","minimum":0}`},

		// MultipleOf constraints
		{"IntegerMultipleOf", types.Int().MultipleOf(5), `{"type":"integer","multipleOf":5}`},
		{"FloatMultipleOf", types.Float().MultipleOf(2.5), `{"type":"number","multipleOf":2.5}`},
		{"IntegerStep", types.Int().Step(3), `{"type":"integer","multipleOf":3}`},

		// Safe integer constraints
		{"IntegerSafe", types.Int().Safe(), `{"type":"integer","minimum":-9007199254740991,"maximum":9007199254740991}`},

		// Specific integer types with their ranges
		{"Int8Constraints", types.Int8().Min(10).Max(100), `{"type":"integer","minimum":10,"maximum":100}`},
		{"Int16Constraints", types.Int16().Min(1000).Max(30000), `{"type":"integer","minimum":1000,"maximum":30000}`},
		{"Int32Constraints", types.Int32().Min(100000).Max(2000000), `{"type":"integer","minimum":100000,"maximum":2000000}`},
		{"Int64Constraints", types.Int64().Min(1000000).Max(9000000000000000), `{"type":"integer","minimum":1000000,"maximum":9000000000000000}`},

		// Unsigned integer types
		{"UintConstraints", types.Uint().Min(10).Max(1000), `{"type":"integer","minimum":10,"maximum":1000}`},
		{"Uint8Constraints", types.Uint8().Min(50).Max(200), `{"type":"integer","minimum":50,"maximum":200}`},
		{"Uint16Constraints", types.Uint16().Min(1000).Max(60000), `{"type":"integer","minimum":1000,"maximum":60000}`},
		{"Uint32Constraints", types.Uint32().Min(100000).Max(4000000000), `{"type":"integer","minimum":100000,"maximum":4000000000}`},
		{"Uint64Constraints", types.Uint64().Min(1000000).Max(9223372036854775807), `{"type":"integer","minimum":1000000,"maximum":9.223372036854776e+18}`},

		// Float types with constraints
		{"Float32Constraints", types.Float32().Min(-1000.5).Max(1000.5), `{"type":"number","minimum":-1000.5,"maximum":1000.5}`},
		{"Float64Constraints", types.Float64().Min(-999999.999).Max(999999.999), `{"type":"number","minimum":-999999.999,"maximum":999999.999}`},

		// Complex constraint combinations
		{"ComplexIntegerConstraints", types.Int().Min(1).Max(100).MultipleOf(5).Positive(), `{"type":"integer","minimum":1,"maximum":100,"multipleOf":5}`},
		{"ComplexFloatConstraints", types.Float().Min(0.1).Max(99.9).NonNegative(), `{"type":"number","minimum":0.1,"maximum":99.9}`},

		// Edge cases with zero
		{"ZeroMinimum", types.Float().Min(0), `{"type":"number","minimum":0}`},
		{"ZeroMaximum", types.Float().Max(0), `{"type":"number","maximum":0}`},
		{"ZeroExclusiveMinimum", types.Float().Gt(0), `{"type":"number","exclusiveMinimum":0}`},
		{"ZeroExclusiveMaximum", types.Float().Lt(0), `{"type":"number","exclusiveMaximum":0}`},

		// Constraint precedence tests (mimicking TypeScript behavior)
		{"GtOverridesGt", types.Float().Gt(5).Gt(10), `{"type":"number","exclusiveMinimum":10}`},
		{"LtOverridesLt", types.Float().Lt(10).Lt(5), `{"type":"number","exclusiveMaximum":5}`},
		{"GteOverridesGt", types.Float().Gt(5).Gte(10), `{"type":"number","minimum":10}`},
		{"LteOverridesLt", types.Float().Lt(10).Lte(5), `{"type":"number","maximum":5}`},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			js, err := ToJSONSchema(c.schema)
			assert.NoError(t, err)
			jsonSchemaBytes, err := json.Marshal(js)
			assert.NoError(t, err)
			assertJSONEquals(t, c.expected, string(jsonSchemaBytes))
		})
	}
}

// =============================================================================
// Slices
// =============================================================================

func TestToJSONSchema_Slices(t *testing.T) {
	t.Run("Simple Array", func(t *testing.T) {
		schema := types.Slice[string](types.String())
		expected := `{"type":"array","items":{"type":"string"}}`
		js, err := ToJSONSchema(schema)
		assert.NoError(t, err)
		jsonSchemaBytes, err := json.Marshal(js)
		assert.NoError(t, err)
		assertJSONEquals(t, expected, string(jsonSchemaBytes))
	})

	t.Run("Array of Numbers", func(t *testing.T) {
		schema := types.Slice[int](types.Int())
		expected := `{"type":"array","items":{"type":"integer"}}`
		js, err := ToJSONSchema(schema)
		assert.NoError(t, err)
		jsonSchemaBytes, err := json.Marshal(js)
		assert.NoError(t, err)
		assertJSONEquals(t, expected, string(jsonSchemaBytes))
	})
}

// =============================================================================
// Arrays
// =============================================================================

func TestToJSONSchema_Arrays(t *testing.T) {
	t.Run("Tuple with Rest", func(t *testing.T) {
		// Tuple: [string, number] followed by boolean rest
		tupleSchema := types.Array([]any{types.String(), types.Float()}, types.Bool())
		expected := `{"type":"array","prefixItems":[{"type":"string"},{"type":"number"}],"items":{"type":"boolean"}}`
		js, err := ToJSONSchema(tupleSchema)
		assert.NoError(t, err)
		jsonSchemaBytes, err := json.Marshal(js)
		assert.NoError(t, err)
		assertJSONEquals(t, expected, string(jsonSchemaBytes))
	})

	t.Run("Fixed Tuple", func(t *testing.T) {
		// Fixed tuple: [string, number]
		tupleSchema := types.Array([]any{types.String(), types.Float()})
		expected := `{"type":"array","prefixItems":[{"type":"string"},{"type":"number"}],"minItems":2,"maxItems":2}`
		js, err := ToJSONSchema(tupleSchema)
		assert.NoError(t, err)
		jsonSchemaBytes, err := json.Marshal(js)
		assert.NoError(t, err)
		assertJSONEquals(t, expected, string(jsonSchemaBytes))
	})
}

// =============================================================================
// UNIONS
// =============================================================================

func TestToJSONSchema_Unions(t *testing.T) {
	t.Run("String or Number", func(t *testing.T) {
		schema := types.Union([]any{types.String(), types.Float()})
		expected := `{"anyOf":[{"type":"string"},{"type":"number"}]}`
		js, err := ToJSONSchema(schema)
		assert.NoError(t, err)
		jsonSchemaBytes, err := json.Marshal(js)
		assert.NoError(t, err)
		assertJSONEquals(t, expected, string(jsonSchemaBytes))
	})

	t.Run("Multiple Types", func(t *testing.T) {
		schema := types.Union([]any{types.String(), types.Int(), types.Bool()})
		expected := `{"anyOf":[{"type":"string"},{"type":"integer"},{"type":"boolean"}]}`
		js, err := ToJSONSchema(schema)
		assert.NoError(t, err)
		jsonSchemaBytes, err := json.Marshal(js)
		assert.NoError(t, err)
		assertJSONEquals(t, expected, string(jsonSchemaBytes))
	})
}

// =============================================================================
// INTERSECTIONS
// =============================================================================

func TestToJSONSchema_Intersections(t *testing.T) {
	t.Run("Object Intersection", func(t *testing.T) {
		schema := types.Intersection(
			types.Object(core.ObjectSchema{"name": types.String()}),
			types.Object(core.ObjectSchema{"age": types.Float()}),
		)
		expected := `{"allOf":[{"type":"object","properties":{"name":{"type":"string"}},"required":["name"],"additionalProperties":false},{"type":"object","properties":{"age":{"type":"number"}},"required":["age"],"additionalProperties":false}]}`
		js, err := ToJSONSchema(schema)
		assert.NoError(t, err)
		jsonSchemaBytes, err := json.Marshal(js)
		assert.NoError(t, err)
		assertJSONEquals(t, expected, string(jsonSchemaBytes))
	})
}

// =============================================================================
// RECORDS
// =============================================================================

func TestToJSONSchema_Records(t *testing.T) {
	t.Run("String to Boolean Record", func(t *testing.T) {
		schema := types.RecordTyped[map[string]bool, map[string]bool](types.String(), types.Bool())
		expected := `{"type":"object","propertyNames":{"type":"string"},"additionalProperties":{"type":"boolean"}}`
		jsonSchema, err := ToJSONSchema(schema)
		assert.NoError(t, err)
		jsonSchemaBytes, err := json.Marshal(jsonSchema)
		assert.NoError(t, err)
		assertJSONEquals(t, expected, string(jsonSchemaBytes))
	})

	t.Run("String to Number Record", func(t *testing.T) {
		schema := types.RecordTyped[map[string]float64, map[string]float64](types.String(), types.Float())
		expected := `{"type":"object","propertyNames":{"type":"string"},"additionalProperties":{"type":"number"}}`
		jsonSchema, err := ToJSONSchema(schema)
		assert.NoError(t, err)
		jsonSchemaBytes, err := json.Marshal(jsonSchema)
		assert.NoError(t, err)
		assertJSONEquals(t, expected, string(jsonSchemaBytes))
	})

	t.Run("LooseRecord with regex key emits patternProperties", func(t *testing.T) {
		// Zod v4 (e01cd02b): loose records with regex key patterns should emit
		// patternProperties instead of propertyNames for more semantic JSON Schema.
		schema := types.LooseRecord(types.String().Regex(regexp.MustCompile("^[a-z]+$")), types.Int())
		expected := `{"type":"object","patternProperties":{"^[a-z]+$":{"type":"integer"}}}`
		jsonSchema, err := ToJSONSchema(schema)
		require.NoError(t, err)
		jsonSchemaBytes, err := json.Marshal(jsonSchema)
		require.NoError(t, err)
		assertJSONEquals(t, expected, string(jsonSchemaBytes))
	})

	t.Run("Non-loose Record with regex key uses propertyNames", func(t *testing.T) {
		// Non-loose records should still use propertyNames even with regex key patterns.
		schema := types.Record(types.String().Regex(regexp.MustCompile("^[a-z]+$")), types.Int())
		expected := `{"type":"object","propertyNames":{"type":"string","pattern":"^[a-z]+$"},"additionalProperties":{"type":"integer"}}`
		jsonSchema, err := ToJSONSchema(schema)
		require.NoError(t, err)
		jsonSchemaBytes, err := json.Marshal(jsonSchema)
		require.NoError(t, err)
		assertJSONEquals(t, expected, string(jsonSchemaBytes))
	})
}

// =============================================================================
// ENUMS
// =============================================================================

func TestToJSONSchema_Enums(t *testing.T) {
	t.Run("String Enum", func(t *testing.T) {
		schema := types.Enum("a", "b", "c")
		expected := `{"type":"string","enum":["a","b","c"]}`
		jsonSchema, err := ToJSONSchema(schema)
		assert.NoError(t, err)
		jsonSchemaBytes, err := json.Marshal(jsonSchema)
		assert.NoError(t, err)
		assertJSONEquals(t, expected, string(jsonSchemaBytes))
	})

	t.Run("Number Enum", func(t *testing.T) {
		schema := types.Enum(1, 2, 3)
		expected := `{"type":"number","enum":[1,2,3]}`
		jsonSchema, err := ToJSONSchema(schema)
		assert.NoError(t, err)
		jsonSchemaBytes, err := json.Marshal(jsonSchema)
		assert.NoError(t, err)
		assertJSONEquals(t, expected, string(jsonSchemaBytes))
	})
}

// =============================================================================
// LITERALS
// =============================================================================

func TestToJSONSchema_Literals(t *testing.T) {
	testCases := []struct {
		name     string
		schema   core.ZodSchema
		expected string
	}{
		{
			name:     "String Literal",
			schema:   types.Literal("hello"),
			expected: `{"type":"string","const":"hello"}`,
		},
		{
			name:     "Number Literal",
			schema:   types.Literal(7),
			expected: `{"type":"number","const":7}`,
		},
		{
			name:     "Boolean Literal",
			schema:   types.Literal(true),
			expected: `{"type":"boolean","const":true}`,
		},
		{
			name:     "False Literal",
			schema:   types.Literal(false),
			expected: `{"type":"boolean","const":false}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			js, err := ToJSONSchema(tc.schema)
			assert.NoError(t, err)
			jsonSchemaBytes, err := json.Marshal(js)
			assert.NoError(t, err)
			assertJSONEquals(t, tc.expected, string(jsonSchemaBytes))
		})
	}
}

// =============================================================================
// OBJECTS
// =============================================================================

func TestToJSONSchema_Objects(t *testing.T) {
	t.Run("Simple Object", func(t *testing.T) {
		schema := types.Object(core.ObjectSchema{
			"name": types.String(),
			"age":  types.Float(),
		})
		expected := `{
			"type": "object",
			"properties": {
				"name": {"type": "string"},
				"age": {"type": "number"}
			},
			"required": ["name", "age"],
			"additionalProperties": false
		}`
		js, err := ToJSONSchema(schema)
		assert.NoError(t, err)
		jsonSchemaBytes, err := json.Marshal(js)
		assert.NoError(t, err)
		assertJSONEquals(t, expected, string(jsonSchemaBytes))
	})

	t.Run("Object with Optional Fields", func(t *testing.T) {
		schema := types.Object(core.ObjectSchema{
			"required":    types.String(),
			"optional":    types.String().Optional(),
			"nonoptional": types.String().Optional().NonOptional(),
		})
		expected := `{
			"type": "object",
			"properties": {
				"required": {"type": "string"},
				"optional": {"type": "string"},
				"nonoptional": {"type": "string"}
			},
			"required": ["required", "nonoptional"],
			"additionalProperties": false
		}`
		js, err := ToJSONSchema(schema)
		assert.NoError(t, err)
		jsonSchemaBytes, err := json.Marshal(js)
		assert.NoError(t, err)
		assertJSONEquals(t, expected, string(jsonSchemaBytes))
	})

	t.Run("Nested Objects", func(t *testing.T) {
		schema := types.Object(core.ObjectSchema{
			"user": types.Object(core.ObjectSchema{
				"name": types.String(),
			}),
		})
		expected := `{
			"type": "object",
			"properties": {
				"user": {
					"type": "object",
					"properties": {
						"name": {"type": "string"}
					},
					"required": ["name"],
					"additionalProperties": false
				}
			},
			"required": ["user"],
			"additionalProperties": false
		}`
		js, err := ToJSONSchema(schema)
		assert.NoError(t, err)
		jsonSchemaBytes, err := json.Marshal(js)
		assert.NoError(t, err)
		assertJSONEquals(t, expected, string(jsonSchemaBytes))
	})

	t.Run("Object with Catchall", func(t *testing.T) {
		schema := types.Object(core.ObjectSchema{
			"name": types.String(),
		}).WithCatchall(types.String())
		expected := `{
			"type": "object",
			"properties": {
				"name": {"type": "string"}
			},
			"required": ["name"],
			"additionalProperties": {
				"type": "string"
			}
		}`
		js, err := ToJSONSchema(schema)
		assert.NoError(t, err)
		jsonSchemaBytes, err := json.Marshal(js)
		assert.NoError(t, err)
		assertJSONEquals(t, expected, string(jsonSchemaBytes))
	})

	t.Run("Strict Object", func(t *testing.T) {
		schema := types.StrictObject(core.ObjectSchema{
			"name": types.String(),
			"age":  types.Float(),
		})
		expected := `{
			"type": "object",
			"properties": {
				"name": {"type": "string"},
				"age": {"type": "number"}
			},
			"required": ["name", "age"],
			"additionalProperties": false
		}`
		js, err := ToJSONSchema(schema)
		assert.NoError(t, err)
		jsonSchemaBytes, err := json.Marshal(js)
		assert.NoError(t, err)
		assertJSONEquals(t, expected, string(jsonSchemaBytes))
	})

	t.Run("Loose Object", func(t *testing.T) {
		schema := types.LooseObject(core.ObjectSchema{
			"name": types.String(),
		})
		expected := `{
			"type": "object",
			"properties": {
				"name": {"type": "string"}
			},
			"required": ["name"],
			"additionalProperties": true
		}`
		js, err := ToJSONSchema(schema)
		assert.NoError(t, err)
		jsonSchemaBytes, err := json.Marshal(js)
		assert.NoError(t, err)
		assertJSONEquals(t, expected, string(jsonSchemaBytes))
	})

	t.Run("Object with Mixed Field Types", func(t *testing.T) {
		schema := types.Object(core.ObjectSchema{
			"id":     types.Int(),
			"name":   types.String(),
			"email":  types.String().Email(),
			"age":    types.Float().Optional(),
			"active": types.Bool(),
			"tags":   types.Slice[string](types.String()),
			"metadata": types.Object(core.ObjectSchema{
				"created": types.String(),
				"updated": types.String().Optional(),
			}),
		})
		expected := `{
			"type": "object",
			"properties": {
				"id": {"type": "integer"},
				"name": {"type": "string"},
				"email": {
					"type": "string",
					"format": "email",
					"pattern": "^[A-Za-z0-9_'+\\-]+([A-Za-z0-9_'+\\-]*\\.[A-Za-z0-9_'+\\-]+)*@[A-Za-z0-9]([A-Za-z0-9\\-]*[A-Za-z0-9])?(\\.[A-Za-z0-9]([A-Za-z0-9\\-]*[A-Za-z0-9])?)*\\.[A-Za-z]{2,}$"
				},
				"age": {"type": "number"},
				"active": {"type": "boolean"},
				"tags": {
					"type": "array",
					"items": {"type": "string"}
				},
				"metadata": {
					"type": "object",
					"properties": {
						"created": {"type": "string"},
						"updated": {"type": "string"}
					},
					"required": ["created"],
					"additionalProperties": false
				}
			},
			"required": ["tags", "name", "metadata", "id", "email", "active"],
			"additionalProperties": false
		}`
		js, err := ToJSONSchema(schema)
		assert.NoError(t, err)
		jsonSchemaBytes, err := json.Marshal(js)
		assert.NoError(t, err)
		assertJSONEquals(t, expected, string(jsonSchemaBytes))
	})
}

// =============================================================================
// ARRAY OF OBJECTS
// =============================================================================

func TestToJSONSchema_ArrayOfObjects(t *testing.T) {
	t.Run("Array of Objects", func(t *testing.T) {
		schema := types.Slice[map[string]any](types.Object(core.ObjectSchema{
			"id": types.Int(),
		}))
		expected := `{
			"type": "array",
			"items": {
				"type": "object",
				"properties": {
					"id": {"type": "integer"}
				},
				"required": ["id"],
				"additionalProperties": false
			}
		}`
		js, err := ToJSONSchema(schema)
		assert.NoError(t, err)
		jsonSchemaBytes, err := json.Marshal(js)
		assert.NoError(t, err)
		assertJSONEquals(t, expected, string(jsonSchemaBytes))
	})
}

// =============================================================================
// LAZY SCHEMAS
// =============================================================================

func TestToJSONSchema_LazySchemas(t *testing.T) {
	t.Run("Lazy String", func(t *testing.T) {
		lazySchema := types.LazyAny(func() any { return types.String() })
		expected := `{"type":"string"}`
		js, err := ToJSONSchema(lazySchema)
		assert.NoError(t, err)
		jsonSchemaBytes, err := json.Marshal(js)
		assert.NoError(t, err)
		assertJSONEquals(t, expected, string(jsonSchemaBytes))
	})

	t.Run("Lazy Object", func(t *testing.T) {
		lazySchema := types.LazyAny(func() any {
			return types.Object(core.ObjectSchema{
				"name": types.String(),
			})
		})
		expected := `{
			"type": "object",
			"properties": {
				"name": {"type": "string"}
			},
			"required": ["name"],
			"additionalProperties": false
		}`
		js, err := ToJSONSchema(lazySchema)
		assert.NoError(t, err)
		jsonSchemaBytes, err := json.Marshal(js)
		assert.NoError(t, err)
		assertJSONEquals(t, expected, string(jsonSchemaBytes))
	})
}

// =============================================================================
// OPTIONAL AND NILABLE
// =============================================================================

func TestToJSONSchema_OptionalAndNilable(t *testing.T) {
	testCases := []struct {
		name     string
		schema   core.ZodSchema
		expected string
	}{
		{
			name:   "Optional String",
			schema: types.String().Optional(),
			expected: `{
				"type": "string"
			}`,
		},
		{
			name:   "Nilable Integer",
			schema: types.Int().Nilable(),
			expected: `{
				"anyOf": [
					{"type": "integer"},
					{"type": "null"}
				]
			}`,
		},
		{
			name:   "Optional and Nilable String",
			schema: types.String().Optional().Nilable(),
			expected: `{
				"anyOf": [
					{"type": "string"},
					{"type": "null"}
				]
			}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonSchema, err := ToJSONSchema(tc.schema)
			assert.NoError(t, err)
			jsonSchemaBytes, err := json.Marshal(jsonSchema)
			assert.NoError(t, err)
			assertJSONEquals(t, tc.expected, string(jsonSchemaBytes))
		})
	}
}

// =============================================================================
// ADVANCED Slices
// =============================================================================

func TestToJSONSchema_AdvancedSlices(t *testing.T) {
	testCases := []struct {
		name     string
		schema   core.ZodSchema
		expected string
	}{
		{
			name:     "Array with Min Items",
			schema:   types.Slice[string](types.String()).Min(2),
			expected: `{"type":"array","items":{"type":"string"},"minItems":2}`,
		},
		{
			name:     "Array with Max Items",
			schema:   types.Slice[string](types.String()).Max(5),
			expected: `{"type":"array","items":{"type":"string"},"maxItems":5}`,
		},
		{
			name:     "Array with Min and Max Items",
			schema:   types.Slice[string](types.String()).Min(2).Max(5),
			expected: `{"type":"array","items":{"type":"string"},"minItems":2,"maxItems":5}`,
		},
		{
			name:     "Array with Exact Length",
			schema:   types.Slice[string](types.String()).Length(3),
			expected: `{"type":"array","items":{"type":"string"},"minItems":3,"maxItems":3}`,
		},
		{
			name:     "Non-empty Array",
			schema:   types.Slice[string](types.String()).NonEmpty(),
			expected: `{"type":"array","items":{"type":"string"},"minItems":1}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonSchema, err := ToJSONSchema(tc.schema)
			assert.NoError(t, err)
			jsonSchemaBytes, err := json.Marshal(jsonSchema)
			assert.NoError(t, err)
			assertJSONEquals(t, tc.expected, string(jsonSchemaBytes))
		})
	}
}

// =============================================================================
// STRING CONSTRAINTS
// =============================================================================

func TestToJSONSchema_StringConstraints(t *testing.T) {
	testCases := []struct {
		name     string
		schema   core.ZodSchema
		expected string
	}{
		{
			name:     "String with StartsWith",
			schema:   types.String().StartsWith("hello"),
			expected: `{"type":"string","pattern":"^hello.*"}`,
		},
		{
			name:     "String with EndsWith",
			schema:   types.String().EndsWith("world"),
			expected: `{"type":"string","pattern":".*world$"}`,
		},
		{
			name:     "String with Includes",
			schema:   types.String().Includes("foo"),
			expected: `{"type":"string","pattern":"foo"}`,
		},
		{
			name:     "String with Includes - Special Chars",
			schema:   types.String().Includes("foo.bar?"),
			expected: `{"type":"string","pattern":"foo\\.bar\\?"}`,
		},
		{
			name:     "String with Regex",
			schema:   types.String().RegexString("^[a-z]+$"),
			expected: `{"type":"string","pattern":"^[a-z]+$"}`,
		},
		{
			name: "Combined String Constraints",
			schema: types.String().
				StartsWith("h").
				EndsWith("d").
				Includes("ell"),
			expected: `{
				"type": "string",
				"allOf": [
					{"pattern": "^h.*"},
					{"pattern": ".*d$"},
					{"pattern": "ell"}
				]
			}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonSchema, err := ToJSONSchema(tc.schema)
			assert.NoError(t, err)
			jsonSchemaBytes, err := json.Marshal(jsonSchema)
			assert.NoError(t, err)
			assertJSONEquals(t, tc.expected, string(jsonSchemaBytes))
		})
	}
}

// =============================================================================
// STRING FORMAT CHAINING TESTS
// =============================================================================

func TestToJSONSchema_StringFormatsChaining(t *testing.T) {
	testCases := []struct {
		name     string
		schema   core.ZodSchema
		expected string
	}{
		{
			name:     "String Email",
			schema:   types.String().Email(),
			expected: `{"type":"string","format":"email","pattern":"^[A-Za-z0-9_'+\\-]+([A-Za-z0-9_'+\\-]*\\.[A-Za-z0-9_'+\\-]+)*@[A-Za-z0-9]([A-Za-z0-9\\-]*[A-Za-z0-9])?(\\.[A-Za-z0-9]([A-Za-z0-9\\-]*[A-Za-z0-9])?)*\\.[A-Za-z]{2,}$"}`,
		},
		{
			name:     "String with Length and Email",
			schema:   types.String().Email().Min(10).Max(50),
			expected: `{"type":"string","format":"email","pattern":"^[A-Za-z0-9_'+\\-]+([A-Za-z0-9_'+\\-]*\\.[A-Za-z0-9_'+\\-]+)*@[A-Za-z0-9]([A-Za-z0-9\\-]*[A-Za-z0-9])?(\\.[A-Za-z0-9]([A-Za-z0-9\\-]*[A-Za-z0-9])?)*\\.[A-Za-z]{2,}$","minLength":10,"maxLength":50}`,
		},
		{
			name:     "String with JSON validation",
			schema:   types.String().JSON(),
			expected: `{"type":"string","contentMediaType":"application/json","pattern":"^[\\s\\S]*$"}`,
		},
		{
			name:   "String with Multiple Pattern Constraints",
			schema: types.String().StartsWith("test").EndsWith(".com").Includes("@"),
			expected: `{
				"type": "string",
				"allOf": [
					{"pattern": "^test.*"},
					{"pattern": ".*\\.com$"},
					{"pattern": "@"}
				]
			}`,
		},
		{
			name:   "String Email with Pattern Constraints",
			schema: types.String().Email().StartsWith("test"),
			expected: `{
				"type": "string",
				"format": "email",
				"allOf": [
					{"pattern": "^[A-Za-z0-9_'+\\-]+([A-Za-z0-9_'+\\-]*\\.[A-Za-z0-9_'+\\-]+)*@[A-Za-z0-9]([A-Za-z0-9\\-]*[A-Za-z0-9])?(\\.[A-Za-z0-9]([A-Za-z0-9\\-]*[A-Za-z0-9])?)*\\.[A-Za-z]{2,}$"},
					{"pattern": "^test.*"}
				]
			}`,
		},
		{
			name:     "String with Min/Max Length",
			schema:   types.String().Min(5).Max(20),
			expected: `{"type":"string","minLength":5,"maxLength":20}`,
		},
		{
			name:     "String with Exact Length",
			schema:   types.String().Length(10),
			expected: `{"type":"string","minLength":10,"maxLength":10}`,
		},
		{
			name:     "String with Custom Regex",
			schema:   types.String().RegexString("^[a-zA-Z0-9]+$"),
			expected: `{"type":"string","pattern":"^[a-zA-Z0-9]+$"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonSchema, err := ToJSONSchema(tc.schema)
			assert.NoError(t, err)
			jsonSchemaBytes, err := json.Marshal(jsonSchema)
			assert.NoError(t, err)
			assertJSONEquals(t, tc.expected, string(jsonSchemaBytes))
		})
	}
}

func TestToJSONSchema_DiscriminatedUnionsAdvanced(t *testing.T) {
	t.Run("Discriminated Union", func(t *testing.T) {
		schema := types.DiscriminatedUnion("type", []any{
			types.Object(core.ObjectSchema{
				"type": types.Literal("a"),
				"a":    types.String(),
			}),
			types.Object(core.ObjectSchema{
				"type": types.Literal("b"),
				"b":    types.Int(),
			}),
		})

		expected := `{
			"oneOf": [
				{
					"type": "object",
					"properties": {
						"type": {"type": "string", "const": "a"},
						"a": {"type": "string"}
					},
					"required": ["type", "a"],
					"additionalProperties": false
				},
				{
					"type": "object",
					"properties": {
						"type": {"type": "string", "const": "b"},
						"b": {"type": "integer"}
					},
					"required": ["type", "b"],
					"additionalProperties": false
				}
			]
		}`

		jsonSchema, err := ToJSONSchema(schema)
		assert.NoError(t, err)
		jsonSchemaBytes, err := json.Marshal(jsonSchema)
		assert.NoError(t, err)
		assertJSONEquals(t, expected, string(jsonSchemaBytes))
	})
}

// =============================================================================
// STRUCTS
// =============================================================================

func TestToJSONSchema_Structs(t *testing.T) {
	// Define test structs
	type User struct {
		Name  string `json:"name"`
		Age   int    `json:"age"`
		Email string `json:"email"`
	}

	type Profile struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
		Bio      string `json:"bio,omitempty"`
		Active   bool   `json:"active"`
	}

	type Company struct {
		Name      string   `json:"name"`
		Employees []User   `json:"employees"`
		Founded   int      `json:"founded"`
		Public    bool     `json:"public"`
		Tags      []string `json:"tags"`
	}

	t.Run("Simple Struct", func(t *testing.T) {
		schema := types.Struct[User]()
		expected := `{
			"type": "object",
			"additionalProperties": false
		}`
		js, err := ToJSONSchema(schema)
		assert.NoError(t, err)
		jsonSchemaBytes, err := json.Marshal(js)
		assert.NoError(t, err)
		assertJSONEquals(t, expected, string(jsonSchemaBytes))
	})

	t.Run("Struct with Field Schema", func(t *testing.T) {
		schema := types.Struct[User](core.StructSchema{
			"name":  types.String().Min(2),
			"age":   types.Int().Min(0).Max(150),
			"email": types.String().Email(),
		})
		expected := `{
			"type": "object",
			"properties": {
				"name": {
					"type": "string",
					"minLength": 2
				},
				"age": {
					"type": "integer",
					"minimum": 0,
					"maximum": 150
				},
				"email": {
					"type": "string",
					"format": "email",
					"pattern": "^[A-Za-z0-9_'+\\-]+([A-Za-z0-9_'+\\-]*\\.[A-Za-z0-9_'+\\-]+)*@[A-Za-z0-9]([A-Za-z0-9\\-]*[A-Za-z0-9])?(\\.[A-Za-z0-9]([A-Za-z0-9\\-]*[A-Za-z0-9])?)*\\.[A-Za-z]{2,}$"
				}
			},
			"required": ["name", "email", "age"],
			"additionalProperties": false
		}`
		js, err := ToJSONSchema(schema)
		assert.NoError(t, err)
		jsonSchemaBytes, err := json.Marshal(js)
		assert.NoError(t, err)
		assertJSONEquals(t, expected, string(jsonSchemaBytes))
	})

	t.Run("Struct with Optional Fields", func(t *testing.T) {
		schema := types.Struct[Profile](core.StructSchema{
			"id":       types.Int().Min(1),
			"username": types.String().Min(3),
			"bio":      types.String().Optional(),
			"active":   types.Bool(),
		})
		expected := `{
			"type": "object",
			"properties": {
				"id": {
					"type": "integer",
					"minimum": 1
				},
				"username": {
					"type": "string",
					"minLength": 3
				},
				"bio": {
					"type": "string"
				},
				"active": {
					"type": "boolean"
				}
			},
			"required": ["username", "id", "active"],
			"additionalProperties": false
		}`
		js, err := ToJSONSchema(schema)
		assert.NoError(t, err)
		jsonSchemaBytes, err := json.Marshal(js)
		assert.NoError(t, err)
		assertJSONEquals(t, expected, string(jsonSchemaBytes))
	})

	t.Run("StructPtr", func(t *testing.T) {
		schema := types.StructPtr[User](core.StructSchema{
			"name":  types.String().Min(1),
			"age":   types.Int().Min(0),
			"email": types.String().Email(),
		})
		expected := `{
			"type": "object",
			"properties": {
				"name": {
					"type": "string",
					"minLength": 1
				},
				"age": {
					"type": "integer",
					"minimum": 0
				},
				"email": {
					"type": "string",
					"format": "email",
					"pattern": "^[A-Za-z0-9_'+\\-]+([A-Za-z0-9_'+\\-]*\\.[A-Za-z0-9_'+\\-]+)*@[A-Za-z0-9]([A-Za-z0-9\\-]*[A-Za-z0-9])?(\\.[A-Za-z0-9]([A-Za-z0-9\\-]*[A-Za-z0-9])?)*\\.[A-Za-z]{2,}$"
				}
			},
			"required": ["name", "email", "age"],
			"additionalProperties": false
		}`
		js, err := ToJSONSchema(schema)
		assert.NoError(t, err)
		jsonSchemaBytes, err := json.Marshal(js)
		assert.NoError(t, err)
		assertJSONEquals(t, expected, string(jsonSchemaBytes))
	})

	t.Run("Nested Struct", func(t *testing.T) {
		schema := types.Struct[Company](core.StructSchema{
			"name": types.String().Min(1),
			"employees": types.Slice[User](types.Struct[User](core.StructSchema{
				"name":  types.String(),
				"age":   types.Int(),
				"email": types.String().Email(),
			})),
			"founded": types.Int().Min(1800).Max(2100),
			"public":  types.Bool(),
			"tags":    types.Slice[string](types.String()),
		})
		expected := `{
			"type": "object",
			"properties": {
				"name": {
					"type": "string",
					"minLength": 1
				},
				"employees": {
					"type": "array",
					"items": {
						"type": "object",
						"properties": {
							"name": {"type": "string"},
							"age": {"type": "integer"},
							"email": {
								"type": "string",
								"format": "email",
								"pattern": "^[A-Za-z0-9_'+\\-]+([A-Za-z0-9_'+\\-]*\\.[A-Za-z0-9_'+\\-]+)*@[A-Za-z0-9]([A-Za-z0-9\\-]*[A-Za-z0-9])?(\\.[A-Za-z0-9]([A-Za-z0-9\\-]*[A-Za-z0-9])?)*\\.[A-Za-z]{2,}$"
							}
						},
						"required": ["name", "email", "age"],
						"additionalProperties": false
					}
				},
				"founded": {
					"type": "integer",
					"minimum": 1800,
					"maximum": 2100
				},
				"public": {
					"type": "boolean"
				},
				"tags": {
					"type": "array",
					"items": {"type": "string"}
				}
			},
			"required": ["tags", "public", "name", "founded", "employees"],
			"additionalProperties": false
		}`
		js, err := ToJSONSchema(schema)
		assert.NoError(t, err)
		jsonSchemaBytes, err := json.Marshal(js)
		assert.NoError(t, err)
		assertJSONEquals(t, expected, string(jsonSchemaBytes))
	})
}

// =============================================================================
// RECURSIVE SCHEMAS
// =============================================================================

func TestToJSONSchema_RecursiveSchemas(t *testing.T) {
	t.Run("Recursive Object with Lazy", func(t *testing.T) {
		type Category struct {
			Name          string     `json:"name"`
			Subcategories []Category `json:"subcategories"`
		}

		var categorySchema core.ZodSchema
		categorySchema = types.Struct[Category](core.StructSchema{
			"name": types.String(),
			"subcategories": types.Slice[Category](types.LazyAny(func() any {
				return categorySchema
			})),
		})

		expected := `{
			"type": "object",
			"properties": {
				"name": {"type": "string"},
				"subcategories": {
					"type": "array",
					"items": {
						"$ref": "#"
					}
				}
			},
			"required": ["subcategories", "name"],
			"additionalProperties": false
		}`
		js, err := ToJSONSchema(categorySchema)
		assert.NoError(t, err)
		jsonSchemaBytes, err := json.Marshal(js)
		assert.NoError(t, err)
		assertJSONEquals(t, expected, string(jsonSchemaBytes))
	})

	t.Run("Simple Recursive Object", func(t *testing.T) {
		type TreeNode struct {
			ID       string    `json:"id"`
			Children *TreeNode `json:"children"`
		}

		var treeSchema core.ZodSchema
		treeSchema = types.Struct[TreeNode](core.StructSchema{
			"id": types.String(),
			"children": types.LazyAny(func() any {
				return treeSchema
			}),
		})

		expected := `{
			"type": "object",
			"properties": {
				"id": {"type": "string"},
				"children": {
					"$ref": "#"
				}
			},
			"required": ["id", "children"],
			"additionalProperties": false
		}`
		js, err := ToJSONSchema(treeSchema)
		assert.NoError(t, err)
		jsonSchemaBytes, err := json.Marshal(js)
		assert.NoError(t, err)
		assertJSONEquals(t, expected, string(jsonSchemaBytes))
	})
}

// =============================================================================
// ADVANCED OBJECT PATTERNS
// =============================================================================

func TestToJSONSchema_AdvancedObjectPatterns(t *testing.T) {
	t.Run("Object with Union Fields", func(t *testing.T) {
		schema := types.Object(core.ObjectSchema{
			"value": types.Union([]any{types.String(), types.Int(), types.Bool()}),
			"type":  types.Enum("string", "number", "boolean"),
		})
		expected := `{
			"type": "object",
			"properties": {
				"value": {
					"anyOf": [
						{"type": "string"},
						{"type": "integer"},
						{"type": "boolean"}
					]
				},
				"type": {
					"type": "string",
					"enum": ["boolean", "number", "string"]
				}
			},
			"required": ["value", "type"],
			"additionalProperties": false
		}`
		js, err := ToJSONSchema(schema)
		assert.NoError(t, err)
		jsonSchemaBytes, err := json.Marshal(js)
		assert.NoError(t, err)
		assertJSONEquals(t, expected, string(jsonSchemaBytes))
	})

	t.Run("Object with Nilable Fields", func(t *testing.T) {
		schema := types.Object(core.ObjectSchema{
			"name":        types.String(),
			"description": types.String().Nilable(),
			"count":       types.Int().Optional().Nilable(),
		})
		expected := `{
			"type": "object",
			"properties": {
				"name": {"type": "string"},
				"description": {
					"anyOf": [
						{"type": "string"},
						{"type": "null"}
					]
				},
				"count": {
					"anyOf": [
						{"type": "integer"},
						{"type": "null"}
					]
				}
			},
			"required": ["name", "description"],
			"additionalProperties": false
		}`
		js, err := ToJSONSchema(schema)
		assert.NoError(t, err)
		jsonSchemaBytes, err := json.Marshal(js)
		assert.NoError(t, err)
		assertJSONEquals(t, expected, string(jsonSchemaBytes))
	})

	t.Run("Object with Array of Objects", func(t *testing.T) {
		schema := types.Object(core.ObjectSchema{
			"users": types.Slice[map[string]any](types.Object(core.ObjectSchema{
				"id":   types.Int(),
				"name": types.String(),
			})),
			"total": types.Int(),
		})
		expected := `{
			"type": "object",
			"properties": {
				"users": {
					"type": "array",
					"items": {
						"type": "object",
						"properties": {
							"id": {"type": "integer"},
							"name": {"type": "string"}
						},
						"required": ["name", "id"],
						"additionalProperties": false
					}
				},
				"total": {"type": "integer"}
			},
			"required": ["users", "total"],
			"additionalProperties": false
		}`
		js, err := ToJSONSchema(schema)
		assert.NoError(t, err)
		jsonSchemaBytes, err := json.Marshal(js)
		assert.NoError(t, err)
		assertJSONEquals(t, expected, string(jsonSchemaBytes))
	})

	t.Run("Object with Record Fields", func(t *testing.T) {
		schema := types.Object(core.ObjectSchema{
			"metadata": types.RecordTyped[map[string]string, map[string]string](types.String(), types.String()),
			"name":     types.String(),
		})
		expected := `{
			"type": "object",
			"properties": {
				"metadata": {
					"type": "object",
					"propertyNames": {"type": "string"},
					"additionalProperties": {"type": "string"}
				},
				"name": {"type": "string"}
			},
			"required": ["name", "metadata"],
			"additionalProperties": false
		}`
		js, err := ToJSONSchema(schema)
		assert.NoError(t, err)
		jsonSchemaBytes, err := json.Marshal(js)
		assert.NoError(t, err)
		assertJSONEquals(t, expected, string(jsonSchemaBytes))
	})
}

func TestToJSONSchemaOverride(t *testing.T) {
	schema := types.String()
	opts := Options{
		Override: func(ctx OverrideContext) {
			ctx.JSONSchema.Title = new("overridden")
		},
	}
	schemaObj, err := ToJSONSchema(schema, opts)
	assert.NoError(t, err)
	jsonBytes, err := json.Marshal(schemaObj)
	assert.NoError(t, err)
	assert.Contains(t, string(jsonBytes), `"title":"overridden"`)
}

func TestToJSONSchemaOverrideWithRefs(t *testing.T) {
	a := types.String().Optional()
	opts := Options{
		Override: func(ctx OverrideContext) {
			// Optional string returns a *ZodString[*string]
			if _, ok := ctx.ZodSchema.(*types.ZodString[*string]); ok {
				ctx.JSONSchema.Title = new("overridden_string")
			}
		},
	}
	schemaObj, err := ToJSONSchema(a, opts)
	assert.NoError(t, err)
	jsonBytes, err := json.Marshal(schemaObj)
	assert.NoError(t, err)
	assert.Contains(t, string(jsonBytes), `"title":"overridden_string"`)
}

func TestToJSONSchemaTransformIO(t *testing.T) {
	mySchema := types.String().Transform(func(s string, ctx *core.RefinementContext) (any, error) {
		return len(s), nil
	})

	// For IO:"output", transform is unrepresentable. With "any", it should be an empty schema.
	outputSchema, err := ToJSONSchema(mySchema, Options{Unrepresentable: "any", IO: "output"})
	assert.NoError(t, err)
	outputJSON, err := json.Marshal(outputSchema)
	assert.NoError(t, err)
	assert.JSONEq(t, `{}`, string(outputJSON), "output of transform should be an empty schema with unrepresentable:any")

	// For IO:"input", it should represent the input schema (string).
	inputSchema, err := ToJSONSchema(mySchema, Options{IO: "input"})
	assert.NoError(t, err)
	inputJSON, err := json.Marshal(inputSchema)
	assert.NoError(t, err)
	assert.Contains(t, string(inputJSON), `"type":"string"`)
}

func TestToJSONSchemaPassthroughSchemas(t *testing.T) {
	Internal := types.Struct[map[string]any](core.ObjectSchema{
		"num": types.Number(),
		"str": types.String(),
	})

	External := types.Struct[map[string]any](core.ObjectSchema{
		"a": Internal,
		"b": Internal.Optional(),
		"c": types.Lazy(func() core.ZodSchema { return Internal }),
	})

	result, err := ToJSONSchema(External, Options{
		Reused: "ref",
	})
	assert.NoError(t, err)

	resultBytes, err := json.Marshal(result)
	assert.NoError(t, err)
	resultStr := string(resultBytes)

	assert.Contains(t, resultStr, `"$defs":`)
	assert.Contains(t, resultStr, `"$ref":"#/$defs/def1"`)
	assert.Equal(t, 2, strings.Count(resultStr, `"$ref":"#/$defs/def1"`))
}

func TestToJSONSchemaExtractSchemasWithID(t *testing.T) {
	name := types.String().Meta(core.GlobalMeta{ID: "name"})
	age := types.Number().Meta(core.GlobalMeta{ID: "age"})

	schema := types.Struct[map[string]any](core.ObjectSchema{
		"first_name":  name,
		"last_name":   name.Nilable(),
		"middle_name": name.Optional(),
		"age":         age,
	})

	result, err := ToJSONSchema(schema)
	assert.NoError(t, err)
	resultBytes, err := json.Marshal(result)
	assert.NoError(t, err)
	resultStr := string(resultBytes)

	assert.Contains(t, resultStr, `"$defs":{"age":{"type":"number"},"name":{"type":"string"}}`)
	assert.Contains(t, resultStr, `"first_name":{"$ref":"#/$defs/name"}`)
	assert.Contains(t, resultStr, `"middle_name":{"$ref":"#/$defs/name"}`)
	assert.Contains(t, resultStr, `"age":{"$ref":"#/$defs/age"}`)
	assert.Contains(t, resultStr, `"last_name":{"anyOf":[{"$ref":"#/$defs/name"},{"type":"null"}]}`)
}

func TestToJSONSchemaUnrepresentableLiteral(t *testing.T) {
	schema := types.Literal[any]([]any{"hello", "world"})

	result, err := ToJSONSchema(schema, Options{Unrepresentable: "any"})
	assert.NoError(t, err)
	var data map[string]any
	resultBytes, err := json.Marshal(result)
	assert.NoError(t, err)
	err = json.Unmarshal(resultBytes, &data)
	assert.NoError(t, err)

	enum, ok := data["enum"].([]any)
	assert.True(t, ok)
	assert.ElementsMatch(t, []any{"hello", "world"}, enum)
}

func TestToJSONSchemaDescribeWithID(t *testing.T) {
	jobID := types.String().Meta(core.GlobalMeta{ID: "jobId"})

	schema := types.Struct[map[string]any](core.ObjectSchema{
		"current":  jobID.Meta(core.GlobalMeta{Description: "Current job"}),
		"previous": jobID.Meta(core.GlobalMeta{Description: "Previous job"}),
	})

	result, err := ToJSONSchema(schema)
	assert.NoError(t, err)

	resultBytes, err := json.Marshal(result)
	assert.NoError(t, err)
	resultStr := string(resultBytes)

	assert.Contains(t, resultStr, `"$defs":{"jobId":{"type":"string"}}`)
	assert.Contains(t, resultStr, `"description":"Current job"`)
	assert.Contains(t, resultStr, `"$ref":"#/$defs/jobId"`)
	assert.Contains(t, resultStr, `"description":"Previous job"`)
}

func TestToJSONSchemaOverwriteID(t *testing.T) {
	jobID := types.String().Meta(core.GlobalMeta{ID: "aaa"})

	schema := types.Struct[map[string]any](core.ObjectSchema{
		"current":  jobID,
		"previous": jobID.Meta(core.GlobalMeta{ID: "bbb"}),
	})

	result, err := ToJSONSchema(schema)
	assert.NoError(t, err)
	resultBytes, err := json.Marshal(result)
	assert.NoError(t, err)
	resultStr := string(resultBytes)

	assert.Regexp(t, regexp.MustCompile(`"\$defs":{.*"aaa":{"type":"string"}.*}`), resultStr)
	assert.Regexp(t, regexp.MustCompile(`"\$defs":{.*"bbb":{.*}.*}`), resultStr)
	assert.Contains(t, resultStr, `"current":{"$ref":"#/$defs/aaa"}`)
	assert.Contains(t, resultStr, `"previous":{"$ref":"#/$defs/bbb"}`)
}

func TestToJSONSchemaInputOutputType(t *testing.T) {
	schema := types.Struct[map[string]any](core.ObjectSchema{
		"a": types.String(),
		"b": types.String().Optional(),
		"c": types.String().Default("hello"),
		"d": types.String().Nilable(),
	})

	inputResult, err := ToJSONSchema(schema, Options{IO: "input"})
	assert.NoError(t, err)
	var inputData map[string]any
	inputResultBytes, err := json.Marshal(inputResult)
	assert.NoError(t, err)
	err = json.Unmarshal(inputResultBytes, &inputData)
	assert.NoError(t, err)
	inputRequired := inputData["required"].([]any)
	assert.ElementsMatch(t, []string{"a", "d"}, inputRequired)

	outputResult, err := ToJSONSchema(schema, Options{IO: "output"})
	assert.NoError(t, err)
	var outputData map[string]any
	outputResultBytes, err := json.Marshal(outputResult)
	assert.NoError(t, err)
	err = json.Unmarshal(outputResultBytes, &outputData)
	assert.NoError(t, err)
	outputRequired := outputData["required"].([]any)
	assert.ElementsMatch(t, []string{"a", "c", "d"}, outputRequired)
}

func TestToJSONSchemaBasicRegistry(t *testing.T) {
	myRegistry := core.NewRegistry[core.GlobalMeta]()

	var User, Post core.ZodSchema

	User = types.Struct[map[string]any](core.ObjectSchema{
		"name": types.String(),
		"posts": types.Lazy(func() core.ZodSchema {
			return types.Array(Post)
		}),
	})

	Post = types.Struct[map[string]any](core.ObjectSchema{
		"title":   types.String(),
		"content": types.String(),
		"author": types.Lazy(func() core.ZodSchema {
			return User
		}),
	})

	myRegistry.Add(User, core.GlobalMeta{ID: "User"})
	myRegistry.Add(Post, core.GlobalMeta{ID: "Post"})

	result, err := ToJSONSchema(myRegistry)
	assert.NoError(t, err)
	resultBytes, err := json.Marshal(result)
	assert.NoError(t, err)
	resultStr := string(resultBytes)

	assert.Contains(t, resultStr, `"$defs":{`)
	assert.Contains(t, resultStr, `"Post":{`)
	assert.Contains(t, resultStr, `"User":{`)

	assert.Contains(t, resultStr, `"author":{"$ref":"#/$defs/User"`)
	assert.Contains(t, resultStr, `"posts":{"items":{"$ref":"#/$defs/Post"},"type":"array"}`)
}

// =============================================================================
// DESCRIPTION OVERRIDE TEST
// =============================================================================

func TestToJSONSchema_OverwriteDescriptions(t *testing.T) {
	field := types.String().Meta(core.GlobalMeta{Description: "a"}).
		Meta(core.GlobalMeta{Description: "b"}).
		Meta(core.GlobalMeta{Description: "c"})

	schema := types.Object(core.ObjectSchema{
		"d": field.Meta(core.GlobalMeta{Description: "d"}),
		"e": field.Meta(core.GlobalMeta{Description: "e"}),
	})

	js, err := ToJSONSchema(schema)
	assert.NoError(t, err)
	jsonBytes, err := json.Marshal(js)
	assert.NoError(t, err)
	resultStr := string(jsonBytes)
	assert.Contains(t, resultStr, "\"description\":\"d\"")
	assert.Contains(t, resultStr, "\"description\":\"e\"")
}
