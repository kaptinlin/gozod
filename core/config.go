package core

// ZodConfig represents global configuration for validation and error handling
type ZodConfig struct {
	// CustomError has highest priority, overrides LocaleError
	CustomError ZodErrorMap
	// LocaleError has lowest priority in error resolution chain
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

// globalConfig is the global configuration instance
var globalConfig = &ZodConfig{}

// Config updates and returns the global configuration.
// It merges the provided config with the existing global settings.
func Config(config *ZodConfig) *ZodConfig {
	if config != nil {
		// Merge fields instead of replacing the whole object.
		// This allows for granular updates.
		if config.CustomError != nil {
			globalConfig.CustomError = config.CustomError
		}
		if config.LocaleError != nil {
			globalConfig.LocaleError = config.LocaleError
		}
	} else {
		// A nil config resets the global configuration.
		globalConfig.CustomError = nil
		globalConfig.LocaleError = nil
	}
	// Return the actual global config.
	return globalConfig
}

// GetConfig returns a read-only copy of the current global configuration.
func GetConfig() *ZodConfig {
	return globalConfig
}
