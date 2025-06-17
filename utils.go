package gozod

import (
	"fmt"
	"math"
	"math/big"
	"math/cmplx"
	"mime/multipart"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cast"
)

// =============================================================================
// TYPE DEFINITIONS
// =============================================================================

// ParsedTypes represents the type of parsed data values
type ParsedTypes string

const (
	ParsedTypeString   ParsedTypes = "string"
	ParsedTypeNumber   ParsedTypes = "number"
	ParsedTypeBigint   ParsedTypes = "bigint"
	ParsedTypeBool     ParsedTypes = "bool"
	ParsedTypeFloat    ParsedTypes = "float"
	ParsedTypeObject   ParsedTypes = "object"
	ParsedTypeFunction ParsedTypes = "function"
	ParsedTypeFile     ParsedTypes = "file"
	ParsedTypeDate     ParsedTypes = "date"
	ParsedTypeArray    ParsedTypes = "array"
	ParsedTypeSlice    ParsedTypes = "slice"
	ParsedTypeMap      ParsedTypes = "map"
	ParsedTypeNaN      ParsedTypes = "nan"
	ParsedTypeNil      ParsedTypes = "nil"
	ParsedTypeComplex  ParsedTypes = "complex"
)

// CachedValue provides lazy evaluation with caching
type CachedValue[T any] struct {
	getter func() T
	value  T
	set    bool
}

func (c *CachedValue[T]) Get() T {
	if !c.set {
		c.value = c.getter()
		c.set = true
		return c.value
	}
	panic("cached value already set")
}

// StructFieldMapping stores mapping information between struct field and schema key
type StructFieldMapping struct {
	FieldName  string // Go struct field name
	SchemaKey  string // Schema key name
	FieldIndex int    // Field index in struct
	IsOptional bool   // Whether field is optional (from omitempty tag)
}

// =============================================================================
// CONSTANTS
// =============================================================================

// AllowsEval checks if eval is allowed in the environment (always false in Go)
var AllowsEval = cached(func() bool {
	return false
})

// NUMBER_FORMAT_RANGES defines numeric format validation ranges
var NUMBER_FORMAT_RANGES = map[string][2]float64{
	"safeint": {-9007199254740991, 9007199254740991},
	"int32":   {-2147483648, 2147483647},
	"uint32":  {0, 4294967295},
	"float32": {-3.4028234663852886e38, 3.4028234663852886e38},
	"float64": {-math.MaxFloat64, math.MaxFloat64},
}

// BIGINT_FORMAT_RANGES defines big integer format validation ranges
var BIGINT_FORMAT_RANGES = map[string][2]int64{
	"int64":  {-9223372036854775808, 9223372036854775807},
	"uint64": {0, 9223372036854775807},
}

// =============================================================================
// CORE UTILITY FUNCTIONS
// =============================================================================

// cached creates a cached value getter
func cached[T any](getter func() T) *CachedValue[T] {
	return &CachedValue[T]{getter: getter}
}

// nullish checks if input is null or undefined (nil in Go context)
func nullish(input interface{}) bool {
	if input == nil {
		return true
	}

	rv := reflect.ValueOf(input)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		return rv.IsNil()
	case reflect.Invalid:
		return true
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128,
		reflect.Array, reflect.String, reflect.Struct, reflect.UnsafePointer:
		return false
	default:
		return false
	}
}

// floatSafeRemainder performs safe floating-point remainder operation
func floatSafeRemainder(val, step float64) float64 {
	valStr := strconv.FormatFloat(val, 'f', -1, 64)
	stepStr := strconv.FormatFloat(step, 'f', -1, 64)

	valParts := strings.Split(valStr, ".")
	stepParts := strings.Split(stepStr, ".")

	valDecCount := 0
	if len(valParts) > 1 {
		valDecCount = len(valParts[1])
	}

	stepDecCount := 0
	if len(stepParts) > 1 {
		stepDecCount = len(stepParts[1])
	}

	decCount := valDecCount
	if stepDecCount > valDecCount {
		decCount = stepDecCount
	}

	multiplier := math.Pow10(decCount)
	valInt := int64(val*multiplier + 0.5)
	stepInt := int64(step*multiplier + 0.5)

	return float64(valInt%stepInt) / multiplier
}

// =============================================================================
// TYPE DETECTION AND PARSING
// =============================================================================

