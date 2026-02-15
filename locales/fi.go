package locales

import (
	"fmt"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// =============================================================================
// FINNISH LOCALE FORMATTER
// =============================================================================

// Finnish sizing info with subject (genitive form)
type finnishSizingInfo struct {
	Unit    string
	Subject string // genitive form for "of X"
}

// SizableFi maps Finnish sizing information.
var SizableFi = map[string]finnishSizingInfo{
	"string": {Unit: "merkkiä", Subject: "merkkijonon"},
	"file":   {Unit: "tavua", Subject: "tiedoston"},
	"array":  {Unit: "alkiota", Subject: "listan"},
	"slice":  {Unit: "alkiota", Subject: "listan"},
	"set":    {Unit: "alkiota", Subject: "joukon"},
	"map":    {Unit: "merkintää", Subject: "kartan"},
	"number": {Unit: "", Subject: "luvun"},
	"bigint": {Unit: "", Subject: "suuren kokonaisluvun"},
	"int":    {Unit: "", Subject: "kokonaisluvun"},
	"date":   {Unit: "", Subject: "päivämäärän"},
}

// FormatNounsFi maps Finnish format noun translations.
var FormatNounsFi = map[string]string{
	"regex":            "säännöllinen lauseke",
	"email":            "sähköpostiosoite",
	"url":              "URL-osoite",
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
	"datetime":         "ISO-aikaleima",
	"date":             "ISO-päivämäärä",
	"time":             "ISO-aika",
	"duration":         "ISO-kesto",
	"ipv4":             "IPv4-osoite",
	"ipv6":             "IPv6-osoite",
	"mac":              "MAC-osoite",
	"cidrv4":           "IPv4-alue",
	"cidrv6":           "IPv6-alue",
	"base64":           "base64-koodattu merkkijono",
	"base64url":        "base64url-koodattu merkkijono",
	"json_string":      "JSON-merkkijono",
	"e164":             "E.164-luku",
	"jwt":              "JWT",
	"template_literal": "templaattimerkkijono",
}

// TypeDictionaryFi maps Finnish type name translations.
var TypeDictionaryFi = map[string]string{
	"nan":       "NaN",
	"number":    "numero",
	"array":     "lista",
	"slice":     "lista",
	"string":    "merkkijono",
	"bool":      "totuusarvo",
	"object":    "objekti",
	"map":       "kartta",
	"nil":       "null",
	"undefined": "määrittelemätön",
	"function":  "funktio",
	"date":      "päivämäärä",
	"file":      "tiedosto",
	"set":       "joukko",
}

// getSizingFi returns Finnish sizing information for a given type
func getSizingFi(origin string) *finnishSizingInfo {
	if _, exists := SizableFi[origin]; exists {
		return new(SizableFi[origin])
	}
	return nil
}

// getFormatNounFi returns Finnish noun for a format name
func getFormatNounFi(format string) string {
	if noun, exists := FormatNounsFi[format]; exists {
		return noun
	}
	return format
}

// getTypeNameFi returns Finnish type name
func getTypeNameFi(typeName string) string {
	if name, exists := TypeDictionaryFi[typeName]; exists {
		return name
	}
	return typeName
}

// formatFi provides Finnish error messages
func formatFi(raw core.ZodRawIssue) string {
	code := raw.Code

	switch code {
	case core.InvalidType:
		expected := mapx.StringOr(raw.Properties, "expected", "")
		expected = getTypeNameFi(expected)
		received := issues.ParsedTypeToString(raw.Input)
		received = getTypeNameFi(received)
		return fmt.Sprintf("Virheellinen tyyppi: odotettiin %s, oli %s", expected, received)

	case core.InvalidValue:
		values := mapx.AnySliceOr(raw.Properties, "values", nil)
		if len(values) == 0 {
			return "Virheellinen arvo"
		}
		if len(values) == 1 {
			return fmt.Sprintf("Virheellinen syöte: täytyy olla %s", issues.StringifyPrimitive(values[0]))
		}
		return fmt.Sprintf("Virheellinen valinta: täytyy olla yksi seuraavista: %s",
			issues.JoinValuesWithSeparator(values, "|"))

	case core.TooBig:
		return formatSizeConstraintFi(raw, false)

	case core.TooSmall:
		return formatSizeConstraintFi(raw, true)

	case core.InvalidFormat:
		format := mapx.StringOr(raw.Properties, "format", "")
		if format == "" {
			return "Virheellinen muoto"
		}
		return formatStringValidationFi(raw, format)

	case core.NotMultipleOf:
		divisor := mapx.AnyOr(raw.Properties, "divisor", nil)
		if divisor == nil {
			return "Virheellinen luku: täytyy olla monikerta"
		}
		return fmt.Sprintf("Virheellinen luku: täytyy olla luvun %v monikerta", divisor)

	case core.UnrecognizedKeys:
		keys := mapx.StringsOr(raw.Properties, "keys", nil)
		if len(keys) == 0 {
			return "Tuntematon avain"
		}
		keyWord := "Tuntematon avain"
		if len(keys) > 1 {
			keyWord = "Tuntemattomat avaimet"
		}
		if keysAny, err := slicex.ToAny(keys); err == nil && !slicex.IsEmpty(keysAny) {
			keysJoined := issues.JoinValuesWithSeparator(keysAny, ", ")
			return fmt.Sprintf("%s: %s", keyWord, keysJoined)
		}
		return keyWord

	case core.InvalidKey:
		return "Virheellinen avain tietueessa"

	case core.InvalidUnion:
		return "Virheellinen unioni"

	case core.InvalidElement:
		return "Virheellinen arvo joukossa"

	case core.MissingRequired:
		fieldName := mapx.StringOr(raw.Properties, "field_name", "")
		fieldType := mapx.StringOr(raw.Properties, "field_type", "kenttä")
		if fieldName == "" {
			return fmt.Sprintf("Pakollinen %s puuttuu", fieldType)
		}
		return fmt.Sprintf("Pakollinen %s puuttuu: %s", fieldType, fieldName)

	case core.TypeConversion:
		fromType := mapx.StringOr(raw.Properties, "from_type", "tuntematon")
		toType := mapx.StringOr(raw.Properties, "to_type", "tuntematon")
		return fmt.Sprintf("Tyyppimuunnos epäonnistui: %s ei voida muuntaa tyypiksi %s", fromType, toType)

	case core.InvalidSchema:
		reason := mapx.StringOr(raw.Properties, "reason", "")
		if reason != "" {
			return fmt.Sprintf("Virheellinen skeema: %s", reason)
		}
		return "Virheellinen skeemamäärittely"

	case core.InvalidDiscriminator:
		field := mapx.StringOr(raw.Properties, "field", "erottelija")
		return fmt.Sprintf("Virheellinen tai puuttuva erottelukenttä: %s", field)

	case core.IncompatibleTypes:
		conflictType := mapx.StringOr(raw.Properties, "conflict_type", "arvot")
		return fmt.Sprintf("Ei voida yhdistää %s: yhteensopimattomat tyypit", conflictType)

	case core.NilPointer:
		return "Tyhjä osoitin havaittu"

	case core.Custom:
		message := mapx.StringOr(raw.Properties, "message", "")
		if message != "" {
			return message
		}
		return "Virheellinen syöte"

	default:
		return "Virheellinen syöte"
	}
}

// formatSizeConstraintFi formats size constraint messages in Finnish
func formatSizeConstraintFi(raw core.ZodRawIssue, isTooSmall bool) string {
	var threshold any
	if isTooSmall {
		threshold = mapx.AnyOr(raw.Properties, "minimum", nil)
	} else {
		threshold = mapx.AnyOr(raw.Properties, "maximum", nil)
	}

	if threshold == nil {
		if isTooSmall {
			return "Liian pieni"
		}
		return "Liian suuri"
	}

	inclusive := mapx.BoolOr(raw.Properties, "inclusive", true)
	sizing := getSizingFi(mapx.StringOr(raw.Properties, "origin", ""))
	thresholdStr := issues.FormatThreshold(threshold)

	// Finnish comparison operators
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
		subject := sizing.Subject
		if isTooSmall {
			if sizing.Unit != "" {
				return fmt.Sprintf("Liian pieni: %s täytyy olla %s%s %s", subject, adj, thresholdStr, sizing.Unit)
			}
			return fmt.Sprintf("Liian pieni: %s täytyy olla %s%s", subject, adj, thresholdStr)
		}
		if sizing.Unit != "" {
			return fmt.Sprintf("Liian suuri: %s täytyy olla %s%s %s", subject, adj, thresholdStr, sizing.Unit)
		}
		return fmt.Sprintf("Liian suuri: %s täytyy olla %s%s", subject, adj, thresholdStr)
	}

	if isTooSmall {
		return fmt.Sprintf("Liian pieni: arvon täytyy olla %s%s", adj, thresholdStr)
	}
	return fmt.Sprintf("Liian suuri: arvon täytyy olla %s%s", adj, thresholdStr)
}

