package locales

import (
	"fmt"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/reflectx"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// =============================================================================
// CHINESE LOCALE FORMATTER
// =============================================================================

// getParsedTypeZh returns Chinese type name for a given input
func getParsedTypeZh(input any) string {
	// Use reflectx.ParsedType directly for comprehensive type detection
	parsedType := reflectx.ParsedType(input)

	//nolint:exhaustive
	switch parsedType {
	case core.ParsedTypeNaN:
		return "非数字(NaN)"
	case core.ParsedTypeNil:
		return "空值(null)"
	case core.ParsedTypeSlice, core.ParsedTypeArray:
		return "数组"
	case core.ParsedTypeMap, core.ParsedTypeObject, core.ParsedTypeStruct:
		return "对象"
	case core.ParsedTypeFloat, core.ParsedTypeNumber:
		return "数字"
	case core.ParsedTypeBigint:
		return "大整数"
	case core.ParsedTypeBool:
		return "布尔值"
	case core.ParsedTypeString:
		return "字符串"
	case core.ParsedTypeFunction:
		return "函数"
	case core.ParsedTypeFile:
		return "文件"
	case core.ParsedTypeDate:
		return "日期"
	case core.ParsedTypeComplex:
		return "复数"
	case core.ParsedTypeEnum:
		return "枚举"
	default:
		return "未知类型"
	}
}

// Chinese sizing info mappings - updated with consistent terminology
var SizableZh = map[string]issues.SizingInfo{
	"string": {Unit: "字符", Verb: "包含"},
	"file":   {Unit: "字节", Verb: "包含"},
	"array":  {Unit: "项", Verb: "包含"},
	"slice":  {Unit: "项", Verb: "包含"},
	"set":    {Unit: "项", Verb: "包含"},
	"object": {Unit: "键", Verb: "包含"},
	"map":    {Unit: "键", Verb: "包含"},
}

// getSizingZh returns Chinese sizing information for a given type
func getSizingZh(origin string) *issues.SizingInfo {
	if info, exists := SizableZh[origin]; exists {
		return &info
	}
	return nil
}

// Chinese format noun mappings - comprehensive coverage for Go and web types
var FormatNounsZh = map[string]string{
	"regex":            "输入",
	"email":            "电子邮件",
	"url":              "URL",
	"emoji":            "表情符号",
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
	"datetime":         "ISO日期时间",
	"date":             "ISO日期",
	"time":             "ISO时间",
	"duration":         "ISO时长",
	"ipv4":             "IPv4地址",
	"ipv6":             "IPv6地址",
	"cidrv4":           "IPv4网段",
	"cidrv6":           "IPv6网段",
	"base64":           "base64编码字符串",
	"base64url":        "base64url编码字符串",
	"json_string":      "JSON字符串",
	"e164":             "E.164号码",
	"jwt":              "JWT",
	"template_literal": "输入",
	// Go-specific formats
	"int8":       "8位整数",
	"int16":      "16位整数",
	"int32":      "32位整数",
	"int64":      "64位整数",
	"uint8":      "8位无符号整数",
	"uint16":     "16位无符号整数",
	"uint32":     "32位无符号整数",
	"uint64":     "64位无符号整数",
	"float32":    "32位浮点数",
	"float64":    "64位浮点数",
	"complex64":  "64位复数",
	"complex128": "128位复数",
}

// getFormatNounZh returns Chinese noun for a format name
func getFormatNounZh(format string) string {
	if noun, exists := FormatNounsZh[format]; exists {
		return noun
	}
	return format
}

// joinValuesZhOr formats value array with Chinese-style conjunction
// Maintains Chinese grammar patterns for value enumeration
// Enhanced with slicex functionality
func joinValuesZhOr(values []any) string {
	if len(values) == 0 || slicex.IsEmpty(values) {
		return ""
	}

	// Use slicex to handle various slice types and get unique values
	if uniqueValues, err := slicex.Unique(values); err == nil {
		if typedValues, err := slicex.ToAny(uniqueValues); err == nil {
			values = typedValues
		}
	}

	quoted := make([]string, len(values))
	for i, v := range values {
		quoted[i] = issues.StringifyPrimitive(v)
	}

	if len(quoted) == 1 {
		return quoted[0]
	}
	if len(quoted) == 2 {
		return quoted[0] + " 或 " + quoted[1]
	}

	last := quoted[len(quoted)-1]
	others := slicex.Join(quoted[:len(quoted)-1], "、")
	return others + " 或 " + last
}

// formatZhCN provides Chinese error messages
// Updated to match TypeScript Zod v4 structure with Chinese localization
func formatZhCN(raw core.ZodRawIssue) string {
	code := raw.Code

	switch code {
	case core.InvalidType:
		expected := mapx.GetStringDefault(raw.Properties, "expected", "")
		received := getParsedTypeZh(raw.Input)
		return fmt.Sprintf("无效输入：期望 %s，实际接收 %s", expected, received)

	case core.InvalidValue:
		values := mapx.GetAnySliceDefault(raw.Properties, "values", nil)
		if len(values) == 0 || slicex.IsEmpty(values) {
			return "无效值"
		}
		if len(values) == 1 {
			return fmt.Sprintf("无效输入：期望 %s", issues.StringifyPrimitive(values[0]))
		}
		return fmt.Sprintf("无效选项：期望 %s", joinValuesZhOr(values))

	case core.TooBig:
		return formatSizeConstraintZh(raw, false)

	case core.TooSmall:
		return formatSizeConstraintZh(raw, true)

	case core.InvalidFormat:
		format := mapx.GetStringDefault(raw.Properties, "format", "")
		if format == "" {
			return "格式无效"
		}
		return formatStringValidationZh(raw, format)

	case core.NotMultipleOf:
		divisor := mapx.GetAnyDefault(raw.Properties, "divisor", nil)
		if divisor == nil {
			return "无效数字：必须是倍数"
		}
		return fmt.Sprintf("无效数字：必须是 %v 的倍数", divisor)

	case core.UnrecognizedKeys:
		keys := mapx.GetStringsDefault(raw.Properties, "keys", nil)
		if len(keys) == 0 || slicex.IsEmpty(keys) {
			return "出现未知的键(key)"
		}
		// Use slicex for better key processing and issues.JoinValuesWithSeparator for formatting
		if keysAny, err := slicex.ToAny(keys); err == nil {
			// Remove duplicates and sort for consistent output
			if uniqueKeys, err := slicex.Unique(keysAny); err == nil {
				if typedKeys, err := slicex.ToAny(uniqueKeys); err == nil {
					keysJoined := issues.JoinValuesWithSeparator(typedKeys, ", ")
					return fmt.Sprintf("出现未知的键(key): %s", keysJoined)
				}
			}
		}
		return "出现未知的键(key)"

	case core.InvalidKey:
		origin := mapx.GetStringDefault(raw.Properties, "origin", "")
		if origin == "" {
			return "键(key)无效"
		}
		return fmt.Sprintf("%s 中的键(key)无效", origin)

	case core.InvalidUnion:
		return "无效输入"

	case core.InvalidElement:
		origin := mapx.GetStringDefault(raw.Properties, "origin", "")
		if origin == "" {
			return "元素无效"
		}
		return fmt.Sprintf("%s 中包含无效值(value)", origin)

	case core.Custom:
		message := mapx.GetStringDefault(raw.Properties, "message", "")
		if message != "" {
			return message
		}
		return "无效输入"

	default:
		return "无效输入"
	}
}

// =============================================================================
// SIZE CONSTRAINT FORMATTING - CHINESE STYLE
// =============================================================================

// formatSizeConstraintZh formats size constraint messages in Chinese
// Maintains natural Chinese grammar while matching TypeScript Zod v4 logic
func formatSizeConstraintZh(raw core.ZodRawIssue, isTooSmall bool) string {
	origin := mapx.GetStringDefault(raw.Properties, "origin", "")
	if origin == "" {
		origin = "值"
	}

	var threshold any
	if isTooSmall {
		threshold = mapx.GetAnyDefault(raw.Properties, "minimum", nil)
	} else {
		threshold = mapx.GetAnyDefault(raw.Properties, "maximum", nil)
	}

	if threshold == nil {
		if isTooSmall {
			return "数值过小"
		}
		return "数值过大"
	}

	inclusive := mapx.GetBoolDefault(raw.Properties, "inclusive", true)
	adj := issues.GetComparisonOperator(inclusive, isTooSmall)
	sizing := getSizingZh(origin)
	thresholdStr := issues.FormatThreshold(threshold)

	if sizing != nil {
		if isTooSmall {
			return fmt.Sprintf("数值过小：期望 %s %s%s %s", origin, adj, thresholdStr, sizing.Unit)
		} else {
			return fmt.Sprintf("数值过大：期望 %s %s%s %s", origin, adj, thresholdStr, sizing.Unit)
		}
	}

	if isTooSmall {
		return fmt.Sprintf("数值过小：期望 %s %s%s", origin, adj, thresholdStr)
	} else {
		return fmt.Sprintf("数值过大：期望 %s %s%s", origin, adj, thresholdStr)
	}
}

// =============================================================================
// STRING FORMAT VALIDATION - CHINESE STYLE
// =============================================================================

// formatStringValidationZh handles string format validation messages in Chinese
// Natural Chinese phrasing while maintaining TypeScript Zod v4 structure
func formatStringValidationZh(raw core.ZodRawIssue, format string) string {
	switch format {
	case "starts_with":
		prefix := mapx.GetStringDefault(raw.Properties, "prefix", "")
		if prefix == "" {
			return "无效字符串：必须以指定前缀开头"
		}
		return fmt.Sprintf("无效字符串：必须以 %s 开头", issues.StringifyPrimitive(prefix))
	case "ends_with":
		suffix := mapx.GetStringDefault(raw.Properties, "suffix", "")
		if suffix == "" {
			return "无效字符串：必须以指定后缀结尾"
		}
		return fmt.Sprintf("无效字符串：必须以 %s 结尾", issues.StringifyPrimitive(suffix))
	case "includes":
		includes := mapx.GetStringDefault(raw.Properties, "includes", "")
		if includes == "" {
			return "无效字符串：必须包含指定子字符串"
		}
		return fmt.Sprintf("无效字符串：必须包含 %s", issues.StringifyPrimitive(includes))
	case "regex":
		pattern := mapx.GetStringDefault(raw.Properties, "pattern", "")
		if pattern == "" {
			return "无效字符串：必须满足正则表达式"
		}
		return fmt.Sprintf("无效字符串：必须满足正则表达式 %s", pattern)
	default:
		noun := getFormatNounZh(format)
		return fmt.Sprintf("无效%s", noun)
	}
}

// ZhCN returns a ZodConfig configured for the Chinese locale.
func ZhCN() *core.ZodConfig {
	return &core.ZodConfig{
		LocaleError: formatZhCN,
	}
}
