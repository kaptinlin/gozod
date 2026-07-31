// Package anyfixture verifies explicit dynamic field support in generated schemas.
package anyfixture

// Values contains explicit dynamic fields accepted by the generator.
type Values struct {
	Any   any         `json:"any" gozod:"required"`
	Empty interface{} `json:"empty" gozod:"required"`
}
