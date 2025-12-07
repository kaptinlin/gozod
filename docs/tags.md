# GoZod Struct Tags Guide

Declarative validation for Go structs using familiar tag syntax with zero-reflection performance optimization.

## 🚀 Quick Start

```go
package main

import (
    "fmt"
    "github.com/kaptinlin/gozod"
)

type User struct {
    Name  string `gozod:"required,min=2,max=50"`
    Email string `gozod:"required,email"`
    Age   int    `gozod:"min=18,max=120"`
}

func main() {
    schema := gozod.FromStruct[User]()
    
    user := User{
        Name:  "Alice Smith",
        Email: "alice@example.com",
        Age:   28,
    }
    
    validatedUser, err := schema.Parse(user)
    if err != nil {
        fmt.Printf("Validation error: %v\n", err)
        return
    }
    
    fmt.Printf("Valid user: %+v\n", validatedUser)
}
```

## 📦 Installation

```bash
go get github.com/kaptinlin/gozod
```

## 🏷️ Tag Syntax

### Core Rules

- **Optional by Default**: Fields are optional unless marked `required`
- **Comma Separated**: `"required,min=2,max=50"`
- **Parameters**: `"min=5"` or `"enum=red green blue"`
- **Skip Validation**: `gozod:"-"` to exclude field completely

```go
type Example struct {
    Required string `gozod:"required"`        // Must be present
    Optional string                           // Optional by default  
    Skipped  string `gozod:"-"`              // Skip validation
    Multiple string `gozod:"required,min=2,max=100,email"`
}
```

### Field Processing

```go
// JSON field name mapping
type User struct {
    FullName string `json:"full_name" gozod:"required,min=2"`
    Email    string `json:"email" gozod:"required,email"`
}

// Schema automatically uses JSON field names for validation paths
schema := gozod.FromStruct[User]()
```

---

## 📝 Available Tag Rules

### String Validation

| Rule | Description | Example |
|------|-------------|---------|
| `required` | Field must be present and non-empty | `gozod:"required"` |
| `min=N` | Minimum string length | `gozod:"min=3"` |
| `max=N` | Maximum string length | `gozod:"max=100"` |
| `length=N` | Exact string length | `gozod:"length=10"` |
| `email` | Valid email format | `gozod:"email"` |
| `url` | Valid URL format | `gozod:"url"` |
| `uuid` | Valid UUID format | `gozod:"uuid"` |
| `regex=pattern` | Custom regex pattern | `gozod:"regex=^[A-Z][a-z]+$"` |

### Numeric Validation

| Rule | Description | Example |
|------|-------------|---------|
| `min=N` | Minimum numeric value | `gozod:"min=0"` |
| `max=N` | Maximum numeric value | `gozod:"max=120"` |
| `positive` | Must be greater than 0 | `gozod:"positive"` |
| `negative` | Must be less than 0 | `gozod:"negative"` |
| `nonnegative` | Must be >= 0 | `gozod:"nonnegative"` |
| `nonpositive` | Must be <= 0 | `gozod:"nonpositive"` |

### Array/Slice Validation

| Rule | Description | Example |
|------|-------------|---------|
| `min=N` | Minimum number of elements | `gozod:"min=1"` |
| `max=N` | Maximum number of elements | `gozod:"max=10"` |
| `length=N` | Exact number of elements | `gozod:"length=5"` |
| `nonempty` | At least one element | `gozod:"nonempty"` |

---

## 🛠️ Practical Examples

### Basic User Validation

```go
type User struct {
    Name     string `gozod:"required,min=2,max=50"`
    Username string `gozod:"required,regex=^[a-zA-Z0-9_]+$"`  
    Email    string `gozod:"required,email"`
    Age      int    `gozod:"required,min=18,max=120"`
    Bio      string `gozod:"max=500"`                         // Optional
}

schema := gozod.FromStruct[User]()

// Valid user
user := User{
    Name:     "Alice Johnson",
    Username: "alice_j",
    Email:    "alice@example.com", 
    Age:      28,
    Bio:      "Software engineer",
}

result, err := schema.Parse(user) // ✅ Success
```

