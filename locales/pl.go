//nolint:misspell // Polish locale contains non-English words
package locales

import (
	"fmt"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// =============================================================================
// POLISH LOCALE FORMATTER
// =============================================================================

// Polish sizing info mappings
var SizablePl = map[string]issues.SizingInfo{
	"string": {Unit: "znaków", Verb: "mieć"},
	"file":   {Unit: "bajtów", Verb: "mieć"},
	"array":  {Unit: "elementów", Verb: "mieć"},
	"slice":  {Unit: "elementów", Verb: "mieć"},
	"set":    {Unit: "elementów", Verb: "mieć"},
	"map":    {Unit: "wpisów", Verb: "mieć"},
}

// Polish format noun mappings
var FormatNounsPl = map[string]string{
	"regex":            "wyrażenie",
	"email":            "adres email",
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
	"datetime":         "data i godzina w formacie ISO",
	"date":             "data w formacie ISO",
	"time":             "godzina w formacie ISO",
	"duration":         "czas trwania ISO",
	"ipv4":             "adres IPv4",
	"ipv6":             "adres IPv6",
	"mac":              "adres MAC",
	"cidrv4":           "zakres IPv4",
	"cidrv6":           "zakres IPv6",
	"base64":           "ciąg znaków zakodowany w formacie base64",
	"base64url":        "ciąg znaków zakodowany w formacie base64url",
	"json_string":      "ciąg znaków w formacie JSON",
	"e164":             "liczba E.164",
	"jwt":              "JWT",
	"template_literal": "wejście",
}

// Polish type dictionary
var TypeDictionaryPl = map[string]string{
	"nan":       "NaN",
	"number":    "liczba",
	"array":     "tablica",
	"slice":     "tablica",
	"string":    "ciąg znaków",
	"bool":      "wartość logiczna",
	"object":    "obiekt",
	"map":       "mapa",
	"nil":       "null",
	"undefined": "undefined",
	"function":  "funkcja",
	"date":      "data",
	"file":      "plik",
	"set":       "zbiór",
}

// getSizingPl returns Polish sizing information for a given type
func getSizingPl(origin string) *issues.SizingInfo {
	if info, exists := SizablePl[origin]; exists {
		return new(info)
	}
	return nil
}

// getFormatNounPl returns Polish noun for a format name
func getFormatNounPl(format string) string {
	if noun, exists := FormatNounsPl[format]; exists {
		return noun
	}
	return format
}

// getTypeNamePl returns Polish type name
func getTypeNamePl(typeName string) string {
	if name, exists := TypeDictionaryPl[typeName]; exists {
		return name
	}
	return typeName
}

// formatPl provides Polish error messages
func formatPl(raw core.ZodRawIssue) string {
	code := raw.Code

	switch code {
	case core.InvalidType:
		expected := mapx.StringOr(raw.Properties, "expected", "")
		expected = getTypeNamePl(expected)
		received := issues.ParsedTypeToString(raw.Input)
		received = getTypeNamePl(received)
		return fmt.Sprintf("Nieprawidłowe dane wejściowe: oczekiwano %s, otrzymano %s", expected, received)

	case core.InvalidValue:
		values := mapx.AnySliceOr(raw.Properties, "values", nil)
		if len(values) == 0 {
			return "Nieprawidłowa wartość"
		}
		if len(values) == 1 {
			return fmt.Sprintf("Nieprawidłowe dane wejściowe: oczekiwano %s", issues.StringifyPrimitive(values[0]))
		}
		return fmt.Sprintf("Nieprawidłowa opcja: oczekiwano jednej z wartości %s",
			issues.JoinValuesWithSeparator(values, "|"))

	case core.TooBig:
		return formatSizeConstraintPl(raw, false)

	case core.TooSmall:
		return formatSizeConstraintPl(raw, true)

	case core.InvalidFormat:
		format := mapx.StringOr(raw.Properties, "format", "")
		if format == "" {
			return "Nieprawidłowy format"
		}
		return formatStringValidationPl(raw, format)

	case core.NotMultipleOf:
		divisor := mapx.AnyOr(raw.Properties, "divisor", nil)
		if divisor == nil {
			return "Nieprawidłowa liczba: musi być wielokrotnością"
		}
		return fmt.Sprintf("Nieprawidłowa liczba: musi być wielokrotnością %v", divisor)

	case core.UnrecognizedKeys:
		keys := mapx.StringsOr(raw.Properties, "keys", nil)
		if len(keys) == 0 {
			return "Nierozpoznany klucz"
		}
		keyWord := "Nierozpoznany klucz"
		if len(keys) > 1 {
			keyWord = "Nierozpoznane klucze"
		}
		if keysAny, err := slicex.ToAny(keys); err == nil && !slicex.IsEmpty(keysAny) {
			keysJoined := issues.JoinValuesWithSeparator(keysAny, ", ")
			return fmt.Sprintf("%s: %s", keyWord, keysJoined)
		}
		return keyWord

	case core.InvalidKey:
		origin := mapx.StringOr(raw.Properties, "origin", "")
		if origin == "" {
			return "Nieprawidłowy klucz"
		}
		return fmt.Sprintf("Nieprawidłowy klucz w %s", origin)

	case core.InvalidUnion:
		return "Nieprawidłowe dane wejściowe"

	case core.InvalidElement:
		origin := mapx.StringOr(raw.Properties, "origin", "")
		if origin == "" {
			return "Nieprawidłowy element"
		}
		return fmt.Sprintf("Nieprawidłowa wartość w %s", origin)

	case core.MissingRequired:
		fieldName := mapx.StringOr(raw.Properties, "field_name", "")
		fieldType := mapx.StringOr(raw.Properties, "field_type", "pole")
		if fieldName == "" {
			return fmt.Sprintf("Brakuje wymaganego %s", fieldType)
		}
		return fmt.Sprintf("Brakuje wymaganego %s: %s", fieldType, fieldName)

	case core.TypeConversion:
		fromType := mapx.StringOr(raw.Properties, "from_type", "nieznany")
		toType := mapx.StringOr(raw.Properties, "to_type", "nieznany")
		return fmt.Sprintf("Konwersja typu nie powiodła się: nie można przekonwertować %s na %s", fromType, toType)

	case core.InvalidSchema:
		reason := mapx.StringOr(raw.Properties, "reason", "")
		if reason != "" {
			return fmt.Sprintf("Nieprawidłowy schemat: %s", reason)
		}
		return "Nieprawidłowa definicja schematu"

	case core.InvalidDiscriminator:
		field := mapx.StringOr(raw.Properties, "field", "dyskryminator")
		return fmt.Sprintf("Nieprawidłowe lub brakujące pole dyskryminatora: %s", field)

	case core.IncompatibleTypes:
		conflictType := mapx.StringOr(raw.Properties, "conflict_type", "wartości")
		return fmt.Sprintf("Nie można scalić %s: niezgodne typy", conflictType)

	case core.NilPointer:
		return "Napotkano pusty wskaźnik"

	case core.Custom:
		message := mapx.StringOr(raw.Properties, "message", "")
		if message != "" {
			return message
		}
		return "Nieprawidłowe dane wejściowe"

	default:
		return "Nieprawidłowe dane wejściowe"
	}
}

// formatSizeConstraintPl formats size constraint messages in Polish
func formatSizeConstraintPl(raw core.ZodRawIssue, isTooSmall bool) string {
	origin := mapx.StringOr(raw.Properties, "origin", "")
	if origin == "" {
		origin = "wartość"
	}

	var threshold any
	if isTooSmall {
		threshold = mapx.AnyOr(raw.Properties, "minimum", nil)
	} else {
		threshold = mapx.AnyOr(raw.Properties, "maximum", nil)
	}

	if threshold == nil {
		if isTooSmall {
			return "Za mała wartość"
		}
		return "Za duża wartość"
	}

	inclusive := mapx.BoolOr(raw.Properties, "inclusive", true)
	sizing := getSizingPl(mapx.StringOr(raw.Properties, "origin", ""))
	thresholdStr := issues.FormatThreshold(threshold)

	// Polish comparison operators
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
			return fmt.Sprintf("Za mała wartość: oczekiwano, że %s będzie mieć %s%s %s", origin, adj, thresholdStr, sizing.Unit)
		}
		return fmt.Sprintf("Za duża wartość: oczekiwano, że %s będzie mieć %s%s %s", origin, adj, thresholdStr, sizing.Unit)
	}

	if isTooSmall {
		return fmt.Sprintf("Za mała wartość: oczekiwano, że %s będzie wynosić %s%s", origin, adj, thresholdStr)
	}
	return fmt.Sprintf("Za duża wartość: oczekiwano, że %s będzie wynosić %s%s", origin, adj, thresholdStr)
}