// formatStringValidationFi handles string format validation messages in Finnish
func formatStringValidationFi(raw core.ZodRawIssue, format string) string {
	switch format {
	case "starts_with":
		prefix := mapx.StringOr(raw.Properties, "prefix", "")
		if prefix == "" {
			return "Virheellinen syöte: täytyy alkaa tietyllä merkkijonolla"
		}
		return fmt.Sprintf("Virheellinen syöte: täytyy alkaa \"%s\"", prefix)
	case "ends_with":
		suffix := mapx.StringOr(raw.Properties, "suffix", "")
		if suffix == "" {
			return "Virheellinen syöte: täytyy loppua tiettyyn merkkijonoon"
		}
		return fmt.Sprintf("Virheellinen syöte: täytyy loppua \"%s\"", suffix)
	case "includes":
		includes := mapx.StringOr(raw.Properties, "includes", "")
		if includes == "" {
			return "Virheellinen syöte: täytyy sisältää tietty merkkijono"
		}
		return fmt.Sprintf("Virheellinen syöte: täytyy sisältää \"%s\"", includes)
	case "regex":
		pattern := mapx.StringOr(raw.Properties, "pattern", "")
		if pattern == "" {
			return "Virheellinen syöte: täytyy vastata säännöllistä lauseketta"
		}
		return fmt.Sprintf("Virheellinen syöte: täytyy vastata säännöllistä lauseketta %s", pattern)
	default:
		noun := getFormatNounFi(format)
		return fmt.Sprintf("Virheellinen %s", noun)
	}
}

// Fi returns a ZodConfig configured for Finnish locale.
func Fi() *core.ZodConfig {
	return &core.ZodConfig{
		LocaleError: formatFi,
	}
}

// FormatMessageFi formats a single issue using Finnish locale
func FormatMessageFi(issue core.ZodRawIssue) string {
	return formatFi(issue)
}
