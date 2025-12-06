package regexes

import (
	"regexp"
)

// IPv4 matches IPv4 addresses
// TypeScript original code:
// export const ipv4: RegExp =
//
//	/^(?:(?:25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9][0-9]|[0-9])\.){3}(?:25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9][0-9]|[0-9])$/;
var IPv4 = regexp.MustCompile(`^(?:(?:25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9][0-9]|[0-9])\.){3}(?:25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9][0-9]|[0-9])$`)

// IPv6 matches IPv6 addresses – comprehensive pattern covering multiple compressed forms.
// TypeScript original code:
// export const ipv6: RegExp =
//
//	/^(([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}|::|([0-9a-fA-F]{1,4})?::([0-9a-fA-F]{1,4}:?){0,6})$/;
var IPv6 = regexp.MustCompile(`^(([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))$`)

// CIDRv4 matches IPv4 CIDR addresses
// TypeScript original code:
// export const cidrv4: RegExp =
//
//	/^((25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9][0-9]|[0-9])\.){3}(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9][0-9]|[0-9])\/([0-9]|[1-2][0-9]|3[0-2])$/;
var CIDRv4 = regexp.MustCompile(`^((25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9][0-9]|[0-9])\.){3}(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9][0-9]|[0-9])\/([0-9]|[1-2][0-9]|3[0-2])$`)

// CIDRv6 matches IPv6 CIDR addresses
// Updated regex to handle more IPv6 address formats with CIDR notation
var CIDRv6 = regexp.MustCompile(`^(([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))\/(12[0-8]|1[01][0-9]|[1-9]?[0-9])$`)

// IP matches any IP address (IPv4 or IPv6)
// TypeScript original code:
// export const ip: RegExp = new RegExp(`(${ipv4.source})|(${ipv6.source})`);
var IP = regexp.MustCompile(`((?:(?:25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9][0-9]|[0-9])\.){3}(?:25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9][0-9]|[0-9]))|((?:(?:[0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}|::|(?:[0-9a-fA-F]{1,4})?::(?:[0-9a-fA-F]{1,4}:?){0,6}))`)

// Hostname matches DNS hostnames
// TypeScript original code:
// export const hostname: RegExp = /^([a-zA-Z0-9-]+\.)*[a-zA-Z0-9-]+$/;
var Hostname = regexp.MustCompile(`^([a-zA-Z0-9-]+\.)*[a-zA-Z0-9-]+$`)

// Domain matches valid domain names with TLD
// TypeScript original code:
// export const domain: RegExp = /^([a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$/;
var Domain = regexp.MustCompile(`^([a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)

// MAC returns a regex for validating MAC addresses with the specified delimiter.
// Matches Zod TypeScript implementation using two branches for case-sensitive matching.
// TypeScript original code:
//
//	export const mac = (delimiter?: string): RegExp => {
//	  const escapedDelim = util.escapeRegex(delimiter ?? ":");
//	  return new RegExp(`^(?:[0-9A-F]{2}${escapedDelim}){5}[0-9A-F]{2}$|^(?:[0-9a-f]{2}${escapedDelim}){5}[0-9a-f]{2}$`);
//	};
//
// The regex uses two alternatives:
// - First branch: uppercase hex digits (0-9A-F)
// - Second branch: lowercase hex digits (0-9a-f)
// This approach provides case-sensitive matching while accepting both cases.
//
// Examples:
//
//	MAC(":")  matches "00:1A:2B:3C:4D:5E" or "00:1a:2b:3c:4d:5e"
//	MAC("-")  matches "00-1A-2B-3C-4D-5E" or "00-1a-2b-3c-4d-5e"
//	MAC(".")  matches "00.1A.2B.3C.4D.5E" or "00.1a.2b.3c.4d.5e"
func MAC(delimiter string) *regexp.Regexp {
	if delimiter == "" {
		delimiter = ":"
	}

	// Escape special regex characters in the delimiter (e.g., "." becomes "\.")
	escaped := regexp.QuoteMeta(delimiter)

	// Build pattern with two branches for uppercase and lowercase hex
	// Format: ^(?:[0-9A-F]{2}DELIM){5}[0-9A-F]{2}$|^(?:[0-9a-f]{2}DELIM){5}[0-9a-f]{2}$
	pattern := "^(?:[0-9A-F]{2}" + escaped + "){5}[0-9A-F]{2}$|^(?:[0-9a-f]{2}" + escaped + "){5}[0-9a-f]{2}$"

	return regexp.MustCompile(pattern)
}

// MACDefault is the default MAC address regex using colon as delimiter
var MACDefault = MAC(":")
