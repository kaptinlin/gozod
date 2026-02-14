package locales

import (
	"fmt"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// =============================================================================
// UKRAINIAN LOCALE FORMATTER
// =============================================================================

// Ukrainian sizing info mappings
var SizableUk = map[string]issues.SizingInfo{
	"string": {Unit: "символів", Verb: "матиме"},
	"file":   {Unit: "байтів", Verb: "матиме"},
	"array":  {Unit: "елементів", Verb: "матиме"},
	"slice":  {Unit: "елементів", Verb: "матиме"},
	"set":    {Unit: "елементів", Verb: "матиме"},
	"map":    {Unit: "записів", Verb: "матиме"},
}

// Ukrainian format noun mappings
var FormatNounsUk = map[string]string{
	"regex":            "вхідні дані",
	"email":            "адреса електронної пошти",
	"url":              "URL",
	"emoji":            "емодзі",
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
	"datetime":         "дата та час ISO",
	"date":             "дата ISO",
	"time":             "час ISO",
	"duration":         "тривалість ISO",
	"ipv4":             "адреса IPv4",
	"ipv6":             "адреса IPv6",
	"mac":              "адреса MAC",
	"cidrv4":           "діапазон IPv4",
	"cidrv6":           "діапазон IPv6",
	"base64":           "рядок у кодуванні base64",
	"base64url":        "рядок у кодуванні base64url",
	"json_string":      "рядок JSON",
	"e164":             "номер E.164",
	"jwt":              "JWT",
	"template_literal": "вхідні дані",
}

// Ukrainian type dictionary
var TypeDictionaryUk = map[string]string{
	"nan":       "NaN",
	"number":    "число",
	"array":     "масив",
	"slice":     "масив",
	"string":    "рядок",
	"bool":      "булеве значення",
	"object":    "об'єкт",
	"map":       "карта",
	"nil":       "null",
	"undefined": "undefined",
	"function":  "функція",
	"date":      "дата",
	"file":      "файл",
	"set":       "множина",
}

// getSizingUk returns Ukrainian sizing information for a given type
func getSizingUk(origin string) *issues.SizingInfo {
	if _, exists := SizableUk[origin]; exists {
		return new(SizableUk[origin])
	}
	return nil
}

// getFormatNounUk returns Ukrainian noun for a format name
func getFormatNounUk(format string) string {
	if noun, exists := FormatNounsUk[format]; exists {
		return noun
	}
	return format
}

// getTypeNameUk returns Ukrainian type name
func getTypeNameUk(typeName string) string {
	if name, exists := TypeDictionaryUk[typeName]; exists {
		return name
	}
	return typeName
}

// formatUk provides Ukrainian error messages
func formatUk(raw core.ZodRawIssue) string {
	code := raw.Code

	switch code {
	case core.InvalidType:
		expected := mapx.StringOr(raw.Properties, "expected", "")
		expected = getTypeNameUk(expected)
		received := issues.ParsedTypeToString(raw.Input)
		received = getTypeNameUk(received)
		return fmt.Sprintf("Неправильні вхідні дані: очікується %s, отримано %s", expected, received)

	case core.InvalidValue:
		values := mapx.AnySliceOr(raw.Properties, "values", nil)
		if len(values) == 0 {
			return "Неправильне значення"
		}
		if len(values) == 1 {
			return fmt.Sprintf("Неправильні вхідні дані: очікується %s", issues.StringifyPrimitive(values[0]))
		}
		return fmt.Sprintf("Неправильна опція: очікується одне з %s",
			issues.JoinValuesWithSeparator(values, "|"))

	case core.TooBig:
		return formatSizeConstraintUk(raw, false)

	case core.TooSmall:
		return formatSizeConstraintUk(raw, true)

	case core.InvalidFormat:
		format := mapx.StringOr(raw.Properties, "format", "")
		if format == "" {
			return "Неправильний формат"
		}
		return formatStringValidationUk(raw, format)

	case core.NotMultipleOf:
		divisor := mapx.AnyOr(raw.Properties, "divisor", nil)
		if divisor == nil {
			return "Неправильне число: повинно бути кратним"
		}
		return fmt.Sprintf("Неправильне число: повинно бути кратним %v", divisor)

	case core.UnrecognizedKeys:
		keys := mapx.StringsOr(raw.Properties, "keys", nil)
		if len(keys) == 0 {
			return "Нерозпізнаний ключ"
		}
		keyWord := "Нерозпізнаний ключ"
		if len(keys) > 1 {
			keyWord = "Нерозпізнані ключі"
		}
		if keysAny, err := slicex.ToAny(keys); err == nil && !slicex.IsEmpty(keysAny) {
			keysJoined := issues.JoinValuesWithSeparator(keysAny, ", ")
			return fmt.Sprintf("%s: %s", keyWord, keysJoined)
		}
		return keyWord

	case core.InvalidKey:
		origin := mapx.StringOr(raw.Properties, "origin", "")
		if origin == "" {
			return "Неправильний ключ"
		}
		return fmt.Sprintf("Неправильний ключ у %s", origin)

	case core.InvalidUnion:
		return "Неправильні вхідні дані"

	case core.InvalidElement:
		origin := mapx.StringOr(raw.Properties, "origin", "")
		if origin == "" {
			return "Неправильний елемент"
		}
		return fmt.Sprintf("Неправильне значення у %s", origin)

	case core.MissingRequired:
		fieldName := mapx.StringOr(raw.Properties, "field_name", "")
		fieldType := mapx.StringOr(raw.Properties, "field_type", "поле")
		if fieldName == "" {
			return fmt.Sprintf("Відсутнє обов'язкове %s", fieldType)
		}
		return fmt.Sprintf("Відсутнє обов'язкове %s: %s", fieldType, fieldName)

	case core.TypeConversion:
		fromType := mapx.StringOr(raw.Properties, "from_type", "невідомий")
		toType := mapx.StringOr(raw.Properties, "to_type", "невідомий")
		return fmt.Sprintf("Помилка перетворення типу: неможливо перетворити %s на %s", fromType, toType)

	case core.InvalidSchema:
		reason := mapx.StringOr(raw.Properties, "reason", "")
		if reason != "" {
			return fmt.Sprintf("Неправильна схема: %s", reason)
		}
		return "Неправильне визначення схеми"

	case core.InvalidDiscriminator:
		field := mapx.StringOr(raw.Properties, "field", "дискримінатор")
		return fmt.Sprintf("Неправильне або відсутнє поле дискримінатора: %s", field)

	case core.IncompatibleTypes:
		conflictType := mapx.StringOr(raw.Properties, "conflict_type", "значення")
		return fmt.Sprintf("Неможливо об'єднати %s: несумісні типи", conflictType)

	case core.NilPointer:
		return "Виявлено нульовий вказівник"

	case core.Custom:
		message := mapx.StringOr(raw.Properties, "message", "")
		if message != "" {
			return message
		}
		return "Неправильні вхідні дані"

	default:
		return "Неправильні вхідні дані"
	}
}

// formatSizeConstraintUk formats size constraint messages in Ukrainian
func formatSizeConstraintUk(raw core.ZodRawIssue, isTooSmall bool) string {
	origin := mapx.StringOr(raw.Properties, "origin", "")
	if origin == "" {
		origin = "значення"
	}

	var threshold any
	if isTooSmall {
		threshold = mapx.AnyOr(raw.Properties, "minimum", nil)
	} else {
		threshold = mapx.AnyOr(raw.Properties, "maximum", nil)
	}

	if threshold == nil {
		if isTooSmall {
			return "Занадто мале"
		}
		return "Занадто велике"
	}

	inclusive := mapx.BoolOr(raw.Properties, "inclusive", true)
	sizing := getSizingUk(mapx.StringOr(raw.Properties, "origin", ""))
	thresholdStr := issues.FormatThreshold(threshold)

	// Ukrainian comparison operators
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
			return fmt.Sprintf("Занадто мале: очікується, що %s %s %s%s %s", origin, sizing.Verb, adj, thresholdStr, sizing.Unit)
		}
		return fmt.Sprintf("Занадто велике: очікується, що %s %s %s%s %s", origin, sizing.Verb, adj, thresholdStr, sizing.Unit)
	}

	if isTooSmall {
		return fmt.Sprintf("Занадто мале: очікується, що %s буде %s%s", origin, adj, thresholdStr)
	}
	return fmt.Sprintf("Занадто велике: очікується, що %s буде %s%s", origin, adj, thresholdStr)
}

