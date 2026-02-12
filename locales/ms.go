package locales

import (
	"fmt"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// =============================================================================
// MALAY LOCALE FORMATTER
// =============================================================================

// Malay sizing info mappings
var SizableMs = map[string]issues.SizingInfo{
	"string": {Unit: "aksara", Verb: "mempunyai"},
	"file":   {Unit: "bait", Verb: "mempunyai"},
	"array":  {Unit: "elemen", Verb: "mempunyai"},
	"slice":  {Unit: "elemen", Verb: "mempunyai"},
	"set":    {Unit: "elemen", Verb: "mempunyai"},
	"map":    {Unit: "entri", Verb: "mempunyai"},
}

// Malay format noun mappings
var FormatNounsMs = map[string]string{
	"regex":            "input",
	"email":            "alamat e-mel",
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
	"datetime":         "tarikh masa ISO",
	"date":             "tarikh ISO",
	"time":             "masa ISO",
	"duration":         "tempoh ISO",
	"ipv4":             "alamat IPv4",
	"ipv6":             "alamat IPv6",
	"mac":              "alamat MAC",
	"cidrv4":           "julat IPv4",
	"cidrv6":           "julat IPv6",
	"base64":           "string dikodkan base64",
	"base64url":        "string dikodkan base64url",
	"json_string":      "string JSON",
	"e164":             "nombor E.164",
	"jwt":              "JWT",
	"template_literal": "input",
}

// Malay type dictionary
var TypeDictionaryMs = map[string]string{
	"nan":       "NaN",
	"number":    "nombor",
	"array":     "senarai",
	"slice":     "senarai",
	"string":    "rentetan",
	"bool":      "boolean",
	"object":    "objek",
	"map":       "peta",
	"nil":       "null",
	"undefined": "tidak ditakrifkan",
	"function":  "fungsi",
	"date":      "tarikh",
	"file":      "fail",
	"set":       "set",
}

// getSizingMs returns Malay sizing information for a given type
func getSizingMs(origin string) *issues.SizingInfo {
	if info, exists := SizableMs[origin]; exists {
		return &info
	}
	return nil
}

// getFormatNounMs returns Malay noun for a format name
func getFormatNounMs(format string) string {
	if noun, exists := FormatNounsMs[format]; exists {
		return noun
	}
	return format
}

// getTypeNameMs returns Malay type name
func getTypeNameMs(typeName string) string {
	if name, exists := TypeDictionaryMs[typeName]; exists {
		return name
	}
	return typeName
}

// formatMs provides Malay error messages
func formatMs(raw core.ZodRawIssue) string {
	code := raw.Code

	switch code {
	case core.InvalidType:
		expected := mapx.StringOr(raw.Properties, "expected", "")
		expected = getTypeNameMs(expected)
		received := issues.ParsedTypeToString(raw.Input)
		received = getTypeNameMs(received)
		return fmt.Sprintf("Input tidak sah: dijangka %s, diterima %s", expected, received)

	case core.InvalidValue:
		values := mapx.AnySliceOr(raw.Properties, "values", nil)
		if len(values) == 0 {
			return "Nilai tidak sah"
		}
		if len(values) == 1 {
			return fmt.Sprintf("Input tidak sah: dijangka %s", issues.StringifyPrimitive(values[0]))
		}
		return fmt.Sprintf("Pilihan tidak sah: dijangka salah satu daripada %s",
			issues.JoinValuesWithSeparator(values, "|"))

	case core.TooBig:
		return formatSizeConstraintMs(raw, false)

	case core.TooSmall:
		return formatSizeConstraintMs(raw, true)

	case core.InvalidFormat:
		format := mapx.StringOr(raw.Properties, "format", "")
		if format == "" {
			return "Format tidak sah"
		}
		return formatStringValidationMs(raw, format)

	case core.NotMultipleOf:
		divisor := mapx.AnyOr(raw.Properties, "divisor", nil)
		if divisor == nil {
			return "Nombor tidak sah: perlu gandaan"
		}
		return fmt.Sprintf("Nombor tidak sah: perlu gandaan %v", divisor)

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
			return "Kunci tidak sah"
		}
		return fmt.Sprintf("Kunci tidak sah dalam %s", origin)

	case core.InvalidUnion:
		return "Input tidak sah"

	case core.InvalidElement:
		origin := mapx.StringOr(raw.Properties, "origin", "")
		if origin == "" {
			return "Elemen tidak sah"
		}
		return fmt.Sprintf("Nilai tidak sah dalam %s", origin)

	case core.MissingRequired:
		fieldName := mapx.StringOr(raw.Properties, "field_name", "")
		fieldType := mapx.StringOr(raw.Properties, "field_type", "medan")
		if fieldName == "" {
			return fmt.Sprintf("%s yang diperlukan tiada", fieldType)
		}
		return fmt.Sprintf("%s yang diperlukan tiada: %s", fieldType, fieldName)

	case core.TypeConversion:
		fromType := mapx.StringOr(raw.Properties, "from_type", "tidak diketahui")
		toType := mapx.StringOr(raw.Properties, "to_type", "tidak diketahui")
		return fmt.Sprintf("Penukaran jenis gagal: tidak dapat menukar %s kepada %s", fromType, toType)

	case core.InvalidSchema:
		reason := mapx.StringOr(raw.Properties, "reason", "")
		if reason != "" {
			return fmt.Sprintf("Skema tidak sah: %s", reason)
		}
		return "Definisi skema tidak sah"

	case core.InvalidDiscriminator:
		field := mapx.StringOr(raw.Properties, "field", "diskriminator")
		return fmt.Sprintf("Medan diskriminator tidak sah atau tiada: %s", field)

	case core.IncompatibleTypes:
		conflictType := mapx.StringOr(raw.Properties, "conflict_type", "nilai")
		return fmt.Sprintf("Tidak dapat menggabungkan %s: jenis tidak serasi", conflictType)

	case core.NilPointer:
		return "Penunjuk null dikesan"

	case core.Custom:
		message := mapx.StringOr(raw.Properties, "message", "")
		if message != "" {
			return message
		}
		return "Input tidak sah"

	default:
		return "Input tidak sah"
	}
}

