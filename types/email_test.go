package types

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Basic Email Validation Tests
// =============================================================================

func TestZodEmail_BasicValidation(t *testing.T) {
	schema := Email()

	// Valid emails based on TypeScript test cases
	validEmails := []string{
		"email@domain.com",
		"firstname.lastname@domain.com",
		"email@subdomain.domain.com",
		"firstname+lastname@domain.com",
		"1234567890@domain.com",
		"email@domain-one.com",
		"_______@domain.com",
		"email@domain.name",
		"email@domain.co.jp",
		"firstname-lastname@domain.com",
		"very.common@example.com",
		"disposable.style.email.with+symbol@example.com",
		"other.email-with-hyphen@example.com",
		"fully-qualified-domain@example.com",
		"user.name+tag+sorting@example.com",
		"x@example.com",
		"mojojojo@asdf.example.com",
		"example-indeed@strange-example.com",
		"example@s.example",
		"user-@example.org",
		"user@my-example.com",
		"a@b.cd",
		"work+user@mail.com",
		"tom@test.te-st.com",
		"something@subdomain.domain-with-hyphens.tld",
		"common'name@domain.com",
		"francois@etu.inp-n7.fr",
	}

	for _, email := range validEmails {
		t.Run("Valid email: "+email, func(t *testing.T) {
			result, err := schema.Parse(email)
			if err != nil {
				t.Errorf("Expected valid email %q to pass, but got error: %v", email, err)
			}
			if result != email {
				t.Errorf("Expected %q, got %q", email, result)
			}
		})
	}

	// Invalid emails based on TypeScript test cases
	invalidEmails := []string{
		// Double @
		"francois@@etu.inp-n7.fr",
		// Quotes not supported
		`"email"@domain.com`,
		`"e asdf sadf ?<>ail"@domain.com`,
		`" "@example.org`,
		`"john..doe"@example.org`,
		`"very.(),:;<>[]\".VERY.\"very@\\ \"very\".unusual"@strange.example.com`,
		// Comma not supported
		"a,b@domain.com",
		// IPv4 not supported by default
		"email@123.123.123.123",
		"email@[123.123.123.123]",
		"postmaster@123.123.123.123",
		"user@[68.185.127.196]",
		"ipv4@[85.129.96.247]",
		"valid@[79.208.229.53]",
		"valid@[255.255.255.255]",
		"valid@[255.0.55.2]",
		// IPv6 not supported by default
		"hgrebert0@[IPv6:4dc8:ac7:ce79:8878:1290:6098:5c50:1f25]",
		"bshapiro4@[IPv6:3669:c709:e981:4884:59a3:75d1:166b:9ae]",
		"jsmith@[IPv6:2001:db8::1]",
		"postmaster@[IPv6:2001:0db8:85a3:0000:0000:8a2e:0370:7334]",
		"postmaster@[IPv6:2001:0db8:85a3:0000:0000:8a2e:0370:192.168.1.1]",
		// Microsoft test cases
		"plainaddress",
		"#@%^%#$@#$@#.com",
		"@domain.com",
		"Joe Smith <email@domain.com>",
		"email.domain.com",
		"email@domain@domain.com",
		".email@domain.com",
		"email.@domain.com",
		"email..email@domain.com",
		"あいうえお@domain.com",
		"email@domain.com (Joe Smith)",
		"email@domain",
		"email@-domain.com",
		"email@111.222.333.44444",
		"email@domain..com",
		"Abc.example.com",
		"A@b@c@example.com",
		"colin..hacks@domain.com",
		`a"b(c)d,e:f;g<h>i[j\k]l@example.com`,
		`just"not"right@example.com`,
		`this is"not\allowed@example.com`,
		`this\ still\"not\\allowed@example.com`,
	}

	for _, email := range invalidEmails {
		t.Run("Invalid email: "+email, func(t *testing.T) {
			_, err := schema.Parse(email)
			if err == nil {
				t.Errorf("Expected invalid email %q to fail, but it passed", email)
			}
		})
	}
}

// =============================================================================
// Email Modifier Tests
// =============================================================================

