package types

import (
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/stretchr/testify/assert"
)

func TestEmoji(t *testing.T) {
	t.Run("valid single emojis", func(t *testing.T) {
		schema := Emoji()
		validEmojis := []string{
			"👋",
			"🍺",
			"💚",
			"🐛",
			"🇹🇷",
			"😀",
			"🎉",
			"❤️",
			"☘",
			"〽️",
		}

		for _, emoji := range validEmojis {
			_, err := schema.Parse(emoji)
			assert.NoError(t, err, "should be valid for emoji %s", emoji)
		}
	})

	t.Run("valid multiple emojis", func(t *testing.T) {
		schema := Emoji()
		validMultiEmojis := []string{
			"😀😁",      // two basic emojis
			"👋👋👋👋",    // four same emojis
			"🍺👩‍🚀🫡",   // complex emojis with ZWJ
			"💚💙💜💛❤️",  // multiple heart emojis
			"🇹🇷🤽🏿‍♂️", // flag + complex emoji
			"🐛🗝🐏🍡🎦🚢🏨💫🎌☘🗡😹🔒🎬➡️🍹🗂🚨⚜🕑〽️🚦🌊🍴💍🍌💰😳🌺🍃", // many emojis
		}

		for _, emoji := range validMultiEmojis {
			_, err := schema.Parse(emoji)
			assert.NoError(t, err, "should be valid for multi-emoji %s", emoji)
		}
	})

	t.Run("invalid emojis", func(t *testing.T) {
		schema := Emoji()
		invalidEmojis := []string{
			":-)",
			"😀 is an emoji",
			"😀stuff",
			"stuff😀",
			"not an emoji",
			"abc",
			"123",
			"hello 🌍 world", // text with emoji
		}

		for _, emoji := range invalidEmojis {
			_, err := schema.Parse(emoji)
			assert.Error(t, err, "should be invalid for emoji %s", emoji)
		}
	})

	t.Run("valid emojis pointer", func(t *testing.T) {
		schema := EmojiPtr()
		validEmoji := "👋"

		_, err := schema.Parse(&validEmoji)
		assert.NoError(t, err)

		// Test multiple emojis with pointer
		multiEmoji := "😀😁🎉"
		_, err = schema.Parse(&multiEmoji)
		assert.NoError(t, err)
	})

	t.Run("invalid emojis pointer", func(t *testing.T) {
		schema := EmojiPtr()
		invalidEmoji := ":-)"

		_, err := schema.Parse(&invalidEmoji)
		assert.Error(t, err)
	})
}

// =============================================================================
// JWT TESTS
// =============================================================================

func TestJWT(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected bool
	}{
		{
			name:     "valid JWT token",
			input:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			expected: true,
		},
		{
			name:     "invalid JWT - missing signature part",
			input:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ",
			expected: false,
		},
		{
			name:     "invalid JWT - malformed token",
			input:    "invalid.jwt.token",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "non-string input",
			input:    123,
			expected: false,
		},
		{
			name:     "single dot",
			input:    ".",
			expected: false,
		},
		{
			name:     "two dots only",
			input:    "..",
			expected: false,
		},
		{
			name:     "invalid base64 in header",
			input:    "invalid_base64.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := JWT()
			result, err := schema.Parse(tt.input)

			if tt.expected {
				if err != nil {
					t.Errorf("JWT().Parse(%v) returned error: %v", tt.input, err)
				} else if str, ok := tt.input.(string); ok && result != str {
					t.Errorf("JWT().Parse(%v) = %v, want %v", tt.input, result, str)
				}
			} else {
				if err == nil {
					t.Errorf("JWT().Parse(%v) expected error but got none", tt.input)
				}
			}
		})
	}
}

func TestJWTPtr(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected bool
	}{
		{
			name:     "valid JWT token pointer",
			input:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			expected: true,
		},
		{
			name:     "nil input",
			input:    nil,
			expected: false,
		},
		{
			name:     "invalid JWT token pointer",
			input:    "invalid.jwt.token",
			expected: false,
		},
		{
			name:     "empty string pointer",
			input:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := JWTPtr()
			result, err := schema.Parse(tt.input)

			if tt.expected {
				if err != nil {
					t.Errorf("JWTPtr().Parse(%v) returned error: %v", tt.input, err)
				} else if str, ok := tt.input.(string); ok && result != nil && *result != str {
					t.Errorf("JWTPtr().Parse(%v) = %v, want %v", tt.input, *result, str)
				}
			} else {
				if err == nil {
					t.Errorf("JWTPtr().Parse(%v) expected error but got none", tt.input)
				}
			}
		})
	}
}

