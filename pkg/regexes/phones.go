package regexes

import "regexp"

// E164 matches E.164 international phone number standard
// TypeScript original code:
// export const e164: RegExp = /^\+(?:[0-9]){6,14}[0-9]$/;
var E164 = regexp.MustCompile(`^\+(?:[0-9]){6,14}[0-9]$`)
