# Deduplication Examples (Go)

## 1) Parameter validation

**Before (repeated checks):**
```go
if strings.TrimSpace(email) == "" {
    return fmt.Errorf("email is required")
}
if strings.TrimSpace(name) == "" {
    return fmt.Errorf("name is required")
}
```

**After (single helper):**
```go
func Required(value, field string) (string, error) {
    v := strings.TrimSpace(value)
    if v == "" {
        return "", fmt.Errorf("%s is required", field)
    }
    return v, nil
}

email, err := validate.Required(input.Email, "email")
name, err := validate.Required(input.Name, "name")
```

## 2) Client initialization

**Before (repeated setup):**
```go
token := os.Getenv("API_KEY")
if token == "" {
    token = cfg.APIKey
}
if token == "" {
    return nil, fmt.Errorf("API key missing")
}
client := NewClient(token)
```

**After (factory helper):**
```go
func Client(cfg *Config) (*Client, error) {
    token := os.Getenv("API_KEY")
    if token == "" {
        token = cfg.APIKey
    }
    if token == "" {
        return nil, fmt.Errorf("API key missing")
    }
    return NewClient(token), nil
}
```

## 3) File write with mkdir

**Before (repeated IO):**
```go
dir := filepath.Dir(path)
if err := os.MkdirAll(dir, 0o755); err != nil {
    return err
}
return os.WriteFile(path, data, 0o644)
```

**After (fs helper):**
```go
func WriteFile(path string, data []byte) error {
    if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
        return err
    }
    return os.WriteFile(path, data, 0o644)
}
```
