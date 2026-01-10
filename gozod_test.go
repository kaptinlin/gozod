package gozod

import (
	"sync"
	"testing"
	"time"

	"github.com/kaptinlin/gozod/locales"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// BASIC TYPE EXPORTS TESTS - Verify all exported types exist
// =============================================================================

func TestPrimitiveTypeExports(t *testing.T) {
	t.Run("string types", func(t *testing.T) {
		_ = String()
		_ = StringPtr()

		result, err := String().Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)
	})

	t.Run("boolean types", func(t *testing.T) {
		_ = Bool()
		_ = BoolPtr()

		result, err := Bool().Parse(true)
		require.NoError(t, err)
		assert.Equal(t, true, result)
	})

	t.Run("integer types", func(t *testing.T) {
		_ = Int()
		_ = IntPtr()
		_ = Int8()
		_ = Int8Ptr()
		_ = Int16()
		_ = Int16Ptr()
		_ = Int32()
		_ = Int32Ptr()
		_ = Int64()
		_ = Int64Ptr()
		_ = Uint()
		_ = UintPtr()
		_ = Uint8()
		_ = Uint8Ptr()
		_ = Uint16()
		_ = Uint16Ptr()
		_ = Uint32()
		_ = Uint32Ptr()
		_ = Uint64()
		_ = Uint64Ptr()

		result, err := Int().Parse(42)
		require.NoError(t, err)
		assert.Equal(t, 42, result)
	})

	t.Run("float types", func(t *testing.T) {
		_ = Float()
		_ = FloatPtr()
		_ = Float32()
		_ = Float32Ptr()
		_ = Float64()
		_ = Float64Ptr()
		_ = Number()
		_ = NumberPtr()

		result, err := Float64().Parse(3.14)
		require.NoError(t, err)
		assert.InDelta(t, 3.14, result, 0.01)
	})

	t.Run("complex types", func(t *testing.T) {
		_ = Complex()
		_ = ComplexPtr()
		_ = Complex64()
		_ = Complex64Ptr()
		_ = Complex128()
		_ = Complex128Ptr()

		result, err := Complex().Parse(complex(3.0, 4.0))
		require.NoError(t, err)
		assert.Equal(t, complex(3.0, 4.0), result)
	})

	t.Run("bigint types", func(t *testing.T) {
		_ = BigInt()
		_ = BigIntPtr()
	})

	t.Run("time types", func(t *testing.T) {
		_ = Time()
		_ = TimePtr()

		now := time.Now()
		result, err := Time().Parse(now)
		require.NoError(t, err)
		assert.Equal(t, now, result)
	})
}

// =============================================================================
// STRING FORMAT TYPE EXPORTS TESTS
// =============================================================================

func TestStringFormatTypeExports(t *testing.T) {
	t.Run("basic string formats", func(t *testing.T) {
		_ = Email()
		_ = EmailPtr()
		_ = Emoji()
		_ = EmojiPtr()
		_ = Base64()
		_ = Base64Ptr()
		_ = Base64URL()
		_ = Base64URLPtr()

		result, err := Email().Parse("test@example.com")
		require.NoError(t, err)
		assert.Equal(t, "test@example.com", result)
	})

	t.Run("network formats", func(t *testing.T) {
		_ = IPv4()
		_ = IPv4Ptr()
		_ = IPv6()
		_ = IPv6Ptr()
		_ = CIDRv4()
		_ = CIDRv4Ptr()
		_ = CIDRv6()
		_ = CIDRv6Ptr()
		_ = URL()
		_ = URLPtr()

		result, err := URL().Parse("https://example.com")
		require.NoError(t, err)
		assert.Equal(t, "https://example.com", result)
	})

	t.Run("iso formats", func(t *testing.T) {
		_ = Iso()
		_ = IsoPtr()
		_ = IsoDateTime()
		_ = IsoDateTimePtr()
		_ = IsoDate()
		_ = IsoDatePtr()
		_ = IsoTime()
		_ = IsoTimePtr()
		_ = IsoDuration()
		_ = IsoDurationPtr()

		// Test ISO precision constants
		_ = PrecisionMinute
		_ = PrecisionSecond
		_ = PrecisionDecisecond
		_ = PrecisionCentisecond
		_ = PrecisionMillisecond
		_ = PrecisionMicrosecond
		_ = PrecisionNanosecond
	})

	t.Run("unique identifier formats", func(t *testing.T) {
		_ = Cuid()
		_ = CuidPtr()
		_ = Cuid2()
		_ = Cuid2Ptr()
		_ = Ulid()
		_ = UlidPtr()
		_ = Xid()
		_ = XidPtr()
		_ = Ksuid()
		_ = KsuidPtr()
		_ = Nanoid()
		_ = NanoidPtr()
		_ = Uuid()
		_ = UuidPtr()
		_ = Uuidv4()
		_ = Uuidv4Ptr()
		_ = Uuidv6()
		_ = Uuidv6Ptr()
		_ = Uuidv7()
		_ = Uuidv7Ptr()
		_ = JWT()
		_ = JWTPtr()
	})
}

