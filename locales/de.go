//nolint:misspell // German locale contains non-English words
package locales

import (
	"fmt"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// =============================================================================
// GERMAN LOCALE FORMATTER
// =============================================================================

// German sizing info mappings
var SizableDe = map[string]issues.SizingInfo{
	"string": {Unit: "Zeichen", Verb: "haben"},
	"file":   {Unit: "Bytes", Verb: "haben"},
	"array":  {Unit: "Elemente", Verb: "haben"},
	"slice":  {Unit: "Elemente", Verb: "haben"},
	"set":    {Unit: "Elemente", Verb: "haben"},
	"map":    {Unit: "Einträge", Verb: "haben"},
}

// German format noun mappings
var FormatNounsDe = map[string]string{
	"regex":            "Eingabe",
	"email":            "E-Mail-Adresse",
	"url":              "URL",
	"emoji":            "Emoji",
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
	"datetime":         "ISO-Datum und -Uhrzeit",
	"date":             "ISO-Datum",
	"time":             "ISO-Uhrzeit",
	"duration":         "ISO-Dauer",
	"ipv4":             "IPv4-Adresse",
	"ipv6":             "IPv6-Adresse",
	"mac":              "MAC-Adresse",
	"cidrv4":           "IPv4-Bereich",
	"cidrv6":           "IPv6-Bereich",
	"base64":           "Base64-codierter String",
	"base64url":        "Base64-URL-codierter String",
	"json_string":      "JSON-String",
	"e164":             "E.164-Nummer",
	"jwt":              "JWT",
	"template_literal": "Eingabe",
}

// German type dictionary
var TypeDictionaryDe = map[string]string{
	"nan":       "NaN",
	"number":    "Zahl",
	"array":     "Array",
	"slice":     "Array",
	"string":    "String",
	"bool":      "Boolean",
	"object":    "Objekt",
	"map":       "Map",
	"nil":       "null",
	"undefined": "undefined",
	"function":  "Funktion",
	"date":      "Datum",
	"file":      "Datei",
	"set":       "Set",
}

// getSizingDe returns German sizing information for a given type
func getSizingDe(origin string) *issues.SizingInfo {
	if info, exists := SizableDe[origin]; exists {
		return &info
	}
	return nil
}

// getFormatNounDe returns German noun for a format name
func getFormatNounDe(format string) string {
	if noun, exists := FormatNounsDe[format]; exists {
		return noun
	}
	return format
}

// getTypeNameDe returns German type name
func getTypeNameDe(typeName string) string {
	if name, exists := TypeDictionaryDe[typeName]; exists {
		return name
	}
	return typeName
}

// formatDe provides German error messages
func formatDe(raw core.ZodRawIssue) string {
	code := raw.Code

	switch code {
	case core.InvalidType:
		expected := mapx.GetStringDefault(raw.Properties, "expected", "")
		expected = getTypeNameDe(expected)
		received := issues.ParsedTypeToString(raw.Input)
		received = getTypeNameDe(received)
		return fmt.Sprintf("Ungültige Eingabe: erwartet %s, erhalten %s", expected, received)

	case core.InvalidValue:
		values := mapx.GetAnySliceDefault(raw.Properties, "values", nil)
		if len(values) == 0 {
			return "Ungültiger Wert"
		}
		if len(values) == 1 {
			return fmt.Sprintf("Ungültige Eingabe: erwartet %s", issues.StringifyPrimitive(values[0]))
		}
		return fmt.Sprintf("Ungültige Option: erwartet eine von %s",
			issues.JoinValuesWithSeparator(values, "|"))

	case core.TooBig:
		return formatSizeConstraintDe(raw, false)

	case core.TooSmall:
		return formatSizeConstraintDe(raw, true)

	case core.InvalidFormat:
		format := mapx.GetStringDefault(raw.Properties, "format", "")
		if format == "" {
			return "Ungültiges Format"
		}
		return formatStringValidationDe(raw, format)

	case core.NotMultipleOf:
		divisor := mapx.GetAnyDefault(raw.Properties, "divisor", nil)
		if divisor == nil {
			return "Ungültige Zahl: muss ein Vielfaches sein"
		}
		return fmt.Sprintf("Ungültige Zahl: muss ein Vielfaches von %v sein", divisor)

	case core.UnrecognizedKeys:
		keys := mapx.GetStringsDefault(raw.Properties, "keys", nil)
		if len(keys) == 0 {
			return "Unbekannter Schlüssel"
		}
		keyWord := "Unbekannter Schlüssel"
		if len(keys) > 1 {
			keyWord = "Unbekannte Schlüssel"
		}
		if keysAny, err := slicex.ToAny(keys); err == nil && !slicex.IsEmpty(keysAny) {
			keysJoined := issues.JoinValuesWithSeparator(keysAny, ", ")
			return fmt.Sprintf("%s: %s", keyWord, keysJoined)
		}
		return keyWord

	case core.InvalidKey:
		origin := mapx.GetStringDefault(raw.Properties, "origin", "")
		if origin == "" {
			return "Ungültiger Schlüssel"
		}
		return fmt.Sprintf("Ungültiger Schlüssel in %s", origin)

	case core.InvalidUnion:
		return "Ungültige Eingabe"

	case core.InvalidElement:
		origin := mapx.GetStringDefault(raw.Properties, "origin", "")
		if origin == "" {
			return "Ungültiges Element"
		}
		return fmt.Sprintf("Ungültiger Wert in %s", origin)

	case core.MissingRequired:
		fieldName := mapx.GetStringDefault(raw.Properties, "field_name", "")
		fieldType := mapx.GetStringDefault(raw.Properties, "field_type", "Feld")
		if fieldName == "" {
			return fmt.Sprintf("Erforderliches %s fehlt", fieldType)
		}
		return fmt.Sprintf("Erforderliches %s fehlt: %s", fieldType, fieldName)

	case core.TypeConversion:
		fromType := mapx.GetStringDefault(raw.Properties, "from_type", "unbekannt")
		toType := mapx.GetStringDefault(raw.Properties, "to_type", "unbekannt")
		return fmt.Sprintf("Typkonvertierung fehlgeschlagen: kann %s nicht in %s konvertieren", fromType, toType)

	case core.InvalidSchema:
		reason := mapx.GetStringDefault(raw.Properties, "reason", "")
		if reason != "" {
			return fmt.Sprintf("Ungültiges Schema: %s", reason)
		}
		return "Ungültige Schemadefinition"

	case core.InvalidDiscriminator:
		field := mapx.GetStringDefault(raw.Properties, "field", "Diskriminator")
		return fmt.Sprintf("Ungültiges oder fehlendes Diskriminatorfeld: %s", field)

	case core.IncompatibleTypes:
		conflictType := mapx.GetStringDefault(raw.Properties, "conflict_type", "Werte")
		return fmt.Sprintf("Kann %s nicht zusammenführen: inkompatible Typen", conflictType)

	case core.NilPointer:
		return "Null-Zeiger erkannt"

	case core.Custom:
		message := mapx.GetStringDefault(raw.Properties, "message", "")
		if message != "" {
			return message
		}
		return "Ungültige Eingabe"

	default:
		return "Ungültige Eingabe"
	}
}

// formatSizeConstraintDe formats size constraint messages in German
func formatSizeConstraintDe(raw core.ZodRawIssue, isTooSmall bool) string {
	origin := mapx.GetStringDefault(raw.Properties, "origin", "")
	if origin == "" {
		origin = "Wert"
	}

	var threshold any
	if isTooSmall {
		threshold = mapx.GetAnyDefault(raw.Properties, "minimum", nil)
	} else {
		threshold = mapx.GetAnyDefault(raw.Properties, "maximum", nil)
	}

	if threshold == nil {
		if isTooSmall {
			return "Zu klein"
		}
		return "Zu groß"
	}

	inclusive := mapx.GetBoolDefault(raw.Properties, "inclusive", true)
	sizing := getSizingDe(origin)
	thresholdStr := issues.FormatThreshold(threshold)

	// German comparison operators
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
			return fmt.Sprintf("Zu klein: erwartet, dass %s %s%s %s hat", origin, adj, thresholdStr, sizing.Unit)
		}
		return fmt.Sprintf("Zu groß: erwartet, dass %s %s%s %s hat", origin, adj, thresholdStr, sizing.Unit)
	}

	if isTooSmall {
		return fmt.Sprintf("Zu klein: erwartet, dass %s %s%s ist", origin, adj, thresholdStr)
	}
	return fmt.Sprintf("Zu groß: erwartet, dass %s %s%s ist", origin, adj, thresholdStr)
}

