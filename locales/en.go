package locales

import (
	"fmt"

	"github.com/kaptinlin/gozod"
)

// Sizable maps types to their sizing information
var sizable = map[string]struct {
	Unit string
	Verb string
}{
	"string": {Unit: "characters", Verb: "to have"},
	"file":   {Unit: "bytes", Verb: "to have"},
	"array":  {Unit: "items", Verb: "to have"},
	"slice":  {Unit: "items", Verb: "to have"},
	"set":    {Unit: "items", Verb: "to have"},
}

// getSizing returns sizing information for a given type
func getSizing(origin string) *struct {
	Unit string
	Verb string
} {
	if info, ok := sizable[origin]; ok {
		return &info
	}
	return nil
}

// parsedType determines the type of a value
func parsedType(data interface{}) string {
	switch gozod.GetParsedType(data) {
	case gozod.ParsedTypeString:
		return "string"
	case gozod.ParsedTypeNumber:
		return "number"
	case gozod.ParsedTypeBigint:
		return "bigint"
	case gozod.ParsedTypeBool:
		return "bool"
	case gozod.ParsedTypeFloat:
		return "float"
	case gozod.ParsedTypeObject:
		return "object"
	case gozod.ParsedTypeFunction:
		return "function"
	case gozod.ParsedTypeFile:
		return "file"
	case gozod.ParsedTypeDate:
		return "date"
	case gozod.ParsedTypeArray:
		return "array"
	case gozod.ParsedTypeSlice:
		return "slice"
	case gozod.ParsedTypeMap:
		return "map"
	case gozod.ParsedTypeNaN:
		return "NaN"
	case gozod.ParsedTypeNil:
		return "nil"
	case gozod.ParsedTypeComplex:
		return "complex"
	default:
		return string(gozod.GetParsedType(data))
	}
}

// Nouns maps string formats to their human-readable names
var nouns = map[string]string{
	"regex":            "input",
	"email":            "email address",
	"url":              "URL",
	"emoji":            "emoji",
	"uuid":             "UUID",
	"uuidv4":           "UUIDv4",
	"uuidv6":           "UUIDv6",
	"nanoid":           "nanoid",
	"guid":             "GUID",
	"cuid":             "cuid",
	"cuid2":            "cuid2",
	"ulid":             "ULID",
	"xid":              "XID",
	"ksuid":            "KSUID",
	"datetime":         "ISO datetime",
	"date":             "ISO date",
	"time":             "ISO time",
	"duration":         "ISO duration",
	"ipv4":             "IPv4 address",
	"ipv6":             "IPv6 address",
	"cidrv4":           "IPv4 range",
	"cidrv6":           "IPv6 range",
	"base64":           "base64-encoded string",
	"base64url":        "base64url-encoded string",
	"json_string":      "JSON string",
	"e164":             "E.164 number",
	"jwt":              "JWT",
	"template_literal": "input",
}

func errorEn(issue gozod.ZodRawIssue) string {
	switch issue.Code {
	case "invalid_type":
		return fmt.Sprintf("Invalid input: expected %s, received %s", issue.GetExpected(), parsedType(issue.Input))
	case "invalid_value":
		values := issue.GetValues()
		if len(values) == 1 {
			return fmt.Sprintf("Invalid input: expected %s", gozod.StringifyPrimitive(values[0]))
		}
		return fmt.Sprintf("Invalid option: expected one of %s", gozod.JoinValues(values, "|"))
	case "too_big":
		adj := "<"
		if issue.GetInclusive() {
			adj = "<="
		}
		origin := issue.GetOrigin()
		sizing := getSizing(origin)
		if sizing != nil {
			if origin == "" {
				origin = "value"
			}
			unit := sizing.Unit
			if unit == "" {
				unit = "elements"
			}
			if origin == "slice" || origin == "array" {
				return fmt.Sprintf("Too big: expected %s to have %s%v %s", origin, adj, issue.GetMaximum(), unit)
			}
			return fmt.Sprintf("Too big: expected %s to have %s%v %s", origin, adj, issue.GetMaximum(), unit)
		}
		if origin == "" {
			origin = "value"
		}
		return fmt.Sprintf("Too big: expected %s to be %s%v", origin, adj, issue.GetMaximum())
	case "too_small":
		adj := ">"
		if issue.GetInclusive() {
			adj = ">="
		}
		origin := issue.GetOrigin()
		sizing := getSizing(origin)
		if sizing != nil {
			// Use consistent format for slice/array types to match main error generation
			if origin == "slice" || origin == "array" {
				return fmt.Sprintf("Too small: expected %s to have %s%v %s", origin, adj, issue.GetMinimum(), sizing.Unit)
			}
			return fmt.Sprintf("Too small: expected %s to have %s%v %s", origin, adj, issue.GetMinimum(), sizing.Unit)
		}
		return fmt.Sprintf("Too small: expected %s to be %s%v", origin, adj, issue.GetMinimum())
	case "invalid_format":
		format := issue.GetFormat()
		if format == "starts_with" {
			return fmt.Sprintf("Invalid string: must start with \"%s\"", issue.GetPrefix())
		}
		if format == "ends_with" {
			return fmt.Sprintf("Invalid string: must end with \"%s\"", issue.GetSuffix())
		}
		if format == "includes" {
			return fmt.Sprintf("Invalid string: must include \"%s\"", issue.GetIncludes())
		}
		if format == "regex" {
			return fmt.Sprintf("Invalid string: must match pattern %s", issue.GetPattern())
		}
		noun := nouns[format]
		if noun == "" {
			noun = format
		}
		return fmt.Sprintf("Invalid %s", noun)
	case "not_multiple_of":
		return fmt.Sprintf("Invalid number: must be a multiple of %v", issue.GetDivisor())
	case "unrecognized_keys":
		keys := issue.GetKeys()
		plural := ""
		if len(keys) > 1 {
			plural = "s"
		}
		keyInterfaces := make([]interface{}, len(keys))
		for i, k := range keys {
			keyInterfaces[i] = k
		}
		return fmt.Sprintf("Unrecognized key%s: %s", plural, gozod.JoinValues(keyInterfaces, ", "))
	case "invalid_key":
		return fmt.Sprintf("Invalid key in %s", issue.GetOrigin())
	case "invalid_union":
		return "Invalid input"
	case "invalid_element":
		return fmt.Sprintf("Invalid value in %s", issue.GetOrigin())
	default:
		return "Invalid input"
	}
}

// En returns the default English locale error map function
func En() func(gozod.ZodRawIssue) string {
	return errorEn
}