// =============================================================================
// COLLECTION TYPE EXPORTS TESTS
// =============================================================================

func TestCollectionTypeExports(t *testing.T) {
	t.Run("array and slice types", func(t *testing.T) {
		_ = Array()
		_ = ArrayPtr()
		_ = Slice[int](Int())
		_ = SlicePtr[int](Int())

		result, err := Slice[int](Int()).Parse([]int{1, 2, 3})
		require.NoError(t, err)
		assert.Equal(t, []int{1, 2, 3}, result)
	})

	t.Run("map and record types", func(t *testing.T) {
		_ = Map(String(), Int())
		_ = MapPtr(String(), Int())
		_ = Record[string](String(), Int())
		_ = RecordPtr[string](String(), Int())

		testMap := map[string]int{"key": 42}
		result, err := Record[string](String(), Int()).Parse(testMap)
		require.NoError(t, err)
		assert.Equal(t, testMap, result)
	})

	t.Run("object types", func(t *testing.T) {
		testSchema := ObjectSchema{"key": String()}
		_ = Object(testSchema)
		_ = ObjectPtr(testSchema)
		_ = StrictObject(testSchema)
		_ = StrictObjectPtr(testSchema)
		_ = LooseObject(testSchema)
		_ = LooseObjectPtr(testSchema)
	})

	t.Run("struct types", func(t *testing.T) {
		type TestStruct struct {
			Name string
			Age  int
		}

		_ = Struct[TestStruct]()
		_ = StructPtr[TestStruct]()

		testData := TestStruct{Name: "Alice", Age: 30}
		result, err := Struct[TestStruct]().Parse(testData)
		require.NoError(t, err)
		assert.Equal(t, testData, result)
	})
}

// =============================================================================
// COMPOSITE TYPE EXPORTS TESTS
// =============================================================================

func TestCompositeTypeExports(t *testing.T) {
	t.Run("union and intersection types", func(t *testing.T) {
		_ = Union([]any{String(), Int()})
		_ = UnionPtr([]any{String(), Int()})
		_ = Intersection(String(), Int())
		_ = IntersectionPtr(String(), Int())
		_ = DiscriminatedUnion("type", []any{String(), Int()})
		_ = DiscriminatedUnionPtr("type", []any{String(), Int()})
	})
}

// =============================================================================
// SPECIAL TYPE EXPORTS TESTS
// =============================================================================

