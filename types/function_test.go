package types

import (
	"strings"
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFunction_BasicFunctionality(t *testing.T) {
	t.Run("valid function inputs", func(t *testing.T) {
		schema := Function()

		fn := func(x int) int { return x * 2 }
		result, err := schema.Parse(fn)
		require.NoError(t, err)
		assert.NotNil(t, result)

		multi := func(a string, b int) string { return a }
		result, err = schema.Parse(multi)
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("invalid type inputs", func(t *testing.T) {
		schema := Function()

		inputs := []any{
			"not a function", 123, 3.14, []int{1, 2, 3}, nil, map[string]int{},
		}
		for _, input := range inputs {
			_, err := schema.Parse(input)
			assert.Error(t, err, "expected error for input: %v", input)
		}
	})

	t.Run("Parse and MustParse", func(t *testing.T) {
		schema := Function()
		fn := func() string { return "test" }

		result, err := schema.Parse(fn)
		require.NoError(t, err)
		assert.NotNil(t, result)

		got := schema.MustParse(fn)
		assert.NotNil(t, got)

		assert.Panics(t, func() { schema.MustParse("invalid") })
	})

	t.Run("custom error message", func(t *testing.T) {
		schema := Function(core.SchemaParams{Error: "Expected a function value"})
		require.NotNil(t, schema)
		assert.Equal(t, core.ZodTypeFunction, schema.internals.Def.Type)

		_, err := schema.Parse("invalid")
		assert.Error(t, err)
	})
}

func TestFunction_TypeSafety(t *testing.T) {
	t.Run("Function returns any type", func(t *testing.T) {
		schema := Function()

		fn := func(x int) int { return x + 1 }
		result, err := schema.Parse(fn)
		require.NoError(t, err)

		if got, ok := result.(func(int) int); ok {
			assert.Equal(t, 6, got(5))
		} else {
			t.Errorf("got type %T, want func(int) int", result)
		}
	})

	t.Run("FunctionPtr returns *any type", func(t *testing.T) {
		schema := FunctionPtr()

		fn := func(x string) string { return x + "!" }
		result, err := schema.Parse(fn)
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("type inference with assignment", func(t *testing.T) {
		valSchema := Function()
		ptrSchema := FunctionPtr()
		fn := func() bool { return true }

		result, err := valSchema.Parse(fn)
		require.NoError(t, err)
		assert.NotNil(t, result)

		result2, err := ptrSchema.Parse(fn)
		require.NoError(t, err)
		assert.NotNil(t, result2)
	})

	t.Run("MustParse type safety", func(t *testing.T) {
		schema := Function()
		fn := func(x float64) float64 { return x * 2 }

		result := schema.MustParse(fn)
		assert.NotNil(t, result)

		if got, ok := result.(func(float64) float64); ok {
			assert.Equal(t, 7.0, got(3.5))
		} else {
			t.Errorf("got type %T, want func(float64) float64", result)
		}
	})
}

func TestFunction_Modifiers(t *testing.T) {
	t.Run("Optional allows nil", func(t *testing.T) {
		schema := Function().Optional()

		fn := func() {}
		result, err := schema.Parse(fn)
		require.NoError(t, err)
		assert.NotNil(t, result)

		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Nilable allows nil", func(t *testing.T) {
		schema := Function().Nilable()

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Default preserves current type", func(t *testing.T) {
		dflt := func() string { return "default" }
		_ = Function().Default(dflt)
		_ = FunctionPtr().Default(dflt)
	})

	t.Run("Prefault preserves current type", func(t *testing.T) {
		pf := func() int { return 42 }
		_ = Function().Prefault(pf)
		_ = FunctionPtr().Prefault(pf)
	})
}

func TestFunction_Chaining(t *testing.T) {
	t.Run("type evolution through chaining", func(t *testing.T) {
		dflt := func() {}
		schema := Function().Default(dflt).Optional()

		fn := func(x int) int { return x * 2 }
		result, err := schema.Parse(fn)
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("complex chaining", func(t *testing.T) {
		dflt := func() string { return "test" }
		schema := FunctionPtr().Nilable().Default(dflt)

		result, err := schema.Parse(func() {})
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("default and prefault chaining", func(t *testing.T) {
		dflt := func() int { return 1 }
		pf := func() int { return 2 }

		schema := Function().Default(dflt).Prefault(pf)

		fn := func() int { return 3 }
		result, err := schema.Parse(fn)
		require.NoError(t, err)
		assert.NotNil(t, result)
	})
}

func TestFunction_DefaultAndPrefault(t *testing.T) {
	t.Run("Default has higher priority than Prefault", func(t *testing.T) {
		dflt := func() string { return "default" }
		pf := func() string { return "prefault" }
		schema := Function().Default(dflt).Prefault(pf)

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.NotNil(t, result)
		if fn, ok := result.(func() string); ok {
			assert.Equal(t, "default", fn())
		} else {
			t.Errorf("got type %T, want func() string", result)
		}
	})

	t.Run("Default short-circuits validation", func(t *testing.T) {
		schema := Function().Refine(func(f any) bool {
			return false
		}, "Should never pass").Default("not-a-function")

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "not-a-function", result)
	})

	t.Run("Prefault goes through full validation", func(t *testing.T) {
		pf := func() int { return 42 }
		schema := Function().Refine(func(f any) bool {
			if fn, ok := f.(func() int); ok {
				return fn() > 0
			}
			return false
		}, "Function must return positive int").Prefault(pf)

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.NotNil(t, result)
		if fn, ok := result.(func() int); ok {
			assert.Equal(t, 42, fn())
		} else {
			t.Errorf("got type %T, want func() int", result)
		}
	})

	t.Run("Prefault only triggered by nil input", func(t *testing.T) {
		pf := func() string { return "prefault" }
		schema := Function().Prefault(pf)

		fn := func() string { return "test" }
		result, err := schema.Parse(fn)
		require.NoError(t, err)
		assert.NotNil(t, result)
		if got, ok := result.(func() string); ok {
			assert.Equal(t, "test", got())
		} else {
			t.Errorf("got type %T, want func() string", result)
		}

		_, err = schema.Parse("not-a-function")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid input: expected function")
	})

	t.Run("DefaultFunc and PrefaultFunc behavior", func(t *testing.T) {
		dfltCalled := false
		pfCalled := false

		dfltProvider := func() any {
			dfltCalled = true
			return func() string { return "default" }
		}
		pfProvider := func() any {
			pfCalled = true
			return func() string { return "prefault" }
		}

		schema := Function().DefaultFunc(dfltProvider).PrefaultFunc(pfProvider)
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.True(t, dfltCalled)
		assert.False(t, pfCalled)
		assert.NotNil(t, result)
		if fn, ok := result.(func() string); ok {
			assert.Equal(t, "default", fn())
		} else {
			t.Errorf("got type %T, want func() string", result)
		}

		dfltCalled = false
		pfCalled = false

		schema2 := Function().PrefaultFunc(pfProvider)
		result2, err := schema2.Parse(nil)
		require.NoError(t, err)
		assert.True(t, pfCalled)
		assert.NotNil(t, result2)
		if fn, ok := result2.(func() string); ok {
			assert.Equal(t, "prefault", fn())
		} else {
			t.Errorf("got type %T, want func() string", result2)
		}
	})

	t.Run("Prefault validation failure", func(t *testing.T) {
		bad := func() int { return -1 }
		schema := Function().Refine(func(f any) bool {
			if fn, ok := f.(func() int); ok {
				return fn() > 0
			}
			return false
		}, "Function must return positive int").Prefault(bad)

		_, err := schema.Parse(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Function must return positive int")
	})

	t.Run("FunctionPtr with Default and Prefault", func(t *testing.T) {
		dflt := func() string { return "default" }
		pf := func() string { return "prefault" }
		schema := FunctionPtr().Default(dflt).Prefault(pf)

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		if fn, ok := (*result).(func() string); ok {
			assert.Equal(t, "default", fn())
		} else {
			t.Errorf("got type %T, want func() string", *result)
		}
	})
}

func TestFunction_Refine(t *testing.T) {
	t.Run("refine validate", func(t *testing.T) {
		schema := Function().Refine(func(f any) bool {
			return f != nil
		})

		result, err := schema.Parse(func() {})
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("refine with custom error message", func(t *testing.T) {
		schema := Function().Refine(func(f any) bool {
			return f != nil
		}, core.SchemaParams{Error: "Function must not be nil"})

		result, err := schema.Parse(func() {})
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("always failing refine", func(t *testing.T) {
		schema := Function().Refine(func(f any) bool {
			return false
		})

		_, err := schema.Parse(func() {})
		assert.Error(t, err)
	})
}

func TestFunction_RefineAny(t *testing.T) {
	t.Run("refineAny function schema", func(t *testing.T) {
		schema := Function().RefineAny(func(v any) bool {
			return v != nil
		})

		result, err := schema.Parse(func() {})
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("refineAny with complex validation", func(t *testing.T) {
		schema := Function().RefineAny(func(v any) bool {
			return v != nil
		})

		result, err := schema.Parse(func(x int) int { return x })
		require.NoError(t, err)
		assert.NotNil(t, result)
	})
}

func TestFunction_TypeSpecificMethods(t *testing.T) {
	t.Run("Input method sets input schema", func(t *testing.T) {
		schema := Function().Input(nil)
		require.NotNil(t, schema)
		assert.Nil(t, schema.internals.Input)
	})

	t.Run("Output method sets output schema", func(t *testing.T) {
		schema := Function().Output(nil)
		require.NotNil(t, schema)
		assert.Nil(t, schema.internals.Output)
	})

	t.Run("Input and Output chaining", func(t *testing.T) {
		schema := Function().Input(nil).Output(nil)
		require.NotNil(t, schema)
		assert.Nil(t, schema.internals.Input)
		assert.Nil(t, schema.internals.Output)
	})

	t.Run("Implement wraps function", func(t *testing.T) {
		schema := Function()
		orig := func(x int) int { return x * 2 }

		wrapped, err := schema.Implement(orig)
		require.NoError(t, err)
		assert.NotNil(t, wrapped)

		if fn, ok := wrapped.(func(int) int); ok {
			assert.Equal(t, 10, fn(5))
		} else {
			t.Errorf("got type %T, want func(int) int", wrapped)
		}
	})

	t.Run("Implement with invalid input", func(t *testing.T) {
		schema := Function()

		_, err := schema.Implement("not a function")
		assert.Error(t, err)

		_, err = schema.Implement(123)
		assert.Error(t, err)

		_, err = schema.Implement(nil)
		assert.Error(t, err)
	})

	t.Run("Implement with input/output validation", func(t *testing.T) {
		schema := Function()

		fn := func(x string) int { return len(x) }
		wrapped, err := schema.Implement(fn)
		require.NoError(t, err)
		assert.NotNil(t, wrapped)
	})
}

func TestFunction_ErrorHandling(t *testing.T) {
	t.Run("invalid type error", func(t *testing.T) {
		schema := Function()

		_, err := schema.Parse("not a function")
		assert.Error(t, err)

		_, err = schema.Parse(123)
		assert.Error(t, err)

		_, err = schema.Parse([]int{1, 2, 3})
		assert.Error(t, err)
	})

	t.Run("custom error message", func(t *testing.T) {
		schema := Function(core.SchemaParams{Error: "Expected a function value"})

		_, err := schema.Parse("not a function")
		assert.Error(t, err)
	})

	t.Run("nil handling without modifiers", func(t *testing.T) {
		schema := Function()

		_, err := schema.Parse(nil)
		assert.Error(t, err)
	})
}

func TestFunction_EdgeCases(t *testing.T) {
	t.Run("nil handling with nilable", func(t *testing.T) {
		schema := Function().Nilable()

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		fn := func() {}
		result, err = schema.Parse(fn)
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("empty context", func(t *testing.T) {
		schema := Function()

		result, err := schema.Parse(func() {})
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("function with various signatures", func(t *testing.T) {
		schema := Function()

		cases := []any{
			func() {},
			func(int) {},
			func() int { return 0 },
			func(int) int { return 0 },
			func(string, bool) (int, error) { return 0, nil },
			func(...int) {},
		}
		for i, fn := range cases {
			result, err := schema.Parse(fn)
			require.NoError(t, err, "case %d should pass", i)
			assert.NotNil(t, result, "case %d should return non-nil", i)
		}
	})

	t.Run("function parameters validation", func(t *testing.T) {
		schema := Function(FunctionParams{Input: nil, Output: nil})
		require.NotNil(t, schema)
		assert.Nil(t, schema.internals.Input)
		assert.Nil(t, schema.internals.Output)
	})
}

func TestFunction_Overwrite(t *testing.T) {
	t.Run("basic function transformation", func(t *testing.T) {
		schema := Function().
			Overwrite(func(fn any) any {
				if f, ok := fn.(func(int) int); ok {
					return func(x int) int { return f(x) + 1 }
				}
				return fn
			})

		orig := func(x int) int { return x * 2 }
		result, err := schema.Parse(orig)
		require.NoError(t, err)

		if got, ok := result.(func(int) int); ok {
			// (5 * 2) + 1 = 11
			assert.Equal(t, 11, got(5))
		} else {
			t.Fatalf("got type %T, want func(int) int", result)
		}
	})

	t.Run("function wrapper transformation", func(t *testing.T) {
		schema := Function().
			Overwrite(func(fn any) any {
				if f, ok := fn.(func(string) string); ok {
					return func(s string) string {
						if s == "" {
							return "default"
						}
						return f(s)
					}
				}
				return fn
			})

		result, err := schema.Parse(strings.ToUpper)
		require.NoError(t, err)

		if got, ok := result.(func(string) string); ok {
			assert.Equal(t, "default", got(""))
			assert.Equal(t, "HELLO", got("hello"))
		} else {
			t.Fatalf("got type %T, want func(string) string", result)
		}
	})

	t.Run("function metadata transformation", func(t *testing.T) {
		counter := func(initial int) func() int {
			n := initial
			return func() int { n++; return n }
		}

		schema := Function().
			Overwrite(func(fn any) any {
				if f, ok := fn.(func() int); ok {
					n := 100
					return func() int {
						f() // advance original
						n++
						return n
					}
				}
				return fn
			})

		result, err := schema.Parse(counter(0))
		require.NoError(t, err)

		if got, ok := result.(func() int); ok {
			assert.Equal(t, 101, got())
			assert.Equal(t, 102, got())
		} else {
			t.Fatalf("got type %T, want func() int", result)
		}
	})

	t.Run("pointer type handling", func(t *testing.T) {
		schema := FunctionPtr().
			Overwrite(func(fn *any) *any {
				if fn == nil {
					dflt := any(func() string { return "default" })
					return &dflt
				}
				if f, ok := (*fn).(func(string) string); ok {
					wrapped := any(func(s string) string {
						return "transformed_" + f(s)
					})
					return &wrapped
				}
				return fn
			})

		orig := any(func(s string) string { return s + "_suffix" })
		result, err := schema.Parse(&orig)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Handle nested pointer structure
		if p1, ok := (*result).(*any); ok {
			if p2, ok := (*p1).(*any); ok {
				if got, ok := (*p2).(func(string) string); ok {
					assert.Equal(t, "transformed_test_suffix", got("test"))
				} else {
					t.Fatalf("got type %T, want func(string) string", *p2)
				}
			} else {
				t.Fatalf("got type %T, want *any", *p1)
			}
		} else {
			t.Fatalf("got type %T, want *any", *result)
		}
	})

	t.Run("type preservation", func(t *testing.T) {
		schema := Function().Overwrite(func(fn any) any { return fn })

		fn := func(x int) int { return x + 1 }
		result, err := schema.Parse(fn)
		require.NoError(t, err)

		if got, ok := result.(func(int) int); ok {
			assert.Equal(t, 6, got(5))
		} else {
			t.Fatalf("got type %T, want func(int) int", result)
		}
	})

	t.Run("chaining with validations", func(t *testing.T) {
		schema := Function().
			Overwrite(func(fn any) any {
				if f, ok := fn.(func(int) int); ok {
					return func(x int) int {
						if r := f(x); r < 0 {
							return 0
						} else {
							return r
						}
					}
				}
				return fn
			}).
			Refine(func(fn any) bool { return fn != nil }, "Function cannot be nil")

		orig := func(x int) int { return x - 10 }
		result, err := schema.Parse(orig)
		require.NoError(t, err)

		if got, ok := result.(func(int) int); ok {
			assert.Equal(t, 0, got(5))  // 5-10=-5, clamped to 0
			assert.Equal(t, 5, got(15)) // 15-10=5
		} else {
			t.Fatalf("got type %T, want func(int) int", result)
		}
	})

	t.Run("error handling preservation", func(t *testing.T) {
		schema := Function().Overwrite(func(fn any) any { return fn })

		_, err := schema.Parse("not a function")
		assert.Error(t, err)

		_, err = schema.Parse(123)
		assert.Error(t, err)
	})

	t.Run("function composition", func(t *testing.T) {
		schema := Function().
			Overwrite(func(fn any) any {
				if f, ok := fn.(func(string) string); ok {
					return func(s string) string { return "prefix_" + f(s) }
				}
				return fn
			})

		result, err := schema.Parse(strings.ToUpper)
		require.NoError(t, err)

		if got, ok := result.(func(string) string); ok {
			assert.Equal(t, "prefix_HELLO", got("hello"))
		} else {
			t.Fatalf("got type %T, want func(string) string", result)
		}
	})
}
