package core

import (
	"sync"
	"testing"
)

func TestSetConfig(t *testing.T) {
	customErr := func(issue ZodRawIssue) string { return "custom" }
	localeErr := func(issue ZodRawIssue) string { return "locale" }

	config := SetConfig(&ZodConfig{
		CustomError: customErr,
		LocaleError: localeErr,
	})

	if config.CustomError == nil {
		t.Error("Expected CustomError to be set")
	}
	if config.LocaleError == nil {
		t.Error("Expected LocaleError to be set")
	}

	retrieved := Config()
	if retrieved.CustomError == nil {
		t.Error("Expected retrieved CustomError to be set")
	}
	if retrieved.LocaleError == nil {
		t.Error("Expected retrieved LocaleError to be set")
	}

	reset := SetConfig(nil)
	if reset.CustomError != nil {
		t.Error("Expected CustomError to be nil after reset")
	}
	if reset.LocaleError != nil {
		t.Error("Expected LocaleError to be nil after reset")
	}
}

func TestSetConfigConcurrent(t *testing.T) {
	t.Parallel()
	SetConfig(nil)

	var wg sync.WaitGroup
	iterations := 100

	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range iterations {
				_ = Config()
			}
		}()
	}

	customErr := func(issue ZodRawIssue) string { return "custom" }
	localeErr := func(issue ZodRawIssue) string { return "locale" }

	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range iterations {
				SetConfig(&ZodConfig{
					CustomError: customErr,
					LocaleError: localeErr,
				})
			}
		}()
	}

	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range iterations {
				if j%2 == 0 {
					Config()
				} else {
					SetConfig(&ZodConfig{CustomError: customErr})
				}
			}
		}()
	}

	wg.Wait()
}

func TestSetConfigPartialUpdate(t *testing.T) {
	SetConfig(nil)

	customErr := func(issue ZodRawIssue) string { return "custom" }
	localeErr := func(issue ZodRawIssue) string { return "locale" }

	SetConfig(&ZodConfig{
		CustomError: customErr,
		LocaleError: localeErr,
	})

	newCustomErr := func(issue ZodRawIssue) string { return "new" }
	result := SetConfig(&ZodConfig{
		CustomError: newCustomErr,
	})

	if result.CustomError == nil {
		t.Error("Expected CustomError to be updated")
	}
	if result.LocaleError == nil {
		t.Error("Expected LocaleError to be preserved")
	}

	retrieved := Config()
	if retrieved.CustomError == nil {
		t.Error("Expected CustomError to be present in Config")
	}
	if retrieved.LocaleError == nil {
		t.Error("Expected LocaleError to be preserved in Config")
	}
}

func BenchmarkConfig(b *testing.B) {
	customErr := func(issue ZodRawIssue) string { return "custom" }
	SetConfig(&ZodConfig{CustomError: customErr})

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		_ = Config()
	}
}

func BenchmarkConfigParallel(b *testing.B) {
	customErr := func(issue ZodRawIssue) string { return "custom" }
	SetConfig(&ZodConfig{CustomError: customErr})

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = Config()
		}
	})
}

func BenchmarkSetConfig(b *testing.B) {
	customErr := func(issue ZodRawIssue) string { return "custom" }

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		SetConfig(&ZodConfig{CustomError: customErr})
	}
}
