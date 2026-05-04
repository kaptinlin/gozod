package validate_test

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kaptinlin/gozod/pkg/validate"
)

func TestURLWithOptions(t *testing.T) {
	t.Parallel()

	hostname := regexp.MustCompile(`^api\.example\.com$`)
	protocol := regexp.MustCompile(`^https$`)

	tests := []struct {
		name    string
		value   any
		options validate.URLOptions
		want    bool
	}{
		{name: "valid URL without constraints", value: "https://example.com/path", want: true},
		{name: "rejects malformed URL", value: "not-a-url", want: false},
		{name: "rejects non string", value: 42, want: false},
		{name: "matches hostname and protocol", value: "https://api.example.com/v1", options: validate.URLOptions{Hostname: hostname, Protocol: protocol}, want: true},
		{name: "rejects hostname mismatch", value: "https://www.example.com/v1", options: validate.URLOptions{Hostname: hostname}, want: false},
		{name: "rejects protocol mismatch", value: "http://api.example.com/v1", options: validate.URLOptions{Protocol: protocol}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, validate.URLWithOptions(tt.value, tt.options))
		})
	}
}

func TestHostname(t *testing.T) {
	t.Parallel()

	longHostname := makeString('a', 254)

	tests := []struct {
		name  string
		value any
		want  bool
	}{
		{name: "simple hostname", value: "example.com", want: true},
		{name: "trailing dot", value: "example.com.", want: true},
		{name: "empty string", value: "", want: false},
		{name: "too long", value: longHostname, want: false},
		{name: "invalid label", value: "-example.com", want: false},
		{name: "non string", value: 123, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, validate.Hostname(tt.value))
		})
	}
}

func TestNetworkIdentifierWrappers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  bool
		want bool
	}{
		{name: "url", got: validate.URL("https://example.com"), want: true},
		{name: "uuid", got: validate.UUID("550e8400-e29b-41d4-a716-446655440000"), want: true},
		{name: "guid accepts versionless uuid shape", got: validate.GUID("550e8400-e29b-01d4-0716-446655440000"), want: true},
		{name: "cuid", got: validate.CUID("cjld2cyuq0000t3rmniod1foy"), want: true},
		{name: "cuid2", got: validate.CUID2("abc123"), want: true},
		{name: "nano id", got: validate.NanoID("V1StGXR8_Z5jdHi6B-myT"), want: true},
		{name: "ulid", got: validate.ULID("01ARZ3NDEKTSV4RRFFQ69G5FAV"), want: true},
		{name: "xid", got: validate.XID("9m4e2mr0ui3e8a215n4g"), want: true},
		{name: "ksuid", got: validate.KSUID("0ujtsYcgvSTl8PAuAdqWYSMnLOv"), want: true},
		{name: "ipv4", got: validate.IPv4("192.168.1.1"), want: true},
		{name: "ipv6", got: validate.IPv6("::1"), want: true},
		{name: "uuid rejects non string", got: validate.UUID(123), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.got)
		})
	}
}

func TestMime(t *testing.T) {
	t.Parallel()

	allowed := []string{"application/json", "text/plain"}

	tests := []struct {
		name  string
		value any
		want  bool
	}{
		{name: "allowed type", value: "application/json", want: true},
		{name: "well formed but not allowed", value: "image/png", want: false},
		{name: "missing separator", value: "application", want: false},
		{name: "missing type", value: "/json", want: false},
		{name: "missing subtype", value: "application/", want: false},
		{name: "non string", value: 123, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, validate.Mime(tt.value, allowed))
		})
	}
}

func makeString(r rune, count int) string {
	out := make([]rune, count)
	for i := range out {
		out[i] = r
	}
	return string(out)
}
