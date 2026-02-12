package regex

import (
	"regexp"
	"strconv"
	"sync"
)

// stringLengthKey identifies a cached string-length regex by its bounds.
type stringLengthKey struct {
	min, max int
}

var (
	stringMu    sync.Mutex
	stringCache = make(map[stringLengthKey]*regexp.Regexp)
)

// StringRegex returns a cached regex matching strings with length in [min, max].
// A non-positive max means no upper bound.
func StringRegex(min, max int) *regexp.Regexp {
	k := stringLengthKey{min, max}

	stringMu.Lock()
	defer stringMu.Unlock()

	if re, ok := stringCache[k]; ok {
		return re
	}

	var pattern string
	if max > 0 {
		pattern = `^[\s\S]{` + strconv.Itoa(min) + `,` + strconv.Itoa(max) + `}$`
	} else {
		pattern = `^[\s\S]{` + strconv.Itoa(min) + `,}$`
	}

	re := regexp.MustCompile(pattern)
	stringCache[k] = re
	return re
}

// Primitive type patterns.
var (
	// String matches any string with no length restrictions.
	String = regexp.MustCompile(`^[\s\S]*$`)

	// Bigint matches big integers with optional trailing 'n' (e.g., "123n").
	Bigint = regexp.MustCompile(`^-?\d+n?$`)

	// Integer matches integers including negative numbers.
	Integer = regexp.MustCompile(`^-?\d+$`)

	// Number matches numbers including decimals and negative numbers.
	Number = regexp.MustCompile(`^-?\d+(?:\.\d+)?$`)

	// Boolean matches boolean values (case-insensitive).
	Boolean = regexp.MustCompile(`(?i)^(true|false)$`)

	// Null matches "null" (case-insensitive).
	Null = regexp.MustCompile(`(?i)^null$`)

	// Undefined matches "undefined" (case-insensitive).
	Undefined = regexp.MustCompile(`(?i)^undefined$`)

	// Lowercase matches strings containing no uppercase letters.
	Lowercase = regexp.MustCompile(`^[^A-Z]*$`)

	// Uppercase matches strings containing no lowercase letters.
	Uppercase = regexp.MustCompile(`^[^a-z]*$`)

	// JSONString matches any string (simplistic; actual validation should be done at runtime).
	JSONString = regexp.MustCompile(`^[\s\S]*$`)
)