### API Request Validation

```go
type CreatePostRequest struct {
    Title    string   `json:"title" gozod:"required,min=3,max=200"`
    Content  string   `json:"content" gozod:"required,min=10"`
    Tags     []string `json:"tags" gozod:"min=1,max=10"`
    AuthorID int      `json:"author_id" gozod:"required,positive"`
    Draft    bool     `json:"draft"`                               // Optional boolean
}

func createPostHandler(w http.ResponseWriter, r *http.Request) {
    var req CreatePostRequest
    json.NewDecoder(r.Body).Decode(&req)
    
    schema := gozod.FromStruct[CreatePostRequest]()
    validatedReq, err := schema.Parse(req)
    if err != nil {
        // Handle validation errors
        writeErrorResponse(w, err)
        return
    }
    
    // Use validated request
    createPost(validatedReq)
}
```

### Configuration Validation

```go
type DatabaseConfig struct {
    Host     string `yaml:"host" gozod:"required"`
    Port     int    `yaml:"port" gozod:"required,min=1,max=65535"`
    Database string `yaml:"database" gozod:"required,min=1"`
    Username string `yaml:"username" gozod:"required"`
    Password string `yaml:"password" gozod:"required,min=8"`
    SSL      bool   `yaml:"ssl"`
    Timeout  int    `yaml:"timeout" gozod:"min=1,max=300"`  // seconds
}

type AppConfig struct {
    Environment string         `yaml:"environment" gozod:"required,regex=^(dev|staging|prod)$"`
    Port        int            `yaml:"port" gozod:"required,min=1000,max=9999"`
    Database    DatabaseConfig `yaml:"database" gozod:"required"`
    Debug       bool           `yaml:"debug"`
}

// Load and validate config
func LoadConfig(path string) (*AppConfig, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    
    var config AppConfig
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, err
    }
    
    schema := gozod.FromStruct[AppConfig]()
    return schema.Parse(config)
}
```

---

## 🔄 Nested Structures

### Basic Nested Validation

```go
type Address struct {
    Street  string `gozod:"required,min=5"`
    City    string `gozod:"required,min=2"`
    Country string `gozod:"required,length=2"`  // ISO country code
    ZipCode string `gozod:"required,regex=^\\d{5}$"`
}

type UserProfile struct {
    Name    string  `gozod:"required,min=2,max=50"`
    Email   string  `gozod:"required,email"`
    Address Address `gozod:"required"`                // Nested struct
    Age     int     `gozod:"required,min=18"`
}

schema := gozod.FromStruct[UserProfile]()

profile := UserProfile{
    Name:  "Bob Smith",
    Email: "bob@example.com",
    Address: Address{
        Street:  "123 Main Street",
        City:    "San Francisco",
        Country: "US",
        ZipCode: "94105",
    },
    Age: 30,
}

result, err := schema.Parse(profile) // ✅ Validates nested structure
```

### Circular References (Automatic Handling)

```go
type User struct {
    Name    string  `gozod:"required,min=2"`
    Email   string  `gozod:"required,email"`
    Friends []*User `gozod:"max=10"`           // Circular reference
}

// GoZod automatically detects and handles circular references
schema := gozod.FromStruct[User]()

alice := &User{
    Name:  "Alice",
    Email: "alice@example.com",
}

bob := &User{
    Name:    "Bob", 
    Email:   "bob@example.com",
    Friends: []*User{alice},  // Reference to Alice
}

alice.Friends = []*User{bob}  // Circular reference

result, err := schema.Parse(*alice) // ✅ No stack overflow
```

---

## 📊 Array and Slice Validation

### Array Element Validation

