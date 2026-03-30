# Documentation Disguised as Code Anti-Patterns

> **Core Principle**: 约束 ≠ 代码 (Constraints ≠ Code)
>
> Development constraints should be enforced through SPEC + tests + lint rules, not translated into code.

## The Deletion Test

**Key Question**: If you delete this code, does program functionality change?

- **No** → Documentation disguised as code (should be deleted or moved to SPEC)
- **Yes** → Runtime/compile-time necessary code (keep it)

## Anti-Pattern Categories

### 1. Unused Type Definitions

**Pattern**: Types defined but never instantiated or used in production code.

**Characteristics**:
- Type exists only in definition file and test files
- No production code instantiates or references the type
- Type represents a concept from SPEC but has no runtime usage

**Detection**: Search for type usage across codebase. If only appears in definition and tests → unused.

**Fix**: Delete the type definition. Keep the concept in SPEC documentation only.

---

### 2. Write-Only Registries

**Pattern**: Registration functions that populate maps/registries, but no code ever queries them.

**Characteristics**:
- Registration function writes to map/registry
- No lookup, query, or iteration functions exist
- Descriptor objects with fields that are never read

**Detection**: Find all write operations, search for read operations. If reads == 0 → write-only.

**Fix**: Delete the registry and registration code. Document the pattern in SPEC if needed.

---

### 3. Validation Functions Never Called

**Pattern**: Validation functions defined but never invoked during actual data processing.

**Characteristics**:
- Validation methods only called in tests
- Production flow bypasses validation
- Schema validation defined but not wired into pipeline

**Detection**: Find validation function definition, search for call sites. If only in tests → never called in production.

**Fix**: Wire validation into production flow, or delete it and rely on framework/schema validation.

---

### 4. Enum Validation Duplication

**Pattern**: Manual validation functions that duplicate constraints already enforced by the framework.

**Characteristics**:
- Framework declares enum constraints (e.g., CLI options, schema definitions)
- Handler manually validates with switch/map lookup
- Both enforce the same constraint

**Detection**: Find enum declaration in framework config, find manual validation in handler. If both exist → duplication.

**Fix**: Delete manual validation. Trust the framework to enforce the constraint.

---

### 5. Mock-Only Interfaces

**Pattern**: Interfaces defined with no production implementations, only test mocks.

**Characteristics**:
- Interface defined in production code
- Only mock/stub implementations exist
- No real implementation in production codebase

**Detection**: Find interface definition, search for implementations excluding test files. If no production implementations → mock-only.

**Fix**: Delete the interface. Move to SPEC as a design note if it represents future functionality.

---

### 6. Premature Abstraction

**Pattern**: Abstraction layers created for a single consumer.

**Characteristics**:
- Service interfaces with only one implementation
- Adapter layers with one caller
- DTO types that duplicate existing types

**Detection**: Count consumers of the abstraction. If consumers == 1 → premature abstraction.

**Fix**: Inline the abstraction into the single consumer. Extract abstraction only when a second consumer appears.

---

### 7. Placeholder Implementations

**Pattern**: Functions that return hardcoded placeholders instead of real data.

**Characteristics**:
- Function returns hardcoded strings like "not implemented", "TODO", "placeholder"
- Function always returns the same value regardless of input
- No actual business logic implemented

**Detection**: Search for hardcoded placeholder strings. Check if function output varies with input.

**Fix**: Implement the function properly, or delete it and return an error stating the feature is not available.

---

### 8. Silent Fallback Validation

**Pattern**: Validation that silently accepts invalid input by falling back to a default.

**Characteristics**:
- Switch statements with default cases that don't return errors
- Validation functions that never return errors
- Invalid input silently converted to default value

**Detection**: Find switch statements with silent default cases. Find validation functions with no error returns.

**Fix**: Return an error for invalid input instead of silently falling back.

---

### 9. Field Existence Checks

**Pattern**: Validation that only checks if a field exists, not if its value is valid.

**Characteristics**:
- Validation checks map key existence only
- No subsequent validation of the value
- Field presence treated as sufficient validation

**Detection**: Find validation code that only checks key existence with no value validation.

**Fix**: Use JSON Schema validation or proper value validation. Delete the existence-only check.

---

### 10. Unused Error Sentinels

**Pattern**: Error constants defined but never returned by any function.

**Characteristics**:
- Error constant defined in production code
- No function returns this error
- Only referenced in tests

**Detection**: Find error definition, search for return sites. If only in definition and tests → unused.

**Fix**: Delete the unused error constant.

---

## Refactoring Strategy

### Phase 1: Zero-Risk Deletions

Delete code that has:
- Zero production usage (only tests reference it)
- No side effects
- No consumers

**Examples**: Unused types, write-only registries, mock-only interfaces

### Phase 2: Validation Wiring

For validation functions that should be called but aren't:
- Wire them into the production flow, OR
- Delete them and rely on framework/schema validation

### Phase 3: Abstraction Inlining

For premature abstractions:
- Inline single-consumer abstractions
- Move internal implementations to consumer's internal package
- Extract abstraction only when second consumer appears

### Phase 4: Documentation Migration

Move constraints from code to SPEC:
- Document valid enum values in SPEC
- Document validation rules in SPEC
- Add lint rules to enforce SPEC constraints

---

## Detection Checklist

When reviewing code, ask:

- [ ] Is this type ever instantiated in production code?
- [ ] Is this registry ever queried (not just written to)?
- [ ] Is this validation function called in the production flow?
- [ ] Does this enum validation duplicate framework constraints?
- [ ] Does this interface have production implementations (not just mocks)?
- [ ] Does this abstraction have multiple consumers?
- [ ] Does this function return real data or placeholders?
- [ ] Does this validation check values or just field existence?
- [ ] Is this error sentinel ever returned by a function?
- [ ] Does this validation reject invalid input or silently fall back?

If any answer is "No", you've found documentation disguised as code.
