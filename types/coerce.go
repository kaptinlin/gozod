package types

import (
	"fmt"
	"math/big"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/pkg/coerce"
)

// =============================================================================
// COERCED TYPE WRAPPERS
// =============================================================================

// ZodCoercedMap provides type-safe coercive map validation
type ZodCoercedMap struct {
	*ZodMap
	keyType   core.ZodType[any, any]
	valueType core.ZodType[any, any]
}

// ZodCoercedRecord provides type-safe coercive record validation
type ZodCoercedRecord struct {
	*ZodRecord
	keyType   core.ZodType[any, any]
	valueType core.ZodType[any, any]
}

// ZodCoercedStruct provides coercive struct validation
type ZodCoercedStruct struct {
	*ZodStruct
}

// ZodCoercedObject provides coercive object validation
type ZodCoercedObject struct {
	*ZodObject
}

// =============================================================================
// COERCIBLE INTERFACE IMPLEMENTATIONS
// =============================================================================

// Parse with coercion for Map
func (z *ZodCoercedMap) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	coercedInput := input
	if coerced, err := coerce.ToMap(input); err == nil {
		coercedInput = coerced
	}

	result, err := z.ZodMap.Parse(coercedInput, ctx...)
	if err != nil {
		return nil, err
	}

	return convertToTypeSafeMap(result, z.keyType, z.valueType), nil
}

func (z *ZodCoercedMap) Coerce(input any) (output any, success bool) {
	if coercible, ok := any(z.ZodMap).(core.Coercible); ok {
		return coercible.Coerce(input)
	}
	return input, false
}

func (z *ZodCoercedMap) GetInternals() *core.ZodTypeInternals {
	return z.ZodMap.GetInternals()
}

func (z *ZodCoercedMap) GetZod() *ZodMapInternals {
	return z.ZodMap.GetZod()
}

// Parse with coercion for Record
func (z *ZodCoercedRecord) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	// Handle the simple case: if the input is already a map with correct types, use it directly
	if resultMap, ok := input.(map[string]int); ok && getTypeName(z.keyType) == "string" && getTypeName(z.valueType) == "int" {
		// Still need to validate using the underlying Record
		_, err := z.ZodRecord.Parse(input, ctx...)
		if err != nil {
			return nil, err
		}
		return resultMap, nil
	}

	// Convert input to generic map format for processing
	var genericMap map[any]any
	switch v := input.(type) {
	case map[any]any:
		genericMap = v
	case map[string]any:
		genericMap = make(map[any]any)
		for k, val := range v {
			genericMap[k] = val
		}
	case map[string]string:
		genericMap = make(map[any]any)
		for k, val := range v {
			genericMap[k] = val
		}
	case map[string]int:
		genericMap = make(map[any]any)
		for k, val := range v {
			genericMap[k] = val
		}
	case map[int]string:
		genericMap = make(map[any]any)
		for k, val := range v {
			genericMap[k] = val
		}
	default:
		// Try struct-to-map conversion for struct types
		if coerced, err := coerce.ToMap(input); err == nil {
			genericMap = coerced
		} else {
			// If we can't convert, fall back to regular Record parsing
			return z.ZodRecord.Parse(input, ctx...)
		}
	}

	// Apply coercion to each key-value pair
	coercedMap := make(map[any]any)
	for k, v := range genericMap {
		// Coerce and validate key
		coercedKey, keyErr := z.keyType.Parse(k, ctx...)
		if keyErr != nil {
			return nil, keyErr
		}

		// Coerce and validate value
		coercedValue, valueErr := z.valueType.Parse(v, ctx...)
		if valueErr != nil {
			return nil, valueErr
		}

		coercedMap[coercedKey] = coercedValue
	}

	// Convert to type-safe map format
	return convertToTypeSafeMap(coercedMap, z.keyType, z.valueType), nil
}

func (z *ZodCoercedRecord) Coerce(input any) (output any, success bool) {
	if coercible, ok := any(z.ZodRecord).(core.Coercible); ok {
		return coercible.Coerce(input)
	}
	return input, false
}

func (z *ZodCoercedRecord) GetInternals() *core.ZodTypeInternals {
	return z.ZodRecord.GetInternals()
}

func (z *ZodCoercedRecord) GetZod() *ZodRecordInternals {
	return z.ZodRecord.GetZod()
}

// Parse with coercion for Struct
func (z *ZodCoercedStruct) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	coercedInput := input
	if coerced, err := coerce.ToObject(input); err == nil {
		coercedInput = coerced
	}

	return z.ZodStruct.Parse(coercedInput, ctx...)
}

func (z *ZodCoercedStruct) Coerce(input any) (output any, success bool) {
	if coercible, ok := any(z.ZodStruct).(core.Coercible); ok {
		return coercible.Coerce(input)
	}
	return input, false
}

func (z *ZodCoercedStruct) GetInternals() *core.ZodTypeInternals {
	return z.ZodStruct.GetInternals()
}

func (z *ZodCoercedStruct) GetZod() *ZodStructInternals {
	return z.ZodStruct.GetZod()
}

