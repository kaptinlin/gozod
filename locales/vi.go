package locales

import (
	"fmt"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// =============================================================================
// VIETNAMESE LOCALE FORMATTER
// =============================================================================

// Vietnamese sizing info mappings
var SizableVi = map[string]issues.SizingInfo{
	"string": {Unit: "ký tự", Verb: "có"},
	"file":   {Unit: "byte", Verb: "có"},
	"array":  {Unit: "phần tử", Verb: "có"},
	"slice":  {Unit: "phần tử", Verb: "có"},
	"set":    {Unit: "phần tử", Verb: "có"},
	"map":    {Unit: "mục", Verb: "có"},
}

// Vietnamese format noun mappings
var FormatNounsVi = map[string]string{
	"regex":            "đầu vào",
	"email":            "địa chỉ email",
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
	"datetime":         "ngày giờ ISO",
	"date":             "ngày ISO",
	"time":             "giờ ISO",
	"duration":         "khoảng thời gian ISO",
	"ipv4":             "địa chỉ IPv4",
	"ipv6":             "địa chỉ IPv6",
	"mac":              "địa chỉ MAC",
	"cidrv4":           "dải IPv4",
	"cidrv6":           "dải IPv6",
	"base64":           "chuỗi mã hóa base64",
	"base64url":        "chuỗi mã hóa base64url",
	"json_string":      "chuỗi JSON",
	"e164":             "số E.164",
	"jwt":              "JWT",
	"template_literal": "đầu vào",
}

// Vietnamese type dictionary
var TypeDictionaryVi = map[string]string{
	"nan":       "NaN",
	"number":    "số",
	"array":     "mảng",
	"slice":     "mảng",
	"string":    "chuỗi",
	"bool":      "boolean",
	"object":    "đối tượng",
	"map":       "bản đồ",
	"nil":       "null",
	"undefined": "không xác định",
	"function":  "hàm",
	"date":      "ngày",
	"file":      "tệp",
	"set":       "tập hợp",
}

// getSizingVi returns Vietnamese sizing information for a given type
func getSizingVi(origin string) *issues.SizingInfo {
	if _, exists := SizableVi[origin]; exists {
		return new(SizableVi[origin])
	}
	return nil
}

// getFormatNounVi returns Vietnamese noun for a format name
func getFormatNounVi(format string) string {
	if noun, exists := FormatNounsVi[format]; exists {
		return noun
	}
	return format
}

// getTypeNameVi returns Vietnamese type name
func getTypeNameVi(typeName string) string {
	if name, exists := TypeDictionaryVi[typeName]; exists {
		return name
	}
	return typeName
}

// formatVi provides Vietnamese error messages
func formatVi(raw core.ZodRawIssue) string {
	code := raw.Code

	switch code {
	case core.InvalidType:
		expected := mapx.StringOr(raw.Properties, "expected", "")
		expected = getTypeNameVi(expected)
		received := issues.ParsedTypeToString(raw.Input)
		received = getTypeNameVi(received)
		return fmt.Sprintf("Đầu vào không hợp lệ: mong đợi %s, nhận được %s", expected, received)

	case core.InvalidValue:
		values := mapx.AnySliceOr(raw.Properties, "values", nil)
		if len(values) == 0 {
			return "Giá trị không hợp lệ"
		}
		if len(values) == 1 {
			return fmt.Sprintf("Đầu vào không hợp lệ: mong đợi %s", issues.StringifyPrimitive(values[0]))
		}
		return fmt.Sprintf("Tùy chọn không hợp lệ: mong đợi một trong các giá trị %s",
			issues.JoinValuesWithSeparator(values, "|"))

	case core.TooBig:
		return formatSizeConstraintVi(raw, false)

	case core.TooSmall:
		return formatSizeConstraintVi(raw, true)

	case core.InvalidFormat:
		format := mapx.StringOr(raw.Properties, "format", "")
		if format == "" {
			return "Định dạng không hợp lệ"
		}
		return formatStringValidationVi(raw, format)

	case core.NotMultipleOf:
		divisor := mapx.AnyOr(raw.Properties, "divisor", nil)
		if divisor == nil {
			return "Số không hợp lệ: phải là bội số"
		}
		return fmt.Sprintf("Số không hợp lệ: phải là bội số của %v", divisor)

	case core.UnrecognizedKeys:
		keys := mapx.StringsOr(raw.Properties, "keys", nil)
		if len(keys) == 0 {
			return "Khóa không được nhận dạng"
		}
		if keysAny, err := slicex.ToAny(keys); err == nil && !slicex.IsEmpty(keysAny) {
			keysJoined := issues.JoinValuesWithSeparator(keysAny, ", ")
			return fmt.Sprintf("Khóa không được nhận dạng: %s", keysJoined)
		}
		return "Khóa không được nhận dạng"

	case core.InvalidKey:
		origin := mapx.StringOr(raw.Properties, "origin", "")
		if origin == "" {
			return "Khóa không hợp lệ"
		}
		return fmt.Sprintf("Khóa không hợp lệ trong %s", origin)

	case core.InvalidUnion:
		return "Đầu vào không hợp lệ"

	case core.InvalidElement:
		origin := mapx.StringOr(raw.Properties, "origin", "")
		if origin == "" {
			return "Phần tử không hợp lệ"
		}
		return fmt.Sprintf("Giá trị không hợp lệ trong %s", origin)

	case core.MissingRequired:
		fieldName := mapx.StringOr(raw.Properties, "field_name", "")
		fieldType := mapx.StringOr(raw.Properties, "field_type", "trường")
		if fieldName == "" {
			return fmt.Sprintf("Thiếu %s bắt buộc", fieldType)
		}
		return fmt.Sprintf("Thiếu %s bắt buộc: %s", fieldType, fieldName)

	case core.TypeConversion:
		fromType := mapx.StringOr(raw.Properties, "from_type", "không xác định")
		toType := mapx.StringOr(raw.Properties, "to_type", "không xác định")
		return fmt.Sprintf("Chuyển đổi kiểu thất bại: không thể chuyển %s sang %s", fromType, toType)

	case core.InvalidSchema:
		reason := mapx.StringOr(raw.Properties, "reason", "")
		if reason != "" {
			return fmt.Sprintf("Lược đồ không hợp lệ: %s", reason)
		}
		return "Định nghĩa lược đồ không hợp lệ"

	case core.InvalidDiscriminator:
		field := mapx.StringOr(raw.Properties, "field", "bộ phân biệt")
		return fmt.Sprintf("Trường phân biệt không hợp lệ hoặc bị thiếu: %s", field)

	case core.IncompatibleTypes:
		conflictType := mapx.StringOr(raw.Properties, "conflict_type", "giá trị")
		return fmt.Sprintf("Không thể hợp nhất %s: kiểu không tương thích", conflictType)

	case core.NilPointer:
		return "Phát hiện con trỏ null"

	case core.Custom:
		message := mapx.StringOr(raw.Properties, "message", "")
		if message != "" {
			return message
		}
		return "Đầu vào không hợp lệ"

	default:
		return "Đầu vào không hợp lệ"
	}
}

// formatSizeConstraintVi formats size constraint messages in Vietnamese
func formatSizeConstraintVi(raw core.ZodRawIssue, isTooSmall bool) string {
	origin := mapx.StringOr(raw.Properties, "origin", "")
	if origin == "" {
		origin = "giá trị"
	}

	var threshold any
	if isTooSmall {
		threshold = mapx.AnyOr(raw.Properties, "minimum", nil)
	} else {
		threshold = mapx.AnyOr(raw.Properties, "maximum", nil)
	}

	if threshold == nil {
		if isTooSmall {
			return "Quá nhỏ"
		}
		return "Quá lớn"
	}

	inclusive := mapx.BoolOr(raw.Properties, "inclusive", true)
	sizing := getSizingVi(mapx.StringOr(raw.Properties, "origin", ""))
	thresholdStr := issues.FormatThreshold(threshold)

	// Vietnamese comparison operators
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
			return fmt.Sprintf("Quá nhỏ: mong đợi %s %s %s%s %s", origin, sizing.Verb, adj, thresholdStr, sizing.Unit)
		}
		return fmt.Sprintf("Quá lớn: mong đợi %s %s %s%s %s", origin, sizing.Verb, adj, thresholdStr, sizing.Unit)
	}

	if isTooSmall {
		return fmt.Sprintf("Quá nhỏ: mong đợi %s %s%s", origin, adj, thresholdStr)
	}
	return fmt.Sprintf("Quá lớn: mong đợi %s %s%s", origin, adj, thresholdStr)
}