```go
type Team struct {
    Name    string   `gozod:"required,min=2,max=50"`
    Members []string `gozod:"required,min=1,max=20"`     // At least 1, max 20
    Skills  []string `gozod:"min=3"`                     // Each member needs 3+ skills
    Scores  []int    `gozod:"nonempty"`                  // Must have scores
}

schema := gozod.FromStruct[Team]()

team := Team{
    Name:    "Backend Team",
    Members: []string{"Alice", "Bob", "Charlie"},
    Skills:  []string{"Go", "Docker", "Kubernetes", "PostgreSQL"},
    Scores:  []int{85, 92, 78},
}

result, err := schema.Parse(team) // ✅ All arrays validated
```

### Nested Array Validation

```go
type Project struct {
    Name  string     `gozod:"required,min=3"`
    Teams []Team     `gozod:"required,min=1,max=5"`  // 1-5 teams
    Tags  [][]string `gozod:"max=10"`                // Nested arrays
}

schema := gozod.FromStruct[Project]()
// Automatically validates each team in the slice
```

---

## 🎯 Advanced Tag Features

### Custom Validation with Refine

For custom validation logic beyond built-in tags, use `.Refine()` after creating the schema:

```go
package main

import (
    "strings"
    "github.com/kaptinlin/gozod"
)

type Company struct {
    Name   string `gozod:"required,min=2"`
    Domain string `gozod:"required"`
    Email  string `gozod:"required,email"`
}

func main() {
    // Create schema from struct tags
    schema := gozod.FromStruct[Company]()
    
    // Add custom validation with Refine
    schemaWithCustomValidation := schema.Refine(func(c Company) bool {
        // Domain must end with .com or .org
        return strings.HasSuffix(c.Domain, ".com") || strings.HasSuffix(c.Domain, ".org")
    }, "Domain must end with .com or .org")
    
    // Valid company
    validCompany := Company{
        Name:   "Acme Corp",
        Domain: "acme.com",
        Email:  "contact@acme.com",
    }
    result, err := schemaWithCustomValidation.Parse(validCompany)  // ✅ Success
    
    // Invalid company
    invalidCompany := Company{
        Name:   "Tech Inc",
        Domain: "tech.io",  // ❌ Not .com or .org
        Email:  "info@tech.io",
    }
    _, err = schemaWithCustomValidation.Parse(invalidCompany)  // ❌ Validation fails
}
```

### Cross-Field Validation

```go
type PasswordForm struct {
    Password        string `gozod:"required,min=8"`
    ConfirmPassword string `gozod:"required"`
}

schema := gozod.FromStruct[PasswordForm]().Refine(func(form PasswordForm) bool {
    return form.Password == form.ConfirmPassword
}, "Passwords must match")
```

---

## ⚡ Performance Optimization

### Code Generation

For maximum performance, use code generation to eliminate reflection:

```go
//go:generate gozodgen

type User struct {
    Name  string `gozod:"required,min=2"`
    Email string `gozod:"required,email"`
    Age   int    `gozod:"required,min=18"`
}

// Generated Schema() method in user_gen.go provides zero-reflection validation
func main() {
    schema := gozod.FromStruct[User]()  // Automatically uses generated code
    
    user := User{Name: "Alice", Email: "alice@example.com", Age: 25}
    result, err := schema.Parse(user)   // 5-10x faster than reflection
}
```

### StrictParse for Known Types

```go
schema := gozod.FromStruct[User]()

// Use StrictParse when input type is guaranteed
user := User{Name: "Alice", Email: "alice@example.com", Age: 25}
result, err := schema.StrictParse(user)  // Optimal performance
```

---

## 🛠️ Error Handling

### Structured Error Information

```go
type User struct {
    Name  string `gozod:"required,min=2,max=50"`
    Email string `gozod:"required,email"`
    Age   int    `gozod:"required,min=18,max=120"`
}

schema := gozod.FromStruct[User]()

invalidUser := User{
    Name:  "A",                    // Too short
    Email: "invalid-email",        // Invalid format
    Age:   15,                     // Too young
}

_, err := schema.Parse(invalidUser)
if err != nil {
    if zodErr, ok := err.(*issues.ZodError); ok {
        // Access structured validation issues
        for _, issue := range zodErr.Issues {
            fmt.Printf("Field: %v, Error: %s\n", issue.Path, issue.Message)
        }
        
        // Pretty formatted errors
        fmt.Println(zodErr.PrettifyError())
        
        // Field-specific errors for forms
        fieldErrors := zodErr.FlattenError()
        for field, errors := range fieldErrors.FieldErrors {
            fmt.Printf("%s: %v\n", field, errors)
        }
    }
}
```

