package testdata

// EmptyStruct has no fields
type EmptyStruct struct{}

// NoTagStruct has fields but no gozod tags
type NoTagStruct struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

// MixedTagStruct has some fields with tags and some without
type MixedTagStruct struct {
	ID          string `json:"id" gozod:"required,uuid"`
	Name        string `json:"name"` // No gozod tag
	Description string // No tags at all
	Active      bool   `gozod:"default=true"` // Only gozod tag
}

// InvalidTagStruct has invalid gozod syntax (for error testing)
type InvalidTagStruct struct {
	Name string `gozod:"invalid=syntax,"`
}
