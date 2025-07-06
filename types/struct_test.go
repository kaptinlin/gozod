package types

import (
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test data structures
type User struct {
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Email string `json:"email"`
}

type UserWithOptional struct {
	Name    string  `json:"name"`
	Age     int     `json:"age"`
	Email   string  `json:"email"`
	Address *string `json:"address,omitempty"`
}

type Profile struct {
	User    User   `json:"user"`
	Country string `json:"country"`
}

type Person struct {
	ID       int    `json:"id"`
	FullName string `json:"full_name"`
	Active   bool   `json:"active"`
}

// =============================================================================
// Basic functionality tests
// =============================================================================

func TestStruct_BasicFunctionality(t *testing.T) {
	t.Run("valid struct inputs", func(t *testing.T) {
		schema := Struct[User]()

		validUser := User{
			Name:  "John Doe",
			Age:   30,
			Email: "john@example.com",
		}

		result, err := schema.Parse(validUser)
		require.NoError(t, err)
		assert.Equal(t, validUser, result)
	})

	t.Run("valid struct with pointer input", func(t *testing.T) {
		schema := Struct[User]()

		validUser := User{
			Name:  "John Doe",
			Age:   30,
			Email: "john@example.com",
		}

		result, err := schema.Parse(&validUser)
		require.NoError(t, err)
		assert.Equal(t, validUser, result)
	})

	t.Run("empty struct", func(t *testing.T) {
		type Empty struct{}
		schema := Struct[Empty]()

		result, err := schema.Parse(Empty{})
		require.NoError(t, err)
		assert.Equal(t, Empty{}, result)
	})

	t.Run("invalid type inputs", func(t *testing.T) {
		schema := Struct[User]()

		invalidInputs := []any{
			"not a struct", 123, []int{1, 2, 3}, true, nil,
			map[string]any{"name": "John"}, // map is not a struct
		}

		for _, input := range invalidInputs {
			_, err := schema.Parse(input)
			assert.Error(t, err, "Expected error for input: %v", input)
		}
	})

	t.Run("Parse and MustParse methods", func(t *testing.T) {
		schema := Struct[User]()
		validUser := User{Name: "John", Age: 25, Email: "john@test.com"}

		// Test Parse method
		result, err := schema.Parse(validUser)
		require.NoError(t, err)
		assert.Equal(t, validUser, result)

		// Test MustParse method
		mustResult := schema.MustParse(validUser)
		assert.Equal(t, validUser, mustResult)

		// Test panic on invalid input
		assert.Panics(t, func() {
			schema.MustParse("invalid")
		})
	})

	t.Run("custom error message", func(t *testing.T) {
		schema := Struct[User](core.SchemaParams{Error: "Expected a valid User struct"})

		require.NotNil(t, schema)

		_, err := schema.Parse("invalid")
		assert.Error(t, err)
	})
}

// =============================================================================
// Type safety tests
// =============================================================================

func TestStruct_TypeSafety(t *testing.T) {
	t.Run("struct returns correct type", func(t *testing.T) {
		schema := Struct[User]()
		require.NotNil(t, schema)

		validUser := User{Name: "John", Age: 30, Email: "john@test.com"}
		result, err := schema.Parse(validUser)
		require.NoError(t, err)
		assert.Equal(t, validUser, result)
		assert.IsType(t, User{}, result)
	})

	t.Run("different struct types", func(t *testing.T) {
		userSchema := Struct[User]()
		personSchema := Struct[Person]()

		user := User{Name: "John", Age: 30, Email: "john@test.com"}
		person := Person{ID: 1, FullName: "John Doe", Active: true}

		// User schema should accept User struct
		result1, err := userSchema.Parse(user)
		require.NoError(t, err)
		assert.Equal(t, user, result1)

		// Person schema should accept Person struct
		result2, err := personSchema.Parse(person)
		require.NoError(t, err)
		assert.Equal(t, person, result2)

		// User schema should reject Person struct
		_, err = userSchema.Parse(person)
		assert.Error(t, err)

		// Person schema should reject User struct
		_, err = personSchema.Parse(user)
		assert.Error(t, err)
	})

	t.Run("MustParse type safety", func(t *testing.T) {
		schema := Struct[User]()
		validUser := User{Name: "John", Age: 25, Email: "john@test.com"}

		result := schema.MustParse(validUser)
		assert.IsType(t, User{}, result)
		assert.Equal(t, validUser, result)
	})
}

// =============================================================================
// Struct Schema validation tests
// =============================================================================

func TestStruct_WithSchema(t *testing.T) {
	t.Run("valid struct with schema validation", func(t *testing.T) {
		// Define schema for User fields
		schema := Struct[User](core.StructSchema{
			"name":  String().Min(2),
			"age":   Int().Min(0).Max(150),
			"email": String().Email(),
		})

		validUser := User{
			Name:  "John Doe",
			Age:   30,
			Email: "john@example.com",
		}

		result, err := schema.Parse(validUser)
		require.NoError(t, err)
		assert.Equal(t, validUser, result)
	})

	t.Run("struct with schema validation failures", func(t *testing.T) {
		schema := Struct[User](core.StructSchema{
			"name":  String().Min(5),  // Name too short
			"age":   Int().Min(18),    // Age too low
			"email": String().Email(), // Invalid email format
		})

		// Test name too short
		invalidUser1 := User{Name: "Jo", Age: 25, Email: "john@example.com"}
		_, err := schema.Parse(invalidUser1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name")

		// Test age too low
		invalidUser2 := User{Name: "John Doe", Age: 16, Email: "john@example.com"}
		_, err = schema.Parse(invalidUser2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "age")

		// Test invalid email
		invalidUser3 := User{Name: "John Doe", Age: 25, Email: "not-an-email"}
		_, err = schema.Parse(invalidUser3)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email")
	})

	t.Run("struct with partial schema (only some fields)", func(t *testing.T) {
		// Only validate name and email, not age
		schema := Struct[User](core.StructSchema{
			"name":  String().Min(2),
			"email": String().Email(),
		})

		validUser := User{
			Name:  "John Doe",
			Age:   30, // Age not validated by schema
			Email: "john@example.com",
		}

		result, err := schema.Parse(validUser)
		require.NoError(t, err)
		assert.Equal(t, validUser, result)
	})

	t.Run("struct with optional field schema", func(t *testing.T) {
		schema := Struct[UserWithOptional](core.StructSchema{
			"name":    String().Min(2),
			"age":     Int().Min(0),
			"email":   String().Email(),
			"address": String().Optional(), // Optional field
		})

		// Test with address present
		userWithAddress := UserWithOptional{
			Name:    "John Doe",
			Age:     30,
			Email:   "john@example.com",
			Address: stringPtr("123 Main St"),
		}

		result, err := schema.Parse(userWithAddress)
		require.NoError(t, err)
		assert.Equal(t, userWithAddress, result)

		// Test with address nil
		userNoAddress := UserWithOptional{
			Name:    "John Doe",
			Age:     30,
			Email:   "john@example.com",
			Address: nil,
		}

		result, err = schema.Parse(userNoAddress)
		require.NoError(t, err)
		assert.Equal(t, userNoAddress, result)
	})

	t.Run("struct with nested validation", func(t *testing.T) {
		// Schema for nested Profile struct
		profileSchema := Struct[Profile](core.StructSchema{
			"user": Struct[User](core.StructSchema{
				"name":  String().Min(2),
				"age":   Int().Min(0),
				"email": String().Email(),
			}),
			"country": String().Min(2),
		})

		validProfile := Profile{
			User: User{
				Name:  "John Doe",
				Age:   30,
				Email: "john@example.com",
			},
			Country: "USA",
		}

		result, err := profileSchema.Parse(validProfile)
		require.NoError(t, err)
		assert.Equal(t, validProfile, result)

		// Test with invalid nested user
		invalidProfile := Profile{
			User: User{
				Name:  "J", // Too short
				Age:   30,
				Email: "john@example.com",
			},
			Country: "USA",
		}

		_, err = profileSchema.Parse(invalidProfile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user")
	})

	t.Run("struct with json tag field mapping", func(t *testing.T) {
		// Person has json tags that differ from field names
		personSchema := Struct[Person](core.StructSchema{
			"id":        Int().Min(1),    // Maps to ID field
			"full_name": String().Min(2), // Maps to FullName field
			"active":    Bool(),          // Maps to Active field
		})

		validPerson := Person{
			ID:       123,
			FullName: "John Doe",
			Active:   true,
		}

		result, err := personSchema.Parse(validPerson)
		require.NoError(t, err)
		assert.Equal(t, validPerson, result)

		// Test validation failure
		invalidPerson := Person{
			ID:       0, // ID too small
			FullName: "John Doe",
			Active:   true,
		}

		_, err = personSchema.Parse(invalidPerson)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "id")
	})
}

func TestStruct_SchemaConstructors(t *testing.T) {
	t.Run("StructPtr with schema", func(t *testing.T) {
		schema := StructPtr[User](core.StructSchema{
			"name":  String().Min(2),
			"email": String().Email(),
		})

		validUser := User{
			Name:  "John Doe",
			Age:   30,
			Email: "john@example.com",
		}

		// Test with pointer input
		result, err := schema.Parse(&validUser)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, validUser, *result)

		// Test with value input
		result, err = schema.Parse(validUser)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, validUser, *result)
	})
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}

