package validate_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kaptinlin/gozod/pkg/validate"
)

func TestJWTWithOptions(t *testing.T) {
	t.Parallel()

	hs256 := "HS256"
	rs256 := "RS256"

	tests := []struct {
		name    string
		value   any
		options validate.JWTOptions
		want    bool
	}{
		{name: "valid jwt", value: jwtToken("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"), want: true},
		{name: "matches expected algorithm", value: jwtToken("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"), options: validate.JWTOptions{Algorithm: &hs256}, want: true},
		{name: "rejects algorithm mismatch", value: jwtToken("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"), options: validate.JWTOptions{Algorithm: &rs256}, want: false},
		{name: "rejects none algorithm", value: jwtToken("eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0"), want: false},
		{name: "rejects non jwt typ", value: jwtToken("eyJhbGciOiJIUzI1NiIsInR5cCI6ImF0K2p3dCJ9"), want: false},
		{name: "rejects malformed token", value: "not-a-token", want: false},
		{name: "rejects non string", value: 123, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, validate.JWTWithOptions(tt.value, tt.options))
		})
	}
}

func jwtToken(encodedHeader string) string {
	return encodedHeader + ".e30.c2lnbmF0dXJl"
}
