package core

import "sync/atomic"

// ZodConfig represents global configuration for validation and error handling.
type ZodConfig struct {
	CustomError ZodErrorMap
	LocaleError ZodErrorMap
}

// clone creates a copy of the ZodConfig.
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

// SetConfig updates and returns the global configuration.
func SetConfig(config *ZodConfig) *ZodConfig {
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

// Config returns a read-only copy of the current global configuration.
func Config() *ZodConfig {
	return globalConfig.Load().clone()
}
