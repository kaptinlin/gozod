package types

import (
	"math/big"
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBigInt_BasicFunctionality(t *testing.T) {
	t.Run("valid big.Int inputs", func(t *testing.T) {
		s := BigInt()

		tests := []struct {
			name  string
			input *big.Int
		}{
			{name: "positive", input: big.NewInt(42)},
			{name: "negative", input: big.NewInt(-123)},
			{name: "zero", input: big.NewInt(0)},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := s.Parse(tt.input)
				require.NoError(t, err)
				assert.Equal(t, tt.input, got)
			})
		}
	})

	t.Run("invalid type inputs", func(t *testing.T) {
		s := BigInt()
		for _, input := range []any{
			"not a bigint", 123, 3.14, []byte{1, 2, 3}, nil,
		} {
			_, err := s.Parse(input)
			assert.Error(t, err, "Parse(%v) = nil error, want error", input)
		}
	})

	t.Run("Parse and MustParse", func(t *testing.T) {
		s := BigInt()
		v := big.NewInt(999)

		got, err := s.Parse(v)
		require.NoError(t, err)
		assert.Equal(t, v, got)

		assert.Equal(t, v, s.MustParse(v))
		assert.Panics(t, func() { s.MustParse("invalid") })
	})
}

func TestBigInt_TypeSafety(t *testing.T) {
	t.Run("BigInt returns *big.Int type", func(t *testing.T) {
		s := BigInt()
		v := big.NewInt(42)
		got, err := s.Parse(v)
		require.NoError(t, err)
		assert.Equal(t, v, got)
		assert.IsType(t, (*big.Int)(nil), got)
	})

	t.Run("BigIntPtr returns **big.Int type", func(t *testing.T) {
		s := BigIntPtr()
		v := big.NewInt(42)
		got, err := s.Parse(v)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, v, *got)
		assert.IsType(t, (**big.Int)(nil), got)
	})

	t.Run("MustParse type safety", func(t *testing.T) {
		v := big.NewInt(123)

		got := BigInt().MustParse(v)
		assert.IsType(t, (*big.Int)(nil), got)
		assert.Equal(t, v, got)

		ptrSchema := BigIntPtr().Nilable().Overwrite(func(bi **big.Int) **big.Int {
			if bi == nil || *bi == nil {
				return nil
			}
			abs := new(big.Int).Abs(*bi)
			return &abs
		})
		ptrGot := ptrSchema.MustParse(v)
		assert.IsType(t, (**big.Int)(nil), ptrGot)
		require.NotNil(t, ptrGot)
		assert.Equal(t, v, *ptrGot)
	})
}

func TestBigInt_Modifiers(t *testing.T) {
	t.Run("Optional always returns **big.Int", func(t *testing.T) {
		s := BigInt().Optional()
		v := big.NewInt(42)
		got, err := s.Parse(v)
		require.NoError(t, err)
		assert.IsType(t, (**big.Int)(nil), got)
		require.NotNil(t, got)
		assert.Equal(t, v, *got)
	})

	t.Run("Nilable always returns **big.Int", func(t *testing.T) {
		s := BigInt().Nilable()
		got, err := s.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("Default preserves current type", func(t *testing.T) {
		v := big.NewInt(100)
		_ = BigInt().Default(v)
		_ = BigIntPtr().Default(v)
	})

	t.Run("Prefault preserves current type", func(t *testing.T) {
		v := big.NewInt(50)
		_ = BigInt().Prefault(v)
		_ = BigIntPtr().Prefault(v)
	})

	t.Run("Default priority over Prefault", func(t *testing.T) {
		def := big.NewInt(100)
		s := BigIntPtr().Min(big.NewInt(150)).Default(def).Prefault(big.NewInt(200))
		got, err := s.Parse(nil)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, def, *got)
	})

	t.Run("Default short-circuit bypasses validation", func(t *testing.T) {
		def := big.NewInt(5)
		s := BigIntPtr().Min(big.NewInt(10)).Default(def)
		got, err := s.Parse(nil)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, def, *got)
	})

	t.Run("Prefault only triggers on nil input", func(t *testing.T) {
		pf := big.NewInt(100)
		s := BigInt().Min(big.NewInt(50)).Prefault(pf)

		got, err := s.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, pf, got)

		_, err = s.Parse(big.NewInt(10))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Too small: expected bigint to be at least 50")
	})

	t.Run("Prefault goes through full validation", func(t *testing.T) {
		s := BigIntPtr().Min(big.NewInt(10)).Prefault(big.NewInt(5))
		_, err := s.Parse(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Too small: expected bigint to be at least 10")
	})

	t.Run("Prefault error handling", func(t *testing.T) {
		s := BigInt().Min(big.NewInt(10)).Prefault(big.NewInt(5))
		_, err := s.Parse(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Too small: expected bigint to be at least 10")
	})

	t.Run("BigIntPtr Prefault only on nil input", func(t *testing.T) {
		pf := big.NewInt(100)
		s := BigIntPtr().Min(big.NewInt(50)).Prefault(pf)

		got, err := s.Parse(nil)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, pf, *got)

		_, err = s.Parse(big.NewInt(10))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Too small: expected bigint to be at least 50")
	})
}

