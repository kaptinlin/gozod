package locales

import (
	"fmt"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// =============================================================================
// URDU LOCALE FORMATTER
// =============================================================================

// Urdu sizing info mappings
var SizableUr = map[string]issues.SizingInfo{
	"string": {Unit: "حروف", Verb: "ہونا"},
	"file":   {Unit: "بائٹس", Verb: "ہونا"},
	"array":  {Unit: "آئٹمز", Verb: "ہونا"},
	"slice":  {Unit: "آئٹمز", Verb: "ہونا"},
	"set":    {Unit: "آئٹمز", Verb: "ہونا"},
	"map":    {Unit: "اندراجات", Verb: "ہونا"},
}

// Urdu format noun mappings
var FormatNounsUr = map[string]string{
	"regex":            "ان پٹ",
	"email":            "ای میل ایڈریس",
	"url":              "یو آر ایل",
	"emoji":            "ایموجی",
	"uuid":             "یو یو آئی ڈی",
	"uuidv4":           "یو یو آئی ڈی وی 4",
	"uuidv6":           "یو یو آئی ڈی وی 6",
	"nanoid":           "نینو آئی ڈی",
	"guid":             "جی یو آئی ڈی",
	"cuid":             "سی یو آئی ڈی",
	"cuid2":            "سی یو آئی ڈی 2",
	"ulid":             "یو ایل آئی ڈی",
	"xid":              "ایکس آئی ڈی",
	"ksuid":            "کے ایس یو آئی ڈی",
	"datetime":         "آئی ایس او ڈیٹ ٹائم",
	"date":             "آئی ایس او تاریخ",
	"time":             "آئی ایس او وقت",
	"duration":         "آئی ایس او مدت",
	"ipv4":             "آئی پی وی 4 ایڈریس",
	"ipv6":             "آئی پی وی 6 ایڈریس",
	"mac":              "میک ایڈریس",
	"cidrv4":           "آئی پی وی 4 رینج",
	"cidrv6":           "آئی پی وی 6 رینج",
	"base64":           "بیس 64 ان کوڈڈ سٹرنگ",
	"base64url":        "بیس 64 یو آر ایل ان کوڈڈ سٹرنگ",
	"json_string":      "جے ایس او این سٹرنگ",
	"e164":             "ای 164 نمبر",
	"jwt":              "جے ڈبلیو ٹی",
	"template_literal": "ان پٹ",
}

// Urdu type dictionary
var TypeDictionaryUr = map[string]string{
	"nan":       "NaN",
	"number":    "نمبر",
	"array":     "آرے",
	"slice":     "آرے",
	"string":    "سٹرنگ",
	"bool":      "بولین",
	"object":    "آبجیکٹ",
	"map":       "میپ",
	"nil":       "نل",
	"null":      "نل",
	"undefined": "غیر متعین",
	"function":  "فنکشن",
	"date":      "تاریخ",
	"file":      "فائل",
	"set":       "سیٹ",
}

// getSizingUr returns Urdu sizing information for a given type
func getSizingUr(origin string) *issues.SizingInfo {
	if _, exists := SizableUr[origin]; exists {
		return new(SizableUr[origin])
	}
	return nil
}

// getFormatNounUr returns Urdu noun for a format name
func getFormatNounUr(format string) string {
	if noun, exists := FormatNounsUr[format]; exists {
		return noun
	}
	return format
}

// getTypeNameUr returns Urdu type name
func getTypeNameUr(typeName string) string {
	if name, exists := TypeDictionaryUr[typeName]; exists {
		return name
	}
	return typeName
}

// formatUr provides Urdu error messages
func formatUr(raw core.ZodRawIssue) string {
	code := raw.Code

	switch code {
	case core.InvalidType:
		expected := mapx.StringOr(raw.Properties, "expected", "")
		expected = getTypeNameUr(expected)
		received := issues.ParsedTypeToString(raw.Input)
		received = getTypeNameUr(received)
		return fmt.Sprintf("غلط ان پٹ: %s متوقع تھا، %s موصول ہوا", expected, received)

	case core.InvalidValue:
		values := mapx.AnySliceOr(raw.Properties, "values", nil)
		if len(values) == 0 {
			return "غلط ویلیو"
		}
		if len(values) == 1 {
			return fmt.Sprintf("غلط ان پٹ: %s متوقع تھا", issues.StringifyPrimitive(values[0]))
		}
		return fmt.Sprintf("غلط آپشن: %s میں سے ایک متوقع تھا",
			issues.JoinValuesWithSeparator(values, "|"))

	case core.TooBig:
		return formatSizeConstraintUr(raw, false)

	case core.TooSmall:
		return formatSizeConstraintUr(raw, true)

	case core.InvalidFormat:
		format := mapx.StringOr(raw.Properties, "format", "")
		if format == "" {
			return "غلط فارمیٹ"
		}
		return formatStringValidationUr(raw, format)

	case core.NotMultipleOf:
		divisor := mapx.AnyOr(raw.Properties, "divisor", nil)
		if divisor == nil {
			return "غلط نمبر: مضاعف ہونا چاہیے"
		}
		return fmt.Sprintf("غلط نمبر: %v کا مضاعف ہونا چاہیے", divisor)

	case core.UnrecognizedKeys:
		keys := mapx.StringsOr(raw.Properties, "keys", nil)
		if len(keys) == 0 {
			return "غیر تسلیم شدہ کی"
		}
		keyWord := "غیر تسلیم شدہ کی"
		if len(keys) > 1 {
			keyWord = "غیر تسلیم شدہ کیز"
		}
		if keysAny, err := slicex.ToAny(keys); err == nil && !slicex.IsEmpty(keysAny) {
			keysJoined := issues.JoinValuesWithSeparator(keysAny, "، ")
			return fmt.Sprintf("%s: %s", keyWord, keysJoined)
		}
		return keyWord

	case core.InvalidKey:
		origin := mapx.StringOr(raw.Properties, "origin", "")
		if origin == "" {
			return "غلط کی"
		}
		return fmt.Sprintf("%s میں غلط کی", origin)

	case core.InvalidUnion:
		return "غلط ان پٹ"

	case core.InvalidElement:
		origin := mapx.StringOr(raw.Properties, "origin", "")
		if origin == "" {
			return "غلط عنصر"
		}
		return fmt.Sprintf("%s میں غلط ویلیو", origin)

	case core.MissingRequired:
		fieldName := mapx.StringOr(raw.Properties, "field_name", "")
		fieldType := mapx.StringOr(raw.Properties, "field_type", "فیلڈ")
		if fieldName == "" {
			return fmt.Sprintf("مطلوبہ %s غائب ہے", fieldType)
		}
		return fmt.Sprintf("مطلوبہ %s غائب ہے: %s", fieldType, fieldName)

	case core.TypeConversion:
		fromType := mapx.StringOr(raw.Properties, "from_type", "نامعلوم")
		toType := mapx.StringOr(raw.Properties, "to_type", "نامعلوم")
		return fmt.Sprintf("ٹائپ کنورژن ناکام: %s کو %s میں تبدیل نہیں کیا جا سکتا", fromType, toType)

	case core.InvalidSchema:
		reason := mapx.StringOr(raw.Properties, "reason", "")
		if reason != "" {
			return fmt.Sprintf("غلط اسکیما: %s", reason)
		}
		return "غلط اسکیما ڈیفینیشن"

	case core.InvalidDiscriminator:
		field := mapx.StringOr(raw.Properties, "field", "ڈسکریمینیٹر")
		return fmt.Sprintf("غلط یا غائب ڈسکریمینیٹر فیلڈ: %s", field)

	case core.IncompatibleTypes:
		conflictType := mapx.StringOr(raw.Properties, "conflict_type", "ویلیوز")
		return fmt.Sprintf("%s کو ضم نہیں کیا جا سکتا: غیر موافق ٹائپس", conflictType)

	case core.NilPointer:
		return "نل پوائنٹر پایا گیا"

	case core.Custom:
		message := mapx.StringOr(raw.Properties, "message", "")
		if message != "" {
			return message
		}
		return "غلط ان پٹ"

	default:
		return "غلط ان پٹ"
	}
}

// formatSizeConstraintUr formats size constraint messages in Urdu
func formatSizeConstraintUr(raw core.ZodRawIssue, isTooSmall bool) string {
	origin := mapx.StringOr(raw.Properties, "origin", "")
	if origin == "" {
		origin = "ویلیو"
	}

	var threshold any
	if isTooSmall {
		threshold = mapx.AnyOr(raw.Properties, "minimum", nil)
	} else {
		threshold = mapx.AnyOr(raw.Properties, "maximum", nil)
	}

	if threshold == nil {
		if isTooSmall {
			return "بہت چھوٹا"
		}
		return "بہت بڑا"
	}

	inclusive := mapx.BoolOr(raw.Properties, "inclusive", true)
	sizing := getSizingUr(mapx.StringOr(raw.Properties, "origin", ""))
	thresholdStr := issues.FormatThreshold(threshold)

	// Urdu comparison operators
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
			return fmt.Sprintf("بہت چھوٹا: %s کے %s%s %s ہونے متوقع تھے", origin, adj, thresholdStr, sizing.Unit)
		}
		return fmt.Sprintf("بہت بڑا: %s کے %s%s %s ہونے متوقع تھے", origin, adj, thresholdStr, sizing.Unit)
	}

	if isTooSmall {
		return fmt.Sprintf("بہت چھوٹا: %s کا %s%s ہونا متوقع تھا", origin, adj, thresholdStr)
	}
	return fmt.Sprintf("بہت بڑا: %s کا %s%s ہونا متوقع تھا", origin, adj, thresholdStr)
}

