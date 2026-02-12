//nolint:misspell // Portuguese locale contains non-English words
package locales

import (
	"fmt"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// =============================================================================
// PORTUGUESE LOCALE FORMATTER
// =============================================================================

// Portuguese sizing info mappings
var SizablePt = map[string]issues.SizingInfo{
	"string": {Unit: "caracteres", Verb: "ter"},
	"file":   {Unit: "bytes", Verb: "ter"},
	"array":  {Unit: "itens", Verb: "ter"},
	"slice":  {Unit: "itens", Verb: "ter"},
	"set":    {Unit: "itens", Verb: "ter"},
	"map":    {Unit: "entradas", Verb: "ter"},
}

// Portuguese format noun mappings
var FormatNounsPt = map[string]string{
	"regex":            "padrão",
	"email":            "endereço de e-mail",
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
	"datetime":         "data e hora ISO",
	"date":             "data ISO",
	"time":             "hora ISO",
	"duration":         "duração ISO",
	"ipv4":             "endereço IPv4",
	"ipv6":             "endereço IPv6",
	"mac":              "endereço MAC",
	"cidrv4":           "faixa de IPv4",
	"cidrv6":           "faixa de IPv6",
	"base64":           "texto codificado em base64",
	"base64url":        "URL codificada em base64",
	"json_string":      "texto JSON",
	"e164":             "número E.164",
	"jwt":              "JWT",
	"template_literal": "entrada",
}

// Portuguese type dictionary
var TypeDictionaryPt = map[string]string{
	"nan":       "NaN",
	"number":    "número",
	"array":     "array",
	"slice":     "array",
	"string":    "texto",
	"bool":      "booleano",
	"object":    "objeto",
	"map":       "mapa",
	"nil":       "nulo",
	"undefined": "indefinido",
	"function":  "função",
	"date":      "data",
	"file":      "arquivo",
	"set":       "conjunto",
}

// getSizingPt returns Portuguese sizing information for a given type
func getSizingPt(origin string) *issues.SizingInfo {
	if info, exists := SizablePt[origin]; exists {
		return &info
	}
	return nil
}

// getFormatNounPt returns Portuguese noun for a format name
func getFormatNounPt(format string) string {
	if noun, exists := FormatNounsPt[format]; exists {
		return noun
	}
	return format
}

// getTypeNamePt returns Portuguese type name
func getTypeNamePt(typeName string) string {
	if name, exists := TypeDictionaryPt[typeName]; exists {
		return name
	}
	return typeName
}

// formatPt provides Portuguese error messages
func formatPt(raw core.ZodRawIssue) string {
	code := raw.Code

	switch code {
	case core.InvalidType:
		expected := mapx.StringOr(raw.Properties, "expected", "")
		expected = getTypeNamePt(expected)
		received := issues.ParsedTypeToString(raw.Input)
		received = getTypeNamePt(received)
		return fmt.Sprintf("Tipo inválido: esperado %s, recebido %s", expected, received)

	case core.InvalidValue:
		values := mapx.AnySliceOr(raw.Properties, "values", nil)
		if len(values) == 0 {
			return "Valor inválido"
		}
		if len(values) == 1 {
			return fmt.Sprintf("Entrada inválida: esperado %s", issues.StringifyPrimitive(values[0]))
		}
		return fmt.Sprintf("Opção inválida: esperada uma das %s",
			issues.JoinValuesWithSeparator(values, "|"))

	case core.TooBig:
		return formatSizeConstraintPt(raw, false)

	case core.TooSmall:
		return formatSizeConstraintPt(raw, true)

	case core.InvalidFormat:
		format := mapx.StringOr(raw.Properties, "format", "")
		if format == "" {
			return "Formato inválido"
		}
		return formatStringValidationPt(raw, format)

	case core.NotMultipleOf:
		divisor := mapx.AnyOr(raw.Properties, "divisor", nil)
		if divisor == nil {
			return "Número inválido: deve ser múltiplo"
		}
		return fmt.Sprintf("Número inválido: deve ser múltiplo de %v", divisor)

	case core.UnrecognizedKeys:
		keys := mapx.StringsOr(raw.Properties, "keys", nil)
		if len(keys) == 0 {
			return "Chave desconhecida"
		}
		keyWord := "Chave desconhecida"
		if len(keys) > 1 {
			keyWord = "Chaves desconhecidas"
		}
		if keysAny, err := slicex.ToAny(keys); err == nil && !slicex.IsEmpty(keysAny) {
			keysJoined := issues.JoinValuesWithSeparator(keysAny, ", ")
			return fmt.Sprintf("%s: %s", keyWord, keysJoined)
		}
		return keyWord

	case core.InvalidKey:
		origin := mapx.StringOr(raw.Properties, "origin", "")
		if origin == "" {
			return "Chave inválida"
		}
		return fmt.Sprintf("Chave inválida em %s", origin)

	case core.InvalidUnion:
		return "Entrada inválida"

	case core.InvalidElement:
		origin := mapx.StringOr(raw.Properties, "origin", "")
		if origin == "" {
			return "Elemento inválido"
		}
		return fmt.Sprintf("Valor inválido em %s", origin)

	case core.MissingRequired:
		fieldName := mapx.StringOr(raw.Properties, "field_name", "")
		fieldType := mapx.StringOr(raw.Properties, "field_type", "campo")
		if fieldName == "" {
			return fmt.Sprintf("Falta %s obrigatório", fieldType)
		}
		return fmt.Sprintf("Falta %s obrigatório: %s", fieldType, fieldName)

	case core.TypeConversion:
		fromType := mapx.StringOr(raw.Properties, "from_type", "desconhecido")
		toType := mapx.StringOr(raw.Properties, "to_type", "desconhecido")
		return fmt.Sprintf("Falha na conversão de tipo: não é possível converter %s para %s", fromType, toType)

	case core.InvalidSchema:
		reason := mapx.StringOr(raw.Properties, "reason", "")
		if reason != "" {
			return fmt.Sprintf("Esquema inválido: %s", reason)
		}
		return "Definição de esquema inválida"

	case core.InvalidDiscriminator:
		field := mapx.StringOr(raw.Properties, "field", "discriminador")
		return fmt.Sprintf("Campo discriminador inválido ou ausente: %s", field)

	case core.IncompatibleTypes:
		conflictType := mapx.StringOr(raw.Properties, "conflict_type", "valores")
		return fmt.Sprintf("Não é possível mesclar %s: tipos incompatíveis", conflictType)

	case core.NilPointer:
		return "Ponteiro nulo encontrado"

	case core.Custom:
		message := mapx.StringOr(raw.Properties, "message", "")
		if message != "" {
			return message
		}
		return "Entrada inválida"

	default:
		return "Campo inválido"
	}
}

// formatSizeConstraintPt formats size constraint messages in Portuguese
func formatSizeConstraintPt(raw core.ZodRawIssue, isTooSmall bool) string {
	origin := mapx.StringOr(raw.Properties, "origin", "")
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
			return "Muito pequeno"
		}
		return "Muito grande"
	}

	inclusive := mapx.BoolOr(raw.Properties, "inclusive", true)
	sizing := getSizingPt(mapx.StringOr(raw.Properties, "origin", ""))
	thresholdStr := issues.FormatThreshold(threshold)

	// Portuguese comparison operators
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
			return fmt.Sprintf("Muito pequeno: esperado que %s tivesse %s%s %s", origin, adj, thresholdStr, sizing.Unit)
		}
		return fmt.Sprintf("Muito grande: esperado que %s tivesse %s%s %s", origin, adj, thresholdStr, sizing.Unit)
	}

	if isTooSmall {
		return fmt.Sprintf("Muito pequeno: esperado que %s fosse %s%s", origin, adj, thresholdStr)
	}
	return fmt.Sprintf("Muito grande: esperado que %s fosse %s%s", origin, adj, thresholdStr)
}