func TestSpecialTypeExports(t *testing.T) {
	t.Run("any and unknown types", func(t *testing.T) {
		_ = Any()
		_ = AnyPtr()
		_ = Unknown()
		_ = UnknownPtr()

		result, err := Any().Parse("anything")
		require.NoError(t, err)
		assert.Equal(t, "anything", result)
	})

	t.Run("never and nil types", func(t *testing.T) {
		_ = Never()
		_ = NeverPtr()
		_ = Nil()
		_ = NilPtr()
	})

	t.Run("file and function types", func(t *testing.T) {
		_ = File()
		_ = FilePtr()
		_ = Function()
		_ = FunctionPtr()
	})

	t.Run("string bool types", func(t *testing.T) {
		_ = StringBool()
		_ = StringBoolPtr()

		result, err := StringBool().Parse("true")
		require.NoError(t, err)
		assert.Equal(t, true, result)
	})

	t.Run("literal and enum types", func(t *testing.T) {
		_ = Literal("hello")
		_ = LiteralPtr("hello")
		_ = LiteralOf([]string{"a", "b", "c"})
		_ = LiteralPtrOf([]string{"a", "b", "c"})

		_ = Enum("red", "green", "blue")
		_ = EnumPtr("red", "green", "blue")
		_ = EnumSlice([]string{"red", "green", "blue"})
		_ = EnumSlicePtr([]string{"red", "green", "blue"})
		_ = EnumMap(map[string]int{"low": 1, "high": 2})
		_ = EnumMapPtr(map[string]int{"low": 1, "high": 2})

		result, err := Enum("red", "green", "blue").Parse("red")
		require.NoError(t, err)
		assert.Equal(t, "red", result)
	})

	t.Run("lazy types", func(t *testing.T) {
		_ = LazyAny(func() any { return String() })
		_ = LazyPtr(func() any { return String() })
		_ = Lazy(func() *ZodString[string] { return String() })
	})
}

// =============================================================================
// TYPE ALIAS EXPORTS TESTS
// =============================================================================

func TestTypeAliasExports(t *testing.T) {
	t.Run("type aliases exist", func(t *testing.T) {
		// Test that all type aliases are exported and can be used
		// We don't test exact type compatibility, just that they exist

		// Primitive type aliases
		_ = (*ZodString[string])(nil)
		_ = (*ZodBool[bool])(nil)
		_ = (*ZodInteger[int64, int64])(nil)
		_ = (*ZodFloat[float64, float64])(nil)
		_ = (*ZodComplex[complex128])(nil)
		_ = (*ZodTime[time.Time])(nil)

		// String format type aliases
		_ = (*ZodEmail[string])(nil)
		_ = (*ZodEmoji[string])(nil)
		_ = (*ZodBase64[string])(nil)
		_ = (*ZodBase64URL[string])(nil)
		_ = (*ZodIPv4[string])(nil)
		_ = (*ZodIPv6[string])(nil)
		_ = (*ZodCIDRv4[string])(nil)
		_ = (*ZodCIDRv6[string])(nil)
		_ = (*ZodURL[string])(nil)
		_ = (*ZodIso[string])(nil)
		_ = (*ZodCUID[string])(nil)
		_ = (*ZodCUID2[string])(nil)
		_ = (*ZodULID[string])(nil)
		_ = (*ZodXID[string])(nil)
		_ = (*ZodKSUID[string])(nil)
		_ = (*ZodNanoID[string])(nil)
		_ = (*ZodUUID[string])(nil)
		_ = (*ZodJWT[string])(nil)

		// Collection type aliases
		type TestStruct struct{ Value string }
		_ = (*ZodArray[any, []any])(nil)
		_ = (*ZodSlice[int, []int])(nil)
		_ = (*ZodMap[any, map[any]any])(nil)
		_ = (*ZodRecord[any, map[string]any])(nil)
		_ = (*ZodObject[any, any])(nil)
		_ = (*ZodStruct[TestStruct, TestStruct])(nil)

		// Composite type aliases
		_ = (*ZodUnion[any, any])(nil)
		_ = (*ZodIntersection[any, any])(nil)
		_ = (*ZodDiscriminatedUnion[any, any])(nil)

		// Special type aliases
		_ = (*ZodAny[any, any])(nil)
		_ = (*ZodUnknown[any, any])(nil)
		_ = (*ZodNever[any, any])(nil)
		_ = (*ZodNil[any])(nil)
		_ = (*ZodFile[any, any])(nil)
		_ = (*ZodFunction[func()])(nil)
		_ = (*ZodStringBool[bool])(nil)
		_ = (*ZodLazy[func()])(nil)
		_ = (*ZodEnum[string, string])(nil)
		_ = (*ZodLiteral[string, string])(nil)
	})
}

