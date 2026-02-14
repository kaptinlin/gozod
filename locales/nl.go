//nolint:misspell // Dutch locale contains non-English words
package locales

import (
	"fmt"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// =============================================================================
// DUTCH LOCALE FORMATTER
// =============================================================================

// Dutch sizing info mappings
var SizableNl = map[string]issues.SizingInfo{
	"string": {Unit: "tekens", Verb: "heeft"},
	"file":   {Unit: "bytes", Verb: "heeft"},
	"array":  {Unit: "elementen", Verb: "heeft"},
	"slice":  {Unit: "elementen", Verb: "heeft"},
	"set":    {Unit: "elementen", Verb: "heeft"},
	"map":    {Unit: "items", Verb: "heeft"},
}

// Dutch format noun mappings
var FormatNounsNl = map[string]string{
	"regex":            "invoer",
	"email":            "emailadres",
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
	"datetime":         "ISO datum en tijd",
	"date":             "ISO datum",
	"time":             "ISO tijd",
	"duration":         "ISO duur",
	"ipv4":             "IPv4-adres",
	"ipv6":             "IPv6-adres",
	"mac":              "MAC-adres",
	"cidrv4":           "IPv4-bereik",
	"cidrv6":           "IPv6-bereik",
	"base64":           "base64-gecodeerde tekst",
	"base64url":        "base64 URL-gecodeerde tekst",
	"json_string":      "JSON string",
	"e164":             "E.164-nummer",
	"jwt":              "JWT",
	"template_literal": "invoer",
}

// Dutch type dictionary
var TypeDictionaryNl = map[string]string{
	"nan":       "NaN",
	"number":    "getal",
	"array":     "array",
	"slice":     "array",
	"string":    "tekst",
	"bool":      "boolean",
	"object":    "object",
	"map":       "map",
	"nil":       "null",
	"undefined": "ongedefinieerd",
	"function":  "functie",
	"date":      "datum",
	"file":      "bestand",
	"set":       "set",
}

// getSizingNl returns Dutch sizing information for a given type
func getSizingNl(origin string) *issues.SizingInfo {
	if _, exists := SizableNl[origin]; exists {
		return new(SizableNl[origin])
	}
	return nil
}

// getFormatNounNl returns Dutch noun for a format name
func getFormatNounNl(format string) string {
	if noun, exists := FormatNounsNl[format]; exists {
		return noun
	}
	return format
}

// getTypeNameNl returns Dutch type name
func getTypeNameNl(typeName string) string {
	if name, exists := TypeDictionaryNl[typeName]; exists {
		return name
	}
	return typeName
}

// getDutchSizeAdjective returns the appropriate Dutch adjective for size constraints
func getDutchSizeAdjective(origin string, isTooSmall bool) string {
	if isTooSmall {
		switch origin {
		case "date":
			return "vroeg"
		case "string":
			return "kort"
		default:
			return "klein"
		}
	}
	switch origin {
	case "date":
		return "laat"
	case "string":
		return "lang"
	default:
		return "groot"
	}
}

// formatNl provides Dutch error messages
func formatNl(raw core.ZodRawIssue) string {
	code := raw.Code

	switch code {
	case core.InvalidType:
		expected := mapx.StringOr(raw.Properties, "expected", "")
		expected = getTypeNameNl(expected)
		received := issues.ParsedTypeToString(raw.Input)
		received = getTypeNameNl(received)
		return fmt.Sprintf("Ongeldige invoer: verwacht %s, ontving %s", expected, received)

	case core.InvalidValue:
		values := mapx.AnySliceOr(raw.Properties, "values", nil)
		if len(values) == 0 {
			return "Ongeldige waarde"
		}
		if len(values) == 1 {
			return fmt.Sprintf("Ongeldige invoer: verwacht %s", issues.StringifyPrimitive(values[0]))
		}
		return fmt.Sprintf("Ongeldige optie: verwacht één van %s",
			issues.JoinValuesWithSeparator(values, "|"))

	case core.TooBig:
		return formatSizeConstraintNl(raw, false)

	case core.TooSmall:
		return formatSizeConstraintNl(raw, true)

	case core.InvalidFormat:
		format := mapx.StringOr(raw.Properties, "format", "")
		if format == "" {
			return "Ongeldig formaat"
		}
		return formatStringValidationNl(raw, format)

	case core.NotMultipleOf:
		divisor := mapx.AnyOr(raw.Properties, "divisor", nil)
		if divisor == nil {
			return "Ongeldig getal: moet een veelvoud zijn"
		}
		return fmt.Sprintf("Ongeldig getal: moet een veelvoud van %v zijn", divisor)

	case core.UnrecognizedKeys:
		keys := mapx.StringsOr(raw.Properties, "keys", nil)
		if len(keys) == 0 {
			return "Onbekende key"
		}
		keyWord := "Onbekende key"
		if len(keys) > 1 {
			keyWord = "Onbekende keys"
		}
		if keysAny, err := slicex.ToAny(keys); err == nil && !slicex.IsEmpty(keysAny) {
			keysJoined := issues.JoinValuesWithSeparator(keysAny, ", ")
			return fmt.Sprintf("%s: %s", keyWord, keysJoined)
		}
		return keyWord

	case core.InvalidKey:
		origin := mapx.StringOr(raw.Properties, "origin", "")
		if origin == "" {
			return "Ongeldige key"
		}
		return fmt.Sprintf("Ongeldige key in %s", origin)

	case core.InvalidUnion:
		return "Ongeldige invoer"

	case core.InvalidElement:
		origin := mapx.StringOr(raw.Properties, "origin", "")
		if origin == "" {
			return "Ongeldig element"
		}
		return fmt.Sprintf("Ongeldige waarde in %s", origin)

	case core.MissingRequired:
		fieldName := mapx.StringOr(raw.Properties, "field_name", "")
		fieldType := mapx.StringOr(raw.Properties, "field_type", "veld")
		if fieldName == "" {
			return fmt.Sprintf("Verplicht %s ontbreekt", fieldType)
		}
		return fmt.Sprintf("Verplicht %s ontbreekt: %s", fieldType, fieldName)

	case core.TypeConversion:
		fromType := mapx.StringOr(raw.Properties, "from_type", "onbekend")
		toType := mapx.StringOr(raw.Properties, "to_type", "onbekend")
		return fmt.Sprintf("Typeconversie mislukt: kan %s niet naar %s converteren", fromType, toType)

	case core.InvalidSchema:
		reason := mapx.StringOr(raw.Properties, "reason", "")
		if reason != "" {
			return fmt.Sprintf("Ongeldig schema: %s", reason)
		}
		return "Ongeldige schemadefinitie"

	case core.InvalidDiscriminator:
		field := mapx.StringOr(raw.Properties, "field", "discriminator")
		return fmt.Sprintf("Ongeldig of ontbrekend discriminatorveld: %s", field)

	case core.IncompatibleTypes:
		conflictType := mapx.StringOr(raw.Properties, "conflict_type", "waarden")
		return fmt.Sprintf("Kan %s niet samenvoegen: incompatibele types", conflictType)

	case core.NilPointer:
		return "Null pointer aangetroffen"

	case core.Custom:
		message := mapx.StringOr(raw.Properties, "message", "")
		if message != "" {
			return message
		}
		return "Ongeldige invoer"

	default:
		return "Ongeldige invoer"
	}
}

// formatSizeConstraintNl formats size constraint messages in Dutch
func formatSizeConstraintNl(raw core.ZodRawIssue, isTooSmall bool) string {
	origin := mapx.StringOr(raw.Properties, "origin", "")
	if origin == "" {
		origin = "waarde"
	}

	var threshold any
	if isTooSmall {
		threshold = mapx.AnyOr(raw.Properties, "minimum", nil)
	} else {
		threshold = mapx.AnyOr(raw.Properties, "maximum", nil)
	}

	if threshold == nil {
		sizeAdj := getDutchSizeAdjective(origin, isTooSmall)
		return fmt.Sprintf("Te %s", sizeAdj)
	}

	inclusive := mapx.BoolOr(raw.Properties, "inclusive", true)
	sizing := getSizingNl(mapx.StringOr(raw.Properties, "origin", ""))
	thresholdStr := issues.FormatThreshold(threshold)
	sizeAdj := getDutchSizeAdjective(origin, isTooSmall)

	// Dutch comparison operators
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
		return fmt.Sprintf("Te %s: verwacht dat %s %s%s %s %s", sizeAdj, origin, adj, thresholdStr, sizing.Unit, sizing.Verb)
	}

	return fmt.Sprintf("Te %s: verwacht dat %s %s%s is", sizeAdj, origin, adj, thresholdStr)
}