// GetParsedType determines the parsed type of data
func GetParsedType(data interface{}) ParsedTypes {
	if data == nil {
		return ParsedTypeNil
	}

	switch v := data.(type) {
	case string:
		return ParsedTypeString
	case bool:
		return ParsedTypeBool
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return ParsedTypeNumber
	case float32:
		if math.IsNaN(float64(v)) {
			return ParsedTypeNaN
		}
		return ParsedTypeFloat
	case float64:
		if math.IsNaN(v) {
			return ParsedTypeNaN
		}
		return ParsedTypeFloat
	case complex64, complex128:
		return ParsedTypeComplex
	case func():
		return ParsedTypeFunction
	default:
		rv := reflect.ValueOf(data)
		switch rv.Kind() {
		case reflect.Array:
			return ParsedTypeArray
		case reflect.Slice:
			return ParsedTypeSlice
		case reflect.Map:
			if rv.Type().Key().Kind() == reflect.String {
				return ParsedTypeObject
			}
			return ParsedTypeMap
		case reflect.Struct:
			if _, ok := data.(interface{ String() string }); ok {
				return ParsedTypeDate
			}
			return ParsedTypeObject
		case reflect.Chan:
			return ParsedTypeObject
		case reflect.Invalid, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
			reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128,
			reflect.Func, reflect.Interface, reflect.Ptr, reflect.String, reflect.UnsafePointer:
			return ParsedTypeObject
		default:
			return ParsedTypeObject
		}
	}
}

// =============================================================================
// STRING AND FORMATTING UTILITIES
// =============================================================================

// JoinValues joins primitive values with a separator
func JoinValues(array []interface{}, separator string) string {
	if separator == "" {
		separator = "|"
	}

	strs := make([]string, len(array))
	for i, val := range array {
		strs[i] = StringifyPrimitive(val)
	}
	return strings.Join(strs, separator)
}

// StringifyPrimitive converts a primitive value to string representation
func StringifyPrimitive(value interface{}) string {
	if value == nil {
		return "null"
	}

	switch v := value.(type) {
	case string:
		return fmt.Sprintf(`"%s"`, v)
	case int64, uint64:
		return fmt.Sprintf("%v", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// escapeRegex escapes special regex characters
func escapeRegex(str string) string {
	re := regexp.MustCompile(`[.*+?^${}()|[\]\\]`)
	return re.ReplaceAllStringFunc(str, func(match string) string {
		return "\\" + match
	})
}

// =============================================================================
// ORIGIN DETECTION UTILITIES
// =============================================================================

// getSizableOrigin determines origin for sizable values
func getSizableOrigin(input interface{}) string {
	if input == nil {
		return "unknown"
	}

	switch input.(type) {
	case *multipart.FileHeader, *os.File:
		return "file"
	}

	rv := reflect.ValueOf(input)
	switch rv.Kind() {
	case reflect.Map:
		return "map"
	case reflect.Slice:
		typeName := rv.Type().String()
		if strings.Contains(strings.ToLower(typeName), "file") {
			return "file"
		}
		return "set"
	case reflect.Invalid, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.Array,
		reflect.Chan, reflect.Func, reflect.Interface, reflect.Ptr, reflect.String, reflect.Struct, reflect.UnsafePointer:
		return "unknown"
	default:
		return "unknown"
	}
}

// getLengthableOrigin determines origin for values with length
func getLengthableOrigin(input interface{}) string {
	if input == nil {
		return "unknown"
	}

	switch input.(type) {
	case string:
		return "string"
	default:
		rv := reflect.ValueOf(input)
		if rv.Kind() == reflect.Slice {
			return "slice"
		} else if rv.Kind() == reflect.Array {
			return "array"
		}
		return "unknown"
	}
}

// =============================================================================
// COMPARISON AND NUMERIC UTILITIES
// =============================================================================

// deepEqual performs deep equality comparison for literal values
func deepEqual(a, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}

// compareNumeric compares two numeric values and returns -1, 0, or 1
func compareNumeric(a, b interface{}) int {
	// Handle complex number comparisons by magnitude
	aComplex64, aIsComplex64 := a.(complex64)
	aComplex128, aIsComplex128 := a.(complex128)
	bComplex64, bIsComplex64 := b.(complex64)
	bComplex128, bIsComplex128 := b.(complex128)

	if aIsComplex64 || aIsComplex128 || bIsComplex64 || bIsComplex128 {
		var aMagnitude, bMagnitude float64
		switch {
		case aIsComplex64:
			aMagnitude = float64(cmplx.Abs(complex128(aComplex64)))
		case aIsComplex128:
			aMagnitude = cmplx.Abs(aComplex128)
		default:
			aMagnitude = toFloat64(a)
		}

		switch {
		case bIsComplex64:
			bMagnitude = float64(cmplx.Abs(complex128(bComplex64)))
		case bIsComplex128:
			bMagnitude = cmplx.Abs(bComplex128)
		default:
			bMagnitude = toFloat64(b)
		}

		switch {
		case aMagnitude < bMagnitude:
			return -1
		case aMagnitude > bMagnitude:
			return 1
		default:
			return 0
		}
	}

	// Handle big.Int comparisons
	aBigInt, aIsBigInt := a.(*big.Int)
	bBigInt, bIsBigInt := b.(*big.Int)

	switch {
	case aIsBigInt && bIsBigInt:
		return aBigInt.Cmp(bBigInt)
	case aIsBigInt:
		bFloat := toFloat64(b)
		if bFloat == float64(int64(bFloat)) {
			bBigIntConverted := big.NewInt(int64(bFloat))
			return aBigInt.Cmp(bBigIntConverted)
		} else {
			aFloat, _ := aBigInt.Float64()
			switch {
			case aFloat < bFloat:
				return -1
			case aFloat > bFloat:
				return 1
			default:
				return 0
			}
		}
	case bIsBigInt:
		aFloat := toFloat64(a)
		if aFloat == float64(int64(aFloat)) {
			aBigIntConverted := big.NewInt(int64(aFloat))
			return aBigIntConverted.Cmp(bBigInt)
		} else {
			bFloat, _ := bBigInt.Float64()
			switch {
			case aFloat < bFloat:
				return -1
			case aFloat > bFloat:
				return 1
			default:
				return 0
			}
		}
	default:
		aFloat := toFloat64(a)
		bFloat := toFloat64(b)
		switch {
		case aFloat < bFloat:
			return -1
		case aFloat > bFloat:
			return 1
		default:
			return 0
		}
	}
}

// toFloat64 converts various numeric types to float64
func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case *big.Int:
		f, _ := val.Float64()
		return f
	case int:
		return float64(val)
	case int8:
		return float64(val)
	case int16:
		return float64(val)
	case int32:
		return float64(val)
	case int64:
		return float64(val)
	case uint:
		return float64(val)
	case uint8:
		return float64(val)
	case uint16:
		return float64(val)
	case uint32:
		return float64(val)
	case uint64:
		return float64(val)
	case float32:
		return float64(val)
	case float64:
		return val
	default:
		return 0
	}
}

// =============================================================================
// LENGTH AND SIZE UTILITIES
// =============================================================================

// hasLength checks if a value has a length property
func hasLength(v interface{}) bool {
	if v == nil {
		return false
	}

	switch v.(type) {
	case string:
		return true
	default:
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Array, reflect.Slice:
			return true
		case reflect.Invalid, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
			reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128,
			reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.String, reflect.Struct, reflect.UnsafePointer:
			return false
		default:
			return false
		}
	}
}

