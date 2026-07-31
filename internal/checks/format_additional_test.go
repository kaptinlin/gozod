package checks

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/pkg/validate"
)

func TestFormatCheckConstructors_ValidateAndAnnotate(t *testing.T) {
	t.Parallel()

	emailPattern := regexp.MustCompile(`^[^@]+@example\.com$`)
	hostnamePattern := regexp.MustCompile(`^api\.example\.com$`)
	protocolPattern := regexp.MustCompile(`^https$`)
	secondPrecision := 0

	tests := []struct {
		name        string
		check       core.ZodCheck
		valid       any
		invalid     any
		wantCode    core.IssueCode
		wantBag     map[string]any
		wantPattern bool
	}{
		{
			name:        "email with custom pattern",
			check:       EmailWithPattern(emailPattern),
			valid:       "admin@example.com",
			invalid:     "admin@example.org",
			wantCode:    core.InvalidFormat,
			wantBag:     map[string]any{"type": "string"},
			wantPattern: true,
		},
		{
			name:        "url with options",
			check:       URLWithOptions(validate.URLOptions{Hostname: hostnamePattern, Protocol: protocolPattern}),
			valid:       "https://api.example.com/v1",
			invalid:     "http://api.example.com/v1",
			wantCode:    core.InvalidFormat,
			wantBag:     map[string]any{"type": "string", "hostnamePattern": hostnamePattern.String(), "protocolPattern": protocolPattern.String()},
			wantPattern: true,
		},
		{name: "ipv4", check: IPv4(), valid: "192.168.0.1", invalid: "999.999.999.999", wantCode: core.InvalidFormat, wantBag: map[string]any{"type": "string"}, wantPattern: true},
		{name: "ipv6", check: IPv6(), valid: "2001:db8::1", invalid: "192.168.0.1", wantCode: core.InvalidFormat, wantBag: map[string]any{"type": "string"}, wantPattern: true},
		{name: "hostname", check: Hostname(), valid: "example.com", invalid: "-example.com", wantCode: core.InvalidFormat, wantBag: map[string]any{"type": "string"}, wantPattern: true},
		{name: "mac default delimiter", check: MAC(), valid: "00:1A:2B:3C:4D:5E", invalid: "00-1A-2B-3C-4D-5E", wantCode: core.InvalidFormat, wantBag: map[string]any{"type": "string"}, wantPattern: true},
		{name: "mac custom delimiter", check: MACWithDelimiter("-"), valid: "00-1A-2B-3C-4D-5E", invalid: "00:1A:2B:3C:4D:5E", wantCode: core.InvalidFormat, wantBag: map[string]any{"type": "string"}, wantPattern: true},
		{name: "cidrv4", check: CIDRv4(), valid: "192.168.0.0/24", invalid: "192.168.0.1", wantCode: core.InvalidFormat, wantBag: map[string]any{"type": "string"}, wantPattern: true},
		{name: "cidrv6", check: CIDRv6(), valid: "2001:db8::/32", invalid: "2001:db8::1", wantCode: core.InvalidFormat, wantBag: map[string]any{"type": "string"}, wantPattern: true},
		{name: "base64", check: Base64(), valid: "SGVsbG8=", invalid: "!!!", wantCode: core.InvalidFormat, wantBag: map[string]any{"type": "string", "contentEncoding": "base64"}, wantPattern: true},
		{name: "base64url", check: Base64URL(), valid: "SGVsbG8", invalid: "!!!", wantCode: core.InvalidFormat, wantBag: map[string]any{"type": "string", "contentEncoding": "base64url"}, wantPattern: true},
		{name: "jwt", check: JWT(), valid: testJWT("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"), invalid: "not-a-token", wantCode: core.InvalidFormat, wantBag: map[string]any{"format": "jwt", "type": "string"}},
		{name: "jwt with algorithm", check: JWTWithAlgorithm("HS256"), valid: testJWT("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"), invalid: testJWT("eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9"), wantCode: core.InvalidFormat, wantBag: map[string]any{"format": "jwt", "type": "string"}},
		{name: "e164", check: E164(), valid: "+14155552671", invalid: "4155552671", wantCode: core.InvalidFormat, wantBag: map[string]any{"type": "string"}, wantPattern: true},
		{name: "iso datetime", check: ISODateTime(), valid: "2024-02-29T12:30:45Z", invalid: "2024-02-29T12:30:45", wantCode: core.InvalidFormat, wantBag: map[string]any{"type": "string"}, wantPattern: true},
		{name: "iso datetime options", check: ISODateTimeWithOptions(validate.ISODateTimeOptions{Local: true}), valid: "2024-02-29T12:30:45", invalid: "not-a-date", wantCode: core.InvalidFormat, wantBag: map[string]any{"type": "string"}, wantPattern: true},
		{name: "iso date", check: ISODate(), valid: "2024-02-29", invalid: "2023-02-29", wantCode: core.InvalidFormat, wantBag: map[string]any{"type": "string"}, wantPattern: true},
		{name: "iso time options", check: ISOTimeWithOptions(validate.ISOTimeOptions{Precision: &secondPrecision}), valid: "12:30:45", invalid: "12:30:45.1", wantCode: core.InvalidFormat, wantBag: map[string]any{"type": "string"}, wantPattern: true},
		{name: "iso time", check: ISOTime(), valid: "12:30:45.123", invalid: "25:00:00", wantCode: core.InvalidFormat, wantBag: map[string]any{"type": "string"}, wantPattern: true},
		{name: "iso duration", check: ISODuration(), valid: "P1Y2M3DT4H5M6S", invalid: "PT", wantCode: core.InvalidFormat, wantBag: map[string]any{"type": "string"}, wantPattern: true},
		{name: "cuid", check: CUID(), valid: "cjld2cyuq0000t3rmniod1foy", invalid: "short", wantCode: core.InvalidFormat, wantBag: map[string]any{"type": "string"}, wantPattern: true},
		{name: "cuid2", check: CUID2(), valid: "abc123", invalid: "ABC123", wantCode: core.InvalidFormat, wantBag: map[string]any{"type": "string"}, wantPattern: true},
		{name: "ulid", check: ULID(), valid: "01ARZ3NDEKTSV4RRFFQ69G5FAV", invalid: "not-a-ulid", wantCode: core.InvalidFormat, wantBag: map[string]any{"type": "string"}, wantPattern: true},
		{name: "xid", check: XID(), valid: "9m4e2mr0ui3e8a215n4g", invalid: "not-an-xid", wantCode: core.InvalidFormat, wantBag: map[string]any{"type": "string"}, wantPattern: true},
		{name: "ksuid", check: KSUID(), valid: "0ujtsYcgvSTl8PAuAdqWYSMnLOv", invalid: "not-a-ksuid", wantCode: core.InvalidFormat, wantBag: map[string]any{"type": "string"}, wantPattern: true},
		{name: "nanoid", check: NanoID(), valid: "V1StGXR8_Z5jdHi6B-myT", invalid: "too-short", wantCode: core.InvalidFormat, wantBag: map[string]any{"type": "string"}, wantPattern: true},
		{name: "json", check: JSON(), valid: `{"ok": true}`, invalid: `{"ok":`, wantCode: core.InvalidFormat, wantBag: map[string]any{"type": "string", "contentMediaType": "application/json"}, wantPattern: true},
		{name: "emoji", check: Emoji(), valid: "😀", invalid: "text", wantCode: core.InvalidFormat, wantBag: map[string]any{"type": "string"}, wantPattern: true},
		{name: "uuid", check: UUID(), valid: "550e8400-e29b-41d4-a716-446655440000", invalid: "not-a-uuid", wantCode: core.InvalidFormat, wantBag: map[string]any{"type": "string"}, wantPattern: true},
		{name: "guid", check: GUID(), valid: "550e8400-e29b-01d4-0716-446655440000", invalid: "not-a-guid", wantCode: core.InvalidFormat, wantBag: map[string]any{"type": "string"}, wantPattern: true},
		{name: "uuidv4", check: UUIDv4(), valid: "550e8400-e29b-41d4-a716-446655440000", invalid: "550e8400-e29b-71d4-a716-446655440000", wantCode: core.InvalidFormat, wantBag: map[string]any{"type": "string"}, wantPattern: true},
		{name: "uuid6", check: UUID6(), valid: "550e8400-e29b-61d4-a716-446655440000", invalid: "550e8400-e29b-41d4-a716-446655440000", wantCode: core.InvalidFormat, wantBag: map[string]any{"type": "string"}, wantPattern: true},
		{name: "uuid7", check: UUID7(), valid: "550e8400-e29b-71d4-a716-446655440000", invalid: "550e8400-e29b-41d4-a716-446655440000", wantCode: core.InvalidFormat, wantBag: map[string]any{"type": "string"}, wantPattern: true},
		{name: "hex", check: Hex(), valid: "deadBEEF", invalid: "not-hex", wantCode: core.InvalidFormat, wantBag: map[string]any{"type": "string"}, wantPattern: true},
		{name: "md5", check: MD5(), valid: strings.Repeat("a", 32), invalid: strings.Repeat("a", 31), wantCode: core.InvalidFormat, wantBag: map[string]any{"type": "string"}, wantPattern: true},
		{name: "sha1", check: SHA1(), valid: strings.Repeat("a", 40), invalid: strings.Repeat("a", 39), wantCode: core.InvalidFormat, wantBag: map[string]any{"type": "string"}, wantPattern: true},
		{name: "sha256", check: SHA256(), valid: strings.Repeat("a", 64), invalid: strings.Repeat("a", 63), wantCode: core.InvalidFormat, wantBag: map[string]any{"type": "string"}, wantPattern: true},
		{name: "sha384", check: SHA384(), valid: strings.Repeat("a", 96), invalid: strings.Repeat("a", 95), wantCode: core.InvalidFormat, wantBag: map[string]any{"type": "string"}, wantPattern: true},
		{name: "sha512", check: SHA512(), valid: strings.Repeat("a", 128), invalid: strings.Repeat("a", 127), wantCode: core.InvalidFormat, wantBag: map[string]any{"type": "string"}, wantPattern: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			requireCheckAccepts(t, tt.check, tt.valid)
			requireCheckRejects(t, tt.check, tt.invalid, tt.wantCode)

			schema := newCheckAttachSchema(core.ZodTypeString)
			bag := applyCheckAttach(t, tt.check, schema)
			for key, want := range tt.wantBag {
				assert.Equal(t, want, bag[key])
			}
			if tt.wantPattern {
				assert.NotContains(t, bag, "format")
				patterns, ok := bag["patterns"].([]string)
				require.True(t, ok)
				assert.NotEmpty(t, patterns)
			}
		})
	}
}

