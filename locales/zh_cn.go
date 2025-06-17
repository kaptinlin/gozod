package locales

import (
	"fmt"

	"github.com/kaptinlin/gozod"
)

// Sizable maps types to their sizing information in Chinese
var sizableZh = map[string]struct {
	Unit string
	Verb string
}{
	"string": {Unit: "字符", Verb: "包含"},
	"file":   {Unit: "字节", Verb: "包含"},
	"array":  {Unit: "项", Verb: "包含"},
	"slice":  {Unit: "项", Verb: "包含"},
	"set":    {Unit: "项", Verb: "包含"},
}

// getSizingZh returns sizing information for a given type in Chinese
func getSizingZh(origin string) *struct {
	Unit string
	Verb string
} {
	if info, ok := sizableZh[origin]; ok {
		return &info
	}
	return nil
}

// parsedTypeZh determines the type of a value in Chinese
func parsedTypeZh(data interface{}) string {
	switch gozod.GetParsedType(data) {
	case gozod.ParsedTypeString:
		return "字符串"
	case gozod.ParsedTypeNumber:
		return "数字"
	case gozod.ParsedTypeBigint:
		return "大整数"
	case gozod.ParsedTypeBool:
		return "布尔值"
	case gozod.ParsedTypeFloat:
		return "浮点数"
	case gozod.ParsedTypeObject:
		return "对象"
	case gozod.ParsedTypeFunction:
		return "函数"
	case gozod.ParsedTypeFile:
		return "文件"
	case gozod.ParsedTypeDate:
		return "日期"
	case gozod.ParsedTypeArray:
		return "数组"
	case gozod.ParsedTypeSlice:
		return "切片"
	case gozod.ParsedTypeMap:
		return "映射"
	case gozod.ParsedTypeNaN:
		return "非数字(NaN)"
	case gozod.ParsedTypeNil:
		return "空值(nil)"
	case gozod.ParsedTypeComplex:
		return "复数"
	default:
		return string(gozod.GetParsedType(data))
	}
}

// Nouns maps string formats to their Chinese names
var nounsZh = map[string]string{
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
}

func errorZhCN(issue gozod.ZodRawIssue) string {
	switch issue.Code {
	case "invalid_type":
		return fmt.Sprintf("无效输入：期望 %s，实际接收 %s", issue.GetExpected(), parsedTypeZh(issue.Input))
	case "invalid_value":
		values := issue.GetValues()
		if len(values) == 1 {
			return fmt.Sprintf("无效输入：期望 %s", gozod.StringifyPrimitive(values[0]))
		}
		return fmt.Sprintf("无效选项：期望以下之一 %s", gozod.JoinValues(values, "|"))
	case "too_big":
		adj := "<"
		if issue.GetInclusive() {
			adj = "<="
		}
		origin := issue.GetOrigin()
		sizing := getSizingZh(origin)
		if sizing != nil {
			if origin == "" {
				origin = "值"
			}
			unit := sizing.Unit
			if unit == "" {
				unit = "个元素"
			}
			// 为 slice/array 类型使用一致的格式以匹配主错误生成逻辑
			if origin == "slice" || origin == "array" {
				// 保持与测试期望一致的格式
				return fmt.Sprintf("数值过大：期望 %s %s%v %s", origin, adj, issue.GetMaximum(), unit)
			}
			return fmt.Sprintf("数值过大：期望 %s %s%v %s", origin, adj, issue.GetMaximum(), unit)
		}
		if origin == "" {
			origin = "值"
		}
		return fmt.Sprintf("数值过大：期望 %s %s%v", origin, adj, issue.GetMaximum())
	case "too_small":
		adj := ">"
		if issue.GetInclusive() {
			adj = ">="
		}
		origin := issue.GetOrigin()
		sizing := getSizingZh(origin)
		if sizing != nil {
			if origin == "slice" || origin == "array" {
				return fmt.Sprintf("数值过小：期望 %s %s%v %s", origin, adj, issue.GetMinimum(), sizing.Unit)
			}
			return fmt.Sprintf("数值过小：期望 %s %s%v %s", origin, adj, issue.GetMinimum(), sizing.Unit)
		}
		return fmt.Sprintf("数值过小：期望 %s %s%v", origin, adj, issue.GetMinimum())
	case "invalid_format":
		format := issue.GetFormat()
		if format == "starts_with" {
			return fmt.Sprintf("无效字符串：必须以 \"%s\" 开头", issue.GetPrefix())
		}
		if format == "ends_with" {
			return fmt.Sprintf("无效字符串：必须以 \"%s\" 结尾", issue.GetSuffix())
		}
		if format == "includes" {
			return fmt.Sprintf("无效字符串：必须包含 \"%s\"", issue.GetIncludes())
		}
		if format == "regex" {
			return fmt.Sprintf("无效字符串：必须满足正则表达式 %s", issue.GetPattern())
		}
		noun := nounsZh[format]
		if noun == "" {
			noun = format
		}
		return fmt.Sprintf("无效%s", noun)
	case "not_multiple_of":
		return fmt.Sprintf("无效数字：必须是 %v 的倍数", issue.GetDivisor())
	case "unrecognized_keys":
		keys := issue.GetKeys()
		keyInterfaces := make([]interface{}, len(keys))
		for i, k := range keys {
			keyInterfaces[i] = k
		}
		return fmt.Sprintf("出现未知的键(key): %s", gozod.JoinValues(keyInterfaces, ", "))
	case "invalid_key":
		return fmt.Sprintf("%s 中的键(key)无效", issue.GetOrigin())
	case "invalid_union":
		return "无效输入"
	case "invalid_element":
		return fmt.Sprintf("%s 中包含无效值(value)", issue.GetOrigin())
	default:
		return "无效输入"
	}
}

// ZhCN returns the default Chinese locale error map function
func ZhCN() func(gozod.ZodRawIssue) string {
	return errorZhCN
}
