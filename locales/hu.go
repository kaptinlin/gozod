package locales

import (
	"fmt"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// =============================================================================
// HUNGARIAN LOCALE FORMATTER
// =============================================================================

// Hungarian sizing info mappings
var SizableHu = map[string]issues.SizingInfo{
	"string": {Unit: "karakter", Verb: "legyen"},
	"file":   {Unit: "byte", Verb: "legyen"},
	"array":  {Unit: "elem", Verb: "legyen"},
	"slice":  {Unit: "elem", Verb: "legyen"},
	"set":    {Unit: "elem", Verb: "legyen"},
	"map":    {Unit: "bejegyzés", Verb: "legyen"},
}

// Hungarian format noun mappings
var FormatNounsHu = map[string]string{
	"regex":            "bemenet",
	"email":            "email cím",
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
	"datetime":         "ISO időbélyeg",
	"date":             "ISO dátum",
	"time":             "ISO idő",
	"duration":         "ISO időintervallum",
	"ipv4":             "IPv4 cím",
	"ipv6":             "IPv6 cím",
	"mac":              "MAC cím",
	"cidrv4":           "IPv4 tartomány",
	"cidrv6":           "IPv6 tartomány",
	"base64":           "base64-kódolt string",
	"base64url":        "base64url-kódolt string",
	"json_string":      "JSON string",
	"e164":             "E.164 szám",
	"jwt":              "JWT",
	"template_literal": "bemenet",
}

// Hungarian type dictionary
var TypeDictionaryHu = map[string]string{
	"nan":       "NaN",
	"number":    "szám",
	"array":     "tömb",
	"slice":     "tömb",
	"string":    "szöveg",
	"bool":      "logikai érték",
	"object":    "objektum",
	"map":       "térkép",
	"nil":       "null",
	"undefined": "meghatározatlan",
	"function":  "függvény",
	"date":      "dátum",
	"file":      "fájl",
	"set":       "halmaz",
}

// getSizingHu returns Hungarian sizing information for a given type
func getSizingHu(origin string) *issues.SizingInfo {
	if info, exists := SizableHu[origin]; exists {
		return new(info)
	}
	return nil
}

// getFormatNounHu returns Hungarian noun for a format name
func getFormatNounHu(format string) string {
	if noun, exists := FormatNounsHu[format]; exists {
		return noun
	}
	return format
}

// getTypeNameHu returns Hungarian type name
func getTypeNameHu(typeName string) string {
	if name, exists := TypeDictionaryHu[typeName]; exists {
		return name
	}
	return typeName
}

// formatHu provides Hungarian error messages
func formatHu(raw core.ZodRawIssue) string {
	code := raw.Code

	switch code {
	case core.InvalidType:
		expected := mapx.StringOr(raw.Properties, "expected", "")
		expected = getTypeNameHu(expected)
		received := issues.ParsedTypeToString(raw.Input)
		received = getTypeNameHu(received)
		return fmt.Sprintf("Érvénytelen bemenet: a várt érték %s, a kapott érték %s", expected, received)

	case core.InvalidValue:
		values := mapx.AnySliceOr(raw.Properties, "values", nil)
		if len(values) == 0 {
			return "Érvénytelen érték"
		}
		if len(values) == 1 {
			return fmt.Sprintf("Érvénytelen bemenet: a várt érték %s", issues.StringifyPrimitive(values[0]))
		}
		return fmt.Sprintf("Érvénytelen opció: valamelyik érték várt %s",
			issues.JoinValuesWithSeparator(values, "|"))

	case core.TooBig:
		return formatSizeConstraintHu(raw, false)

	case core.TooSmall:
		return formatSizeConstraintHu(raw, true)

	case core.InvalidFormat:
		format := mapx.StringOr(raw.Properties, "format", "")
		if format == "" {
			return "Érvénytelen formátum"
		}
		return formatStringValidationHu(raw, format)

	case core.NotMultipleOf:
		divisor := mapx.AnyOr(raw.Properties, "divisor", nil)
		if divisor == nil {
			return "Érvénytelen szám: többszörösének kell lennie"
		}
		return fmt.Sprintf("Érvénytelen szám: %v többszörösének kell lennie", divisor)

	case core.UnrecognizedKeys:
		keys := mapx.StringsOr(raw.Properties, "keys", nil)
		if len(keys) == 0 {
			return "Ismeretlen kulcs"
		}
		keyWord := "Ismeretlen kulcs"
		if len(keys) > 1 {
			keyWord = "Ismeretlen kulcsok"
		}
		if keysAny, err := slicex.ToAny(keys); err == nil && !slicex.IsEmpty(keysAny) {
			keysJoined := issues.JoinValuesWithSeparator(keysAny, ", ")
			return fmt.Sprintf("%s: %s", keyWord, keysJoined)
		}
		return keyWord

	case core.InvalidKey:
		origin := mapx.StringOr(raw.Properties, "origin", "")
		if origin == "" {
			return "Érvénytelen kulcs"
		}
		return fmt.Sprintf("Érvénytelen kulcs %s", origin)

	case core.InvalidUnion:
		return "Érvénytelen bemenet"

	case core.InvalidElement:
		origin := mapx.StringOr(raw.Properties, "origin", "")
		if origin == "" {
			return "Érvénytelen elem"
		}
		return fmt.Sprintf("Érvénytelen érték: %s", origin)

	case core.MissingRequired:
		fieldName := mapx.StringOr(raw.Properties, "field_name", "")
		fieldType := mapx.StringOr(raw.Properties, "field_type", "mező")
		if fieldName == "" {
			return fmt.Sprintf("Kötelező %s hiányzik", fieldType)
		}
		return fmt.Sprintf("Kötelező %s hiányzik: %s", fieldType, fieldName)

	case core.TypeConversion:
		fromType := mapx.StringOr(raw.Properties, "from_type", "ismeretlen")
		toType := mapx.StringOr(raw.Properties, "to_type", "ismeretlen")
		return fmt.Sprintf("Típuskonverzió sikertelen: %s nem konvertálható %s típusra", fromType, toType)

	case core.InvalidSchema:
		reason := mapx.StringOr(raw.Properties, "reason", "")
		if reason != "" {
			return fmt.Sprintf("Érvénytelen séma: %s", reason)
		}
		return "Érvénytelen sémadefiníció"

	case core.InvalidDiscriminator:
		field := mapx.StringOr(raw.Properties, "field", "diszkriminátor")
		return fmt.Sprintf("Érvénytelen vagy hiányzó diszkriminátor mező: %s", field)

	case core.IncompatibleTypes:
		conflictType := mapx.StringOr(raw.Properties, "conflict_type", "értékek")
		return fmt.Sprintf("Nem lehet egyesíteni %s: inkompatibilis típusok", conflictType)

	case core.NilPointer:
		return "Null mutató észlelve"

	case core.Custom:
		message := mapx.StringOr(raw.Properties, "message", "")
		if message != "" {
			return message
		}
		return "Érvénytelen bemenet"

	default:
		return "Érvénytelen bemenet"
	}
}

// formatSizeConstraintHu formats size constraint messages in Hungarian
func formatSizeConstraintHu(raw core.ZodRawIssue, isTooSmall bool) string {
	origin := mapx.StringOr(raw.Properties, "origin", "")
	if origin == "" {
		origin = "érték"
	}

	var threshold any
	if isTooSmall {
		threshold = mapx.AnyOr(raw.Properties, "minimum", nil)
	} else {
		threshold = mapx.AnyOr(raw.Properties, "maximum", nil)
	}

	if threshold == nil {
		if isTooSmall {
			return "Túl kicsi"
		}
		return "Túl nagy"
	}

	inclusive := mapx.BoolOr(raw.Properties, "inclusive", true)
	sizing := getSizingHu(mapx.StringOr(raw.Properties, "origin", ""))
	thresholdStr := issues.FormatThreshold(threshold)

	// Hungarian comparison operators
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
			return fmt.Sprintf("Túl kicsi: a bemeneti érték %s mérete túl kicsi %s%s %s", origin, adj, thresholdStr, sizing.Unit)
		}
		return fmt.Sprintf("Túl nagy: %s mérete túl nagy %s%s %s", origin, adj, thresholdStr, sizing.Unit)
	}

	if isTooSmall {
		return fmt.Sprintf("Túl kicsi: a bemeneti érték %s túl kicsi %s%s", origin, adj, thresholdStr)
	}
	return fmt.Sprintf("Túl nagy: a bemeneti érték %s túl nagy: %s%s", origin, adj, thresholdStr)
}

