// Package main demonstrates GoZod struct tag validation usage
// This example covers the most important features in a simple, practical way.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/kaptinlin/gozod"
)

// =============================================================================
// EXAMPLE STRUCTS WITH VALIDATION TAGS
// =============================================================================

// User demonstrates basic struct tag validation
type User struct {
	Name  string `gozod:"required,min=2,max=50" json:"name"`
	Email string `gozod:"required,email" json:"email"`
	Age   int    `gozod:"min=18,max=120" json:"age"`
	Bio   string `gozod:"max=500" json:"bio"` // Optional field
}

// Address demonstrates nested struct validation
type Address struct {
	Street  string `gozod:"required,min=5" json:"street"`
	City    string `gozod:"required,min=2" json:"city"`
	Country string `gozod:"required,length=2" json:"country"` // ISO country code
}

// Profile demonstrates nested validation
type Profile struct {
	User    User    `json:"user"`    // Deep validation
	Address Address `json:"address"` // Deep validation
}

// UpdateUserRequest demonstrates pointer field validation for PATCH operations
type UpdateUserRequest struct {
	Name  *string `gozod:"nilable,min=2,max=50" json:"name,omitempty"`
	Email *string `gozod:"nilable,email" json:"email,omitempty"`
	Age   *int    `gozod:"nilable,min=18,max=120" json:"age,omitempty"`
}

// ComprehensiveUser demonstrates various tag features
type ComprehensiveUser struct {
	// Identity fields
	Name     string `gozod:"required,min=2,max=50" json:"name"`
	Username string `gozod:"required,min=3,max=30,regex=^[a-zA-Z0-9_]+$" json:"username"`
	Email    string `gozod:"required,email" json:"email"`

	// Optional fields
	Website string `gozod:"optional" json:"website,omitempty"`
	Bio     string `gozod:"max=500" json:"bio,omitempty"`

	// Numeric validation
	Age   int     `gozod:"required,min=18,max=120" json:"age"`
	Score float64 `gozod:"optional" json:"score"`

	// Array validation
	Tags   []string `gozod:"optional" json:"tags"`
	Skills []string `gozod:"optional" json:"skills"`

	// Enum validation
	Status string `gozod:"required" json:"status"`
	Role   string `gozod:"required" json:"role"`
	Theme  string `gozod:"optional" json:"theme"`

	// Format validation
	UserID string `gozod:"required" json:"user_id"`
	Phone  string `gozod:"optional" json:"phone,omitempty"`

	// Time validation
	CreatedAt time.Time  `gozod:"required" json:"created_at"`
	UpdatedAt *time.Time `gozod:"nilable" json:"updated_at,omitempty"`

	// Default values
	IsActive   bool `gozod:"optional" json:"is_active"`
	LoginCount int  `gozod:"optional,min=0" json:"login_count"`
}

// =============================================================================
// MAIN DEMONSTRATION
// =============================================================================

func main() {
	fmt.Println("🏷️  GoZod Struct Tags - Simplified Examples")
	fmt.Println("==========================================")

	// Example 1: Basic struct validation
	fmt.Println("\n📝 Example 1: Basic User Validation")
	demonstrateBasicValidation()

	// Example 2: Nested struct validation
	fmt.Println("\n🏗️  Example 2: Nested Struct Validation")
	demonstrateNestedValidation()

	// Example 3: Pointer field validation (PATCH operations)
	fmt.Println("\n👉 Example 3: Pointer Field Validation (PATCH)")
	demonstratePointerFields()

	// Example 4: Comprehensive tag features
	fmt.Println("\n🌟 Example 4: Comprehensive Tag Features")
	demonstrateComprehensiveValidation()

	// Example 5: Error handling
	fmt.Println("\n🐛 Example 5: Error Handling")
	demonstrateErrorHandling()

	// Example 6: JSON integration
	fmt.Println("\n📄 Example 6: JSON Integration")
	demonstrateJSONIntegration()

	// Example 7: Business logic integration
	fmt.Println("\n🔗 Example 7: Business Logic Integration")
	demonstrateBusinessLogicIntegration()

	fmt.Println("\n✅ All examples completed successfully!")
}

// =============================================================================
// DEMONSTRATION FUNCTIONS
// =============================================================================