// formatStringValidationNl handles string format validation messages in Dutch
func formatStringValidationNl(raw core.ZodRawIssue, format string) string {
	switch format {
	case "starts_with":
		prefix := mapx.StringOr(raw.Properties, "prefix", "")
		if prefix == "" {
			return "Ongeldige tekst: moet met het opgegeven voorvoegsel beginnen"
		}
		return fmt.Sprintf("Ongeldige tekst: moet met \"%s\" beginnen", prefix)
	case "ends_with":
		suffix := mapx.StringOr(raw.Properties, "suffix", "")
		if suffix == "" {
			return "Ongeldige tekst: moet op het opgegeven achtervoegsel eindigen"
		}
		return fmt.Sprintf("Ongeldige tekst: moet op \"%s\" eindigen", suffix)
	case "includes":
		includes := mapx.StringOr(raw.Properties, "includes", "")
		if includes == "" {
			return "Ongeldige tekst: moet de opgegeven substring bevatten"
		}
		return fmt.Sprintf("Ongeldige tekst: moet \"%s\" bevatten", includes)
	case "regex":
		pattern := mapx.StringOr(raw.Properties, "pattern", "")
		if pattern == "" {
			return "Ongeldige tekst: moet overeenkomen met patroon"
		}
		return fmt.Sprintf("Ongeldige tekst: moet overeenkomen met patroon %s", pattern)
	default:
		noun := getFormatNounNl(format)
		return fmt.Sprintf("Ongeldig: %s", noun)
	}
}

// Nl returns a ZodConfig configured for Dutch locale.
func Nl() *core.ZodConfig {
	return &core.ZodConfig{
		LocaleError: formatNl,
	}
}

// FormatMessageNl formats a single issue using Dutch locale
func FormatMessageNl(issue core.ZodRawIssue) string {
	return formatNl(issue)
}
