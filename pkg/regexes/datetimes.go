package regexes

import (
	"fmt"
	"regexp"
	"strings"
)

// Duration matches ISO 8601-1 duration regex simplified for Go (without lookaheads)
// TypeScript original code:
// export const duration: RegExp =
//
//	/^P(?:(\d+W)|(?!.*W)(?=\d|T\d)(\d+Y)?(\d+M)?(\d+D)?(T(?=\d)(\d+H)?(\d+M)?(\d+([.,]\d+)?S)?)?)$/;
//
// Note: Go's regexp doesn't support lookaheads, so this is a simplified version
// This version matches the basic structure but needs additional validation
var Duration = regexp.MustCompile(`^P(?:\d+W|(?:\d+Y)?(?:\d+M)?(?:\d+D)?(?:T(?:\d+H)?(?:\d+M)?(?:\d+(?:[.,]\d+)?S)?)?)$`)

// ExtendedDuration implements ISO 8601-2 extensions simplified for Go
// TypeScript original code:
// export const extendedDuration: RegExp =
//
//	/^[-+]?P(?!$)(?:(?:[-+]?\d+Y)|(?:[-+]?\d+[.,]\d+Y$))?(?:(?:[-+]?\d+M)|(?:[-+]?\d+[.,]\d+M$))?(?:(?:[-+]?\d+W)|(?:[-+]?\d+[.,]\d+W$))?(?:(?:[-+]?\d+D)|(?:[-+]?\d+[.,]\d+D$))?(?:T(?=[\d+-])(?:(?:[-+]?\d+H)|(?:[-+]?\d+[.,]\d+H$))?(?:(?:[-+]?\d+M)|(?:[-+]?\d+[.,]\d+M$))?(?:[-+]?\d+(?:[.,]\d+)?S)?)??$/;
//
// Note: Go's regexp doesn't support lookaheads, so this is a simplified version
var ExtendedDuration = regexp.MustCompile(`^[-+]?P([-+]?\d+[.,]?\d*[YMWD]|T[-+]?\d+[.,]?\d*[HMS])+$`)

// dateSource provides date pattern base component
// TypeScript original code:
// const dateSource = `((\\d\\d[2468][048]|\\d\\d[13579][26]|\\d\\d0[48]|[02468][048]00|[13579][26]00)-02-29|\\d{4}-((0[13578]|1[02])-(0[1-9]|[12]\\d|3[01])|(0[469]|11)-(0[1-9]|[12]\\d|30)|(02)-(0[1-9]|1\\d|2[0-8])))`;
var dateSource = `((\d\d[2468][048]|\d\d[13579][26]|\d\d0[48]|[02468][048]00|[13579][26]00)-02-29|\d{4}-((0[13578]|1[02])-(0[1-9]|[12]\d|3[01])|(0[469]|11)-(0[1-9]|[12]\d|30)|(02)-(0[1-9]|1\d|2[0-8])))`

// Date matches ISO 8601 date format (YYYY-MM-DD)
// TypeScript original code:
// export const date: RegExp = new RegExp(`^${dateSource}$`);
var Date = regexp.MustCompile(`^` + dateSource + `$`)

// TimeOptions defines parameters for time regex pattern
// TypeScript original code:
//
//	function timeSource(args: { precision?: number | null }) {
//	  let regex = `([01]\\d|2[0-3]):[0-5]\\d:[0-5]\\d`;
//	  if (args.precision) {
//	    regex = `${regex}\\.\\d{${args.precision}}`;
//	  } else if (args.precision == null) {
//	    regex = `${regex}(\\.\\d+)?`;
//	  }
//	  return regex;
//	}
type TimeOptions struct {
	// Precision specifies number of decimal places for seconds
	// If nil, matches any number of decimal places
	// If 0 or negative, no decimal places allowed
	Precision *int
}

