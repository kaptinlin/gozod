package testdata

// ValidatorStruct tests all validators (covers generateValidatorChain)
type ValidatorStruct struct {
	// String validators
	Email   string  `json:"email" gozod:"email"`
	URL     string  `json:"url" gozod:"url"`
	UUID    string  `json:"uuid" gozod:"uuid"`
	IPv4    string  `json:"ipv4" gozod:"ipv4"`
	IPv6    string  `json:"ipv6" gozod:"ipv6"`
	Regex   string  `json:"regex" gozod:"regex=^[A-Z]+$"`
	Trim    string  `json:"trim" gozod:"trim"`
	Lower   string  `json:"lower" gozod:"lowercase"`
	Upper   string  `json:"upper" gozod:"uppercase"`
	Nilable *string `json:"nilable" gozod:"nilable"`

	// Numeric validators
	Gt  int `json:"gt" gozod:"gt=0"`
	Gte int `json:"gte" gozod:"gte=0"`
	Lt  int `json:"lt" gozod:"lt=100"`
	Lte int `json:"lte" gozod:"lte=100"`
}

// DefaultStruct tests default values (covers generateDefaultValue)
type DefaultStruct struct {
	Str      string            `json:"str" gozod:"default=hello"`
	Num      int               `json:"num" gozod:"default=42"`
	Float    float64           `json:"float" gozod:"default=3.14"`
	Bool     bool              `json:"bool" gozod:"default=true"`
	Slice    []string          `json:"slice" gozod:"default=[\"a\",\"b\"]"`
	IntSlice []int             `json:"int_slice" gozod:"default=[1,2,3]"`
	Map      map[string]string `json:"map" gozod:"default={\"k\":\"v\"}"`
}

// PrefaultStruct tests prefault values (covers generatePrefaultValue)
type PrefaultStruct struct {
	Str   string            `json:"str" gozod:"prefault=world"`
	Num   int               `json:"num" gozod:"prefault=100"`
	Slice []string          `json:"slice" gozod:"prefault=[\"x\",\"y\"]"`
	Map   map[string]string `json:"map" gozod:"prefault={\"foo\":\"bar\"}"`
}

// AllTypesStruct tests all Go basic types (covers typesToReflectType)
type AllTypesStruct struct {
	// Integer types
	Int   int   `json:"int" gozod:"required"`
	Int8  int8  `json:"int8" gozod:"required"`
	Int16 int16 `json:"int16" gozod:"required"`
	Int32 int32 `json:"int32" gozod:"required"`
	Int64 int64 `json:"int64" gozod:"required"`

	// Unsigned integer types
	Uint   uint   `json:"uint" gozod:"required"`
	Uint8  uint8  `json:"uint8" gozod:"required"`
	Uint16 uint16 `json:"uint16" gozod:"required"`
	Uint32 uint32 `json:"uint32" gozod:"required"`
	Uint64 uint64 `json:"uint64" gozod:"required"`

	// Float types
	Float32 float32 `json:"float32" gozod:"gt=0"`
	Float64 float64 `json:"float64" gozod:"gt=0"`

	// Complex types
	Complex64  complex64  `json:"complex64" gozod:""`
	Complex128 complex128 `json:"complex128" gozod:""`

	// Boolean
	Bool bool `json:"bool" gozod:"default=false"`

	// Pointer types
	PtrString *string `json:"ptr_string" gozod:"min=1"`
	PtrInt    *int    `json:"ptr_int" gozod:"gt=0"`
	PtrBool   *bool   `json:"ptr_bool" gozod:"default=true"`
}

// FloatDefaultStruct tests float slice defaults
type FloatDefaultStruct struct {
	FloatSlice []float64 `json:"float_slice" gozod:"default=[1.1,2.2]"`
	BoolSlice  []bool    `json:"bool_slice" gozod:"default=[true,false]"`
}

// InterfaceMapStruct tests map[string]interface{} defaults
type InterfaceMapStruct struct {
	Data map[string]interface{} `json:"data" gozod:"default={\"a\":1,\"b\":\"two\",\"c\":true}"`
}

// FloatPrefaultStruct tests float slice prefaults
type FloatPrefaultStruct struct {
	FloatSlice []float64 `json:"float_slice" gozod:"prefault=[3.3,4.4]"`
	BoolSlice  []bool    `json:"bool_slice" gozod:"prefault=[false,true]"`
}

// InterfaceMapPrefaultStruct tests map[string]interface{} prefaults
type InterfaceMapPrefaultStruct struct {
	Data map[string]interface{} `json:"data" gozod:"prefault={\"x\":99,\"y\":\"test\"}"`
}