// Parse with coercion for Object
func (z *ZodCoercedObject) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	coercedInput := input
	if coerced, err := coerce.ToObject(input); err == nil {
		coercedInput = coerced
	}

	return z.ZodObject.Parse(coercedInput, ctx...)
}

func (z *ZodCoercedObject) Coerce(input any) (output any, success bool) {
	if coercible, ok := any(z.ZodObject).(core.Coercible); ok {
		return coercible.Coerce(input)
	}
	return input, false
}

func (z *ZodCoercedObject) GetInternals() *core.ZodTypeInternals {
	return z.ZodObject.GetInternals()
}

func (z *ZodCoercedObject) GetZod() *ZodObjectInternals {
	if internals := z.ZodObject.GetInternals(); internals != nil {
		return &ZodObjectInternals{
			ZodTypeInternals: *internals,
		}
	}
	return nil
}

// =============================================================================
// COERCE NAMESPACE
// =============================================================================

// CoerceNamespace provides coercive validation factory functions
type CoerceNamespace struct{}

// Global coerce namespace instance
var Coerce = &CoerceNamespace{}

// =============================================================================
// PRIMITIVE TYPE COERCERS
// =============================================================================

// String creates coercive string schema
func (c *CoerceNamespace) String(params ...any) *ZodString {
	schema := String(params...)
	c.enableCoercionForType(schema)
	return schema
}

// Number creates coercive number schema (float64)
func (c *CoerceNamespace) Number(params ...any) *ZodFloat[float64] {
	schema := Float64(params...)
	c.enableCoercionForType(schema)
	return schema
}

// Bool creates coercive boolean schema
func (c *CoerceNamespace) Bool(params ...any) *ZodBool {
	schema := Bool(params...)
	c.enableCoercionForType(schema)
	return schema
}

// BigInt creates coercive big integer schema
func (c *CoerceNamespace) BigInt(params ...any) *ZodBigInt {
	schema := BigInt(params...)
	c.enableCoercionForType(schema)
	return schema
}

// Complex64 creates coercive complex64 schema
func (c *CoerceNamespace) Complex64(params ...any) *ZodComplex[complex64] {
	schema := Complex64(params...)
	c.enableCoercionForType(schema)
	return schema
}

// Complex128 creates coercive complex128 schema
func (c *CoerceNamespace) Complex128(params ...any) *ZodComplex[complex128] {
	schema := Complex128(params...)
	c.enableCoercionForType(schema)
	return schema
}

// =============================================================================
// COLLECTION TYPE COERCERS
// =============================================================================

// Map creates coercive Map schema with type-safe wrapper
func (c *CoerceNamespace) Map(keySchema, valueSchema core.ZodType[any, any], params ...any) *ZodCoercedMap {
	schema := Map(keySchema, valueSchema, params...)
	c.enableCoercionForType(schema)
	c.enableCoercionForType(keySchema)
	c.enableCoercionForType(valueSchema)

	return &ZodCoercedMap{
		ZodMap:    schema,
		keyType:   keySchema,
		valueType: valueSchema,
	}
}

// Record creates coercive Record schema with type-safe wrapper
func (c *CoerceNamespace) Record(keyType, valueType core.ZodType[any, any], params ...any) *ZodCoercedRecord {
	schema := Record(keyType, valueType, params...)
	c.enableCoercionForType(schema)
	c.enableCoercionForType(keyType)
	c.enableCoercionForType(valueType)

	return &ZodCoercedRecord{
		ZodRecord: schema,
		keyType:   keyType,
		valueType: valueType,
	}
}

// Object creates coercive Object schema with type-safe wrapper
func (c *CoerceNamespace) Object(shape core.ObjectSchema, params ...any) *ZodCoercedObject {
	schema := Object(shape, params...)
	c.enableCoercionForType(schema)

	for _, fieldSchema := range shape {
		c.enableCoercionForType(fieldSchema)
	}

	return &ZodCoercedObject{
		ZodObject: schema,
	}
}

// Struct creates coercive Struct schema with type-safe wrapper
func (c *CoerceNamespace) Struct(shape core.StructSchema, params ...any) *ZodCoercedStruct {
	schema := Struct(shape, params...)
	c.enableCoercionForType(schema)

	for _, fieldSchema := range shape {
		c.enableCoercionForType(fieldSchema)
	}

	return &ZodCoercedStruct{
		ZodStruct: schema,
	}
}

// =============================================================================
// TYPE CONVERSION UTILITIES
// =============================================================================