func TestZodEmail_Modifiers(t *testing.T) {
	t.Run("Optional", func(t *testing.T) {
		schema := Email().Optional()

		// Should accept valid email
		result, err := schema.Parse("test@example.com")
		if err != nil {
			t.Errorf("Expected valid email to pass: %v", err)
		}
		if result == nil || *result != "test@example.com" {
			t.Errorf("Expected 'test@example.com', got %v", result)
		}

		// Should accept nil
		result, err = schema.Parse(nil)
		if err != nil {
			t.Errorf("Expected nil to pass for optional: %v", err)
		}
		if result != nil {
			t.Errorf("Expected nil, got %v", result)
		}
	})

	t.Run("Default", func(t *testing.T) {
		schema := Email().Default("default@example.com")

		// Should use default for nil
		result, err := schema.Parse(nil)
		if err != nil {
			t.Errorf("Expected nil to use default: %v", err)
		}
		if result != "default@example.com" {
			t.Errorf("Expected 'default@example.com', got %q", result)
		}

		// Should use provided value when valid
		result, err = schema.Parse("test@example.com")
		if err != nil {
			t.Errorf("Expected valid email to pass: %v", err)
		}
		if result != "test@example.com" {
			t.Errorf("Expected 'test@example.com', got %q", result)
		}
	})

	t.Run("DefaultFunc", func(t *testing.T) {
		schema := Email().DefaultFunc(func() string {
			return "dynamic@example.com"
		})

		result, err := schema.Parse(nil)
		if err != nil {
			t.Errorf("Expected nil to use default function: %v", err)
		}
		if result != "dynamic@example.com" {
			t.Errorf("Expected 'dynamic@example.com', got %q", result)
		}
	})
}

// =============================================================================
// Email Default and Prefault Tests
// =============================================================================

func TestEmail_DefaultAndPrefault(t *testing.T) {
	t.Run("Default has higher priority than Prefault", func(t *testing.T) {
		// When both Default and Prefault are set, Default should take precedence for nil input
		schema := Email().Default("default@example.com").Prefault("prefault@example.com")

		result, err := schema.Parse(nil)
		assert.NoError(t, err)
		assert.Equal(t, "default@example.com", result)
	})

	t.Run("Default short-circuits validation", func(t *testing.T) {
		// Default value should bypass email validation constraints
		schema := Email().Refine(func(s string) bool {
			return false // Always fail refinement
		}, "Should never pass").Default("default@example.com")

		result, err := schema.Parse(nil)
		assert.NoError(t, err)
		assert.Equal(t, "default@example.com", result)
	})

	t.Run("Prefault requires full validation", func(t *testing.T) {
		// Prefault value must pass all email validation including refinements
		schema := Email().Refine(func(s string) bool {
			return s == "valid@example.com"
		}, "Must be valid@example.com").Prefault("valid@example.com")

		result, err := schema.Parse(nil)
		assert.NoError(t, err)
		assert.Equal(t, "valid@example.com", result)
	})

	t.Run("Prefault only triggers for nil input", func(t *testing.T) {
		// Non-nil input that fails validation should not trigger Prefault
		schema := Email().Prefault("prefault@example.com")

		// Invalid email should fail without triggering Prefault
		_, err := schema.Parse("invalid-email")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid email")

		// Valid email should pass normally
		result, err := schema.Parse("test@example.com")
		assert.NoError(t, err)
		assert.Equal(t, "test@example.com", result)
	})

	t.Run("DefaultFunc and PrefaultFunc behavior", func(t *testing.T) {
		// Test function call behavior and priority
		defaultCalled := false
		prefaultCalled := false

		schema := Email().
			DefaultFunc(func() string {
				defaultCalled = true
				return "default@example.com"
			}).
			PrefaultFunc(func() string {
				prefaultCalled = true
				return "prefault@example.com"
			})

		result, err := schema.Parse(nil)
		assert.NoError(t, err)
		assert.Equal(t, "default@example.com", result)
		assert.True(t, defaultCalled, "DefaultFunc should be called")
		assert.False(t, prefaultCalled, "PrefaultFunc should not be called when Default is present")
	})

	t.Run("Prefault validation failure", func(t *testing.T) {
		// When Prefault value fails validation, should return error
		schema := Email().Refine(func(s string) bool {
			return false // Always fail
		}, "Refinement failed").Prefault("prefault@example.com")

		_, err := schema.Parse(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Refinement")
	})
}

// =============================================================================
// Email Type Safety Tests
// =============================================================================

func TestZodEmail_TypeSafety(t *testing.T) {
	t.Run("String type", func(t *testing.T) {
		var _ *ZodEmail[string] = Email()
		var _ *ZodEmail[string] = EmailTyped[string]()
	})

	t.Run("Pointer type", func(t *testing.T) {
		var _ *ZodEmail[*string] = EmailPtr()
		var _ *ZodEmail[*string] = EmailTyped[*string]()
		var _ *ZodEmail[*string] = Email().Optional()
	})

	t.Run("Non-string types should not compile", func(t *testing.T) {
		// These should cause compilation errors if uncommented:
		// var _ *ZodEmail[int] = EmailTyped[int]()
		// var _ *ZodEmail[bool] = EmailTyped[bool]()
	})
}

// =============================================================================
// Email Pattern Tests
// =============================================================================

func TestZodEmail_Patterns(t *testing.T) {
	testCases := []struct {
		name         string
		schema       *ZodEmail[string]
		validEmail   string
		invalidEmail string
	}{
		{
			name:         "Default pattern",
			schema:       Email(),
			validEmail:   "test@example.com",
			invalidEmail: "invalid-email",
		},
		{
			name:         "HTML5 pattern",
			schema:       Email().Html5(),
			validEmail:   "test@example.com",
			invalidEmail: "invalid-email",
		},
		{
			name:         "RFC5322 pattern",
			schema:       Email().Rfc5322(),
			validEmail:   "test@example.com",
			invalidEmail: "invalid-email",
		},
		{
			name:         "Unicode pattern",
			schema:       Email().Unicode(),
			validEmail:   "test@example.com",
			invalidEmail: "invalid-email",
		},
		{
			name:         "Browser pattern",
			schema:       Email().Browser(),
			validEmail:   "test@example.com",
			invalidEmail: "invalid-email",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test valid email
			result, err := tc.schema.Parse(tc.validEmail)
			if err != nil {
				t.Errorf("Expected valid email to pass: %v", err)
			}
			if result != tc.validEmail {
				t.Errorf("Expected %q, got %q", tc.validEmail, result)
			}

			// Test invalid email
			_, err = tc.schema.Parse(tc.invalidEmail)
			if err == nil {
				t.Errorf("Expected invalid email to fail")
			}
		})
	}
}