func demonstrateBasicValidation() {
	schema := gozod.FromStruct[User]()

	// Valid user
	validUser := User{
		Name:  "Alice Johnson",
		Email: "alice@example.com",
		Age:   28,
		Bio:   "Software engineer passionate about Go",
	}

	result, err := schema.Parse(validUser)
	if err != nil {
		log.Printf("❌ Validation failed: %v", err)
	} else {
		fmt.Printf("✅ Valid user: %s (%s), Age: %d\n", result.Name, result.Email, result.Age)
	}

	// Invalid user - demonstrates validation errors
	invalidUser := User{
		Name:  "A",                      // Too short (min=2)
		Email: "not-an-email",           // Invalid format
		Age:   16,                       // Too young (min=18)
		Bio:   strings.Repeat("x", 600), // Too long (max=500)
	}

	_, err = schema.Parse(invalidUser)
	if err != nil {
		fmt.Printf("❌ Expected validation errors found\n")
	}
}

func demonstrateNestedValidation() {
	schema := gozod.FromStruct[Profile]()

	// Create valid nested structure
	profile := Profile{
		User: User{
			Name:  "Bob Smith",
			Email: "bob@example.com",
			Age:   32,
			Bio:   "Product manager",
		},
		Address: Address{
			Street:  "123 Main Street",
			City:    "San Francisco",
			Country: "US",
		},
	}

	result, err := schema.Parse(profile)
	if err != nil {
		log.Printf("❌ Validation failed: %v", err)
	} else {
		fmt.Printf("✅ Valid profile:\n")
		fmt.Printf("   User: %s (%s)\n", result.User.Name, result.User.Email)
		fmt.Printf("   Address: %s, %s, %s\n",
			result.Address.Street, result.Address.City, result.Address.Country)
	}
}

func demonstratePointerFields() {
	schema := gozod.FromStruct[UpdateUserRequest]()

	// Partial update - only some fields provided
	nameUpdate := "Charlie Brown"
	ageUpdate := 25

	updateRequest := UpdateUserRequest{
		Name: &nameUpdate,
		Age:  &ageUpdate,
		// Email omitted - will be nil
	}

	result, err := schema.Parse(updateRequest)
	if err != nil {
		log.Printf("❌ Validation failed: %v", err)
	} else {
		fmt.Printf("✅ Valid partial update:\n")
		if result.Name != nil {
			fmt.Printf("   Name: %s\n", *result.Name)
		}
		if result.Age != nil {
			fmt.Printf("   Age: %d\n", *result.Age)
		}
		if result.Email == nil {
			fmt.Printf("   Email: not provided (nil)\n")
		}
	}
}

func demonstrateComprehensiveValidation() {
	schema := gozod.FromStruct[ComprehensiveUser]()

	// Create comprehensive user with all features
	user := ComprehensiveUser{
		Name:       "Sarah Connor",
		Username:   "sarah_c",
		Email:      "sarah@resistance.com",
		Website:    "https://resistance.com",
		Bio:        "Resistance leader",
		Age:        35,
		Score:      92.5,
		Tags:       []string{"leadership", "combat", "strategy"},
		Skills:     []string{"Weapons", "Tactics", "Survival"},
		Status:     "active",
		Role:       "admin",
		Theme:      "dark",
		UserID:     "550e8400-e29b-41d4-a716-446655440000",
		Phone:      "+1 (555) 123-4567",
		CreatedAt:  time.Now(),
		IsActive:   true,
		LoginCount: 42,
	}

	result, err := schema.Parse(user)
	if err != nil {
		fmt.Printf("❌ Comprehensive validation failed: %v\n", err)
		return
	}

	fmt.Printf("✅ Comprehensive validation successful!\n")
	fmt.Printf("   User: %s (@%s)\n", result.Name, result.Username)
	fmt.Printf("   Role: %s, Status: %s\n", result.Role, result.Status)
	fmt.Printf("   Score: %.1f, Tags: %v\n", result.Score, result.Tags)
	fmt.Printf("   Active: %t, Login Count: %d\n", result.IsActive, result.LoginCount)
}

