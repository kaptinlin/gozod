package locales

import (
	"fmt"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// =============================================================================
// HEBREW LOCALE FORMATTER
// =============================================================================

// Hebrew type info with grammatical gender
type hebrewTypeInfo struct {
	Label  string
	Gender string // "m" for masculine, "f" for feminine
}

// Hebrew sizing info with size labels
type hebrewSizingInfo struct {
	Unit       string
	ShortLabel string // for "too small"
	LongLabel  string // for "too big"
}

// Hebrew type names with gender
var TypeNamesHe = map[string]hebrewTypeInfo{
	"string":    {Label: "מחרוזת", Gender: "f"},
	"number":    {Label: "מספר", Gender: "m"},
	"bool":      {Label: "ערך בוליאני", Gender: "m"},
	"boolean":   {Label: "ערך בוליאני", Gender: "m"},
	"bigint":    {Label: "BigInt", Gender: "m"},
	"date":      {Label: "תאריך", Gender: "m"},
	"array":     {Label: "מערך", Gender: "m"},
	"slice":     {Label: "מערך", Gender: "m"},
	"object":    {Label: "אובייקט", Gender: "m"},
	"nil":       {Label: "ערך ריק (null)", Gender: "m"},
	"undefined": {Label: "ערך לא מוגדר (undefined)", Gender: "m"},
	"symbol":    {Label: "סימבול (Symbol)", Gender: "m"},
	"function":  {Label: "פונקציה", Gender: "f"},
	"map":       {Label: "מפה (Map)", Gender: "f"},
	"set":       {Label: "קבוצה (Set)", Gender: "f"},
	"file":      {Label: "קובץ", Gender: "m"},
	"promise":   {Label: "Promise", Gender: "m"},
	"NaN":       {Label: "NaN", Gender: "m"},
	"unknown":   {Label: "ערך לא ידוע", Gender: "m"},
	"value":     {Label: "ערך", Gender: "m"},
}

// Hebrew sizing info mappings
var SizableHe = map[string]hebrewSizingInfo{
	"string": {Unit: "תווים", ShortLabel: "קצר", LongLabel: "ארוך"},
	"file":   {Unit: "בייטים", ShortLabel: "קטן", LongLabel: "גדול"},
	"array":  {Unit: "פריטים", ShortLabel: "קטן", LongLabel: "גדול"},
	"slice":  {Unit: "פריטים", ShortLabel: "קטן", LongLabel: "גדול"},
	"set":    {Unit: "פריטים", ShortLabel: "קטן", LongLabel: "גדול"},
	"map":    {Unit: "פריטים", ShortLabel: "קטן", LongLabel: "גדול"},
	"number": {Unit: "", ShortLabel: "קטן", LongLabel: "גדול"},
}

// Hebrew format noun mappings with gender
var FormatNounsHe = map[string]hebrewTypeInfo{
	"regex":            {Label: "קלט", Gender: "m"},
	"email":            {Label: "כתובת אימייל", Gender: "f"},
	"url":              {Label: "כתובת רשת", Gender: "f"},
	"emoji":            {Label: "אימוג'י", Gender: "m"},
	"uuid":             {Label: "UUID", Gender: "m"},
	"uuidv4":           {Label: "UUIDv4", Gender: "m"},
	"uuidv6":           {Label: "UUIDv6", Gender: "m"},
	"nanoid":           {Label: "nanoid", Gender: "m"},
	"guid":             {Label: "GUID", Gender: "m"},
	"cuid":             {Label: "cuid", Gender: "m"},
	"cuid2":            {Label: "cuid2", Gender: "m"},
	"ulid":             {Label: "ULID", Gender: "m"},
	"xid":              {Label: "XID", Gender: "m"},
	"ksuid":            {Label: "KSUID", Gender: "m"},
	"datetime":         {Label: "תאריך וזמן ISO", Gender: "m"},
	"date":             {Label: "תאריך ISO", Gender: "m"},
	"time":             {Label: "זמן ISO", Gender: "m"},
	"duration":         {Label: "משך זמן ISO", Gender: "m"},
	"ipv4":             {Label: "כתובת IPv4", Gender: "f"},
	"ipv6":             {Label: "כתובת IPv6", Gender: "f"},
	"mac":              {Label: "כתובת MAC", Gender: "f"},
	"cidrv4":           {Label: "טווח IPv4", Gender: "m"},
	"cidrv6":           {Label: "טווח IPv6", Gender: "m"},
	"base64":           {Label: "מחרוזת בבסיס 64", Gender: "f"},
	"base64url":        {Label: "מחרוזת בבסיס 64 לכתובות רשת", Gender: "f"},
	"json_string":      {Label: "מחרוזת JSON", Gender: "f"},
	"e164":             {Label: "מספר E.164", Gender: "m"},
	"jwt":              {Label: "JWT", Gender: "m"},
	"template_literal": {Label: "קלט", Gender: "m"},
}

// getSizingHe returns Hebrew sizing information for a given type
func getSizingHe(origin string) *hebrewSizingInfo {
	if _, exists := SizableHe[origin]; exists {
		return new(SizableHe[origin])
	}
	return nil
}

// getTypeLabelHe returns Hebrew label for a type
func getTypeLabelHe(typeName string) string {
	if info, exists := TypeNamesHe[typeName]; exists {
		return info.Label
	}
	return typeName
}

// getTypeGenderHe returns Hebrew grammatical gender for a type
func getTypeGenderHe(typeName string) string {
	if info, exists := TypeNamesHe[typeName]; exists {
		return info.Gender
	}
	return "m" // default to masculine
}

// getVerbForHe returns the appropriate verb form based on gender
func getVerbForHe(typeName string) string {
	gender := getTypeGenderHe(typeName)
	if gender == "f" {
		return "צריכה להיות"
	}
	return "צריך להיות"
}

// withDefiniteHe adds the Hebrew definite article
func withDefiniteHe(typeName string) string {
	return "ה" + getTypeLabelHe(typeName)
}

// formatHe provides Hebrew error messages
func formatHe(raw core.ZodRawIssue) string {
	code := raw.Code

	switch code {
	case core.InvalidType:
		expected := mapx.StringOr(raw.Properties, "expected", "")
		expected = getTypeLabelHe(expected)
		received := issues.ParsedTypeToString(raw.Input)
		received = getTypeLabelHe(received)
		return fmt.Sprintf("קלט לא תקין: צריך להיות %s, התקבל %s", expected, received)

	case core.InvalidValue:
		values := mapx.AnySliceOr(raw.Properties, "values", nil)
		if len(values) == 0 {
			return "ערך לא תקין"
		}
		if len(values) == 1 {
			return fmt.Sprintf("ערך לא תקין: הערך חייב להיות %s", issues.StringifyPrimitive(values[0]))
		}
		if len(values) == 2 {
			return fmt.Sprintf("ערך לא תקין: האפשרויות המתאימות הן %s או %s",
				issues.StringifyPrimitive(values[0]),
				issues.StringifyPrimitive(values[1]))
		}
		// For 3+ values
		return fmt.Sprintf("ערך לא תקין: האפשרויות המתאימות הן %s",
			issues.JoinValuesWithSeparator(values, ", "))

	case core.TooBig:
		return formatSizeConstraintHe(raw, false)

	case core.TooSmall:
		return formatSizeConstraintHe(raw, true)

	case core.InvalidFormat:
		format := mapx.StringOr(raw.Properties, "format", "")
		if format == "" {
			return "פורמט לא תקין"
		}
		return formatStringValidationHe(raw, format)

	case core.NotMultipleOf:
		divisor := mapx.AnyOr(raw.Properties, "divisor", nil)
		if divisor == nil {
			return "מספר לא תקין: חייב להיות מכפלה"
		}
		return fmt.Sprintf("מספר לא תקין: חייב להיות מכפלה של %v", divisor)

	case core.UnrecognizedKeys:
		keys := mapx.StringsOr(raw.Properties, "keys", nil)
		if len(keys) == 0 {
			return "מפתח לא מזוהה"
		}
		keyWord := "מפתח לא מזוהה"
		if len(keys) > 1 {
			keyWord = "מפתחות לא מזוהים"
		}
		if keysAny, err := slicex.ToAny(keys); err == nil && !slicex.IsEmpty(keysAny) {
			keysJoined := issues.JoinValuesWithSeparator(keysAny, ", ")
			return fmt.Sprintf("%s: %s", keyWord, keysJoined)
		}
		return keyWord

	case core.InvalidKey:
		return "שדה לא תקין באובייקט"

	case core.InvalidUnion:
		return "קלט לא תקין"

	case core.InvalidElement:
		origin := mapx.StringOr(raw.Properties, "origin", "")
		if origin == "" {
			return "ערך לא תקין"
		}
		return fmt.Sprintf("ערך לא תקין ב%s", withDefiniteHe(origin))

	case core.MissingRequired:
		fieldName := mapx.StringOr(raw.Properties, "field_name", "")
		fieldType := mapx.StringOr(raw.Properties, "field_type", "שדה")
		if fieldName == "" {
			return fmt.Sprintf("%s נדרש חסר", fieldType)
		}
		return fmt.Sprintf("%s נדרש חסר: %s", fieldType, fieldName)

	case core.TypeConversion:
		fromType := mapx.StringOr(raw.Properties, "from_type", "לא ידוע")
		toType := mapx.StringOr(raw.Properties, "to_type", "לא ידוע")
		return fmt.Sprintf("המרת סוג נכשלה: לא ניתן להמיר %s ל-%s", fromType, toType)

	case core.InvalidSchema:
		reason := mapx.StringOr(raw.Properties, "reason", "")
		if reason != "" {
			return fmt.Sprintf("סכמה לא תקינה: %s", reason)
		}
		return "הגדרת סכמה לא תקינה"

	case core.InvalidDiscriminator:
		field := mapx.StringOr(raw.Properties, "field", "מפריד")
		return fmt.Sprintf("שדה מפריד לא תקין או חסר: %s", field)

	case core.IncompatibleTypes:
		conflictType := mapx.StringOr(raw.Properties, "conflict_type", "ערכים")
		return fmt.Sprintf("לא ניתן למזג %s: סוגים לא תואמים", conflictType)

	case core.NilPointer:
		return "זוהה מצביע ריק"

	case core.Custom:
		message := mapx.StringOr(raw.Properties, "message", "")
		if message != "" {
			return message
		}
		return "קלט לא תקין"

	default:
		return "קלט לא תקין"
	}
}

// formatSizeConstraintHe formats size constraint messages in Hebrew
func formatSizeConstraintHe(raw core.ZodRawIssue, isTooSmall bool) string {
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
			return "קטן מדי"
		}
		return "גדול מדי"
	}

	inclusive := mapx.BoolOr(raw.Properties, "inclusive", true)
	sizing := getSizingHe(origin)
	thresholdStr := issues.FormatThreshold(threshold)
	subject := withDefiniteHe(origin)

	// Handle specific types with natural Hebrew
	if origin == "string" {
		if sizing != nil {
			if isTooSmall {
				comparison := thresholdStr + " " + sizing.Unit
				if inclusive {
					comparison += " או יותר"
				} else {
					comparison = "לפחות " + comparison
				}
				return fmt.Sprintf("%s מדי: %s צריכה להכיל %s", sizing.ShortLabel, subject, comparison)
			}
			comparison := thresholdStr + " " + sizing.Unit
			if inclusive {
				comparison += " או פחות"
			} else {
				comparison = "לכל היותר " + comparison
			}
			return fmt.Sprintf("%s מדי: %s צריכה להכיל %s", sizing.LongLabel, subject, comparison)
		}
	}

	if origin == "number" {
		if isTooSmall {
			var comparison string
			if inclusive {
				comparison = fmt.Sprintf("גדול או שווה ל-%s", thresholdStr)
			} else {
				comparison = fmt.Sprintf("גדול מ-%s", thresholdStr)
			}
			return fmt.Sprintf("קטן מדי: %s צריך להיות %s", subject, comparison)
		}
		var comparison string
		if inclusive {
			comparison = fmt.Sprintf("קטן או שווה ל-%s", thresholdStr)
		} else {
			comparison = fmt.Sprintf("קטן מ-%s", thresholdStr)
		}
		return fmt.Sprintf("גדול מדי: %s צריך להיות %s", subject, comparison)
	}

	if origin == "array" || origin == "slice" || origin == "set" {
		verb := "צריך"
		if origin == "set" {
			verb = "צריכה"
		}
		if sizing != nil {
			if isTooSmall {
				comparison := thresholdStr + " " + sizing.Unit
				if inclusive {
					comparison += " או יותר"
				} else {
					comparison = "יותר מ-" + thresholdStr + " " + sizing.Unit
				}
				return fmt.Sprintf("קטן מדי: %s %s להכיל %s", subject, verb, comparison)
			}
			comparison := thresholdStr + " " + sizing.Unit
			if inclusive {
				comparison += " או פחות"
			} else {
				comparison = "פחות מ-" + thresholdStr + " " + sizing.Unit
			}
			return fmt.Sprintf("גדול מדי: %s %s להכיל %s", subject, verb, comparison)
		}
	}

	// Default format
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

	sizeLabel := "גדול"
	if isTooSmall {
		sizeLabel = "קטן"
	}
	if sizing != nil {
		if isTooSmall {
			sizeLabel = sizing.ShortLabel
		} else {
			sizeLabel = sizing.LongLabel
		}
	}

	be := getVerbForHe(origin)
	if sizing != nil && sizing.Unit != "" {
		return fmt.Sprintf("%s מדי: %s %s %s%s %s", sizeLabel, subject, be, adj, thresholdStr, sizing.Unit)
	}
	return fmt.Sprintf("%s מדי: %s %s %s%s", sizeLabel, subject, be, adj, thresholdStr)
}