func TestZodEmail_CustomPattern(t *testing.T) {
	// Create a very strict pattern that only allows specific format
	strictPattern := regexp.MustCompile(`^[a-z]+@[a-z]+\.com$`)
	schema := Email(strictPattern)

	// This should pass the strict pattern
	result, err := schema.Parse("test@example.com")
	if err != nil {
		t.Errorf("Expected email to pass strict pattern: %v", err)
	}
	if result != "test@example.com" {
		t.Errorf("Expected 'test@example.com', got %q", result)
	}

	// This should fail the strict pattern (contains numbers)
	_, err = schema.Parse("test123@example.com")
	if err == nil {
		t.Errorf("Expected email with numbers to fail strict pattern")
	}
}

func TestZodEmail_PresetPatterns(t *testing.T) {
	t.Run("Html5 method", func(t *testing.T) {
		schema := Email().Html5()
		result, err := schema.Parse("test@example.com")
		if err != nil {
			t.Errorf("Expected valid email: %v", err)
		}
		if result != "test@example.com" {
			t.Errorf("Expected 'test@example.com', got %q", result)
		}
	})

	t.Run("Rfc5322 method", func(t *testing.T) {
		schema := Email().Rfc5322()
		result, err := schema.Parse("test@example.com")
		if err != nil {
			t.Errorf("Expected valid email: %v", err)
		}
		if result != "test@example.com" {
			t.Errorf("Expected 'test@example.com', got %q", result)
		}
	})

	t.Run("Unicode method", func(t *testing.T) {
		schema := Email().Unicode()
		result, err := schema.Parse("test@example.com")
		if err != nil {
			t.Errorf("Expected valid email: %v", err)
		}
		if result != "test@example.com" {
			t.Errorf("Expected 'test@example.com', got %q", result)
		}
	})

	t.Run("Browser method", func(t *testing.T) {
		schema := Email().Browser()
		result, err := schema.Parse("test@example.com")
		if err != nil {
			t.Errorf("Expected valid email: %v", err)
		}
		if result != "test@example.com" {
			t.Errorf("Expected 'test@example.com', got %q", result)
		}
	})
}

// =============================================================================
// Email Options Tests
// =============================================================================

