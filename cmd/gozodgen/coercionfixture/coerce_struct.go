// Package coercionfixture verifies generated coercion parity.
package coercionfixture

// CoerceStruct exercises generated primitive coercion.
type CoerceStruct struct {
	Name   string  `json:"name" gozod:"required,coerce"`
	Email  string  `json:"email" gozod:"required,coerce,email"`
	Age    int     `json:"age" gozod:"required,coerce,gte=1"`
	Active bool    `json:"active" gozod:"required,coerce"`
	Score  float64 `json:"score" gozod:"required,coerce"`
}
