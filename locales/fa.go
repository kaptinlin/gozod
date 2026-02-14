//nolint:staticcheck // Persian locale contains Unicode format characters
package locales

import (
	"fmt"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// =============================================================================
// PERSIAN (FARSI) LOCALE FORMATTER
// =============================================================================

// Persian sizing info mappings
var SizableFa = map[string]issues.SizingInfo{
	"string": {Unit: "کاراکتر", Verb: "داشته باشد"},
	"file":   {Unit: "بایت", Verb: "داشته باشد"},
	"array":  {Unit: "آیتم", Verb: "داشته باشد"},
	"slice":  {Unit: "آیتم", Verb: "داشته باشد"},
	"set":    {Unit: "آیتم", Verb: "داشته باشد"},
	"map":    {Unit: "ورودی", Verb: "داشته باشد"},
}

// Persian format noun mappings
var FormatNounsFa = map[string]string{
	"regex":            "ورودی",
	"email":            "آدرس ایمیل",
	"url":              "URL",
	"emoji":            "ایموجی",
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
	"datetime":         "تاریخ و زمان ایزو",
	"date":             "تاریخ ایزو",
	"time":             "زمان ایزو",
	"duration":         "مدت زمان ایزو",
	"ipv4":             "آدرس IPv4",
	"ipv6":             "آدرس IPv6",
	"mac":              "آدرس MAC",
	"cidrv4":           "دامنه IPv4",
	"cidrv6":           "دامنه IPv6",
	"base64":           "رشته base64",
	"base64url":        "رشته base64url",
	"json_string":      "رشته JSON",
	"e164":             "عدد E.164",
	"jwt":              "JWT",
	"template_literal": "ورودی",
}

// Persian type dictionary
var TypeDictionaryFa = map[string]string{
	"nan":       "NaN",
	"number":    "عدد",
	"array":     "آرایه",
	"slice":     "آرایه",
	"string":    "رشته",
	"bool":      "بولی",
	"object":    "شیء",
	"map":       "نقشه",
	"nil":       "تهی",
	"undefined": "تعریف نشده",
	"function":  "تابع",
	"date":      "تاریخ",
	"file":      "فایل",
	"set":       "مجموعه",
}

// getSizingFa returns Persian sizing information for a given type
func getSizingFa(origin string) *issues.SizingInfo {
	if _, exists := SizableFa[origin]; exists {
		return new(SizableFa[origin])
	}
	return nil
}

// getFormatNounFa returns Persian noun for a format name
func getFormatNounFa(format string) string {
	if noun, exists := FormatNounsFa[format]; exists {
		return noun
	}
	return format
}

// getTypeNameFa returns Persian type name
func getTypeNameFa(typeName string) string {
	if name, exists := TypeDictionaryFa[typeName]; exists {
		return name
	}
	return typeName
}

// formatFa provides Persian error messages
func formatFa(raw core.ZodRawIssue) string {
	code := raw.Code

	switch code {
	case core.InvalidType:
		expected := mapx.StringOr(raw.Properties, "expected", "")
		expected = getTypeNameFa(expected)
		received := issues.ParsedTypeToString(raw.Input)
		received = getTypeNameFa(received)
		return fmt.Sprintf("ورودی نامعتبر: می‌بایست %s می‌بود، %s دریافت شد", expected, received)

	case core.InvalidValue:
		values := mapx.AnySliceOr(raw.Properties, "values", nil)
		if len(values) == 0 {
			return "مقدار نامعتبر"
		}
		if len(values) == 1 {
			return fmt.Sprintf("ورودی نامعتبر: می‌بایست %s می‌بود", issues.StringifyPrimitive(values[0]))
		}
		return fmt.Sprintf("گزینه نامعتبر: می‌بایست یکی از %s می‌بود",
			issues.JoinValuesWithSeparator(values, "|"))

	case core.TooBig:
		return formatSizeConstraintFa(raw, false)

	case core.TooSmall:
		return formatSizeConstraintFa(raw, true)

	case core.InvalidFormat:
		format := mapx.StringOr(raw.Properties, "format", "")
		if format == "" {
			return "فرمت نامعتبر"
		}
		return formatStringValidationFa(raw, format)

	case core.NotMultipleOf:
		divisor := mapx.AnyOr(raw.Properties, "divisor", nil)
		if divisor == nil {
			return "عدد نامعتبر: باید مضرب باشد"
		}
		return fmt.Sprintf("عدد نامعتبر: باید مضرب %v باشد", divisor)

	case core.UnrecognizedKeys:
		keys := mapx.StringsOr(raw.Properties, "keys", nil)
		if len(keys) == 0 {
			return "کلید ناشناس"
		}
		keyWord := "کلید ناشناس"
		if len(keys) > 1 {
			keyWord = "کلیدهای ناشناس"
		}
		if keysAny, err := slicex.ToAny(keys); err == nil && !slicex.IsEmpty(keysAny) {
			keysJoined := issues.JoinValuesWithSeparator(keysAny, ", ")
			return fmt.Sprintf("%s: %s", keyWord, keysJoined)
		}
		return keyWord

	case core.InvalidKey:
		origin := mapx.StringOr(raw.Properties, "origin", "")
		if origin == "" {
			return "کلید نامعتبر"
		}
		return fmt.Sprintf("کلید ناشناس در %s", origin)

	case core.InvalidUnion:
		return "ورودی نامعتبر"

	case core.InvalidElement:
		origin := mapx.StringOr(raw.Properties, "origin", "")
		if origin == "" {
			return "عنصر نامعتبر"
		}
		return fmt.Sprintf("مقدار نامعتبر در %s", origin)

	case core.MissingRequired:
		fieldName := mapx.StringOr(raw.Properties, "field_name", "")
		fieldType := mapx.StringOr(raw.Properties, "field_type", "فیلد")
		if fieldName == "" {
			return fmt.Sprintf("%s اجباری وجود ندارد", fieldType)
		}
		return fmt.Sprintf("%s اجباری وجود ندارد: %s", fieldType, fieldName)

	case core.TypeConversion:
		fromType := mapx.StringOr(raw.Properties, "from_type", "ناشناخته")
		toType := mapx.StringOr(raw.Properties, "to_type", "ناشناخته")
		return fmt.Sprintf("تبدیل نوع ناموفق: نمی‌توان %s را به %s تبدیل کرد", fromType, toType)

	case core.InvalidSchema:
		reason := mapx.StringOr(raw.Properties, "reason", "")
		if reason != "" {
			return fmt.Sprintf("اسکیما نامعتبر: %s", reason)
		}
		return "تعریف اسکیما نامعتبر"

	case core.InvalidDiscriminator:
		field := mapx.StringOr(raw.Properties, "field", "تفکیک‌کننده")
		return fmt.Sprintf("فیلد تفکیک‌کننده نامعتبر یا ناموجود: %s", field)

	case core.IncompatibleTypes:
		conflictType := mapx.StringOr(raw.Properties, "conflict_type", "مقادیر")
		return fmt.Sprintf("نمی‌توان %s را ادغام کرد: انواع ناسازگار", conflictType)

	case core.NilPointer:
		return "اشاره‌گر تهی یافت شد"

	case core.Custom:
		message := mapx.StringOr(raw.Properties, "message", "")
		if message != "" {
			return message
		}
		return "ورودی نامعتبر"

	default:
		return "ورودی نامعتبر"
	}
}

// formatSizeConstraintFa formats size constraint messages in Persian
func formatSizeConstraintFa(raw core.ZodRawIssue, isTooSmall bool) string {
	origin := mapx.StringOr(raw.Properties, "origin", "")
	if origin == "" {
		origin = "مقدار"
	}

	var threshold any
	if isTooSmall {
		threshold = mapx.AnyOr(raw.Properties, "minimum", nil)
	} else {
		threshold = mapx.AnyOr(raw.Properties, "maximum", nil)
	}

	if threshold == nil {
		if isTooSmall {
			return "خیلی کوچک"
		}
		return "خیلی بزرگ"
	}

	inclusive := mapx.BoolOr(raw.Properties, "inclusive", true)
	sizing := getSizingFa(mapx.StringOr(raw.Properties, "origin", ""))
	thresholdStr := issues.FormatThreshold(threshold)

	// Persian comparison operators
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
			return fmt.Sprintf("خیلی کوچک: %s باید %s%s %s باشد", origin, adj, thresholdStr, sizing.Unit)
		}
		return fmt.Sprintf("خیلی بزرگ: %s باید %s%s %s باشد", origin, adj, thresholdStr, sizing.Unit)
	}

	if isTooSmall {
		return fmt.Sprintf("خیلی کوچک: %s باید %s%s باشد", origin, adj, thresholdStr)
	}
	return fmt.Sprintf("خیلی بزرگ: %s باید %s%s باشد", origin, adj, thresholdStr)
}

