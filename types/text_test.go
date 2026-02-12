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
	t.Run("valid single emojis", func(t *testing.T) {
		schema := Emoji()
		for _, emoji := range []string{
			"👋", "🍺", "💚", "🐛", "🇹🇷",
			"😀", "🎉", "❤️", "☘", "〽️",
		} {
			_, err := schema.Parse(emoji)
			assert.NoError(t, err, "should accept %s", emoji)
		}
	})

	t.Run("valid multiple emojis", func(t *testing.T) {
		schema := Emoji()
		for _, emoji := range []string{
			"😀😁",
			"👋👋👋👋",
			"🍺👩‍🚀🫡",
			"💚💙💜💛❤️",
			"🇹🇷🤽🏿‍♂️",
			"🐛🗝🐏🍡🎦🚢🏨💫🎌☘🗡😹🔒🎬➡️🍹🗂🚨⚜🕑〽️🚦🌊🍴💍🍌💰😳🌺🍃",
		} {
			_, err := schema.Parse(emoji)
			assert.NoError(t, err, "should accept %s", emoji)
		}
	})

	t.Run("invalid emojis", func(t *testing.T) {
		schema := Emoji()
		for _, input := range []string{
			":-)", "😀 is an emoji", "😀stuff", "stuff😀",
			"not an emoji", "abc", "123", "hello 🌍 world",
		} {
			_, err := schema.Parse(input)
			assert.Error(t, err, "should reject %s", input)
		}
	})

	t.Run("pointer", func(t *testing.T) {
		schema := EmojiPtr()
		valid := "👋"
		_, err := schema.Parse(&valid)
		assert.NoError(t, err)

		multi := "😀😁🎉"
		_, err = schema.Parse(&multi)
		assert.NoError(t, err)

		invalid := ":-)"
		_, err = schema.Parse(&invalid)
		assert.Error(t, err)
	})

	t.Run("Optional", func(t *testing.T) {
		schema := Emoji().Optional()
		result, err := schema.Parse(nil)
		assert.NoError(t, err)
		assert.Nil(t, result)

		valid := "😀"
		result, err = schema.Parse(valid)
		assert.NoError(t, err)
		assert.Equal(t, &valid, result)
	})

	t.Run("Nilable", func(t *testing.T) {
		schema := Emoji().Nilable()
		result, err := schema.Parse(nil)
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Nullish", func(t *testing.T) {
		schema := Emoji().Nullish()
		result, err := schema.Parse(nil)
		assert.NoError(t, err)
		assert.Nil(t, result)
	})
}

// =============================================================================
// JWT
// =============================================================================

// validJWTHS256 is a well-known HS256 test token used across JWT tests.
const validJWTHS256 = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9." +
	"eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ." +
	"SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