// getLength gets the length of a value
func getLength(v interface{}) int {
	if v == nil {
		return 0
	}

	switch val := v.(type) {
	case string:
		return len(val)
	default:
		rv := reflect.ValueOf(val)
		switch rv.Kind() {
		case reflect.Array, reflect.Slice:
			return rv.Len()
		case reflect.Invalid, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
			reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128,
			reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.String, reflect.Struct, reflect.UnsafePointer:
			return 0
		default:
			return 0
		}
	}
}

// hasSize checks if a value has a size property
func hasSize(v interface{}) bool {
	if v == nil {
		return false
	}

	switch v.(type) {
	case *multipart.FileHeader, *os.File:
		return true
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Map, reflect.Slice, reflect.Array:
		return true
	case reflect.Invalid, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128,
		reflect.Chan, reflect.Func, reflect.Interface, reflect.Ptr, reflect.String, reflect.Struct, reflect.UnsafePointer:
		return false
	default:
		return false
	}
}

// getSize gets the size of a value
func getSize(v interface{}) int {
	if v == nil {
		return 0
	}

	switch file := v.(type) {
	case *multipart.FileHeader:
		return int(file.Size)
	case *os.File:
		if stat, err := file.Stat(); err == nil {
			return int(stat.Size())
		}
		return 0
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Map, reflect.Slice, reflect.Array:
		return rv.Len()
	case reflect.Invalid, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128,
		reflect.Chan, reflect.Func, reflect.Interface, reflect.Ptr, reflect.String, reflect.Struct, reflect.UnsafePointer:
		return 0
	default:
		return 0
	}
}

// =============================================================================
// TYPE CONVERSION UTILITIES
// =============================================================================

