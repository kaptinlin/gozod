package structx

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// TEST TYPES
// =============================================================================

type testUser struct {
	Name    string `json:"name"`
	Age     int    `json:"age"`
	Email   string `json:"email,omitempty"`
	private string //nolint:unused // unexported field for testing
}

type testConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type testNoTags struct {
	Name string
	Age  int
}

type testSkipField struct {
	Name   string `json:"name"`
	Secret string `json:"-"`
}

// =============================================================================
// STRUCT CONVERSION TESTS
// =============================================================================

func TestToMap(t *testing.T) {
	t.Run("nil input returns error", func(t *testing.T) {
		_, err := ToMap(nil)
		assert.ErrorIs(t, err, ErrInvalidStructInput)
	})

	t.Run("non-struct returns error", func(t *testing.T) {
		_, err := ToMap("not a struct")
		assert.ErrorIs(t, err, ErrInvalidStructInput)
	})

	t.Run("struct converts correctly", func(t *testing.T) {
		user := testUser{Name: "Alice", Age: 30, Email: "alice@example.com"}
		got, err := ToMap(user)
		require.NoError(t, err)

		assert.Equal(t, "Alice", got["name"])
		assert.Equal(t, 30, got["age"])
		assert.Equal(t, "alice@example.com", got["email"])
	})

	t.Run("pointer to struct converts correctly", func(t *testing.T) {
		user := &testUser{Name: "Bob", Age: 25}
		got, err := ToMap(user)
		require.NoError(t, err)
		assert.Equal(t, "Bob", got["name"])
	})

	t.Run("nil pointer returns error", func(t *testing.T) {
		var user *testUser
		_, err := ToMap(user)
		assert.ErrorIs(t, err, ErrInvalidStructInput)
	})
}

func TestFromMap(t *testing.T) {
	t.Run("non-struct type returns error", func(t *testing.T) {
		data := map[string]any{"a": 1}
		_, err := FromMap(data, reflect.TypeOf("string"))
		assert.ErrorIs(t, err, ErrTargetTypeMustBeStruct)
	})

	t.Run("map converts to struct", func(t *testing.T) {
		data := map[string]any{"name": "Alice", "age": 30}
		result, err := FromMap(data, reflect.TypeOf(testUser{}))
		require.NoError(t, err)

		user, ok := result.(testUser)
		require.True(t, ok)
		assert.Equal(t, "Alice", user.Name)
		assert.Equal(t, 30, user.Age)
	})

	t.Run("pointer type converts correctly", func(t *testing.T) {
		data := map[string]any{"host": "localhost", "port": 8080}
		result, err := FromMap(data, reflect.TypeOf(&testConfig{}))
		require.NoError(t, err)

		config, ok := result.(testConfig)
		require.True(t, ok)
		assert.Equal(t, "localhost", config.Host)
	})
}

func TestMarshal(t *testing.T) {
	t.Run("nil input returns nil", func(t *testing.T) {
		assert.Nil(t, Marshal(nil))
	})

	t.Run("nil pointer returns nil", func(t *testing.T) {
		var user *testUser
		assert.Nil(t, Marshal(user))
	})

	t.Run("non-struct returns nil", func(t *testing.T) {
		assert.Nil(t, Marshal("string"))
	})

	t.Run("struct marshals correctly", func(t *testing.T) {
		user := testUser{Name: "Alice", Age: 30}
		got := Marshal(user)

		require.NotNil(t, got)
		assert.Equal(t, "Alice", got["name"])
		assert.Equal(t, 30, got["age"])
	})

	t.Run("struct without json tags uses field names", func(t *testing.T) {
		s := testNoTags{Name: "Bob", Age: 25}
		got := Marshal(s)

		assert.Equal(t, "Bob", got["Name"])
		assert.Equal(t, 25, got["Age"])
	})

	t.Run("json:- skips field", func(t *testing.T) {
		s := testSkipField{Name: "Alice", Secret: "password"}
		got := Marshal(s)

		assert.Equal(t, "Alice", got["name"])
		assert.NotContains(t, got, "Secret")
		assert.NotContains(t, got, "-")
	})

	t.Run("unexported fields are skipped", func(t *testing.T) {
		user := testUser{Name: "Alice", Age: 30}
		got := Marshal(user)
		assert.NotContains(t, got, "private")
	})
}

func TestUnmarshal(t *testing.T) {
	t.Run("non-struct type returns error", func(t *testing.T) {
		data := map[string]any{"a": 1}
		_, err := Unmarshal(data, reflect.TypeOf("string"))
		assert.ErrorIs(t, err, ErrTargetTypeMustBeStruct)
	})

	t.Run("correctly unmarshals data", func(t *testing.T) {
		data := map[string]any{"name": "Alice", "age": 30}
		result, err := Unmarshal(data, reflect.TypeOf(testUser{}))
		require.NoError(t, err)

		user, ok := result.(testUser)
		require.True(t, ok)
		assert.Equal(t, "Alice", user.Name)
	})

	t.Run("handles missing fields gracefully", func(t *testing.T) {
		data := map[string]any{"name": "Bob"}
		result, err := Unmarshal(data, reflect.TypeOf(testUser{}))
		require.NoError(t, err)

		user := result.(testUser)
		assert.Equal(t, "Bob", user.Name)
		assert.Equal(t, 0, user.Age) // zero value
	})

	t.Run("handles nil values gracefully", func(t *testing.T) {
		data := map[string]any{"name": nil, "age": 25}
		result, err := Unmarshal(data, reflect.TypeOf(testUser{}))
		require.NoError(t, err)

		user := result.(testUser)
		assert.Empty(t, user.Name)
		assert.Equal(t, 25, user.Age)
	})

	t.Run("handles type conversion", func(t *testing.T) {
		data := map[string]any{"name": "Alice", "age": int64(30)}
		result, err := Unmarshal(data, reflect.TypeOf(testUser{}))
		require.NoError(t, err)

		user := result.(testUser)
		assert.Equal(t, 30, user.Age)
	})
}

func TestGetFieldName(t *testing.T) {
	type testStruct struct {
		WithTag    string `json:"custom_name"`
		WithEmpty  string `json:",omitempty"`
		WithDash   string `json:"-"`
		WithoutTag string
	}

	st := reflect.TypeOf(testStruct{})

	tests := []struct {
		fieldName string
		want      string
	}{
		{"WithTag", "custom_name"},
		{"WithEmpty", "WithEmpty"},
		{"WithDash", "-"},
		{"WithoutTag", "WithoutTag"},
	}

	for _, tt := range tests {
		t.Run(tt.fieldName, func(t *testing.T) {
			field, _ := st.FieldByName(tt.fieldName)
			got := getFieldName(field)
			assert.Equal(t, tt.want, got)
		})
	}
}
