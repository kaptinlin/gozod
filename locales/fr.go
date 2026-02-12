//nolint:misspell // French locale contains non-English words
package locales

import (
	"fmt"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// =============================================================================
// FRENCH LOCALE FORMATTER
// =============================================================================

// French sizing info mappings
var SizableFr = map[string]issues.SizingInfo{
	"string": {Unit: "caractères", Verb: "avoir"},
	"file":   {Unit: "octets", Verb: "avoir"},
	"array":  {Unit: "éléments", Verb: "avoir"},
	"slice":  {Unit: "éléments", Verb: "avoir"},
	"set":    {Unit: "éléments", Verb: "avoir"},
	"map":    {Unit: "entrées", Verb: "avoir"},
}

// French format noun mappings
var FormatNounsFr = map[string]string{
	"regex":            "entrée",
	"email":            "adresse e-mail",
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
	"datetime":         "date et heure ISO",
	"date":             "date ISO",
	"time":             "heure ISO",
	"duration":         "durée ISO",
	"ipv4":             "adresse IPv4",
	"ipv6":             "adresse IPv6",
	"mac":              "adresse MAC",
	"cidrv4":           "plage IPv4",
	"cidrv6":           "plage IPv6",
	"base64":           "chaîne encodée en base64",
	"base64url":        "chaîne encodée en base64url",
	"json_string":      "chaîne JSON",
	"e164":             "numéro E.164",
	"jwt":              "JWT",
	"template_literal": "entrée",
}

// French type dictionary
var TypeDictionaryFr = map[string]string{
	"nan":       "NaN",
	"number":    "nombre",
	"array":     "tableau",
	"slice":     "tableau",
	"string":    "chaîne",
	"bool":      "booléen",
	"object":    "objet",
	"map":       "map",
	"nil":       "null",
	"undefined": "undefined",
	"function":  "fonction",
	"date":      "date",
	"file":      "fichier",
	"set":       "ensemble",
}

// getSizingFr returns French sizing information for a given type
func getSizingFr(origin string) *issues.SizingInfo {
	if info, exists := SizableFr[origin]; exists {
		return &info
	}
	return nil
}

// getFormatNounFr returns French noun for a format name
func getFormatNounFr(format string) string {
	if noun, exists := FormatNounsFr[format]; exists {
		return noun
	}
	return format
}

// getTypeNameFr returns French type name
func getTypeNameFr(typeName string) string {
	if name, exists := TypeDictionaryFr[typeName]; exists {
		return name
	}
	return typeName
}

// formatFr provides French error messages
func formatFr(raw core.ZodRawIssue) string {
	code := raw.Code

	switch code {
	case core.InvalidType:
		expected := mapx.StringOr(raw.Properties, "expected", "")
		expected = getTypeNameFr(expected)
		received := issues.ParsedTypeToString(raw.Input)
		received = getTypeNameFr(received)
		return fmt.Sprintf("Entrée invalide : %s attendu, %s reçu", expected, received)

	case core.InvalidValue:
		values := mapx.AnySliceOr(raw.Properties, "values", nil)
		if len(values) == 0 {
			return "Valeur invalide"
		}
		if len(values) == 1 {
			return fmt.Sprintf("Entrée invalide : %s attendu", issues.StringifyPrimitive(values[0]))
		}
		return fmt.Sprintf("Option invalide : une valeur parmi %s attendue",
			issues.JoinValuesWithSeparator(values, "|"))

	case core.TooBig:
		return formatSizeConstraintFr(raw, false)

	case core.TooSmall:
		return formatSizeConstraintFr(raw, true)

	case core.InvalidFormat:
		format := mapx.StringOr(raw.Properties, "format", "")
		if format == "" {
			return "Format invalide"
		}
		return formatStringValidationFr(raw, format)

	case core.NotMultipleOf:
		divisor := mapx.AnyOr(raw.Properties, "divisor", nil)
		if divisor == nil {
			return "Nombre invalide : doit être un multiple"
		}
		return fmt.Sprintf("Nombre invalide : doit être un multiple de %v", divisor)

	case core.UnrecognizedKeys:
		keys := mapx.StringsOr(raw.Properties, "keys", nil)
		if len(keys) == 0 {
			return "Clé non reconnue"
		}
		keyWord := "Clé non reconnue"
		if len(keys) > 1 {
			keyWord = "Clés non reconnues"
		}
		if keysAny, err := slicex.ToAny(keys); err == nil && !slicex.IsEmpty(keysAny) {
			keysJoined := issues.JoinValuesWithSeparator(keysAny, ", ")
			return fmt.Sprintf("%s : %s", keyWord, keysJoined)
		}
		return keyWord

	case core.InvalidKey:
		origin := mapx.StringOr(raw.Properties, "origin", "")
		if origin == "" {
			return "Clé invalide"
		}
		return fmt.Sprintf("Clé invalide dans %s", origin)

	case core.InvalidUnion:
		return "Entrée invalide"

	case core.InvalidElement:
		origin := mapx.StringOr(raw.Properties, "origin", "")
		if origin == "" {
			return "Élément invalide"
		}
		return fmt.Sprintf("Valeur invalide dans %s", origin)

	case core.MissingRequired:
		fieldName := mapx.StringOr(raw.Properties, "field_name", "")
		fieldType := mapx.StringOr(raw.Properties, "field_type", "champ")
		if fieldName == "" {
			return fmt.Sprintf("%s requis manquant", fieldType)
		}
		return fmt.Sprintf("%s requis manquant : %s", fieldType, fieldName)

	case core.TypeConversion:
		fromType := mapx.StringOr(raw.Properties, "from_type", "inconnu")
		toType := mapx.StringOr(raw.Properties, "to_type", "inconnu")
		return fmt.Sprintf("Échec de la conversion de type : impossible de convertir %s en %s", fromType, toType)

	case core.InvalidSchema:
		reason := mapx.StringOr(raw.Properties, "reason", "")
		if reason != "" {
			return fmt.Sprintf("Schéma invalide : %s", reason)
		}
		return "Définition de schéma invalide"

	case core.InvalidDiscriminator:
		field := mapx.StringOr(raw.Properties, "field", "discriminateur")
		return fmt.Sprintf("Champ discriminateur invalide ou manquant : %s", field)

	case core.IncompatibleTypes:
		conflictType := mapx.StringOr(raw.Properties, "conflict_type", "valeurs")
		return fmt.Sprintf("Impossible de fusionner %s : types incompatibles", conflictType)

	case core.NilPointer:
		return "Pointeur nul rencontré"

	case core.Custom:
		message := mapx.StringOr(raw.Properties, "message", "")
		if message != "" {
			return message
		}
		return "Entrée invalide"

	default:
		return "Entrée invalide"
	}
}

// formatSizeConstraintFr formats size constraint messages in French
func formatSizeConstraintFr(raw core.ZodRawIssue, isTooSmall bool) string {
	origin := mapx.StringOr(raw.Properties, "origin", "")
	if origin == "" {
		origin = "valeur"
	}

	var threshold any
	if isTooSmall {
		threshold = mapx.AnyOr(raw.Properties, "minimum", nil)
	} else {
		threshold = mapx.AnyOr(raw.Properties, "maximum", nil)
	}

	if threshold == nil {
		if isTooSmall {
			return "Trop petit"
		}
		return "Trop grand"
	}

	inclusive := mapx.BoolOr(raw.Properties, "inclusive", true)
	sizing := getSizingFr(origin)
	thresholdStr := issues.FormatThreshold(threshold)

	// French comparison operators
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
			return fmt.Sprintf("Trop petit : %s doit %s %s%s %s", origin, sizing.Verb, adj, thresholdStr, sizing.Unit)
		}
		return fmt.Sprintf("Trop grand : %s doit %s %s%s %s", origin, sizing.Verb, adj, thresholdStr, sizing.Unit)
	}

	if isTooSmall {
		return fmt.Sprintf("Trop petit : %s doit être %s%s", origin, adj, thresholdStr)
	}
	return fmt.Sprintf("Trop grand : %s doit être %s%s", origin, adj, thresholdStr)
}