func demonstrateErrorHandling() {
	schema := gozod.FromStruct[ComprehensiveUser]()

	// Create user with multiple validation errors
	invalidUser := ComprehensiveUser{
		Name:       "A",             // Too short (min=2)
		Username:   "invalid-user!", // Invalid characters
		Email:      "not-an-email",  // Invalid format
		Age:        15,              // Too young (min=18)
		Score:      150.0,           // Too high (max=100)
		Tags:       []string{},      // Empty array (min=1)
		Skills:     []string{},      // Empty array (nonempty)
		Status:     "unknown",       // Invalid enum
		Role:       "superuser",     // Invalid enum
		UserID:     "not-a-uuid",    // Invalid UUID
		CreatedAt:  time.Time{},     // Zero time for required field
		LoginCount: -5,              // Negative (min=0)
	}

	_, err := schema.Parse(invalidUser)
	if err != nil {
		fmt.Printf("❌ Multiple validation errors (expected):\n")

		// Parse structured error for detailed reporting
		var zodErr *gozod.ZodError
		if errors.As(err, &zodErr) {
			fmt.Printf("   Found %d validation issues:\n", len(zodErr.Issues))

			maxErrors := len(zodErr.Issues)
			if maxErrors > 5 {
				maxErrors = 5
			}
			for i, issue := range zodErr.Issues[:maxErrors] { // Show first 5 errors
				fieldPath := "root"
				if len(issue.Path) > 0 {
					pathStrs := make([]string, len(issue.Path))
					for j, p := range issue.Path {
						pathStrs[j] = fmt.Sprintf("%v", p)
					}
					fieldPath = strings.Join(pathStrs, ".")
				}

				fmt.Printf("   %d. Field '%s': %s\n", i+1, fieldPath, issue.Message)
				if input := issue.Input; input != nil {
					fmt.Printf("      Input: %v\n", input)
				}
			}

			if len(zodErr.Issues) > 5 {
				fmt.Printf("   ... and %d more errors\n", len(zodErr.Issues)-5)
			}
		}
	}
}

func demonstrateJSONIntegration() {
	schema := gozod.FromStruct[ComprehensiveUser]()

	// Simulate JSON API request payload
	jsonPayload := `{
		"name": "John Wick",
		"username": "john_wick",
		"email": "john@continental.com",
		"website": "https://continental-hotel.com",
		"bio": "Professional problem solver",
		"age": 45,
		"score": 98.5,
		"tags": ["assassin", "professional"],
		"skills": ["Marksmanship", "Combat"],
		"status": "active",
		"role": "user",
		"theme": "dark",
		"user_id": "550e8400-e29b-41d4-a716-446655440000",
		"phone": "+1 (555) 999-0001",
		"created_at": "2024-01-15T10:30:00Z",
		"is_active": true,
		"login_count": 100
	}`

	// Parse JSON to struct
	var user ComprehensiveUser
	if err := json.Unmarshal([]byte(jsonPayload), &user); err != nil {
		fmt.Printf("❌ JSON parsing failed: %v\n", err)
		return
	}

	// Validate parsed JSON data
	validatedUser, err := schema.Parse(user)
	if err != nil {
		fmt.Printf("❌ JSON validation failed: %v\n", err)
		return
	}

	fmt.Printf("✅ JSON integration successful!\n")
	fmt.Printf("   Parsed and validated user from JSON:\n")
	fmt.Printf("   Name: %s (@%s)\n", validatedUser.Name, validatedUser.Username)
	fmt.Printf("   Email: %s\n", validatedUser.Email)
	fmt.Printf("   Score: %.1f, Skills: %v\n", validatedUser.Score, validatedUser.Skills)
}

func demonstrateBusinessLogicIntegration() {
	// Combine tag validation with custom business rules
	schema := gozod.FromStruct[User]().
		Refine(func(u User) bool {
			// Business rule: users named "admin" must be at least 25
			return !strings.EqualFold(u.Name, "admin") || u.Age >= 25
		}, "Admin users must be at least 25 years old")

	// Test valid admin user
	adminUser := User{
		Name:  "Admin",
		Email: "admin@company.com",
		Age:   30,
		Bio:   "System administrator",
	}

	result, err := schema.Parse(adminUser)
	if err != nil {
		log.Printf("❌ Validation failed: %v", err)
	} else {
		fmt.Printf("✅ Valid admin user: %s (age %d)\n", result.Name, result.Age)
	}

	// Test invalid admin user (too young)
	youngAdmin := User{
		Name:  "Admin",
		Email: "young@company.com",
		Age:   22,
		Bio:   "Junior admin",
	}

	_, err = schema.Parse(youngAdmin)
	if err != nil {
		fmt.Printf("❌ Expected business rule error for young admin\n")
	}

	// Demonstrate optional chaining
	optionalSchema := gozod.FromStruct[User]().Optional()

	// Test with nil input
	result2, err := optionalSchema.Parse(nil)
	if err != nil {
		log.Printf("❌ Optional validation failed: %v", err)
	} else if result2 == nil {
		fmt.Printf("✅ Optional schema handles nil input correctly\n")
	}
}
