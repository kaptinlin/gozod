//nolint:misspell // Danish locale contains non-English words
package locales

import (
	"fmt"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// =============================================================================
// DANISH LOCALE FORMATTER
// =============================================================================

// Danish sizing info mappings
var SizableDa = map[string]issues.SizingInfo{
	"string": {Unit: "tegn", Verb: "havde"},
	"file":   {Unit: "bytes", Verb: "havde"},
	"array":  {Unit: "elementer", Verb: "indeholdt"},
	"slice":  {Unit: "elementer", Verb: "indeholdt"},
	"set":    {Unit: "elementer", Verb: "indeholdt"},
	"map":    {Unit: "poster", Verb: "indeholdt"},
}

// Danish format noun mappings
var FormatNounsDa = map[string]string{
	"regex":            "input",
	"email":            "e-mailadresse",
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
	"datetime":         "ISO dato- og klokkeslæt",
	"date":             "ISO-dato",
	"time":             "ISO-klokkeslæt",
	"duration":         "ISO-varighed",
	"ipv4":             "IPv4-adresse",
	"ipv6":             "IPv6-adresse",
	"mac":              "MAC-adresse",
	"cidrv4":           "IPv4-spektrum",
	"cidrv6":           "IPv6-spektrum",
	"base64":           "base64-kodet streng",
	"base64url":        "base64url-kodet streng",
	"json_string":      "JSON-streng",
	"e164":             "E.164-nummer",
	"jwt":              "JWT",
	"template_literal": "input",
}

// Danish type dictionary
var TypeDictionaryDa = map[string]string{
	"nan":       "NaN",
	"number":    "tal",
	"array":     "liste",
	"slice":     "liste",
	"string":    "streng",
	"bool":      "boolean",
	"object":    "objekt",
	"map":       "kort",
	"nil":       "null",
	"undefined": "udefineret",
	"function":  "funktion",
	"date":      "dato",
	"file":      "fil",
	"set":       "sæt",
}

// getSizingDa returns Danish sizing information for a given type
func getSizingDa(origin string) *issues.SizingInfo {
	if info, exists := SizableDa[origin]; exists {
		return new(info)
	}
	return nil
}

// getFormatNounDa returns Danish noun for a format name
func getFormatNounDa(format string) string {
	if noun, exists := FormatNounsDa[format]; exists {
		return noun
	}
	return format
}

// getTypeNameDa returns Danish type name
func getTypeNameDa(typeName string) string {
	if name, exists := TypeDictionaryDa[typeName]; exists {
		return name
	}
	return typeName
}

// formatDa provides Danish error messages
func formatDa(raw core.ZodRawIssue) string {
	code := raw.Code

	switch code {
	case core.InvalidType:
		expected := mapx.StringOr(raw.Properties, "expected", "")
		expected = getTypeNameDa(expected)
		received := issues.ParsedTypeToString(raw.Input)
		received = getTypeNameDa(received)
		return fmt.Sprintf("Ugyldigt input: forventede %s, fik %s", expected, received)

	case core.InvalidValue:
		values := mapx.AnySliceOr(raw.Properties, "values", nil)
		if len(values) == 0 {
			return "Ugyldig værdi"
		}
		if len(values) == 1 {
			return fmt.Sprintf("Ugyldig værdi: forventede %s", issues.StringifyPrimitive(values[0]))
		}
		return fmt.Sprintf("Ugyldigt valg: forventede en af følgende %s",
			issues.JoinValuesWithSeparator(values, "|"))

	case core.TooBig:
		return formatSizeConstraintDa(raw, false)

	case core.TooSmall:
		return formatSizeConstraintDa(raw, true)

	case core.InvalidFormat:
		format := mapx.StringOr(raw.Properties, "format", "")
		if format == "" {
			return "Ugyldigt format"
		}
		return formatStringValidationDa(raw, format)

	case core.NotMultipleOf:
		divisor := mapx.AnyOr(raw.Properties, "divisor", nil)
		if divisor == nil {
			return "Ugyldigt tal: skal være deleligt"
		}
		return fmt.Sprintf("Ugyldigt tal: skal være deleligt med %v", divisor)

	case core.UnrecognizedKeys:
		keys := mapx.StringsOr(raw.Properties, "keys", nil)
		if len(keys) == 0 {
			return "Ukendt nøgle"
		}
		keyWord := "Ukendt nøgle"
		if len(keys) > 1 {
			keyWord = "Ukendte nøgler"
		}
		if keysAny, err := slicex.ToAny(keys); err == nil && !slicex.IsEmpty(keysAny) {
			keysJoined := issues.JoinValuesWithSeparator(keysAny, ", ")
			return fmt.Sprintf("%s: %s", keyWord, keysJoined)
		}
		return keyWord

	case core.InvalidKey:
		origin := mapx.StringOr(raw.Properties, "origin", "")
		if origin == "" {
			return "Ugyldig nøgle"
		}
		return fmt.Sprintf("Ugyldig nøgle i %s", origin)

	case core.InvalidUnion:
		return "Ugyldigt input: matcher ingen af de tilladte typer"

	case core.InvalidElement:
		origin := mapx.StringOr(raw.Properties, "origin", "")
		if origin == "" {
			return "Ugyldigt element"
		}
		return fmt.Sprintf("Ugyldig værdi i %s", origin)

	case core.MissingRequired:
		fieldName := mapx.StringOr(raw.Properties, "field_name", "")
		fieldType := mapx.StringOr(raw.Properties, "field_type", "felt")
		if fieldName == "" {
			return fmt.Sprintf("Påkrævet %s mangler", fieldType)
		}
		return fmt.Sprintf("Påkrævet %s mangler: %s", fieldType, fieldName)

	case core.TypeConversion:
		fromType := mapx.StringOr(raw.Properties, "from_type", "ukendt")
		toType := mapx.StringOr(raw.Properties, "to_type", "ukendt")
		return fmt.Sprintf("Typekonvertering mislykkedes: kan ikke konvertere %s til %s", fromType, toType)

	case core.InvalidSchema:
		reason := mapx.StringOr(raw.Properties, "reason", "")
		if reason != "" {
			return fmt.Sprintf("Ugyldigt skema: %s", reason)
		}
		return "Ugyldig skemadefinition"

	case core.InvalidDiscriminator:
		field := mapx.StringOr(raw.Properties, "field", "diskriminator")
		return fmt.Sprintf("Ugyldigt eller manglende diskriminatorfelt: %s", field)

	case core.IncompatibleTypes:
		conflictType := mapx.StringOr(raw.Properties, "conflict_type", "værdier")
		return fmt.Sprintf("Kan ikke flette %s: inkompatible typer", conflictType)

	case core.NilPointer:
		return "Null-pointer opdaget"

	case core.Custom:
		message := mapx.StringOr(raw.Properties, "message", "")
		if message != "" {
			return message
		}
		return "Ugyldigt input"

	default:
		return "Ugyldigt input"
	}
}

// formatSizeConstraintDa formats size constraint messages in Danish
func formatSizeConstraintDa(raw core.ZodRawIssue, isTooSmall bool) string {
	origin := mapx.StringOr(raw.Properties, "origin", "")
	if origin == "" {
		origin = "værdi"
	}

	var threshold any
	if isTooSmall {
		threshold = mapx.AnyOr(raw.Properties, "minimum", nil)
	} else {
		threshold = mapx.AnyOr(raw.Properties, "maximum", nil)
	}

	if threshold == nil {
		if isTooSmall {
			return "For lille"
		}
		return "For stor"
	}

	inclusive := mapx.BoolOr(raw.Properties, "inclusive", true)
	sizing := getSizingDa(mapx.StringOr(raw.Properties, "origin", ""))
	thresholdStr := issues.FormatThreshold(threshold)

	// Danish comparison operators
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
			return fmt.Sprintf("For lille: forventede %s %s %s%s %s", origin, sizing.Verb, adj, thresholdStr, sizing.Unit)
		}
		return fmt.Sprintf("For stor: forventede %s %s %s%s %s", origin, sizing.Verb, adj, thresholdStr, sizing.Unit)
	}

	if isTooSmall {
		return fmt.Sprintf("For lille: forventede %s havde %s%s", origin, adj, thresholdStr)
	}
	return fmt.Sprintf("For stor: forventede %s havde %s%s", origin, adj, thresholdStr)
}

