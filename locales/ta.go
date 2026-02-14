package locales

import (
	"fmt"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// =============================================================================
// TAMIL LOCALE FORMATTER
// =============================================================================

// Tamil sizing info mappings
var SizableTa = map[string]issues.SizingInfo{
	"string": {Unit: "எழுத்துக்கள்", Verb: "கொண்டிருக்க வேண்டும்"},
	"file":   {Unit: "பைட்டுகள்", Verb: "கொண்டிருக்க வேண்டும்"},
	"array":  {Unit: "உறுப்புகள்", Verb: "கொண்டிருக்க வேண்டும்"},
	"slice":  {Unit: "உறுப்புகள்", Verb: "கொண்டிருக்க வேண்டும்"},
	"set":    {Unit: "உறுப்புகள்", Verb: "கொண்டிருக்க வேண்டும்"},
	"map":    {Unit: "உள்ளீடுகள்", Verb: "கொண்டிருக்க வேண்டும்"},
}

// Tamil format noun mappings
var FormatNounsTa = map[string]string{
	"regex":            "உள்ளீடு",
	"email":            "மின்னஞ்சல் முகவரி",
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
	"datetime":         "ISO தேதி நேரம்",
	"date":             "ISO தேதி",
	"time":             "ISO நேரம்",
	"duration":         "ISO கால அளவு",
	"ipv4":             "IPv4 முகவரி",
	"ipv6":             "IPv6 முகவரி",
	"mac":              "MAC முகவரி",
	"cidrv4":           "IPv4 வரம்பு",
	"cidrv6":           "IPv6 வரம்பு",
	"base64":           "base64-encoded சரம்",
	"base64url":        "base64url-encoded சரம்",
	"json_string":      "JSON சரம்",
	"e164":             "E.164 எண்",
	"jwt":              "JWT",
	"template_literal": "உள்ளீடு",
}

// Tamil type dictionary
var TypeDictionaryTa = map[string]string{
	"nan":       "NaN",
	"number":    "எண்",
	"array":     "அணி",
	"slice":     "அணி",
	"string":    "சரம்",
	"bool":      "பூலியன்",
	"object":    "பொருள்",
	"map":       "வரைபடம்",
	"nil":       "வெறுமை",
	"null":      "வெறுமை",
	"undefined": "வரையறுக்கப்படாதது",
	"function":  "செயல்பாடு",
	"date":      "தேதி",
	"file":      "கோப்பு",
	"set":       "தொகுப்பு",
}

// getSizingTa returns Tamil sizing information for a given type
func getSizingTa(origin string) *issues.SizingInfo {
	if _, exists := SizableTa[origin]; exists {
		return new(SizableTa[origin])
	}
	return nil
}

// getFormatNounTa returns Tamil noun for a format name
func getFormatNounTa(format string) string {
	if noun, exists := FormatNounsTa[format]; exists {
		return noun
	}
	return format
}

// getTypeNameTa returns Tamil type name
func getTypeNameTa(typeName string) string {
	if name, exists := TypeDictionaryTa[typeName]; exists {
		return name
	}
	return typeName
}

// formatTa provides Tamil error messages
func formatTa(raw core.ZodRawIssue) string {
	code := raw.Code

	switch code {
	case core.InvalidType:
		expected := mapx.StringOr(raw.Properties, "expected", "")
		expected = getTypeNameTa(expected)
		received := issues.ParsedTypeToString(raw.Input)
		received = getTypeNameTa(received)
		return fmt.Sprintf("தவறான உள்ளீடு: எதிர்பார்க்கப்பட்டது %s, பெறப்பட்டது %s", expected, received)

	case core.InvalidValue:
		values := mapx.AnySliceOr(raw.Properties, "values", nil)
		if len(values) == 0 {
			return "தவறான மதிப்பு"
		}
		if len(values) == 1 {
			return fmt.Sprintf("தவறான உள்ளீடு: எதிர்பார்க்கப்பட்டது %s", issues.StringifyPrimitive(values[0]))
		}
		return fmt.Sprintf("தவறான விருப்பம்: எதிர்பார்க்கப்பட்டது %s இல் ஒன்று",
			issues.JoinValuesWithSeparator(values, "|"))

	case core.TooBig:
		return formatSizeConstraintTa(raw, false)

	case core.TooSmall:
		return formatSizeConstraintTa(raw, true)

	case core.InvalidFormat:
		format := mapx.StringOr(raw.Properties, "format", "")
		if format == "" {
			return "தவறான வடிவம்"
		}
		return formatStringValidationTa(raw, format)

	case core.NotMultipleOf:
		divisor := mapx.AnyOr(raw.Properties, "divisor", nil)
		if divisor == nil {
			return "தவறான எண்: பலமாக இருக்க வேண்டும்"
		}
		return fmt.Sprintf("தவறான எண்: %v இன் பலமாக இருக்க வேண்டும்", divisor)

	case core.UnrecognizedKeys:
		keys := mapx.StringsOr(raw.Properties, "keys", nil)
		if len(keys) == 0 {
			return "அடையாளம் தெரியாத விசை"
		}
		keyWord := "அடையாளம் தெரியாத விசை"
		if len(keys) > 1 {
			keyWord = "அடையாளம் தெரியாத விசைகள்"
		}
		if keysAny, err := slicex.ToAny(keys); err == nil && !slicex.IsEmpty(keysAny) {
			keysJoined := issues.JoinValuesWithSeparator(keysAny, ", ")
			return fmt.Sprintf("%s: %s", keyWord, keysJoined)
		}
		return keyWord

	case core.InvalidKey:
		origin := mapx.StringOr(raw.Properties, "origin", "")
		if origin == "" {
			return "தவறான விசை"
		}
		return fmt.Sprintf("%s இல் தவறான விசை", origin)

	case core.InvalidUnion:
		return "தவறான உள்ளீடு"

	case core.InvalidElement:
		origin := mapx.StringOr(raw.Properties, "origin", "")
		if origin == "" {
			return "தவறான உறுப்பு"
		}
		return fmt.Sprintf("%s இல் தவறான மதிப்பு", origin)

	case core.MissingRequired:
		fieldName := mapx.StringOr(raw.Properties, "field_name", "")
		fieldType := mapx.StringOr(raw.Properties, "field_type", "புலம்")
		if fieldName == "" {
			return fmt.Sprintf("தேவையான %s இல்லை", fieldType)
		}
		return fmt.Sprintf("தேவையான %s இல்லை: %s", fieldType, fieldName)

	case core.TypeConversion:
		fromType := mapx.StringOr(raw.Properties, "from_type", "தெரியாதது")
		toType := mapx.StringOr(raw.Properties, "to_type", "தெரியாதது")
		return fmt.Sprintf("வகை மாற்றம் தோல்வியடைந்தது: %s ஐ %s ஆக மாற்ற முடியவில்லை", fromType, toType)

	case core.InvalidSchema:
		reason := mapx.StringOr(raw.Properties, "reason", "")
		if reason != "" {
			return fmt.Sprintf("தவறான திட்டம்: %s", reason)
		}
		return "தவறான திட்ட வரையறை"

	case core.InvalidDiscriminator:
		field := mapx.StringOr(raw.Properties, "field", "வேறுபாடு காட்டி")
		return fmt.Sprintf("தவறான அல்லது இல்லாத வேறுபாடு காட்டி புலம்: %s", field)

	case core.IncompatibleTypes:
		conflictType := mapx.StringOr(raw.Properties, "conflict_type", "மதிப்புகள்")
		return fmt.Sprintf("%s ஐ இணைக்க முடியவில்லை: பொருந்தாத வகைகள்", conflictType)

	case core.NilPointer:
		return "வெற்று சுட்டி கண்டறியப்பட்டது"

	case core.Custom:
		message := mapx.StringOr(raw.Properties, "message", "")
		if message != "" {
			return message
		}
		return "தவறான உள்ளீடு"

	default:
		return "தவறான உள்ளீடு"
	}
}

// formatSizeConstraintTa formats size constraint messages in Tamil
func formatSizeConstraintTa(raw core.ZodRawIssue, isTooSmall bool) string {
	origin := mapx.StringOr(raw.Properties, "origin", "")
	if origin == "" {
		origin = "மதிப்பு"
	}

	var threshold any
	if isTooSmall {
		threshold = mapx.AnyOr(raw.Properties, "minimum", nil)
	} else {
		threshold = mapx.AnyOr(raw.Properties, "maximum", nil)
	}

	if threshold == nil {
		if isTooSmall {
			return "மிகச் சிறியது"
		}
		return "மிக பெரியது"
	}

	inclusive := mapx.BoolOr(raw.Properties, "inclusive", true)
	sizing := getSizingTa(mapx.StringOr(raw.Properties, "origin", ""))
	thresholdStr := issues.FormatThreshold(threshold)

	// Tamil comparison operators
	var adj string
	if inclusive {
		if isTooSmall {
			adj = ">="
		} else {
			adj = "<="
		}
	} else {
		if isTooSmall {
			adj = ">"
		} else {
			adj = "<"
		}
	}

	if sizing != nil {
		if isTooSmall {
			return fmt.Sprintf("மிகச் சிறியது: எதிர்பார்க்கப்பட்டது %s %s%s %s ஆக இருக்க வேண்டும்", origin, adj, thresholdStr, sizing.Unit)
		}
		return fmt.Sprintf("மிக பெரியது: எதிர்பார்க்கப்பட்டது %s %s%s %s ஆக இருக்க வேண்டும்", origin, adj, thresholdStr, sizing.Unit)
	}

	if isTooSmall {
		return fmt.Sprintf("மிகச் சிறியது: எதிர்பார்க்கப்பட்டது %s %s%s ஆக இருக்க வேண்டும்", origin, adj, thresholdStr)
	}
	return fmt.Sprintf("மிக பெரியது: எதிர்பார்க்கப்பட்டது %s %s%s ஆக இருக்க வேண்டும்", origin, adj, thresholdStr)
}

// formatStringValidationTa handles string format validation messages in Tamil
func formatStringValidationTa(raw core.ZodRawIssue, format string) string {
	switch format {
	case "starts_with":
		prefix := mapx.StringOr(raw.Properties, "prefix", "")
		if prefix == "" {
			return "தவறான சரம்: குறிப்பிட்ட முன்னொட்டில் தொடங்க வேண்டும்"
		}
		return fmt.Sprintf("தவறான சரம்: \"%s\" இல் தொடங்க வேண்டும்", prefix)
	case "ends_with":
		suffix := mapx.StringOr(raw.Properties, "suffix", "")
		if suffix == "" {
			return "தவறான சரம்: குறிப்பிட்ட பின்னொட்டில் முடிவடைய வேண்டும்"
		}
		return fmt.Sprintf("தவறான சரம்: \"%s\" இல் முடிவடைய வேண்டும்", suffix)
	case "includes":
		includes := mapx.StringOr(raw.Properties, "includes", "")
		if includes == "" {
			return "தவறான சரம்: குறிப்பிட்ட துணை சரத்தை உள்ளடக்க வேண்டும்"
		}
		return fmt.Sprintf("தவறான சரம்: \"%s\" ஐ உள்ளடக்க வேண்டும்", includes)
	case "regex":
		pattern := mapx.StringOr(raw.Properties, "pattern", "")
		if pattern == "" {
			return "தவறான சரம்: முறைபாட்டுடன் பொருந்த வேண்டும்"
		}
		return fmt.Sprintf("தவறான சரம்: %s முறைபாட்டுடன் பொருந்த வேண்டும்", pattern)
	default:
		noun := getFormatNounTa(format)
		return fmt.Sprintf("தவறான %s", noun)
	}
}

// Ta returns a ZodConfig configured for Tamil locale.
func Ta() *core.ZodConfig {
	return &core.ZodConfig{
		LocaleError: formatTa,
	}
}

// FormatMessageTa formats a single issue using Tamil locale
func FormatMessageTa(issue core.ZodRawIssue) string {
	return formatTa(issue)
}
