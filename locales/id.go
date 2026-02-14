package locales

import (
	"fmt"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// =============================================================================
// INDONESIAN LOCALE FORMATTER
// =============================================================================

// Indonesian sizing info mappings
var SizableId = map[string]issues.SizingInfo{
	"string": {Unit: "karakter", Verb: "memiliki"},
	"file":   {Unit: "byte", Verb: "memiliki"},
	"array":  {Unit: "item", Verb: "memiliki"},
	"slice":  {Unit: "item", Verb: "memiliki"},
	"set":    {Unit: "item", Verb: "memiliki"},
	"map":    {Unit: "entri", Verb: "memiliki"},
}

// Indonesian format noun mappings
var FormatNounsId = map[string]string{
	"regex":            "input",
	"email":            "alamat email",
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
	"datetime":         "tanggal dan waktu format ISO",
	"date":             "tanggal format ISO",
	"time":             "jam format ISO",
	"duration":         "durasi format ISO",
	"ipv4":             "alamat IPv4",
	"ipv6":             "alamat IPv6",
	"mac":              "alamat MAC",
	"cidrv4":           "rentang alamat IPv4",
	"cidrv6":           "rentang alamat IPv6",
	"base64":           "string dengan enkode base64",
	"base64url":        "string dengan enkode base64url",
	"json_string":      "string JSON",
	"e164":             "angka E.164",
	"jwt":              "JWT",
	"template_literal": "input",
}

// Indonesian type dictionary
var TypeDictionaryId = map[string]string{
	"nan":       "NaN",
	"number":    "angka",
	"array":     "array",
	"slice":     "array",
	"string":    "string",
	"bool":      "boolean",
	"object":    "objek",
	"map":       "peta",
	"nil":       "null",
	"undefined": "tidak terdefinisi",
	"function":  "fungsi",
	"date":      "tanggal",
	"file":      "file",
	"set":       "set",
}

// getSizingId returns Indonesian sizing information for a given type
func getSizingId(origin string) *issues.SizingInfo {
	if _, exists := SizableId[origin]; exists {
		return new(SizableId[origin])
	}
	return nil
}

// getFormatNounId returns Indonesian noun for a format name
func getFormatNounId(format string) string {
	if noun, exists := FormatNounsId[format]; exists {
		return noun
	}
	return format
}

// getTypeNameId returns Indonesian type name
func getTypeNameId(typeName string) string {
	if name, exists := TypeDictionaryId[typeName]; exists {
		return name
	}
	return typeName
}

// formatId provides Indonesian error messages
func formatId(raw core.ZodRawIssue) string {
	code := raw.Code

	switch code {
	case core.InvalidType:
		expected := mapx.StringOr(raw.Properties, "expected", "")
		expected = getTypeNameId(expected)
		received := issues.ParsedTypeToString(raw.Input)
		received = getTypeNameId(received)
		return fmt.Sprintf("Input tidak valid: diharapkan %s, diterima %s", expected, received)

	case core.InvalidValue:
		values := mapx.AnySliceOr(raw.Properties, "values", nil)
		if len(values) == 0 {
			return "Nilai tidak valid"
		}
		if len(values) == 1 {
			return fmt.Sprintf("Input tidak valid: diharapkan %s", issues.StringifyPrimitive(values[0]))
		}
		return fmt.Sprintf("Pilihan tidak valid: diharapkan salah satu dari %s",
			issues.JoinValuesWithSeparator(values, "|"))

	case core.TooBig:
		return formatSizeConstraintId(raw, false)

	case core.TooSmall:
		return formatSizeConstraintId(raw, true)

	case core.InvalidFormat:
		format := mapx.StringOr(raw.Properties, "format", "")
		if format == "" {
			return "Format tidak valid"
		}
		return formatStringValidationId(raw, format)

	case core.NotMultipleOf:
		divisor := mapx.AnyOr(raw.Properties, "divisor", nil)
		if divisor == nil {
			return "Angka tidak valid: harus kelipatan"
		}
		return fmt.Sprintf("Angka tidak valid: harus kelipatan dari %v", divisor)

	case core.UnrecognizedKeys:
		keys := mapx.StringsOr(raw.Properties, "keys", nil)
		if len(keys) == 0 {
			return "Kunci tidak dikenali"
		}
		if keysAny, err := slicex.ToAny(keys); err == nil && !slicex.IsEmpty(keysAny) {
			keysJoined := issues.JoinValuesWithSeparator(keysAny, ", ")
			return fmt.Sprintf("Kunci tidak dikenali: %s", keysJoined)
		}
		return "Kunci tidak dikenali"

	case core.InvalidKey:
		origin := mapx.StringOr(raw.Properties, "origin", "")
		if origin == "" {
			return "Kunci tidak valid"
		}
		return fmt.Sprintf("Kunci tidak valid di %s", origin)

	case core.InvalidUnion:
		return "Input tidak valid"

	case core.InvalidElement:
		origin := mapx.StringOr(raw.Properties, "origin", "")
		if origin == "" {
			return "Elemen tidak valid"
		}
		return fmt.Sprintf("Nilai tidak valid di %s", origin)

	case core.MissingRequired:
		fieldName := mapx.StringOr(raw.Properties, "field_name", "")
		fieldType := mapx.StringOr(raw.Properties, "field_type", "field")
		if fieldName == "" {
			return fmt.Sprintf("%s wajib tidak ada", fieldType)
		}
		return fmt.Sprintf("%s wajib tidak ada: %s", fieldType, fieldName)

	case core.TypeConversion:
		fromType := mapx.StringOr(raw.Properties, "from_type", "tidak diketahui")
		toType := mapx.StringOr(raw.Properties, "to_type", "tidak diketahui")
		return fmt.Sprintf("Konversi tipe gagal: tidak dapat mengkonversi %s ke %s", fromType, toType)

	case core.InvalidSchema:
		reason := mapx.StringOr(raw.Properties, "reason", "")
		if reason != "" {
			return fmt.Sprintf("Skema tidak valid: %s", reason)
		}
		return "Definisi skema tidak valid"

	case core.InvalidDiscriminator:
		field := mapx.StringOr(raw.Properties, "field", "diskriminator")
		return fmt.Sprintf("Field diskriminator tidak valid atau tidak ada: %s", field)

	case core.IncompatibleTypes:
		conflictType := mapx.StringOr(raw.Properties, "conflict_type", "nilai")
		return fmt.Sprintf("Tidak dapat menggabungkan %s: tipe tidak kompatibel", conflictType)

	case core.NilPointer:
		return "Pointer null terdeteksi"

	case core.Custom:
		message := mapx.StringOr(raw.Properties, "message", "")
		if message != "" {
			return message
		}
		return "Input tidak valid"

	default:
		return "Input tidak valid"
	}
}

// formatSizeConstraintId formats size constraint messages in Indonesian
func formatSizeConstraintId(raw core.ZodRawIssue, isTooSmall bool) string {
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
			return "Terlalu kecil"
		}
		return "Terlalu besar"
	}

	inclusive := mapx.BoolOr(raw.Properties, "inclusive", true)
	sizing := getSizingId(mapx.StringOr(raw.Properties, "origin", ""))
	thresholdStr := issues.FormatThreshold(threshold)

	// Indonesian comparison operators
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
			return fmt.Sprintf("Terlalu kecil: diharapkan %s memiliki %s%s %s", origin, adj, thresholdStr, sizing.Unit)
		}
		return fmt.Sprintf("Terlalu besar: diharapkan %s memiliki %s%s %s", origin, adj, thresholdStr, sizing.Unit)
	}

	if isTooSmall {
		return fmt.Sprintf("Terlalu kecil: diharapkan %s menjadi %s%s", origin, adj, thresholdStr)
	}
	return fmt.Sprintf("Terlalu besar: diharapkan %s menjadi %s%s", origin, adj, thresholdStr)
}

