package validate

import (
	"regexp"

	"github.com/kaptinlin/gozod/pkg/coerce"
	"github.com/kaptinlin/gozod/pkg/reflectx"
)

// matchString extracts a string from value and checks it against pattern.
func matchString(value any, pattern *regexp.Regexp) bool {
	str, ok := reflectx.StringVal(value)
	if !ok {
		return false
	}
	return pattern.MatchString(str)
}

// toFloat64 converts any value to float64 using the coerce package.
func toFloat64(value any) float64 {
	result, err := coerce.ToFloat64(value)
	if err != nil {
		return 0
	}
	return result
}

// collectionSize returns the size of a collection (map, slice, array, string).
func collectionSize(value any) (int, bool) {
	switch m := value.(type) {
	case map[string]any:
		return len(m), true
	case map[any]any:
		return len(m), true
	default:
		if size, ok := reflectx.Size(value); ok {
			return size, true
		}
		return reflectx.Length(value)
	}
}