// formatStringValidationVi handles string format validation messages in Vietnamese
func formatStringValidationVi(raw core.ZodRawIssue, format string) string {
	switch format {
	case "starts_with":
		prefix := mapx.StringOr(raw.Properties, "prefix", "")
		if prefix == "" {
			return "Chuỗi không hợp lệ: phải bắt đầu bằng tiền tố được chỉ định"
		}
		return fmt.Sprintf("Chuỗi không hợp lệ: phải bắt đầu bằng \"%s\"", prefix)
	case "ends_with":
		suffix := mapx.StringOr(raw.Properties, "suffix", "")
		if suffix == "" {
			return "Chuỗi không hợp lệ: phải kết thúc bằng hậu tố được chỉ định"
		}
		return fmt.Sprintf("Chuỗi không hợp lệ: phải kết thúc bằng \"%s\"", suffix)
	case "includes":
		includes := mapx.StringOr(raw.Properties, "includes", "")
		if includes == "" {
			return "Chuỗi không hợp lệ: phải bao gồm chuỗi con được chỉ định"
		}
		return fmt.Sprintf("Chuỗi không hợp lệ: phải bao gồm \"%s\"", includes)
	case "regex":
		pattern := mapx.StringOr(raw.Properties, "pattern", "")
		if pattern == "" {
			return "Chuỗi không hợp lệ: phải khớp với mẫu"
		}
		return fmt.Sprintf("Chuỗi không hợp lệ: phải khớp với mẫu %s", pattern)
	default:
		noun := getFormatNounVi(format)
		return fmt.Sprintf("%s không hợp lệ", noun)
	}
}

// Vi returns a ZodConfig configured for Vietnamese locale.
func Vi() *core.ZodConfig {
	return &core.ZodConfig{
		LocaleError: formatVi,
	}
}

// FormatMessageVi formats a single issue using Vietnamese locale
func FormatMessageVi(issue core.ZodRawIssue) string {
	return formatVi(issue)
}