// convertToMap converts various object types to map[string]interface{}
func convertToMap(value interface{}) map[string]interface{} {
	if value == nil {
		return nil
	}

	if m, ok := value.(map[string]interface{}); ok {
		return m
	}

	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() == reflect.Struct {
		result := make(map[string]interface{})
		rt := rv.Type()

		for i := 0; i < rv.NumField(); i++ {
			field := rt.Field(i)
			fieldValue := rv.Field(i)

			if !field.IsExported() {
				continue
			}

			fieldName := field.Name
			if jsonTag := field.Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
				fieldName = jsonTag
			}

			result[fieldName] = fieldValue.Interface()
		}

		return result
	}

	if rv.Kind() == reflect.Map && rv.Type().Key().Kind() == reflect.String {
		result := make(map[string]interface{})
		for _, key := range rv.MapKeys() {
			result[key.String()] = rv.MapIndex(key).Interface()
		}
		return result
	}

	return nil
}

// convertToMapWithTags converts a struct to map using tag priority
func convertToMapWithTags(value interface{}, tagPriority []string) map[string]interface{} {
	if value == nil {
		return nil
	}

	v := reflect.ValueOf(value)
	t := reflect.TypeOf(value)

	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
		t = t.Elem()
	}

	if v.Kind() != reflect.Struct {
		return convertToMap(value)
	}

	result := make(map[string]interface{})

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		if !fieldType.IsExported() {
			continue
		}

		var key string
		for _, tagName := range tagPriority {
			if tag := fieldType.Tag.Get(tagName); tag != "" {
				if parts := strings.Split(tag, ","); len(parts) > 0 && parts[0] != "" && parts[0] != "-" {
					key = parts[0]
					break
				}
			}
		}

		if key == "" {
			key = strings.ToLower(fieldType.Name)
		}

		if field.CanInterface() {
			result[key] = field.Interface()
		}
	}

	return result
}

// convertToSlice converts a value to []interface{} if it's a slice or array type
func convertToSlice(value interface{}) ([]interface{}, bool) {
	if value == nil {
		return []interface{}{}, true
	}

	rv := reflect.ValueOf(value)

	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return nil, false
	}

	length := rv.Len()
	slice := make([]interface{}, length)
	for i := 0; i < length; i++ {
		slice[i] = rv.Index(i).Interface()
	}

	return slice, true
}

// =============================================================================
// TYPE COERCION UTILITIES
// =============================================================================

// coerceToString attempts to convert a value to string using spf13/cast
func coerceToString(value interface{}) (string, bool) {
	if value == nil {
		return "", false
	}

	str, err := cast.ToStringE(value)
	if err != nil {
		return "", false
	}

	return str, true
}

// coerceToBool attempts to convert a value to boolean using spf13/cast
func coerceToBool(value interface{}) (bool, bool) {
	if value == nil {
		return false, false
	}

	if reflect.ValueOf(value).Kind() == reflect.Ptr && reflect.ValueOf(value).IsNil() {
		return false, false
	}

	if str, ok := value.(string); ok {
		str = strings.TrimSpace(strings.ToLower(str))
		if str == "" {
			return false, true
		}
		switch str {
		case "true", "1", "yes", "on", "enabled", "y":
			return true, true
		case "false", "0", "no", "off", "disabled", "n":
			return false, true
		default:
			return false, false
		}
	}

	val, err := cast.ToBoolE(value)
	if err != nil {
		return false, false
	}

	return val, true
}

// coerceToFloat64 attempts to convert a value to float64 using spf13/cast
func coerceToFloat64(value interface{}) (float64, bool) {
	if value == nil {
		return 0, false
	}

	if reflect.ValueOf(value).Kind() == reflect.Ptr && reflect.ValueOf(value).IsNil() {
		return 0, false
	}

	val, err := cast.ToFloat64E(value)
	if err != nil {
		return 0, false
	}

	return val, true
}

// coerceToBigInt attempts to coerce a value to big.Int using spf13/cast
func coerceToBigInt(value interface{}) (*big.Int, bool) {
	if value == nil {
		return nil, false
	}

	switch v := value.(type) {
	case *big.Int:
		return new(big.Int).Set(v), true
	case bool:
		if v {
			return big.NewInt(1), true
		}
		return big.NewInt(0), true
	case string:
		v = strings.TrimSpace(v)
		if v == "" {
			return nil, false
		}
		for _, char := range v {
			if char == '.' {
				return nil, false
			}
		}
		if result, ok := new(big.Int).SetString(v, 10); ok {
			return result, true
		}
		return nil, false
	case float32, float64:
		floatVal := toFloat64(v)
		if floatVal == float64(int64(floatVal)) {
			return big.NewInt(int64(floatVal)), true
		}
		return nil, false
	default:
		if intVal, err := cast.ToInt64E(value); err == nil {
			return big.NewInt(intVal), true
		}

		if strVal, err := cast.ToStringE(value); err == nil {
			if result, ok := new(big.Int).SetString(strVal, 10); ok {
				return result, true
			}
		}

		return nil, false
	}
}