// formatStringValidationDe handles string format validation messages in German
func formatStringValidationDe(raw core.ZodRawIssue, format string) string {
	switch format {
	case "starts_with":
		prefix := mapx.GetStringDefault(raw.Properties, "prefix", "")
		if prefix == "" {
			return "Ungültiger String: muss mit dem angegebenen Präfix beginnen"
		}
		return fmt.Sprintf("Ungültiger String: muss mit \"%s\" beginnen", prefix)
	case "ends_with":
		suffix := mapx.GetStringDefault(raw.Properties, "suffix", "")
		if suffix == "" {
			return "Ungültiger String: muss mit dem angegebenen Suffix enden"
		}
		return fmt.Sprintf("Ungültiger String: muss mit \"%s\" enden", suffix)
	case "includes":
		includes := mapx.GetStringDefault(raw.Properties, "includes", "")
		if includes == "" {
			return "Ungültiger String: muss den angegebenen Teilstring enthalten"
		}
		return fmt.Sprintf("Ungültiger String: muss \"%s\" enthalten", includes)
	case "regex":
		pattern := mapx.GetStringDefault(raw.Properties, "pattern", "")
		if pattern == "" {
			return "Ungültiger String: muss dem Muster entsprechen"
		}
		return fmt.Sprintf("Ungültiger String: muss dem Muster %s entsprechen", pattern)
	default:
		noun := getFormatNounDe(format)
		return fmt.Sprintf("Ungültig: %s", noun)
	}
}

// De returns a ZodConfig configured for German locale.
func De() *core.ZodConfig {
	return &core.ZodConfig{
		LocaleError: formatDe,
	}
}

// FormatMessageDe formats a single issue using German locale
func FormatMessageDe(issue core.ZodRawIssue) string {
	return formatDe(issue)
}