// timeSource creates time regex pattern based on precision options
// TypeScript original code:
//
//	function timeSource(args: { precision?: number | null }) {
//	  let regex = `([01]\\d|2[0-3]):[0-5]\\d:[0-5]\\d`;
//	  if (args.precision) {
//	    regex = `${regex}\\.\\d{${args.precision}}`;
//	  } else if (args.precision == null) {
//	    regex = `${regex}(\\.\\d+)?`;
//	  }
//	  return regex;
//	}
func timeSource(precision *int) string {
	// Base pattern supports both HH:MM and HH:MM:SS formats
	// TypeScript Zod 4 supports: "03:15", "03:15:00", "03:15:00.9999999"
	regex := `([01]\d|2[0-3]):[0-5]\d(:[0-5]\d)?`

	if precision == nil {
		// Allow optional seconds and optional fractional seconds
		return fmt.Sprintf(`%s(\.\d+)?`, regex)
	}

	if *precision == -1 {
		// Minute precision only (HH:MM) - no seconds allowed
		return `([01]\d|2[0-3]):[0-5]\d`
	}

	if *precision == 0 {
		// Second precision required (HH:MM:SS) - no fractional seconds
		return `([01]\d|2[0-3]):[0-5]\d:[0-5]\d`
	}

	if *precision > 0 {
		// Specific fractional precision required (HH:MM:SS.sss)
		return fmt.Sprintf(`([01]\d|2[0-3]):[0-5]\d:[0-5]\d\.\d{%d}`, *precision)
	}

	return regex
}

// Time returns a regex for matching ISO 8601 time format
// TypeScript original code:
//
//	export function time(args: {
//	  precision?: number | null;
//	}): RegExp {
//
//	  return new RegExp(`^${timeSource(args)}$`);
//	}
func Time(opts TimeOptions) *regexp.Regexp {
	pattern := `^` + timeSource(opts.Precision) + `$`
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		// Fallback to default time pattern if compilation fails
		return DefaultTime
	}
	return compiled
}

// DefaultTime is the time regex with any decimal precision
// Updated to support both HH:MM and HH:MM:SS formats like TypeScript Zod 4
var DefaultTime = regexp.MustCompile(`^` + timeSource(nil) + `$`)

// DatetimeOptions defines parameters for datetime regex pattern
// TypeScript original code:
//
//	export function datetime(args: {
//	  precision?: number | null;
//	  offset?: boolean;
//	  local?: boolean;
//	}): RegExp
type DatetimeOptions struct {
	// Precision specifies number of decimal places for seconds
	// If nil, matches any number of decimal places
	// If 0 or negative, no decimal places allowed
	Precision *int

	// Offset if true, allows timezone offsets like +01:00
	Offset bool

	// Local if true, makes the 'Z' timezone marker optional
	Local bool
}

// Datetime returns a regex for matching ISO 8601 datetime format
// TypeScript original code:
//
//	export function datetime(args: {
//	  precision?: number | null;
//	  offset?: boolean;
//	  local?: boolean;
//	}): RegExp {
//
//	  let regex = `${dateSource}T${timeSource(args)}`;
//	  const opts: string[] = [];
//	  opts.push(args.local ? `Z?` : `Z`);
//	  if (args.offset) opts.push(`([+-]\\d{2}:?\\d{2})`);
//	  regex = `${regex}(${opts.join("|")})`;
//	  return new RegExp(`^${regex}$`);
//	}
func Datetime(options DatetimeOptions) *regexp.Regexp {
	regex := dateSource + `T` + timeSource(options.Precision)

	// Handle timezone offset options
	var opts []string
	if options.Local {
		opts = append(opts, `Z?`)
	} else {
		opts = append(opts, `Z`)
	}

	if options.Offset {
		// TypeScript Zod 4 requires colon in offset format: +02:00, not +0200 or +02
		opts = append(opts, `([+-]\d{2}:\d{2})`)
	}

	if len(opts) > 0 {
		regex = fmt.Sprintf(`%s(%s)`, regex, strings.Join(opts, "|"))
	}

	pattern := `^` + regex + `$`
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		// Fallback to default datetime pattern if compilation fails
		return DefaultDatetime
	}
	return compiled
}

// DefaultDatetime is the datetime regex with Z timezone only (no offsets by default)
// Updated to match TypeScript Zod 4 default behavior: only Z allowed, no offsets
var DefaultDatetime = regexp.MustCompile(`^` + dateSource + `T` + timeSource(nil) + `Z$`)
