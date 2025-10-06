package core

import "sync/atomic"

// ZodConfig represents global configuration for validation and error handling
type ZodConfig struct {
	CustomError ZodErrorMap
	LocaleError ZodErrorMap
}

// GetCustomError returns the custom error map from ZodConfig
func (c *ZodConfig) GetCustomError() ZodErrorMap {
	if c == nil {
		return nil
	}
	return c.CustomError
}

// GetLocaleError returns the locale error map from ZodConfig
func (c *ZodConfig) GetLocaleError() ZodErrorMap {
	if c == nil {
		return nil
	}
	return c.LocaleError
}

var globalConfig atomic.Pointer[ZodConfig]

func init() {
	// Initialize with empty config
	globalConfig.Store(&ZodConfig{})
}

// Config updates and returns the global configuration
func Config(config *ZodConfig) *ZodConfig {
	if config == nil {
		// Reset to empty config
		newConfig := &ZodConfig{}
		globalConfig.Store(newConfig)
		return &ZodConfig{}
	}

	// Get current config
	current := globalConfig.Load()

	// Create new config (immutable update pattern)
	newConfig := &ZodConfig{
		CustomError: current.CustomError,
		LocaleError: current.LocaleError,
	}

	// Apply updates
	if config.CustomError != nil {
		newConfig.CustomError = config.CustomError
	}
	if config.LocaleError != nil {
		newConfig.LocaleError = config.LocaleError
	}

	// Atomic swap
	globalConfig.Store(newConfig)

	// Return copy
	return &ZodConfig{
		CustomError: newConfig.CustomError,
		LocaleError: newConfig.LocaleError,
	}
}

// GetConfig returns a read-only copy of the current global configuration
func GetConfig() *ZodConfig {
	cfg := globalConfig.Load()
	// Return a copy to prevent external mutation
	return &ZodConfig{
		CustomError: cfg.CustomError,
		LocaleError: cfg.LocaleError,
	}
}