// formatStringValidationPl handles string format validation messages in Polish
func formatStringValidationPl(raw core.ZodRawIssue, format string) string {
	switch format {
	case "starts_with":
		prefix := mapx.StringOr(raw.Properties, "prefix", "")
		if prefix == "" {
			return "Nieprawidłowy ciąg znaków: musi zaczynać się od określonego prefiksu"
		}
		return fmt.Sprintf("Nieprawidłowy ciąg znaków: musi zaczynać się od \"%s\"", prefix)
	case "ends_with":
		suffix := mapx.StringOr(raw.Properties, "suffix", "")
		if suffix == "" {
			return "Nieprawidłowy ciąg znaków: musi kończyć się określonym sufiksem"
		}
		return fmt.Sprintf("Nieprawidłowy ciąg znaków: musi kończyć się na \"%s\"", suffix)
	case "includes":
		includes := mapx.StringOr(raw.Properties, "includes", "")
		if includes == "" {
			return "Nieprawidłowy ciąg znaków: musi zawierać określony podciąg"
		}
		return fmt.Sprintf("Nieprawidłowy ciąg znaków: musi zawierać \"%s\"", includes)
	case "regex":
		pattern := mapx.StringOr(raw.Properties, "pattern", "")
		if pattern == "" {
			return "Nieprawidłowy ciąg znaków: musi odpowiadać wzorcowi"
		}
		return fmt.Sprintf("Nieprawidłowy ciąg znaków: musi odpowiadać wzorcowi %s", pattern)
	default:
		noun := getFormatNounPl(format)
		return fmt.Sprintf("Nieprawidłowy %s", noun)
	}
}

// Pl returns a ZodConfig configured for Polish locale.
func Pl() *core.ZodConfig {
	return &core.ZodConfig{
		LocaleError: formatPl,
	}
}

// FormatMessagePl formats a single issue using Polish locale
func FormatMessagePl(issue core.ZodRawIssue) string {
	return formatPl(issue)
}
