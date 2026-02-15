package locales

import (
	"fmt"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// =============================================================================
// KOREAN LOCALE FORMATTER
// =============================================================================

// SizableKo maps Korean sizing information.
var SizableKo = map[string]issues.SizingInfo{
	"string": {Unit: "문자", Verb: ""},
	"file":   {Unit: "바이트", Verb: ""},
	"array":  {Unit: "개", Verb: ""},
	"slice":  {Unit: "개", Verb: ""},
	"set":    {Unit: "개", Verb: ""},
	"map":    {Unit: "개", Verb: ""},
}

// FormatNounsKo maps Korean format noun translations.
var FormatNounsKo = map[string]string{
	"regex":            "입력",
	"email":            "이메일 주소",
	"url":              "URL",
	"emoji":            "이모지",
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
	"datetime":         "ISO 날짜시간",
	"date":             "ISO 날짜",
	"time":             "ISO 시간",
	"duration":         "ISO 기간",
	"ipv4":             "IPv4 주소",
	"ipv6":             "IPv6 주소",
	"mac":              "MAC 주소",
	"cidrv4":           "IPv4 범위",
	"cidrv6":           "IPv6 범위",
	"base64":           "base64 인코딩 문자열",
	"base64url":        "base64url 인코딩 문자열",
	"json_string":      "JSON 문자열",
	"e164":             "E.164 번호",
	"jwt":              "JWT",
	"template_literal": "입력",
}

// TypeDictionaryKo maps Korean type name translations.
var TypeDictionaryKo = map[string]string{
	"nan":       "NaN",
	"number":    "숫자",
	"array":     "배열",
	"slice":     "배열",
	"string":    "문자열",
	"bool":      "불리언",
	"object":    "객체",
	"map":       "맵",
	"nil":       "null",
	"undefined": "undefined",
	"function":  "함수",
	"date":      "날짜",
	"file":      "파일",
	"set":       "세트",
}

// getSizingKo returns Korean sizing information for a given type
func getSizingKo(origin string) *issues.SizingInfo {
	if _, exists := SizableKo[origin]; exists {
		return new(SizableKo[origin])
	}
	return nil
}

// getFormatNounKo returns Korean noun for a format name
func getFormatNounKo(format string) string {
	if noun, exists := FormatNounsKo[format]; exists {
		return noun
	}
	return format
}

// getTypeNameKo returns Korean type name
func getTypeNameKo(typeName string) string {
	if name, exists := TypeDictionaryKo[typeName]; exists {
		return name
	}
	return typeName
}

// formatKo provides Korean error messages
func formatKo(raw core.ZodRawIssue) string {
	code := raw.Code

	switch code {
	case core.InvalidType:
		expected := mapx.StringOr(raw.Properties, "expected", "")
		expected = getTypeNameKo(expected)
		received := issues.ParsedTypeToString(raw.Input)
		received = getTypeNameKo(received)
		return fmt.Sprintf("잘못된 입력: 예상 타입은 %s, 받은 타입은 %s입니다", expected, received)

	case core.InvalidValue:
		values := mapx.AnySliceOr(raw.Properties, "values", nil)
		if len(values) == 0 {
			return "잘못된 값"
		}
		if len(values) == 1 {
			return fmt.Sprintf("잘못된 입력: 값은 %s 이어야 합니다", issues.StringifyPrimitive(values[0]))
		}
		return fmt.Sprintf("잘못된 옵션: %s 중 하나여야 합니다",
			issues.JoinValuesWithSeparator(values, " 또는 "))

	case core.TooBig:
		return formatSizeConstraintKo(raw, false)

	case core.TooSmall:
		return formatSizeConstraintKo(raw, true)

	case core.InvalidFormat:
		format := mapx.StringOr(raw.Properties, "format", "")
		if format == "" {
			return "잘못된 형식"
		}
		return formatStringValidationKo(raw, format)

	case core.NotMultipleOf:
		divisor := mapx.AnyOr(raw.Properties, "divisor", nil)
		if divisor == nil {
			return "잘못된 숫자: 배수여야 합니다"
		}
		return fmt.Sprintf("잘못된 숫자: %v의 배수여야 합니다", divisor)

	case core.UnrecognizedKeys:
		keys := mapx.StringsOr(raw.Properties, "keys", nil)
		if len(keys) == 0 {
			return "인식할 수 없는 키"
		}
		if keysAny, err := slicex.ToAny(keys); err == nil && !slicex.IsEmpty(keysAny) {
			keysJoined := issues.JoinValuesWithSeparator(keysAny, ", ")
			return fmt.Sprintf("인식할 수 없는 키: %s", keysJoined)
		}
		return "인식할 수 없는 키"

	case core.InvalidKey:
		origin := mapx.StringOr(raw.Properties, "origin", "")
		if origin == "" {
			return "잘못된 키"
		}
		return fmt.Sprintf("잘못된 키: %s", origin)

	case core.InvalidUnion:
		return "잘못된 입력"

	case core.InvalidElement:
		origin := mapx.StringOr(raw.Properties, "origin", "")
		if origin == "" {
			return "잘못된 요소"
		}
		return fmt.Sprintf("잘못된 값: %s", origin)

	case core.MissingRequired:
		fieldName := mapx.StringOr(raw.Properties, "field_name", "")
		fieldType := mapx.StringOr(raw.Properties, "field_type", "필드")
		if fieldName == "" {
			return fmt.Sprintf("필수 %s이(가) 없습니다", fieldType)
		}
		return fmt.Sprintf("필수 %s: %s이(가) 없습니다", fieldType, fieldName)

	case core.TypeConversion:
		fromType := mapx.StringOr(raw.Properties, "from_type", "알 수 없음")
		toType := mapx.StringOr(raw.Properties, "to_type", "알 수 없음")
		return fmt.Sprintf("타입 변환 실패: %s을(를) %s(으)로 변환할 수 없습니다", fromType, toType)

	case core.InvalidSchema:
		reason := mapx.StringOr(raw.Properties, "reason", "")
		if reason != "" {
			return fmt.Sprintf("잘못된 스키마: %s", reason)
		}
		return "잘못된 스키마 정의"

	case core.InvalidDiscriminator:
		field := mapx.StringOr(raw.Properties, "field", "판별자")
		return fmt.Sprintf("잘못되거나 누락된 판별자 필드: %s", field)

	case core.IncompatibleTypes:
		conflictType := mapx.StringOr(raw.Properties, "conflict_type", "값")
		return fmt.Sprintf("%s을(를) 병합할 수 없습니다: 호환되지 않는 타입", conflictType)

	case core.NilPointer:
		return "nil 포인터가 발견되었습니다"

	case core.Custom:
		message := mapx.StringOr(raw.Properties, "message", "")
		if message != "" {
			return message
		}
		return "잘못된 입력"

	default:
		return "잘못된 입력"
	}
}

// formatSizeConstraintKo formats size constraint messages in Korean
func formatSizeConstraintKo(raw core.ZodRawIssue, isTooSmall bool) string {
	origin := mapx.StringOr(raw.Properties, "origin", "")
	if origin == "" {
		origin = "값"
	}

	var threshold any
	if isTooSmall {
		threshold = mapx.AnyOr(raw.Properties, "minimum", nil)
	} else {
		threshold = mapx.AnyOr(raw.Properties, "maximum", nil)
	}

	if threshold == nil {
		if isTooSmall {
			return "너무 작습니다"
		}
		return "너무 큽니다"
	}

	inclusive := mapx.BoolOr(raw.Properties, "inclusive", true)
	sizing := getSizingKo(origin)
	thresholdStr := issues.FormatThreshold(threshold)

	// Korean comparison expressions
	var adj, suffix string
	if isTooSmall {
		if inclusive {
			adj = "이상"
			suffix = "이어야 합니다"
		} else {
			adj = "초과"
			suffix = "여야 합니다"
		}
	} else {
		if inclusive {
			adj = "이하"
			suffix = "여야 합니다"
		} else {
			adj = "미만"
			suffix = "이어야 합니다"
		}
	}

	if sizing != nil {
		if isTooSmall {
			return fmt.Sprintf("%s이(가) 너무 작습니다: %s%s %s%s", origin, thresholdStr, sizing.Unit, adj, suffix)
		}
		return fmt.Sprintf("%s이(가) 너무 큽니다: %s%s %s%s", origin, thresholdStr, sizing.Unit, adj, suffix)
	}

	if isTooSmall {
		return fmt.Sprintf("%s이(가) 너무 작습니다: %s %s%s", origin, thresholdStr, adj, suffix)
	}
	return fmt.Sprintf("%s이(가) 너무 큽니다: %s %s%s", origin, thresholdStr, adj, suffix)
}

// formatStringValidationKo handles string format validation messages in Korean
func formatStringValidationKo(raw core.ZodRawIssue, format string) string {
	switch format {
	case "starts_with":
		prefix := mapx.StringOr(raw.Properties, "prefix", "")
		if prefix == "" {
			return "잘못된 문자열: 지정된 접두사로 시작해야 합니다"
		}
		return fmt.Sprintf("잘못된 문자열: \"%s\"(으)로 시작해야 합니다", prefix)
	case "ends_with":
		suffix := mapx.StringOr(raw.Properties, "suffix", "")
		if suffix == "" {
			return "잘못된 문자열: 지정된 접미사로 끝나야 합니다"
		}
		return fmt.Sprintf("잘못된 문자열: \"%s\"(으)로 끝나야 합니다", suffix)
	case "includes":
		includes := mapx.StringOr(raw.Properties, "includes", "")
		if includes == "" {
			return "잘못된 문자열: 지정된 문자열을 포함해야 합니다"
		}
		return fmt.Sprintf("잘못된 문자열: \"%s\"을(를) 포함해야 합니다", includes)
	case "regex":
		pattern := mapx.StringOr(raw.Properties, "pattern", "")
		if pattern == "" {
			return "잘못된 문자열: 패턴과 일치해야 합니다"
		}
		return fmt.Sprintf("잘못된 문자열: 정규식 %s 패턴과 일치해야 합니다", pattern)
	default:
		noun := getFormatNounKo(format)
		return fmt.Sprintf("잘못된 %s", noun)
	}
}

// Ko returns a ZodConfig configured for Korean locale.
func Ko() *core.ZodConfig {
	return &core.ZodConfig{
		LocaleError: formatKo,
	}
}

// FormatMessageKo formats a single issue using Korean locale
func FormatMessageKo(issue core.ZodRawIssue) string {
	return formatKo(issue)
}
