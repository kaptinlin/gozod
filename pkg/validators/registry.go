package validators

import (
	"errors"
	"fmt"
	"sync"
)

// Static error variables
var (
	ErrValidatorAlreadyExists = errors.New("validator already exists")
)

// ValidatorRegistry manages registered validators
type ValidatorRegistry struct {
	validators map[string]any // validator name -> validator instance
	mu         sync.RWMutex
}

// Global registry instance
var registry = &ValidatorRegistry{
	validators: make(map[string]any),
}

// Register registers a validator instance by its name
func Register[T any](validator ZodValidator[T]) error {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	name := validator.Name()
	if _, exists := registry.validators[name]; exists {
		return fmt.Errorf("%w: %s", ErrValidatorAlreadyExists, name)
	}

	registry.validators[name] = validator
	return nil
}

// Get retrieves a validator by name and type
func Get[T any](name string) (ZodValidator[T], bool) {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	if validator, exists := registry.validators[name]; exists {
		if typedValidator, ok := validator.(ZodValidator[T]); ok {
			return typedValidator, true
		}
	}
	return nil, false
}

// GetAny retrieves a validator by name without type checking
func GetAny(name string) (any, bool) {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	validator, exists := registry.validators[name]
	return validator, exists
}

// ListValidators returns all registered validator names
func ListValidators() []string {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	names := make([]string, 0, len(registry.validators))
	for name := range registry.validators {
		names = append(names, name)
	}
	return names
}

// Unregister removes a validator by name
func Unregister(name string) {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	delete(registry.validators, name)
}

// Clear removes all registered validators
func Clear() {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	clear(registry.validators)
}
