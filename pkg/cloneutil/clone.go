// Package cloneutil centralizes deep-copy behavior for GoZod runtime values
// and schema internals while preserving library-specific semantics.
package cloneutil

import (
	"math/big"
	"reflect"
	"time"

	"github.com/kaptinlin/deepclone"
)

// Clone copies arbitrary values while preserving GoZod's existing semantics for
// runtime defaults and schema internals. It handles time/math-bits values
// explicitly, then delegates general cloning to deepclone.
func Clone(v any) any {
	if v == nil {
		return nil
	}

	if cloned, ok := cloneSpecialValue(reflect.ValueOf(v)); ok {
		return cloned.Interface()
	}
	return deepclone.Clone(v)
}

func cloneSpecialValue(v reflect.Value) (reflect.Value, bool) {
	switch val := v.Interface().(type) {
	case time.Time:
		return reflect.ValueOf(val), true
	case *time.Time:
		if val == nil {
			return reflect.Zero(v.Type()), true
		}
		cloned := *val
		return reflect.ValueOf(&cloned), true
	case big.Int:
		var cloned big.Int
		cloned.Set(&val)
		return reflect.ValueOf(cloned), true
	case *big.Int:
		if val == nil {
			return reflect.Zero(v.Type()), true
		}
		cloned := new(big.Int)
		cloned.Set(val)
		return reflect.ValueOf(cloned), true
	case big.Float:
		var cloned big.Float
		cloned.Copy(&val)
		return reflect.ValueOf(cloned), true
	case *big.Float:
		if val == nil {
			return reflect.Zero(v.Type()), true
		}
		cloned := new(big.Float)
		cloned.Copy(val)
		return reflect.ValueOf(cloned), true
	case big.Rat:
		var cloned big.Rat
		cloned.Set(&val)
		return reflect.ValueOf(cloned), true
	case *big.Rat:
		if val == nil {
			return reflect.Zero(v.Type()), true
		}
		cloned := new(big.Rat)
		cloned.Set(val)
		return reflect.ValueOf(cloned), true
	default:
		return reflect.Value{}, false
	}
}
