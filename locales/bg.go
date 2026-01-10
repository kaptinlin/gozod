package locales

import (
	"fmt"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// =============================================================================
// BULGARIAN LOCALE FORMATTER
// =============================================================================

// Bulgarian sizing info mappings
var SizableBg = map[string]issues.SizingInfo{
	"string": {Unit: "символа", Verb: "да съдържа"},
	"file":   {Unit: "байта", Verb: "да съдържа"},
	"array":  {Unit: "елемента", Verb: "да съдържа"},
	"slice":  {Unit: "елемента", Verb: "да съдържа"},
	"set":    {Unit: "елемента", Verb: "да съдържа"},
	"map":    {Unit: "записа", Verb: "да съдържа"},
}

// Bulgarian format noun mappings with gender
var FormatNounsBg = map[string]string{
	"regex":            "вход",
	"email":            "имейл адрес",
	"url":              "URL",
	"emoji":            "емоджи",
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
	"datetime":         "ISO време",
	"date":             "ISO дата",
	"time":             "ISO време",
	"duration":         "ISO продължителност",
	"ipv4":             "IPv4 адрес",
	"ipv6":             "IPv6 адрес",
	"mac":              "MAC адрес",
	"cidrv4":           "IPv4 диапазон",
	"cidrv6":           "IPv6 диапазон",
	"base64":           "base64-кодиран низ",
	"base64url":        "base64url-кодиран низ",
	"json_string":      "JSON низ",
	"e164":             "E.164 номер",
	"jwt":              "JWT",
	"template_literal": "вход",
}

// Bulgarian type dictionary
var TypeDictionaryBg = map[string]string{
	"nan":       "NaN",
	"number":    "число",
	"array":     "масив",
	"slice":     "масив",
	"string":    "низ",
	"bool":      "булево",
	"object":    "обект",
	"map":       "карта",
	"nil":       "null",
	"undefined": "неопределено",
	"function":  "функция",
	"date":      "дата",
	"file":      "файл",
	"set":       "множество",
}

// Bulgarian gender adjectives for format validation
var FormatGenderBg = map[string]string{
	"emoji":    "Невалидно",
	"datetime": "Невалидно",
	"date":     "Невалидна",
	"time":     "Невалидно",
	"duration": "Невалидна",
}

// getSizingBg returns Bulgarian sizing information for a given type
func getSizingBg(origin string) *issues.SizingInfo {
	if info, exists := SizableBg[origin]; exists {
		return &info
	}
	return nil
}

// getFormatNounBg returns Bulgarian noun for a format name
func getFormatNounBg(format string) string {
	if noun, exists := FormatNounsBg[format]; exists {
		return noun
	}
	return format
}

// getTypeNameBg returns Bulgarian type name
func getTypeNameBg(typeName string) string {
	if name, exists := TypeDictionaryBg[typeName]; exists {
		return name
	}
	return typeName
}

// getInvalidAdjBg returns the correct gender form of "invalid" in Bulgarian
func getInvalidAdjBg(format string) string {
	if adj, exists := FormatGenderBg[format]; exists {
		return adj
	}
	return "Невалиден"
}

// formatBg provides Bulgarian error messages
func formatBg(raw core.ZodRawIssue) string {
	code := raw.Code

	switch code {
	case core.InvalidType:
		expected := mapx.GetStringDefault(raw.Properties, "expected", "")
		expected = getTypeNameBg(expected)
		received := issues.ParsedTypeToString(raw.Input)
		received = getTypeNameBg(received)
		return fmt.Sprintf("Невалиден вход: очакван %s, получен %s", expected, received)

	case core.InvalidValue:
		values := mapx.GetAnySliceDefault(raw.Properties, "values", nil)
		if len(values) == 0 {
			return "Невалидна стойност"
		}
		if len(values) == 1 {
			return fmt.Sprintf("Невалиден вход: очакван %s", issues.StringifyPrimitive(values[0]))
		}
		return fmt.Sprintf("Невалидна опция: очаквано едно от %s",
			issues.JoinValuesWithSeparator(values, "|"))

	case core.TooBig:
		return formatSizeConstraintBg(raw, false)

	case core.TooSmall:
		return formatSizeConstraintBg(raw, true)

	case core.InvalidFormat:
		format := mapx.GetStringDefault(raw.Properties, "format", "")
		if format == "" {
			return "Невалиден формат"
		}
		return formatStringValidationBg(raw, format)

	case core.NotMultipleOf:
		divisor := mapx.GetAnyDefault(raw.Properties, "divisor", nil)
		if divisor == nil {
			return "Невалидно число: трябва да бъде кратно"
		}
		return fmt.Sprintf("Невалидно число: трябва да бъде кратно на %v", divisor)

	case core.UnrecognizedKeys:
		keys := mapx.GetStringsDefault(raw.Properties, "keys", nil)
		if len(keys) == 0 {
			return "Неразпознат ключ"
		}
		keyWord := "Неразпознат ключ"
		if len(keys) > 1 {
			keyWord = "Неразпознати ключове"
		}
		if keysAny, err := slicex.ToAny(keys); err == nil && !slicex.IsEmpty(keysAny) {
			keysJoined := issues.JoinValuesWithSeparator(keysAny, ", ")
			return fmt.Sprintf("%s: %s", keyWord, keysJoined)
		}
		return keyWord

	case core.InvalidKey:
		origin := mapx.GetStringDefault(raw.Properties, "origin", "")
		if origin == "" {
			return "Невалиден ключ"
		}
		return fmt.Sprintf("Невалиден ключ в %s", origin)

	case core.InvalidUnion:
		return "Невалиден вход"

	case core.InvalidElement:
		origin := mapx.GetStringDefault(raw.Properties, "origin", "")
		if origin == "" {
			return "Невалиден елемент"
		}
		return fmt.Sprintf("Невалидна стойност в %s", origin)

	case core.MissingRequired:
		fieldName := mapx.GetStringDefault(raw.Properties, "field_name", "")
		fieldType := mapx.GetStringDefault(raw.Properties, "field_type", "поле")
		if fieldName == "" {
			return fmt.Sprintf("Липсва задължително %s", fieldType)
		}
		return fmt.Sprintf("Липсва задължително %s: %s", fieldType, fieldName)

	case core.TypeConversion:
		fromType := mapx.GetStringDefault(raw.Properties, "from_type", "неизвестен")
		toType := mapx.GetStringDefault(raw.Properties, "to_type", "неизвестен")
		return fmt.Sprintf("Неуспешно преобразуване на тип: не може да се преобразува %s в %s", fromType, toType)

	case core.InvalidSchema:
		reason := mapx.GetStringDefault(raw.Properties, "reason", "")
		if reason != "" {
			return fmt.Sprintf("Невалидна схема: %s", reason)
		}
		return "Невалидна дефиниция на схема"

	case core.InvalidDiscriminator:
		field := mapx.GetStringDefault(raw.Properties, "field", "дискриминатор")
		return fmt.Sprintf("Невалидно или липсващо дискриминаторно поле: %s", field)

	case core.IncompatibleTypes:
		conflictType := mapx.GetStringDefault(raw.Properties, "conflict_type", "стойности")
		return fmt.Sprintf("Не може да се обедини %s: несъвместими типове", conflictType)

	case core.NilPointer:
		return "Открит нулев указател"

	case core.Custom:
		message := mapx.GetStringDefault(raw.Properties, "message", "")
		if message != "" {
			return message
		}
		return "Невалиден вход"

	default:
		return "Невалиден вход"
	}
}

// formatSizeConstraintBg formats size constraint messages in Bulgarian
func formatSizeConstraintBg(raw core.ZodRawIssue, isTooSmall bool) string {
	origin := mapx.GetStringDefault(raw.Properties, "origin", "")
	if origin == "" {
		origin = "стойност"
	}

	var threshold any
	if isTooSmall {
		threshold = mapx.GetAnyDefault(raw.Properties, "minimum", nil)
	} else {
		threshold = mapx.GetAnyDefault(raw.Properties, "maximum", nil)
	}

	if threshold == nil {
		if isTooSmall {
			return "Твърде малко"
		}
		return "Твърде голямо"
	}

	inclusive := mapx.GetBoolDefault(raw.Properties, "inclusive", true)
	sizing := getSizingBg(mapx.GetStringDefault(raw.Properties, "origin", ""))
	thresholdStr := issues.FormatThreshold(threshold)

	// Bulgarian comparison operators
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
			return fmt.Sprintf("Твърде малко: очаква се %s да съдържа %s%s %s", origin, adj, thresholdStr, sizing.Unit)
		}
		return fmt.Sprintf("Твърде голямо: очаква се %s да съдържа %s%s %s", origin, adj, thresholdStr, sizing.Unit)
	}

	if isTooSmall {
		return fmt.Sprintf("Твърде малко: очаква се %s да бъде %s%s", origin, adj, thresholdStr)
	}
	return fmt.Sprintf("Твърде голямо: очаква се %s да бъде %s%s", origin, adj, thresholdStr)
}