// convertToTypeSafeMap converts generic map to type-safe map based on key/value types
func convertToTypeSafeMap(input any, keyType, valueType core.ZodType[any, any]) any {
	var genericMap map[any]any

	switch v := input.(type) {
	case map[any]any:
		genericMap = v
	case map[string]any:
		genericMap = make(map[any]any)
		for k, val := range v {
			genericMap[k] = val
		}
	case map[string]string:
		genericMap = make(map[any]any)
		for k, val := range v {
			genericMap[k] = val
		}
	case map[string]int:
		genericMap = make(map[any]any)
		for k, val := range v {
			genericMap[k] = val
		}
	case map[int]string:
		genericMap = make(map[any]any)
		for k, val := range v {
			genericMap[k] = val
		}
	default:
		return input
	}

	// If input is already empty, return empty map of correct type
	if len(genericMap) == 0 {
		keyTypeName := getTypeName(keyType)
		valueTypeName := getTypeName(valueType)

		if keyTypeName == "string" && valueTypeName == "int" {
			return make(map[string]int)
		}
		if keyTypeName == "string" && valueTypeName == "string" {
			return make(map[string]string)
		}
		// Return input as fallback for other types
		return input
	}

	keyTypeName := getTypeName(keyType)
	valueTypeName := getTypeName(valueType)

	// Handle most common case: map[string]T
	if keyTypeName == "string" {
		switch valueTypeName {
		case "string":
			// Convert all keys to string; keep/convert their values to string
			result := make(map[string]string)
			for k, v := range genericMap {
				keyStr := fmt.Sprint(k)
				if valueStr, ok := v.(string); ok {
					result[keyStr] = valueStr
				} else {
					// Use fmt.Sprint to safely convert any value
					result[keyStr] = fmt.Sprint(v)
				}
			}
			return result
		case "int":
			// Convert values to int; support strings and other numeric types
			result := make(map[string]int)
			for k, v := range genericMap {
				keyStr := fmt.Sprint(k)
				if valueInt, ok := v.(int); ok {
					result[keyStr] = valueInt
					continue
				}
				if parsed, err := coerce.ToInteger[int](v); err == nil {
					result[keyStr] = parsed
				} else {
					// Fallback to 0 when conversion fails, keep the key in the map
					result[keyStr] = 0
				}
			}
			return result
		case "int8":
			result := make(map[string]int8)
			for k, v := range genericMap {
				if keyStr, ok := k.(string); ok {
					if valueInt, ok := v.(int8); ok {
						result[keyStr] = valueInt
					}
				}
			}
			return result
		case "int16":
			result := make(map[string]int16)
			for k, v := range genericMap {
				if keyStr, ok := k.(string); ok {
					if valueInt, ok := v.(int16); ok {
						result[keyStr] = valueInt
					}
				}
			}
			return result
		case "int32":
			result := make(map[string]int32)
			for k, v := range genericMap {
				if keyStr, ok := k.(string); ok {
					if valueInt, ok := v.(int32); ok {
						result[keyStr] = valueInt
					}
				}
			}
			return result
		case "int64":
			result := make(map[string]int64)
			for k, v := range genericMap {
				if keyStr, ok := k.(string); ok {
					if valueInt, ok := v.(int64); ok {
						result[keyStr] = valueInt
					}
				}
			}
			return result
		case "float32":
			result := make(map[string]float32)
			for k, v := range genericMap {
				if keyStr, ok := k.(string); ok {
					if valueFloat, ok := v.(float32); ok {
						result[keyStr] = valueFloat
					}
				}
			}
			return result
		case "float64":
			result := make(map[string]float64)
			for k, v := range genericMap {
				if keyStr, ok := k.(string); ok {
					if valueFloat, ok := v.(float64); ok {
						result[keyStr] = valueFloat
					}
				}
			}
			return result
		case "bool":
			result := make(map[string]bool)
			for k, v := range genericMap {
				if keyStr, ok := k.(string); ok {
					if valueBool, ok := v.(bool); ok {
						result[keyStr] = valueBool
					}
				}
			}
			return result
		case "complex64":
			result := make(map[string]complex64)
			for k, v := range genericMap {
				if keyStr, ok := k.(string); ok {
					if valueComplex, ok := v.(complex64); ok {
						result[keyStr] = valueComplex
					}
				}
			}
			return result
		case "complex128":
			result := make(map[string]complex128)
			for k, v := range genericMap {
				if keyStr, ok := k.(string); ok {
					if valueComplex, ok := v.(complex128); ok {
						result[keyStr] = valueComplex
					}
				}
			}
			return result
		case "big.Int":
			result := make(map[string]*big.Int)
			for k, v := range genericMap {
				if keyStr, ok := k.(string); ok {
					if valueInt, ok := v.(*big.Int); ok {
						result[keyStr] = valueInt
					}
				}
			}
			return result
		default:
			result := make(map[string]any)
			for k, v := range genericMap {
				if keyStr, ok := k.(string); ok {
					result[keyStr] = v
				}
			}
			return result
		}
	}

	return input
}

// getTypeName extracts type name from schema internals
func getTypeName(schema core.ZodType[any, any]) string {
	internals := schema.GetInternals()
	return internals.Type
}

// enableCoercionForType sets coercion flag for given schema
func (c *CoerceNamespace) enableCoercionForType(schema core.ZodType[any, any]) {
	internals := schema.GetInternals()
	if internals == nil {
		return
	}

	if engine.ShouldCoerce(internals.Bag) {
		return
	}

	if internals.Bag == nil {
		internals.Bag = make(map[string]any)
	}
	internals.Bag["coerce"] = true
}