func TestZodEmail_WithOptions(t *testing.T) {
	t.Run("Custom pattern via simplified API", func(t *testing.T) {
		customPattern := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		schema := Email(customPattern)

		result, err := schema.Parse("test@example.com")
		if err != nil {
			t.Errorf("Expected valid email: %v", err)
		}
		if result != "test@example.com" {
			t.Errorf("Expected 'test@example.com', got %q", result)
		}
	})

	t.Run("Preset pattern via method", func(t *testing.T) {
		schema := Email().Html5()

		result, err := schema.Parse("test@example.com")
		if err != nil {
			t.Errorf("Expected valid email: %v", err)
		}
		if result != "test@example.com" {
			t.Errorf("Expected 'test@example.com', got %q", result)
		}
	})
}

// =============================================================================
// Real World Usage Tests
// =============================================================================

func TestEmail_RealWorldUsage(t *testing.T) {
	t.Run("User registration form", func(t *testing.T) {
		type UserForm struct {
			Email       string  `json:"email"`
			BackupEmail *string `json:"backup_email,omitempty"`
		}

		emailSchema := Email("Email is required")
		backupEmailSchema := Email().Optional()

		// Test primary email
		primaryEmail := "user@example.com"
		validatedEmail, err := emailSchema.Parse(primaryEmail)
		if err != nil {
			t.Errorf("Expected valid primary email: %v", err)
		}

		// Test backup email (optional)
		backupEmail := "backup@example.com"
		validatedBackup, err := backupEmailSchema.Parse(backupEmail)
		if err != nil {
			t.Errorf("Expected valid backup email: %v", err)
		}

		// Test nil backup email
		nilBackup, err := backupEmailSchema.Parse(nil)
		if err != nil {
			t.Errorf("Expected nil backup email to be valid: %v", err)
		}

		form := UserForm{
			Email:       validatedEmail,
			BackupEmail: validatedBackup,
		}

		if form.Email != "user@example.com" {
			t.Errorf("Expected 'user@example.com', got %q", form.Email)
		}
		if form.BackupEmail == nil || *form.BackupEmail != "backup@example.com" {
			t.Errorf("Expected 'backup@example.com', got %v", form.BackupEmail)
		}

		// Test with nil backup
		formWithNilBackup := UserForm{
			Email:       validatedEmail,
			BackupEmail: nilBackup,
		}
		if formWithNilBackup.BackupEmail != nil {
			t.Errorf("Expected nil backup email, got %v", formWithNilBackup.BackupEmail)
		}
	})
}

// =============================================================================
// Error Handling Tests
// =============================================================================

