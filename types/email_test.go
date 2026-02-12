package types

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestZodEmail_BasicValidation(t *testing.T) {
	schema := Email()

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
		t.Run("valid/"+email, func(t *testing.T) {
			result, err := schema.Parse(email)
			require.NoError(t, err)
			assert.Equal(t, email, result)
		})
	}

	invalidEmails := []string{
		"francois@@etu.inp-n7.fr",
		`"email"@domain.com`,
		`"e asdf sadf ?<>ail"@domain.com`,
		`" "@example.org`,
		`"john..doe"@example.org`,
		`"very.(),:;<>[]\".VERY.\"very@\\ \"very\".unusual"@strange.example.com`,
		"a,b@domain.com",
		"email@123.123.123.123",
		"email@[123.123.123.123]",
		"postmaster@123.123.123.123",
		"user@[68.185.127.196]",
		"ipv4@[85.129.96.247]",
		"valid@[79.208.229.53]",
		"valid@[255.255.255.255]",
		"valid@[255.0.55.2]",
		"hgrebert0@[IPv6:4dc8:ac7:ce79:8878:1290:6098:5c50:1f25]",
		"bshapiro4@[IPv6:3669:c709:e981:4884:59a3:75d1:166b:9ae]",
		"jsmith@[IPv6:2001:db8::1]",
		"postmaster@[IPv6:2001:0db8:85a3:0000:0000:8a2e:0370:7334]",
		"postmaster@[IPv6:2001:0db8:85a3:0000:0000:8a2e:0370:192.168.1.1]",
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
		t.Run("invalid/"+email, func(t *testing.T) {
			_, err := schema.Parse(email)
			assert.Error(t, err, "expected %q to be rejected", email)
		})
	}
}

func TestZodEmail_Modifiers(t *testing.T) {
	t.Run("Optional", func(t *testing.T) {
		schema := Email().Optional()

		result, err := schema.Parse("test@example.com")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "test@example.com", *result)

		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Default", func(t *testing.T) {
		schema := Email().Default("default@example.com")

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "default@example.com", result)

		result, err = schema.Parse("test@example.com")
		require.NoError(t, err)
		assert.Equal(t, "test@example.com", result)
	})

	t.Run("DefaultFunc", func(t *testing.T) {
		schema := Email().DefaultFunc(func() string {
			return "dynamic@example.com"
		})

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "dynamic@example.com", result)
	})
}

func TestZodEmail_DefaultAndPrefault(t *testing.T) {
	t.Run("Default has higher priority than Prefault", func(t *testing.T) {
		schema := Email().Default("default@example.com").Prefault("prefault@example.com")

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "default@example.com", result)
	})

	t.Run("Default short-circuits validation", func(t *testing.T) {
		schema := Email().Refine(func(s string) bool {
			return false
		}, "Should never pass").Default("default@example.com")

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "default@example.com", result)
	})

	t.Run("Prefault requires full validation", func(t *testing.T) {
		schema := Email().Refine(func(s string) bool {
			return s == "valid@example.com"
		}, "Must be valid@example.com").Prefault("valid@example.com")

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "valid@example.com", result)
	})

	t.Run("Prefault only triggers for nil input", func(t *testing.T) {
		schema := Email().Prefault("prefault@example.com")

		_, err := schema.Parse("invalid-email")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid email")

		result, err := schema.Parse("test@example.com")
		require.NoError(t, err)
		assert.Equal(t, "test@example.com", result)
	})

	t.Run("DefaultFunc and PrefaultFunc behavior", func(t *testing.T) {
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
		require.NoError(t, err)
		assert.Equal(t, "default@example.com", result)
		assert.True(t, defaultCalled, "DefaultFunc should be called")
		assert.False(t, prefaultCalled, "PrefaultFunc should not be called when Default is present")
	})

	t.Run("Prefault validation failure", func(t *testing.T) {
		schema := Email().Refine(func(s string) bool {
			return false
		}, "Refinement failed").Prefault("prefault@example.com")

		_, err := schema.Parse(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Refinement")
	})
}

func TestZodEmail_TypeSafety(t *testing.T) {
	t.Run("String type", func(t *testing.T) {
		_ = Email()
		_ = EmailTyped[string]()
	})

	t.Run("Pointer type", func(t *testing.T) {
		_ = EmailPtr()
		_ = EmailTyped[*string]()
		_ = Email().Optional()
	})
}

func TestZodEmail_Patterns(t *testing.T) {
	tests := []struct {
		name    string
		schema  *ZodEmail[string]
		valid   string
		invalid string
	}{
		{"Default", Email(), "test@example.com", "invalid-email"},
		{"HTML5", Email().Html5(), "test@example.com", "invalid-email"},
		{"RFC5322", Email().Rfc5322(), "test@example.com", "invalid-email"},
		{"Unicode", Email().Unicode(), "test@example.com", "invalid-email"},
		{"Browser", Email().Browser(), "test@example.com", "invalid-email"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.schema.Parse(tt.valid)
			require.NoError(t, err)
			assert.Equal(t, tt.valid, result)

			_, err = tt.schema.Parse(tt.invalid)
			assert.Error(t, err)
		})
	}
}

func TestZodEmail_CustomPattern(t *testing.T) {
	strictPattern := regexp.MustCompile(`^[a-z]+@[a-z]+\.com$`)
	schema := Email(strictPattern)

	result, err := schema.Parse("test@example.com")
	require.NoError(t, err)
	assert.Equal(t, "test@example.com", result)

	_, err = schema.Parse("test123@example.com")
	assert.Error(t, err)
}

