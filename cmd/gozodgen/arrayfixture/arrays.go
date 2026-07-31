// Package arrayfixture verifies generated fixed-array behavior.
package arrayfixture

// Arrays distinguishes fixed arrays from dynamically sized slices.
type Arrays struct {
	Fixed           [3]string  `json:"fixed" gozod:"required"`
	Dynamic         []string   `json:"dynamic" gozod:"required"`
	DefaultValue    [2]string  `json:"default_value" gozod:"default=[\"left\",\"right\"]"`
	DefaultPointer  *[2]string `json:"default_pointer" gozod:"default=[\"north\",\"south\"]"`
	PrefaultValue   [2]string  `json:"prefault_value" gozod:"prefault=[\"up\",\"down\"]"`
	PrefaultPointer *[2]string `json:"prefault_pointer" gozod:"prefault=[\"east\",\"west\"]"`
}

// CompositeArrays exercises nested and pointer-backed fixed arrays.
type CompositeArrays struct {
	Nested         [2][3]string `json:"nested" gozod:"required"`
	Pointer        *[2]string   `json:"pointer" gozod:"required"`
	DefaultNested  [2][3]string `json:"default_nested" gozod:"default=[[\"a\",\"b\",\"c\"],[\"d\",\"e\",\"f\"]]"`
	PrefaultNested [2][3]string `json:"prefault_nested" gozod:"prefault=[[\"g\",\"h\",\"i\"],[\"j\",\"k\",\"l\"]]"`
}
