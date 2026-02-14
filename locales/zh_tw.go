package locales

import (
	"fmt"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// =============================================================================
// TRADITIONAL CHINESE (TAIWAN) LOCALE FORMATTER
// =============================================================================

// Traditional Chinese sizing info mappings
var SizableZhTw = map[string]issues.SizingInfo{
	"string": {Unit: "字元", Verb: "擁有"},
	"file":   {Unit: "位元組", Verb: "擁有"},
	"array":  {Unit: "項目", Verb: "擁有"},
	"slice":  {Unit: "項目", Verb: "擁有"},
	"set":    {Unit: "項目", Verb: "擁有"},
	"map":    {Unit: "項目", Verb: "擁有"},
}

// Traditional Chinese format noun mappings
var FormatNounsZhTw = map[string]string{
	"regex":            "輸入",
	"email":            "郵件地址",
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
	"datetime":         "ISO 日期時間",
	"date":             "ISO 日期",
	"time":             "ISO 時間",
	"duration":         "ISO 期間",
	"ipv4":             "IPv4 位址",
	"ipv6":             "IPv6 位址",
	"mac":              "MAC 位址",
	"cidrv4":           "IPv4 範圍",
	"cidrv6":           "IPv6 範圍",
	"base64":           "base64 編碼字串",
	"base64url":        "base64url 編碼字串",
	"json_string":      "JSON 字串",
	"e164":             "E.164 數值",
	"jwt":              "JWT",
	"template_literal": "輸入",
}

// Traditional Chinese type dictionary
var TypeDictionaryZhTw = map[string]string{
	"nan":       "NaN",
	"number":    "數字",
	"array":     "陣列",
	"slice":     "陣列",
	"string":    "字串",
	"bool":      "布林值",
	"object":    "物件",
	"map":       "對應表",
	"nil":       "null",
	"undefined": "undefined",
	"function":  "函式",
	"date":      "日期",
	"file":      "檔案",
	"set":       "集合",
}

// getSizingZhTw returns Traditional Chinese sizing information for a given type
func getSizingZhTw(origin string) *issues.SizingInfo {
	if _, exists := SizableZhTw[origin]; exists {
		return new(SizableZhTw[origin])
	}
	return nil
}

// getFormatNounZhTw returns Traditional Chinese noun for a format name
func getFormatNounZhTw(format string) string {
	if noun, exists := FormatNounsZhTw[format]; exists {
		return noun
	}
	return format
}

// getTypeNameZhTw returns Traditional Chinese type name
func getTypeNameZhTw(typeName string) string {
	if name, exists := TypeDictionaryZhTw[typeName]; exists {
		return name
	}
	return typeName
}

// formatZhTw provides Traditional Chinese error messages
func formatZhTw(raw core.ZodRawIssue) string {
	code := raw.Code

	switch code {
	case core.InvalidType:
		expected := mapx.StringOr(raw.Properties, "expected", "")
		expected = getTypeNameZhTw(expected)
		received := issues.ParsedTypeToString(raw.Input)
		received = getTypeNameZhTw(received)
		return fmt.Sprintf("無效的輸入值：預期為 %s，但收到 %s", expected, received)

	case core.InvalidValue:
		values := mapx.AnySliceOr(raw.Properties, "values", nil)
		if len(values) == 0 {
			return "無效的數值"
		}
		if len(values) == 1 {
			return fmt.Sprintf("無效的輸入值：預期為 %s", issues.StringifyPrimitive(values[0]))
		}
		return fmt.Sprintf("無效的選項：預期為以下其中之一 %s",
			issues.JoinValuesWithSeparator(values, "|"))

	case core.TooBig:
		return formatSizeConstraintZhTw(raw, false)

	case core.TooSmall:
		return formatSizeConstraintZhTw(raw, true)

	case core.InvalidFormat:
		format := mapx.StringOr(raw.Properties, "format", "")
		if format == "" {
			return "無效的格式"
		}
		return formatStringValidationZhTw(raw, format)

	case core.NotMultipleOf:
		divisor := mapx.AnyOr(raw.Properties, "divisor", nil)
		if divisor == nil {
			return "無效的數字：必須為倍數"
		}
		return fmt.Sprintf("無效的數字：必須為 %v 的倍數", divisor)

	case core.UnrecognizedKeys:
		keys := mapx.StringsOr(raw.Properties, "keys", nil)
		if len(keys) == 0 {
			return "無法識別的鍵值"
		}
		if keysAny, err := slicex.ToAny(keys); err == nil && !slicex.IsEmpty(keysAny) {
			keysJoined := issues.JoinValuesWithSeparator(keysAny, "、")
			return fmt.Sprintf("無法識別的鍵值：%s", keysJoined)
		}
		return "無法識別的鍵值"

	case core.InvalidKey:
		origin := mapx.StringOr(raw.Properties, "origin", "")
		if origin == "" {
			return "無效的鍵值"
		}
		return fmt.Sprintf("%s 中有無效的鍵值", origin)

	case core.InvalidUnion:
		return "無效的輸入值"

	case core.InvalidElement:
		origin := mapx.StringOr(raw.Properties, "origin", "")
		if origin == "" {
			return "無效的元素"
		}
		return fmt.Sprintf("%s 中有無效的值", origin)

	case core.MissingRequired:
		fieldName := mapx.StringOr(raw.Properties, "field_name", "")
		fieldType := mapx.StringOr(raw.Properties, "field_type", "欄位")
		if fieldName == "" {
			return fmt.Sprintf("缺少必填的%s", fieldType)
		}
		return fmt.Sprintf("缺少必填的%s：%s", fieldType, fieldName)

	case core.TypeConversion:
		fromType := mapx.StringOr(raw.Properties, "from_type", "未知")
		toType := mapx.StringOr(raw.Properties, "to_type", "未知")
		return fmt.Sprintf("型別轉換失敗：無法將 %s 轉換為 %s", fromType, toType)

	case core.InvalidSchema:
		reason := mapx.StringOr(raw.Properties, "reason", "")
		if reason != "" {
			return fmt.Sprintf("無效的結構描述：%s", reason)
		}
		return "無效的結構描述定義"

	case core.InvalidDiscriminator:
		field := mapx.StringOr(raw.Properties, "field", "辨別器")
		return fmt.Sprintf("無效或缺少辨別器欄位：%s", field)

	case core.IncompatibleTypes:
		conflictType := mapx.StringOr(raw.Properties, "conflict_type", "數值")
		return fmt.Sprintf("無法合併 %s：型別不相容", conflictType)

	case core.NilPointer:
		return "偵測到空指標"

	case core.Custom:
		message := mapx.StringOr(raw.Properties, "message", "")
		if message != "" {
			return message
		}
		return "無效的輸入值"

	default:
		return "無效的輸入值"
	}
}

// formatSizeConstraintZhTw formats size constraint messages in Traditional Chinese
func formatSizeConstraintZhTw(raw core.ZodRawIssue, isTooSmall bool) string {
	origin := mapx.StringOr(raw.Properties, "origin", "")
	if origin == "" {
		origin = "值"
	}

	var threshold any
	if isTooSmall {
		threshold = mapx.AnyOr(raw.Properties, "minimum", nil)
	} else {
		threshold = mapx.AnyOr(raw.Properties, "maximum", nil)
	}

	if threshold == nil {
		if isTooSmall {
			return "數值過小"
		}
		return "數值過大"
	}

	inclusive := mapx.BoolOr(raw.Properties, "inclusive", true)
	sizing := getSizingZhTw(mapx.StringOr(raw.Properties, "origin", ""))
	thresholdStr := issues.FormatThreshold(threshold)

	// Traditional Chinese comparison operators
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
			return fmt.Sprintf("數值過小：預期 %s 應為 %s%s %s", origin, adj, thresholdStr, sizing.Unit)
		}
		return fmt.Sprintf("數值過大：預期 %s 應為 %s%s %s", origin, adj, thresholdStr, sizing.Unit)
	}

	if isTooSmall {
		return fmt.Sprintf("數值過小：預期 %s 應為 %s%s", origin, adj, thresholdStr)
	}
	return fmt.Sprintf("數值過大：預期 %s 應為 %s%s", origin, adj, thresholdStr)
}