// formatStringValidationFr handles string format validation messages in French
func formatStringValidationFr(raw core.ZodRawIssue, format string) string {
	switch format {
	case "starts_with":
		prefix := mapx.StringOr(raw.Properties, "prefix", "")
		if prefix == "" {
			return "Chaîne invalide : doit commencer par le préfixe spécifié"
		}
		return fmt.Sprintf("Chaîne invalide : doit commencer par \"%s\"", prefix)
	case "ends_with":
		suffix := mapx.StringOr(raw.Properties, "suffix", "")
		if suffix == "" {
			return "Chaîne invalide : doit se terminer par le suffixe spécifié"
		}
		return fmt.Sprintf("Chaîne invalide : doit se terminer par \"%s\"", suffix)
	case "includes":
		includes := mapx.StringOr(raw.Properties, "includes", "")
		if includes == "" {
			return "Chaîne invalide : doit inclure la sous-chaîne spécifiée"
		}
		return fmt.Sprintf("Chaîne invalide : doit inclure \"%s\"", includes)
	case "regex":
		pattern := mapx.StringOr(raw.Properties, "pattern", "")
		if pattern == "" {
			return "Chaîne invalide : doit correspondre au modèle"
		}
		return fmt.Sprintf("Chaîne invalide : doit correspondre au modèle %s", pattern)
	default:
		noun := getFormatNounFr(format)
		return fmt.Sprintf("%s invalide", noun)
	}
}

// Fr returns a ZodConfig configured for French locale.
func Fr() *core.ZodConfig {
	return &core.ZodConfig{
		LocaleError: formatFr,
	}
}

// FormatMessageFr formats a single issue using French locale
func FormatMessageFr(issue core.ZodRawIssue) string {
	return formatFr(issue)
}