func TestZodEmail_UnicodeEmails(t *testing.T) {
	schema := Email().Unicode()

	validEmails := []string{
		"test@example.com",
		"用户@例子.com",
		"пользователь@пример.com",
	}

	for _, email := range validEmails {
		t.Run(email, func(t *testing.T) {
			result, err := schema.Parse(email)
			require.NoError(t, err)
			assert.Equal(t, email, result)
		})
	}
}

func TestZodEmail_BrowserPattern(t *testing.T) {
	schema := Email().Browser()

	invalidEmails := []string{
		"invalid-email",
		"@example.com",
		"user@",
		"user space@example.com",
	}

	for _, email := range invalidEmails {
		t.Run(email, func(t *testing.T) {
			_, err := schema.Parse(email)
			assert.Error(t, err)
		})
	}
}

func TestZodEmail_WithOptions(t *testing.T) {
	t.Run("Custom pattern via constructor", func(t *testing.T) {
		customPattern := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		schema := Email(customPattern)

		result, err := schema.Parse("test@example.com")
		require.NoError(t, err)
		assert.Equal(t, "test@example.com", result)
	})

	t.Run("Preset pattern via method", func(t *testing.T) {
		schema := Email().Html5()

		result, err := schema.Parse("test@example.com")
		require.NoError(t, err)
		assert.Equal(t, "test@example.com", result)
	})
}

func TestZodEmail_RealWorldUsage(t *testing.T) {
	t.Run("User registration form", func(t *testing.T) {
		type UserForm struct {
			Email       string  `json:"email"`
			BackupEmail *string `json:"backup_email,omitempty"`
		}

		emailSchema := Email("Email is required")
		backupSchema := Email().Optional()

		primary, err := emailSchema.Parse("user@example.com")
		require.NoError(t, err)

		backup, err := backupSchema.Parse("backup@example.com")
		require.NoError(t, err)

		nilBackup, err := backupSchema.Parse(nil)
		require.NoError(t, err)

		form := UserForm{Email: primary, BackupEmail: backup}
		assert.Equal(t, "user@example.com", form.Email)
		require.NotNil(t, form.BackupEmail)
		assert.Equal(t, "backup@example.com", *form.BackupEmail)

		formNil := UserForm{Email: primary, BackupEmail: nilBackup}
		assert.Nil(t, formNil.BackupEmail)
	})
}

func TestZodEmail_ErrorHandling(t *testing.T) {
	schema := Email("Custom email error message")

	_, err := schema.Parse("invalid-email")
	require.Error(t, err)

	invalidInputs := []any{123, true, []string{"test@example.com"}, map[string]string{"email": "test@example.com"}}

	for _, input := range invalidInputs {
		_, err := schema.Parse(input)
		assert.Error(t, err, "expected error for input %v", input)
	}
}

func TestZodEmail_Integration(t *testing.T) {
	t.Run("Chaining modifiers", func(t *testing.T) {
		schema := Email().Default("default@example.com").Optional()

		result, err := schema.Parse("test@example.com")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "test@example.com", *result)

		result, err = schema.Parse(nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "default@example.com", *result)
	})

	t.Run("String validations", func(t *testing.T) {
		schema := Email().Min(10).Max(50)

		result, err := schema.Parse("test@example.com")
		require.NoError(t, err)
		assert.Equal(t, "test@example.com", result)

		_, err = schema.Parse("a@b.co")
		assert.Error(t, err)

		_, err = schema.Parse("verylongusernamethatexceedslimit@verylongdomainthatexceedslimit.com")
		assert.Error(t, err)
	})
}

func TestZodEmail_BackwardCompatibility(t *testing.T) {
	schemas := map[string]*ZodEmail[string]{
		"Email()":              Email(),
		"EmailTyped[string]()": EmailTyped[string](),
		"Email().Html5()":      Email().Html5(),
		"Email().Rfc5322()":    Email().Rfc5322(),
		"Email().Unicode()":    Email().Unicode(),
		"Email().Browser()":    Email().Browser(),
	}

	for name, schema := range schemas {
		t.Run(name, func(t *testing.T) {
			result, err := schema.Parse("test@example.com")
			require.NoError(t, err)
			assert.Equal(t, "test@example.com", result)
		})
	}

	ptrSchemas := map[string]*ZodEmail[*string]{
		"EmailPtr()":            EmailPtr(),
		"EmailTyped[*string]()": EmailTyped[*string](),
	}

	for name, schema := range ptrSchemas {
		t.Run(name+" rejects nil", func(t *testing.T) {
			_, err := schema.Parse(nil)
			assert.Error(t, err)
		})
	}

	optionalSchemas := map[string]*ZodEmail[*string]{
		"Email().Optional()":    Email().Optional(),
		"EmailPtr().Optional()": EmailPtr().Optional(),
		"EmailPtr().Nilable()":  EmailPtr().Nilable(),
	}

	for name, schema := range optionalSchemas {
		t.Run(name+" accepts nil", func(t *testing.T) {
			result, err := schema.Parse(nil)
			require.NoError(t, err)
			assert.Nil(t, result)
		})
	}
}

func TestZodEmail_EdgeCases(t *testing.T) {
	schema := Email()

	tests := []struct {
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := schema.Parse(tt.input)
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestZodEmail_Constructors(t *testing.T) {
	t.Run("Email()", func(t *testing.T) {
		result, err := Email().Parse("test@example.com")
		require.NoError(t, err)
		assert.Equal(t, "test@example.com", result)
	})

	t.Run("EmailPtr()", func(t *testing.T) {
		result, err := EmailPtr().Parse("test@example.com")
		require.NoError(t, err)
		assert.Equal(t, "test@example.com", *result)
	})

	t.Run("EmailTyped()", func(t *testing.T) {
		result, err := EmailTyped[string]().Parse("test@example.com")
		require.NoError(t, err)
		assert.Equal(t, "test@example.com", result)
	})
}