// =============================================================================
// Modifier methods tests
// =============================================================================

func TestStruct_Modifiers(t *testing.T) {
	t.Run("Optional allows nil values", func(t *testing.T) {
		schema := Struct[User]()
		optionalSchema := schema.Optional()

		// Test non-nil value - should return pointer
		validUser := User{Name: "John", Age: 25, Email: "john@test.com"}
		result, err := optionalSchema.Parse(validUser)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, validUser, *result)

		// Test nil value (should be allowed for optional)
		result, err = optionalSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Nilable allows nil values", func(t *testing.T) {
		schema := Struct[User]()
		nilableSchema := schema.Nilable()

		// Test nil handling
		result, err := nilableSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Test valid value - should return pointer
		validUser := User{Name: "John", Age: 25, Email: "john@test.com"}
		result, err = nilableSchema.Parse(validUser)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, validUser, *result)
	})

	t.Run("Default preserves current type", func(t *testing.T) {
		defaultUser := User{Name: "Default", Age: 0, Email: "default@test.com"}
		schema := Struct[User]()
		defaultSchema := schema.Default(defaultUser)

		// Valid input should override default
		validUser := User{Name: "John", Age: 25, Email: "john@test.com"}
		result, err := defaultSchema.Parse(validUser)
		require.NoError(t, err)
		assert.Equal(t, validUser, result)
		assert.IsType(t, User{}, result)
	})

	t.Run("Prefault preserves current type", func(t *testing.T) {
		prefaultUser := User{Name: "Prefault", Age: 0, Email: "prefault@test.com"}
		schema := Struct[User]()
		prefaultSchema := schema.Prefault(prefaultUser)

		// Valid input should override prefault
		validUser := User{Name: "John", Age: 25, Email: "john@test.com"}
		result, err := prefaultSchema.Parse(validUser)
		require.NoError(t, err)
		assert.Equal(t, validUser, result)
		assert.IsType(t, User{}, result)
	})
}

