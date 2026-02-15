package types

import (
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGUID(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		s := GUID()
		valid := []string{
			"123e4567-e89b-12d3-a456-426655440000",
			"6ba7b810-9dad-11d1-80b4-00c04fd430c8",
			"d9428888-122b-469b-84de-d6e0a77858b7",
			"00000000-0000-0000-0000-000000000000",
			"FFFFFFFF-FFFF-FFFF-FFFF-FFFFFFFFFFFF",
			"aBcDeF01-1234-5678-90Ab-CdEf12345678",
		}
		for _, v := range valid {
			got, err := s.Parse(v)
			require.NoError(t, err, "Parse(%q)", v)
			assert.Equal(t, v, got)
			assert.IsType(t, "", got)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		s := GUID()
		invalid := []any{
			"not-a-guid", 123, true, "",
			"123e4567-e89b-12d3-a456",
			"123e4567-e89b-12d3-a456-42665544",
			"123e4567e89b12d3a456426655440000",
			"g23e4567-e89b-12d3-a456-426655440000",
		}
		for _, v := range invalid {
			_, err := s.Parse(v)
			assert.Error(t, err, "Parse(%v)", v)
		}
	})

	t.Run("pointer", func(t *testing.T) {
		want := "123e4567-e89b-12d3-a456-426655440000"
		got, err := GUIDPtr().Parse(want)
		require.NoError(t, err)
		assert.IsType(t, (*string)(nil), got)
		require.NotNil(t, got)
		assert.Equal(t, want, *got)
	})

	t.Run("optional", func(t *testing.T) {
		got, err := GUID().Optional().Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("default", func(t *testing.T) {
		want := "123e4567-e89b-12d3-a456-426655440000"
		got, err := GUID().Default(want).Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("StrictParse", func(t *testing.T) {
		s := GUID()
		want := "123e4567-e89b-12d3-a456-426655440000"
		got, err := s.StrictParse(want)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("MustStrictParse", func(t *testing.T) {
		s := GUID()
		want := "123e4567-e89b-12d3-a456-426655440000"
		got := s.MustStrictParse(want)
		assert.Equal(t, want, got)

		assert.Panics(t, func() {
			s.MustStrictParse("invalid")
		})
	})

	t.Run("custom error", func(t *testing.T) {
		s := GUID(core.SchemaParams{Error: "Invalid GUID format"})
		_, err := s.Parse("invalid")
		assert.Error(t, err)
	})

	t.Run("chaining with min length", func(t *testing.T) {
		got, err := GUID().Min(36).Parse("123e4567-e89b-12d3-a456-426655440000")
		require.NoError(t, err)
		assert.Equal(t, "123e4567-e89b-12d3-a456-426655440000", got)

		_, err = GUID().Min(37).Parse("123e4567-e89b-12d3-a456-426655440000")
		assert.Error(t, err)
	})
}

func TestCUID(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		s := CUID()
		valid := []string{"ck7q2g3ak0001psr9pbfgx83z", "clegxv4h2000008ld9bs2a2vr"}
		for _, v := range valid {
			got, err := s.Parse(v)
			require.NoError(t, err)
			assert.Equal(t, v, got)
			assert.IsType(t, "", got)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		s := CUID()
		invalid := []any{"not-a-cuid", 123, true, "", "ck7q2g3ak0001psr9pbfgx83z-"}
		for _, v := range invalid {
			_, err := s.Parse(v)
			assert.Error(t, err, "Parse(%v)", v)
		}
	})

	t.Run("pointer", func(t *testing.T) {
		want := "ck7q2g3ak0001psr9pbfgx83z"
		got, err := CUIDPtr().Parse(want)
		require.NoError(t, err)
		assert.IsType(t, (*string)(nil), got)
		require.NotNil(t, got)
		assert.Equal(t, want, *got)
	})

	t.Run("optional", func(t *testing.T) {
		got, err := CUID().Optional().Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("default", func(t *testing.T) {
		want := "ck7q2g3ak0001psr9pbfgx83z"
		got, err := CUID().Default(want).Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("chaining with min length", func(t *testing.T) {
		got, err := CUID().Min(25).Parse("ck7q2g3ak0001psr9pbfgx83z")
		require.NoError(t, err)
		assert.Equal(t, "ck7q2g3ak0001psr9pbfgx83z", got)

		_, err = CUID().Min(26).Parse("ck7q2g3ak0001psr9pbfgx83z")
		assert.Error(t, err)
	})

	t.Run("custom error", func(t *testing.T) {
		s := CUID(core.SchemaParams{Error: "Invalid CUID format"})
		_, err := s.Parse("invalid")
		assert.Error(t, err)
	})
}

func TestCUID2(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		s := CUID2()
		valid := []string{"ahkiaa2j7k63ufbpq688f3id", "b3his4na8s72mb275evbr83d"}
		for _, v := range valid {
			got, err := s.Parse(v)
			require.NoError(t, err)
			assert.Equal(t, v, got)
			assert.IsType(t, "", got)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		s := CUID2()
		invalid := []any{"not-a-cuid2", "CUID2-WITH-UPPER", 123, true, ""}
		for _, v := range invalid {
			_, err := s.Parse(v)
			assert.Error(t, err, "Parse(%v)", v)
		}
	})

	t.Run("pointer", func(t *testing.T) {
		want := "ahkiaa2j7k63ufbpq688f3id"
		got, err := CUID2Ptr().Parse(want)
		require.NoError(t, err)
		assert.IsType(t, (*string)(nil), got)
		require.NotNil(t, got)
		assert.Equal(t, want, *got)
	})
}

func TestULID(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		s := ULID()
		valid := []string{"01H8XGJWBWBAQ2XF7JGRB4B4B4", "01H8XGJWBYJ5M8M4J0Z5J3B3B3"}
		for _, v := range valid {
			got, err := s.Parse(v)
			require.NoError(t, err)
			assert.Equal(t, v, got)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		s := ULID()
		invalid := []any{"01H8XGJWBYJ5M8M4J0Z5J3B3B3I", "not-a-ulid", 123, true, ""}
		for _, v := range invalid {
			_, err := s.Parse(v)
			assert.Error(t, err, "Parse(%v)", v)
		}
	})
}

func TestXID(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		s := XID()
		valid := []string{"chc2vbt2mcc0000abs80", "chc2vbt2mcc0000abs8g", "CHC2VBT2MCC0000ABS8G"}
		for _, v := range valid {
			got, err := s.Parse(v)
			require.NoError(t, err)
			assert.Equal(t, v, got)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		s := XID()
		invalid := []any{"not-an-xid", 123, true, ""}
		for _, v := range invalid {
			_, err := s.Parse(v)
			assert.Error(t, err, "Parse(%v)", v)
		}
	})
}

func TestKSUID(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		s := KSUID()
		valid := []string{"24Stb6wH3k7W5PpkBN24q4jCnsa", "24Stb6wH3k7W5PpkBN24q4jCnsb"}
		for _, v := range valid {
			got, err := s.Parse(v)
			require.NoError(t, err)
			assert.Equal(t, v, got)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		s := KSUID()
		invalid := []any{"not-a-ksuid", "24Stb6wH3k7W5PpkBN24q4jCnsa-", 123, true, ""}
		for _, v := range invalid {
			_, err := s.Parse(v)
			assert.Error(t, err, "Parse(%v)", v)
		}
	})
}

func TestNanoID(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		s := NanoID()
		valid := []string{"J_ATaM-qZ8-Qk3Y-bY5c2", "c_U1bA-qZ8-Qk3Y-bY5c2"}
		for _, v := range valid {
			got, err := s.Parse(v)
			require.NoError(t, err)
			assert.Equal(t, v, got)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		s := NanoID()
		invalid := []any{"not-a-nanoid", "J_ATaM-qZ8-Qk3Y-bY5c2!", 123, true, ""}
		for _, v := range invalid {
			_, err := s.Parse(v)
			assert.Error(t, err, "Parse(%v)", v)
		}
	})
}

func TestUUID(t *testing.T) {
	t.Run("valid generic", func(t *testing.T) {
		s := UUID()
		valid := []string{
			"123e4567-e89b-12d3-a456-426655440000",
			"6ba7b810-9dad-11d1-80b4-00c04fd430c8",
			"d9428888-122b-469b-84de-d6e0a77858b7",
			"1ee9b3b0-0b4a-6290-9b49-0242ac120002",
			"01890de4-7f13-7d5a-a439-c5f10a8a71d1",
		}
		for _, v := range valid {
			got, err := s.Parse(v)
			require.NoError(t, err)
			assert.Equal(t, v, got)
		}
	})

	t.Run("invalid generic", func(t *testing.T) {
		s := UUID()
		invalid := []any{"not-a-uuid", 123, true, ""}
		for _, v := range invalid {
			_, err := s.Parse(v)
			assert.Error(t, err, "Parse(%v)", v)
		}
	})

	t.Run("version v4", func(t *testing.T) {
		valid := "d9428888-122b-469b-84de-d6e0a77858b7"
		invalid := "d9428888-122b-169b-84de-d6e0a77858b7"

		s1 := UUID("v4")
		_, err := s1.Parse(valid)
		require.NoError(t, err)
		_, err = s1.Parse(invalid)
		require.Error(t, err)

		s2 := UUIDv4()
		_, err = s2.Parse(valid)
		require.NoError(t, err)
		_, err = s2.Parse(invalid)
		require.Error(t, err)
	})

	t.Run("version v6", func(t *testing.T) {
		valid := "1ee9b3b0-0b4a-6290-9b49-0242ac120002"
		invalid := "1ee9b3b0-0b4a-1290-9b49-0242ac120002"

		s1 := UUID("v6")
		_, err := s1.Parse(valid)
		require.NoError(t, err)
		_, err = s1.Parse(invalid)
		require.Error(t, err)

		s2 := UUIDv6()
		_, err = s2.Parse(valid)
		require.NoError(t, err)
		_, err = s2.Parse(invalid)
		require.Error(t, err)
	})

	t.Run("version v7", func(t *testing.T) {
		valid := "01890de4-7f13-7d5a-a439-c5f10a8a71d1"
		invalid := "01890de4-7f13-4d5a-a439-c5f10a8a71d1"

		s1 := UUID("v7")
		_, err := s1.Parse(valid)
		require.NoError(t, err)
		_, err = s1.Parse(invalid)
		require.Error(t, err)

		s2 := UUIDv7()
		_, err = s2.Parse(valid)
		require.NoError(t, err)
		_, err = s2.Parse(invalid)
		require.Error(t, err)
	})

	t.Run("optional with version", func(t *testing.T) {
		got, err := UUIDv4().Optional().Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("default with generic", func(t *testing.T) {
		want := "123e4567-e89b-12d3-a456-426655440000"
		got, err := UUID().Default(want).Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})
}

func TestIDs_DefaultAndPrefault(t *testing.T) {
	t.Run("default has higher priority than prefault", func(t *testing.T) {
		got1, err := CUID().Default("ck7q2g3ak0001psr9pbfgx83z").Prefault("clegxv4h2000008ld9bs2a2vr").Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "ck7q2g3ak0001psr9pbfgx83z", got1)

		got2, err := UUID().Default("123e4567-e89b-12d3-a456-426655440000").Prefault("d9428888-122b-469b-84de-d6e0a77858b7").Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "123e4567-e89b-12d3-a456-426655440000", got2)

		got3, err := CUIDPtr().Default("ck7q2g3ak0001psr9pbfgx83z").Prefault("clegxv4h2000008ld9bs2a2vr").Parse(nil)
		require.NoError(t, err)
		require.NotNil(t, got3)
		assert.Equal(t, "ck7q2g3ak0001psr9pbfgx83z", *got3)
	})

	t.Run("default short-circuits validation", func(t *testing.T) {
		got1, err := CUID().Min(30).Default("short").Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "short", got1)

		got2, err := UUID().Refine(func(s string) bool {
			return false
		}, "Should never pass").Default("invalid-uuid-format").Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "invalid-uuid-format", got2)

		got3, err := ULID().Default("not-a-ulid").Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "not-a-ulid", got3)
	})

	t.Run("prefault goes through full validation", func(t *testing.T) {
		got1, err := CUID().Min(25).Prefault("ck7q2g3ak0001psr9pbfgx83z").Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "ck7q2g3ak0001psr9pbfgx83z", got1)

		got2, err := UUIDv4().Prefault("d9428888-122b-469b-84de-d6e0a77858b7").Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "d9428888-122b-469b-84de-d6e0a77858b7", got2)

		got3, err := ULID().Prefault("01H8XGJWBWBAQ2XF7JGRB4B4B4").Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "01H8XGJWBWBAQ2XF7JGRB4B4B4", got3)
	})

	t.Run("prefault only triggered by nil input", func(t *testing.T) {
		s := CUID().Prefault("ck7q2g3ak0001psr9pbfgx83z")

		got, err := s.Parse("clegxv4h2000008ld9bs2a2vr")
		require.NoError(t, err)
		assert.Equal(t, "clegxv4h2000008ld9bs2a2vr", got)

		_, err = s.Parse("invalid-cuid")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid cuid")
	})

	t.Run("DefaultFunc and PrefaultFunc", func(t *testing.T) {
		defaultCalled := false
		prefaultCalled := false

		dfn := func() string {
			defaultCalled = true
			return "ck7q2g3ak0001psr9pbfgx83z"
		}
		pfn := func() string {
			prefaultCalled = true
			return "clegxv4h2000008ld9bs2a2vr"
		}

		got1, err := CUID().DefaultFunc(dfn).PrefaultFunc(pfn).Parse(nil)
		require.NoError(t, err)
		assert.True(t, defaultCalled)
		assert.False(t, prefaultCalled)
		assert.Equal(t, "ck7q2g3ak0001psr9pbfgx83z", got1)

		defaultCalled = false
		prefaultCalled = false

		got2, err := CUID().PrefaultFunc(pfn).Parse(nil)
		require.NoError(t, err)
		assert.True(t, prefaultCalled)
		assert.Equal(t, "clegxv4h2000008ld9bs2a2vr", got2)
	})

	t.Run("prefault validation failure", func(t *testing.T) {
		_, err := CUID().Prefault("invalid-cuid").Parse(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid cuid")

		_, err = UUIDv4().Prefault("01890de4-7f13-7d5a-a439-c5f10a8a71d1").Parse(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid uuid")

		_, err = ULID().Prefault("invalid-ulid").Parse(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid ulid")
	})

	t.Run("all ID types with default and prefault", func(t *testing.T) {
		got1, err := CUID2().Default("ahkiaa2j7k63ufbpq688f3id").Prefault("b3his4na8s72mb275evbr83d").Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "ahkiaa2j7k63ufbpq688f3id", got1)

		got2, err := XID().Default("chc2vbt2mcc0000abs80").Prefault("chc2vbt2mcc0000abs8g").Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "chc2vbt2mcc0000abs80", got2)

		got3, err := KSUID().Default("24Stb6wH3k7W5PpkBN24q4jCnsa").Prefault("24Stb6wH3k7W5PpkBN24q4jCnsb").Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "24Stb6wH3k7W5PpkBN24q4jCnsa", got3)

		got4, err := NanoID().Default("J_ATaM-qZ8-Qk3Y-bY5c2").Prefault("c_U1bA-qZ8-Qk3Y-bY5c2").Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "J_ATaM-qZ8-Qk3Y-bY5c2", got4)
	})
}
