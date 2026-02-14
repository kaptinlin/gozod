package locales

import (
	"fmt"
	"math"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// =============================================================================
// RUSSIAN LOCALE FORMATTER
// =============================================================================

// getRussianPlural returns the correct plural form for Russian
// Russian has 3 plural forms: one (1, 21, 31...), few (2-4, 22-24...), many (5-20, 25-30...)
func getRussianPlural(count int, one, few, many string) string {
	absCount := int(math.Abs(float64(count)))
	lastDigit := absCount % 10
	lastTwoDigits := absCount % 100

	if lastTwoDigits >= 11 && lastTwoDigits <= 19 {
		return many
	}

	if lastDigit == 1 {
		return one
	}

	if lastDigit >= 2 && lastDigit <= 4 {
		return few
	}

	return many
}

// RussianSizable represents sizing info with Russian plural forms
type RussianSizable struct {
	UnitOne  string
	UnitFew  string
	UnitMany string
	Verb     string
}

// Russian sizing info mappings with plural forms
var SizableRu = map[string]RussianSizable{
	"string": {UnitOne: "символ", UnitFew: "символа", UnitMany: "символов", Verb: "иметь"},
	"file":   {UnitOne: "байт", UnitFew: "байта", UnitMany: "байт", Verb: "иметь"},
	"array":  {UnitOne: "элемент", UnitFew: "элемента", UnitMany: "элементов", Verb: "иметь"},
	"slice":  {UnitOne: "элемент", UnitFew: "элемента", UnitMany: "элементов", Verb: "иметь"},
	"set":    {UnitOne: "элемент", UnitFew: "элемента", UnitMany: "элементов", Verb: "иметь"},
	"map":    {UnitOne: "запись", UnitFew: "записи", UnitMany: "записей", Verb: "иметь"},
}

// Russian format noun mappings
var FormatNounsRu = map[string]string{
	"regex":            "ввод",
	"email":            "email адрес",
	"url":              "URL",
	"emoji":            "эмодзи",
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
	"datetime":         "ISO дата и время",
	"date":             "ISO дата",
	"time":             "ISO время",
	"duration":         "ISO длительность",
	"ipv4":             "IPv4 адрес",
	"ipv6":             "IPv6 адрес",
	"mac":              "MAC адрес",
	"cidrv4":           "IPv4 диапазон",
	"cidrv6":           "IPv6 диапазон",
	"base64":           "строка в формате base64",
	"base64url":        "строка в формате base64url",
	"json_string":      "JSON строка",
	"e164":             "номер E.164",
	"jwt":              "JWT",
	"template_literal": "ввод",
}

// Russian type dictionary
var TypeDictionaryRu = map[string]string{
	"nan":       "NaN",
	"number":    "число",
	"array":     "массив",
	"slice":     "массив",
	"string":    "строка",
	"bool":      "логическое значение",
	"object":    "объект",
	"map":       "карта",
	"nil":       "null",
	"undefined": "undefined",
	"function":  "функция",
	"date":      "дата",
	"file":      "файл",
	"set":       "множество",
}

// getSizingRu returns Russian sizing information for a given type
func getSizingRu(origin string) *RussianSizable {
	if _, exists := SizableRu[origin]; exists {
		return new(SizableRu[origin])
	}
	return nil
}

// getFormatNounRu returns Russian noun for a format name
func getFormatNounRu(format string) string {
	if noun, exists := FormatNounsRu[format]; exists {
		return noun
	}
	return format
}

// getTypeNameRu returns Russian type name
func getTypeNameRu(typeName string) string {
	if name, exists := TypeDictionaryRu[typeName]; exists {
		return name
	}
	return typeName
}

// formatRu provides Russian error messages
func formatRu(raw core.ZodRawIssue) string {
	code := raw.Code

	switch code {
	case core.InvalidType:
		expected := mapx.StringOr(raw.Properties, "expected", "")
		expected = getTypeNameRu(expected)
		received := issues.ParsedTypeToString(raw.Input)
		received = getTypeNameRu(received)
		return fmt.Sprintf("Неверный ввод: ожидалось %s, получено %s", expected, received)

	case core.InvalidValue:
		values := mapx.AnySliceOr(raw.Properties, "values", nil)
		if len(values) == 0 {
			return "Неверное значение"
		}
		if len(values) == 1 {
			return fmt.Sprintf("Неверный ввод: ожидалось %s", issues.StringifyPrimitive(values[0]))
		}
		return fmt.Sprintf("Неверный вариант: ожидалось одно из %s",
			issues.JoinValuesWithSeparator(values, "|"))

	case core.TooBig:
		return formatSizeConstraintRu(raw, false)

	case core.TooSmall:
		return formatSizeConstraintRu(raw, true)

	case core.InvalidFormat:
		format := mapx.StringOr(raw.Properties, "format", "")
		if format == "" {
			return "Неверный формат"
		}
		return formatStringValidationRu(raw, format)

	case core.NotMultipleOf:
		divisor := mapx.AnyOr(raw.Properties, "divisor", nil)
		if divisor == nil {
			return "Неверное число: должно быть кратным"
		}
		return fmt.Sprintf("Неверное число: должно быть кратным %v", divisor)

	case core.UnrecognizedKeys:
		keys := mapx.StringsOr(raw.Properties, "keys", nil)
		if len(keys) == 0 {
			return "Нераспознанный ключ"
		}
		keyWord := "Нераспознанный ключ"
		if len(keys) > 1 {
			keyWord = "Нераспознанные ключи"
		}
		if keysAny, err := slicex.ToAny(keys); err == nil && !slicex.IsEmpty(keysAny) {
			keysJoined := issues.JoinValuesWithSeparator(keysAny, ", ")
			return fmt.Sprintf("%s: %s", keyWord, keysJoined)
		}
		return keyWord

	case core.InvalidKey:
		origin := mapx.StringOr(raw.Properties, "origin", "")
		if origin == "" {
			return "Неверный ключ"
		}
		return fmt.Sprintf("Неверный ключ в %s", origin)

	case core.InvalidUnion:
		return "Неверные входные данные"

	case core.InvalidElement:
		origin := mapx.StringOr(raw.Properties, "origin", "")
		if origin == "" {
			return "Неверный элемент"
		}
		return fmt.Sprintf("Неверное значение в %s", origin)

	case core.MissingRequired:
		fieldName := mapx.StringOr(raw.Properties, "field_name", "")
		fieldType := mapx.StringOr(raw.Properties, "field_type", "поле")
		if fieldName == "" {
			return fmt.Sprintf("Отсутствует обязательное %s", fieldType)
		}
		return fmt.Sprintf("Отсутствует обязательное %s: %s", fieldType, fieldName)

	case core.TypeConversion:
		fromType := mapx.StringOr(raw.Properties, "from_type", "неизвестный")
		toType := mapx.StringOr(raw.Properties, "to_type", "неизвестный")
		return fmt.Sprintf("Ошибка преобразования типа: невозможно преобразовать %s в %s", fromType, toType)

	case core.InvalidSchema:
		reason := mapx.StringOr(raw.Properties, "reason", "")
		if reason != "" {
			return fmt.Sprintf("Неверная схема: %s", reason)
		}
		return "Неверное определение схемы"

	case core.InvalidDiscriminator:
		field := mapx.StringOr(raw.Properties, "field", "дискриминатор")
		return fmt.Sprintf("Неверное или отсутствующее поле дискриминатора: %s", field)

	case core.IncompatibleTypes:
		conflictType := mapx.StringOr(raw.Properties, "conflict_type", "значения")
		return fmt.Sprintf("Невозможно объединить %s: несовместимые типы", conflictType)

	case core.NilPointer:
		return "Обнаружен нулевой указатель"

	case core.Custom:
		message := mapx.StringOr(raw.Properties, "message", "")
		if message != "" {
			return message
		}
		return "Неверные входные данные"

	default:
		return "Неверные входные данные"
	}
}

// formatSizeConstraintRu formats size constraint messages in Russian with proper pluralization
func formatSizeConstraintRu(raw core.ZodRawIssue, isTooSmall bool) string {
	origin := mapx.StringOr(raw.Properties, "origin", "")
	if origin == "" {
		origin = "значение"
	}

	var threshold any
	if isTooSmall {
		threshold = mapx.AnyOr(raw.Properties, "minimum", nil)
	} else {
		threshold = mapx.AnyOr(raw.Properties, "maximum", nil)
	}

	if threshold == nil {
		if isTooSmall {
			return "Слишком маленькое значение"
		}
		return "Слишком большое значение"
	}

	inclusive := mapx.BoolOr(raw.Properties, "inclusive", true)
	sizing := getSizingRu(origin)
	thresholdStr := issues.FormatThreshold(threshold)

	// Russian comparison operators
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
		// Get the numeric value for plural form
		thresholdInt := 0
		switch v := threshold.(type) {
		case int:
			thresholdInt = v
		case int64:
			thresholdInt = int(v)
		case float64:
			thresholdInt = int(v)
		}
		unit := getRussianPlural(thresholdInt, sizing.UnitOne, sizing.UnitFew, sizing.UnitMany)

		if isTooSmall {
			return fmt.Sprintf("Слишком маленькое значение: ожидалось, что %s будет иметь %s%s %s", origin, adj, thresholdStr, unit)
		}
		return fmt.Sprintf("Слишком большое значение: ожидалось, что %s будет иметь %s%s %s", origin, adj, thresholdStr, unit)
	}

	if isTooSmall {
		return fmt.Sprintf("Слишком маленькое значение: ожидалось, что %s будет %s%s", origin, adj, thresholdStr)
	}
	return fmt.Sprintf("Слишком большое значение: ожидалось, что %s будет %s%s", origin, adj, thresholdStr)
}