// formatStringValidationFa handles string format validation messages in Persian
func formatStringValidationFa(raw core.ZodRawIssue, format string) string {
	switch format {
	case "starts_with":
		prefix := mapx.StringOr(raw.Properties, "prefix", "")
		if prefix == "" {
			return "رشته نامعتبر: باید با پیشوند مشخص شده شروع شود"
		}
		return fmt.Sprintf("رشته نامعتبر: باید با \"%s\" شروع شود", prefix)
	case "ends_with":
		suffix := mapx.StringOr(raw.Properties, "suffix", "")
		if suffix == "" {
			return "رشته نامعتبر: باید با پسوند مشخص شده تمام شود"
		}
		return fmt.Sprintf("رشته نامعتبر: باید با \"%s\" تمام شود", suffix)
	case "includes":
		includes := mapx.StringOr(raw.Properties, "includes", "")
		if includes == "" {
			return "رشته نامعتبر: باید شامل زیررشته مشخص شده باشد"
		}
		return fmt.Sprintf("رشته نامعتبر: باید شامل \"%s\" باشد", includes)
	case "regex":
		pattern := mapx.StringOr(raw.Properties, "pattern", "")
		if pattern == "" {
			return "رشته نامعتبر: باید با الگو مطابقت داشته باشد"
		}
		return fmt.Sprintf("رشته نامعتبر: باید با الگوی %s مطابقت داشته باشد", pattern)
	default:
		noun := getFormatNounFa(format)
		return fmt.Sprintf("%s نامعتبر", noun)
	}
}

// Fa returns a ZodConfig configured for Persian (Farsi) locale.
func Fa() *core.ZodConfig {
	return &core.ZodConfig{
		LocaleError: formatFa,
	}
}

// FormatMessageFa formats a single issue using Persian locale
func FormatMessageFa(issue core.ZodRawIssue) string {
	return formatFa(issue)
}
