package regex

import (
	"regexp"
	"strconv"
	"strings"
)

// Duration matches ISO 8601-1 duration format.
// Matches P<weeks>W or P<years>Y<months>M<days>D(T<hours>H<minutes>M<seconds>S).
var Duration = regexp.MustCompile(`^P(?:(\d+W)|(\d+Y)?(\d+M)?(\d+D)?(?:T(\d+H)?(\d+M)?(\d+(?:[.,]\d+)?S)?)?)$`)

// ExtendedDuration matches ISO 8601-2 extended duration format (simplified for Go).
var ExtendedDuration = regexp.MustCompile(`^[-+]?P(?:[-+]?\d+[.,]?\d*[YMWD])*(?:T(?:[-+]?\d+[.,]?\d*[HMS])*)?$`)

// datePattern is the base date pattern component for ISO 8601 dates.
const datePattern = `(?:(?:\d\d[2468][048]|\d\d[13579][26]|\d\d0[48]|[02468][048]00|[13579][26]00)-02-29|\d{4}-(?:(?:0[13578]|1[02])-(?:0[1-9]|[12]\d|3[01])|(?:0[469]|11)-(?:0[1-9]|[12]\d|30)|(?:02)-(?:0[1-9]|1\d|2[0-8])))`

// Date matches ISO 8601 date format (YYYY-MM-DD) with leap year validation.
var Date = regexp.MustCompile(`^` + datePattern + `$`)

// TimeOptions defines parameters for the time regex pattern.
type TimeOptions struct {
	// Precision specifies the number of decimal places for seconds.
	// nil: any number of decimal places; 0: no decimals; -1: minute precision only.
	Precision *int
}

// timePattern returns the time regex pattern string based on precision.
func timePattern(precision *int) string {
	if precision == nil {
		return `(?:[01]\d|2[0-3]):[0-5]\d(?::[0-5]\d(?:\.\d+)?)?`
	}

	switch p := *precision; {
	case p == -1:
		return `(?:[01]\d|2[0-3]):[0-5]\d`
	case p == 0:
		return `(?:[01]\d|2[0-3]):[0-5]\d:[0-5]\d`
	case p > 0:
		return `(?:[01]\d|2[0-3]):[0-5]\d:[0-5]\d\.\d{` + strconv.Itoa(p) + `}`
	default:
		return `(?:[01]\d|2[0-3]):[0-5]\d(?::[0-5]\d)?`
	}
}

// Time returns a regex for matching ISO 8601 time format.
func Time(opts TimeOptions) *regexp.Regexp {
	return regexp.MustCompile(`^` + timePattern(opts.Precision) + `$`)
}

// DefaultTime matches ISO 8601 time (HH:MM or HH:MM:SS with optional fractional seconds).
var DefaultTime = regexp.MustCompile(`^(?:[01]\d|2[0-3]):[0-5]\d(?::[0-5]\d(?:\.\d+)?)?$`)

// DatetimeOptions defines parameters for the datetime regex pattern.
type DatetimeOptions struct {
	// Precision specifies the number of decimal places for seconds.
	// nil: any number of decimal places; 0: no decimals; -1: minute precision only.
	Precision *int

	// Offset allows timezone offsets like +01:00 when true.
	Offset bool

	// Local makes the 'Z' timezone marker optional when true.
	Local bool
}

// Datetime returns a regex for matching ISO 8601 datetime format.
func Datetime(options DatetimeOptions) *regexp.Regexp {
	pat := datePattern + `T` + timePattern(options.Precision)

	var suffixes []string
	if options.Local {
		suffixes = append(suffixes, `Z?`)
	} else {
		suffixes = append(suffixes, `Z`)
	}

	if options.Offset {
		suffixes = append(suffixes, `([+-](?:[01]\d|2[0-3]):[0-5]\d)`)
	}

	pat += `(` + strings.Join(suffixes, "|") + `)`

	return regexp.MustCompile(`^` + pat + `$`)
}

// DefaultDatetime matches ISO 8601 datetime with Z or offset timezone.
var DefaultDatetime = regexp.MustCompile(`^` + datePattern + `T(?:` + timePattern(nil) + `(?:Z|[+-](?:[01]\d|2[0-3]):[0-5]\d))$`)
