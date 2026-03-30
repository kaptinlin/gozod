package core

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetConfig(t *testing.T) {
	customErr := func(issue ZodRawIssue) string { return "custom" }
	localeErr := func(issue ZodRawIssue) string { return "locale" }

	config := SetConfig(&ZodConfig{
		CustomError: customErr,
		LocaleError: localeErr,
	})

	require.NotNil(t, config.CustomError, "Expected CustomError to be set")
	require.NotNil(t, config.LocaleError, "Expected LocaleError to be set")

	retrieved := Config()
	assert.NotNil(t, retrieved.CustomError, "Expected retrieved CustomError to be set")
	assert.NotNil(t, retrieved.LocaleError, "Expected retrieved LocaleError to be set")

	reset := SetConfig(nil)
	assert.Nil(t, reset.CustomError, "Expected CustomError to be nil after reset")
	assert.Nil(t, reset.LocaleError, "Expected LocaleError to be nil after reset")
}

func TestSetConfigConcurrent(t *testing.T) {
	t.Parallel()
	SetConfig(nil)

	var wg sync.WaitGroup
	iterations := 100

	for range 10 {
		wg.Go(func() {
			for range iterations {
				_ = Config()
			}
		})
	}

	customErr := func(issue ZodRawIssue) string { return "custom" }
	localeErr := func(issue ZodRawIssue) string { return "locale" }

	for range 5 {
		wg.Go(func() {
			for range iterations {
				SetConfig(&ZodConfig{
					CustomError: customErr,
					LocaleError: localeErr,
				})
			}
		})
	}

	for range 5 {
		wg.Go(func() {
			for j := range iterations {
				if j%2 == 0 {
					Config()
				} else {
					SetConfig(&ZodConfig{CustomError: customErr})
				}
			}
		})
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

	assert.NotNil(t, result.CustomError, "Expected CustomError to be updated")
	assert.NotNil(t, result.LocaleError, "Expected LocaleError to be preserved")

	retrieved := Config()
	assert.NotNil(t, retrieved.CustomError, "Expected CustomError to be present in Config")
	assert.NotNil(t, retrieved.LocaleError, "Expected LocaleError to be preserved in Config")
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