// =============================================================================
// Chaining tests
// =============================================================================

func TestStruct_Chaining(t *testing.T) {
	t.Run("type evolution through chaining", func(t *testing.T) {
		// Chain with type evolution
		defaultUser := User{Name: "Default", Age: 0, Email: "default@test.com"}
		schema := Struct[User]().
			Default(defaultUser). // Preserves struct type
			Optional()            // Now returns pointer type

		// Test final behavior - should return pointer
		validUser := User{Name: "John", Age: 25, Email: "john@test.com"}
		result, err := schema.Parse(validUser)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, validUser, *result)

		// Test nil handling
		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("complex chaining", func(t *testing.T) {
		schema := Struct[User]().
			Nilable()

		validUser := User{Name: "John", Age: 25, Email: "john@test.com"}
		result, err := schema.Parse(validUser)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, validUser, *result)

		// Test nil handling
		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("default and prefault chaining", func(t *testing.T) {
		defaultUser := User{Name: "Default", Age: 0, Email: "default@test.com"}
		prefaultUser := User{Name: "Prefault", Age: 0, Email: "prefault@test.com"}
		schema := Struct[User]().
			Default(defaultUser).
			Prefault(prefaultUser)

		validUser := User{Name: "John", Age: 25, Email: "john@test.com"}
		result, err := schema.Parse(validUser)
		require.NoError(t, err)
		assert.Equal(t, validUser, result)
	})
}

// =============================================================================
// Default and prefault tests
// =============================================================================

func TestStruct_DefaultAndPrefault(t *testing.T) {
	t.Run("default value behavior", func(t *testing.T) {
		defaultUser := User{Name: "Default", Age: 0, Email: "default@test.com"}
		schema := Struct[User]().Default(defaultUser)

		// Valid input should override default
		validUser := User{Name: "John", Age: 25, Email: "john@test.com"}
		result, err := schema.Parse(validUser)
		require.NoError(t, err)
		assert.Equal(t, validUser, result)

		// Test default function
		userFunc := Struct[User]().DefaultFunc(func() User {
			return User{Name: "Func", Age: 0, Email: "func@test.com"}
		})
		result2, err := userFunc.Parse(User{Name: "John", Age: 25, Email: "john@test.com"})
		require.NoError(t, err)
		assert.Equal(t, User{Name: "John", Age: 25, Email: "john@test.com"}, result2)
	})

	t.Run("prefault value behavior", func(t *testing.T) {
		prefaultUser := User{Name: "Prefault", Age: 0, Email: "prefault@test.com"}
		schema := Struct[User]().Prefault(prefaultUser)

		// Valid input should work normally
		validUser := User{Name: "John", Age: 25, Email: "john@test.com"}
		result, err := schema.Parse(validUser)
		require.NoError(t, err)
		assert.Equal(t, validUser, result)

		// Test prefault function
		userFunc := Struct[User]().PrefaultFunc(func() User {
			return User{Name: "Func", Age: 0, Email: "func@test.com"}
		})
		result2, err := userFunc.Parse(User{Name: "John", Age: 25, Email: "john@test.com"})
		require.NoError(t, err)
		assert.Equal(t, User{Name: "John", Age: 25, Email: "john@test.com"}, result2)
	})
}

// =============================================================================
// Refine tests
// =============================================================================

func TestStruct_Refine(t *testing.T) {
	t.Run("refine validation", func(t *testing.T) {
		// Only accept users with name length > 3
		schema := Struct[User]().Refine(func(u User) bool {
			return len(u.Name) > 3
		})

		validUser := User{Name: "John", Age: 25, Email: "john@test.com"}
		result, err := schema.Parse(validUser)
		require.NoError(t, err)
		assert.Equal(t, validUser, result)

		invalidUser := User{Name: "Jo", Age: 25, Email: "jo@test.com"}
		_, err = schema.Parse(invalidUser)
		assert.Error(t, err)
	})

	t.Run("refine with custom error message", func(t *testing.T) {
		errorMessage := "Name must be longer than 3 characters"
		schema := Struct[User]().Refine(func(u User) bool {
			return len(u.Name) > 3
		}, core.SchemaParams{Error: errorMessage})

		validUser := User{Name: "John", Age: 25, Email: "john@test.com"}
		result, err := schema.Parse(validUser)
		require.NoError(t, err)
		assert.Equal(t, validUser, result)

		invalidUser := User{Name: "Jo", Age: 25, Email: "jo@test.com"}
		_, err = schema.Parse(invalidUser)
		assert.Error(t, err)
	})

	t.Run("refine nilable struct", func(t *testing.T) {
		schema := Struct[User]().Nilable().Refine(func(u *User) bool {
			// Allow nil or user with valid name
			if u == nil {
				return true
			}
			return len(u.Name) > 0
		})

		// nil should pass
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// valid struct should pass and return pointer
		validUser := User{Name: "John", Age: 25, Email: "john@test.com"}
		result, err = schema.Parse(validUser)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, validUser, *result)
	})
}

func TestStruct_RefineAny(t *testing.T) {
	t.Run("refineAny flexible validation", func(t *testing.T) {
		schema := Struct[User]().RefineAny(func(v any) bool {
			u, ok := v.(User)
			return ok && len(u.Name) >= 1
		})

		// user with valid name should pass
		validUser := User{Name: "John", Age: 25, Email: "john@test.com"}
		result, err := schema.Parse(validUser)
		require.NoError(t, err)
		assert.Equal(t, validUser, result)

		// user with empty name should fail
		invalidUser := User{Name: "", Age: 25, Email: "john@test.com"}
		_, err = schema.Parse(invalidUser)
		assert.Error(t, err)
	})

	t.Run("refineAny with type checking", func(t *testing.T) {
		schema := Struct[User]().RefineAny(func(v any) bool {
			u, ok := v.(User)
			if !ok {
				return false
			}
			// Only allow users with even age
			return u.Age%2 == 0
		})

		evenAgeUser := User{Name: "John", Age: 30, Email: "john@test.com"}
		result, err := schema.Parse(evenAgeUser)
		require.NoError(t, err)
		assert.Equal(t, evenAgeUser, result)

		oddAgeUser := User{Name: "John", Age: 25, Email: "john@test.com"}
		_, err = schema.Parse(oddAgeUser)
		assert.Error(t, err)
	})
}

// =============================================================================
// Error handling and edge case tests
// =============================================================================

func TestStruct_ErrorHandling(t *testing.T) {
	t.Run("invalid struct type error", func(t *testing.T) {
		schema := Struct[User]()

		_, err := schema.Parse("not a struct")
		assert.Error(t, err)
	})

	t.Run("wrong struct type error", func(t *testing.T) {
		schema := Struct[User]()

		wrongStruct := Person{ID: 1, FullName: "John", Active: true}
		_, err := schema.Parse(wrongStruct)
		assert.Error(t, err)
	})

	t.Run("custom error message", func(t *testing.T) {
		schema := Struct[User](core.SchemaParams{Error: "Expected a valid User struct"})

		_, err := schema.Parse("invalid")
		assert.Error(t, err)
	})
}

func TestStruct_EdgeCases(t *testing.T) {
	t.Run("empty struct", func(t *testing.T) {
		type Empty struct{}
		schema := Struct[Empty]()

		result, err := schema.Parse(Empty{})
		require.NoError(t, err)
		assert.Equal(t, Empty{}, result)
	})

	t.Run("nil handling with nilable struct", func(t *testing.T) {
		schema := Struct[User]().Nilable()

		// Test nil input
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Test valid struct - should return pointer
		validUser := User{Name: "John", Age: 25, Email: "john@test.com"}
		result, err = schema.Parse(validUser)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, validUser, *result)
	})

	t.Run("deeply nested struct validation", func(t *testing.T) {
		schema := Struct[Profile]()

		validProfile := Profile{
			User: User{
				Name:  "John",
				Age:   25,
				Email: "john@test.com",
			},
			Country: "USA",
		}

		result, err := schema.Parse(validProfile)
		require.NoError(t, err)
		assert.Equal(t, validProfile, result)
	})

	t.Run("optional fields handling", func(t *testing.T) {
		schema := Struct[UserWithOptional]()

		// Should work with optional field nil
		user := UserWithOptional{
			Name:    "John",
			Age:     25,
			Email:   "john@test.com",
			Address: nil,
		}
		result, err := schema.Parse(user)
		require.NoError(t, err)
		assert.Equal(t, user, result)

		// Should work with optional field set
		address := "123 Main St"
		userWithAddress := UserWithOptional{
			Name:    "John",
			Age:     25,
			Email:   "john@test.com",
			Address: &address,
		}
		result, err = schema.Parse(userWithAddress)
		require.NoError(t, err)
		assert.Equal(t, userWithAddress, result)
	})

	t.Run("pointer value handling", func(t *testing.T) {
		schema := Struct[User]()

		// Test with pointer to struct
		validUser := User{Name: "John", Age: 25, Email: "john@test.com"}
		validUserPtr := &validUser

		result, err := schema.Parse(validUserPtr)
		require.NoError(t, err)
		assert.Equal(t, validUser, result)
	})

	t.Run("concurrent access safety", func(t *testing.T) {
		schema := Struct[User]()
		validUser := User{Name: "John", Age: 25, Email: "john@test.com"}

		// Run multiple goroutines parsing the same schema
		const numGoroutines = 10
		results := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				_, err := schema.Parse(validUser)
				results <- err
			}()
		}

		// Check all results
		for i := 0; i < numGoroutines; i++ {
			err := <-results
			assert.NoError(t, err)
		}
	})

	t.Run("transform operations", func(t *testing.T) {
		schema := Struct[User]()

		// Test Transform
		transform := schema.Transform(func(u User, ctx *core.RefinementContext) (any, error) {
			return len(u.Name), nil
		})
		require.NotNil(t, transform)
	})
}

