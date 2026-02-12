//nolint:misspell // Norwegian locale contains non-English words
package locales

import (
	"fmt"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// =============================================================================
// NORWEGIAN LOCALE FORMATTER
// =============================================================================

// Norwegian sizing info mappings
var SizableNo = map[string]issues.SizingInfo{
	"string": {Unit: "tegn", Verb: "å ha"},
	"file":   {Unit: "bytes", Verb: "å ha"},
	"array":  {Unit: "elementer", Verb: "å inneholde"},
	"slice":  {Unit: "elementer", Verb: "å inneholde"},
	"set":    {Unit: "elementer", Verb: "å inneholde"},
	"map":    {Unit: "oppføringer", Verb: "å inneholde"},
}

// Norwegian format noun mappings
var FormatNounsNo = map[string]string{
	"regex":            "input",
	"email":            "e-postadresse",
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
	"datetime":         "ISO dato- og klokkeslett",
	"date":             "ISO-dato",
	"time":             "ISO-klokkeslett",
	"duration":         "ISO-varighet",
	"ipv4":             "IPv4-adresse",
	"ipv6":             "IPv6-adresse",
	"mac":              "MAC-adresse",
	"cidrv4":           "IPv4-spekter",
	"cidrv6":           "IPv6-spekter",
	"base64":           "base64-enkodet streng",
	"base64url":        "base64url-enkodet streng",
	"json_string":      "JSON-streng",
	"e164":             "E.164-nummer",
	"jwt":              "JWT",
	"template_literal": "input",
}

// Norwegian type dictionary
var TypeDictionaryNo = map[string]string{
	"nan":       "NaN",
	"number":    "tall",
	"array":     "liste",
	"slice":     "liste",
	"string":    "streng",
	"bool":      "boolsk verdi",
	"object":    "objekt",
	"map":       "kart",
	"nil":       "null",
	"undefined": "udefinert",
	"function":  "funksjon",
	"date":      "dato",
	"file":      "fil",
	"set":       "sett",
}

// getSizingNo returns Norwegian sizing information for a given type
func getSizingNo(origin string) *issues.SizingInfo {
	if info, exists := SizableNo[origin]; exists {
		return &info
	}
	return nil
}

// getFormatNounNo returns Norwegian noun for a format name
func getFormatNounNo(format string) string {
	if noun, exists := FormatNounsNo[format]; exists {
		return noun
	}
	return format
}

// getTypeNameNo returns Norwegian type name
func getTypeNameNo(typeName string) string {
	if name, exists := TypeDictionaryNo[typeName]; exists {
		return name
	}
	return typeName
}

// formatNo provides Norwegian error messages
func formatNo(raw core.ZodRawIssue) string {
	code := raw.Code

	switch code {
	case core.InvalidType:
		expected := mapx.StringOr(raw.Properties, "expected", "")
		expected = getTypeNameNo(expected)
		received := issues.ParsedTypeToString(raw.Input)
		received = getTypeNameNo(received)
		return fmt.Sprintf("Ugyldig input: forventet %s, fikk %s", expected, received)

	case core.InvalidValue:
		values := mapx.AnySliceOr(raw.Properties, "values", nil)
		if len(values) == 0 {
			return "Ugyldig verdi"
		}
		if len(values) == 1 {
			return fmt.Sprintf("Ugyldig verdi: forventet %s", issues.StringifyPrimitive(values[0]))
		}
		return fmt.Sprintf("Ugyldig valg: forventet en av %s",
			issues.JoinValuesWithSeparator(values, "|"))

	case core.TooBig:
		return formatSizeConstraintNo(raw, false)

	case core.TooSmall:
		return formatSizeConstraintNo(raw, true)

	case core.InvalidFormat:
		format := mapx.StringOr(raw.Properties, "format", "")
		if format == "" {
			return "Ugyldig format"
		}
		return formatStringValidationNo(raw, format)

	case core.NotMultipleOf:
		divisor := mapx.AnyOr(raw.Properties, "divisor", nil)
		if divisor == nil {
			return "Ugyldig tall: må være et multiplum"
		}
		return fmt.Sprintf("Ugyldig tall: må være et multiplum av %v", divisor)

	case core.UnrecognizedKeys:
		keys := mapx.StringsOr(raw.Properties, "keys", nil)
		if len(keys) == 0 {
			return "Ukjent nøkkel"
		}
		keyWord := "Ukjent nøkkel"
		if len(keys) > 1 {
			keyWord = "Ukjente nøkler"
		}
		if keysAny, err := slicex.ToAny(keys); err == nil && !slicex.IsEmpty(keysAny) {
			keysJoined := issues.JoinValuesWithSeparator(keysAny, ", ")
			return fmt.Sprintf("%s: %s", keyWord, keysJoined)
		}
		return keyWord

	case core.InvalidKey:
		origin := mapx.StringOr(raw.Properties, "origin", "")
		if origin == "" {
			return "Ugyldig nøkkel"
		}
		return fmt.Sprintf("Ugyldig nøkkel i %s", origin)

	case core.InvalidUnion:
		return "Ugyldig input"

	case core.InvalidElement:
		origin := mapx.StringOr(raw.Properties, "origin", "")
		if origin == "" {
			return "Ugyldig element"
		}
		return fmt.Sprintf("Ugyldig verdi i %s", origin)

	case core.MissingRequired:
		fieldName := mapx.StringOr(raw.Properties, "field_name", "")
		fieldType := mapx.StringOr(raw.Properties, "field_type", "felt")
		if fieldName == "" {
			return fmt.Sprintf("Påkrevd %s mangler", fieldType)
		}
		return fmt.Sprintf("Påkrevd %s mangler: %s", fieldType, fieldName)

	case core.TypeConversion:
		fromType := mapx.StringOr(raw.Properties, "from_type", "ukjent")
		toType := mapx.StringOr(raw.Properties, "to_type", "ukjent")
		return fmt.Sprintf("Typekonvertering mislyktes: kan ikke konvertere %s til %s", fromType, toType)

	case core.InvalidSchema:
		reason := mapx.StringOr(raw.Properties, "reason", "")
		if reason != "" {
			return fmt.Sprintf("Ugyldig skjema: %s", reason)
		}
		return "Ugyldig skjemadefinisjon"

	case core.InvalidDiscriminator:
		field := mapx.StringOr(raw.Properties, "field", "diskriminator")
		return fmt.Sprintf("Ugyldig eller manglende diskriminatorfelt: %s", field)

	case core.IncompatibleTypes:
		conflictType := mapx.StringOr(raw.Properties, "conflict_type", "verdier")
		return fmt.Sprintf("Kan ikke slå sammen %s: inkompatible typer", conflictType)

	case core.NilPointer:
		return "Null-peker oppdaget"

	case core.Custom:
		message := mapx.StringOr(raw.Properties, "message", "")
		if message != "" {
			return message
		}
		return "Ugyldig input"

	default:
		return "Ugyldig input"
	}
}

// formatSizeConstraintNo formats size constraint messages in Norwegian
func formatSizeConstraintNo(raw core.ZodRawIssue, isTooSmall bool) string {
	origin := mapx.StringOr(raw.Properties, "origin", "")
	if origin == "" {
		origin = "value"
	}

	var threshold any
	if isTooSmall {
		threshold = mapx.AnyOr(raw.Properties, "minimum", nil)
	} else {
		threshold = mapx.AnyOr(raw.Properties, "maximum", nil)
	}

	if threshold == nil {
		if isTooSmall {
			return "For lite(n)"
		}
		return "For stor(t)"
	}

	inclusive := mapx.BoolOr(raw.Properties, "inclusive", true)
	sizing := getSizingNo(mapx.StringOr(raw.Properties, "origin", ""))
	thresholdStr := issues.FormatThreshold(threshold)

	// Norwegian comparison operators
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
			return fmt.Sprintf("For lite(n): forventet %s til å ha %s%s %s", origin, adj, thresholdStr, sizing.Unit)
		}
		return fmt.Sprintf("For stor(t): forventet %s til å ha %s%s %s", origin, adj, thresholdStr, sizing.Unit)
	}

	if isTooSmall {
		return fmt.Sprintf("For lite(n): forventet %s til å ha %s%s", origin, adj, thresholdStr)
	}
	return fmt.Sprintf("For stor(t): forventet %s til å ha %s%s", origin, adj, thresholdStr)
}

