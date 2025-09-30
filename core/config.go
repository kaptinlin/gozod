package core

import "sync"

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

var (
	globalConfig   = &ZodConfig{}
	globalConfigMu sync.RWMutex
)

// Config updates and returns the global configuration
func Config(config *ZodConfig) *ZodConfig {
	globalConfigMu.Lock()
	defer globalConfigMu.Unlock()

	if config != nil {
		if config.CustomError != nil {
			globalConfig.CustomError = config.CustomError
		}
		if config.LocaleError != nil {
			globalConfig.LocaleError = config.LocaleError
		}
	} else {
		globalConfig.CustomError = nil
		globalConfig.LocaleError = nil
	}

	return &ZodConfig{
		CustomError: globalConfig.CustomError,
		LocaleError: globalConfig.LocaleError,
	}
}

// GetConfig returns a read-only copy of the current global configuration
func GetConfig() *ZodConfig {
	globalConfigMu.RLock()
	defer globalConfigMu.RUnlock()

	return &ZodConfig{
		CustomError: globalConfig.CustomError,
		LocaleError: globalConfig.LocaleError,
	}
}
