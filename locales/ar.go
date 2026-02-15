package locales

import (
	"fmt"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// =============================================================================
// ARABIC LOCALE FORMATTER
// =============================================================================

// SizableAr maps Arabic sizing information.
var SizableAr = map[string]issues.SizingInfo{
	"string": {Unit: "حرف", Verb: "أن يحوي"},
	"file":   {Unit: "بايت", Verb: "أن يحوي"},
	"array":  {Unit: "عنصر", Verb: "أن يحوي"},
	"slice":  {Unit: "عنصر", Verb: "أن يحوي"},
	"set":    {Unit: "عنصر", Verb: "أن يحوي"},
	"map":    {Unit: "مدخل", Verb: "أن يحوي"},
}

// FormatNounsAr maps Arabic format noun translations.
var FormatNounsAr = map[string]string{
	"regex":            "مدخل",
	"email":            "بريد إلكتروني",
	"url":              "رابط",
	"emoji":            "إيموجي",
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
	"datetime":         "تاريخ ووقت بمعيار ISO",
	"date":             "تاريخ بمعيار ISO",
	"time":             "وقت بمعيار ISO",
	"duration":         "مدة بمعيار ISO",
	"ipv4":             "عنوان IPv4",
	"ipv6":             "عنوان IPv6",
	"mac":              "عنوان MAC",
	"cidrv4":           "مدى عناوين بصيغة IPv4",
	"cidrv6":           "مدى عناوين بصيغة IPv6",
	"base64":           "نَص بترميز base64",
	"base64url":        "نَص بترميز base64url",
	"json_string":      "نَص على هيئة JSON",
	"e164":             "رقم هاتف بمعيار E.164",
	"jwt":              "JWT",
	"template_literal": "مدخل",
}

// TypeDictionaryAr maps Arabic type name translations.
var TypeDictionaryAr = map[string]string{
	"nan":       "NaN",
	"number":    "رقم",
	"array":     "مصفوفة",
	"slice":     "مصفوفة",
	"string":    "نَص",
	"bool":      "قيمة منطقية",
	"object":    "كائن",
	"map":       "خريطة",
	"nil":       "فارغ",
	"undefined": "غير معرّف",
	"function":  "دالة",
	"date":      "تاريخ",
	"file":      "ملف",
	"set":       "مجموعة",
}

// getSizingAr returns Arabic sizing information for a given type
func getSizingAr(origin string) *issues.SizingInfo {
	if _, exists := SizableAr[origin]; exists {
		return new(SizableAr[origin])
	}
	return nil
}

// getFormatNounAr returns Arabic noun for a format name
func getFormatNounAr(format string) string {
	if noun, exists := FormatNounsAr[format]; exists {
		return noun
	}
	return format
}

// getTypeNameAr returns Arabic type name
func getTypeNameAr(typeName string) string {
	if name, exists := TypeDictionaryAr[typeName]; exists {
		return name
	}
	return typeName
}

// formatAr provides Arabic error messages
func formatAr(raw core.ZodRawIssue) string {
	code := raw.Code

	switch code {
	case core.InvalidType:
		expected := mapx.StringOr(raw.Properties, "expected", "")
		expected = getTypeNameAr(expected)
		received := issues.ParsedTypeToString(raw.Input)
		received = getTypeNameAr(received)
		return fmt.Sprintf("مدخلات غير مقبولة: يفترض إدخال %s، ولكن تم إدخال %s", expected, received)

	case core.InvalidValue:
		values := mapx.AnySliceOr(raw.Properties, "values", nil)
		if len(values) == 0 {
			return "قيمة غير مقبولة"
		}
		if len(values) == 1 {
			return fmt.Sprintf("مدخلات غير مقبولة: يفترض إدخال %s", issues.StringifyPrimitive(values[0]))
		}
		return fmt.Sprintf("اختيار غير مقبول: يتوقع انتقاء أحد هذه الخيارات: %s",
			issues.JoinValuesWithSeparator(values, "|"))

	case core.TooBig:
		return formatSizeConstraintAr(raw, false)

	case core.TooSmall:
		return formatSizeConstraintAr(raw, true)

	case core.InvalidFormat:
		format := mapx.StringOr(raw.Properties, "format", "")
		if format == "" {
			return "صيغة غير مقبولة"
		}
		return formatStringValidationAr(raw, format)

	case core.NotMultipleOf:
		divisor := mapx.AnyOr(raw.Properties, "divisor", nil)
		if divisor == nil {
			return "رقم غير مقبول: يجب أن يكون من المضاعفات"
		}
		return fmt.Sprintf("رقم غير مقبول: يجب أن يكون من مضاعفات %v", divisor)

	case core.UnrecognizedKeys:
		keys := mapx.StringsOr(raw.Properties, "keys", nil)
		if len(keys) == 0 {
			return "معرف غريب"
		}
		keyWord := "معرف غريب"
		if len(keys) > 1 {
			keyWord = "معرفات غريبة"
		}
		if keysAny, err := slicex.ToAny(keys); err == nil && !slicex.IsEmpty(keysAny) {
			keysJoined := issues.JoinValuesWithSeparator(keysAny, "، ")
			return fmt.Sprintf("%s: %s", keyWord, keysJoined)
		}
		return keyWord

	case core.InvalidKey:
		origin := mapx.StringOr(raw.Properties, "origin", "")
		if origin == "" {
			return "معرف غير مقبول"
		}
		return fmt.Sprintf("معرف غير مقبول في %s", origin)

	case core.InvalidUnion:
		return "مدخل غير مقبول"

	case core.InvalidElement:
		origin := mapx.StringOr(raw.Properties, "origin", "")
		if origin == "" {
			return "عنصر غير مقبول"
		}
		return fmt.Sprintf("مدخل غير مقبول في %s", origin)

	case core.MissingRequired:
		fieldName := mapx.StringOr(raw.Properties, "field_name", "")
		fieldType := mapx.StringOr(raw.Properties, "field_type", "حقل")
		if fieldName == "" {
			return fmt.Sprintf("%s مطلوب مفقود", fieldType)
		}
		return fmt.Sprintf("%s مطلوب مفقود: %s", fieldType, fieldName)

	case core.TypeConversion:
		fromType := mapx.StringOr(raw.Properties, "from_type", "غير معروف")
		toType := mapx.StringOr(raw.Properties, "to_type", "غير معروف")
		return fmt.Sprintf("فشل تحويل النوع: لا يمكن تحويل %s إلى %s", fromType, toType)

	case core.InvalidSchema:
		reason := mapx.StringOr(raw.Properties, "reason", "")
		if reason != "" {
			return fmt.Sprintf("مخطط غير مقبول: %s", reason)
		}
		return "تعريف مخطط غير مقبول"

	case core.InvalidDiscriminator:
		field := mapx.StringOr(raw.Properties, "field", "المميز")
		return fmt.Sprintf("حقل مميز غير مقبول أو مفقود: %s", field)

	case core.IncompatibleTypes:
		conflictType := mapx.StringOr(raw.Properties, "conflict_type", "القيم")
		return fmt.Sprintf("لا يمكن دمج %s: أنواع غير متوافقة", conflictType)

	case core.NilPointer:
		return "تم اكتشاف مؤشر فارغ"

	case core.Custom:
		message := mapx.StringOr(raw.Properties, "message", "")
		if message != "" {
			return message
		}
		return "مدخل غير مقبول"

	default:
		return "مدخل غير مقبول"
	}
}

// formatSizeConstraintAr formats size constraint messages in Arabic
func formatSizeConstraintAr(raw core.ZodRawIssue, isTooSmall bool) string {
	origin := mapx.StringOr(raw.Properties, "origin", "")
	if origin == "" {
		origin = "القيمة"
	}

	var threshold any
	if isTooSmall {
		threshold = mapx.AnyOr(raw.Properties, "minimum", nil)
	} else {
		threshold = mapx.AnyOr(raw.Properties, "maximum", nil)
	}

	if threshold == nil {
		if isTooSmall {
			return "أصغر من اللازم"
		}
		return "أكبر من اللازم"
	}

	inclusive := mapx.BoolOr(raw.Properties, "inclusive", true)
	sizing := getSizingAr(mapx.StringOr(raw.Properties, "origin", ""))
	thresholdStr := issues.FormatThreshold(threshold)

	// Arabic comparison operators
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
			return fmt.Sprintf("أصغر من اللازم: يفترض لـ %s أن يكون %s %s %s", origin, adj, thresholdStr, sizing.Unit)
		}
		return fmt.Sprintf("أكبر من اللازم: يفترض أن تكون %s %s %s %s", origin, adj, thresholdStr, sizing.Unit)
	}

	if isTooSmall {
		return fmt.Sprintf("أصغر من اللازم: يفترض لـ %s أن يكون %s %s", origin, adj, thresholdStr)
	}
	return fmt.Sprintf("أكبر من اللازم: يفترض أن تكون %s %s %s", origin, adj, thresholdStr)
}

