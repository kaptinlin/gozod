package regexes

import (
	"regexp"
)

// Emoji matches emoji characters (simplified Go implementation)
// TypeScript original code:
// export const _emoji = `^(\\p{Extended_Pictographic}|\\p{Emoji_Component})+$`;
//
//	export function emoji(): RegExp {
//	  return new RegExp(_emoji, "u");
//	}
//
// Note: Go's regexp doesn't support Unicode properties, so this is a simplified approximation
var Emoji = regexp.MustCompile(`^[\x{1F600}-\x{1F64F}|\x{1F300}-\x{1F5FF}|\x{1F680}-\x{1F6FF}|\x{1F700}-\x{1F77F}|\x{1F780}-\x{1F7FF}|\x{1F800}-\x{1F8FF}|\x{1F900}-\x{1F9FF}|\x{1FA00}-\x{1FA6F}|\x{1FA70}-\x{1FAFF}|\x{2600}-\x{26FF}|\x{2700}-\x{27BF}]+$`)

// JWT matches strings that follow the JWT format (header.payload.signature)
// Note: TypeScript Zod doesn't have a JWT regex export - this is a Go-specific utility
var JWT = regexp.MustCompile(`^[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+\.[A-Za-z0-9-_]*$`)
