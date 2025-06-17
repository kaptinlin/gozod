package gozod

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

// Config updates and returns the global configuration
func Config(config *ZodConfig) *ZodConfig {
	if config != nil {
		// Update global config in place
		// Always update the fields, even if they are nil (to allow clearing)
		globalConfig.CustomError = config.CustomError
		globalConfig.LocaleError = config.LocaleError
	} else {
		// When config is nil, reset the global configuration
		globalConfig.CustomError = nil
		globalConfig.LocaleError = nil
	}
	// Return the actual global config, not a copy
	return globalConfig
}

// GetConfig returns the current global configuration
func GetConfig() *ZodConfig {
	return globalConfig
}