### Custom Error Messages

```go
type User struct {
    Name  string `gozod:"required,min=2" error:"Name must be at least 2 characters"`
    Email string `gozod:"required,email" error:"Please provide a valid email address"`
    Age   int    `gozod:"required,min=18" error:"Must be 18 or older"`
}

// Custom error messages will be used in validation failures
```

---

## 🏢 Real-World Examples

### E-commerce Product Validation

```go
type Money struct {
    Amount   int    `gozod:"required,nonnegative"`  // Cents
    Currency string `gozod:"required,length=3"`     // ISO currency code
}

type Product struct {
    ID          string   `json:"id" gozod:"required,uuid"`
    Name        string   `json:"name" gozod:"required,min=3,max=200"`
    Description string   `json:"description" gozod:"max=2000"`
    Price       Money    `json:"price" gozod:"required"`
    Category    string   `json:"category" gozod:"required,min=2"`
    Tags        []string `json:"tags" gozod:"max=20"`
    InStock     bool     `json:"in_stock"`
    Weight      int      `json:"weight" gozod:"positive"`  // grams
}

func createProduct(w http.ResponseWriter, r *http.Request) {
    var product Product
    if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
        http.Error(w, "Invalid JSON", 400)
        return
    }
    
    schema := gozod.FromStruct[Product]()
    validProduct, err := schema.Parse(product)
    if err != nil {
        writeValidationError(w, err)
        return
    }
    
    // Save validated product
    savedProduct := saveProduct(validProduct)
    json.NewEncoder(w).Encode(savedProduct)
}
```

### User Registration Form

```go
type RegistrationForm struct {
    FirstName       string `form:"first_name" gozod:"required,min=2,max=50"`
    LastName        string `form:"last_name" gozod:"required,min=2,max=50"`
    Username        string `form:"username" gozod:"required,min=3,max=20,regex=^[a-zA-Z0-9_]+$"`
    Email           string `form:"email" gozod:"required,email"`
    Password        string `form:"password" gozod:"required,min=8"`
    ConfirmPassword string `form:"confirm_password" gozod:"required"`
    Age             int    `form:"age" gozod:"required,min=13,max=120"`
    AgreeToTerms    bool   `form:"agree_to_terms" gozod:"required"`
    Newsletter      bool   `form:"newsletter"`  // Optional
}

func registerUser(c *gin.Context) {
    var form RegistrationForm
    if err := c.ShouldBind(&form); err != nil {
        c.JSON(400, gin.H{"error": "Form binding failed"})
        return
    }
    
    // Additional custom validation
    if form.Password != form.ConfirmPassword {
        c.JSON(400, gin.H{"error": "Passwords do not match"})
        return
    }
    
    schema := gozod.FromStruct[RegistrationForm]()
    validForm, err := schema.Parse(form)
    if err != nil {
        c.JSON(400, gin.H{"validation_errors": formatValidationErrors(err)})
        return
    }
    
    // Create user account
    user := createUserAccount(validForm)
    c.JSON(201, gin.H{"user": user})
}
```

### Multi-tenant SaaS Configuration

