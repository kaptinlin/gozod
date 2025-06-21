package engine

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// =============================================================================
// SCHEMA TYPE INITIALIZATION
// =============================================================================

// InitZodType initializes the common fields of a ZodType
// This function sets up the basic internal structure for any schema type
// Enhanced version using mapx and slicex for better data management
func InitZodType[T core.ZodType[any, any]](schema T, def *core.ZodTypeDef) {
	internals := schema.GetInternals()

	// Initialize base internals with version and type information
	internals.Version = core.Version
	internals.Type = def.Type
	internals.Error = def.Error

	// Use mapx to manage internal state
	if internals.Bag == nil {
		internals.Bag = make(map[string]any)
	}

	// Use mapx to set configuration
	mapx.Set(internals.Bag, "version", core.Version)
	mapx.Set(internals.Bag, "type", def.Type)

	// Use slicex to safely copy checks
	if !slicex.IsEmpty(def.Checks) {
		if clonedChecks, err := slicex.ToTyped[core.ZodCheck](def.Checks); err == nil {
			internals.Checks = clonedChecks
		} else {
			// Fallback to manual copy
			internals.Checks = make([]core.ZodCheck, len(def.Checks))
			copy(internals.Checks, def.Checks)
		}
	} else {
		internals.Checks = make([]core.ZodCheck, 0)
	}

	// Initialize Values map if not already done
	// This map stores valid literal values for literal type validation
	if internals.Values == nil {
		internals.Values = make(map[any]struct{})
	}

	// Run onattach callbacks for all checks
	// This allows checks to perform any initialization they need when attached to a schema
	for _, check := range internals.Checks {
		if check != nil {
			if checkInternals := check.GetZod(); checkInternals != nil {
				for _, fn := range checkInternals.OnAttach {
					fn(any(schema).(core.ZodType[any, any]))
				}
			}
		}
	}
}

// NewBaseZodTypeInternals creates basic ZodTypeInternals structure
// Provides a foundation for custom schema type implementations
func NewBaseZodTypeInternals(typeName string) core.ZodTypeInternals {
	return core.ZodTypeInternals{
		Version: core.Version,
		Type:    typeName,
		Checks:  make([]core.ZodCheck, 0),
		Values:  make(map[any]struct{}),
		Bag:     make(map[string]any),
	}
}

// =============================================================================
// SCHEMA OPERATIONS
// =============================================================================

// AddCheck adds a validation check to any ZodType and returns new instance
// This is a generic function that works with any schema type implementing ZodType
// Enhanced version using slicex and mapx for better data management
func AddCheck[T interface{ GetInternals() *core.ZodTypeInternals }](schema T, check core.ZodCheck) core.ZodType[any, any] {
	internals := schema.GetInternals()

	// 1. 直接复制旧 checks 并追加新 check，避免 slicex.Append 类型转换问题
	newChecks := append(append([]core.ZodCheck(nil), internals.Checks...), check)

	// 2. 构造新的类型定义
	newDef := &core.ZodTypeDef{
		Type:   internals.Type,
		Error:  internals.Error,
		Checks: newChecks,
	}

	// 3. 通过原 Constructor 生成新 schema
	if internals.Constructor == nil {
		panic(fmt.Sprintf("No constructor found for type: %T", schema))
	}

	newSchema := internals.Constructor(newDef)
	newInternals := newSchema.GetInternals()

	// 4. 继承关键状态位
	newInternals.Nilable = internals.Nilable
	newInternals.Optional = internals.Optional
	newInternals.Coerce = internals.Coerce
	newInternals.Pattern = internals.Pattern

	// 5. 深拷贝 Values 与 Bag
	if len(internals.Values) > 0 {
		newInternals.Values = make(map[any]struct{}, len(internals.Values))
		for k, v := range internals.Values {
			newInternals.Values[k] = v
		}
	}
	if len(internals.Bag) > 0 {
		if newInternals.Bag == nil {
			newInternals.Bag = make(map[string]any)
		}
		newInternals.Bag = mapx.Merge(newInternals.Bag, internals.Bag)
	}

	// 6. 如果实现 Cloneable，复制特定状态
	if cloneable, ok := newSchema.(core.Cloneable); ok {
		if source, ok := any(schema).(core.Cloneable); ok {
			cloneable.CloneFrom(source)
		}
	}

	// 7. 执行 OnAttach 回调
	if check != nil {
		if ci := check.GetZod(); ci != nil {
			for _, fn := range ci.OnAttach {
				fn(newSchema)
			}
		}
	}

	return newSchema
}

// Clone creates a new instance of any ZodType with optional definition modifications
// This generic function provides deep cloning for any schema type
// Enhanced version using mapx for better state management
func Clone[T interface{ GetInternals() *core.ZodTypeInternals }](schema T, modifyDef func(*core.ZodTypeDef)) core.ZodType[any, any] {
	internals := schema.GetInternals()

	// Use slicex to safely copy checks
	var newChecks []core.ZodCheck
	if clonedChecks, err := slicex.ToTyped[core.ZodCheck](internals.Checks); err == nil {
		newChecks = clonedChecks
	} else {
		// Fallback to manual copy
		newChecks = append(make([]core.ZodCheck, len(internals.Checks)), internals.Checks...)
	}

	// Create new type definition as a base for cloning
	newDef := &core.ZodTypeDef{
		Type:   internals.Type,
		Error:  internals.Error,
		Checks: newChecks,
	}

	// Apply modifications if provided
	// This allows customization during cloning process
	if modifyDef != nil {
		modifyDef(newDef)
	}

	// Use existing constructor to create new instance
	if internals.Constructor != nil {
		newSchema := internals.Constructor(newDef)
		newInternals := newSchema.GetInternals()

		// Preserve important state flags
		newInternals.Nilable = internals.Nilable
		newInternals.Optional = internals.Optional
		newInternals.Coerce = internals.Coerce

		// Preserve pattern state
		if internals.Pattern != nil {
			newInternals.Pattern = internals.Pattern
		}

		// Use mapx to preserve values state with deep copy
		if len(internals.Values) > 0 {
			newInternals.Values = make(map[any]struct{}, len(internals.Values))
			for k, v := range internals.Values {
				newInternals.Values[k] = v
			}
		}

		// Use mapx to preserve bag state
		if len(internals.Bag) > 0 {
			if newInternals.Bag == nil {
				newInternals.Bag = make(map[string]any)
			}
			newInternals.Bag = mapx.Merge(newInternals.Bag, internals.Bag)
		}

		// Use Cloneable interface to copy type-specific state
		if cloneable, ok := newSchema.(core.Cloneable); ok {
			if sourceAny := any(schema); sourceAny != nil {
				if sourceCloneable, ok := sourceAny.(core.Cloneable); ok {
					cloneable.CloneFrom(sourceCloneable)
				}
			}
		}

		return newSchema
	}

	panic(fmt.Sprintf("No constructor found for type: %T", schema))
}

// IsOptionalField checks if a field is optional by checking its type
func IsOptionalField(schema any) bool {
	if schema == nil {
		return false
	}

	// Use reflection to check if it's a ZodOptional, ZodDefault, or ZodPrefault type
	schemaType := reflect.TypeOf(schema)
	if schemaType == nil {
		return false
	}

	// Check if the type name contains "Optional", "Default", or "Prefault"
	typeName := schemaType.String()
	return strings.Contains(typeName, "Optional") ||
		strings.Contains(typeName, "Default") ||
		strings.Contains(typeName, "Prefault")
}