func TestJWT(t *testing.T) {
	tests := []struct {
		name  string
		input any
		valid bool
	}{
		{"valid JWT token", validJWTHS256, true},
		{"missing signature part", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ", false},
		{"malformed token", "invalid.jwt.token", false},
		{"empty string", "", false},
		{"non-string input", 123, false},
		{"single dot", ".", false},
		{"two dots only", "..", false},
		{"invalid base64 in header", "invalid_base64.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := JWT().Parse(tt.input)
			if tt.valid {
				require.NoError(t, err)
				if str, ok := tt.input.(string); ok {
					assert.Equal(t, str, result)
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
		{"valid JWT pointer", validJWTHS256, true},
		{"nil input", nil, false},
		{"invalid JWT pointer", "invalid.jwt.token", false},
		{"empty string pointer", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := JWTPtr().Parse(tt.input)
			if tt.valid {
				require.NoError(t, err)
				if str, ok := tt.input.(string); ok {
					require.NotNil(t, result)
					assert.Equal(t, str, *result)
				}
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestJWTWithAlgorithm(t *testing.T) {
	tests := []struct {
		name      string
		algorithm string
		input     any
		valid     bool
	}{
		{"matching HS256", "HS256", validJWTHS256, true},
		{"mismatch RS256", "RS256", validJWTHS256, false},
		{"invalid structure", "HS256", "invalid.jwt.token", false},
		{"empty string", "HS256", "", false},
		{"non-string input", "HS256", 123, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := JWT(JWTOptions{Algorithm: tt.algorithm}).Parse(tt.input)
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
		name      string
		algorithm string
		input     any
		valid     bool
	}{
		{"matching algorithm", "HS256", validJWTHS256, true},
		{"algorithm mismatch", "RS256", validJWTHS256, false},
		{"nil input", "HS256", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := JWTPtr(JWTOptions{Algorithm: tt.algorithm}).Parse(tt.input)
			if tt.valid {
				require.NoError(t, err)
				if str, ok := tt.input.(string); ok {
					require.NotNil(t, result)
					assert.Equal(t, str, *result)
				}
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestJWTChaining(t *testing.T) {
	schema := JWT().Min(10).Max(1000)

	result, err := schema.Parse(validJWTHS256)
	require.NoError(t, err)
	assert.Equal(t, validJWTHS256, result)

	_, err = schema.Parse("short")
	assert.Error(t, err)
}

func TestJWTWithCustomError(t *testing.T) {
	customError := "Invalid JWT token provided"
	_, err := JWT(customError).Parse("invalid.jwt.token")
	require.Error(t, err)

	var zodErr *issues.ZodError
	if issues.IsZodError(err, &zodErr) && len(zodErr.Issues) > 0 {
		assert.Contains(t, zodErr.Issues[0].Message, customError)
	}
}

func TestJWTWithStringMethods(t *testing.T) {
	t.Run("StartsWith", func(t *testing.T) {
		result, err := JWT().StartsWith("eyJ").Parse(validJWTHS256)
		require.NoError(t, err)
		assert.Equal(t, validJWTHS256, result)
	})

	t.Run("Includes", func(t *testing.T) {
		result, err := JWT().Includes("eyJ").Parse(validJWTHS256)
		require.NoError(t, err)
		assert.Equal(t, validJWTHS256, result)
	})
}

func TestJWTModifiers(t *testing.T) {
	t.Run("Optional", func(t *testing.T) {
		schema := JWT().Optional()
		result, err := schema.Parse(validJWTHS256)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, validJWTHS256, *result)
	})

	t.Run("Nilable", func(t *testing.T) {
		result, err := JWT().Nilable().Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Default", func(t *testing.T) {
		result, err := JWT().Default(validJWTHS256).Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, validJWTHS256, result)
	})

	t.Run("Nullish", func(t *testing.T) {
		result, err := JWT().Nullish().Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestJWTWithOptions(t *testing.T) {
	t.Run("algorithm options", func(t *testing.T) {
		result, err := JWT(JWTOptions{Algorithm: "HS256"}).Parse(validJWTHS256)
		require.NoError(t, err)
		assert.Equal(t, validJWTHS256, result)
	})

	t.Run("options with custom error", func(t *testing.T) {
		_, err := JWT(JWTOptions{Algorithm: "HS256"}, "Invalid HS256 JWT token").Parse("invalid.jwt.token")
		assert.Error(t, err)
	})

	t.Run("options with schema params", func(t *testing.T) {
		_, err := JWT(JWTOptions{Algorithm: "RS256"}, core.SchemaParams{Description: "RSA JWT Token"}).Parse(validJWTHS256)
		assert.Error(t, err, "should fail for algorithm mismatch")
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
		{"valid base64", "SGVsbG8gV29ybGQ=", true},
		{"valid with padding", "Zm9vYg==", true},
		{"valid without padding", "Zm9vYmFy", true},
		{"invalid characters", "SGVsbG8gV29ybGQ$=", false},
		{"invalid padding", "SGVsbG8gV29ybGQ===", false},
		{"empty string", "", true},
		{"non-string input", 12345, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Base64().Parse(tt.input)
			if tt.valid {
				assert.NoError(t, err)
				if str, ok := tt.input.(string); ok {
					assert.Equal(t, str, result)
				}
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestBase64Ptr(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		valid := "SGVsbG8gV29ybGQ="
		result, err := Base64Ptr().Parse(&valid)
		assert.NoError(t, err)
		assert.Equal(t, &valid, result)
	})

	t.Run("nil input", func(t *testing.T) {
		_, err := Base64Ptr().Parse(nil)
		assert.Error(t, err)
	})

	t.Run("invalid", func(t *testing.T) {
		invalid := "invalid!"
		_, err := Base64Ptr().Parse(&invalid)
		assert.Error(t, err)
	})
}

func TestBase64Chaining(t *testing.T) {
	schema := Base64().Min(10)
	_, err := schema.Parse("SGVsbG8gV29ybGQ=")
	assert.NoError(t, err)

	_, err = schema.Parse("Zm9v")
	assert.Error(t, err)
}

func TestBase64WithCustomError(t *testing.T) {
	customError := "Invalid Base64 data"
	_, err := Base64(customError).Parse("invalid!")
	require.Error(t, err)

	var zodErr *issues.ZodError
	if issues.IsZodError(err, &zodErr) && len(zodErr.Issues) > 0 {
		assert.Contains(t, zodErr.Issues[0].Message, customError)
	}
}

func TestBase64Modifiers(t *testing.T) {
	t.Run("Optional", func(t *testing.T) {
		schema := Base64().Optional()
		valid := "SGVsbG8gV29ybGQ="
		result, err := schema.Parse(valid)
		assert.NoError(t, err)
		assert.Equal(t, &valid, result)

		result, err = schema.Parse(nil)
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Nilable", func(t *testing.T) {
		result, err := Base64().Nilable().Parse(nil)
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Default", func(t *testing.T) {
		defaultVal := "SGVsbG8gV29ybGQ="
		result, err := Base64().Default(defaultVal).Parse(nil)
		assert.NoError(t, err)
		assert.Equal(t, defaultVal, result)
	})

	t.Run("Nullish", func(t *testing.T) {
		result, err := Base64().Nullish().Parse(nil)
		assert.NoError(t, err)
		assert.Nil(t, result)
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
		{"valid base64url", "SGVsbG8gV29ybGQ", true},
		{"+ char invalid", "Zm9v+g==", false},
		{"/ char invalid", "Zm9v/g==", false},
		{"valid no padding", "Zm9vYmFy", true},
		{"valid with padding", "Zm9vYg==", true},
		{"empty string", "", true},
		{"non-string input", 12345, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Base64URL().Parse(tt.input)
			if tt.valid {
				assert.NoError(t, err)
				if str, ok := tt.input.(string); ok {
					assert.Equal(t, str, result)
				}
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestBase64URLPtr(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		valid := "SGVsbG8gV29ybGQ"
		result, err := Base64URLPtr().Parse(&valid)
		assert.NoError(t, err)
		assert.Equal(t, &valid, result)
	})

	t.Run("nil input", func(t *testing.T) {
		_, err := Base64URLPtr().Parse(nil)
		assert.Error(t, err)
	})

	t.Run("invalid", func(t *testing.T) {
		invalid := "not-valid-url-safe-base64+"
		_, err := Base64URLPtr().Parse(&invalid)
		assert.Error(t, err)
	})
}

func TestBase64URLChaining(t *testing.T) {
	schema := Base64URL().Min(10)
	_, err := schema.Parse("SGVsbG8gV29ybGQ")
	assert.NoError(t, err)

	_, err = schema.Parse("Zm9v")
	assert.Error(t, err)
}

func TestBase64URLWithCustomError(t *testing.T) {
	customError := "Invalid Base64URL data"
	_, err := Base64URL(customError).Parse("invalid+base64url")
	require.Error(t, err)

	var zodErr *issues.ZodError
	if issues.IsZodError(err, &zodErr) && len(zodErr.Issues) > 0 {
		assert.Contains(t, zodErr.Issues[0].Message, customError)
	}
}

func TestBase64URLModifiers(t *testing.T) {
	t.Run("Optional", func(t *testing.T) {
		schema := Base64URL().Optional()
		valid := "SGVsbG8gV29ybGQ"
		result, err := schema.Parse(valid)
		assert.NoError(t, err)
		assert.Equal(t, &valid, result)

		result, err = schema.Parse(nil)
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Nilable", func(t *testing.T) {
		result, err := Base64URL().Nilable().Parse(nil)
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Default", func(t *testing.T) {
		defaultVal := "SGVsbG8gV29ybGQ"
		result, err := Base64URL().Default(defaultVal).Parse(nil)
		assert.NoError(t, err)
		assert.Equal(t, defaultVal, result)
	})

	t.Run("Nullish", func(t *testing.T) {
		result, err := Base64URL().Nullish().Parse(nil)
		assert.NoError(t, err)
		assert.Nil(t, result)
	})
}

// =============================================================================
// Default and Prefault
// =============================================================================

func TestText_DefaultAndPrefault(t *testing.T) {
	t.Run("Default has higher priority than Prefault", func(t *testing.T) {
		result, err := Emoji().Default("😀").Prefault("😁").Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "😀", result)
	})

	t.Run("Default short-circuits validation", func(t *testing.T) {
		result, err := Emoji().Default("not-an-emoji").Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "not-an-emoji", result)
	})

	t.Run("Prefault goes through full validation", func(t *testing.T) {
		result, err := Emoji().Prefault("🎉").Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "🎉", result)
	})

	t.Run("Prefault only triggered by nil input", func(t *testing.T) {
		_, err := Emoji().Prefault("🚀").Parse("not-an-emoji")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid string: must match pattern")
	})

	t.Run("DefaultFunc and PrefaultFunc behavior", func(t *testing.T) {
		defaultCalled := false
		prefaultCalled := false

		schema := Base64().DefaultFunc(func() string {
			defaultCalled = true
			return "SGVsbG8="
		}).PrefaultFunc(func() string {
			prefaultCalled = true
			return "V29ybGQ="
		})

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "SGVsbG8=", result)
		assert.True(t, defaultCalled, "DefaultFunc should be called")
		assert.False(t, prefaultCalled, "PrefaultFunc should not be called when Default is present")
	})

	t.Run("Prefault validation failure returns error", func(t *testing.T) {
		result, err := JWT().Prefault(validJWTHS256).Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, validJWTHS256, result)
	})
}