// =============================================================================
// OPTIONS AND CONSTANTS TESTS
// =============================================================================

func TestOptionsAndConstants(t *testing.T) {
	t.Run("options types", func(t *testing.T) {
		var _ URLOptions
		var _ JWTOptions
		var _ IsoDatetimeOptions
		var _ IsoTimeOptions
	})

	t.Run("issue code constants", func(t *testing.T) {
		_ = IssueInvalidType
		_ = IssueInvalidValue
		_ = IssueInvalidFormat
		_ = IssueInvalidUnion
		_ = IssueInvalidKey
		_ = IssueInvalidElement
		_ = IssueTooBig
		_ = IssueTooSmall
		_ = IssueNotMultipleOf
		_ = IssueUnrecognizedKeys
		_ = IssueCustom
	})
}

// =============================================================================
// ERROR HANDLING TESTS
// =============================================================================

func TestErrorHandling(t *testing.T) {
	t.Run("error types", func(t *testing.T) {
		// Test that error types are exported
		var _ ZodError
		var _ ZodIssue
		var _ ZodRawIssue

		// Test error utility function
		_ = IsZodError
	})

	t.Run("config functions", func(t *testing.T) {
		// Test that config functions are exported
		_ = Config
		_ = GetConfig
	})
}

// =============================================================================
// I18N AND LOCALES TESTS
// =============================================================================

func TestInternationalization(t *testing.T) {
	// Defer resetting the config to avoid affecting other tests
	defer Config(nil)

	t.Run("should return Chinese error messages when locale is set to zh-CN", func(t *testing.T) {
		// Set the global locale to Simplified Chinese
		Config(locales.ZhCN())

		// Schema that will fail
		schema := String().Min(5)

		// Parse invalid data
		_, err := schema.Parse("hi")
		require.Error(t, err)

		// Assert the error message is in Chinese
		assert.Contains(t, err.Error(), "数值过小", "Error message should be in Chinese")
	})

	t.Run("should return English error messages when locale is set to en", func(t *testing.T) {
		// Set the global locale to English
		Config(locales.EN())

		// Schema that will fail
		schema := String().Min(5)

		// Parse invalid data
		_, err := schema.Parse("hi")
		require.Error(t, err)

		// Assert the error message is in English
		assert.Contains(t, err.Error(), "Too small", "Error message should be in English")
	})

	t.Run("should return English error messages by default", func(t *testing.T) {
		// Reset config to default
		Config(nil)

		// Schema that will fail
		schema := String().Min(5)

		// Parse invalid data
		_, err := schema.Parse("hi")
		require.Error(t, err)

		// Assert the error message is in English
		assert.Contains(t, err.Error(), "Too small", "Default error message should be in English")
	})
}

// =============================================================================
// REGISTRY TESTS
// =============================================================================

type fieldMeta struct {
	Title       string            `json:"title"`
	Description string            `json:"description,omitempty"`
	Examples    []any             `json:"examples,omitempty"`
	Extra       map[string]string `json:"extra,omitempty"`
}

func TestRegistryCRUD(t *testing.T) {
	reg := NewRegistry[fieldMeta]()

	// Define a simple schema and associated metadata
	nameSchema := String().Min(1)
	meta := fieldMeta{Title: "User Name", Description: "The user's full name"}

	// Initially registry should not contain the schema
	if reg.Has(nameSchema) {
		t.Fatalf("expected Has to be false for new registry entry")
	}

	// Add metadata and verify
	reg.Add(nameSchema, meta)
	if !reg.Has(nameSchema) {
		t.Fatalf("expected Has to be true after Add")
	}

	got, ok := reg.Get(nameSchema)
	if !ok {
		t.Fatalf("expected Get to return true after Add")
	}
	if got.Title != meta.Title {
		t.Errorf("unexpected metadata title: got=%q want=%q", got.Title, meta.Title)
	}

	// Remove entry and verify it's gone
	reg.Remove(nameSchema)
	if reg.Has(nameSchema) {
		t.Fatalf("expected Has to be false after Remove")
	}
}

