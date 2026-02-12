//nolint:misspell // Italian locale contains non-English words
package locales

import (
	"fmt"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// =============================================================================
// ITALIAN LOCALE FORMATTER
// =============================================================================

// Italian sizing info mappings
var SizableIt = map[string]issues.SizingInfo{
	"string": {Unit: "caratteri", Verb: "avere"},
	"file":   {Unit: "byte", Verb: "avere"},
	"array":  {Unit: "elementi", Verb: "avere"},
	"slice":  {Unit: "elementi", Verb: "avere"},
	"set":    {Unit: "elementi", Verb: "avere"},
	"map":    {Unit: "voci", Verb: "avere"},
}

// Italian format noun mappings
var FormatNounsIt = map[string]string{
	"regex":            "input",
	"email":            "indirizzo email",
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
	"datetime":         "data e ora ISO",
	"date":             "data ISO",
	"time":             "ora ISO",
	"duration":         "durata ISO",
	"ipv4":             "indirizzo IPv4",
	"ipv6":             "indirizzo IPv6",
	"mac":              "indirizzo MAC",
	"cidrv4":           "intervallo IPv4",
	"cidrv6":           "intervallo IPv6",
	"base64":           "stringa codificata in base64",
	"base64url":        "URL codificata in base64",
	"json_string":      "stringa JSON",
	"e164":             "numero E.164",
	"jwt":              "JWT",
	"template_literal": "input",
}

// Italian type dictionary
var TypeDictionaryIt = map[string]string{
	"nan":       "NaN",
	"number":    "numero",
	"array":     "vettore",
	"slice":     "vettore",
	"string":    "stringa",
	"bool":      "booleano",
	"object":    "oggetto",
	"map":       "mappa",
	"nil":       "nullo",
	"undefined": "indefinito",
	"function":  "funzione",
	"date":      "data",
	"file":      "file",
	"set":       "insieme",
}

// getSizingIt returns Italian sizing information for a given type
func getSizingIt(origin string) *issues.SizingInfo {
	if info, exists := SizableIt[origin]; exists {
		return &info
	}
	return nil
}

// getFormatNounIt returns Italian noun for a format name
func getFormatNounIt(format string) string {
	if noun, exists := FormatNounsIt[format]; exists {
		return noun
	}
	return format
}

// getTypeNameIt returns Italian type name
func getTypeNameIt(typeName string) string {
	if name, exists := TypeDictionaryIt[typeName]; exists {
		return name
	}
	return typeName
}

// formatIt provides Italian error messages
func formatIt(raw core.ZodRawIssue) string {
	code := raw.Code

	switch code {
	case core.InvalidType:
		expected := mapx.StringOr(raw.Properties, "expected", "")
		expected = getTypeNameIt(expected)
		received := issues.ParsedTypeToString(raw.Input)
		received = getTypeNameIt(received)
		return fmt.Sprintf("Input non valido: atteso %s, ricevuto %s", expected, received)

	case core.InvalidValue:
		values := mapx.AnySliceOr(raw.Properties, "values", nil)
		if len(values) == 0 {
			return "Valore non valido"
		}
		if len(values) == 1 {
			return fmt.Sprintf("Input non valido: atteso %s", issues.StringifyPrimitive(values[0]))
		}
		return fmt.Sprintf("Opzione non valida: atteso uno tra %s",
			issues.JoinValuesWithSeparator(values, "|"))

	case core.TooBig:
		return formatSizeConstraintIt(raw, false)

	case core.TooSmall:
		return formatSizeConstraintIt(raw, true)

	case core.InvalidFormat:
		format := mapx.StringOr(raw.Properties, "format", "")
		if format == "" {
			return "Formato non valido"
		}
		return formatStringValidationIt(raw, format)

	case core.NotMultipleOf:
		divisor := mapx.AnyOr(raw.Properties, "divisor", nil)
		if divisor == nil {
			return "Numero non valido: deve essere un multiplo"
		}
		return fmt.Sprintf("Numero non valido: deve essere un multiplo di %v", divisor)

	case core.UnrecognizedKeys:
		keys := mapx.StringsOr(raw.Properties, "keys", nil)
		if len(keys) == 0 {
			return "Chiave non riconosciuta"
		}
		keyWord := "Chiave non riconosciuta"
		if len(keys) > 1 {
			keyWord = "Chiavi non riconosciute"
		}
		if keysAny, err := slicex.ToAny(keys); err == nil && !slicex.IsEmpty(keysAny) {
			keysJoined := issues.JoinValuesWithSeparator(keysAny, ", ")
			return fmt.Sprintf("%s: %s", keyWord, keysJoined)
		}
		return keyWord

	case core.InvalidKey:
		origin := mapx.StringOr(raw.Properties, "origin", "")
		if origin == "" {
			return "Chiave non valida"
		}
		return fmt.Sprintf("Chiave non valida in %s", origin)

	case core.InvalidUnion:
		return "Input non valido"

	case core.InvalidElement:
		origin := mapx.StringOr(raw.Properties, "origin", "")
		if origin == "" {
			return "Elemento non valido"
		}
		return fmt.Sprintf("Valore non valido in %s", origin)

	case core.MissingRequired:
		fieldName := mapx.StringOr(raw.Properties, "field_name", "")
		fieldType := mapx.StringOr(raw.Properties, "field_type", "campo")
		if fieldName == "" {
			return fmt.Sprintf("Manca %s obbligatorio", fieldType)
		}
		return fmt.Sprintf("Manca %s obbligatorio: %s", fieldType, fieldName)

	case core.TypeConversion:
		fromType := mapx.StringOr(raw.Properties, "from_type", "sconosciuto")
		toType := mapx.StringOr(raw.Properties, "to_type", "sconosciuto")
		return fmt.Sprintf("Conversione di tipo fallita: impossibile convertire %s in %s", fromType, toType)

	case core.InvalidSchema:
		reason := mapx.StringOr(raw.Properties, "reason", "")
		if reason != "" {
			return fmt.Sprintf("Schema non valido: %s", reason)
		}
		return "Definizione dello schema non valida"

	case core.InvalidDiscriminator:
		field := mapx.StringOr(raw.Properties, "field", "discriminatore")
		return fmt.Sprintf("Campo discriminatore non valido o mancante: %s", field)

	case core.IncompatibleTypes:
		conflictType := mapx.StringOr(raw.Properties, "conflict_type", "valori")
		return fmt.Sprintf("Impossibile unire %s: tipi incompatibili", conflictType)

	case core.NilPointer:
		return "Puntatore nullo rilevato"

	case core.Custom:
		message := mapx.StringOr(raw.Properties, "message", "")
		if message != "" {
			return message
		}
		return "Input non valido"

	default:
		return "Input non valido"
	}
}

// formatSizeConstraintIt formats size constraint messages in Italian
func formatSizeConstraintIt(raw core.ZodRawIssue, isTooSmall bool) string {
	origin := mapx.StringOr(raw.Properties, "origin", "")
	if origin == "" {
		origin = "valore"
	}

	var threshold any
	if isTooSmall {
		threshold = mapx.AnyOr(raw.Properties, "minimum", nil)
	} else {
		threshold = mapx.AnyOr(raw.Properties, "maximum", nil)
	}

	if threshold == nil {
		if isTooSmall {
			return "Troppo piccolo"
		}
		return "Troppo grande"
	}

	inclusive := mapx.BoolOr(raw.Properties, "inclusive", true)
	sizing := getSizingIt(mapx.StringOr(raw.Properties, "origin", ""))
	thresholdStr := issues.FormatThreshold(threshold)

	// Italian comparison operators
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
			return fmt.Sprintf("Troppo piccolo: %s deve avere %s%s %s", origin, adj, thresholdStr, sizing.Unit)
		}
		return fmt.Sprintf("Troppo grande: %s deve avere %s%s %s", origin, adj, thresholdStr, sizing.Unit)
	}

	if isTooSmall {
		return fmt.Sprintf("Troppo piccolo: %s deve essere %s%s", origin, adj, thresholdStr)
	}
	return fmt.Sprintf("Troppo grande: %s deve essere %s%s", origin, adj, thresholdStr)
}