// formatStringValidationHe handles string format validation messages in Hebrew
func formatStringValidationHe(raw core.ZodRawIssue, format string) string {
	switch format {
	case "starts_with":
		prefix := mapx.StringOr(raw.Properties, "prefix", "")
		if prefix == "" {
			return "המחרוזת חייבת להתחיל בקידומת מסוימת"
		}
		return fmt.Sprintf("המחרוזת חייבת להתחיל ב \"%s\"", prefix)
	case "ends_with":
		suffix := mapx.StringOr(raw.Properties, "suffix", "")
		if suffix == "" {
			return "המחרוזת חייבת להסתיים בסיומת מסוימת"
		}
		return fmt.Sprintf("המחרוזת חייבת להסתיים ב \"%s\"", suffix)
	case "includes":
		includes := mapx.StringOr(raw.Properties, "includes", "")
		if includes == "" {
			return "המחרוזת חייבת לכלול מחרוזת מסוימת"
		}
		return fmt.Sprintf("המחרוזת חייבת לכלול \"%s\"", includes)
	case "regex":
		pattern := mapx.StringOr(raw.Properties, "pattern", "")
		if pattern == "" {
			return "המחרוזת חייבת להתאים לתבנית"
		}
		return fmt.Sprintf("המחרוזת חייבת להתאים לתבנית %s", pattern)
	default:
		if info, exists := FormatNounsHe[format]; exists {
			adjective := "תקין"
			if info.Gender == "f" {
				adjective = "תקינה"
			}
			return fmt.Sprintf("%s לא %s", info.Label, adjective)
		}
		return fmt.Sprintf("%s לא תקין", format)
	}
}

// He returns a ZodConfig configured for Hebrew locale.
func He() *core.ZodConfig {
	return &core.ZodConfig{
		LocaleError: formatHe,
	}
}

// FormatMessageHe formats a single issue using Hebrew locale
func FormatMessageHe(issue core.ZodRawIssue) string {
	return formatHe(issue)
}