func TestJWTWithAlgorithm(t *testing.T) {
	tests := []struct {
		name      string
		algorithm string
		input     any
		expected  bool
	}{
		{
			name:      "valid JWT with matching HS256 algorithm",
			algorithm: "HS256",
			input:     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			expected:  true,
		},
		{
			name:      "valid JWT structure but algorithm mismatch (RS256 expected but HS256 in token)",
			algorithm: "RS256",
			input:     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			expected:  false,
		},
		{
			name:      "invalid JWT token structure",
			algorithm: "HS256",
			input:     "invalid.jwt.token",
			expected:  false,
		},
		{
			name:      "empty string with algorithm constraint",
			algorithm: "HS256",
			input:     "",
			expected:  false,
		},
		{
			name:      "non-string input with algorithm constraint",
			algorithm: "HS256",
			input:     123,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := JWT(JWTOptions{Algorithm: tt.algorithm})
			result, err := schema.Parse(tt.input)

			if tt.expected {
				if err != nil {
					t.Errorf("JWT(...).Parse(%v) returned error: %v", tt.input, err)
				} else if str, ok := tt.input.(string); ok && result != str {
					t.Errorf("JWT(...).Parse(%v) = %v, want %v", tt.input, result, str)
				}
			} else {
				if err == nil {
					t.Errorf("JWT(...).Parse(%v) expected error but got none", tt.input)
				}
			}
		})
	}
}

func TestJWTPtrWithAlgorithm(t *testing.T) {
	tests := []struct {
		name      string
		algorithm string
		input     any
		expected  bool
	}{
		{
			name:      "valid JWT token pointer with matching algorithm",
			algorithm: "HS256",
			input:     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			expected:  true,
		},
		{
			name:      "algorithm mismatch with pointer type",
			algorithm: "RS256",
			input:     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			expected:  false,
		},
		{
			name:      "nil input with algorithm constraint",
			algorithm: "HS256",
			input:     nil,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := JWTPtr(JWTOptions{Algorithm: tt.algorithm})
			result, err := schema.Parse(tt.input)

			if tt.expected {
				if err != nil {
					t.Errorf("JWTPtr(...).Parse(%v) returned error: %v", tt.input, err)
				} else if str, ok := tt.input.(string); ok && result != nil && *result != str {
					t.Errorf("JWTPtr(...).Parse(%v) = %v, want %v", tt.input, *result, str)
				}
			} else {
				if err == nil {
					t.Errorf("JWTPtr(...).Parse(%v) expected error but got none", tt.input)
				}
			}
		})
	}
}

func TestJWTChaining(t *testing.T) {
	// Test that JWT type supports method chaining with string validation methods
	schema := JWT().Min(10).Max(1000)

	validJWT := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

	result, err := schema.Parse(validJWT)
	if err != nil {
		t.Errorf("JWT().Min(10).Max(1000).Parse(%v) returned error: %v", validJWT, err)
	}

	if result != validJWT {
		t.Errorf("JWT().Min(10).Max(1000).Parse(%v) = %v, want %v", validJWT, result, validJWT)
	}

	// Test that length validation works with JWT
	shortJWT := "short"
	_, err = schema.Parse(shortJWT)
	if err == nil {
		t.Error("Expected error for JWT that's too short")
	}
}

func TestJWTWithCustomError(t *testing.T) {
	customError := "Invalid JWT token provided"
	schema := JWT(customError)

	_, err := schema.Parse("invalid.jwt.token")
	if err == nil {
		t.Error("Expected error for invalid JWT")
		return
	}

	var zodErr2 *issues.ZodError
	if issues.IsZodError(err, &zodErr2) {
		if len(zodErr2.Issues) > 0 {
			assert.Contains(t, zodErr2.Issues[0].Message, customError)
		}
	}
}

func TestJWTWithStringMethods(t *testing.T) {
	t.Run("JWT with StartsWith validation", func(t *testing.T) {
		// Create a JWT schema that requires token to start with "eyJ"
		schema := JWT().StartsWith("eyJ")

		validJWT := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

		result, err := schema.Parse(validJWT)
		if err != nil {
			t.Errorf("JWT().StartsWith('eyJ').Parse(%v) returned error: %v", validJWT, err)
		}

		if result != validJWT {
			t.Errorf("JWT().StartsWith('eyJ').Parse(%v) = %v, want %v", validJWT, result, validJWT)
		}
	})

	t.Run("JWT with Includes validation", func(t *testing.T) {
		// Create a JWT schema that requires token to include "eyJ"
		schema := JWT().Includes("eyJ")

		validJWT := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

		result, err := schema.Parse(validJWT)
		if err != nil {
			t.Errorf("JWT().Includes('eyJ').Parse(%v) returned error: %v", validJWT, err)
		}

		if result != validJWT {
			t.Errorf("JWT().Includes('eyJ').Parse(%v) = %v, want %v", validJWT, result, validJWT)
		}
	})
}