func TestGlobalRegistry(t *testing.T) {
	emailSchema := Email()

	meta := GlobalMeta{Title: "Email Address", Description: "A valid e-mail address"}

	// Ensure clean state by removing any pre-existing entry (if present)
	GlobalRegistry.Remove(emailSchema)

	// Store metadata in the global registry
	GlobalRegistry.Add(emailSchema, meta)

	if !GlobalRegistry.Has(emailSchema) {
		t.Fatalf("expected schema to be present in GlobalRegistry")
	}

	got, ok := GlobalRegistry.Get(emailSchema)
	if !ok {
		t.Fatalf("expected Get to succeed in GlobalRegistry")
	}
	if got.Title != meta.Title {
		t.Errorf("unexpected title in GlobalRegistry: got=%q want=%q", got.Title, meta.Title)
	}

	// Cleanup to avoid side-effects on other tests
	GlobalRegistry.Remove(emailSchema)
}

func TestRegistryConcurrentAccess(t *testing.T) {
	reg := NewRegistry[int]()

	// We will concurrently add and read entries to ensure thread-safety
	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines * 2)

	schema := Int()

	// Concurrent writers
	for i := 0; i < goroutines; i++ {
		go func(val int) {
			defer wg.Done()
			reg.Add(schema, val)
		}(i)
	}

	// Concurrent readers
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			_, _ = reg.Get(schema)
		}()
	}

	wg.Wait()

	// Registry should still contain a value for the schema
	if !reg.Has(schema) {
		t.Fatalf("registry lost metadata under concurrent access")
	}
}

// =============================================================================
// APPLY FUNCTION TESTS
// =============================================================================