func TestBigInt_Validations(t *testing.T) {
	t.Run("Min validation", func(t *testing.T) {
		s := BigInt().Min(big.NewInt(10))
		tests := []struct {
			name    string
			input   int64
			wantErr bool
		}{
			{name: "above minimum", input: 15},
			{name: "at minimum", input: 10},
			{name: "below minimum", input: 5, wantErr: true},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := s.Parse(big.NewInt(tt.input))
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					require.NoError(t, err)
					assert.Equal(t, big.NewInt(tt.input), got)
				}
			})
		}
	})

	t.Run("Max validation", func(t *testing.T) {
		s := BigInt().Max(big.NewInt(100))
		tests := []struct {
			name    string
			input   int64
			wantErr bool
		}{
			{name: "below maximum", input: 50},
			{name: "at maximum", input: 100},
			{name: "above maximum", input: 150, wantErr: true},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := s.Parse(big.NewInt(tt.input))
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					require.NoError(t, err)
					assert.Equal(t, big.NewInt(tt.input), got)
				}
			})
		}
	})

	t.Run("Positive validation", func(t *testing.T) {
		s := BigInt().Positive()
		tests := []struct {
			name    string
			input   int64
			wantErr bool
		}{
			{name: "positive", input: 42},
			{name: "zero", input: 0, wantErr: true},
			{name: "negative", input: -1, wantErr: true},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := s.Parse(big.NewInt(tt.input))
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					require.NoError(t, err)
					assert.Equal(t, big.NewInt(tt.input), got)
				}
			})
		}
	})

	t.Run("Negative validation", func(t *testing.T) {
		s := BigInt().Negative()
		tests := []struct {
			name    string
			input   int64
			wantErr bool
		}{
			{name: "negative", input: -42},
			{name: "zero", input: 0, wantErr: true},
			{name: "positive", input: 1, wantErr: true},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := s.Parse(big.NewInt(tt.input))
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					require.NoError(t, err)
					assert.Equal(t, big.NewInt(tt.input), got)
				}
			})
		}
	})
}

func TestBigInt_Coercion(t *testing.T) {
	t.Run("basic coercion", func(t *testing.T) {
		s := CoercedBigInt()
		tests := []struct {
			name  string
			input any
			want  string
		}{
			{name: "string", input: "42", want: "42"},
			{name: "int", input: int(42), want: "42"},
			{name: "int64", input: int64(84), want: "84"},
			{name: "uint", input: uint(100), want: "100"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := s.Parse(tt.input)
				require.NoError(t, err)
				assert.Equal(t, tt.want, got.String())
			})
		}
	})

	t.Run("large number coercion", func(t *testing.T) {
		s := CoercedBigInt()
		want := "123456789012345678901234567890"
		got, err := s.Parse(want)
		require.NoError(t, err)
		assert.Equal(t, want, got.String())
	})

	t.Run("coercion with validation", func(t *testing.T) {
		s := CoercedBigInt().Min(big.NewInt(5))

		got, err := s.Parse("10")
		require.NoError(t, err)
		assert.Equal(t, "10", got.String())

		_, err = s.Parse("3")
		assert.Error(t, err)
	})
}

func TestBigInt_Chaining(t *testing.T) {
	t.Run("chain multiple validations", func(t *testing.T) {
		s := BigInt().Min(big.NewInt(10)).Max(big.NewInt(100)).Positive()

		got, err := s.Parse(big.NewInt(50))
		require.NoError(t, err)
		assert.Equal(t, big.NewInt(50), got)

		_, err = s.Parse(big.NewInt(5))
		assert.Error(t, err)

		_, err = s.Parse(big.NewInt(150))
		assert.Error(t, err)
	})
}

