package checks

import (
	"testing"

	"github.com/kaptinlin/gozod/core"
)

func TestFormatChecks(t *testing.T) {
	t.Run("Email validates email format", func(t *testing.T) {
		check := Email()

		// Test valid case
		payload := &core.ParsePayload{
			Value:  "test@example.com",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 0 {
			t.Errorf("Expected no issues for valid email, got %d", len(payload.Issues))
		}

		// Test invalid case
		payload = &core.ParsePayload{
			Value:  "not-an-email",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 1 {
			t.Errorf("Expected 1 issue for invalid email, got %d", len(payload.Issues))
		}
	})

	t.Run("URL validates URL format", func(t *testing.T) {
		check := URL()

		// Test valid case
		payload := &core.ParsePayload{
			Value:  "https://example.com",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 0 {
			t.Errorf("Expected no issues for valid URL, got %d", len(payload.Issues))
		}

		// Test invalid case
		payload = &core.ParsePayload{
			Value:  "not-a-url",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 1 {
			t.Errorf("Expected 1 issue for invalid URL, got %d", len(payload.Issues))
		}
	})

	t.Run("UUID validates UUID format", func(t *testing.T) {
		check := UUID()

		// Test valid case
		payload := &core.ParsePayload{
			Value:  "123e4567-e89b-12d3-a456-426614174000",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 0 {
			t.Errorf("Expected no issues for valid UUID, got %d", len(payload.Issues))
		}

		// Test invalid case
		payload = &core.ParsePayload{
			Value:  "not-a-uuid",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 1 {
			t.Errorf("Expected 1 issue for invalid UUID, got %d", len(payload.Issues))
		}
	})

	t.Run("GUID validates GUID format", func(t *testing.T) {
		check := GUID()

		// Test valid case - GUID format is same as UUID
		payload := &core.ParsePayload{
			Value:  "123e4567-e89b-12d3-a456-426614174000",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 0 {
			t.Errorf("Expected no issues for valid GUID, got %d", len(payload.Issues))
		}

		// Test invalid case
		payload = &core.ParsePayload{
			Value:  "not-a-guid",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 1 {
			t.Errorf("Expected 1 issue for invalid GUID, got %d", len(payload.Issues))
		}
	})
}

func TestIPAddressChecks(t *testing.T) {
	t.Run("IPv4 validates IPv4 format", func(t *testing.T) {
		check := IPv4()

		// Test valid case
		payload := &core.ParsePayload{
			Value:  "192.168.1.1",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 0 {
			t.Errorf("Expected no issues for valid IPv4, got %d", len(payload.Issues))
		}

		// Test invalid case
		payload = &core.ParsePayload{
			Value:  "192.168.1.256",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 1 {
			t.Errorf("Expected 1 issue for invalid IPv4, got %d", len(payload.Issues))
		}
	})

	t.Run("IPv6 validates IPv6 format", func(t *testing.T) {
		check := IPv6()

		// Test valid case
		payload := &core.ParsePayload{
			Value:  "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 0 {
			t.Errorf("Expected no issues for valid IPv6, got %d", len(payload.Issues))
		}

		// Test invalid case
		payload = &core.ParsePayload{
			Value:  "not-an-ipv6",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 1 {
			t.Errorf("Expected 1 issue for invalid IPv6, got %d", len(payload.Issues))
		}
	})
}

func TestCIDRChecks(t *testing.T) {
	t.Run("CIDRv4 validates IPv4 CIDR notation", func(t *testing.T) {
		check := CIDRv4()

		// Test valid case
		payload := &core.ParsePayload{
			Value:  "192.168.1.0/24",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 0 {
			t.Errorf("Expected no issues for valid CIDRv4, got %d", len(payload.Issues))
		}

		// Test invalid case
		payload = &core.ParsePayload{
			Value:  "192.168.1.1",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 1 {
			t.Errorf("Expected 1 issue for invalid CIDRv4, got %d", len(payload.Issues))
		}
	})

	t.Run("CIDRv6 validates IPv6 CIDR notation", func(t *testing.T) {
		check := CIDRv6()

		// Test valid case - full IPv6 address with CIDR notation
		payload := &core.ParsePayload{
			Value:  "2001:0db8:85a3:0000:0000:8a2e:0370:7334/64",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 0 {
			t.Errorf("Expected no issues for valid CIDRv6, got %d", len(payload.Issues))
		}

		// Test invalid case - IPv6 address without CIDR notation
		payload = &core.ParsePayload{
			Value:  "2001:db8::1",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 1 {
			t.Errorf("Expected 1 issue for invalid CIDRv6, got %d", len(payload.Issues))
		}
	})
}

func TestEncodingChecks(t *testing.T) {
	t.Run("Base64 validates base64 encoding", func(t *testing.T) {
		check := Base64()

		// Test valid case
		payload := &core.ParsePayload{
			Value:  "SGVsbG8gV29ybGQ=",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 0 {
			t.Errorf("Expected no issues for valid base64, got %d", len(payload.Issues))
		}

		// Test invalid case
		payload = &core.ParsePayload{
			Value:  "not-base64!",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 1 {
			t.Errorf("Expected 1 issue for invalid base64, got %d", len(payload.Issues))
		}
	})

	t.Run("JWT validates JWT token format", func(t *testing.T) {
		check := JWT()

		// Test valid case (simplified JWT structure)
		payload := &core.ParsePayload{
			Value:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 0 {
			t.Errorf("Expected no issues for valid JWT, got %d", len(payload.Issues))
		}

		// Test invalid case - missing dots
		payload = &core.ParsePayload{
			Value:  "not-a-jwt-missing-dots",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 1 {
			t.Errorf("Expected 1 issue for invalid JWT, got %d", len(payload.Issues))
		}
	})
}

func TestPhoneNumberChecks(t *testing.T) {
	t.Run("E164 validates international phone number format", func(t *testing.T) {
		check := E164()

		// Test valid case
		payload := &core.ParsePayload{
			Value:  "+1234567890",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 0 {
			t.Errorf("Expected no issues for valid E164, got %d", len(payload.Issues))
		}

		// Test invalid case
		payload = &core.ParsePayload{
			Value:  "1234567890",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 1 {
			t.Errorf("Expected 1 issue for invalid E164, got %d", len(payload.Issues))
		}
	})
}

func TestDateTimeChecks(t *testing.T) {
	t.Run("ISODateTime validates ISO datetime format", func(t *testing.T) {
		check := ISODateTime()

		// Test valid case
		payload := &core.ParsePayload{
			Value:  "2023-12-31T23:59:59Z",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 0 {
			t.Errorf("Expected no issues for valid ISO datetime, got %d", len(payload.Issues))
		}

		// Test invalid case
		payload = &core.ParsePayload{
			Value:  "not-a-datetime",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 1 {
			t.Errorf("Expected 1 issue for invalid datetime, got %d", len(payload.Issues))
		}
	})

	t.Run("ISODate validates ISO date format", func(t *testing.T) {
		check := ISODate()

		// Test valid case
		payload := &core.ParsePayload{
			Value:  "2023-12-31",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 0 {
			t.Errorf("Expected no issues for valid ISO date, got %d", len(payload.Issues))
		}

		// Test invalid case
		payload = &core.ParsePayload{
			Value:  "not-a-date",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 1 {
			t.Errorf("Expected 1 issue for invalid date, got %d", len(payload.Issues))
		}
	})

	t.Run("ISODuration validates ISO 8601 duration format", func(t *testing.T) {
		check := ISODuration()

		// Test valid case
		payload := &core.ParsePayload{
			Value:  "P1Y2M3DT4H5M6S",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 0 {
			t.Errorf("Expected no issues for valid ISO duration, got %d", len(payload.Issues))
		}

		// Test invalid case
		payload = &core.ParsePayload{
			Value:  "not-a-duration",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 1 {
			t.Errorf("Expected 1 issue for invalid duration, got %d", len(payload.Issues))
		}
	})
}

func TestIDFormatChecks(t *testing.T) {
	t.Run("CUID validates CUID format", func(t *testing.T) {
		check := CUID()

		// Test valid case
		payload := &core.ParsePayload{
			Value:  "cjld2cjxh0000qzrmn831i7rn",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 0 {
			t.Errorf("Expected no issues for valid CUID, got %d", len(payload.Issues))
		}

		// Test invalid case
		payload = &core.ParsePayload{
			Value:  "not-a-cuid",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 1 {
			t.Errorf("Expected 1 issue for invalid CUID, got %d", len(payload.Issues))
		}
	})

	t.Run("ULID validates ULID format", func(t *testing.T) {
		check := ULID()

		// Test valid case
		payload := &core.ParsePayload{
			Value:  "01F4A2E4DK3R9H7A3T6M8N0Q5S",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 0 {
			t.Errorf("Expected no issues for valid ULID, got %d", len(payload.Issues))
		}

		// Test invalid case
		payload = &core.ParsePayload{
			Value:  "not-a-ulid",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 1 {
			t.Errorf("Expected 1 issue for invalid ULID, got %d", len(payload.Issues))
		}
	})

	t.Run("XID validates XID format", func(t *testing.T) {
		check := XID()

		// Test valid case
		payload := &core.ParsePayload{
			Value:  "9m4e2mr0ui3e8a215n4g",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 0 {
			t.Errorf("Expected no issues for valid XID, got %d", len(payload.Issues))
		}

		// Test invalid case
		payload = &core.ParsePayload{
			Value:  "not-a-xid",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 1 {
			t.Errorf("Expected 1 issue for invalid XID, got %d", len(payload.Issues))
		}
	})

	t.Run("KSUID validates KSUID format", func(t *testing.T) {
		check := KSUID()

		// Test valid case
		payload := &core.ParsePayload{
			Value:  "0ujsszwN8NRY24YaXiTIE2VWDTS",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 0 {
			t.Errorf("Expected no issues for valid KSUID, got %d", len(payload.Issues))
		}

		// Test invalid case
		payload = &core.ParsePayload{
			Value:  "not-a-ksuid",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 1 {
			t.Errorf("Expected 1 issue for invalid KSUID, got %d", len(payload.Issues))
		}
	})
}

func TestFormatCustomMessages(t *testing.T) {
	t.Run("Custom error messages work for format validation", func(t *testing.T) {
		check := Email("Must be a valid email address")

		payload := &core.ParsePayload{
			Value:  "invalid-email",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)

		if len(payload.Issues) != 1 {
			t.Fatalf("Expected 1 issue, got %d", len(payload.Issues))
		}

		// Check custom error mapping is set
		internals := check.GetZod()
		if internals.Def.Error == nil {
			t.Error("Expected custom error mapping to be set")
		}
	})
}
