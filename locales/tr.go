package locales

import (
	"fmt"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// =============================================================================
// TURKISH LOCALE FORMATTER
// =============================================================================

// Turkish sizing info mappings
var SizableTr = map[string]issues.SizingInfo{
	"string": {Unit: "karakter", Verb: "olmalı"},
	"file":   {Unit: "bayt", Verb: "olmalı"},
	"array":  {Unit: "öğe", Verb: "olmalı"},
	"slice":  {Unit: "öğe", Verb: "olmalı"},
	"set":    {Unit: "öğe", Verb: "olmalı"},
	"map":    {Unit: "girdi", Verb: "olmalı"},
}

// Turkish format noun mappings
var FormatNounsTr = map[string]string{
	"regex":            "girdi",
	"email":            "e-posta adresi",
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
	"datetime":         "ISO tarih ve saat",
	"date":             "ISO tarih",
	"time":             "ISO saat",
	"duration":         "ISO süre",
	"ipv4":             "IPv4 adresi",
	"ipv6":             "IPv6 adresi",
	"mac":              "MAC adresi",
	"cidrv4":           "IPv4 aralığı",
	"cidrv6":           "IPv6 aralığı",
	"base64":           "base64 ile şifrelenmiş metin",
	"base64url":        "base64url ile şifrelenmiş metin",
	"json_string":      "JSON dizesi",
	"e164":             "E.164 sayısı",
	"jwt":              "JWT",
	"template_literal": "şablon dizesi",
}

// Turkish type dictionary
var TypeDictionaryTr = map[string]string{
	"nan":       "NaN",
	"number":    "sayı",
	"array":     "dizi",
	"slice":     "dizi",
	"string":    "metin",
	"bool":      "mantıksal",
	"object":    "nesne",
	"map":       "harita",
	"nil":       "boş",
	"undefined": "tanımsız",
	"function":  "fonksiyon",
	"date":      "tarih",
	"file":      "dosya",
	"set":       "küme",
}

// getSizingTr returns Turkish sizing information for a given type
func getSizingTr(origin string) *issues.SizingInfo {
	if _, exists := SizableTr[origin]; exists {
		return new(SizableTr[origin])
	}
	return nil
}

// getFormatNounTr returns Turkish noun for a format name
func getFormatNounTr(format string) string {
	if noun, exists := FormatNounsTr[format]; exists {
		return noun
	}
	return format
}

// getTypeNameTr returns Turkish type name
func getTypeNameTr(typeName string) string {
	if name, exists := TypeDictionaryTr[typeName]; exists {
		return name
	}
	return typeName
}

// formatTr provides Turkish error messages
func formatTr(raw core.ZodRawIssue) string {
	code := raw.Code

	switch code {
	case core.InvalidType:
		expected := mapx.StringOr(raw.Properties, "expected", "")
		expected = getTypeNameTr(expected)
		received := issues.ParsedTypeToString(raw.Input)
		received = getTypeNameTr(received)
		return fmt.Sprintf("Geçersiz değer: beklenen %s, alınan %s", expected, received)

	case core.InvalidValue:
		values := mapx.AnySliceOr(raw.Properties, "values", nil)
		if len(values) == 0 {
			return "Geçersiz değer"
		}
		if len(values) == 1 {
			return fmt.Sprintf("Geçersiz değer: beklenen %s", issues.StringifyPrimitive(values[0]))
		}
		return fmt.Sprintf("Geçersiz seçenek: aşağıdakilerden biri olmalı: %s",
			issues.JoinValuesWithSeparator(values, "|"))

	case core.TooBig:
		return formatSizeConstraintTr(raw, false)

	case core.TooSmall:
		return formatSizeConstraintTr(raw, true)

	case core.InvalidFormat:
		format := mapx.StringOr(raw.Properties, "format", "")
		if format == "" {
			return "Geçersiz format"
		}
		return formatStringValidationTr(raw, format)

	case core.NotMultipleOf:
		divisor := mapx.AnyOr(raw.Properties, "divisor", nil)
		if divisor == nil {
			return "Geçersiz sayı: tam bölünebilmeli"
		}
		return fmt.Sprintf("Geçersiz sayı: %v ile tam bölünebilmeli", divisor)

	case core.UnrecognizedKeys:
		keys := mapx.StringsOr(raw.Properties, "keys", nil)
		if len(keys) == 0 {
			return "Tanınmayan anahtar"
		}
		keyWord := "Tanınmayan anahtar"
		if len(keys) > 1 {
			keyWord = "Tanınmayan anahtarlar"
		}
		if keysAny, err := slicex.ToAny(keys); err == nil && !slicex.IsEmpty(keysAny) {
			keysJoined := issues.JoinValuesWithSeparator(keysAny, ", ")
			return fmt.Sprintf("%s: %s", keyWord, keysJoined)
		}
		return keyWord

	case core.InvalidKey:
		origin := mapx.StringOr(raw.Properties, "origin", "")
		if origin == "" {
			return "Geçersiz anahtar"
		}
		return fmt.Sprintf("%s içinde geçersiz anahtar", origin)

	case core.InvalidUnion:
		return "Geçersiz değer"

	case core.InvalidElement:
		origin := mapx.StringOr(raw.Properties, "origin", "")
		if origin == "" {
			return "Geçersiz öğe"
		}
		return fmt.Sprintf("%s içinde geçersiz değer", origin)

	case core.MissingRequired:
		fieldName := mapx.StringOr(raw.Properties, "field_name", "")
		fieldType := mapx.StringOr(raw.Properties, "field_type", "alan")
		if fieldName == "" {
			return fmt.Sprintf("Zorunlu %s eksik", fieldType)
		}
		return fmt.Sprintf("Zorunlu %s eksik: %s", fieldType, fieldName)

	case core.TypeConversion:
		fromType := mapx.StringOr(raw.Properties, "from_type", "bilinmeyen")
		toType := mapx.StringOr(raw.Properties, "to_type", "bilinmeyen")
		return fmt.Sprintf("Tür dönüştürme başarısız: %s türü %s türüne dönüştürülemiyor", fromType, toType)

	case core.InvalidSchema:
		reason := mapx.StringOr(raw.Properties, "reason", "")
		if reason != "" {
			return fmt.Sprintf("Geçersiz şema: %s", reason)
		}
		return "Geçersiz şema tanımı"

	case core.InvalidDiscriminator:
		field := mapx.StringOr(raw.Properties, "field", "ayrımcı")
		return fmt.Sprintf("Geçersiz veya eksik ayrımcı alan: %s", field)

	case core.IncompatibleTypes:
		conflictType := mapx.StringOr(raw.Properties, "conflict_type", "değerler")
		return fmt.Sprintf("%s birleştirilemiyor: uyumsuz türler", conflictType)

	case core.NilPointer:
		return "Boş işaretçi algılandı"

	case core.Custom:
		message := mapx.StringOr(raw.Properties, "message", "")
		if message != "" {
			return message
		}
		return "Geçersiz değer"

	default:
		return "Geçersiz değer"
	}
}

// formatSizeConstraintTr formats size constraint messages in Turkish
func formatSizeConstraintTr(raw core.ZodRawIssue, isTooSmall bool) string {
	origin := mapx.StringOr(raw.Properties, "origin", "")
	if origin == "" {
		origin = "değer"
	}

	var threshold any
	if isTooSmall {
		threshold = mapx.AnyOr(raw.Properties, "minimum", nil)
	} else {
		threshold = mapx.AnyOr(raw.Properties, "maximum", nil)
	}

	if threshold == nil {
		if isTooSmall {
			return "Çok küçük"
		}
		return "Çok büyük"
	}

	inclusive := mapx.BoolOr(raw.Properties, "inclusive", true)
	sizing := getSizingTr(mapx.StringOr(raw.Properties, "origin", ""))
	thresholdStr := issues.FormatThreshold(threshold)

	// Turkish comparison operators
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
			return fmt.Sprintf("Çok küçük: beklenen %s %s%s %s", origin, adj, thresholdStr, sizing.Unit)
		}
		return fmt.Sprintf("Çok büyük: beklenen %s %s%s %s", origin, adj, thresholdStr, sizing.Unit)
	}

	if isTooSmall {
		return fmt.Sprintf("Çok küçük: beklenen %s %s%s", origin, adj, thresholdStr)
	}
	return fmt.Sprintf("Çok büyük: beklenen %s %s%s", origin, adj, thresholdStr)
}

