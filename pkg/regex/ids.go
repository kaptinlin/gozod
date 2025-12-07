package regex

import (
	"regexp"
	"strconv"
)

// CUID matches strings in CUID format
// TypeScript original code:
// export const cuid: RegExp = /^[cC][^\s-]{8,}$/;
var CUID = regexp.MustCompile(`^[cC][^\s-]{8,}$`)

// CUID2 matches strings in CUID2 format
// TypeScript original code:
// export const cuid2: RegExp = /^[0-9a-z]+$/;
var CUID2 = regexp.MustCompile(`^[0-9a-z]+$`)

// ULID matches strings in ULID format
// TypeScript original code:
// export const ulid: RegExp = /^[0-9A-HJKMNP-TV-Za-hjkmnp-tv-z]{26}$/;
var ULID = regexp.MustCompile(`^[0-9A-HJKMNP-TV-Za-hjkmnp-tv-z]{26}$`)

// XID matches strings in XID format
// TypeScript original code:
// export const xid: RegExp = /^[0-9a-vA-V]{20}$/;
var XID = regexp.MustCompile(`^[0-9a-vA-V]{20}$`)

// KSUID matches strings in KSUID format
// TypeScript original code:
// export const ksuid: RegExp = /^[A-Za-z0-9]{27}$/;
var KSUID = regexp.MustCompile(`^[A-Za-z0-9]{27}$`)

// NanoID matches strings in NanoID format
// TypeScript original code:
// export const nanoid: RegExp = /^[a-zA-Z0-9_-]{21}$/;
var NanoID = regexp.MustCompile(`^[a-zA-Z0-9_-]{21}$`)

// GUID matches any UUID-like identifier with 8-4-4-4-12 hex pattern
// TypeScript original code:
// export const guid: RegExp = /^([0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12})$/;
var GUID = regexp.MustCompile(`^([0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12})$`)

// UUID matches RFC 4122 UUID with all versions supported when no version is specified
// TypeScript original code:
//
//	export const uuid = (version?: number | undefined): RegExp => {
//	  if (!version)
//	    return /^([0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-8][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}|00000000-0000-0000-0000-000000000000)$/;
//	  return new RegExp(
//	    `^([0-9a-fA-F]{8}-[0-9a-fA-F]{4}-${version}[0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12})$`
//	  );
//	};
var UUID = regexp.MustCompile(`^([0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-8][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}|00000000-0000-0000-0000-000000000000)$`)

// UUID4 matches version 4 UUIDs
// TypeScript original code:
// export const uuid4: RegExp = uuid(4);
var UUID4 = regexp.MustCompile(`^([0-9a-fA-F]{8}-[0-9a-fA-F]{4}-4[0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12})$`)

// UUID6 matches version 6 UUIDs
// TypeScript original code:
// export const uuid6: RegExp = uuid(6);
var UUID6 = regexp.MustCompile(`^([0-9a-fA-F]{8}-[0-9a-fA-F]{4}-6[0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12})$`)

// UUID7 matches version 7 UUIDs
// TypeScript original code:
// export const uuid7: RegExp = uuid(7);
var UUID7 = regexp.MustCompile(`^([0-9a-fA-F]{8}-[0-9a-fA-F]{4}-7[0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12})$`)

// UUIDForVersion returns a regex matching a specific UUID version (1-8)
// TypeScript original code:
//
//	export const uuid = (version?: number | undefined): RegExp => {
//	  if (!version)
//	    return /^([0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-8][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}|00000000-0000-0000-0000-000000000000)$/;
//	  return new RegExp(
//	    `^([0-9a-fA-F]{8}-[0-9a-fA-F]{4}-${version}[0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12})$`
//	  );
//	};
func UUIDForVersion(version int) *regexp.Regexp {
	if version < 1 || version > 8 {
		return UUID
	}
	return regexp.MustCompile(`^([0-9a-fA-F]{8}-[0-9a-fA-F]{4}-` + strconv.Itoa(version) + `[0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12})$`)
}