// formatStringValidationIt handles string format validation messages in Italian
func formatStringValidationIt(raw core.ZodRawIssue, format string) string {
	switch format {
	case "starts_with":
		prefix := mapx.StringOr(raw.Properties, "prefix", "")
		if prefix == "" {
			return "Stringa non valida: deve iniziare con il prefisso specificato"
		}
		return fmt.Sprintf("Stringa non valida: deve iniziare con \"%s\"", prefix)
	case "ends_with":
		suffix := mapx.StringOr(raw.Properties, "suffix", "")
		if suffix == "" {
			return "Stringa non valida: deve terminare con il suffisso specificato"
		}
		return fmt.Sprintf("Stringa non valida: deve terminare con \"%s\"", suffix)
	case "includes":
		includes := mapx.StringOr(raw.Properties, "includes", "")
		if includes == "" {
			return "Stringa non valida: deve includere la sottostringa specificata"
		}
		return fmt.Sprintf("Stringa non valida: deve includere \"%s\"", includes)
	case "regex":
		pattern := mapx.StringOr(raw.Properties, "pattern", "")
		if pattern == "" {
			return "Stringa non valida: deve corrispondere al pattern"
		}
		return fmt.Sprintf("Stringa non valida: deve corrispondere al pattern %s", pattern)
	default:
		noun := getFormatNounIt(format)
		return fmt.Sprintf("%s non valido", noun)
	}
}

// It returns a ZodConfig configured for Italian locale.
func It() *core.ZodConfig {
	return &core.ZodConfig{
		LocaleError: formatIt,
	}
}

// FormatMessageIt formats a single issue using Italian locale
func FormatMessageIt(issue core.ZodRawIssue) string {
	return formatIt(issue)
}