// formatStringValidationRu handles string format validation messages in Russian
func formatStringValidationRu(raw core.ZodRawIssue, format string) string {
	switch format {
	case "starts_with":
		prefix := mapx.StringOr(raw.Properties, "prefix", "")
		if prefix == "" {
			return "Неверная строка: должна начинаться с указанного префикса"
		}
		return fmt.Sprintf("Неверная строка: должна начинаться с \"%s\"", prefix)
	case "ends_with":
		suffix := mapx.StringOr(raw.Properties, "suffix", "")
		if suffix == "" {
			return "Неверная строка: должна заканчиваться указанным суффиксом"
		}
		return fmt.Sprintf("Неверная строка: должна заканчиваться на \"%s\"", suffix)
	case "includes":
		includes := mapx.StringOr(raw.Properties, "includes", "")
		if includes == "" {
			return "Неверная строка: должна содержать указанную подстроку"
		}
		return fmt.Sprintf("Неверная строка: должна содержать \"%s\"", includes)
	case "regex":
		pattern := mapx.StringOr(raw.Properties, "pattern", "")
		if pattern == "" {
			return "Неверная строка: должна соответствовать шаблону"
		}
		return fmt.Sprintf("Неверная строка: должна соответствовать шаблону %s", pattern)
	default:
		noun := getFormatNounRu(format)
		return fmt.Sprintf("Неверный %s", noun)
	}
}

// Ru returns a ZodConfig configured for Russian locale.
func Ru() *core.ZodConfig {
	return &core.ZodConfig{
		LocaleError: formatRu,
	}
}

// FormatMessageRu formats a single issue using Russian locale
func FormatMessageRu(issue core.ZodRawIssue) string {
	return formatRu(issue)
}