// formatStringValidationZhTw handles string format validation messages in Traditional Chinese
func formatStringValidationZhTw(raw core.ZodRawIssue, format string) string {
	switch format {
	case "starts_with":
		prefix := mapx.StringOr(raw.Properties, "prefix", "")
		if prefix == "" {
			return "無效的字串：必須以指定前綴開頭"
		}
		return fmt.Sprintf("無效的字串：必須以 \"%s\" 開頭", prefix)
	case "ends_with":
		suffix := mapx.StringOr(raw.Properties, "suffix", "")
		if suffix == "" {
			return "無效的字串：必須以指定後綴結尾"
		}
		return fmt.Sprintf("無效的字串：必須以 \"%s\" 結尾", suffix)
	case "includes":
		includes := mapx.StringOr(raw.Properties, "includes", "")
		if includes == "" {
			return "無效的字串：必須包含指定子字串"
		}
		return fmt.Sprintf("無效的字串：必須包含 \"%s\"", includes)
	case "regex":
		pattern := mapx.StringOr(raw.Properties, "pattern", "")
		if pattern == "" {
			return "無效的字串：必須符合格式"
		}
		return fmt.Sprintf("無效的字串：必須符合格式 %s", pattern)
	default:
		noun := getFormatNounZhTw(format)
		return fmt.Sprintf("無效的 %s", noun)
	}
}

// ZhTw returns a ZodConfig configured for Traditional Chinese (Taiwan) locale.
func ZhTw() *core.ZodConfig {
	return &core.ZodConfig{
		LocaleError: formatZhTw,
	}
}

// FormatMessageZhTw formats a single issue using Traditional Chinese locale
func FormatMessageZhTw(issue core.ZodRawIssue) string {
	return formatZhTw(issue)
}
