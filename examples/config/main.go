package main

import (
	"fmt"

	"github.com/go-json-experiment/json"

	"github.com/kaptinlin/gozod"
)

// AppConfig represents application configuration struct.
type AppConfig struct {
	Port     int    `json:"port"`
	Database string `json:"database"`
	Debug    bool   `json:"debug"`
}

func main() {
	fmt.Println("Config validation example")

	// Struct schema with field validation (uses JSON tags).
	cfgSchema := gozod.Struct[AppConfig](gozod.StructSchema{
		"port":     gozod.Int().Min(1).Max(65535),
		"database": gozod.URL(),
		"debug":    gozod.Bool().Default(false),
	})

	// Pretend we loaded this JSON from disk.
	rawJSON := []byte(`{
        "port": 8080,
        "database": "https://db.example.com",
        "debug": true
    }`)

	// Unmarshal into generic map first.
	var data map[string]any
	_ = json.Unmarshal(rawJSON, &data)

	// MustParse will panic on validation failure (startup scenario).
	cfg := cfgSchema.MustParse(data)

	fmt.Printf("✓ config loaded: %+v\n", cfg)
}