// formatStringValidationBg handles string format validation messages in Bulgarian
func formatStringValidationBg(raw core.ZodRawIssue, format string) string {
	switch format {
	case "starts_with":
		prefix := mapx.GetStringDefault(raw.Properties, "prefix", "")
		if prefix == "" {
			return "Невалиден низ: трябва да започва с определен префикс"
		}
		return fmt.Sprintf("Невалиден низ: трябва да започва с \"%s\"", prefix)
	case "ends_with":
		suffix := mapx.GetStringDefault(raw.Properties, "suffix", "")
		if suffix == "" {
			return "Невалиден низ: трябва да завършва с определен суфикс"
		}
		return fmt.Sprintf("Невалиден низ: трябва да завършва с \"%s\"", suffix)
	case "includes":
		includes := mapx.GetStringDefault(raw.Properties, "includes", "")
		if includes == "" {
			return "Невалиден низ: трябва да включва определен подниз"
		}
		return fmt.Sprintf("Невалиден низ: трябва да включва \"%s\"", includes)
	case "regex":
		pattern := mapx.GetStringDefault(raw.Properties, "pattern", "")
		if pattern == "" {
			return "Невалиден низ: трябва да съвпада с шаблона"
		}
		return fmt.Sprintf("Невалиден низ: трябва да съвпада с %s", pattern)
	default:
		noun := getFormatNounBg(format)
		invalidAdj := getInvalidAdjBg(format)
		return fmt.Sprintf("%s %s", invalidAdj, noun)
	}
}

// Bg returns a ZodConfig configured for Bulgarian locale.
func Bg() *core.ZodConfig {
	return &core.ZodConfig{
		LocaleError: formatBg,
	}
}

// FormatMessageBg formats a single issue using Bulgarian locale
func FormatMessageBg(issue core.ZodRawIssue) string {
	return formatBg(issue)
}
