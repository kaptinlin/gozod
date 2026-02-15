package locales

import (
	"fmt"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// =============================================================================
// SPANISH LOCALE FORMATTER
// =============================================================================

// SizableEs maps Spanish sizing information.
var SizableEs = map[string]issues.SizingInfo{
	"string": {Unit: "caracteres", Verb: "tener"},
	"file":   {Unit: "bytes", Verb: "tener"},
	"array":  {Unit: "elementos", Verb: "tener"},
	"slice":  {Unit: "elementos", Verb: "tener"},
	"set":    {Unit: "elementos", Verb: "tener"},
	"map":    {Unit: "entradas", Verb: "tener"},
}

// FormatNounsEs maps Spanish format noun translations.
var FormatNounsEs = map[string]string{
	"regex":            "entrada",
	"email":            "dirección de correo electrónico",
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
	"datetime":         "fecha y hora ISO",
	"date":             "fecha ISO",
	"time":             "hora ISO",
	"duration":         "duración ISO",
	"ipv4":             "dirección IPv4",
	"ipv6":             "dirección IPv6",
	"mac":              "dirección MAC",
	"cidrv4":           "rango IPv4",
	"cidrv6":           "rango IPv6",
	"base64":           "cadena codificada en base64",
	"base64url":        "URL codificada en base64",
	"json_string":      "cadena JSON",
	"e164":             "número E.164",
	"jwt":              "JWT",
	"template_literal": "entrada",
}

// TypeDictionaryEs maps Spanish type name translations.
var TypeDictionaryEs = map[string]string{
	"nan":       "NaN",
	"string":    "texto",
	"number":    "número",
	"boolean":   "booleano",
	"bool":      "booleano",
	"array":     "arreglo",
	"slice":     "arreglo",
	"object":    "objeto",
	"set":       "conjunto",
	"file":      "archivo",
	"date":      "fecha",
	"bigint":    "número grande",
	"symbol":    "símbolo",
	"undefined": "indefinido",
	"nil":       "nulo",
	"null":      "nulo",
	"function":  "función",
	"map":       "mapa",
	"record":    "registro",
	"tuple":     "tupla",
	"enum":      "enumeración",
	"union":     "unión",
	"literal":   "literal",
	"promise":   "promesa",
	"void":      "vacío",
	"never":     "nunca",
	"unknown":   "desconocido",
	"any":       "cualquiera",
}

// getSizingEs returns Spanish sizing information for a given type
func getSizingEs(origin string) *issues.SizingInfo {
	if _, exists := SizableEs[origin]; exists {
		return new(SizableEs[origin])
	}
	return nil
}

// getFormatNounEs returns Spanish noun for a format name
func getFormatNounEs(format string) string {
	if noun, exists := FormatNounsEs[format]; exists {
		return noun
	}
	return format
}

// getTypeNameEs returns Spanish type name
func getTypeNameEs(typeName string) string {
	if name, exists := TypeDictionaryEs[typeName]; exists {
		return name
	}
	return typeName
}

// formatEs provides Spanish error messages
func formatEs(raw core.ZodRawIssue) string {
	code := raw.Code

	switch code {
	case core.InvalidType:
		expected := mapx.StringOr(raw.Properties, "expected", "")
		expected = getTypeNameEs(expected)
		received := issues.ParsedTypeToString(raw.Input)
		received = getTypeNameEs(received)
		return fmt.Sprintf("Entrada inválida: se esperaba %s, recibido %s", expected, received)

	case core.InvalidValue:
		values := mapx.AnySliceOr(raw.Properties, "values", nil)
		if len(values) == 0 {
			return "Valor inválido"
		}
		if len(values) == 1 {
			return fmt.Sprintf("Entrada inválida: se esperaba %s", issues.StringifyPrimitive(values[0]))
		}
		return fmt.Sprintf("Opción inválida: se esperaba una de %s",
			issues.JoinValuesWithSeparator(values, "|"))

	case core.TooBig:
		return formatSizeConstraintEs(raw, false)

	case core.TooSmall:
		return formatSizeConstraintEs(raw, true)

	case core.InvalidFormat:
		format := mapx.StringOr(raw.Properties, "format", "")
		if format == "" {
			return "Formato inválido"
		}
		return formatStringValidationEs(raw, format)

	case core.NotMultipleOf:
		divisor := mapx.AnyOr(raw.Properties, "divisor", nil)
		if divisor == nil {
			return "Número inválido: debe ser múltiplo"
		}
		return fmt.Sprintf("Número inválido: debe ser múltiplo de %v", divisor)

	case core.UnrecognizedKeys:
		keys := mapx.StringsOr(raw.Properties, "keys", nil)
		if len(keys) == 0 {
			return "Llave desconocida"
		}
		keyWord := "Llave desconocida"
		if len(keys) > 1 {
			keyWord = "Llaves desconocidas"
		}
		if keysAny, err := slicex.ToAny(keys); err == nil && !slicex.IsEmpty(keysAny) {
			keysJoined := issues.JoinValuesWithSeparator(keysAny, ", ")
			return fmt.Sprintf("%s: %s", keyWord, keysJoined)
		}
		return keyWord

	case core.InvalidKey:
		origin := mapx.StringOr(raw.Properties, "origin", "")
		origin = getTypeNameEs(origin)
		if origin == "" {
			return "Llave inválida"
		}
		return fmt.Sprintf("Llave inválida en %s", origin)

	case core.InvalidUnion:
		return "Entrada inválida"

	case core.InvalidElement:
		origin := mapx.StringOr(raw.Properties, "origin", "")
		origin = getTypeNameEs(origin)
		if origin == "" {
			return "Elemento inválido"
		}
		return fmt.Sprintf("Valor inválido en %s", origin)

	case core.MissingRequired:
		fieldName := mapx.StringOr(raw.Properties, "field_name", "")
		fieldType := mapx.StringOr(raw.Properties, "field_type", "campo")
		if fieldName == "" {
			return fmt.Sprintf("Falta %s requerido", fieldType)
		}
		return fmt.Sprintf("Falta %s requerido: %s", fieldType, fieldName)

	case core.TypeConversion:
		fromType := mapx.StringOr(raw.Properties, "from_type", "desconocido")
		toType := mapx.StringOr(raw.Properties, "to_type", "desconocido")
		return fmt.Sprintf("Error de conversión de tipo: no se puede convertir %s a %s", fromType, toType)

	case core.InvalidSchema:
		reason := mapx.StringOr(raw.Properties, "reason", "")
		if reason != "" {
			return fmt.Sprintf("Esquema inválido: %s", reason)
		}
		return "Definición de esquema inválida"

	case core.InvalidDiscriminator:
		field := mapx.StringOr(raw.Properties, "field", "discriminador")
		return fmt.Sprintf("Campo discriminador inválido o faltante: %s", field)

	case core.IncompatibleTypes:
		conflictType := mapx.StringOr(raw.Properties, "conflict_type", "valores")
		return fmt.Sprintf("No se pueden fusionar %s: tipos incompatibles", conflictType)

	case core.NilPointer:
		return "Se encontró un puntero nulo"

	case core.Custom:
		message := mapx.StringOr(raw.Properties, "message", "")
		if message != "" {
			return message
		}
		return "Entrada inválida"

	default:
		return "Entrada inválida"
	}
}

// formatSizeConstraintEs formats size constraint messages in Spanish
func formatSizeConstraintEs(raw core.ZodRawIssue, isTooSmall bool) string {
	origin := mapx.StringOr(raw.Properties, "origin", "")
	origin = getTypeNameEs(origin)
	if origin == "" {
		origin = "valor"
	}

	var threshold any
	if isTooSmall {
		threshold = mapx.AnyOr(raw.Properties, "minimum", nil)
	} else {
		threshold = mapx.AnyOr(raw.Properties, "maximum", nil)
	}

	if threshold == nil {
		if isTooSmall {
			return "Demasiado pequeño"
		}
		return "Demasiado grande"
	}

	inclusive := mapx.BoolOr(raw.Properties, "inclusive", true)
	sizing := getSizingEs(mapx.StringOr(raw.Properties, "origin", ""))
	thresholdStr := issues.FormatThreshold(threshold)

	// Spanish comparison operators
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
			return fmt.Sprintf("Demasiado pequeño: se esperaba que %s tuviera %s%s %s", origin, adj, thresholdStr, sizing.Unit)
		}
		return fmt.Sprintf("Demasiado grande: se esperaba que %s tuviera %s%s %s", origin, adj, thresholdStr, sizing.Unit)
	}

	if isTooSmall {
		return fmt.Sprintf("Demasiado pequeño: se esperaba que %s fuera %s%s", origin, adj, thresholdStr)
	}
	return fmt.Sprintf("Demasiado grande: se esperaba que %s fuera %s%s", origin, adj, thresholdStr)
}

