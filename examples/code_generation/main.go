// Package main demonstrates explicit GoZod schema generation from struct tags.
//
// Usage:
//  1. Add //go:generate gozodgen directive to your Go files
//  2. Define structs with gozod struct tags
//  3. Run: go generate ./...
//  4. Generated *_gen.go files contain optimized Schema() methods
//  5. Use generated methods explicitly, for example User{}.Schema()
package main

import (
	"fmt"
	"time"

	"github.com/kaptinlin/gozod"
)

//go:generate go run ../../cmd/gozodgen -suffix=_schema.go

// User demonstrates struct tag validation with code generation
type User struct {
	ID        string    `json:"id" gozod:"required,uuid"`
	Name      string    `json:"name" gozod:"required,min=2,max=100"`
	Email     string    `json:"email" gozod:"required,email"`
	Age       int       `json:"age" gozod:"required,min=18,max=120"`
	Status    string    `json:"status" gozod:"required,enum=active inactive,default=active"`
	CreatedAt time.Time `json:"created_at" gozod:"required,time"`
}

// Product demonstrates complex validation rules
type Product struct {
	ID       string  `json:"id" gozod:"required,uuid"`
	SKU      string  `json:"sku" gozod:"required,min=3,max=50,regex=^[A-Z0-9\\-]+$"`
	Name     string  `json:"name" gozod:"required,min=1,max=200"`
	Price    float64 `json:"price" gozod:"required,gt=0.0"`
	Currency string  `json:"currency" gozod:"required,enum=USD EUR GBP JPY"`
	InStock  bool    `json:"in_stock" gozod:"default=true"`
}

func main() {
	fmt.Println("🚀 GoZod Code Generation Example")
	fmt.Println("=================================")

	// Example: User validation
	fmt.Println("\n👤 User Validation:")

	user := User{
		ID:        "550e8400-e29b-41d4-a716-446655440000",
		Name:      "Alice Johnson",
		Email:     "alice@example.com",
		Age:       28,
		Status:    "active",
		CreatedAt: time.Now(),
	}

	// This example stays runnable before code generation. After running
	// go generate, use User{}.Schema() to call the generated schema explicitly.
	schema := gozod.MustFromStruct[User]()
	result, err := schema.Parse(user)

	if err != nil {
		fmt.Printf("❌ Validation failed: %v\n", err)
	} else {
		fmt.Printf("✅ User validated: %s <%s>\n", result.Name, result.Email)
		fmt.Printf("   Status: %s, Age: %d\n", result.Status, result.Age)
	}

	// Example: Product validation
	fmt.Println("\n🛍️ Product Validation:")

	product := Product{
		ID:       "550e8400-e29b-41d4-a716-446655440001",
		SKU:      "LAPTOP-PRO-2024",
		Name:     "Professional Laptop",
		Price:    1299.99,
		Currency: "USD",
		InStock:  true,
	}

	productSchema := gozod.MustFromStruct[Product]()
	productResult, err := productSchema.Parse(product)

	if err != nil {
		fmt.Printf("❌ Validation failed: %v\n", err)
	} else {
		fmt.Printf("✅ Product validated: %s (%s)\n", productResult.Name, productResult.SKU)
		fmt.Printf("   Price: $%.2f %s\n", productResult.Price, productResult.Currency)
	}

	fmt.Println("\n💡 Code Generation Benefits:")
	fmt.Println("   • Explicit pre-built Schema() methods in generated files")
	fmt.Println("   • No runtime tag reflection in generated schema construction")
	fmt.Println("   • Generated Schema() methods in *_gen.go files")
	fmt.Println("   • Call generated schemas directly, for example User{}.Schema()")

	fmt.Println("\n📚 Next Steps:")
	fmt.Println("   1. Run: go generate ./...")
	fmt.Println("   2. Check generated *_gen.go files")
	fmt.Println("   3. Call the generated Schema() method where you want generated validation")
}
