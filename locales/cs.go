package locales

import (
	"fmt"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// =============================================================================
// CZECH LOCALE FORMATTER
// =============================================================================

// Czech sizing info mappings
var SizableCs = map[string]issues.SizingInfo{
	"string": {Unit: "znaků", Verb: "mít"},
	"file":   {Unit: "bajtů", Verb: "mít"},
	"array":  {Unit: "prvků", Verb: "mít"},
	"slice":  {Unit: "prvků", Verb: "mít"},
	"set":    {Unit: "prvků", Verb: "mít"},
	"map":    {Unit: "záznamů", Verb: "mít"},
}

// Czech format noun mappings
var FormatNounsCs = map[string]string{
	"regex":            "regulární výraz",
	"email":            "e-mailová adresa",
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
	"datetime":         "datum a čas ve formátu ISO",
	"date":             "datum ve formátu ISO",
	"time":             "čas ve formátu ISO",
	"duration":         "doba trvání ISO",
	"ipv4":             "IPv4 adresa",
	"ipv6":             "IPv6 adresa",
	"mac":              "MAC adresa",
	"cidrv4":           "rozsah IPv4",
	"cidrv6":           "rozsah IPv6",
	"base64":           "řetězec zakódovaný ve formátu base64",
	"base64url":        "řetězec zakódovaný ve formátu base64url",
	"json_string":      "řetězec ve formátu JSON",
	"e164":             "číslo E.164",
	"jwt":              "JWT",
	"template_literal": "vstup",
}

// Czech type dictionary
var TypeDictionaryCs = map[string]string{
	"nan":       "NaN",
	"number":    "číslo",
	"array":     "pole",
	"slice":     "pole",
	"string":    "řetězec",
	"bool":      "boolean",
	"object":    "objekt",
	"map":       "mapa",
	"nil":       "null",
	"undefined": "nedefinováno",
	"function":  "funkce",
	"date":      "datum",
	"file":      "soubor",
	"set":       "množina",
}

// getSizingCs returns Czech sizing information for a given type
func getSizingCs(origin string) *issues.SizingInfo {
	if info, exists := SizableCs[origin]; exists {
		return &info
	}
	return nil
}

// getFormatNounCs returns Czech noun for a format name
func getFormatNounCs(format string) string {
	if noun, exists := FormatNounsCs[format]; exists {
		return noun
	}
	return format
}

// getTypeNameCs returns Czech type name
func getTypeNameCs(typeName string) string {
	if name, exists := TypeDictionaryCs[typeName]; exists {
		return name
	}
	return typeName
}

// formatCs provides Czech error messages
func formatCs(raw core.ZodRawIssue) string {
	code := raw.Code

	switch code {
	case core.InvalidType:
		expected := mapx.StringOr(raw.Properties, "expected", "")
		expected = getTypeNameCs(expected)
		received := issues.ParsedTypeToString(raw.Input)
		received = getTypeNameCs(received)
		return fmt.Sprintf("Neplatný vstup: očekáváno %s, obdrženo %s", expected, received)

	case core.InvalidValue:
		values := mapx.AnySliceOr(raw.Properties, "values", nil)
		if len(values) == 0 {
			return "Neplatná hodnota"
		}
		if len(values) == 1 {
			return fmt.Sprintf("Neplatný vstup: očekáváno %s", issues.StringifyPrimitive(values[0]))
		}
		return fmt.Sprintf("Neplatná možnost: očekávána jedna z hodnot %s",
			issues.JoinValuesWithSeparator(values, "|"))

	case core.TooBig:
		return formatSizeConstraintCs(raw, false)

	case core.TooSmall:
		return formatSizeConstraintCs(raw, true)

	case core.InvalidFormat:
		format := mapx.StringOr(raw.Properties, "format", "")
		if format == "" {
			return "Neplatný formát"
		}
		return formatStringValidationCs(raw, format)

	case core.NotMultipleOf:
		divisor := mapx.AnyOr(raw.Properties, "divisor", nil)
		if divisor == nil {
			return "Neplatné číslo: musí být násobkem"
		}
		return fmt.Sprintf("Neplatné číslo: musí být násobkem %v", divisor)

	case core.UnrecognizedKeys:
		keys := mapx.StringsOr(raw.Properties, "keys", nil)
		if len(keys) == 0 {
			return "Neznámý klíč"
		}
		keyWord := "Neznámý klíč"
		if len(keys) > 1 {
			keyWord = "Neznámé klíče"
		}
		if keysAny, err := slicex.ToAny(keys); err == nil && !slicex.IsEmpty(keysAny) {
			keysJoined := issues.JoinValuesWithSeparator(keysAny, ", ")
			return fmt.Sprintf("%s: %s", keyWord, keysJoined)
		}
		return keyWord

	case core.InvalidKey:
		origin := mapx.StringOr(raw.Properties, "origin", "")
		if origin == "" {
			return "Neplatný klíč"
		}
		return fmt.Sprintf("Neplatný klíč v %s", origin)

	case core.InvalidUnion:
		return "Neplatný vstup"

	case core.InvalidElement:
		origin := mapx.StringOr(raw.Properties, "origin", "")
		if origin == "" {
			return "Neplatný prvek"
		}
		return fmt.Sprintf("Neplatná hodnota v %s", origin)

	case core.MissingRequired:
		fieldName := mapx.StringOr(raw.Properties, "field_name", "")
		fieldType := mapx.StringOr(raw.Properties, "field_type", "pole")
		if fieldName == "" {
			return fmt.Sprintf("Chybí povinné %s", fieldType)
		}
		return fmt.Sprintf("Chybí povinné %s: %s", fieldType, fieldName)

	case core.TypeConversion:
		fromType := mapx.StringOr(raw.Properties, "from_type", "neznámý")
		toType := mapx.StringOr(raw.Properties, "to_type", "neznámý")
		return fmt.Sprintf("Převod typu selhal: nelze převést %s na %s", fromType, toType)

	case core.InvalidSchema:
		reason := mapx.StringOr(raw.Properties, "reason", "")
		if reason != "" {
			return fmt.Sprintf("Neplatné schéma: %s", reason)
		}
		return "Neplatná definice schématu"

	case core.InvalidDiscriminator:
		field := mapx.StringOr(raw.Properties, "field", "diskriminátor")
		return fmt.Sprintf("Neplatné nebo chybějící pole diskriminátoru: %s", field)

	case core.IncompatibleTypes:
		conflictType := mapx.StringOr(raw.Properties, "conflict_type", "hodnoty")
		return fmt.Sprintf("Nelze sloučit %s: nekompatibilní typy", conflictType)

	case core.NilPointer:
		return "Zjištěn nulový ukazatel"

	case core.Custom:
		message := mapx.StringOr(raw.Properties, "message", "")
		if message != "" {
			return message
		}
		return "Neplatný vstup"

	default:
		return "Neplatný vstup"
	}
}

// formatSizeConstraintCs formats size constraint messages in Czech
func formatSizeConstraintCs(raw core.ZodRawIssue, isTooSmall bool) string {
	origin := mapx.StringOr(raw.Properties, "origin", "")
	if origin == "" {
		origin = "hodnota"
	}

	var threshold any
	if isTooSmall {
		threshold = mapx.AnyOr(raw.Properties, "minimum", nil)
	} else {
		threshold = mapx.AnyOr(raw.Properties, "maximum", nil)
	}

	if threshold == nil {
		if isTooSmall {
			return "Hodnota je příliš malá"
		}
		return "Hodnota je příliš velká"
	}

	inclusive := mapx.BoolOr(raw.Properties, "inclusive", true)
	sizing := getSizingCs(mapx.StringOr(raw.Properties, "origin", ""))
	thresholdStr := issues.FormatThreshold(threshold)

	// Czech comparison operators
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
			return fmt.Sprintf("Hodnota je příliš malá: %s musí %s %s%s %s", origin, sizing.Verb, adj, thresholdStr, sizing.Unit)
		}
		return fmt.Sprintf("Hodnota je příliš velká: %s musí %s %s%s %s", origin, sizing.Verb, adj, thresholdStr, sizing.Unit)
	}

	if isTooSmall {
		return fmt.Sprintf("Hodnota je příliš malá: %s musí být %s%s", origin, adj, thresholdStr)
	}
	return fmt.Sprintf("Hodnota je příliš velká: %s musí být %s%s", origin, adj, thresholdStr)
}