// coerceToMap try to coerce input to map[interface{}]interface{}
func coerceToMap(input any) (map[interface{}]interface{}, bool) {
	// Use reflection to handle various map types
	val := reflect.ValueOf(input)
	if val.Kind() == reflect.Map {
		result := make(map[interface{}]interface{})
		for _, key := range val.MapKeys() {
			result[key.Interface()] = val.MapIndex(key).Interface()
		}
		return result, true
	}

	// Handle struct to map conversion
	if IsStructType(input) {
		if mapped := convertToMapWithTags(input, []string{"json", "yaml"}); mapped != nil {
			// Convert map[string]interface{} to map[interface{}]interface{}
			result := make(map[interface{}]interface{})
			for k, v := range mapped {
				result[k] = v
			}
			return result, true
		}
	}

	// Try using convertToMap to handle other object types
	if mapped := convertToMap(input); mapped != nil {
		result := make(map[interface{}]interface{})
		for k, v := range mapped {
			result[k] = v
		}
		return result, true
	}

	return nil, false
}

// coerceToObject try to coerce input to map[string]interface{}
func coerceToObject(input any) (map[string]interface{}, bool) {
	// Use existing convertToMap function
	if result := convertToMap(input); result != nil {
		return result, true
	}
	return nil, false
}

// =============================================================================
// STRUCT REFLECTION UTILITIES
// =============================================================================

// getStructFieldMappings analyzes a struct type and returns field mappings
func getStructFieldMappings(structType reflect.Type) map[string]StructFieldMapping {
	mappings := make(map[string]StructFieldMapping)

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		if !field.IsExported() {
			continue
		}

		schemaKey := field.Name
		isOptional := false

		if tag := field.Tag.Get("gozod"); tag != "" {
			parts := strings.Split(tag, ",")
			if parts[0] != "" && parts[0] != "-" {
				schemaKey = parts[0]
			}
			for _, part := range parts[1:] {
				if strings.TrimSpace(part) == "omitempty" {
					isOptional = true
				}
			}
		} else if tag := field.Tag.Get("json"); tag != "" {
			parts := strings.Split(tag, ",")
			if parts[0] != "" && parts[0] != "-" {
				schemaKey = parts[0]
			}
			for _, part := range parts[1:] {
				if strings.TrimSpace(part) == "omitempty" {
					isOptional = true
				}
			}
		}

		schemaKey = strings.ToLower(schemaKey)

		mappings[schemaKey] = StructFieldMapping{
			FieldName:  field.Name,
			SchemaKey:  schemaKey,
			FieldIndex: i,
			IsOptional: isOptional,
		}
	}

	return mappings
}

// mapToStruct converts validated map data back to specified struct type
func mapToStruct(data map[string]interface{}, structType reflect.Type) (interface{}, error) {
	structValue := reflect.New(structType).Elem()
	mappings := getStructFieldMappings(structType)

	for schemaKey, fieldMapping := range mappings {
		if value, exists := data[schemaKey]; exists {
			field := structValue.Field(fieldMapping.FieldIndex)
			if field.CanSet() {
				if err := setFieldValue(field, value); err != nil {
					return nil, fmt.Errorf("%w %s: %w", ErrFailedToSetField, fieldMapping.FieldName, err)
				}
			}
		}
	}

	return structValue.Interface(), nil
}

// setFieldValue sets a struct field value with type conversion support
func setFieldValue(field reflect.Value, value interface{}) error {
	if value == nil {
		if field.Kind() == reflect.Ptr {
			field.Set(reflect.Zero(field.Type()))
			return nil
		}
		return ErrCannotAssignNilToNonPointer
	}

	valueType := reflect.TypeOf(value)
	fieldType := field.Type()

	if valueType.AssignableTo(fieldType) {
		field.Set(reflect.ValueOf(value))
		return nil
	}

	if valueType.ConvertibleTo(fieldType) {
		field.Set(reflect.ValueOf(value).Convert(fieldType))
		return nil
	}

	if fieldType.Kind() == reflect.Ptr {
		if valueType.AssignableTo(fieldType.Elem()) {
			ptrValue := reflect.New(fieldType.Elem())
			ptrValue.Elem().Set(reflect.ValueOf(value))
			field.Set(ptrValue)
			return nil
		}
		if valueType.ConvertibleTo(fieldType.Elem()) {
			ptrValue := reflect.New(fieldType.Elem())
			ptrValue.Elem().Set(reflect.ValueOf(value).Convert(fieldType.Elem()))
			field.Set(ptrValue)
			return nil
		}
	}

	return fmt.Errorf("%w %v to %v", ErrCannotConvertType, valueType, fieldType)
}