// formatStringValidationUk handles string format validation messages in Ukrainian
func formatStringValidationUk(raw core.ZodRawIssue, format string) string {
	switch format {
	case "starts_with":
		prefix := mapx.StringOr(raw.Properties, "prefix", "")
		if prefix == "" {
			return "Неправильний рядок: повинен починатися з вказаного префікса"
		}
		return fmt.Sprintf("Неправильний рядок: повинен починатися з \"%s\"", prefix)
	case "ends_with":
		suffix := mapx.StringOr(raw.Properties, "suffix", "")
		if suffix == "" {
			return "Неправильний рядок: повинен закінчуватися вказаним суфіксом"
		}
		return fmt.Sprintf("Неправильний рядок: повинен закінчуватися на \"%s\"", suffix)
	case "includes":
		includes := mapx.StringOr(raw.Properties, "includes", "")
		if includes == "" {
			return "Неправильний рядок: повинен містити вказаний підрядок"
		}
		return fmt.Sprintf("Неправильний рядок: повинен містити \"%s\"", includes)
	case "regex":
		pattern := mapx.StringOr(raw.Properties, "pattern", "")
		if pattern == "" {
			return "Неправильний рядок: повинен відповідати шаблону"
		}
		return fmt.Sprintf("Неправильний рядок: повинен відповідати шаблону %s", pattern)
	default:
		noun := getFormatNounUk(format)
		return fmt.Sprintf("Неправильний %s", noun)
	}
}

// Uk returns a ZodConfig configured for Ukrainian locale.
func Uk() *core.ZodConfig {
	return &core.ZodConfig{
		LocaleError: formatUk,
	}
}

// FormatMessageUk formats a single issue using Ukrainian locale
func FormatMessageUk(issue core.ZodRawIssue) string {
	return formatUk(issue)
}
