// Package main demonstrates GoZod code generation for zero-overhead struct tag validation.
//
// Usage:
//  1. Add //go:generate gozodgen directive to your Go files
//  2. Define structs with gozod struct tags
//  3. Run: go generate ./...
//  4. Generated *_gen.go files contain optimized Schema() methods
//  5. FromStruct automatically detects and uses generated methods
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

	// In real usage after running go generate, this would use generated Schema() method
	schema := gozod.FromStruct[User]()
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

	productSchema := gozod.FromStruct[Product]()
	productResult, err := productSchema.Parse(product)

	if err != nil {
		fmt.Printf("❌ Validation failed: %v\n", err)
	} else {
		fmt.Printf("✅ Product validated: %s (%s)\n", productResult.Name, productResult.SKU)
		fmt.Printf("   Price: $%.2f %s\n", productResult.Price, productResult.Currency)
	}

	fmt.Println("\n💡 Code Generation Benefits:")
	fmt.Println("   • 5-10x faster validation (zero reflection)")
	fmt.Println("   • 50-70% reduction in memory allocations")
	fmt.Println("   • Generated Schema() methods in *_gen.go files")
	fmt.Println("   • Automatic detection by FromStruct")

	fmt.Println("\n📚 Next Steps:")
	fmt.Println("   1. Run: go generate ./...")
	fmt.Println("   2. Check generated *_gen.go files")
	fmt.Println("   3. Use gozod.FromStruct[T]() for validation")
}