// formatStringValidationHu handles string format validation messages in Hungarian
func formatStringValidationHu(raw core.ZodRawIssue, format string) string {
	switch format {
	case "starts_with":
		prefix := mapx.StringOr(raw.Properties, "prefix", "")
		if prefix == "" {
			return "Érvénytelen string: megadott értékkel kell kezdődnie"
		}
		return fmt.Sprintf("Érvénytelen string: \"%s\" értékkel kell kezdődnie", prefix)
	case "ends_with":
		suffix := mapx.StringOr(raw.Properties, "suffix", "")
		if suffix == "" {
			return "Érvénytelen string: megadott értékkel kell végződnie"
		}
		return fmt.Sprintf("Érvénytelen string: \"%s\" értékkel kell végződnie", suffix)
	case "includes":
		includes := mapx.StringOr(raw.Properties, "includes", "")
		if includes == "" {
			return "Érvénytelen string: megadott értéket kell tartalmaznia"
		}
		return fmt.Sprintf("Érvénytelen string: \"%s\" értéket kell tartalmaznia", includes)
	case "regex":
		pattern := mapx.StringOr(raw.Properties, "pattern", "")
		if pattern == "" {
			return "Érvénytelen string: mintának kell megfelelnie"
		}
		return fmt.Sprintf("Érvénytelen string: %s mintának kell megfelelnie", pattern)
	default:
		noun := getFormatNounHu(format)
		return fmt.Sprintf("Érvénytelen %s", noun)
	}
}

// Hu returns a ZodConfig configured for Hungarian locale.
func Hu() *core.ZodConfig {
	return &core.ZodConfig{
		LocaleError: formatHu,
	}
}

// FormatMessageHu formats a single issue using Hungarian locale
func FormatMessageHu(issue core.ZodRawIssue) string {
	return formatHu(issue)
}