// IsStructType checks if the given value is a struct type
func IsStructType(value interface{}) bool {
	if value == nil {
		return false
	}

	t := reflect.TypeOf(value)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return t.Kind() == reflect.Struct
}

// =============================================================================
// VALIDATION HELPER UTILITIES
// =============================================================================

// newBaseZodTypeInternals creates a basic ZodTypeInternals with common initialization
func newBaseZodTypeInternals(typeName string) ZodTypeInternals {
	return ZodTypeInternals{
		Version: Version,
		Type:    typeName,
		Checks:  make([]ZodCheck, 0),
		Values:  make(map[interface{}]struct{}),
		Bag:     make(map[string]interface{}),
	}
}

// createErrorMap creates a ZodErrorMap from various error types
func createErrorMap(errorParam interface{}) *ZodErrorMap {
	if errorParam == nil {
		return nil
	}

	switch err := errorParam.(type) {
	case string:
		errorMsg := err
		errorMap := ZodErrorMap(func(ZodRawIssue) string {
			return errorMsg
		})
		return &errorMap
	case ZodErrorMap:
		return &err
	case *ZodErrorMap:
		return err
	case func(ZodRawIssue) string:
		errorMap := ZodErrorMap(err)
		return &errorMap
	default:
		return nil
	}
}

// convertRawIssuesToIssues converts a slice of ZodRawIssue to a slice of ZodIssue
func convertRawIssuesToIssues(rawIssues []ZodRawIssue, ctx *ParseContext) []ZodIssue {
	issues := make([]ZodIssue, len(rawIssues))
	for i, raw := range rawIssues {
		issues[i] = FinalizeIssue(raw, ctx, GetConfig())
	}
	return issues
}

// isOptionalField checks if a schema represents an optional field
func isOptionalField(schema ZodType[any, any]) bool {
	if _, ok := schema.(*ZodOptional[ZodType[any, any]]); ok {
		return true
	}

	internals := schema.GetInternals()
	if internals != nil && internals.Optional {
		return true
	}

	if internals != nil && internals.OptIn == "optional" {
		return true
	}

	return false
}

// createMissingKeyIssue creates an issue for missing required keys
func createMissingKeyIssue(key string, options ...func(*ZodRawIssue)) ZodRawIssue {
	issue := ZodRawIssue{
		Code:    "required",
		Message: fmt.Sprintf("Required field '%s' is missing", key),
		Path:    []interface{}{key},
	}

	for _, option := range options {
		option(&issue)
	}

	return issue
}

// processFieldPath adds a field name to the beginning of an issue path
func processFieldPath(issues []ZodRawIssue, fieldName string) []ZodRawIssue {
	for i := range issues {
		newPath := make([]interface{}, len(issues[i].Path)+1)
		newPath[0] = fieldName
		copy(newPath[1:], issues[i].Path)
		issues[i].Path = newPath
	}
	return issues
}

// createUnrecognizedKeysIssue creates an issue for unrecognized keys
func createUnrecognizedKeysIssue(keys []string, options ...func(*ZodRawIssue)) ZodRawIssue {
	pathElements := make([]interface{}, len(keys))
	for i, key := range keys {
		pathElements[i] = key
	}

	issue := ZodRawIssue{
		Code:    string(UnrecognizedKeys),
		Message: fmt.Sprintf("Unrecognized keys: %v", keys),
		Path:    pathElements,
	}

	for _, option := range options {
		option(&issue)
	}

	return issue
}