func TestISODateBounds_ValidateAndAnnotate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		check      core.ZodCheck
		valid      any
		invalid    any
		wantCode   core.IssueCode
		wantKey    string
		wantValue  any
		wantOrigin string
	}{
		{name: "minimum date", check: ISODateMin("2024-01-01"), valid: "2024-01-01", invalid: "2023-12-31", wantCode: core.TooSmall, wantKey: "minimum", wantValue: "2024-01-01", wantOrigin: "date"},
		{name: "maximum date", check: ISODateMax("2024-12-31"), valid: "2024-12-31", invalid: "2025-01-01", wantCode: core.TooBig, wantKey: "maximum", wantValue: "2024-12-31", wantOrigin: "date"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			requireCheckAccepts(t, tt.check, tt.valid)
			issue := requireCheckRejects(t, tt.check, tt.invalid, tt.wantCode)
			assert.Equal(t, tt.wantOrigin, issue.Origin())

			schema := newCheckAttachSchema(core.ZodTypeString)
			bag := applyCheckAttach(t, tt.check, schema)
			assert.Equal(t, tt.wantValue, bag[tt.wantKey])
		})
	}
}

func TestEmailWithPattern_RejectsNonString(t *testing.T) {
	t.Parallel()

	issue := requireCheckRejects(t, EmailWithPattern(regexp.MustCompile(`.+@example\.com`)), 123, core.InvalidType)
	assert.Equal(t, core.ZodTypeString, issue.Expected())
}

func testJWT(encodedHeader string) string {
	return encodedHeader + ".e30.c2lnbmF0dXJl"
}