func TestEmail_ErrorHandling(t *testing.T) {
	schema := Email("Custom email error message")

	_, err := schema.Parse("invalid-email")
	if err == nil {
		t.Error("Expected error for invalid email")
	}

	// Test with different invalid types
	invalidInputs := []any{
		123,
		true,
		[]string{"test@example.com"},
		map[string]string{"email": "test@example.com"},
	}

	for _, input := range invalidInputs {
		t.Run("Invalid type", func(t *testing.T) {
			_, err := schema.Parse(input)
			if err == nil {
				t.Errorf("Expected error for non-string input: %v", input)
			}
		})
	}
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestEmail_Integration(t *testing.T) {
	t.Run("Chaining modifiers", func(t *testing.T) {
		schema := Email().
			Default("default@example.com").
			Optional()

		// Test with valid email
		result, err := schema.Parse("test@example.com")
		if err != nil {
			t.Errorf("Expected valid email: %v", err)
		}
		if result == nil || *result != "test@example.com" {
			t.Errorf("Expected 'test@example.com', got %v", result)
		}

		// Test with nil (Default should short-circuit)
		result, err = schema.Parse(nil)
		if err != nil {
			t.Errorf("Expected default to be used: %v", err)
		}
		if result == nil || *result != "default@example.com" {
			t.Errorf("Expected 'default@example.com' (Default short-circuit), got %v", result)
		}
	})
}

// =============================================================================
// Backward Compatibility Tests
// =============================================================================

func TestEmail_BackwardCompatibility(t *testing.T) {
	// Test that all constructor functions work as expected
	schemas := map[string]*ZodEmail[string]{
		"Email()":              Email(),
		"EmailTyped[string]()": EmailTyped[string](),
		"Email().Html5()":      Email().Html5(),
		"Email().Rfc5322()":    Email().Rfc5322(),
		"Email().Unicode()":    Email().Unicode(),
		"Email().Browser()":    Email().Browser(),
	}

	validEmail := "test@example.com"

	for name, schema := range schemas {
		t.Run(name, func(t *testing.T) {
			result, err := schema.Parse(validEmail)
			if err != nil {
				t.Errorf("Expected %s to work with valid email: %v", name, err)
			}
			if result != validEmail {
				t.Errorf("Expected %q, got %q", validEmail, result)
			}
		})
	}

	// Test pointer types - they should NOT accept nil unless Optional/Nilable
	ptrSchemas := map[string]*ZodEmail[*string]{
		"EmailPtr()":            EmailPtr(),
		"EmailTyped[*string]()": EmailTyped[*string](),
	}

	for name, schema := range ptrSchemas {
		t.Run(name+" with nil", func(t *testing.T) {
			_, err := schema.Parse(nil)
			if err == nil {
				t.Errorf("Expected %s to reject nil (pointer types should not accept nil unless Optional/Nilable)", name)
			}
		})
	}

	// Test Optional pointer types - they should accept nil
	optionalPtrSchemas := map[string]*ZodEmail[*string]{
		"Email().Optional()":    Email().Optional(),
		"EmailPtr().Optional()": EmailPtr().Optional(),
		"EmailPtr().Nilable()":  EmailPtr().Nilable(),
	}

	for name, schema := range optionalPtrSchemas {
		t.Run(name+" with nil", func(t *testing.T) {
			result, err := schema.Parse(nil)
			if err != nil {
				t.Errorf("Expected %s to work with nil: %v", name, err)
			}
			if result != nil {
				t.Errorf("Expected nil, got %v", result)
			}
		})
	}
}

// =============================================================================
// Edge Cases and Validation Tests
// =============================================================================

func TestEmail_EdgeCasesAndValidation(t *testing.T) {
	schema := Email()

	edgeCases := []struct {
		name  string
		input string
		valid bool
	}{
		{"Empty string", "", false},
		{"Only @", "@", false},
		{"Only domain", "example.com", false},
		{"Only local part", "test", false},
		{"Multiple @", "test@domain@com", false},
		{"Starting with @", "@test.com", false},
		{"Ending with @", "test@", false},
		{"Space in email", "test @example.com", false},
		{"Tab in email", "test\t@example.com", false},
		{"Newline in email", "test\n@example.com", false},
		{"Very long email", string(make([]byte, 1000)) + "@example.com", false},
		{"Minimum valid", "a@b.co", true},
	}

	for _, tc := range edgeCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := schema.Parse(tc.input)
			if tc.valid && err != nil {
				t.Errorf("Expected %q to be valid, but got error: %v", tc.input, err)
			}
			if !tc.valid && err == nil {
				t.Errorf("Expected %q to be invalid, but it passed", tc.input)
			}
		})
	}
}

