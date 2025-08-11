package testdata

// Product represents a complex structure with various field types
type Product struct {
	ID       string   `json:"id" gozod:"required,uuid"`
	SKU      string   `json:"sku" gozod:"required,regex=^[A-Z0-9\\-]+$"`
	Name     string   `json:"name" gozod:"required,min=1,max=200"`
	Price    float64  `json:"price" gozod:"required,gt=0.0"`
	Currency string   `json:"currency" gozod:"required,enum=USD EUR GBP"`
	Tags     []string `json:"tags" gozod:"min=0,max=10"`
	Active   *bool    `json:"active" gozod:"default=true"`
}

// Category represents a category with optional fields
type Category struct {
	ID          string  `json:"id" gozod:"required,uuid"`
	Name        string  `json:"name" gozod:"required,min=2"`
	Description *string `json:"description" gozod:"max=500"`
	ParentID    *string `json:"parent_id" gozod:"uuid"`
}