// formatStringValidationPt handles string format validation messages in Portuguese
func formatStringValidationPt(raw core.ZodRawIssue, format string) string {
	switch format {
	case "starts_with":
		prefix := mapx.StringOr(raw.Properties, "prefix", "")
		if prefix == "" {
			return "Texto inválido: deve começar com o prefixo especificado"
		}
		return fmt.Sprintf("Texto inválido: deve começar com \"%s\"", prefix)
	case "ends_with":
		suffix := mapx.StringOr(raw.Properties, "suffix", "")
		if suffix == "" {
			return "Texto inválido: deve terminar com o sufixo especificado"
		}
		return fmt.Sprintf("Texto inválido: deve terminar com \"%s\"", suffix)
	case "includes":
		includes := mapx.StringOr(raw.Properties, "includes", "")
		if includes == "" {
			return "Texto inválido: deve incluir a substring especificada"
		}
		return fmt.Sprintf("Texto inválido: deve incluir \"%s\"", includes)
	case "regex":
		pattern := mapx.StringOr(raw.Properties, "pattern", "")
		if pattern == "" {
			return "Texto inválido: deve corresponder ao padrão"
		}
		return fmt.Sprintf("Texto inválido: deve corresponder ao padrão %s", pattern)
	default:
		noun := getFormatNounPt(format)
		return fmt.Sprintf("%s inválido", noun)
	}
}

// Pt returns a ZodConfig configured for Portuguese locale.
func Pt() *core.ZodConfig {
	return &core.ZodConfig{
		LocaleError: formatPt,
	}
}

// FormatMessagePt formats a single issue using Portuguese locale
func FormatMessagePt(issue core.ZodRawIssue) string {
	return formatPt(issue)
}
