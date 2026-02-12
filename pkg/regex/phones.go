package regex

import "regexp"

// E164 matches E.164 international phone numbers (e.g., "+14155552671").
var E164 = regexp.MustCompile(`^\+[1-9]\d{6,14}$`)
