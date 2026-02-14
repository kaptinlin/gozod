// Package core provides the foundation contracts for the gozod validation library,
// including interfaces, types, and constants used throughout the system.
package core

import "sync/atomic"

// ZodConfig represents global configuration for validation and error handling.
type ZodConfig struct {
	CustomError ZodErrorMap
	LocaleError ZodErrorMap
}

// clone creates a copy of the ZodConfig.
// This is an unexported helper method for internal use.
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
// A nil argument resets the configuration to its zero state.
func SetConfig(config *ZodConfig) *ZodConfig {
	if config == nil {
		newConfig := &ZodConfig{}
		globalConfig.Store(newConfig)
		return newConfig.clone()
	}

	current := globalConfig.Load()
	newConfig := current.clone()

	if config.CustomError != nil {
		newConfig.CustomError = config.CustomError
	}
	if config.LocaleError != nil {
		newConfig.LocaleError = config.LocaleError
	}

	globalConfig.Store(newConfig)
	return newConfig.clone()
}

// Config returns a read-only copy of the current global configuration.
func Config() *ZodConfig {
	return globalConfig.Load().clone()
}
