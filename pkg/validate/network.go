package validate

import (
	"net"
	"net/url"
	"regexp"
	"slices"
	"strings"

	"github.com/kaptinlin/gozod/pkg/reflectx"
	"github.com/kaptinlin/gozod/pkg/regex"
)

// URL reports whether the string is a valid URL format.
func URL(value any) bool { return matchString(value, regex.URL) }

// Hostname reports whether the string is a valid DNS hostname.
func Hostname(value any) bool {
	str, ok := reflectx.StringVal(value)
	if !ok || len(str) == 0 || len(str) > 253 {
		return false
	}
	return regex.Hostname.MatchString(str)
}

// URLWithOptions reports whether the string is a valid URL with the given constraints.
func URLWithOptions(value any, options URLOptions) bool {
	str, ok := reflectx.StringVal(value)
	if !ok {
		return false
	}
	if !regex.URL.MatchString(str) {
		return false
	}
	if options.Hostname == nil && options.Protocol == nil {
		return true
	}
	u, err := url.Parse(str)
	if err != nil {
		return false
	}
	if options.Hostname != nil && !options.Hostname.MatchString(u.Hostname()) {
		return false
	}
	if options.Protocol != nil && !options.Protocol.MatchString(u.Scheme) {
		return false
	}
	return true
}

// UUID reports whether the string is a valid UUID format.
func UUID(value any) bool { return matchString(value, regex.UUID) }

// GUID reports whether the string is a valid GUID format.
func GUID(value any) bool { return matchString(value, regex.GUID) }

// CUID reports whether the string is a valid CUID format.
func CUID(value any) bool { return matchString(value, regex.CUID) }

// CUID2 reports whether the string is a valid CUID2 format.
func CUID2(value any) bool { return matchString(value, regex.CUID2) }

// NanoID reports whether the string is a valid NanoID format.
func NanoID(value any) bool { return matchString(value, regex.NanoID) }

// ULID reports whether the string is a valid ULID format.
func ULID(value any) bool { return matchString(value, regex.ULID) }

// XID reports whether the string is a valid XID format.
func XID(value any) bool { return matchString(value, regex.XID) }

// KSUID reports whether the string is a valid KSUID format.
func KSUID(value any) bool { return matchString(value, regex.KSUID) }

// IPv4 reports whether the string is a valid IPv4 address.
func IPv4(value any) bool { return matchString(value, regex.IPv4) }

// IPv6 reports whether the string is a valid IPv6 address.
func IPv6(value any) bool { return matchString(value, regex.IPv6) }

// CIDR reports whether the string is valid CIDR notation.
func CIDR(value any, version int) bool {
	str, ok := reflectx.StringVal(value)
	if !ok {
		return false
	}
	if !strings.Contains(str, "/") {
		return false
	}
	_, ipnet, err := net.ParseCIDR(str)
	if err != nil {
		return false
	}
	if version == 0 {
		return true
	}
	isV4 := ipnet.IP.To4() != nil
	return (version == 4 && isV4) || (version == 6 && !isV4)
}

// CIDRv4 reports whether the string is a valid IPv4 CIDR notation.
func CIDRv4(value any) bool { return CIDR(value, 4) }

// CIDRv6 reports whether the string is a valid IPv6 CIDR notation.
func CIDRv6(value any) bool { return CIDR(value, 6) }

// MAC reports whether the string is a valid MAC address using the default delimiter.
func MAC(value any) bool {
	return MACWithOptions(value, MACOptions{Delimiter: ":"})
}

// MACWithOptions reports whether the string is a valid MAC address with a custom delimiter.
func MACWithOptions(value any, opts MACOptions) bool {
	str, ok := reflectx.StringVal(value)
	if !ok {
		return false
	}
	delim := opts.Delimiter
	if delim == "" {
		delim = ":"
	}
	return regex.MAC(delim).MatchString(str)
}

// Mime reports whether the value is one of the allowed MIME types.
func Mime(value any, allowedTypes []string) bool {
	str, ok := reflectx.StringVal(value)
	if !ok {
		return false
	}
	typePart, subtype, hasSep := strings.Cut(str, "/")
	if !hasSep || typePart == "" || subtype == "" {
		return false
	}
	return slices.Contains(allowedTypes, str)
}

var _ = regexp.MustCompile
