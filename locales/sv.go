//nolint:misspell // Swedish locale contains non-English words
package locales

import (
	"fmt"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// =============================================================================
// SWEDISH LOCALE FORMATTER
// =============================================================================

// Swedish sizing info mappings
var SizableSv = map[string]issues.SizingInfo{
	"string": {Unit: "tecken", Verb: "att ha"},
	"file":   {Unit: "bytes", Verb: "att ha"},
	"array":  {Unit: "objekt", Verb: "att innehålla"},
	"slice":  {Unit: "objekt", Verb: "att innehålla"},
	"set":    {Unit: "objekt", Verb: "att innehålla"},
	"map":    {Unit: "poster", Verb: "att innehålla"},
}

// Swedish format noun mappings
var FormatNounsSv = map[string]string{
	"regex":            "reguljärt uttryck",
	"email":            "e-postadress",
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
	"datetime":         "ISO-datum och tid",
	"date":             "ISO-datum",
	"time":             "ISO-tid",
	"duration":         "ISO-varaktighet",
	"ipv4":             "IPv4-adress",
	"ipv6":             "IPv6-adress",
	"mac":              "MAC-adress",
	"cidrv4":           "IPv4-spektrum",
	"cidrv6":           "IPv6-spektrum",
	"base64":           "base64-kodad sträng",
	"base64url":        "base64url-kodad sträng",
	"json_string":      "JSON-sträng",
	"e164":             "E.164-nummer",
	"jwt":              "JWT",
	"template_literal": "mall-literal",
}

// Swedish type dictionary
var TypeDictionarySv = map[string]string{
	"nan":       "NaN",
	"number":    "antal",
	"array":     "lista",
	"slice":     "lista",
	"string":    "sträng",
	"bool":      "boolean",
	"object":    "objekt",
	"map":       "karta",
	"nil":       "null",
	"undefined": "odefinierad",
	"function":  "funktion",
	"date":      "datum",
	"file":      "fil",
	"set":       "mängd",
}

// getSizingSv returns Swedish sizing information for a given type
func getSizingSv(origin string) *issues.SizingInfo {
	if _, exists := SizableSv[origin]; exists {
		return new(SizableSv[origin])
	}
	return nil
}

// getFormatNounSv returns Swedish noun for a format name
func getFormatNounSv(format string) string {
	if noun, exists := FormatNounsSv[format]; exists {
		return noun
	}
	return format
}

// getTypeNameSv returns Swedish type name
func getTypeNameSv(typeName string) string {
	if name, exists := TypeDictionarySv[typeName]; exists {
		return name
	}
	return typeName
}

// formatSv provides Swedish error messages
func formatSv(raw core.ZodRawIssue) string {
	code := raw.Code

	switch code {
	case core.InvalidType:
		expected := mapx.StringOr(raw.Properties, "expected", "")
		expected = getTypeNameSv(expected)
		received := issues.ParsedTypeToString(raw.Input)
		received = getTypeNameSv(received)
		return fmt.Sprintf("Ogiltig inmatning: förväntat %s, fick %s", expected, received)

	case core.InvalidValue:
		values := mapx.AnySliceOr(raw.Properties, "values", nil)
		if len(values) == 0 {
			return "Ogiltigt värde"
		}
		if len(values) == 1 {
			return fmt.Sprintf("Ogiltig inmatning: förväntat %s", issues.StringifyPrimitive(values[0]))
		}
		return fmt.Sprintf("Ogiltigt val: förväntade en av %s",
			issues.JoinValuesWithSeparator(values, "|"))

	case core.TooBig:
		return formatSizeConstraintSv(raw, false)

	case core.TooSmall:
		return formatSizeConstraintSv(raw, true)

	case core.InvalidFormat:
		format := mapx.StringOr(raw.Properties, "format", "")
		if format == "" {
			return "Ogiltigt format"
		}
		return formatStringValidationSv(raw, format)

	case core.NotMultipleOf:
		divisor := mapx.AnyOr(raw.Properties, "divisor", nil)
		if divisor == nil {
			return "Ogiltigt tal: måste vara en multipel"
		}
		return fmt.Sprintf("Ogiltigt tal: måste vara en multipel av %v", divisor)

	case core.UnrecognizedKeys:
		keys := mapx.StringsOr(raw.Properties, "keys", nil)
		if len(keys) == 0 {
			return "Okänd nyckel"
		}
		keyWord := "Okänd nyckel"
		if len(keys) > 1 {
			keyWord = "Okända nycklar"
		}
		if keysAny, err := slicex.ToAny(keys); err == nil && !slicex.IsEmpty(keysAny) {
			keysJoined := issues.JoinValuesWithSeparator(keysAny, ", ")
			return fmt.Sprintf("%s: %s", keyWord, keysJoined)
		}
		return keyWord

	case core.InvalidKey:
		origin := mapx.StringOr(raw.Properties, "origin", "")
		if origin == "" {
			return "Ogiltig nyckel"
		}
		return fmt.Sprintf("Ogiltig nyckel i %s", origin)

	case core.InvalidUnion:
		return "Ogiltig input"

	case core.InvalidElement:
		origin := mapx.StringOr(raw.Properties, "origin", "")
		if origin == "" {
			return "Ogiltigt element"
		}
		return fmt.Sprintf("Ogiltigt värde i %s", origin)

	case core.MissingRequired:
		fieldName := mapx.StringOr(raw.Properties, "field_name", "")
		fieldType := mapx.StringOr(raw.Properties, "field_type", "fält")
		if fieldName == "" {
			return fmt.Sprintf("Obligatoriskt %s saknas", fieldType)
		}
		return fmt.Sprintf("Obligatoriskt %s saknas: %s", fieldType, fieldName)

	case core.TypeConversion:
		fromType := mapx.StringOr(raw.Properties, "from_type", "okänd")
		toType := mapx.StringOr(raw.Properties, "to_type", "okänd")
		return fmt.Sprintf("Typkonvertering misslyckades: kan inte konvertera %s till %s", fromType, toType)

	case core.InvalidSchema:
		reason := mapx.StringOr(raw.Properties, "reason", "")
		if reason != "" {
			return fmt.Sprintf("Ogiltigt schema: %s", reason)
		}
		return "Ogiltig schemadefinition"

	case core.InvalidDiscriminator:
		field := mapx.StringOr(raw.Properties, "field", "diskriminator")
		return fmt.Sprintf("Ogiltigt eller saknat diskriminatorfält: %s", field)

	case core.IncompatibleTypes:
		conflictType := mapx.StringOr(raw.Properties, "conflict_type", "värden")
		return fmt.Sprintf("Kan inte slå samman %s: inkompatibla typer", conflictType)

	case core.NilPointer:
		return "Null-pekare upptäckt"

	case core.Custom:
		message := mapx.StringOr(raw.Properties, "message", "")
		if message != "" {
			return message
		}
		return "Ogiltig input"

	default:
		return "Ogiltig input"
	}
}

// formatSizeConstraintSv formats size constraint messages in Swedish
func formatSizeConstraintSv(raw core.ZodRawIssue, isTooSmall bool) string {
	origin := mapx.StringOr(raw.Properties, "origin", "")
	if origin == "" {
		origin = "värdet"
	}

	var threshold any
	if isTooSmall {
		threshold = mapx.AnyOr(raw.Properties, "minimum", nil)
	} else {
		threshold = mapx.AnyOr(raw.Properties, "maximum", nil)
	}

	if threshold == nil {
		if isTooSmall {
			return "För lite(t)"
		}
		return "För stor(t)"
	}

	inclusive := mapx.BoolOr(raw.Properties, "inclusive", true)
	sizing := getSizingSv(mapx.StringOr(raw.Properties, "origin", ""))
	thresholdStr := issues.FormatThreshold(threshold)

	// Swedish comparison operators
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
			return fmt.Sprintf("För lite(t): förväntade %s %s %s%s %s", origin, sizing.Verb, adj, thresholdStr, sizing.Unit)
		}
		return fmt.Sprintf("För stor(t): förväntade %s %s %s%s %s", origin, sizing.Verb, adj, thresholdStr, sizing.Unit)
	}

	if isTooSmall {
		return fmt.Sprintf("För lite(t): förväntade %s att ha %s%s", origin, adj, thresholdStr)
	}
	return fmt.Sprintf("För stor(t): förväntade %s att ha %s%s", origin, adj, thresholdStr)
}