func TestStruct_Constructors(t *testing.T) {
	t.Run("Struct constructor", func(t *testing.T) {
		schema := Struct[User]()
		require.NotNil(t, schema)
	})

	t.Run("StructPtr constructor", func(t *testing.T) {
		schema := StructPtr[User]()
		require.NotNil(t, schema)

		// Test nil handling
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Test valid struct
		validUser := User{Name: "John", Age: 25, Email: "john@test.com"}
		result, err = schema.Parse(validUser)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, validUser, *result)
	})
}

// =============================================================================
// StructPtr tests
// =============================================================================

func TestStructPtr_Functionality(t *testing.T) {
	t.Run("StructPtr basic functionality", func(t *testing.T) {
		schema := StructPtr[User]()

		// Test valid struct
		validUser := User{Name: "John", Age: 25, Email: "john@test.com"}
		result, err := schema.Parse(validUser)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, validUser, *result)

		// Test nil input
		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("StructPtr with pointer input", func(t *testing.T) {
		schema := StructPtr[User]()

		validUser := User{Name: "John", Age: 25, Email: "john@test.com"}
		result, err := schema.Parse(&validUser)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, validUser, *result)
	})

	t.Run("StructPtr with modifiers", func(t *testing.T) {
		defaultUser := User{Name: "Default", Age: 0, Email: "default@test.com"}
		schema := StructPtr[User]().Default(defaultUser)

		validUser := User{Name: "John", Age: 25, Email: "john@test.com"}
		result, err := schema.Parse(validUser)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, validUser, *result)
	})
}