func TestEmailValidation(t *testing.T) {
	t.Run("Basic email validation", func(t *testing.T) {
		schema := Email()

		// Valid emails
		validEmails := []string{
			"test@example.com",
			"user.name@domain.co.uk",
			"user+tag@example.org",
			"user_name@example.com",
		}

		for _, email := range validEmails {
			t.Run("Valid: "+email, func(t *testing.T) {
				result, err := schema.Parse(email)
				assert.NoError(t, err)
				assert.Equal(t, email, result)
			})
		}

		// Invalid emails
		invalidEmails := []string{
			"invalid-email",
			"@example.com",
			"user@",
			"user..name@example.com",
			"user@.com",
		}

		for _, email := range invalidEmails {
			t.Run("Invalid: "+email, func(t *testing.T) {
				_, err := schema.Parse(email)
				assert.Error(t, err)
			})
		}
	})

	t.Run("HTML5 email validation", func(t *testing.T) {
		schema := Email().Html5()

		// Valid HTML5 emails
		validEmails := []string{
			"test@example.com",
			"user.name@domain.co.uk",
			"user+tag@example.org",
		}

		for _, email := range validEmails {
			t.Run("Valid HTML5: "+email, func(t *testing.T) {
				result, err := schema.Parse(email)
				assert.NoError(t, err)
				assert.Equal(t, email, result)
			})
		}
	})

	t.Run("RFC5322 email validation", func(t *testing.T) {
		schema := Email().Rfc5322()

		// Valid RFC5322 emails
		validEmails := []string{
			"test@example.com",
			"user.name@domain.co.uk",
			"user+tag@example.org",
		}

		for _, email := range validEmails {
			t.Run("Valid RFC5322: "+email, func(t *testing.T) {
				result, err := schema.Parse(email)
				assert.NoError(t, err)
				assert.Equal(t, email, result)
			})
		}
	})

	t.Run("Unicode email validation", func(t *testing.T) {
		schema := Email().Unicode()

		// Valid Unicode emails that match the simplified regex
		validEmails := []string{
			"test@example.com",
			"用户@例子.com",               // Use .com instead of .测试
			"пользователь@пример.com", // Use .com instead of .тест
		}

		for _, email := range validEmails {
			t.Run("Valid Unicode: "+email, func(t *testing.T) {
				result, err := schema.Parse(email)
				assert.NoError(t, err)
				assert.Equal(t, email, result)
			})
		}
	})

	t.Run("Browser email validation", func(t *testing.T) {
		schema := Email().Browser()

		// Valid browser emails (more permissive)
		validEmails := []string{
			"test@example.com",
			"user@domain.co.uk",
			"simple@test.org",
		}

		for _, email := range validEmails {
			t.Run("Valid Browser: "+email, func(t *testing.T) {
				result, err := schema.Parse(email)
				assert.NoError(t, err)
				assert.Equal(t, email, result)
			})
		}

		// Invalid browser emails
		invalidEmails := []string{
			"invalid-email",
			"@example.com",
			"user@",
			"user space@example.com",
		}

		for _, email := range invalidEmails {
			t.Run("Invalid Browser: "+email, func(t *testing.T) {
				_, err := schema.Parse(email)
				assert.Error(t, err)
			})
		}
	})

	t.Run("Email with modifiers", func(t *testing.T) {
		t.Run("Optional email", func(t *testing.T) {
			schema := Email().Optional()

			// Test with valid email
			result, err := schema.Parse("test@example.com")
			require.NoError(t, err)
			assert.Equal(t, "test@example.com", *result)

			// Test with nil (should be valid for optional)
			result, err = schema.Parse(nil)
			require.NoError(t, err)
			assert.Nil(t, result)
		})

		t.Run("Email with default", func(t *testing.T) {
			schema := Email().Default("default@example.com")

			// Test with valid email
			result, err := schema.Parse("test@example.com")
			require.NoError(t, err)
			assert.Equal(t, "test@example.com", result)

			// Test with nil (should return default)
			result, err = schema.Parse(nil)
			require.NoError(t, err)
			assert.Equal(t, "default@example.com", result)
		})

		t.Run("Email with string validations", func(t *testing.T) {
			schema := Email().Min(10).Max(50)

			// Test with valid email of appropriate length
			result, err := schema.Parse("test@example.com")
			require.NoError(t, err)
			assert.Equal(t, "test@example.com", result)

			// Test with too short email
			_, err = schema.Parse("a@b.co")
			assert.Error(t, err)

			// Test with too long email
			longEmail := "verylongusernamethatexceedslimit@verylongdomainthatexceedslimit.com"
			_, err = schema.Parse(longEmail)
			assert.Error(t, err)
		})
	})

	t.Run("Email with custom regex", func(t *testing.T) {
		// Test that custom regex still works with the Regex method
		schema := Email()

		// Test basic functionality
		result, err := schema.Parse("test@example.com")
		require.NoError(t, err)
		assert.Equal(t, "test@example.com", result)
	})
}

func TestEmailConstructors(t *testing.T) {
	t.Run("Email() constructor", func(t *testing.T) {
		schema := Email()
		result, err := schema.Parse("test@example.com")
		require.NoError(t, err)
		assert.Equal(t, "test@example.com", result)
	})

	t.Run("EmailPtr() constructor", func(t *testing.T) {
		schema := EmailPtr()
		result, err := schema.Parse("test@example.com")
		require.NoError(t, err)
		assert.Equal(t, "test@example.com", *result)
	})

	t.Run("EmailTyped() constructor", func(t *testing.T) {
		schema := EmailTyped[string]()
		result, err := schema.Parse("test@example.com")
		require.NoError(t, err)
		assert.Equal(t, "test@example.com", result)
	})
}