// formatStringValidationId handles string format validation messages in Indonesian
func formatStringValidationId(raw core.ZodRawIssue, format string) string {
	switch format {
	case "starts_with":
		prefix := mapx.StringOr(raw.Properties, "prefix", "")
		if prefix == "" {
			return "String tidak valid: harus dimulai dengan prefix yang ditentukan"
		}
		return fmt.Sprintf("String tidak valid: harus dimulai dengan \"%s\"", prefix)
	case "ends_with":
		suffix := mapx.StringOr(raw.Properties, "suffix", "")
		if suffix == "" {
			return "String tidak valid: harus berakhir dengan suffix yang ditentukan"
		}
		return fmt.Sprintf("String tidak valid: harus berakhir dengan \"%s\"", suffix)
	case "includes":
		includes := mapx.StringOr(raw.Properties, "includes", "")
		if includes == "" {
			return "String tidak valid: harus menyertakan substring yang ditentukan"
		}
		return fmt.Sprintf("String tidak valid: harus menyertakan \"%s\"", includes)
	case "regex":
		pattern := mapx.StringOr(raw.Properties, "pattern", "")
		if pattern == "" {
			return "String tidak valid: harus sesuai pola"
		}
		return fmt.Sprintf("String tidak valid: harus sesuai pola %s", pattern)
	default:
		noun := getFormatNounId(format)
		return fmt.Sprintf("%s tidak valid", noun)
	}
}

// Id returns a ZodConfig configured for Indonesian locale.
func Id() *core.ZodConfig {
	return &core.ZodConfig{
		LocaleError: formatId,
	}
}

// FormatMessageId formats a single issue using Indonesian locale
func FormatMessageId(issue core.ZodRawIssue) string {
	return formatId(issue)
}