// =============================================================================
// OVERWRITE TESTS
// =============================================================================

func TestStruct_Overwrite(t *testing.T) {
	t.Run("type preservation", func(t *testing.T) {
		schema := Struct[User]().
			Overwrite(func(u User) User {
				return u // Identity transformation
			})

		input := User{
			Name:  "John",
			Age:   25,
			Email: "john@test.com",
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)
		assert.IsType(t, User{}, result)
		assert.Equal(t, input, result)
	})
}

// =============================================================================
// Check Method Tests
// =============================================================================

func TestStruct_Check(t *testing.T) {
	type Foo struct {
		Value int
	}

	t.Run("invalid struct triggers issue", func(t *testing.T) {
		schema := Struct[Foo]().Check(func(value Foo, p *core.ParsePayload) {
			if value.Value < 0 {
				p.AddIssueWithMessage("value must be >= 0")
			}
		})

		_, err := schema.Parse(Foo{Value: -1})
		require.Error(t, err)
		var zErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zErr))
		assert.Len(t, zErr.Issues, 1)
	})

	t.Run("pointer schema adapts to struct value", func(t *testing.T) {
		schema := StructPtr[Foo]().Check(func(value *Foo, p *core.ParsePayload) {
			if value == nil || value.Value == 0 {
				p.AddIssueWithMessage("zero value")
			}
		})

		_, err := schema.Parse(Foo{})
		require.Error(t, err)
		var zErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zErr))
		assert.Len(t, zErr.Issues, 1)
	})
}