// validateSliceElements validates each element of a slice with the given schema
func validateSliceElements(slice []interface{}, elementSchema ZodType[any, any], basePath []interface{}, ctx *ParseContext, originalValue interface{}) (interface{}, []ZodRawIssue) {
	length := len(slice)
	validatedSlice := make([]interface{}, length)
	var issues []ZodRawIssue
	hasElementChanges := false

	for i, element := range slice {
		elementPayload := &ParsePayload{
			Value:  element,
			Path:   append(basePath, i),
			Issues: make([]ZodRawIssue, 0),
		}

		elementResult := elementSchema.GetInternals().Parse(elementPayload, ctx)

		if len(elementResult.Issues) > 0 {
			for _, issue := range elementResult.Issues {
				if len(issue.Path) == 0 {
					issue.Path = elementPayload.Path
				}
				issues = append(issues, issue)
			}
		} else {
			validatedSlice[i] = elementResult.Value
			if !deepEqual(elementResult.Value, element) {
				hasElementChanges = true
			}
		}
	}

	if len(issues) > 0 {
		return nil, issues
	}

	if !hasElementChanges {
		rv := reflect.ValueOf(originalValue)
		if rv.Kind() == reflect.Slice {
			return originalValue, nil
		} else if rv.Kind() == reflect.Array {
			elemType := rv.Type().Elem()
			newSlice := reflect.MakeSlice(reflect.SliceOf(elemType), length, length)
			for i := 0; i < length; i++ {
				newSlice.Index(i).Set(rv.Index(i))
			}
			return newSlice.Interface(), nil
		}
	}

	return validatedSlice, nil
}

// runChecksOnValue runs validation checks on a value
func runChecksOnValue(value interface{}, checks []ZodCheck, payload *ParsePayload, ctx *ParseContext) {
	for _, check := range checks {
		if check != nil {
			if checkInternals := check.GetZod(); checkInternals != nil {
				checkPayload := &ParsePayload{
					Value:  value,
					Path:   payload.Path,
					Issues: make([]ZodRawIssue, 0),
				}

				checkInternals.Check(checkPayload)
				if len(checkPayload.Issues) > 0 {
					payload.Issues = append(payload.Issues, checkPayload.Issues...)
				}

				if len(checkPayload.Issues) > 0 && checkInternals.Def.Abort {
					break
				}
			}
		}
	}
}

// =============================================================================
// OBJECT UTILITIES
// =============================================================================

// getObjectKeys extracts keys from an object (map or struct)
func getObjectKeys(input interface{}) []interface{} {
	if mapped := convertToMap(input); mapped != nil {
		keys := make([]interface{}, 0, len(mapped))
		for key := range mapped {
			keys = append(keys, key)
		}
		return keys
	}

	v := reflect.ValueOf(input)

	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	var keys []interface{}

	switch v.Kind() {
	case reflect.Map:
		mapKeys := v.MapKeys()
		for _, key := range mapKeys {
			keys = append(keys, key.Interface())
		}
	case reflect.Struct:
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			if field.IsExported() {
				keys = append(keys, field.Name)
			}
		}
	case reflect.Invalid, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.Array,
		reflect.Chan, reflect.Func, reflect.Interface, reflect.Ptr, reflect.Slice, reflect.String, reflect.UnsafePointer:
		// These types don't have extractable keys
	}

	return keys
}

// getObjectValue gets a value from an object by key
func getObjectValue(input interface{}, key interface{}) (interface{}, bool) {
	if mapped := convertToMap(input); mapped != nil {
		if keyStr, ok := key.(string); ok {
			if value, exists := mapped[keyStr]; exists {
				return value, true
			}
		}
		return nil, false
	}

	v := reflect.ValueOf(input)

	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil, false
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Map:
		keyValue := reflect.ValueOf(key)
		mapValue := v.MapIndex(keyValue)
		if !mapValue.IsValid() {
			return nil, false
		}
		return mapValue.Interface(), true
	case reflect.Struct:
		if keyStr, ok := key.(string); ok {
			fieldValue := v.FieldByName(keyStr)
			if !fieldValue.IsValid() || !fieldValue.CanInterface() {
				return nil, false
			}
			return fieldValue.Interface(), true
		}
		return nil, false
	case reflect.Invalid, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.Array,
		reflect.Chan, reflect.Func, reflect.Interface, reflect.Ptr, reflect.Slice, reflect.String, reflect.UnsafePointer:
		return nil, false
	default:
		return nil, false
	}
}

// createInvalidKeyIssue creates an invalid key issue
func createInvalidKeyIssue(input interface{}, key interface{}, keyIssues []ZodRawIssue, options ...func(*ZodRawIssue)) ZodRawIssue {
	issue := NewRawIssue(
		string(InvalidKey),
		input,
		WithOrigin("record"),
		WithPath([]interface{}{key}),
	)

	if issue.Properties == nil {
		issue.Properties = make(map[string]interface{})
	}
	issue.Properties["issues"] = keyIssues

	for _, opt := range options {
		opt(&issue)
	}

	return issue
}

// =============================================================================
// MISSING UTILITY FUNCTIONS
// =============================================================================

// getNumericOrigin determines the origin string for a numeric value
func getNumericOrigin(value interface{}) string {
	switch value.(type) {
	case *big.Int:
		return "bigint"
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return "number"
	case float32, float64:
		return "number"
	default:
		return "number" // Default to number for Go
	}
}

