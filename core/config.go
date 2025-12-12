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

// clone creates a copy of the ZodConfig
func (c *ZodConfig) clone() *ZodConfig {
	if c == nil {
		return &ZodConfig{}
	}
	return &ZodConfig{
		CustomError: c.CustomError,
		LocaleError: c.LocaleError,
	}
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
	return newConfig.clone()
}

// GetConfig returns a read-only copy of the current global configuration
func GetConfig() *ZodConfig {
	return globalConfig.Load().clone()
}
