package types

import (
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Emoji
// =============================================================================

func TestEmoji(t *testing.T) {
	t.Run("valid_single", func(t *testing.T) {
		s := Emoji()
		for _, v := range []string{
			"👋", "🍺", "💚", "🐛", "🇹🇷",
			"😀", "🎉", "❤️", "☘", "〽️",
		} {
			_, err := s.Parse(v)
			assert.NoError(t, err, "Emoji().Parse(%q) returned unexpected error", v)
		}
	})

	t.Run("valid_multiple", func(t *testing.T) {
		s := Emoji()
		for _, v := range []string{
			"😀😁",
			"👋👋👋👋",
			"🍺👩‍🚀🫡",
			"💚💙💜💛❤️",
			"🇹🇷🤽🏿‍♂️",
			"🐛🗝🐏🍡🎦🚢🏨💫🎌☘🗡😹🔒🎬➡️🍹🗂🚨⚜🕑〽️🚦🌊🍴💍🍌💰😳🌺🍃",
		} {
			_, err := s.Parse(v)
			assert.NoError(t, err, "Emoji().Parse(%q) returned unexpected error", v)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		s := Emoji()
		for _, v := range []string{
			":-)", "😀 is an emoji", "😀stuff", "stuff😀",
			"not an emoji", "abc", "123", "hello 🌍 world",
		} {
			_, err := s.Parse(v)
			assert.Error(t, err, "Emoji().Parse(%q) = _, nil; want error", v)
		}
	})

	t.Run("pointer", func(t *testing.T) {
		s := EmojiPtr()
		v := "👋"
		_, err := s.Parse(&v)
		assert.NoError(t, err)

		multi := "😀😁🎉"
		_, err = s.Parse(&multi)
		assert.NoError(t, err)

		bad := ":-)"
		_, err = s.Parse(&bad)
		assert.Error(t, err)
	})

	t.Run("Optional", func(t *testing.T) {
		s := Emoji().Optional()
		got, err := s.Parse(nil)
		assert.NoError(t, err)
		assert.Nil(t, got)

		v := "😀"
		got, err = s.Parse(v)
		assert.NoError(t, err)
		assert.Equal(t, &v, got)
	})

	t.Run("Nilable", func(t *testing.T) {
		got, err := Emoji().Nilable().Parse(nil)
		assert.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("Nullish", func(t *testing.T) {
		got, err := Emoji().Nullish().Parse(nil)
		assert.NoError(t, err)
		assert.Nil(t, got)
	})
}

// =============================================================================
// JWT
// =============================================================================

// testJWTHS256 is a well-known HS256 test token used across JWT tests.
const testJWTHS256 = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9." +
	"eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ." +
	"SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

func TestJWT(t *testing.T) {
	tests := []struct {
		name  string
		input any
		valid bool
	}{
		{"valid_token", testJWTHS256, true},
		{"missing_signature", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ", false},
		{"malformed", "invalid.jwt.token", false},
		{"empty", "", false},
		{"non_string", 123, false},
		{"single_dot", ".", false},
		{"two_dots", "..", false},
		{"invalid_base64_header", "invalid_base64.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := JWT().Parse(tt.input)
			if tt.valid {
				require.NoError(t, err)
				if s, ok := tt.input.(string); ok {
					assert.Equal(t, s, got)
				}
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestJWTPtr(t *testing.T) {
	tests := []struct {
		name  string
		input any
		valid bool
	}{
		{"valid", testJWTHS256, true},
		{"nil", nil, false},
		{"invalid", "invalid.jwt.token", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := JWTPtr().Parse(tt.input)
			if tt.valid {
				require.NoError(t, err)
				if s, ok := tt.input.(string); ok {
					require.NotNil(t, got)
					assert.Equal(t, s, *got)
				}
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestJWTWithAlgorithm(t *testing.T) {
	tests := []struct {
		name  string
		alg   string
		input any
		valid bool
	}{
		{"match_HS256", "HS256", testJWTHS256, true},
		{"mismatch_RS256", "RS256", testJWTHS256, false},
		{"invalid_structure", "HS256", "invalid.jwt.token", false},
		{"empty", "HS256", "", false},
		{"non_string", "HS256", 123, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := JWT(JWTOptions{Algorithm: tt.alg}).Parse(tt.input)
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestJWTPtrWithAlgorithm(t *testing.T) {
	tests := []struct {
		name  string
		alg   string
		input any
		valid bool
	}{
		{"match", "HS256", testJWTHS256, true},
		{"mismatch", "RS256", testJWTHS256, false},
		{"nil", "HS256", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := JWTPtr(JWTOptions{Algorithm: tt.alg}).Parse(tt.input)
			if tt.valid {
				require.NoError(t, err)
				if s, ok := tt.input.(string); ok {
					require.NotNil(t, got)
					assert.Equal(t, s, *got)
				}
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestJWTChaining(t *testing.T) {
	s := JWT().Min(10).Max(1000)

	got, err := s.Parse(testJWTHS256)
	require.NoError(t, err)
	assert.Equal(t, testJWTHS256, got)

	_, err = s.Parse("short")
	assert.Error(t, err)
}

func TestJWTCustomError(t *testing.T) {
	msg := "Invalid JWT token provided"
	_, err := JWT(msg).Parse("invalid.jwt.token")
	require.Error(t, err)

	var zodErr *issues.ZodError
	if issues.IsZodError(err, &zodErr) && len(zodErr.Issues) > 0 {
		assert.Contains(t, zodErr.Issues[0].Message, msg)
	}
}

func TestJWTStringMethods(t *testing.T) {
	t.Run("StartsWith", func(t *testing.T) {
		got, err := JWT().StartsWith("eyJ").Parse(testJWTHS256)
		require.NoError(t, err)
		assert.Equal(t, testJWTHS256, got)
	})

	t.Run("Includes", func(t *testing.T) {
		got, err := JWT().Includes("eyJ").Parse(testJWTHS256)
		require.NoError(t, err)
		assert.Equal(t, testJWTHS256, got)
	})
}

func TestJWTModifiers(t *testing.T) {
	t.Run("Optional", func(t *testing.T) {
		got, err := JWT().Optional().Parse(testJWTHS256)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, testJWTHS256, *got)
	})

	t.Run("Nilable", func(t *testing.T) {
		got, err := JWT().Nilable().Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("Default", func(t *testing.T) {
		got, err := JWT().Default(testJWTHS256).Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, testJWTHS256, got)
	})

	t.Run("Nullish", func(t *testing.T) {
		got, err := JWT().Nullish().Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}

func TestJWTOptions(t *testing.T) {
	t.Run("algorithm", func(t *testing.T) {
		got, err := JWT(JWTOptions{Algorithm: "HS256"}).Parse(testJWTHS256)
		require.NoError(t, err)
		assert.Equal(t, testJWTHS256, got)
	})

	t.Run("with_custom_error", func(t *testing.T) {
		_, err := JWT(JWTOptions{Algorithm: "HS256"}, "Invalid HS256 JWT token").Parse("invalid.jwt.token")
		assert.Error(t, err)
	})

	t.Run("with_schema_params", func(t *testing.T) {
		_, err := JWT(JWTOptions{Algorithm: "RS256"}, core.SchemaParams{Description: "RSA JWT Token"}).Parse(testJWTHS256)
		assert.Error(t, err, "JWT(RS256).Parse(HS256 token) = _, nil; want algorithm mismatch error")
	})
}

// =============================================================================
// Base64
// =============================================================================

func TestBase64(t *testing.T) {
	tests := []struct {
		name  string
		input any
		valid bool
	}{
		{"valid", "SGVsbG8gV29ybGQ=", true},
		{"valid_padded", "Zm9vYg==", true},
		{"valid_no_padding", "Zm9vYmFy", true},
		{"invalid_chars", "SGVsbG8gV29ybGQ$=", false},
		{"invalid_padding", "SGVsbG8gV29ybGQ===", false},
		{"empty", "", true},
		{"non_string", 12345, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Base64().Parse(tt.input)
			if tt.valid {
				assert.NoError(t, err)
				if s, ok := tt.input.(string); ok {
					assert.Equal(t, s, got)
				}
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestBase64Ptr(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		v := "SGVsbG8gV29ybGQ="
		got, err := Base64Ptr().Parse(&v)
		assert.NoError(t, err)
		assert.Equal(t, &v, got)
	})

	t.Run("nil", func(t *testing.T) {
		_, err := Base64Ptr().Parse(nil)
		assert.Error(t, err)
	})

	t.Run("invalid", func(t *testing.T) {
		v := "invalid!"
		_, err := Base64Ptr().Parse(&v)
		assert.Error(t, err)
	})
}

func TestBase64Chaining(t *testing.T) {
	s := Base64().Min(10)
	_, err := s.Parse("SGVsbG8gV29ybGQ=")
	assert.NoError(t, err)

	_, err = s.Parse("Zm9v")
	assert.Error(t, err)
}

func TestBase64CustomError(t *testing.T) {
	msg := "Invalid Base64 data"
	_, err := Base64(msg).Parse("invalid!")
	require.Error(t, err)

	var zodErr *issues.ZodError
	if issues.IsZodError(err, &zodErr) && len(zodErr.Issues) > 0 {
		assert.Contains(t, zodErr.Issues[0].Message, msg)
	}
}

func TestBase64Modifiers(t *testing.T) {
	t.Run("Optional", func(t *testing.T) {
		s := Base64().Optional()
		v := "SGVsbG8gV29ybGQ="
		got, err := s.Parse(v)
		assert.NoError(t, err)
		assert.Equal(t, &v, got)

		got, err = s.Parse(nil)
		assert.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("Nilable", func(t *testing.T) {
		got, err := Base64().Nilable().Parse(nil)
		assert.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("Default", func(t *testing.T) {
		def := "SGVsbG8gV29ybGQ="
		got, err := Base64().Default(def).Parse(nil)
		assert.NoError(t, err)
		assert.Equal(t, def, got)
	})

	t.Run("Nullish", func(t *testing.T) {
		got, err := Base64().Nullish().Parse(nil)
		assert.NoError(t, err)
		assert.Nil(t, got)
	})
}

// =============================================================================
// Base64URL
// =============================================================================

func TestBase64URL(t *testing.T) {
	tests := []struct {
		name  string
		input any
		valid bool
	}{
		{"valid", "SGVsbG8gV29ybGQ", true},
		{"plus_invalid", "Zm9v+g==", false},
		{"slash_invalid", "Zm9v/g==", false},
		{"valid_no_padding", "Zm9vYmFy", true},
		{"valid_padded", "Zm9vYg==", true},
		{"empty", "", true},
		{"non_string", 12345, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Base64URL().Parse(tt.input)
			if tt.valid {
				assert.NoError(t, err)
				if s, ok := tt.input.(string); ok {
					assert.Equal(t, s, got)
				}
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestBase64URLPtr(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		v := "SGVsbG8gV29ybGQ"
		got, err := Base64URLPtr().Parse(&v)
		assert.NoError(t, err)
		assert.Equal(t, &v, got)
	})

	t.Run("nil", func(t *testing.T) {
		_, err := Base64URLPtr().Parse(nil)
		assert.Error(t, err)
	})

	t.Run("invalid", func(t *testing.T) {
		v := "not-valid-url-safe-base64+"
		_, err := Base64URLPtr().Parse(&v)
		assert.Error(t, err)
	})
}

func TestBase64URLChaining(t *testing.T) {
	s := Base64URL().Min(10)
	_, err := s.Parse("SGVsbG8gV29ybGQ")
	assert.NoError(t, err)

	_, err = s.Parse("Zm9v")
	assert.Error(t, err)
}

func TestBase64URLCustomError(t *testing.T) {
	msg := "Invalid Base64URL data"
	_, err := Base64URL(msg).Parse("invalid+base64url")
	require.Error(t, err)

	var zodErr *issues.ZodError
	if issues.IsZodError(err, &zodErr) && len(zodErr.Issues) > 0 {
		assert.Contains(t, zodErr.Issues[0].Message, msg)
	}
}

func TestBase64URLModifiers(t *testing.T) {
	t.Run("Optional", func(t *testing.T) {
		s := Base64URL().Optional()
		v := "SGVsbG8gV29ybGQ"
		got, err := s.Parse(v)
		assert.NoError(t, err)
		assert.Equal(t, &v, got)

		got, err = s.Parse(nil)
		assert.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("Nilable", func(t *testing.T) {
		got, err := Base64URL().Nilable().Parse(nil)
		assert.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("Default", func(t *testing.T) {
		def := "SGVsbG8gV29ybGQ"
		got, err := Base64URL().Default(def).Parse(nil)
		assert.NoError(t, err)
		assert.Equal(t, def, got)
	})

	t.Run("Nullish", func(t *testing.T) {
		got, err := Base64URL().Nullish().Parse(nil)
		assert.NoError(t, err)
		assert.Nil(t, got)
	})
}

// =============================================================================
// Hex
// =============================================================================

func TestHex(t *testing.T) {
	tests := []struct {
		name  string
		input any
		valid bool
	}{
		{"valid_lower", "deadbeef", true},
		{"valid_upper", "DEADBEEF", true},
		{"valid_mixed", "DeAdBeEf", true},
		{"valid_digits", "0123456789", true},
		{"empty", "", true},
		{"invalid_chars", "xyz123", false},
		{"non_string", 12345, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Hex().Parse(tt.input)
			if tt.valid {
				assert.NoError(t, err)
				if s, ok := tt.input.(string); ok {
					assert.Equal(t, s, got)
				}
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestHexPtr(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		v := "deadbeef"
		got, err := HexPtr().Parse(&v)
		assert.NoError(t, err)
		assert.Equal(t, &v, got)
	})

	t.Run("nil", func(t *testing.T) {
		_, err := HexPtr().Parse(nil)
		assert.Error(t, err)
	})

	t.Run("invalid", func(t *testing.T) {
		v := "xyz"
		_, err := HexPtr().Parse(&v)
		assert.Error(t, err)
	})
}

func TestHexModifiers(t *testing.T) {
	t.Run("Optional", func(t *testing.T) {
		got, err := Hex().Optional().Parse(nil)
		assert.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("Nilable", func(t *testing.T) {
		got, err := Hex().Nilable().Parse(nil)
		assert.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("Nullish", func(t *testing.T) {
		got, err := Hex().Nullish().Parse(nil)
		assert.NoError(t, err)
		assert.Nil(t, got)
	})
}

// =============================================================================
// Default and Prefault
// =============================================================================

func TestTextDefaultAndPrefault(t *testing.T) {
	t.Run("default_over_prefault", func(t *testing.T) {
		got, err := Emoji().Default("😀").Prefault("😁").Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "😀", got)
	})

	t.Run("default_short_circuits", func(t *testing.T) {
		got, err := Emoji().Default("not-an-emoji").Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "not-an-emoji", got)
	})

	t.Run("prefault_validates", func(t *testing.T) {
		got, err := Emoji().Prefault("🎉").Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "🎉", got)
	})

	t.Run("prefault_nil_only", func(t *testing.T) {
		_, err := Emoji().Prefault("🚀").Parse("not-an-emoji")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid string: must match pattern")
	})

	t.Run("DefaultFunc_PrefaultFunc", func(t *testing.T) {
		defCalled := false
		preCalled := false

		s := Base64().DefaultFunc(func() string {
			defCalled = true
			return "SGVsbG8="
		}).PrefaultFunc(func() string {
			preCalled = true
			return "V29ybGQ="
		})

		got, err := s.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "SGVsbG8=", got)
		assert.True(t, defCalled, "DefaultFunc should be called")
		assert.False(t, preCalled, "PrefaultFunc should not be called when Default is present")
	})

	t.Run("prefault_jwt", func(t *testing.T) {
		got, err := JWT().Prefault(testJWTHS256).Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, testJWTHS256, got)
	})
}