// formatStringValidationCs handles string format validation messages in Czech
func formatStringValidationCs(raw core.ZodRawIssue, format string) string {
	switch format {
	case "starts_with":
		prefix := mapx.StringOr(raw.Properties, "prefix", "")
		if prefix == "" {
			return "Neplatný řetězec: musí začínat zadaným prefixem"
		}
		return fmt.Sprintf("Neplatný řetězec: musí začínat na \"%s\"", prefix)
	case "ends_with":
		suffix := mapx.StringOr(raw.Properties, "suffix", "")
		if suffix == "" {
			return "Neplatný řetězec: musí končit zadaným sufixem"
		}
		return fmt.Sprintf("Neplatný řetězec: musí končit na \"%s\"", suffix)
	case "includes":
		includes := mapx.StringOr(raw.Properties, "includes", "")
		if includes == "" {
			return "Neplatný řetězec: musí obsahovat zadaný podřetězec"
		}
		return fmt.Sprintf("Neplatný řetězec: musí obsahovat \"%s\"", includes)
	case "regex":
		pattern := mapx.StringOr(raw.Properties, "pattern", "")
		if pattern == "" {
			return "Neplatný řetězec: musí odpovídat vzoru"
		}
		return fmt.Sprintf("Neplatný řetězec: musí odpovídat vzoru %s", pattern)
	default:
		noun := getFormatNounCs(format)
		return fmt.Sprintf("Neplatný formát %s", noun)
	}
}

// Cs returns a ZodConfig configured for Czech locale.
func Cs() *core.ZodConfig {
	return &core.ZodConfig{
		LocaleError: formatCs,
	}
}

// FormatMessageCs formats a single issue using Czech locale
func FormatMessageCs(issue core.ZodRawIssue) string {
	return formatCs(issue)
}