// formatStringValidationEs handles string format validation messages in Spanish
func formatStringValidationEs(raw core.ZodRawIssue, format string) string {
	switch format {
	case "starts_with":
		prefix := mapx.StringOr(raw.Properties, "prefix", "")
		if prefix == "" {
			return "Cadena inválida: debe comenzar con el prefijo especificado"
		}
		return fmt.Sprintf("Cadena inválida: debe comenzar con \"%s\"", prefix)
	case "ends_with":
		suffix := mapx.StringOr(raw.Properties, "suffix", "")
		if suffix == "" {
			return "Cadena inválida: debe terminar con el sufijo especificado"
		}
		return fmt.Sprintf("Cadena inválida: debe terminar en \"%s\"", suffix)
	case "includes":
		includes := mapx.StringOr(raw.Properties, "includes", "")
		if includes == "" {
			return "Cadena inválida: debe incluir la subcadena especificada"
		}
		return fmt.Sprintf("Cadena inválida: debe incluir \"%s\"", includes)
	case "regex":
		pattern := mapx.StringOr(raw.Properties, "pattern", "")
		if pattern == "" {
			return "Cadena inválida: debe coincidir con el patrón"
		}
		return fmt.Sprintf("Cadena inválida: debe coincidir con el patrón %s", pattern)
	default:
		noun := getFormatNounEs(format)
		return fmt.Sprintf("Inválido %s", noun)
	}
}

// Es returns a ZodConfig configured for Spanish locale.
func Es() *core.ZodConfig {
	return &core.ZodConfig{
		LocaleError: formatEs,
	}
}

// FormatMessageEs formats a single issue using Spanish locale
func FormatMessageEs(issue core.ZodRawIssue) string {
	return formatEs(issue)
}
