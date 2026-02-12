package core

import (
	"sync"
	"testing"
)

func TestConfig(t *testing.T) {
	// Test setting config
	customErr := func(issue ZodRawIssue) string { return "custom" }
	localeErr := func(issue ZodRawIssue) string { return "locale" }

	config := Config(&ZodConfig{
		CustomError: customErr,
		LocaleError: localeErr,
	})

	if config.CustomError == nil {
		t.Error("Expected CustomError to be set")
	}
	if config.LocaleError == nil {
		t.Error("Expected LocaleError to be set")
	}

	// Test getting config
	retrieved := GetConfig()
	if retrieved.CustomError == nil {
		t.Error("Expected retrieved CustomError to be set")
	}
	if retrieved.LocaleError == nil {
		t.Error("Expected retrieved LocaleError to be set")
	}

	// Test resetting config
	reset := Config(nil)
	if reset.CustomError != nil {
		t.Error("Expected CustomError to be nil after reset")
	}
	if reset.LocaleError != nil {
		t.Error("Expected LocaleError to be nil after reset")
	}
}

func TestConfigConcurrent(t *testing.T) {
	t.Parallel()

	// Reset to clean state
	Config(nil)

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent readers
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range iterations {
				cfg := GetConfig()
				_ = cfg // Use the config
			}
		}()
	}

	// Concurrent writers
	customErr := func(issue ZodRawIssue) string { return "custom" }
	localeErr := func(issue ZodRawIssue) string { return "locale" }

	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range iterations {
				Config(&ZodConfig{
					CustomError: customErr,
					LocaleError: localeErr,
				})
			}
		}()
	}

	// Mixed readers and writers
	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range iterations {
				if j%2 == 0 {
					GetConfig()
				} else {
					Config(&ZodConfig{CustomError: customErr})
				}
			}
		}()
	}

	wg.Wait()
}

func TestConfigPartialUpdate(t *testing.T) {
	// Reset to clean state
	Config(nil)

	customErr := func(issue ZodRawIssue) string { return "custom" }
	localeErr := func(issue ZodRawIssue) string { return "locale" }

	// Set both
	Config(&ZodConfig{
		CustomError: customErr,
		LocaleError: localeErr,
	})

	// Update only CustomError (should preserve LocaleError)
	newCustomErr := func(issue ZodRawIssue) string { return "new" }
	result := Config(&ZodConfig{
		CustomError: newCustomErr,
	})

	if result.CustomError == nil {
		t.Error("Expected CustomError to be updated")
	}
	if result.LocaleError == nil {
		t.Error("Expected LocaleError to be preserved")
	}

	// Verify GetConfig returns the same
	retrieved := GetConfig()
	if retrieved.CustomError == nil {
		t.Error("Expected CustomError to be present in GetConfig")
	}
	if retrieved.LocaleError == nil {
		t.Error("Expected LocaleError to be preserved in GetConfig")
	}
}

func BenchmarkGetConfig(b *testing.B) {
	customErr := func(issue ZodRawIssue) string { return "custom" }
	Config(&ZodConfig{CustomError: customErr})

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		_ = GetConfig()
	}
}

func BenchmarkGetConfigParallel(b *testing.B) {
	customErr := func(issue ZodRawIssue) string { return "custom" }
	Config(&ZodConfig{CustomError: customErr})

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = GetConfig()
		}
	})
}

func BenchmarkConfig(b *testing.B) {
	customErr := func(issue ZodRawIssue) string { return "custom" }

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		Config(&ZodConfig{CustomError: customErr})
	}
}