// formatStringValidationSv handles string format validation messages in Swedish
func formatStringValidationSv(raw core.ZodRawIssue, format string) string {
	switch format {
	case "starts_with":
		prefix := mapx.StringOr(raw.Properties, "prefix", "")
		if prefix == "" {
			return "Ogiltig sträng: måste börja med angivet prefix"
		}
		return fmt.Sprintf("Ogiltig sträng: måste börja med \"%s\"", prefix)
	case "ends_with":
		suffix := mapx.StringOr(raw.Properties, "suffix", "")
		if suffix == "" {
			return "Ogiltig sträng: måste sluta med angivet suffix"
		}
		return fmt.Sprintf("Ogiltig sträng: måste sluta med \"%s\"", suffix)
	case "includes":
		includes := mapx.StringOr(raw.Properties, "includes", "")
		if includes == "" {
			return "Ogiltig sträng: måste innehålla angiven delsträng"
		}
		return fmt.Sprintf("Ogiltig sträng: måste innehålla \"%s\"", includes)
	case "regex":
		pattern := mapx.StringOr(raw.Properties, "pattern", "")
		if pattern == "" {
			return "Ogiltig sträng: måste matcha mönstret"
		}
		return fmt.Sprintf("Ogiltig sträng: måste matcha mönstret \"%s\"", pattern)
	default:
		noun := getFormatNounSv(format)
		return fmt.Sprintf("Ogiltig(t) %s", noun)
	}
}

// Sv returns a ZodConfig configured for Swedish locale.
func Sv() *core.ZodConfig {
	return &core.ZodConfig{
		LocaleError: formatSv,
	}
}

// FormatMessageSv formats a single issue using Swedish locale
func FormatMessageSv(issue core.ZodRawIssue) string {
	return formatSv(issue)
}
