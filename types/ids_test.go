package types

import (
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// CUID Tests
// =============================================================================

func TestCUID(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		schema := Cuid()
		validCUIDs := []string{"ck7q2g3ak0001psr9pbfgx83z", "clegxv4h2000008ld9bs2a2vr"}
		for _, cuid := range validCUIDs {
			result, err := schema.Parse(cuid)
			require.NoError(t, err)
			assert.Equal(t, cuid, result)
			assert.IsType(t, "", result)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		schema := Cuid()
		invalidInputs := []any{"not-a-cuid", 123, true, "", "ck7q2g3ak0001psr9pbfgx83z-"}
		for _, input := range invalidInputs {
			_, err := schema.Parse(input)
			assert.Error(t, err, "should reject: %v", input)
		}
	})

	t.Run("pointer", func(t *testing.T) {
		schema := CuidPtr()
		cuidStr := "ck7q2g3ak0001psr9pbfgx83z"
		result, err := schema.Parse(cuidStr)
		require.NoError(t, err)
		assert.IsType(t, (*string)(nil), result)
		require.NotNil(t, result)
		assert.Equal(t, cuidStr, *result)
	})

	t.Run("optional", func(t *testing.T) {
		schema := Cuid().Optional()
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("default", func(t *testing.T) {
		defaultValue := "ck7q2g3ak0001psr9pbfgx83z"
		schema := Cuid().Default(defaultValue)
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, defaultValue, result)
	})

	t.Run("chaining with min length", func(t *testing.T) {
		// A CUID is 25 chars long
		schema := Cuid().Min(25)
		result, err := schema.Parse("ck7q2g3ak0001psr9pbfgx83z")
		require.NoError(t, err)
		assert.Equal(t, "ck7q2g3ak0001psr9pbfgx83z", result)

		_, err = Cuid().Min(26).Parse("ck7q2g3ak0001psr9pbfgx83z")
		assert.Error(t, err)
	})

	t.Run("custom error", func(t *testing.T) {
		customError := "Invalid CUID format"
		schema := Cuid(core.SchemaParams{Error: customError})
		_, err := schema.Parse("invalid")
		assert.Error(t, err)
	})
}

// =============================================================================
// CUID2 Tests
// =============================================================================

func TestCUID2(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		schema := Cuid2()
		validCUIDs := []string{"ahkiaa2j7k63ufbpq688f3id", "b3his4na8s72mb275evbr83d"}
		for _, cuid := range validCUIDs {
			result, err := schema.Parse(cuid)
			require.NoError(t, err)
			assert.Equal(t, cuid, result)
			assert.IsType(t, "", result)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		schema := Cuid2()
		invalidInputs := []any{"not-a-cuid2", "CUID2-WITH-UPPER", 123, true, ""}
		for _, input := range invalidInputs {
			_, err := schema.Parse(input)
			assert.Error(t, err, "should reject: %v", input)
		}
	})

	t.Run("pointer", func(t *testing.T) {
		schema := Cuid2Ptr()
		cuid2Str := "ahkiaa2j7k63ufbpq688f3id"
		result, err := schema.Parse(cuid2Str)
		require.NoError(t, err)
		assert.IsType(t, (*string)(nil), result)
		require.NotNil(t, result)
		assert.Equal(t, cuid2Str, *result)
	})
}

// =============================================================================
// ULID Tests
// =============================================================================

func TestULID(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		schema := Ulid()
		validULIDs := []string{"01H8XGJWBWBAQ2XF7JGRB4B4B4", "01H8XGJWBYJ5M8M4J0Z5J3B3B3"}
		for _, ulid := range validULIDs {
			result, err := schema.Parse(ulid)
			require.NoError(t, err)
			assert.Equal(t, ulid, result)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		schema := Ulid()
		invalidInputs := []any{"01H8XGJWBYJ5M8M4J0Z5J3B3B3I", "not-a-ulid", 123, true, ""}
		for _, input := range invalidInputs {
			_, err := schema.Parse(input)
			assert.Error(t, err, "should reject: %v", input)
		}
	})
}

// =============================================================================
// XID Tests
// =============================================================================

func TestXID(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		schema := Xid()
		validXIDs := []string{"chc2vbt2mcc0000abs80", "chc2vbt2mcc0000abs8g", "CHC2VBT2MCC0000ABS8G"}
		for _, xid := range validXIDs {
			result, err := schema.Parse(xid)
			require.NoError(t, err)
			assert.Equal(t, xid, result)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		schema := Xid()
		invalidInputs := []any{"not-an-xid", 123, true, ""}
		for _, input := range invalidInputs {
			_, err := schema.Parse(input)
			assert.Error(t, err, "should reject: %v", input)
		}
	})
}

// =============================================================================
// KSUID Tests
// =============================================================================

