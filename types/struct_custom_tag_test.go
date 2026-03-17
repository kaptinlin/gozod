package types

import (
	"testing"
)

func TestFromStructWithCustomTag(t *testing.T) {
	type User struct {
		Name  string `validate:"required,min=2,max=50"`
		Email string `validate:"required,email"`
		Age   int    `validate:"min=18,max=120"`
	}

	schema := FromStruct[User](WithTagName("validate"))

	t.Run("valid user", func(t *testing.T) {
		user := User{
			Name:  "Alice",
			Email: "alice@example.com",
			Age:   25,
		}

		result, err := schema.Parse(user)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if result.Name != "Alice" {
			t.Errorf("expected Name=Alice, got %s", result.Name)
		}
	})

	t.Run("invalid email", func(t *testing.T) {
		user := User{
			Name:  "Bob",
			Email: "invalid",
			Age:   30,
		}

		_, err := schema.Parse(user)
		if err == nil {
			t.Fatal("expected validation error")
		}
	})

	t.Run("age too young", func(t *testing.T) {
		user := User{
			Name:  "Charlie",
			Email: "charlie@example.com",
			Age:   15,
		}

		_, err := schema.Parse(user)
		if err == nil {
			t.Fatal("expected validation error")
		}
	})
}

func TestFromStructPtrWithCustomTag(t *testing.T) {
	type Config struct {
		Host string `validate:"required"`
		Port int    `validate:"required,min=1,max=65535"`
	}

	schema := FromStructPtr[Config](WithTagName("validate"))

	t.Run("valid config", func(t *testing.T) {
		cfg := Config{
			Host: "localhost",
			Port: 8080,
		}

		result, err := schema.Parse(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if result.Host != "localhost" {
			t.Errorf("expected Host=localhost, got %s", result.Host)
		}
	})
}

func TestFromStructDefaultTag(t *testing.T) {
	type Product struct {
		Name  string `gozod:"required,min=3"`
		Price int    `gozod:"required,positive"`
	}

	// Without WithTagName, should use default "gozod"
	schema := FromStruct[Product]()

	t.Run("valid product", func(t *testing.T) {
		product := Product{
			Name:  "Widget",
			Price: 100,
		}

		result, err := schema.Parse(product)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if result.Name != "Widget" {
			t.Errorf("expected Name=Widget, got %s", result.Name)
		}
	})
}