// formatStringValidationUr handles string format validation messages in Urdu
func formatStringValidationUr(raw core.ZodRawIssue, format string) string {
	switch format {
	case "starts_with":
		prefix := mapx.StringOr(raw.Properties, "prefix", "")
		if prefix == "" {
			return "غلط سٹرنگ: مخصوص پریفکس سے شروع ہونا چاہیے"
		}
		return fmt.Sprintf("غلط سٹرنگ: \"%s\" سے شروع ہونا چاہیے", prefix)
	case "ends_with":
		suffix := mapx.StringOr(raw.Properties, "suffix", "")
		if suffix == "" {
			return "غلط سٹرنگ: مخصوص سفکس پر ختم ہونا چاہیے"
		}
		return fmt.Sprintf("غلط سٹرنگ: \"%s\" پر ختم ہونا چاہیے", suffix)
	case "includes":
		includes := mapx.StringOr(raw.Properties, "includes", "")
		if includes == "" {
			return "غلط سٹرنگ: مخصوص سب سٹرنگ شامل ہونا چاہیے"
		}
		return fmt.Sprintf("غلط سٹرنگ: \"%s\" شامل ہونا چاہیے", includes)
	case "regex":
		pattern := mapx.StringOr(raw.Properties, "pattern", "")
		if pattern == "" {
			return "غلط سٹرنگ: پیٹرن سے میچ ہونا چاہیے"
		}
		return fmt.Sprintf("غلط سٹرنگ: پیٹرن %s سے میچ ہونا چاہیے", pattern)
	default:
		noun := getFormatNounUr(format)
		return fmt.Sprintf("غلط %s", noun)
	}
}

// Ur returns a ZodConfig configured for Urdu locale.
func Ur() *core.ZodConfig {
	return &core.ZodConfig{
		LocaleError: formatUr,
	}
}

// FormatMessageUr formats a single issue using Urdu locale
func FormatMessageUr(issue core.ZodRawIssue) string {
	return formatUr(issue)
}