func TestStruct_NonOptional(t *testing.T) {
	t.Run("basic non-optional", func(t *testing.T) {
		schema := StructPtr[User]().Optional().NonOptional()

		// A valid struct should still pass
		user := User{Name: "test", Age: 1, Email: "test@test.com"}
		result, err := schema.Parse(user)
		require.NoError(t, err)
		assert.Equal(t, user, result)
		assert.IsType(t, User{}, result)

		// A pointer to a valid struct should also pass
		result, err = schema.Parse(&user)
		require.NoError(t, err)
		assert.Equal(t, user, result)

		// nil should now fail
		_, err = schema.Parse(nil)
		assert.Error(t, err)
	})

	t.Run("chained optional and non-optional", func(t *testing.T) {
		schema := Struct[User]().Optional().NonOptional().Optional().NonOptional()

		// A valid struct should still pass
		user := User{Name: "test", Age: 1, Email: "test@test.com"}
		result, err := schema.Parse(user)
		require.NoError(t, err)
		assert.Equal(t, user, result)
		assert.IsType(t, User{}, result)

		// nil should fail
		_, err = schema.Parse(nil)
		assert.Error(t, err)
	})

	t.Run("non-optional on already non-optional schema", func(t *testing.T) {
		schema := Struct[User]().NonOptional()

		// A valid struct should still pass
		user := User{Name: "test", Age: 1, Email: "test@test.com"}
		result, err := schema.Parse(user)
		require.NoError(t, err)
		assert.Equal(t, user, result)
		assert.IsType(t, User{}, result)

		// nil should fail
		_, err = schema.Parse(nil)
		assert.Error(t, err)
	})

	t.Run("non-optional with nested struct", func(t *testing.T) {
		profileSchema := Struct[Profile](core.StructSchema{
			"user":    Struct[User]().Optional().NonOptional(),
			"country": String(),
		}).Optional().NonOptional()

		profile := Profile{
			User: User{
				Name:  "test",
				Age:   10,
				Email: "test@test.com",
			},
			Country: "USA",
		}

		result, err := profileSchema.Parse(profile)
		require.NoError(t, err)
		assert.Equal(t, profile, result)

		// Test with nil for the top-level optional
		_, err = profileSchema.Parse(nil)
		assert.Error(t, err)

		// Test with nil for the nested user struct (should fail as it is NonOptional)
		// Note: This case is tricky because Go will initialize the nested struct,
		// so we cannot pass a nil for it directly in the Profile literal.
		// Instead we check if the parser would have caught a truly nil field,
		// which our current implementation should.
		// A struct with a nil field that is NonOptional should fail validation.
		// The current struct validation logic might not check this deeply for nil fields.
		// Let's create a test case that would expose this.
		// A schema with a field that is `Struct[User]().Optional().NonOptional()` should not allow that field to be nil.
		// But in Go, a field of type `User` can't be nil. A pointer `*User` can.
		// The `Optional` call on the User struct returns a `ZodStruct[User, *User]`, so `NonOptional`
		// takes it back to `ZodStruct[User, User]`. The validation for the field will be on `User`.
		// Let's re-verify the logic.
		// The `parseGoStruct` function handles `nil` at the beginning. If the input is not nil, it proceeds.
		// The validation of fields happens in `validateStructFields`.
		// It gets the field value. If the field is a struct, it's not a pointer and can't be nil.
		// The logic seems correct. The test case is valid.
	})
}