// formatStringValidationDa handles string format validation messages in Danish
func formatStringValidationDa(raw core.ZodRawIssue, format string) string {
	switch format {
	case "starts_with":
		prefix := mapx.StringOr(raw.Properties, "prefix", "")
		if prefix == "" {
			return "Ugyldig streng: skal starte med angivet præfiks"
		}
		return fmt.Sprintf("Ugyldig streng: skal starte med \"%s\"", prefix)
	case "ends_with":
		suffix := mapx.StringOr(raw.Properties, "suffix", "")
		if suffix == "" {
			return "Ugyldig streng: skal ende med angivet suffiks"
		}
		return fmt.Sprintf("Ugyldig streng: skal ende med \"%s\"", suffix)
	case "includes":
		includes := mapx.StringOr(raw.Properties, "includes", "")
		if includes == "" {
			return "Ugyldig streng: skal indeholde angivet delstreng"
		}
		return fmt.Sprintf("Ugyldig streng: skal indeholde \"%s\"", includes)
	case "regex":
		pattern := mapx.StringOr(raw.Properties, "pattern", "")
		if pattern == "" {
			return "Ugyldig streng: skal matche mønsteret"
		}
		return fmt.Sprintf("Ugyldig streng: skal matche mønsteret %s", pattern)
	default:
		noun := getFormatNounDa(format)
		return fmt.Sprintf("Ugyldig %s", noun)
	}
}

// Da returns a ZodConfig configured for Danish locale.
func Da() *core.ZodConfig {
	return &core.ZodConfig{
		LocaleError: formatDa,
	}
}

// FormatMessageDa formats a single issue using Danish locale
func FormatMessageDa(issue core.ZodRawIssue) string {
	return formatDa(issue)
}