func TestBigInt_Transform(t *testing.T) {
	t.Run("basic transform", func(t *testing.T) {
		tr := BigInt().Transform(func(v *big.Int, _ *core.RefinementContext) (any, error) {
			return v.String(), nil
		})
		got, err := tr.Parse(big.NewInt(42))
		require.NoError(t, err)
		assert.Equal(t, "42", got)
	})
}

func TestBigInt_Overwrite(t *testing.T) {
	t.Run("basic transformations", func(t *testing.T) {
		s := BigInt().Overwrite(func(bi *big.Int) *big.Int {
			return new(big.Int).Abs(bi)
		})

		got, err := s.Parse(big.NewInt(42))
		require.NoError(t, err)
		assert.Equal(t, int64(42), got.Int64())

		got, err = s.Parse(big.NewInt(-42))
		require.NoError(t, err)
		assert.Equal(t, int64(42), got.Int64())
	})

	t.Run("arithmetic transformations", func(t *testing.T) {
		double := BigInt().Overwrite(func(bi *big.Int) *big.Int {
			return new(big.Int).Mul(bi, big.NewInt(2))
		})
		got, err := double.Parse(big.NewInt(21))
		require.NoError(t, err)
		assert.Equal(t, int64(42), got.Int64())

		addTen := BigInt().Overwrite(func(bi *big.Int) *big.Int {
			return new(big.Int).Add(bi, big.NewInt(10))
		})
		got, err = addTen.Parse(big.NewInt(5))
		require.NoError(t, err)
		assert.Equal(t, int64(15), got.Int64())
	})

	t.Run("modular arithmetic", func(t *testing.T) {
		s := BigInt().Overwrite(func(bi *big.Int) *big.Int {
			return new(big.Int).Mod(bi, big.NewInt(10))
		})

		got, err := s.Parse(big.NewInt(123))
		require.NoError(t, err)
		assert.Equal(t, int64(3), got.Int64())

		got, err = s.Parse(big.NewInt(-17))
		require.NoError(t, err)
		assert.Equal(t, int64(3), got.Int64())
	})

	t.Run("power and root operations", func(t *testing.T) {
		square := BigInt().Overwrite(func(bi *big.Int) *big.Int {
			return new(big.Int).Mul(bi, bi)
		})
		got, err := square.Parse(big.NewInt(7))
		require.NoError(t, err)
		assert.Equal(t, int64(49), got.Int64())

		cube := BigInt().Overwrite(func(bi *big.Int) *big.Int {
			return new(big.Int).Exp(bi, big.NewInt(3), nil)
		})
		got, err = cube.Parse(big.NewInt(4))
		require.NoError(t, err)
		assert.Equal(t, int64(64), got.Int64())
	})

	t.Run("conditional transformations", func(t *testing.T) {
		clamp := BigInt().Overwrite(func(bi *big.Int) *big.Int {
			if bi.Cmp(big.NewInt(0)) < 0 {
				return big.NewInt(0)
			}
			if bi.Cmp(big.NewInt(100)) > 0 {
				return big.NewInt(100)
			}
			return new(big.Int).Set(bi)
		})

		got, err := clamp.Parse(big.NewInt(-10))
		require.NoError(t, err)
		assert.Equal(t, int64(0), got.Int64())

		got, err = clamp.Parse(big.NewInt(50))
		require.NoError(t, err)
		assert.Equal(t, int64(50), got.Int64())

		got, err = clamp.Parse(big.NewInt(150))
		require.NoError(t, err)
		assert.Equal(t, int64(100), got.Int64())
	})

	t.Run("chaining with validations", func(t *testing.T) {
		s := BigInt().
			Overwrite(func(bi *big.Int) *big.Int { return new(big.Int).Mul(bi, big.NewInt(2)) }).
			Min(big.NewInt(10))

		got, err := s.Parse(big.NewInt(7))
		require.NoError(t, err)
		assert.Equal(t, int64(14), got.Int64())

		_, err = s.Parse(big.NewInt(2))
		assert.Error(t, err)
	})

	t.Run("multiple overwrite calls", func(t *testing.T) {
		s := BigInt().
			Overwrite(func(bi *big.Int) *big.Int { return new(big.Int).Abs(bi) }).
			Overwrite(func(bi *big.Int) *big.Int { return new(big.Int).Mul(bi, big.NewInt(3)) }).
			Overwrite(func(bi *big.Int) *big.Int { return new(big.Int).Add(bi, big.NewInt(1)) })

		// -5 -> 5 -> 15 -> 16
		got, err := s.Parse(big.NewInt(-5))
		require.NoError(t, err)
		assert.Equal(t, int64(16), got.Int64())

		// 3 -> 3 -> 9 -> 10
		got, err = s.Parse(big.NewInt(3))
		require.NoError(t, err)
		assert.Equal(t, int64(10), got.Int64())
	})

	t.Run("large number handling", func(t *testing.T) {
		large := new(big.Int)
		large.SetString("999999999999999999999999999999", 10)
		s := BigInt().Overwrite(func(bi *big.Int) *big.Int {
			return new(big.Int).Add(bi, large)
		})

		input := new(big.Int)
		input.SetString("123456789012345678901234567890", 10)
		got, err := s.Parse(input)
		require.NoError(t, err)

		want := new(big.Int)
		want.SetString("1123456789012345678901234567889", 10)
		assert.Equal(t, 0, got.Cmp(want))
	})

	t.Run("type preservation", func(t *testing.T) {
		base := BigInt()
		abs := base.Overwrite(func(bi *big.Int) *big.Int {
			return new(big.Int).Abs(bi)
		})
		v := big.NewInt(-42)

		got1, err := base.Parse(v)
		require.NoError(t, err)
		got2, err := abs.Parse(v)
		require.NoError(t, err)

		assert.IsType(t, (*big.Int)(nil), got1)
		assert.IsType(t, (*big.Int)(nil), got2)
	})

	t.Run("pointer type handling", func(t *testing.T) {
		s := BigIntPtr().Nilable().Overwrite(func(bi **big.Int) **big.Int {
			if bi == nil || *bi == nil {
				return nil
			}
			a := new(big.Int).Abs(*bi)
			return &a
		})

		got, err := s.Parse(big.NewInt(-42))
		require.NoError(t, err)
		require.NotNil(t, got)
		require.NotNil(t, *got)
		assert.Equal(t, int64(42), (*got).Int64())

		got, err = s.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("coerced bigint overwrite", func(t *testing.T) {
		s := CoercedBigInt().Overwrite(func(bi *big.Int) *big.Int {
			return new(big.Int).Mul(bi, bi)
		})

		got, err := s.Parse("7")
		require.NoError(t, err)
		assert.Equal(t, int64(49), got.Int64())

		got, err = s.Parse(6)
		require.NoError(t, err)
		assert.Equal(t, int64(36), got.Int64())

		got, err = s.Parse(5.0)
		require.NoError(t, err)
		assert.Equal(t, int64(25), got.Int64())
	})

	t.Run("error handling", func(t *testing.T) {
		s := BigInt().Overwrite(func(bi *big.Int) *big.Int {
			return new(big.Int).Abs(bi)
		})
		for _, input := range []any{"not a number", 3.14, nil, []int{1, 2, 3}} {
			_, err := s.Parse(input)
			assert.Error(t, err, "Parse(%v) = nil error, want error", input)
		}
	})

	t.Run("immutability", func(t *testing.T) {
		orig := big.NewInt(42)
		cp := new(big.Int).Set(orig)

		s := BigInt().Overwrite(func(bi *big.Int) *big.Int {
			return new(big.Int).Neg(bi)
		})
		got, err := s.Parse(orig)
		require.NoError(t, err)
		assert.Equal(t, int64(-42), got.Int64())
		assert.Equal(t, 0, orig.Cmp(cp))
	})
}

func TestBigInt_NonOptional(t *testing.T) {
	s := BigInt().NonOptional()

	v := big.NewInt(123)
	got, err := s.Parse(v)
	require.NoError(t, err)
	assert.Equal(t, v, got)
	assert.IsType(t, (*big.Int)(nil), got)

	_, err = s.Parse(nil)
	assert.Error(t, err)
	var zErr *issues.ZodError
	if issues.IsZodError(err, &zErr) {
		assert.Equal(t, core.InvalidType, zErr.Issues[0].Code)
		assert.Equal(t, core.ZodTypeNonOptional, zErr.Issues[0].Expected)
	}

	chain := BigInt().Optional().NonOptional()
	_, err = chain.Parse(nil)
	assert.Error(t, err)

	obj := Object(map[string]core.ZodSchema{
		"id": BigInt().Optional().NonOptional(),
	})
	_, err = obj.Parse(map[string]any{"id": big.NewInt(999)})
	require.NoError(t, err)
	_, err = obj.Parse(map[string]any{"id": nil})
	assert.Error(t, err)

	ps := BigIntPtr().NonOptional()
	vv := big.NewInt(456)
	got2, err := ps.Parse(&vv)
	require.NoError(t, err)
	assert.Equal(t, vv, got2)
	_, err = ps.Parse(nil)
	assert.Error(t, err)
}

func TestBigInt_StrictParse(t *testing.T) {
	t.Run("basic functionality", func(t *testing.T) {
		s := BigInt()

		v := big.NewInt(12345)
		got, err := s.StrictParse(v)
		require.NoError(t, err)
		assert.Equal(t, v, got)
		assert.IsType(t, (*big.Int)(nil), got)

		neg, err := s.StrictParse(big.NewInt(-9876))
		require.NoError(t, err)
		assert.Equal(t, big.NewInt(-9876), neg)

		zero, err := s.StrictParse(big.NewInt(0))
		require.NoError(t, err)
		assert.Equal(t, big.NewInt(0), zero)
	})

	t.Run("with validation constraints", func(t *testing.T) {
		s := BigInt().Refine(func(b *big.Int) bool {
			return b.Cmp(big.NewInt(100)) >= 0
		}, "Must be at least 100")

		got, err := s.StrictParse(big.NewInt(150))
		require.NoError(t, err)
		assert.Equal(t, big.NewInt(150), got)

		_, err = s.StrictParse(big.NewInt(50))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Must be at least 100")
	})

	t.Run("with pointer types", func(t *testing.T) {
		s := BigIntPtr()
		v := big.NewInt(999)
		got, err := s.StrictParse(&v)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, v, *got)
		assert.IsType(t, (**big.Int)(nil), got)
	})

	t.Run("with default values", func(t *testing.T) {
		def := big.NewInt(42)
		s := BigIntPtr().Default(def)
		var nilPtr **big.Int

		got, err := s.StrictParse(nilPtr)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, def, *got)
	})

	t.Run("with prefault values", func(t *testing.T) {
		pf := big.NewInt(1000)
		s := BigIntPtr().Refine(func(b **big.Int) bool {
			return b != nil && *b != nil && (*b).Cmp(big.NewInt(500)) >= 0
		}, "Must be at least 500").Prefault(pf)

		small := big.NewInt(100)
		_, err := s.StrictParse(&small)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Must be at least 500")

		got, err := s.StrictParse(nil)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, pf, *got)
	})

	t.Run("large numbers", func(t *testing.T) {
		s := BigInt()

		large := new(big.Int)
		large.SetString("123456789012345678901234567890", 10)
		got, err := s.StrictParse(large)
		require.NoError(t, err)
		assert.Equal(t, large, got)

		negLarge := new(big.Int)
		negLarge.SetString("-987654321098765432109876543210", 10)
		got, err = s.StrictParse(negLarge)
		require.NoError(t, err)
		assert.Equal(t, negLarge, got)
	})
}

