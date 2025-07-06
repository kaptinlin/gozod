package regexes

import "regexp"

// Emoji matches emoji characters (Go-compatible implementation)
// Note: Go's regexp doesn't support Unicode properties like \p{Extended_Pictographic}
// This pattern covers comprehensive emoji ranges based on Unicode standards
var Emoji = regexp.MustCompile(`^[` +
	// Basic emoji ranges (core emoji blocks)
	`\x{1F600}-\x{1F64F}` + // Emoticons (smileys & emotion)
	`\x{1F300}-\x{1F5FF}` + // Miscellaneous Symbols and Pictographs
	`\x{1F680}-\x{1F6FF}` + // Transport and Map Symbols
	`\x{1F700}-\x{1F77F}` + // Alchemical Symbols
	`\x{1F780}-\x{1F7FF}` + // Geometric Shapes Extended
	`\x{1F800}-\x{1F8FF}` + // Supplemental Arrows-C
	`\x{1F900}-\x{1F9FF}` + // Supplemental Symbols and Pictographs
	`\x{1FA00}-\x{1FA6F}` + // Chess Symbols
	`\x{1FA70}-\x{1FAFF}` + // Symbols and Pictographs Extended-A

	// Symbol ranges
	`\x{2600}-\x{26FF}` + // Miscellaneous Symbols
	`\x{2700}-\x{27BF}` + // Dingbats

	// Regional indicators (flags)
	`\x{1F1E6}-\x{1F1FF}` + // Regional Indicator Symbols

	// CJK and other symbol ranges
	`\x{3000}-\x{303F}` + // CJK Symbols and Punctuation
	`\x{3200}-\x{32FF}` + // Enclosed CJK Letters and Months

	// Specific emoji characters
	`\x{1F004}` + // Mahjong Tile Red Dragon
	`\x{1F0CF}` + // Playing Card Black Joker
	`\x{1F18E}` + // Negative Squared AB
	`\x{1F191}-\x{1F19A}` + // Squared symbols
	`\x{1F201}` + // Squared Katakana Koko
	`\x{1F21A}` + // Squared CJK Unified Ideograph-7121
	`\x{1F22F}` + // Squared CJK Unified Ideograph-6307
	`\x{1F232}-\x{1F236}` + // Squared CJK symbols
	`\x{1F238}-\x{1F23A}` + // Squared CJK symbols
	`\x{1F250}` + // Circled Ideograph Advantage
	`\x{1F251}` + // Circled Ideograph Accept

	// Modifiers and combining characters
	`\x{1F3FB}-\x{1F3FF}` + // Emoji Modifier Fitzpatrick Type
	`\x{200D}` + // Zero Width Joiner (ZWJ)
	`\x{FE0F}` + // Variation Selector-16 (emoji presentation)
	`]+$`)