// formatStringValidationTr handles string format validation messages in Turkish
func formatStringValidationTr(raw core.ZodRawIssue, format string) string {
	switch format {
	case "starts_with":
		prefix := mapx.StringOr(raw.Properties, "prefix", "")
		if prefix == "" {
			return "Geçersiz metin: belirtilen önek ile başlamalı"
		}
		return fmt.Sprintf("Geçersiz metin: \"%s\" ile başlamalı", prefix)
	case "ends_with":
		suffix := mapx.StringOr(raw.Properties, "suffix", "")
		if suffix == "" {
			return "Geçersiz metin: belirtilen sonek ile bitmeli"
		}
		return fmt.Sprintf("Geçersiz metin: \"%s\" ile bitmeli", suffix)
	case "includes":
		includes := mapx.StringOr(raw.Properties, "includes", "")
		if includes == "" {
			return "Geçersiz metin: belirtilen alt dizeyi içermeli"
		}
		return fmt.Sprintf("Geçersiz metin: \"%s\" içermeli", includes)
	case "regex":
		pattern := mapx.StringOr(raw.Properties, "pattern", "")
		if pattern == "" {
			return "Geçersiz metin: desene uymalı"
		}
		return fmt.Sprintf("Geçersiz metin: %s desenine uymalı", pattern)
	default:
		noun := getFormatNounTr(format)
		return fmt.Sprintf("Geçersiz %s", noun)
	}
}

// Tr returns a ZodConfig configured for Turkish locale.
func Tr() *core.ZodConfig {
	return &core.ZodConfig{
		LocaleError: formatTr,
	}
}

// FormatMessageTr formats a single issue using Turkish locale
func FormatMessageTr(issue core.ZodRawIssue) string {
	return formatTr(issue)
}
