// Package presencefixture verifies generated struct-field presence parity.
package presencefixture

// Presence exercises presence independently from value, pointer, and fallback behavior.
type Presence struct {
	RequiredValue           string  `json:"required_value" gozod:"required,min=2"`
	OptionalValue           string  `json:"optional_value" gozod:"min=2"`
	RequiredPointer         *string `json:"required_pointer" gozod:"required,min=2"`
	OptionalPointer         *string `json:"optional_pointer" gozod:"min=2"`
	RequiredNilableValue    string  `json:"required_nilable_value" gozod:"required,nilable,min=2"`
	OptionalNilableValue    string  `json:"optional_nilable_value" gozod:"nilable,min=2"`
	RequiredNilablePointer  *string `json:"required_nilable_pointer" gozod:"required,nilable,min=2"`
	OptionalNilablePointer  *string `json:"optional_nilable_pointer" gozod:"nilable,min=2"`
	RequiredDefaultValue    string  `json:"required_default_value" gozod:"required,default=required"`
	OptionalDefaultValue    string  `json:"optional_default_value" gozod:"default=optional"`
	RequiredDefaultPointer  *string `json:"required_default_pointer" gozod:"required,default=required"`
	OptionalDefaultPointer  *string `json:"optional_default_pointer" gozod:"default=optional"`
	RequiredPrefaultValue   string  `json:"required_prefault_value" gozod:"required,prefault=required,min=2"`
	OptionalPrefaultValue   string  `json:"optional_prefault_value" gozod:"prefault=optional,min=2"`
	RequiredPrefaultPointer *string `json:"required_prefault_pointer" gozod:"required,prefault=required,min=2"`
	OptionalPrefaultPointer *string `json:"optional_prefault_pointer" gozod:"prefault=optional,min=2"`
}