func TestApply(t *testing.T) {
	t.Run("basic apply with same schema type", func(t *testing.T) {
		// Define a reusable modifier function
		addMinMax := func(s *ZodString[string]) *ZodString[string] {
			return s.Min(1).Max(100)
		}

		// Apply it to a schema
		schema := Apply(String(), addMinMax)

		// Test validation
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		// Test min constraint
		_, err = schema.Parse("")
		require.Error(t, err)

		// Test max constraint (should fail for string over 100 chars)
		longStr := make([]byte, 101)
		for i := range longStr {
			longStr[i] = 'a'
		}
		_, err = schema.Parse(string(longStr))
		require.Error(t, err)
	})

	t.Run("apply with type transformation", func(t *testing.T) {
		// Function that converts string schema to optional
		makeOptional := func(s *ZodString[string]) *ZodString[*string] {
			return s.Optional()
		}

		// Apply it
		schema := Apply(String().Min(1), makeOptional)

		// nil should be accepted for optional
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Valid string should work
		result, err = schema.Parse("test")
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "test", *result)
	})

	t.Run("apply with integer schema", func(t *testing.T) {
		// Define common integer checks
		addIntChecks := func(s *ZodInteger[int64, int64]) *ZodInteger[int64, int64] {
			return s.Min(0).Max(100)
		}

		schema := Apply(Int64(), addIntChecks)

		result, err := schema.Parse(int64(50))
		require.NoError(t, err)
		assert.Equal(t, int64(50), result)

		// Negative should fail
		_, err = schema.Parse(int64(-1))
		require.Error(t, err)

		// Over 100 should fail
		_, err = schema.Parse(int64(101))
		require.Error(t, err)
	})

	t.Run("chained apply calls", func(t *testing.T) {
		addMin := func(s *ZodString[string]) *ZodString[string] {
			return s.Min(2)
		}

		addMax := func(s *ZodString[string]) *ZodString[string] {
			return s.Max(10)
		}

		// Chain multiple apply calls
		schema := Apply(Apply(String(), addMin), addMax)

		// Valid string
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		// Too short
		_, err = schema.Parse("a")
		require.Error(t, err)

		// Too long
		_, err = schema.Parse("this is way too long")
		require.Error(t, err)
	})

	t.Run("apply followed by method chaining", func(t *testing.T) {
		addBase := func(s *ZodString[string]) *ZodString[string] {
			return s.Min(1)
		}

		// Apply then chain additional methods
		schema := Apply(String(), addBase).Max(50).Trim()

		result, err := schema.Parse("  hello  ")
		require.NoError(t, err)
		assert.Equal(t, "hello", result) // Trimmed

		_, err = schema.Parse("")
		require.Error(t, err)
	})

	t.Run("apply returns callback result type", func(t *testing.T) {
		// Function that returns a completely different type
		convertToInt := func(_ *ZodString[string]) *ZodInteger[int, int] {
			return Int().Min(0)
		}

		// The return type is now ZodInteger, not ZodString
		schema := Apply(String(), convertToInt)

		result, err := schema.Parse(42)
		require.NoError(t, err)
		assert.Equal(t, 42, result)

		_, err = schema.Parse(-1)
		require.Error(t, err)
	})

	t.Run("apply with object schema", func(t *testing.T) {
		addStrict := func(s *ZodObject[map[string]any, map[string]any]) *ZodObject[map[string]any, map[string]any] {
			return s.Strict()
		}

		schema := Apply(Object(ObjectSchema{
			"name": String(),
		}), addStrict)

		// Valid object
		result, err := schema.Parse(map[string]any{"name": "Alice"})
		require.NoError(t, err)
		assert.Equal(t, "Alice", result["name"])

		// Extra key should fail (strict mode)
		_, err = schema.Parse(map[string]any{"name": "Alice", "extra": "value"})
		require.Error(t, err)
	})

	t.Run("apply combined with nullable wrapper", func(t *testing.T) {
		// Zod v4 pattern: z.nullable(z.number().apply(setCommonNumberChecks))
		setChecks := func(s *ZodInteger[int, int]) *ZodInteger[int, int] {
			return s.Min(0).Max(100)
		}

		// Apply first, then make nilable
		schema := Apply(Int(), setChecks).Nilable()

		// Valid range
		result, err := schema.Parse(50)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, 50, *result)

		// nil is accepted
		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Out of range fails
		_, err = schema.Parse(-1)
		require.Error(t, err)

		_, err = schema.Parse(101)
		require.Error(t, err)
	})

	t.Run("apply callback return value is the result", func(t *testing.T) {
		// Zod v4: The callback's return value becomes the apply's return value
		// This tests that Apply returns exactly what the callback returns

		// Create a distinct schema via callback
		distinctSchema := Int().Min(10)
		getDistinctSchema := func(_ *ZodString[string]) *ZodInteger[int, int] {
			return distinctSchema
		}

		result := Apply(String(), getDistinctSchema)

		// The result should be the same schema instance from callback
		assert.Equal(t, distinctSchema.GetInternals(), result.GetInternals())
	})

	t.Run("apply with slice schema", func(t *testing.T) {
		addNonEmpty := func(s *ZodSlice[string, []string]) *ZodSlice[string, []string] {
			return s.NonEmpty().Max(5)
		}

		schema := Apply(Slice[string](String()), addNonEmpty)

		// Valid: non-empty within max
		result, err := schema.Parse([]any{"a", "b", "c"})
		require.NoError(t, err)
		assert.Len(t, result, 3)

		// Invalid: empty
		_, err = schema.Parse([]any{})
		require.Error(t, err)

		// Invalid: too many
		_, err = schema.Parse([]any{"1", "2", "3", "4", "5", "6"})
		require.Error(t, err)
	})

	t.Run("apply preserves schema immutability", func(t *testing.T) {
		addMin := func(s *ZodString[string]) *ZodString[string] {
			return s.Min(5)
		}

		original := String()
		applied := Apply(original, addMin)

		// Original should not have min constraint
		_, err := original.Parse("hi")
		require.NoError(t, err)

		// Applied should have min constraint
		_, err = applied.Parse("hi")
		require.Error(t, err)

		_, err = applied.Parse("hello")
		require.NoError(t, err)
	})
}