// formatStringValidationAr handles string format validation messages in Arabic
func formatStringValidationAr(raw core.ZodRawIssue, format string) string {
	switch format {
	case "starts_with":
		prefix := mapx.StringOr(raw.Properties, "prefix", "")
		if prefix == "" {
			return "نَص غير مقبول: يجب أن يبدأ بالبادئة المحددة"
		}
		return fmt.Sprintf("نَص غير مقبول: يجب أن يبدأ بـ \"%s\"", prefix)
	case "ends_with":
		suffix := mapx.StringOr(raw.Properties, "suffix", "")
		if suffix == "" {
			return "نَص غير مقبول: يجب أن ينتهي باللاحقة المحددة"
		}
		return fmt.Sprintf("نَص غير مقبول: يجب أن ينتهي بـ \"%s\"", suffix)
	case "includes":
		includes := mapx.StringOr(raw.Properties, "includes", "")
		if includes == "" {
			return "نَص غير مقبول: يجب أن يتضمَّن السلسلة الفرعية المحددة"
		}
		return fmt.Sprintf("نَص غير مقبول: يجب أن يتضمَّن \"%s\"", includes)
	case "regex":
		pattern := mapx.StringOr(raw.Properties, "pattern", "")
		if pattern == "" {
			return "نَص غير مقبول: يجب أن يطابق النمط"
		}
		return fmt.Sprintf("نَص غير مقبول: يجب أن يطابق النمط %s", pattern)
	default:
		noun := getFormatNounAr(format)
		return fmt.Sprintf("%s غير مقبول", noun)
	}
}

// Ar returns a ZodConfig configured for Arabic locale.
func Ar() *core.ZodConfig {
	return &core.ZodConfig{
		LocaleError: formatAr,
	}
}

// FormatMessageAr formats a single issue using Arabic locale
func FormatMessageAr(issue core.ZodRawIssue) string {
	return formatAr(issue)
}