// formatStringValidationNo handles string format validation messages in Norwegian
func formatStringValidationNo(raw core.ZodRawIssue, format string) string {
	switch format {
	case "starts_with":
		prefix := mapx.StringOr(raw.Properties, "prefix", "")
		if prefix == "" {
			return "Ugyldig streng: må starte med angitt prefiks"
		}
		return fmt.Sprintf("Ugyldig streng: må starte med \"%s\"", prefix)
	case "ends_with":
		suffix := mapx.StringOr(raw.Properties, "suffix", "")
		if suffix == "" {
			return "Ugyldig streng: må ende med angitt suffiks"
		}
		return fmt.Sprintf("Ugyldig streng: må ende med \"%s\"", suffix)
	case "includes":
		includes := mapx.StringOr(raw.Properties, "includes", "")
		if includes == "" {
			return "Ugyldig streng: må inneholde angitt delstreng"
		}
		return fmt.Sprintf("Ugyldig streng: må inneholde \"%s\"", includes)
	case "regex":
		pattern := mapx.StringOr(raw.Properties, "pattern", "")
		if pattern == "" {
			return "Ugyldig streng: må matche mønsteret"
		}
		return fmt.Sprintf("Ugyldig streng: må matche mønsteret %s", pattern)
	default:
		noun := getFormatNounNo(format)
		return fmt.Sprintf("Ugyldig %s", noun)
	}
}

// No returns a ZodConfig configured for Norwegian locale.
func No() *core.ZodConfig {
	return &core.ZodConfig{
		LocaleError: formatNo,
	}
}

// FormatMessageNo formats a single issue using Norwegian locale
func FormatMessageNo(issue core.ZodRawIssue) string {
	return formatNo(issue)
}