// formatSizeConstraintMs formats size constraint messages in Malay
func formatSizeConstraintMs(raw core.ZodRawIssue, isTooSmall bool) string {
	origin := mapx.StringOr(raw.Properties, "origin", "")
	if origin == "" {
		origin = "nilai"
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
	sizing := getSizingMs(mapx.StringOr(raw.Properties, "origin", ""))
	thresholdStr := issues.FormatThreshold(threshold)

	// Malay comparison operators
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
			return fmt.Sprintf("Terlalu kecil: dijangka %s %s %s%s %s", origin, sizing.Verb, adj, thresholdStr, sizing.Unit)
		}
		return fmt.Sprintf("Terlalu besar: dijangka %s %s %s%s %s", origin, sizing.Verb, adj, thresholdStr, sizing.Unit)
	}

	if isTooSmall {
		return fmt.Sprintf("Terlalu kecil: dijangka %s adalah %s%s", origin, adj, thresholdStr)
	}
	return fmt.Sprintf("Terlalu besar: dijangka %s adalah %s%s", origin, adj, thresholdStr)
}

// formatStringValidationMs handles string format validation messages in Malay
func formatStringValidationMs(raw core.ZodRawIssue, format string) string {
	switch format {
	case "starts_with":
		prefix := mapx.StringOr(raw.Properties, "prefix", "")
		if prefix == "" {
			return "String tidak sah: mesti bermula dengan awalan tertentu"
		}
		return fmt.Sprintf("String tidak sah: mesti bermula dengan \"%s\"", prefix)
	case "ends_with":
		suffix := mapx.StringOr(raw.Properties, "suffix", "")
		if suffix == "" {
			return "String tidak sah: mesti berakhir dengan akhiran tertentu"
		}
		return fmt.Sprintf("String tidak sah: mesti berakhir dengan \"%s\"", suffix)
	case "includes":
		includes := mapx.StringOr(raw.Properties, "includes", "")
		if includes == "" {
			return "String tidak sah: mesti mengandungi substring tertentu"
		}
		return fmt.Sprintf("String tidak sah: mesti mengandungi \"%s\"", includes)
	case "regex":
		pattern := mapx.StringOr(raw.Properties, "pattern", "")
		if pattern == "" {
			return "String tidak sah: mesti sepadan dengan corak"
		}
		return fmt.Sprintf("String tidak sah: mesti sepadan dengan corak %s", pattern)
	default:
		noun := getFormatNounMs(format)
		return fmt.Sprintf("%s tidak sah", noun)
	}
}

// Ms returns a ZodConfig configured for Malay locale.
func Ms() *core.ZodConfig {
	return &core.ZodConfig{
		LocaleError: formatMs,
	}
}

// FormatMessageMs formats a single issue using Malay locale
func FormatMessageMs(issue core.ZodRawIssue) string {
	return formatMs(issue)
}