// coerceToValueSet attempts to coerce input to match any value in a set
func coerceToValueSet(input interface{}, validValues map[interface{}]struct{}) (interface{}, bool) {
	// Direct match
	if _, exists := validValues[input]; exists {
		return input, true
	}

	// Try string conversion
	if coercedStr, ok := coerceToString(input); ok {
		if _, exists := validValues[coercedStr]; exists {
			return coercedStr, true
		}
	}

	// Try boolean conversion
	if coercedBool, ok := coerceToBool(input); ok {
		if _, exists := validValues[coercedBool]; exists {
			return coercedBool, true
		}
	}

	// Try numeric conversion
	if coercedFloat, ok := coerceToFloat64(input); ok {
		// Try both float64 and int versions
		if _, exists := validValues[coercedFloat]; exists {
			return coercedFloat, true
		}

		// Try as integer if it's a whole number
		if coercedFloat == float64(int64(coercedFloat)) {
			intVal := int64(coercedFloat)
			if _, exists := validValues[intVal]; exists {
				return intVal, true
			}
			// Also try as int
			if intVal <= int64(int(^uint(0)>>1)) && intVal >= int64(-int(^uint(0)>>1)-1) {
				intValInt := int(intVal)
				if _, exists := validValues[intValInt]; exists {
					return intValInt, true
				}
			}
		}
	}

	return input, false
}

// coerceToLiteralValue attempts to coerce input to match a literal value
func coerceToLiteralValue(input interface{}, allowedValues []interface{}) (interface{}, bool) {
	// Direct match check first
	for _, allowedValue := range allowedValues {
		if deepEqual(input, allowedValue) {
			return allowedValue, true
		}
	}

	// Try coercion for each allowed value
	for _, allowedValue := range allowedValues {
		switch expectedType := allowedValue.(type) {
		case string:
			// Try to coerce to string
			if coercedStr, ok := coerceToString(input); ok && coercedStr == expectedType {
				return expectedType, true
			}
		case bool:
			// Try to coerce to boolean
			if coercedBool, ok := coerceToBool(input); ok && coercedBool == expectedType {
				return expectedType, true
			}
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			// Try to coerce to number
			if coercedFloat, ok := coerceToFloat64(input); ok {
				// Compare with the expected numeric value
				expectedFloat := toFloat64(expectedType)
				if coercedFloat == expectedFloat {
					return expectedType, true
				}
			}
		}
	}

	return input, false
}

// =============================================================================
// EXTENSION UTILITIES
// =============================================================================

// Extend adds properties to a schema
func Extend(schema interface{}, shape map[string]interface{}) interface{} {
	newShape := make(map[string]interface{})

	if schemaMap, ok := schema.(map[string]interface{}); ok {
		if originalShape, exists := schemaMap["shape"].(map[string]interface{}); exists {
			for k, v := range originalShape {
				newShape[k] = v
			}
		}
	}

	for k, v := range shape {
		newShape[k] = v
	}

	return newShape
}

// Merge combines two schemas
func Merge(a, b interface{}) interface{} {
	newShape := make(map[string]interface{})

	if aMap, ok := a.(map[string]interface{}); ok {
		if aShape, exists := aMap["shape"].(map[string]interface{}); exists {
			for k, v := range aShape {
				newShape[k] = v
			}
		}
	}

	if bMap, ok := b.(map[string]interface{}); ok {
		if bShape, exists := bMap["shape"].(map[string]interface{}); exists {
			for k, v := range bShape {
				newShape[k] = v
			}
		}
	}

	return newShape
}

// Partial makes all fields optional
func Partial(class interface{}, schema interface{}, mask map[string]interface{}) interface{} {
	newShape := make(map[string]interface{})

	if schemaMap, ok := schema.(map[string]interface{}); ok {
		if oldShape, exists := schemaMap["shape"].(map[string]interface{}); exists {
			if mask != nil {
				for key, enabled := range mask {
					if _, exists := oldShape[key]; !exists {
						panic(fmt.Sprintf("Unrecognized key: \"%s\"", key))
					}
					if enabled != nil && enabled != false {
						newShape[key] = map[string]interface{}{
							"type":      "optional",
							"innerType": oldShape[key],
						}
					} else {
						newShape[key] = oldShape[key]
					}
				}
			} else {
				for key, value := range oldShape {
					newShape[key] = map[string]interface{}{
						"type":      "optional",
						"innerType": value,
					}
				}
			}
		}
	}

	return newShape
}