func TestJWTModifiers(t *testing.T) {
	t.Run("JWT Optional", func(t *testing.T) {
		schema := JWT().Optional()

		// Test with valid JWT
		validJWT := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
		result, err := schema.Parse(validJWT)
		if err != nil {
			t.Errorf("JWT().Optional().Parse(%v) returned error: %v", validJWT, err)
		}
		if result == nil || *result != validJWT {
			t.Errorf("JWT().Optional().Parse(%v) = %v, want %v", validJWT, result, validJWT)
		}
	})

	t.Run("JWT Nilable", func(t *testing.T) {
		schema := JWT().Nilable()

		// Test with nil input
		result, err := schema.Parse(nil)
		if err != nil {
			t.Errorf("JWT().Nilable().Parse(nil) returned error: %v", err)
		}
		if result != nil {
			t.Errorf("JWT().Nilable().Parse(nil) = %v, want nil", result)
		}
	})

	t.Run("JWT with Default", func(t *testing.T) {
		defaultJWT := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
		schema := JWT().Default(defaultJWT)

		// Test with undefined input (should use default)
		result, err := schema.Parse(nil)
		if err != nil {
			t.Errorf("JWT().Default().Parse(nil) returned error: %v", err)
		}
		if result != defaultJWT {
			t.Errorf("JWT().Default().Parse(nil) = %v, want %v", result, defaultJWT)
		}
	})
}

func TestJWTWithOptions(t *testing.T) {
	t.Run("JWT with algorithm options", func(t *testing.T) {
		schema := JWT(JWTOptions{Algorithm: "HS256"})

		validJWT := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

		result, err := schema.Parse(validJWT)
		if err != nil {
			t.Errorf("JWT(JWTOptions{Algorithm: HS256}).Parse(%v) returned error: %v", validJWT, err)
		}
		if result != validJWT {
			t.Errorf("JWT(JWTOptions{Algorithm: HS256}).Parse(%v) = %v, want %v", validJWT, result, validJWT)
		}
	})

	t.Run("JWT with options and custom error", func(t *testing.T) {
		customError := "Invalid HS256 JWT token"
		schema := JWT(JWTOptions{Algorithm: "HS256"}, customError)

		_, err := schema.Parse("invalid.jwt.token")
		if err == nil {
			t.Error("Expected error for invalid JWT")
		}
	})

	t.Run("JWT with options and schema params", func(t *testing.T) {
		schemaParams := core.SchemaParams{Description: "RSA JWT Token"}
		schema := JWT(JWTOptions{Algorithm: "RS256"}, schemaParams)

		// This should fail because token uses HS256 but we expect RS256
		validJWT := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

		_, err := schema.Parse(validJWT)
		if err == nil {
			t.Error("Expected error for algorithm mismatch")
		}
	})
}

// =============================================================================
// Base64 TESTS
// =============================================================================

