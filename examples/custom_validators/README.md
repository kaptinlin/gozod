# Custom Validators Example

This example demonstrates how to create and use custom validators with GoZod.

## Features Demonstrated

1. **Creating custom validators** - Implement your own validation logic
2. **Parameterized validators** - Validators that accept configuration parameters
3. **Validator registration** - Register validators directly in your application
4. **Struct tag integration** - Use custom validators in struct tags
5. **Programmatic usage** - Use validators without struct tags

## Example Validators

The example includes three types of custom validators:

### 1. Basic Validator (UniqueUsernameValidator)
- **Type**: `string`
- **Tag**: `unique_username`
- **Purpose**: Validates username uniqueness (simulated database check)

### 2. Parameterized Int Validator (MinValueValidator)
- **Type**: `int`
- **Tag**: `min_value=N`
- **Purpose**: Validates minimum integer value
- **Example**: `min_value=18` for age validation

### 3. Parameterized String Validator (PrefixValidator)
- **Type**: `string`
- **Tag**: `prefix=PREFIX`
- **Purpose**: Validates string prefix
- **Example**: `prefix=PROD` for product SKUs

## Usage

### Struct Tag Usage
```go
type User struct {
    Username string `gozod:"required,unique_username"`
    Email    string `gozod:"required,email"`
    Age      int    `gozod:"required,min_value=18"`
}

type Product struct {
    SKU   string `gozod:"required,prefix=PROD"`
    Name  string `gozod:"required,min=3,max=100"`
    Price int    `gozod:"required,min_value=1"`
}
```

### Registration
```go
func registerCustomValidators() {
    validators.Register(&UniqueUsernameValidator{})
    validators.Register(&PrefixValidator{})
    validators.Register(&MinValueValidator{})
}
```

### Programmatic Usage
```go
// Get validator from registry
validator, _ := validators.Get[string]("unique_username")

// Use with schema
schema := gozod.String().
    Min(3).
    Max(20).
    Refine(validators.ToRefineFn(validator))
```

## Running the Example

```bash
cd examples/custom_validators
go run main.go
```

## Output

The example demonstrates:
- ✅ Successful validation of valid data
- ✅ Rejection of invalid data with clear error messages
- ✅ Both struct tag and programmatic usage patterns

## Creating Your Own Validators

To create a custom validator:

1. **Implement the interface**:
```go
type MyValidator struct{}

func (v *MyValidator) Name() string { return "my_validator" }
func (v *MyValidator) Validate(value T) bool { /* logic */ }
func (v *MyValidator) ErrorMessage(ctx *core.ParseContext) string { /* message */ }
```

2. **For parameterized validators**, also implement:
```go
func (v *MyValidator) ValidateParam(value T, param string) bool { /* logic */ }
func (v *MyValidator) ErrorMessageWithParam(ctx *core.ParseContext, param string) string { /* message */ }
```

3. **Register at startup**:
```go
validators.Register(&MyValidator{})
```

4. **Use in struct tags or programmatically**:
```go
type MyStruct struct {
    Field string `gozod:"my_validator"`
}
```