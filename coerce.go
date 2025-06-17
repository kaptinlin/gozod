package gozod

import (
	"math/big"
)

// =============================================================================
// COERCED TYPE WRAPPERS
// =============================================================================

// ZodCoercedMap provides type-safe coercive map validation
type ZodCoercedMap struct {
	*ZodMap
	keyType   ZodType[any, any]
	valueType ZodType[any, any]
}

// ZodCoercedRecord provides type-safe coercive record validation
type ZodCoercedRecord struct {
	*ZodRecord
	keyType   ZodType[any, any]
	valueType ZodType[any, any]
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
func (z *ZodCoercedMap) Parse(input any, ctx ...*ParseContext) (any, error) {
	coercedInput := input
	if coerced, ok := coerceToMap(input); ok {
		coercedInput = coerced
	}

	result, err := z.ZodMap.Parse(coercedInput, ctx...)
	if err != nil {
		return nil, err
	}

	return convertToTypeSafeMap(result, z.keyType, z.valueType), nil
}

func (z *ZodCoercedMap) Coerce(input interface{}) (output interface{}, success bool) {
	if coercible, ok := interface{}(z.ZodMap).(Coercible); ok {
		return coercible.Coerce(input)
	}
	return input, false
}

func (z *ZodCoercedMap) GetInternals() *ZodTypeInternals {
	return z.ZodMap.GetInternals()
}

func (z *ZodCoercedMap) GetZod() *ZodMapInternals {
	return z.ZodMap.GetZod()
}

// Parse with coercion for Record
func (z *ZodCoercedRecord) Parse(input any, ctx ...*ParseContext) (any, error) {
	coercedInput := input
	if mapped := convertToMap(input); mapped != nil {
		coercedInput = make(map[interface{}]interface{})
		for k, v := range mapped {
			coercedInput.(map[interface{}]interface{})[k] = v
		}
	}

	result, err := z.ZodRecord.Parse(coercedInput, ctx...)
	if err != nil {
		return nil, err
	}

	return convertToTypeSafeMap(result, z.keyType, z.valueType), nil
}

func (z *ZodCoercedRecord) Coerce(input interface{}) (output interface{}, success bool) {
	if coercible, ok := interface{}(z.ZodRecord).(Coercible); ok {
		return coercible.Coerce(input)
	}
	return input, false
}

func (z *ZodCoercedRecord) GetInternals() *ZodTypeInternals {
	return z.ZodRecord.GetInternals()
}

func (z *ZodCoercedRecord) GetZod() *ZodRecordInternals {
	return z.ZodRecord.GetZod()
}

// Parse with coercion for Struct
func (z *ZodCoercedStruct) Parse(input any, ctx ...*ParseContext) (any, error) {
	coercedInput := input
	if coerced, ok := coerceToObject(input); ok {
		coercedInput = coerced
	}

	return z.ZodStruct.Parse(coercedInput, ctx...)
}

func (z *ZodCoercedStruct) Coerce(input interface{}) (output interface{}, success bool) {
	if coercible, ok := interface{}(z.ZodStruct).(Coercible); ok {
		return coercible.Coerce(input)
	}
	return input, false
}

func (z *ZodCoercedStruct) GetInternals() *ZodTypeInternals {
	return z.ZodStruct.GetInternals()
}

func (z *ZodCoercedStruct) GetZod() *ZodStructInternals {
	return z.ZodStruct.GetZod()
}

// Parse with coercion for Object
func (z *ZodCoercedObject) Parse(input any, ctx ...*ParseContext) (any, error) {
	coercedInput := input
	if coerced, ok := coerceToObject(input); ok {
		coercedInput = coerced
	}

	return z.ZodObject.Parse(coercedInput, ctx...)
}

func (z *ZodCoercedObject) Coerce(input interface{}) (output interface{}, success bool) {
	if coercible, ok := interface{}(z.ZodObject).(Coercible); ok {
		return coercible.Coerce(input)
	}
	return input, false
}

func (z *ZodCoercedObject) GetInternals() *ZodTypeInternals {
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
func (c *CoerceNamespace) String(params ...SchemaParams) *ZodString {
	schema := String(params...)
	c.enableCoercionForType(schema)
	return schema
}

// Number creates coercive number schema (float64)
func (c *CoerceNamespace) Number(params ...SchemaParams) *ZodFloat[float64] {
	schema := Float64(params...)
	c.enableCoercionForType(schema)
	return schema
}

// Bool creates coercive boolean schema
func (c *CoerceNamespace) Bool(params ...SchemaParams) *ZodBool {
	schema := Bool(params...)
	c.enableCoercionForType(schema)
	return schema
}

// BigInt creates coercive big integer schema
func (c *CoerceNamespace) BigInt(params ...SchemaParams) *ZodBigInt {
	schema := BigInt(params...)
	c.enableCoercionForType(schema)
	return schema
}

// Complex64 creates coercive complex64 schema
func (c *CoerceNamespace) Complex64(params ...SchemaParams) *ZodComplex[complex64] {
	schema := Complex64(params...)
	c.enableCoercionForType(schema)
	return schema
}

// Complex128 creates coercive complex128 schema
func (c *CoerceNamespace) Complex128(params ...SchemaParams) *ZodComplex[complex128] {
	schema := Complex128(params...)
	c.enableCoercionForType(schema)
	return schema
}

// =============================================================================
// COLLECTION TYPE COERCERS
// =============================================================================

// Map creates coercive Map schema with type-safe wrapper
func (c *CoerceNamespace) Map(keySchema, valueSchema ZodType[any, any], params ...SchemaParams) *ZodCoercedMap {
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
func (c *CoerceNamespace) Record(keyType, valueType ZodType[any, any], params ...SchemaParams) *ZodCoercedRecord {
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
func (c *CoerceNamespace) Object(shape ObjectSchema, params ...SchemaParams) *ZodCoercedObject {
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
func (c *CoerceNamespace) Struct(shape ObjectSchema, params ...SchemaParams) *ZodCoercedStruct {
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
func convertToTypeSafeMap(input any, keyType, valueType ZodType[any, any]) any {
	var genericMap map[interface{}]interface{}

	switch v := input.(type) {
	case map[interface{}]interface{}:
		genericMap = v
	case map[string]interface{}:
		genericMap = make(map[interface{}]interface{})
		for k, val := range v {
			genericMap[k] = val
		}
	case map[string]string:
		genericMap = make(map[interface{}]interface{})
		for k, val := range v {
			genericMap[k] = val
		}
	case map[string]int:
		genericMap = make(map[interface{}]interface{})
		for k, val := range v {
			genericMap[k] = val
		}
	case map[int]string:
		genericMap = make(map[interface{}]interface{})
		for k, val := range v {
			genericMap[k] = val
		}
	default:
		return input
	}

	keyTypeName := getTypeName(keyType)
	valueTypeName := getTypeName(valueType)

	// Handle most common case: map[string]T
	if keyTypeName == "string" {
		switch valueTypeName {
		case "string":
			result := make(map[string]string)
			for k, v := range genericMap {
				if keyStr, ok := k.(string); ok {
					if valueStr, ok := v.(string); ok {
						result[keyStr] = valueStr
					}
				}
			}
			return result
		case "int":
			result := make(map[string]int)
			for k, v := range genericMap {
				if keyStr, ok := k.(string); ok {
					if valueInt, ok := v.(int); ok {
						result[keyStr] = valueInt
					}
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
			result := make(map[string]interface{})
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
func getTypeName(schema ZodType[any, any]) string {
	internals := schema.GetInternals()
	return internals.Type
}

// enableCoercionForType sets coercion flag for given schema
func (c *CoerceNamespace) enableCoercionForType(schema ZodType[any, any]) {
	internals := schema.GetInternals()
	if internals == nil {
		return
	}

	if shouldCoerce(internals.Bag) {
		return
	}

	if internals.Bag == nil {
		internals.Bag = make(map[string]interface{})
	}
	internals.Bag["coerce"] = true
}
