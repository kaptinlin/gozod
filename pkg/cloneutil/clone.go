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
// runtime defaults and schema internals. It uses deepclone as the general
// fallback, but handles time/math-bits values explicitly to avoid regressions.
func Clone(v any) any {
	if v == nil {
		return nil
	}
	return cloneValue(reflect.ValueOf(v)).Interface()
}

func cloneValue(v reflect.Value) reflect.Value {
	if !v.IsValid() {
		return v
	}

	if cloned, ok := cloneSpecialValue(v); ok {
		return cloned
	}

	switch v.Kind() {
	case reflect.Interface:
		if v.IsNil() {
			return reflect.Zero(v.Type())
		}
		cloned := cloneValue(v.Elem())
		out := reflect.New(v.Type()).Elem()
		out.Set(cloned)
		return out

	case reflect.Pointer:
		if v.IsNil() {
			return reflect.Zero(v.Type())
		}
		cloned := reflect.New(v.Type().Elem())
		cloned.Elem().Set(cloneValue(v.Elem()))
		return cloned

	case reflect.Slice:
		if v.IsNil() {
			return reflect.Zero(v.Type())
		}
		cloned := reflect.MakeSlice(v.Type(), v.Len(), v.Cap())
		for i := range v.Len() {
			cloned.Index(i).Set(cloneValue(v.Index(i)))
		}
		return cloned

	case reflect.Array:
		cloned := reflect.New(v.Type()).Elem()
		for i := range v.Len() {
			cloned.Index(i).Set(cloneValue(v.Index(i)))
		}
		return cloned

	case reflect.Map:
		if v.IsNil() {
			return reflect.Zero(v.Type())
		}
		cloned := reflect.MakeMapWithSize(v.Type(), v.Len())
		iter := v.MapRange()
		for iter.Next() {
			cloned.SetMapIndex(cloneValue(iter.Key()), cloneValue(iter.Value()))
		}
		return cloned

	case reflect.Struct:
		cloned := reflect.New(v.Type()).Elem()
		cloned.Set(v)
		for i := range v.NumField() {
			field := cloned.Field(i)
			if !field.CanSet() {
				continue
			}
			field.Set(cloneValue(v.Field(i)))
		}
		return cloned

	default:
		return reflect.ValueOf(deepclone.Clone(v.Interface()))
	}
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