```go
type FeatureFlags struct {
    Analytics   bool `yaml:"analytics"`
    CustomTheme bool `yaml:"custom_theme"`
    APIAccess   bool `yaml:"api_access"`
    UserLimit   int  `yaml:"user_limit" gozod:"min=1,max=1000"`
}

type TenantConfig struct {
    TenantID     string       `yaml:"tenant_id" gozod:"required,uuid"`
    Name         string       `yaml:"name" gozod:"required,min=2,max=100"`
    Domain       string       `yaml:"domain" gozod:"required,regex=^[a-z0-9\\-]+\\.[a-z]{2,}$"`
    AdminEmail   string       `yaml:"admin_email" gozod:"required,email"`
    Plan         string       `yaml:"plan" gozod:"required,regex=^(starter|pro|enterprise)$"`
    Features     FeatureFlags `yaml:"features" gozod:"required"`
    Active       bool         `yaml:"active"`
    BillingEmail string       `yaml:"billing_email" gozod:"email"`  // Optional
    WebhookURL   string       `yaml:"webhook_url" gozod:"url"`      // Optional
}

func validateTenantConfig(configPath string) (*TenantConfig, error) {
    data, err := os.ReadFile(configPath)
    if err != nil {
        return nil, fmt.Errorf("reading config: %w", err)
    }
    
    var config TenantConfig
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, fmt.Errorf("parsing YAML: %w", err)
    }
    
    schema := gozod.FromStruct[TenantConfig]()
    validConfig, err := schema.Parse(config)
    if err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    
    return &validConfig, nil
}
```

---

## 🔧 Integration Patterns

### Gin Web Framework

```go
func setupValidatedRoutes(r *gin.Engine) {
    r.POST("/users", func(c *gin.Context) {
        var user User
        if err := c.ShouldBindJSON(&user); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }
        
        schema := gozod.FromStruct[User]()
        if validUser, err := schema.Parse(user); err != nil {
            c.JSON(422, gin.H{"validation_error": err.Error()})
        } else {
            result := createUser(validUser)
            c.JSON(201, result)
        }
    })
}
```

### Fiber Framework

```go
func setupFiberRoutes(app *fiber.App) {
    app.Post("/products", func(c *fiber.Ctx) error {
        var product Product
        if err := c.BodyParser(&product); err != nil {
            return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
        }
        
        schema := gozod.FromStruct[Product]()
        validProduct, err := schema.Parse(product)
        if err != nil {
            return c.Status(422).JSON(fiber.Map{"validation_error": err.Error()})
        }
        
        result := saveProduct(validProduct)
        return c.Status(201).JSON(result)
    })
}
```

---

## 📋 Tag Reference Summary

### Quick Reference Table

| Category | Tag | Description | Example |
|----------|-----|-------------|---------|
| **Presence** | `required` | Field must be present | `gozod:"required"` |
| | `-` | Skip field validation | `gozod:"-"` |
| **String** | `min=N` | Minimum length | `gozod:"min=3"` |
| | `max=N` | Maximum length | `gozod:"max=100"` |
| | `length=N` | Exact length | `gozod:"length=10"` |
| | `email` | Email format | `gozod:"email"` |
| | `url` | URL format | `gozod:"url"` |
| | `uuid` | UUID format | `gozod:"uuid"` |
| | `regex=pattern` | Custom pattern | `gozod:"regex=^[A-Z]+$"` |
| **Numeric** | `min=N` | Minimum value | `gozod:"min=0"` |
| | `max=N` | Maximum value | `gozod:"max=120"` |
| | `positive` | Greater than 0 | `gozod:"positive"` |
| | `negative` | Less than 0 | `gozod:"negative"` |
| | `nonnegative` | Greater or equal to 0 | `gozod:"nonnegative"` |
| | `nonpositive` | Less or equal to 0 | `gozod:"nonpositive"` |
| **Arrays** | `min=N` | Minimum elements | `gozod:"min=1"` |
| | `max=N` | Maximum elements | `gozod:"max=10"` |
| | `length=N` | Exact elements | `gozod:"length=5"` |
| | `nonempty` | At least one element | `gozod:"nonempty"` |

---

This comprehensive guide covers all aspects of GoZod struct tags, from basic usage to advanced patterns and real-world integration examples. The tag system provides a declarative, maintainable way to define validation rules while maintaining full type safety and optimal performance.