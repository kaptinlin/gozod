package validate_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kaptinlin/gozod/pkg/validate"
)

func TestEncodingAndHashWrappers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  bool
		want bool
	}{
		{name: "email", got: validate.Email("user@example.com"), want: true},
		{name: "base64", got: validate.Base64("SGVsbG8="), want: true},
		{name: "base64url", got: validate.Base64URL("SGVsbG8"), want: true},
		{name: "hex", got: validate.Hex("deadBEEF"), want: true},
		{name: "md5", got: validate.MD5Hex(strings.Repeat("a", 32)), want: true},
		{name: "sha1", got: validate.SHA1Hex(strings.Repeat("a", 40)), want: true},
		{name: "sha256", got: validate.SHA256Hex(strings.Repeat("a", 64)), want: true},
		{name: "sha384", got: validate.SHA384Hex(strings.Repeat("a", 96)), want: true},
		{name: "sha512", got: validate.SHA512Hex(strings.Repeat("a", 128)), want: true},
		{name: "phone", got: validate.E164("+14155552671"), want: true},
		{name: "emoji", got: validate.Emoji("😀"), want: true},
		{name: "email rejects invalid", got: validate.Email("not-email"), want: false},
		{name: "base64 rejects invalid", got: validate.Base64("!!!"), want: false},
		{name: "hash rejects non string", got: validate.SHA256Hex(123), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.got)
		})
	}
}