func TestBase64(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected bool
	}{
		{
			name:     "valid base64 string",
			input:    "SGVsbG8gV29ybGQ=", // "Hello World"
			expected: true,
		},
		{
			name:     "valid base64 with padding",
			input:    "Zm9vYg==", // "foob"
			expected: true,
		},
		{
			name:     "valid base64 without padding",
			input:    "Zm9vYmFy", // "foobar"
			expected: true,
		},
		{
			name:     "invalid base64 characters",
			input:    "SGVsbG8gV29ybGQ$=",
			expected: false,
		},
		{
			name:     "invalid base64 padding",
			input:    "SGVsbG8gV29ybGQ===",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: true, // Empty string is valid base64
		},
		{
			name:     "non-string input",
			input:    12345,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := Base64()
			result, err := schema.Parse(tt.input)

			if tt.expected {
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
	t.Run("valid base64 pointer", func(t *testing.T) {
		schema := Base64Ptr()
		valid := "SGVsbG8gV29ybGQ="
		result, err := schema.Parse(&valid)
		assert.NoError(t, err)
		assert.Equal(t, &valid, result)
	})

	t.Run("nil input", func(t *testing.T) {
		schema := Base64Ptr()
		_, err := schema.Parse(nil)
		assert.Error(t, err) // By default, ptr is not nilable
	})

	t.Run("invalid base64 pointer", func(t *testing.T) {
		schema := Base64Ptr()
		invalid := "invalid!"
		_, err := schema.Parse(&invalid)
		assert.Error(t, err)
	})
}

func TestBase64Chaining(t *testing.T) {
	schema := Base64().Min(10) // "SGVsbG8gV29ybGQ=" is 16 chars long
	valid := "SGVsbG8gV29ybGQ="
	_, err := schema.Parse(valid)
	assert.NoError(t, err)

	shortValid := "Zm9v" // "foo" is 4 chars long
	_, err = schema.Parse(shortValid)
	assert.Error(t, err)
}

func TestBase64WithCustomError(t *testing.T) {
	customError := "Invalid Base64 data"
	schema := Base64(customError)
	_, err := schema.Parse("invalid!")
	assert.Error(t, err)

	var zodErr3 *issues.ZodError
	if issues.IsZodError(err, &zodErr3) {
		if len(zodErr3.Issues) > 0 {
			assert.Contains(t, zodErr3.Issues[0].Message, customError)
		}
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
		schema := Base64().Nilable()
		result, err := schema.Parse(nil)
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Default", func(t *testing.T) {
		defaultVal := "SGVsbG8gV29ybGQ="
		schema := Base64().Default(defaultVal)
		result, err := schema.Parse(nil)
		assert.NoError(t, err)
		assert.Equal(t, defaultVal, result)
	})
}

// =============================================================================
// Base64URL TESTS
// =============================================================================

func TestBase64URL(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected bool
	}{
		{
			name:     "valid base64url string",
			input:    "SGVsbG8gV29ybGQ", // "Hello World" in base64url
			expected: true,
		},
		{
			name:     "base64 with + char is invalid in base64url",
			input:    "Zm9v+g==", // contains '+'
			expected: false,
		},
		{
			name:     "base64 with / char is invalid in base64url",
			input:    "Zm9v/g==", // contains '/'
			expected: false,
		},
		{
			name:     "valid base64url (no padding)",
			input:    "Zm9vYmFy", // "foobar"
			expected: true,
		},
		{
			name:     "valid base64url (with padding)",
			input:    "Zm9vYg==", // "foob"
			expected: true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: true, // Empty string is valid base64url
		},
		{
			name:     "non-string input",
			input:    12345,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := Base64URL()
			result, err := schema.Parse(tt.input)

			if tt.expected {
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
	t.Run("valid base64url pointer", func(t *testing.T) {
		schema := Base64URLPtr()
		valid := "SGVsbG8gV29ybGQ"
		result, err := schema.Parse(&valid)
		assert.NoError(t, err)
		assert.Equal(t, &valid, result)
	})

	t.Run("nil input", func(t *testing.T) {
		schema := Base64URLPtr()
		_, err := schema.Parse(nil)
		assert.Error(t, err)
	})

	t.Run("invalid base64url pointer", func(t *testing.T) {
		schema := Base64URLPtr()
		invalid := "not-valid-url-safe-base64+"
		_, err := schema.Parse(&invalid)
		assert.Error(t, err)
	})
}

func TestBase64URLChaining(t *testing.T) {
	schema := Base64URL().Min(10) // "SGVsbG8gV29ybGQ" is 15 chars
	valid := "SGVsbG8gV29ybGQ"
	_, err := schema.Parse(valid)
	assert.NoError(t, err)

	shortValid := "Zm9v" // "foo" is 4 chars long
	_, err = schema.Parse(shortValid)
	assert.Error(t, err)
}

func TestBase64URLWithCustomError(t *testing.T) {
	customError := "Invalid Base64URL data"
	schema := Base64URL(customError)
	_, err := schema.Parse("invalid+base64url")
	assert.Error(t, err)

	var zodErr4 *issues.ZodError
	if issues.IsZodError(err, &zodErr4) {
		if len(zodErr4.Issues) > 0 {
			assert.Contains(t, zodErr4.Issues[0].Message, customError)
		}
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
		schema := Base64URL().Nilable()
		result, err := schema.Parse(nil)
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Default", func(t *testing.T) {
		defaultVal := "SGVsbG8gV29ybGQ"
		schema := Base64URL().Default(defaultVal)
		result, err := schema.Parse(nil)
		assert.NoError(t, err)
		assert.Equal(t, defaultVal, result)
	})
}