func TestBigInt_MustStrictParse(t *testing.T) {
	t.Run("basic functionality", func(t *testing.T) {
		s := BigInt()

		got := s.MustStrictParse(big.NewInt(54321))
		assert.Equal(t, big.NewInt(54321), got)
		assert.IsType(t, (*big.Int)(nil), got)

		assert.Equal(t, big.NewInt(0), s.MustStrictParse(big.NewInt(0)))
		assert.Equal(t, big.NewInt(-11111), s.MustStrictParse(big.NewInt(-11111)))
	})

	t.Run("panic behavior", func(t *testing.T) {
		s := BigInt().Refine(func(b *big.Int) bool {
			return b.Cmp(big.NewInt(1000)) >= 0
		}, "Must be at least 1000")

		assert.Panics(t, func() { s.MustStrictParse(big.NewInt(500)) })
	})

	t.Run("with pointer types", func(t *testing.T) {
		s := BigIntPtr()
		v := big.NewInt(77777)
		got := s.MustStrictParse(&v)
		require.NotNil(t, got)
		assert.Equal(t, v, *got)
		assert.IsType(t, (**big.Int)(nil), got)
	})

	t.Run("with default values", func(t *testing.T) {
		def := big.NewInt(88888)
		s := BigIntPtr().Default(def)
		var nilPtr **big.Int

		got := s.MustStrictParse(nilPtr)
		require.NotNil(t, got)
		assert.Equal(t, def, *got)
	})

	t.Run("large numbers", func(t *testing.T) {
		s := BigInt()

		large := new(big.Int)
		large.SetString("999888777666555444333222111000", 10)
		assert.Equal(t, large, s.MustStrictParse(large))

		negLarge := new(big.Int)
		negLarge.SetString("-111222333444555666777888999000", 10)
		assert.Equal(t, negLarge, s.MustStrictParse(negLarge))
	})
}
