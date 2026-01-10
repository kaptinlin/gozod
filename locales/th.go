package locales

import (
	"fmt"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// =============================================================================
// THAI LOCALE FORMATTER
// =============================================================================

// Thai sizing info mappings
var SizableTh = map[string]issues.SizingInfo{
	"string": {Unit: "ตัวอักษร", Verb: "ควรมี"},
	"file":   {Unit: "ไบต์", Verb: "ควรมี"},
	"array":  {Unit: "รายการ", Verb: "ควรมี"},
	"slice":  {Unit: "รายการ", Verb: "ควรมี"},
	"set":    {Unit: "รายการ", Verb: "ควรมี"},
	"map":    {Unit: "รายการ", Verb: "ควรมี"},
}

// Thai format noun mappings
var FormatNounsTh = map[string]string{
	"regex":            "ข้อมูลที่ป้อน",
	"email":            "ที่อยู่อีเมล",
	"url":              "URL",
	"emoji":            "อิโมจิ",
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
	"datetime":         "วันที่เวลาแบบ ISO",
	"date":             "วันที่แบบ ISO",
	"time":             "เวลาแบบ ISO",
	"duration":         "ช่วงเวลาแบบ ISO",
	"ipv4":             "ที่อยู่ IPv4",
	"ipv6":             "ที่อยู่ IPv6",
	"mac":              "ที่อยู่ MAC",
	"cidrv4":           "ช่วง IP แบบ IPv4",
	"cidrv6":           "ช่วง IP แบบ IPv6",
	"base64":           "ข้อความแบบ Base64",
	"base64url":        "ข้อความแบบ Base64 สำหรับ URL",
	"json_string":      "ข้อความแบบ JSON",
	"e164":             "เบอร์โทรศัพท์ระหว่างประเทศ (E.164)",
	"jwt":              "โทเคน JWT",
	"template_literal": "ข้อมูลที่ป้อน",
}

// Thai type dictionary
var TypeDictionaryTh = map[string]string{
	"nan":       "NaN",
	"number":    "ตัวเลข",
	"array":     "อาร์เรย์ (Array)",
	"slice":     "อาร์เรย์ (Array)",
	"string":    "ข้อความ",
	"bool":      "บูลีน",
	"object":    "อ็อบเจกต์",
	"map":       "แผนที่",
	"nil":       "ไม่มีค่า (null)",
	"undefined": "ไม่ได้กำหนดค่า",
	"function":  "ฟังก์ชัน",
	"date":      "วันที่",
	"file":      "ไฟล์",
	"set":       "เซต",
}

// getSizingTh returns Thai sizing information for a given type
func getSizingTh(origin string) *issues.SizingInfo {
	if info, exists := SizableTh[origin]; exists {
		return &info
	}
	return nil
}

// getFormatNounTh returns Thai noun for a format name
func getFormatNounTh(format string) string {
	if noun, exists := FormatNounsTh[format]; exists {
		return noun
	}
	return format
}

// getTypeNameTh returns Thai type name
func getTypeNameTh(typeName string) string {
	if name, exists := TypeDictionaryTh[typeName]; exists {
		return name
	}
	return typeName
}

// formatTh provides Thai error messages
func formatTh(raw core.ZodRawIssue) string {
	code := raw.Code

	switch code {
	case core.InvalidType:
		expected := mapx.GetStringDefault(raw.Properties, "expected", "")
		expected = getTypeNameTh(expected)
		received := issues.ParsedTypeToString(raw.Input)
		received = getTypeNameTh(received)
		return fmt.Sprintf("ประเภทข้อมูลไม่ถูกต้อง: ควรเป็น %s แต่ได้รับ %s", expected, received)

	case core.InvalidValue:
		values := mapx.GetAnySliceDefault(raw.Properties, "values", nil)
		if len(values) == 0 {
			return "ค่าไม่ถูกต้อง"
		}
		if len(values) == 1 {
			return fmt.Sprintf("ค่าไม่ถูกต้อง: ควรเป็น %s", issues.StringifyPrimitive(values[0]))
		}
		return fmt.Sprintf("ตัวเลือกไม่ถูกต้อง: ควรเป็นหนึ่งใน %s",
			issues.JoinValuesWithSeparator(values, "|"))

	case core.TooBig:
		return formatSizeConstraintTh(raw, false)

	case core.TooSmall:
		return formatSizeConstraintTh(raw, true)

	case core.InvalidFormat:
		format := mapx.GetStringDefault(raw.Properties, "format", "")
		if format == "" {
			return "รูปแบบไม่ถูกต้อง"
		}
		return formatStringValidationTh(raw, format)

	case core.NotMultipleOf:
		divisor := mapx.GetAnyDefault(raw.Properties, "divisor", nil)
		if divisor == nil {
			return "ตัวเลขไม่ถูกต้อง: ต้องเป็นจำนวนที่หารลงตัว"
		}
		return fmt.Sprintf("ตัวเลขไม่ถูกต้อง: ต้องเป็นจำนวนที่หารด้วย %v ได้ลงตัว", divisor)

	case core.UnrecognizedKeys:
		keys := mapx.GetStringsDefault(raw.Properties, "keys", nil)
		if len(keys) == 0 {
			return "พบคีย์ที่ไม่รู้จัก"
		}
		if keysAny, err := slicex.ToAny(keys); err == nil && !slicex.IsEmpty(keysAny) {
			keysJoined := issues.JoinValuesWithSeparator(keysAny, ", ")
			return fmt.Sprintf("พบคีย์ที่ไม่รู้จัก: %s", keysJoined)
		}
		return "พบคีย์ที่ไม่รู้จัก"

	case core.InvalidKey:
		origin := mapx.GetStringDefault(raw.Properties, "origin", "")
		if origin == "" {
			return "คีย์ไม่ถูกต้อง"
		}
		return fmt.Sprintf("คีย์ไม่ถูกต้องใน %s", origin)

	case core.InvalidUnion:
		return "ข้อมูลไม่ถูกต้อง: ไม่ตรงกับรูปแบบยูเนียนที่กำหนดไว้"

	case core.InvalidElement:
		origin := mapx.GetStringDefault(raw.Properties, "origin", "")
		if origin == "" {
			return "รายการไม่ถูกต้อง"
		}
		return fmt.Sprintf("ข้อมูลไม่ถูกต้องใน %s", origin)

	case core.MissingRequired:
		fieldName := mapx.GetStringDefault(raw.Properties, "field_name", "")
		fieldType := mapx.GetStringDefault(raw.Properties, "field_type", "ฟิลด์")
		if fieldName == "" {
			return fmt.Sprintf("ไม่พบ%sที่จำเป็น", fieldType)
		}
		return fmt.Sprintf("ไม่พบ%sที่จำเป็น: %s", fieldType, fieldName)

	case core.TypeConversion:
		fromType := mapx.GetStringDefault(raw.Properties, "from_type", "ไม่ทราบ")
		toType := mapx.GetStringDefault(raw.Properties, "to_type", "ไม่ทราบ")
		return fmt.Sprintf("การแปลงประเภทข้อมูลล้มเหลว: ไม่สามารถแปลง %s เป็น %s ได้", fromType, toType)

	case core.InvalidSchema:
		reason := mapx.GetStringDefault(raw.Properties, "reason", "")
		if reason != "" {
			return fmt.Sprintf("สคีมาไม่ถูกต้อง: %s", reason)
		}
		return "การกำหนดสคีมาไม่ถูกต้อง"

	case core.InvalidDiscriminator:
		field := mapx.GetStringDefault(raw.Properties, "field", "ตัวแบ่งแยก")
		return fmt.Sprintf("ฟิลด์ตัวแบ่งแยกไม่ถูกต้องหรือไม่มี: %s", field)

	case core.IncompatibleTypes:
		conflictType := mapx.GetStringDefault(raw.Properties, "conflict_type", "ค่า")
		return fmt.Sprintf("ไม่สามารถรวม %s ได้: ประเภทข้อมูลไม่เข้ากัน", conflictType)

	case core.NilPointer:
		return "พบตัวชี้ที่ไม่มีค่า (null pointer)"

	case core.Custom:
		message := mapx.GetStringDefault(raw.Properties, "message", "")
		if message != "" {
			return message
		}
		return "ข้อมูลไม่ถูกต้อง"

	default:
		return "ข้อมูลไม่ถูกต้อง"
	}
}

// formatSizeConstraintTh formats size constraint messages in Thai
func formatSizeConstraintTh(raw core.ZodRawIssue, isTooSmall bool) string {
	origin := mapx.GetStringDefault(raw.Properties, "origin", "")
	if origin == "" {
		origin = "ค่า"
	}

	var threshold any
	if isTooSmall {
		threshold = mapx.GetAnyDefault(raw.Properties, "minimum", nil)
	} else {
		threshold = mapx.GetAnyDefault(raw.Properties, "maximum", nil)
	}

	if threshold == nil {
		if isTooSmall {
			return "น้อยกว่ากำหนด"
		}
		return "เกินกำหนด"
	}

	inclusive := mapx.GetBoolDefault(raw.Properties, "inclusive", true)
	sizing := getSizingTh(mapx.GetStringDefault(raw.Properties, "origin", ""))
	thresholdStr := issues.FormatThreshold(threshold)

	// Thai comparison words
	var adj string
	if isTooSmall {
		if inclusive {
			adj = "อย่างน้อย"
		} else {
			adj = "มากกว่า"
		}
	} else {
		if inclusive {
			adj = "ไม่เกิน"
		} else {
			adj = "น้อยกว่า"
		}
	}

	if sizing != nil {
		if isTooSmall {
			return fmt.Sprintf("น้อยกว่ากำหนด: %s ควรมี%s %s %s", origin, adj, thresholdStr, sizing.Unit)
		}
		return fmt.Sprintf("เกินกำหนด: %s ควรมี%s %s %s", origin, adj, thresholdStr, sizing.Unit)
	}

	if isTooSmall {
		return fmt.Sprintf("น้อยกว่ากำหนด: %s ควรมี%s %s", origin, adj, thresholdStr)
	}
	return fmt.Sprintf("เกินกำหนด: %s ควรมี%s %s", origin, adj, thresholdStr)
}

// formatStringValidationTh handles string format validation messages in Thai
func formatStringValidationTh(raw core.ZodRawIssue, format string) string {
	switch format {
	case "starts_with":
		prefix := mapx.GetStringDefault(raw.Properties, "prefix", "")
		if prefix == "" {
			return "รูปแบบไม่ถูกต้อง: ข้อความต้องขึ้นต้นด้วยคำนำหน้าที่กำหนด"
		}
		return fmt.Sprintf("รูปแบบไม่ถูกต้อง: ข้อความต้องขึ้นต้นด้วย \"%s\"", prefix)
	case "ends_with":
		suffix := mapx.GetStringDefault(raw.Properties, "suffix", "")
		if suffix == "" {
			return "รูปแบบไม่ถูกต้อง: ข้อความต้องลงท้ายด้วยคำลงท้ายที่กำหนด"
		}
		return fmt.Sprintf("รูปแบบไม่ถูกต้อง: ข้อความต้องลงท้ายด้วย \"%s\"", suffix)
	case "includes":
		includes := mapx.GetStringDefault(raw.Properties, "includes", "")
		if includes == "" {
			return "รูปแบบไม่ถูกต้อง: ข้อความต้องมีข้อความย่อยที่กำหนดอยู่ในข้อความ"
		}
		return fmt.Sprintf("รูปแบบไม่ถูกต้อง: ข้อความต้องมี \"%s\" อยู่ในข้อความ", includes)
	case "regex":
		pattern := mapx.GetStringDefault(raw.Properties, "pattern", "")
		if pattern == "" {
			return "รูปแบบไม่ถูกต้อง: ต้องตรงกับรูปแบบที่กำหนด"
		}
		return fmt.Sprintf("รูปแบบไม่ถูกต้อง: ต้องตรงกับรูปแบบที่กำหนด %s", pattern)
	default:
		noun := getFormatNounTh(format)
		return fmt.Sprintf("รูปแบบไม่ถูกต้อง: %s", noun)
	}
}

// Th returns a ZodConfig configured for Thai locale.
func Th() *core.ZodConfig {
	return &core.ZodConfig{
		LocaleError: formatTh,
	}
}

// FormatMessageTh formats a single issue using Thai locale
func FormatMessageTh(issue core.ZodRawIssue) string {
	return formatTh(issue)
}
