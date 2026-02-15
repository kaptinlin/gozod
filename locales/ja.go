package locales

import (
	"fmt"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// =============================================================================
// JAPANESE LOCALE FORMATTER
// =============================================================================

// SizableJa maps Japanese sizing information.
var SizableJa = map[string]issues.SizingInfo{
	"string": {Unit: "文字", Verb: "である"},
	"file":   {Unit: "バイト", Verb: "である"},
	"array":  {Unit: "要素", Verb: "である"},
	"slice":  {Unit: "要素", Verb: "である"},
	"set":    {Unit: "要素", Verb: "である"},
	"map":    {Unit: "エントリ", Verb: "である"},
}

// FormatNounsJa maps Japanese format noun translations.
var FormatNounsJa = map[string]string{
	"regex":            "入力値",
	"email":            "メールアドレス",
	"url":              "URL",
	"emoji":            "絵文字",
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
	"datetime":         "ISO日時",
	"date":             "ISO日付",
	"time":             "ISO時刻",
	"duration":         "ISO期間",
	"ipv4":             "IPv4アドレス",
	"ipv6":             "IPv6アドレス",
	"mac":              "MACアドレス",
	"cidrv4":           "IPv4範囲",
	"cidrv6":           "IPv6範囲",
	"base64":           "base64エンコード文字列",
	"base64url":        "base64urlエンコード文字列",
	"json_string":      "JSON文字列",
	"e164":             "E.164番号",
	"jwt":              "JWT",
	"template_literal": "入力値",
}

// TypeDictionaryJa maps Japanese type name translations.
var TypeDictionaryJa = map[string]string{
	"nan":       "NaN",
	"number":    "数値",
	"array":     "配列",
	"slice":     "配列",
	"string":    "文字列",
	"bool":      "真偽値",
	"object":    "オブジェクト",
	"map":       "マップ",
	"nil":       "null",
	"undefined": "undefined",
	"function":  "関数",
	"date":      "日付",
	"file":      "ファイル",
	"set":       "セット",
}

// getSizingJa returns Japanese sizing information for a given type
func getSizingJa(origin string) *issues.SizingInfo {
	if _, exists := SizableJa[origin]; exists {
		return new(SizableJa[origin])
	}
	return nil
}

// getFormatNounJa returns Japanese noun for a format name
func getFormatNounJa(format string) string {
	if noun, exists := FormatNounsJa[format]; exists {
		return noun
	}
	return format
}

// getTypeNameJa returns Japanese type name
func getTypeNameJa(typeName string) string {
	if name, exists := TypeDictionaryJa[typeName]; exists {
		return name
	}
	return typeName
}

// formatJa provides Japanese error messages
func formatJa(raw core.ZodRawIssue) string {
	code := raw.Code

	switch code {
	case core.InvalidType:
		expected := mapx.StringOr(raw.Properties, "expected", "")
		expected = getTypeNameJa(expected)
		received := issues.ParsedTypeToString(raw.Input)
		received = getTypeNameJa(received)
		return fmt.Sprintf("無効な入力: %sが期待されましたが、%sが入力されました", expected, received)

	case core.InvalidValue:
		values := mapx.AnySliceOr(raw.Properties, "values", nil)
		if len(values) == 0 {
			return "無効な値"
		}
		if len(values) == 1 {
			return fmt.Sprintf("無効な入力: %sが期待されました", issues.StringifyPrimitive(values[0]))
		}
		return fmt.Sprintf("無効な選択: %sのいずれかである必要があります",
			issues.JoinValuesWithSeparator(values, "、"))

	case core.TooBig:
		return formatSizeConstraintJa(raw, false)

	case core.TooSmall:
		return formatSizeConstraintJa(raw, true)

	case core.InvalidFormat:
		format := mapx.StringOr(raw.Properties, "format", "")
		if format == "" {
			return "無効な形式"
		}
		return formatStringValidationJa(raw, format)

	case core.NotMultipleOf:
		divisor := mapx.AnyOr(raw.Properties, "divisor", nil)
		if divisor == nil {
			return "無効な数値: 倍数である必要があります"
		}
		return fmt.Sprintf("無効な数値: %vの倍数である必要があります", divisor)

	case core.UnrecognizedKeys:
		keys := mapx.StringsOr(raw.Properties, "keys", nil)
		if len(keys) == 0 {
			return "認識されていないキー"
		}
		suffix := ""
		if len(keys) > 1 {
			suffix = "群"
		}
		if keysAny, err := slicex.ToAny(keys); err == nil && !slicex.IsEmpty(keysAny) {
			keysJoined := issues.JoinValuesWithSeparator(keysAny, "、")
			return fmt.Sprintf("認識されていないキー%s: %s", suffix, keysJoined)
		}
		return "認識されていないキー"

	case core.InvalidKey:
		origin := mapx.StringOr(raw.Properties, "origin", "")
		if origin == "" {
			return "無効なキー"
		}
		return fmt.Sprintf("%s内の無効なキー", origin)

	case core.InvalidUnion:
		return "無効な入力"

	case core.InvalidElement:
		origin := mapx.StringOr(raw.Properties, "origin", "")
		if origin == "" {
			return "無効な要素"
		}
		return fmt.Sprintf("%s内の無効な値", origin)

	case core.MissingRequired:
		fieldName := mapx.StringOr(raw.Properties, "field_name", "")
		fieldType := mapx.StringOr(raw.Properties, "field_type", "フィールド")
		if fieldName == "" {
			return fmt.Sprintf("必須の%sがありません", fieldType)
		}
		return fmt.Sprintf("必須の%s: %sがありません", fieldType, fieldName)

	case core.TypeConversion:
		fromType := mapx.StringOr(raw.Properties, "from_type", "不明")
		toType := mapx.StringOr(raw.Properties, "to_type", "不明")
		return fmt.Sprintf("型変換エラー: %sを%sに変換できません", fromType, toType)

	case core.InvalidSchema:
		reason := mapx.StringOr(raw.Properties, "reason", "")
		if reason != "" {
			return fmt.Sprintf("無効なスキーマ: %s", reason)
		}
		return "無効なスキーマ定義"

	case core.InvalidDiscriminator:
		field := mapx.StringOr(raw.Properties, "field", "判別フィールド")
		return fmt.Sprintf("無効または欠落している判別フィールド: %s", field)

	case core.IncompatibleTypes:
		conflictType := mapx.StringOr(raw.Properties, "conflict_type", "値")
		return fmt.Sprintf("%sをマージできません: 互換性のない型", conflictType)

	case core.NilPointer:
		return "nilポインタが検出されました"

	case core.Custom:
		message := mapx.StringOr(raw.Properties, "message", "")
		if message != "" {
			return message
		}
		return "無効な入力"

	default:
		return "無効な入力"
	}
}

// formatSizeConstraintJa formats size constraint messages in Japanese
func formatSizeConstraintJa(raw core.ZodRawIssue, isTooSmall bool) string {
	origin := mapx.StringOr(raw.Properties, "origin", "")
	if origin == "" {
		origin = "値"
	}

	var threshold any
	if isTooSmall {
		threshold = mapx.AnyOr(raw.Properties, "minimum", nil)
	} else {
		threshold = mapx.AnyOr(raw.Properties, "maximum", nil)
	}

	if threshold == nil {
		if isTooSmall {
			return "小さすぎる値"
		}
		return "大きすぎる値"
	}

	inclusive := mapx.BoolOr(raw.Properties, "inclusive", true)
	sizing := getSizingJa(origin)
	thresholdStr := issues.FormatThreshold(threshold)

	// Japanese comparison expressions
	var adj string
	if isTooSmall {
		if inclusive {
			adj = "以上である"
		} else {
			adj = "より大きい"
		}
	} else {
		if inclusive {
			adj = "以下である"
		} else {
			adj = "より小さい"
		}
	}

	if sizing != nil {
		if isTooSmall {
			return fmt.Sprintf("小さすぎる値: %sは%s%s%s必要があります", origin, thresholdStr, sizing.Unit, adj)
		}
		return fmt.Sprintf("大きすぎる値: %sは%s%s%s必要があります", origin, thresholdStr, sizing.Unit, adj)
	}

	if isTooSmall {
		return fmt.Sprintf("小さすぎる値: %sは%s%s必要があります", origin, thresholdStr, adj)
	}
	return fmt.Sprintf("大きすぎる値: %sは%s%s必要があります", origin, thresholdStr, adj)
}

// formatStringValidationJa handles string format validation messages in Japanese
func formatStringValidationJa(raw core.ZodRawIssue, format string) string {
	switch format {
	case "starts_with":
		prefix := mapx.StringOr(raw.Properties, "prefix", "")
		if prefix == "" {
			return "無効な文字列: 指定された接頭辞で始まる必要があります"
		}
		return fmt.Sprintf("無効な文字列: \"%s\"で始まる必要があります", prefix)
	case "ends_with":
		suffix := mapx.StringOr(raw.Properties, "suffix", "")
		if suffix == "" {
			return "無効な文字列: 指定された接尾辞で終わる必要があります"
		}
		return fmt.Sprintf("無効な文字列: \"%s\"で終わる必要があります", suffix)
	case "includes":
		includes := mapx.StringOr(raw.Properties, "includes", "")
		if includes == "" {
			return "無効な文字列: 指定された文字列を含む必要があります"
		}
		return fmt.Sprintf("無効な文字列: \"%s\"を含む必要があります", includes)
	case "regex":
		pattern := mapx.StringOr(raw.Properties, "pattern", "")
		if pattern == "" {
			return "無効な文字列: パターンに一致する必要があります"
		}
		return fmt.Sprintf("無効な文字列: パターン%sに一致する必要があります", pattern)
	default:
		noun := getFormatNounJa(format)
		return fmt.Sprintf("無効な%s", noun)
	}
}

// Ja returns a ZodConfig configured for Japanese locale.
func Ja() *core.ZodConfig {
	return &core.ZodConfig{
		LocaleError: formatJa,
	}
}

// FormatMessageJa formats a single issue using Japanese locale
func FormatMessageJa(issue core.ZodRawIssue) string {
	return formatJa(issue)
}