func TestKSUID(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		schema := Ksuid()
		validKSUIDs := []string{"24Stb6wH3k7W5PpkBN24q4jCnsa", "24Stb6wH3k7W5PpkBN24q4jCnsb"}
		for _, ksuid := range validKSUIDs {
			result, err := schema.Parse(ksuid)
			require.NoError(t, err)
			assert.Equal(t, ksuid, result)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		schema := Ksuid()
		invalidInputs := []any{"not-a-ksuid", "24Stb6wH3k7W5PpkBN24q4jCnsa-", 123, true, ""}
		for _, input := range invalidInputs {
			_, err := schema.Parse(input)
			assert.Error(t, err, "should reject: %v", input)
		}
	})
}

// =============================================================================
// NanoID Tests
// =============================================================================

func TestNanoID(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		schema := Nanoid()
		validNanoIDs := []string{"J_ATaM-qZ8-Qk3Y-bY5c2", "c_U1bA-qZ8-Qk3Y-bY5c2"}
		for _, nanoid := range validNanoIDs {
			result, err := schema.Parse(nanoid)
			require.NoError(t, err)
			assert.Equal(t, nanoid, result)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		schema := Nanoid()
		invalidInputs := []any{"not-a-nanoid", "J_ATaM-qZ8-Qk3Y-bY5c2!", 123, true, ""}
		for _, input := range invalidInputs {
			_, err := schema.Parse(input)
			assert.Error(t, err, "should reject: %v", input)
		}
	})
}

// =============================================================================
// UUID Tests
// =============================================================================

func TestUUID(t *testing.T) {
	t.Run("valid generic", func(t *testing.T) {
		schema := Uuid()
		validUUIDs := []string{
			"123e4567-e89b-12d3-a456-426655440000", // v1
			"6ba7b810-9dad-11d1-80b4-00c04fd430c8",
			"d9428888-122b-469b-84de-d6e0a77858b7", // v4
			"1ee9b3b0-0b4a-6290-9b49-0242ac120002", // v6
			"01890de4-7f13-7d5a-a439-c5f10a8a71d1", // v7
		}
		for _, uuid := range validUUIDs {
			result, err := schema.Parse(uuid)
			require.NoError(t, err)
			assert.Equal(t, uuid, result)
		}
	})

	t.Run("invalid generic", func(t *testing.T) {
		schema := Uuid()
		invalidInputs := []any{"not-a-uuid", 123, true, ""}
		for _, input := range invalidInputs {
			_, err := schema.Parse(input)
			assert.Error(t, err, "should reject: %v", input)
		}
	})

	t.Run("version v4", func(t *testing.T) {
		valid := "d9428888-122b-469b-84de-d6e0a77858b7"
		invalid := "d9428888-122b-169b-84de-d6e0a77858b7" // Invalid version

		// Test with versioned constructor
		schema1 := Uuid("v4")
		_, err := schema1.Parse(valid)
		require.NoError(t, err)
		_, err = schema1.Parse(invalid)
		require.Error(t, err)

		// Test with convenience function
		schema2 := Uuidv4()
		_, err = schema2.Parse(valid)
		require.NoError(t, err)
		_, err = schema2.Parse(invalid)
		require.Error(t, err)
	})

	t.Run("version v6", func(t *testing.T) {
		valid := "1ee9b3b0-0b4a-6290-9b49-0242ac120002"
		invalid := "1ee9b3b0-0b4a-1290-9b49-0242ac120002" // Invalid version

		schema1 := Uuid("v6")
		_, err := schema1.Parse(valid)
		require.NoError(t, err)
		_, err = schema1.Parse(invalid)
		require.Error(t, err)

		schema2 := Uuidv6()
		_, err = schema2.Parse(valid)
		require.NoError(t, err)
		_, err = schema2.Parse(invalid)
		require.Error(t, err)
	})

	t.Run("version v7", func(t *testing.T) {
		valid := "01890de4-7f13-7d5a-a439-c5f10a8a71d1"
		invalid := "01890de4-7f13-4d5a-a439-c5f10a8a71d1" // Invalid version

		schema1 := Uuid("v7")
		_, err := schema1.Parse(valid)
		require.NoError(t, err)
		_, err = schema1.Parse(invalid)
		require.Error(t, err)

		schema2 := Uuidv7()
		_, err = schema2.Parse(valid)
		require.NoError(t, err)
		_, err = schema2.Parse(invalid)
		require.Error(t, err)
	})

	t.Run("optional with version", func(t *testing.T) {
		schema := Uuidv4().Optional()
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("default with generic", func(t *testing.T) {
		defaultValue := "123e4567-e89b-12d3-a456-426655440000"
		schema := Uuid().Default(defaultValue)
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, defaultValue, result)
	})
}
